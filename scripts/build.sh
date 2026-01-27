#!/bin/bash

# BeeCount Cloud 构建脚本（Linux/macOS）

set -e

# 默认版本号
VERSION="latest"

# 检查命令行参数
if [ $# -eq 1 ]; then
    VERSION="$1"
fi

# 项目根目录
ROOT_DIR=$(pwd)

# 服务列表和对应的可执行文件名
SERVICES=(gateway config auth business storage log firewall beecount)

# 服务到可执行文件的映射
declare -A SERVICE_TO_EXECUTABLE
SERVICE_TO_EXECUTABLE["gateway"]="gateway"
SERVICE_TO_EXECUTABLE["config"]="config"
SERVICE_TO_EXECUTABLE["auth"]="auth"
SERVICE_TO_EXECUTABLE["business"]="business"
SERVICE_TO_EXECUTABLE["storage"]="storage"
SERVICE_TO_EXECUTABLE["log"]="log"
SERVICE_TO_EXECUTABLE["firewall"]="firewall"
SERVICE_TO_EXECUTABLE["beecount"]="BeeCount-Cloud"

# 构建输出目录
BUILD_DIR="$ROOT_DIR/build"
APP_DIR="$BUILD_DIR/BeeCount-Cloud"
BIN_DIR="$APP_DIR/bin"

echo "=== BeeCount Cloud 构建脚本 ==="

# 创建构建目录
mkdir -p "$BUILD_DIR"
mkdir -p "$APP_DIR"
mkdir -p "$BIN_DIR"

echo "✅ 构建目录已创建: $BUILD_DIR"
echo "✅ 应用目录已创建: $APP_DIR"
echo "✅ 二进制文件目录已创建: $BIN_DIR"

# 构建所有服务
for service in "${SERVICES[@]}"; do
    executable=${SERVICE_TO_EXECUTABLE[$service]}
    echo "\n=== 构建服务: $service -> $executable ==="
    cd "$ROOT_DIR/services/$service"
    
    # 安装依赖
    go mod tidy
    
    # 构建服务
    go build -ldflags="-s -w" -o "$BIN_DIR/$executable" ./cmd
    
    echo "✅ $service 构建完成，生成 $executable"
done

# 复制资源文件到构建目录
echo "\n=== 复制资源文件 ==="

# 复制web目录
WEB_SRC="$ROOT_DIR/web"
WEB_DST="$APP_DIR/web"
if [ -d "$WEB_SRC" ]; then
    rm -rf "$WEB_DST" 2>/dev/null || true
    cp -r "$WEB_SRC" "$WEB_DST"
    echo "✅ 已复制 web 目录"
fi

# 复制config目录
CONFIG_SRC="$ROOT_DIR/config"
CONFIG_DST="$APP_DIR/config"
if [ -d "$CONFIG_SRC" ]; then
    rm -rf "$CONFIG_DST" 2>/dev/null || true
    cp -r "$CONFIG_SRC" "$CONFIG_DST"
    echo "✅ 已复制 config 目录"
fi

# 复制i18n目录
I18N_SRC="$ROOT_DIR/services/beecount/i18n"
I18N_DST="$APP_DIR/i18n"
if [ -d "$I18N_SRC" ]; then
    rm -rf "$I18N_DST" 2>/dev/null || true
    cp -r "$I18N_SRC" "$I18N_DST"
    echo "✅ 已复制 i18n 目录"
fi

# 压缩打包
echo "\n=== 压缩打包 ==="

ZIP_NAME="BeeCount-Cloud-$VERSION.zip"
ZIP_PATH="$BUILD_DIR/$ZIP_NAME"

# 检查是否已存在相同名称的压缩包，若存在则删除
if [ -f "$ZIP_PATH" ]; then
    rm -f "$ZIP_PATH"
    echo "⚠️  已删除同名压缩包: $ZIP_NAME"
fi

# 关键：确保压缩包内顶层目录是 "BeeCount-Cloud"
# 方法：在 BUILD_DIR 中创建压缩包，包含 BeeCount-Cloud 目录
cd "$BUILD_DIR"
zip -r "$ZIP_NAME" "BeeCount-Cloud"
cd "$ROOT_DIR"

echo "✅ 已创建压缩包: $ZIP_PATH"

# 清理中间文件夹
echo "\n=== 清理中间文件 ==="

rm -rf "$APP_DIR"
echo "✅ 已清理应用目录: $APP_DIR"

echo "\n=== 构建完成 ==="
echo "构建产物: $ZIP_PATH"
echo "压缩包大小: $(du -h "$ZIP_PATH" | cut -f1)"
ls -la "$ZIP_PATH"