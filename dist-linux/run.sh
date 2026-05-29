#!/usr/bin/env bash
# 在 Linux 服务器上前台启动 apk-landing（调试用；正式建议用 systemd，见 DEPLOY.md）。
# 用法：bash run.sh
set -euo pipefail
cd "$(dirname "$0")"

echo "============================================"
echo " 启动 apk-landing （监听 0.0.0.0:5001）"
echo " 公网经 nginx /ulink 反代到本服务"
echo " Ctrl+C 停止"
echo "============================================"

exec ./apk-landing -f etc/config.yaml
