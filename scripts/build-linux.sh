#!/usr/bin/env bash
# 构建 Linux 命令行版；需 Go 1.27.3+、Node 22+、C/C++ 编译器、SpiderMonkey 与 libwebp 静态库
set -euo pipefail
: "${SPIDERMONKEY_DIR:?set SPIDERMONKEY_DIR}"
: "${LIBWEBP_DIR:?set LIBWEBP_DIR}"
VERSION="${VERSION:-v0.3.15-uncut.1}"
export CGO_ENABLED=1
export CGO_CPPFLAGS="-I$SPIDERMONKEY_DIR/include -I$LIBWEBP_DIR/include"
export CGO_CXXFLAGS="$CGO_CPPFLAGS"
export CGO_LDFLAGS="-L$SPIDERMONKEY_DIR/lib -L$LIBWEBP_DIR/lib"
if [ ! -f internal/webui/dist/index.html ]; then
  npm ci --prefix web
  npm run build --prefix web
fi
go build -trimpath -ldflags "-X github.com/Mag1cFall/AIStudio2API/internal/app.version=$VERSION" -o aistudio2api ./cmd/aistudio2api
echo "built ./aistudio2api ($VERSION)"
