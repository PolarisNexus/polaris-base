#!/usr/bin/env bash
# APISIX 健康检查脚本
# 通过 bash 内置 /dev/tcp 发真实 HTTP 请求并校验状态码
# 覆盖 TCP + HTTP 响应 + 状态码三层，避免只测端口 listen 的盲点
set -euo pipefail
exec 3<>/dev/tcp/localhost/9080
printf 'GET /health HTTP/1.0\r\nHost: localhost\r\n\r\n' >&3
grep -q '200 OK' <&3
