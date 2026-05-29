#!/usr/bin/env bash
# 在 Mac/Linux 上交叉编译出 Windows 可执行文件，连同配置/模板一起打包到 dist/。
# 用法：bash scripts/build-windows.sh
set -euo pipefail
cd "$(dirname "$0")/.."

OUT=dist
rm -rf "$OUT"
mkdir -p "$OUT"

echo ">> 交叉编译 windows/amd64 ..."
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 \
  GOPROXY=https://goproxy.cn,direct \
  go build -o "$OUT/apk-landing.exe" .

# 模板已通过 //go:embed 编进 exe，无需单独拷贝 templates/
mkdir -p "$OUT/etc"
cp etc/config.yaml "$OUT/etc/config.yaml"
cp scripts/run.bat "$OUT/run.bat"

echo ">> 打包完成，dist/ 内容："
ls -la "$OUT"
echo ""
echo "把整个 dist/ 目录拷到 Windows VM，双击 run.bat 即可启动。"
