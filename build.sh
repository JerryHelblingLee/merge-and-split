#!/bin/sh

# 用法：sh build.sh [系统] [架构]
# 默认系统为 linux，默认架构为 arm64（适配手机）
# 版本号固定为 1.0.0-nightly，打包时通过压缩包文件名区分版本

VERSION="1.0.0-nightly"
OS="${1:-linux}"
ARCH="${2:-arm64}"

# 判断系统，Windows 需要 .exe 后缀
if [ "$OS" = "windows" ]; then
  OUTPUT="merge-and-split.exe"
else
  OUTPUT="merge-and-split"
fi

# 编译完成后压缩包的文件名
ARCHIVE="merge-and-split-${VERSION}-${OS}-${ARCH}.7z"

echo "正在编译: 版本=$VERSION | 系统=$OS | 架构=$ARCH"

# 纯静态编译，不依赖系统动态库
if CGO_ENABLED=0 GOOS="$OS" GOARCH="$ARCH" go build -ldflags="-s -w -X main.version=$VERSION" -o "$OUTPUT" .; then
  echo "编译完成: $OUTPUT"
  echo "打包为: $ARCHIVE"
  7z a "$ARCHIVE" "$OUTPUT"
  echo "打包完成，可直接上传 Releases"
else
  echo "编译失败，请检查 Go 环境和代码"
  exit 1
fi
