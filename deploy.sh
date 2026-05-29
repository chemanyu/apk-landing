#!/usr/bin/env bash
# 一键部署脚本（在 Linux 服务器上执行）。
# 服务器只需 git，无需 Go —— 二进制已在本机交叉编译并提交进 git。
#
# 流程：git pull → 校验产物 → 安装/更新 systemd 服务 → 重启 → 健康检查
#
# 用法：
#   cd /home/sysadmin/data/apk-landing
#   ./deploy.sh
set -euo pipefail

# ── 配置 ────────────────────────────────────────────────────────
APP_NAME="apk-landing"
REPO_DIR="$(cd "$(dirname "$0")" && pwd)"   # 仓库根目录 = 脚本所在目录
DEPLOY_DIR="$REPO_DIR/dist-linux"           # 编译产物目录（service 路径与此一致）
PORT=5001
# ───────────────────────────────────────────────────────────────

cd "$REPO_DIR"

echo "==> [1/5] 拉取最新代码 (git pull)"
git pull --ff-only

echo "==> [2/5] 校验编译产物"
if [ ! -f "$DEPLOY_DIR/$APP_NAME" ]; then
  echo "❌ 找不到 $DEPLOY_DIR/$APP_NAME"
  echo "   请在本机执行：bash scripts/build-linux.sh && git add -A && git commit -m deploy && git push"
  exit 1
fi
chmod +x "$DEPLOY_DIR/$APP_NAME"
echo "    产物: $($DEPLOY_DIR/$APP_NAME --help >/dev/null 2>&1; echo OK) $(ls -la "$DEPLOY_DIR/$APP_NAME" | awk '{print $5" bytes "$6" "$7" "$8}')"

echo "==> [3/5] 安装/更新 systemd 服务"
sudo cp "$DEPLOY_DIR/${APP_NAME}.service" "/etc/systemd/system/${APP_NAME}.service"
sudo systemctl daemon-reload
sudo systemctl enable "$APP_NAME" >/dev/null 2>&1 || true

echo "==> [4/5] 重启服务"
sudo systemctl restart "$APP_NAME"
sleep 2

echo "==> [5/5] 健康检查"
sudo systemctl --no-pager --full status "$APP_NAME" | head -n 6 || true
echo "----"
if curl -fsS "http://127.0.0.1:${PORT}/ping" >/dev/null; then
  echo "✅ 部署成功：http://127.0.0.1:${PORT}/ping 正常"
else
  echo "❌ 健康检查失败，看日志：journalctl -u ${APP_NAME} -n 50 --no-pager"
  exit 1
fi
