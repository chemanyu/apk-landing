#!/usr/bin/env bash
# 在 Mac/Linux 上交叉编译出 Linux 可执行文件，连同配置一起打包到 dist-linux/。
# 用法：bash scripts/build-linux.sh
set -euo pipefail
cd "$(dirname "$0")/.."

OUT=dist-linux
rm -rf "$OUT"
mkdir -p "$OUT/etc"

echo ">> 交叉编译 linux/amd64 ..."
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
  GOPROXY=https://goproxy.cn,direct \
  go build -o "$OUT/apk-landing" .

# 模板已通过 //go:embed 编进二进制，无需单独拷贝 templates/
cp etc/config.yaml "$OUT/etc/config.yaml"
cp scripts/run.sh "$OUT/run.sh"
cp deploy/apk-landing.service "$OUT/apk-landing.service"
chmod +x "$OUT/apk-landing" "$OUT/run.sh"

echo ">> 打包完成，dist-linux/ 内容："
ls -la "$OUT"
echo ""
echo "把整个 dist-linux/ 拷到 Linux 服务器，详见 DEPLOY.md。"
