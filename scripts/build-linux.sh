#!/usr/bin/env bash
# 构建 Linux 命令行版；需先安装 Go 1.27+、Node 22+、gcc，以及 SpiderMonkey/libwebp 静态库
set -euo pipefail

: "${SPIDERMONKEY_DIR:?请设置 SPIDERMONKEY_DIR}"
: "${LIBWEBP_DIR:?请设置 LIBWEBP_DIR}"
VERSION="${VERSION:-v0.3.1-uncut.5}"

export CGO_ENABLED=1
export CGO_CPPFLAGS="-I$SPIDERMONKEY_DIR/include -I$LIBWEBP_DIR/include"
export CGO_CXXFLAGS="$CGO_CPPFLAGS"
export CGO_LDFLAGS="-L$SPIDERMONKEY_DIR/lib -L$LIBWEBP_DIR/lib"
export GOFLAGS="${GOFLAGS:--p=4}"

if [ ! -f internal/webui/dist/index.html ]; then
  npm ci --prefix web
  npm run build --prefix web
fi

go build -trimpath \
  -ldflags "-X github.com/Mag1cFall/AIStudio2API/internal/app.version=$VERSION" \
  -o aistudio2api ./cmd/aistudio2api

echo "built ./aistudio2api (version $VERSION)"
