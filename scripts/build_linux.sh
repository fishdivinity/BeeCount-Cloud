#!/bin/bash
# BeeCount Cloud Linux构建脚本
# 支持动态库分离，避免生成过大的单个可执行文件

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 变量定义
BINARY_NAME="beecount-cloud"
CMD_DIR="./cmd/server"
MAIN_FILE="$CMD_DIR/main.go"
BUILD_DIR="./build"
LIB_DIR="./lib/linux/amd64"
VERSION=${VERSION:-"latest"}

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}BeeCount Cloud Linux构建脚本${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""

# 检查CGO是否启用
if [ "$(go env CGO_ENABLED)" != "1" ]; then
    echo -e "${YELLOW}警告: CGO未启用，正在启用...${NC}"
    go env -w CGO_ENABLED=1
fi

# 检查必要的工具
echo -e "${GREEN}检查构建环境...${NC}"
command -v gcc >/dev/null 2>&1 || { echo -e "${RED}错误: gcc未安装${NC}"; exit 1; }
command -v go >/dev/null 2>&1 || { echo -e "${RED}错误: go未安装${NC}"; exit 1; }

# 检查SQLite开发库
if ! pkg-config --exists sqlite3 2>/dev/null; then
    echo -e "${YELLOW}警告: SQLite开发库未安装${NC}"
    echo -e "${YELLOW}请运行: sudo apt-get install libsqlite3-dev${NC}"
    echo -e "${YELLOW}继续构建，但可能无法使用SQLite数据库${NC}"
fi

# 创建构建目录
echo -e "${GREEN}创建构建目录...${NC}"
mkdir -p "$BUILD_DIR"
mkdir -p "$LIB_DIR"

# 整理依赖
echo -e "${GREEN}整理依赖...${NC}"
go mod tidy
go mod download

# 构建应用程序（启用CGO，使用动态链接）
echo -e "${GREEN}构建应用程序...${NC}"
CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build \
    -o "$BUILD_DIR/$BINARY_NAME" \
    -ldflags="-w -s -linkmode external" \
    -buildmode=exe \
    "$MAIN_FILE"

if [ $? -eq 0 ]; then
    echo -e "${GREEN}构建成功: $BUILD_DIR/$BINARY_NAME${NC}"
else
    echo -e "${RED}构建失败${NC}"
    exit 1
fi

# 复制依赖的动态库（如果存在）
echo -e "${GREEN}复制依赖的动态库...${NC}"

# 检查并复制SQLite库
if pkg-config --exists sqlite3 2>/dev/null; then
    SQLITE_LIB=$(pkg-config --libs-only-L sqlite3 | sed 's/-L//g')
    if [ -n "$SQLITE_LIB" ]; then
        # 复制libsqlite3.so
        find "$SQLITE_LIB" -name "libsqlite3.so*" -type f 2>/dev/null | while read lib; do
            if [ -f "$lib" ]; then
                cp -v "$lib" "$LIB_DIR/"
            fi
        done
    fi
fi

# 检查并复制其他依赖的动态库
if [ -f "$BUILD_DIR/$BINARY_NAME" ]; then
    # 使用ldd查找依赖的动态库
    ldd "$BUILD_DIR/$BINARY_NAME" 2>/dev/null | grep "=>" | awk '{print $3}' | while read lib; do
        if [ -f "$lib" ]; then
            # 只复制非系统库
            if [[ ! "$lib" =~ ^/lib/|^/usr/lib/ ]]; then
                cp -v "$lib" "$LIB_DIR/" 2>/dev/null || true
            fi
        fi
    done
fi

# 创建部署包
echo -e "${GREEN}创建部署包...${NC}"
DEPLOY_DIR="$BUILD_DIR/deploy"
mkdir -p "$DEPLOY_DIR"

# 复制可执行文件
cp "$BUILD_DIR/$BINARY_NAME" "$DEPLOY_DIR/"

# 复制动态库
if [ "$(ls -A $LIB_DIR 2>/dev/null)" ]; then
    cp -r "$LIB_DIR" "$DEPLOY_DIR/lib"
fi

# 复制配置文件
cp config.yaml "$DEPLOY_DIR/"

# 复制ReDoc静态文件
if [ -d "web/redoc" ]; then
    cp -r web/redoc "$DEPLOY_DIR/"
    echo -e "${GREEN}已复制: ReDoc静态文件${NC}"
fi

# 创建启动脚本
cat > "$DEPLOY_DIR/start.sh" << 'EOF'
#!/bin/bash
# BeeCount Cloud启动脚本

# 设置库路径
export LD_LIBRARY_PATH="$(dirname "$0")/lib:$LD_LIBRARY_PATH"

# 启动服务
exec "$(dirname "$0")/beecount-cloud"
EOF
chmod +x "$DEPLOY_DIR/start.sh"

# 创建部署说明
cat > "$DEPLOY_DIR/DEPLOY.md" << 'EOF'
# BeeCount Cloud 部署说明
## 文件说明

- `beecount-cloud`: 主程序可执行文件
- `lib/`: 依赖的动态库文件
- `config.yaml`: 配置文件
- `start.sh`: 启动脚本
- `web/redoc/`: ReDoc API文档（可选）
## 部署步骤
1. 将所有文件上传到服务器
2. 根据需要修改 `config.yaml`
3. 运行 `chmod +x beecount-cloud start.sh`
4. 运行 `./start.sh` 启动服务
## 访问API文档
启动服务后，访问以下地址查看API文档：
- ReDoc界面: http://localhost:8080/docs
- OpenAPI JSON: http://localhost:8080/swagger/swagger.json
## 环境要求

- Linux 64位系统
- GCC运行时库
- SQLite运行时库（如果使用SQLite数据库）
## 手动启动
如果需要手动启动，确保设置库路径：
```bash
export LD_LIBRARY_PATH=/path/to/lib:$LD_LIBRARY_PATH
./beecount-cloud
```
## 配置说明
- API文档默认启用，可通过 `config.yaml` 中的 `server.docs.enabled` 关闭
- 或设置环境变量 `SERVER_DOCS_ENABLED=false` 关闭
EOF

# 打包
echo -e "${GREEN}打包部署文件...${NC}"
cd "$BUILD_DIR"
tar -czf "${BINARY_NAME}-${VERSION}-linux-amd64.tar.gz" deploy/
cd - > /dev/null

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}构建完成！${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo -e "可执行文件: ${GREEN}$BUILD_DIR/$BINARY_NAME${NC}"
echo -e "动态库目录: ${GREEN}$LIB_DIR${NC}"
echo -e "部署包: ${GREEN}$BUILD_DIR/${BINARY_NAME}-${VERSION}-linux-amd64.tar.gz${NC}"
echo ""
echo -e "${YELLOW}部署说明: $DEPLOY_DIR/DEPLOY.md${NC}"
echo ""
