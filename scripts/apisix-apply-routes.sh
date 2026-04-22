#!/usr/bin/env bash
# ============================================================
# apisix-apply-routes.sh — 将 components/apisix/routes/*.yaml 中的
# 结构配置幂等写入 etcd（通过 APISIX Admin API）。
#
# 用法:
#   ./scripts/apisix-apply-routes.sh              # 使用默认值
#   APISIX_ADMIN_URL=http://host:9180 \
#   APISIX_ADMIN_KEY=your-key \
#     ./scripts/apisix-apply-routes.sh
#
# 幂等：相同 id 的资源走 PUT，会被整体覆盖。
# 不会删除 etcd 中 YAML 未声明的资源（避免误删 UI 运行时改动）。
# 如需严格同步清理，加 --prune 标志（未实现，需显式确认后再做）。
# ============================================================
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROUTES_DIR="${ROUTES_DIR:-${SCRIPT_DIR}/../components/apisix/routes}"
APISIX_ADMIN_URL="${APISIX_ADMIN_URL:-http://localhost:9180}"
APISIX_ADMIN_KEY="${APISIX_ADMIN_KEY:-edd1c9f034335f136f87ad84b625c8f1}"

ADMIN="${APISIX_ADMIN_URL}/apisix/admin"

# 使用 docker 容器跑 yq，免去宿主机依赖
yq() {
  docker run --rm -i mikefarah/yq:4.45.1 "$@"
}

# 将 YAML 单个资源节点转为 JSON（剥离 id 字段）
to_json_body() {
  local yaml_chunk="$1"
  echo "$yaml_chunk" | yq -o=json 'del(.id)'
}

put_resource() {
  local resource="$1"  # routes / upstreams / services / ssl / consumers / global_rules
  local id="$2"
  local payload="$3"

  local http_code
  http_code=$(curl -s -o /tmp/apisix_apply_err.$$ -w '%{http_code}' \
    -X PUT "${ADMIN}/${resource}/${id}" \
    -H "X-API-KEY: ${APISIX_ADMIN_KEY}" \
    -H "Content-Type: application/json" \
    -d "${payload}")

  if [[ "${http_code}" =~ ^2 ]]; then
    echo "  [OK]  ${resource}/${id} (${http_code})"
    rm -f /tmp/apisix_apply_err.$$
  else
    echo "  [FAIL] ${resource}/${id} (${http_code})"
    cat /tmp/apisix_apply_err.$$ >&2
    echo "" >&2
    rm -f /tmp/apisix_apply_err.$$
    return 1
  fi
}

# Admin API 资源名到 YAML 顶层 key 的映射
# YAML key 复数形式（对齐 APISIX 社区习惯），Admin API 亦为复数
RESOURCES=(routes upstreams services ssl consumers global_rules)

apply_file() {
  local file="$1"
  echo "--> ${file}"
  for resource in "${RESOURCES[@]}"; do
    # 某资源类型在当前文件的条目数
    local count
    count=$(yq ".${resource} | length // 0" < "$file")
    if [[ "$count" -eq 0 ]]; then
      continue
    fi
    for i in $(seq 0 $((count - 1))); do
      local id
      id=$(yq -r ".${resource}[$i].id" < "$file")
      if [[ -z "$id" || "$id" == "null" ]]; then
        echo "  [SKIP] ${resource}[$i] missing id in ${file}" >&2
        continue
      fi
      local body
      body=$(yq -o=json "del(.${resource}[$i].id) | .${resource}[$i]" < "$file")
      put_resource "$resource" "$id" "$body"
    done
  done
}

echo "==> Apply APISIX routes to etcd via Admin API"
echo "    Source:    ${ROUTES_DIR}"
echo "    Admin URL: ${APISIX_ADMIN_URL}"
echo ""

shopt -s nullglob
files=("${ROUTES_DIR}"/*.yaml "${ROUTES_DIR}"/*.yml)
if [[ ${#files[@]} -eq 0 ]]; then
  echo "No YAML files found in ${ROUTES_DIR}" >&2
  exit 1
fi

for f in "${files[@]}"; do
  apply_file "$f"
done

echo ""
echo "==> Verify"
route_count=$(curl -s "${ADMIN}/routes" \
  -H "X-API-KEY: ${APISIX_ADMIN_KEY}" \
  | python3 -c "import sys,json; print(json.load(sys.stdin).get('total',0))" 2>/dev/null \
  || echo "?")
echo "    Total routes in etcd: ${route_count}"

echo ""
echo "==> Smoke test: GET /health"
health_code=$(curl -s -o /dev/null -w '%{http_code}' http://localhost:9080/health 2>/dev/null || echo "000")
if [[ "${health_code}" == "200" ]]; then
  echo "    [PASS] /health -> ${health_code}"
else
  echo "    [WARN] /health -> ${health_code} (APISIX not reachable or route not yet loaded)"
fi
