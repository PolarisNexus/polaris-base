#!/usr/bin/env bash
# ============================================================
# apisix-export-routes.sh — 从 etcd 导出 APISIX 结构配置为 YAML。
#
# 用途：运维在 platform-admin UI 中改了结构配置（Route/Upstream/SSL/
# Consumer/Service/GlobalRule）后，导出当前状态 → diff `components/
# apisix/routes/` → 人工确认 → 提交 PR 回流 Git 源。
#
# 用法:
#   ./scripts/apisix-export-routes.sh                  # 输出到 stdout
#   ./scripts/apisix-export-routes.sh > snapshot.yaml
#   APISIX_ADMIN_URL=http://host:9180 \
#   APISIX_ADMIN_KEY=your-key \
#     ./scripts/apisix-export-routes.sh
# ============================================================
set -euo pipefail

APISIX_ADMIN_URL="${APISIX_ADMIN_URL:-http://localhost:9180}"
APISIX_ADMIN_KEY="${APISIX_ADMIN_KEY:-edd1c9f034335f136f87ad84b625c8f1}"
ADMIN="${APISIX_ADMIN_URL}/apisix/admin"

yq() {
  docker run --rm -i mikefarah/yq:4.45.1 "$@"
}

# 从 Admin API list 响应中抽取资源数组 → 过滤到用户关心字段 → 作为 YAML 输出
fetch_to_yaml() {
  local admin_path="$1"  # routes / upstreams / services / ssl / consumers / global_rules
  local yaml_key="$2"

  local json
  json=$(curl -s "${ADMIN}/${admin_path}" -H "X-API-KEY: ${APISIX_ADMIN_KEY}")

  # APISIX 返回: { total: N, list: [ { key, value: {id, ...}, createdIndex, modifiedIndex }, ... ] }
  # 我们只要 value 部分，并按 id 排序
  echo "$json" | yq -P -o=yaml \
    ".list // [] | map(.value) | sort_by(.id) | { \"${yaml_key}\": . }"
}

{
  echo "# Exported from APISIX etcd via Admin API"
  echo "# Source: ${APISIX_ADMIN_URL}"
  echo "# Generated: $(date -u +%Y-%m-%dT%H:%M:%SZ)"
  echo ""
  fetch_to_yaml "routes"        "routes"
  echo ""
  fetch_to_yaml "upstreams"     "upstreams"
  echo ""
  fetch_to_yaml "services"      "services"
  echo ""
  fetch_to_yaml "ssl"           "ssl"
  echo ""
  fetch_to_yaml "consumers"     "consumers"
  echo ""
  fetch_to_yaml "global_rules"  "global_rules"
}
