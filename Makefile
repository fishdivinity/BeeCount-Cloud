# BeeCount Cloud Makefile

.PHONY: help build run test clean tidy docker-build docker-run build-windows build-linux build-darwin

# 默认目标
.DEFAULT_GOAL := help

# 变量定义
BINARY_NAME=beecount-cloud
CMD_DIR=./cmd/server
MAIN_FILE=$(CMD_DIR)/main.go
BUILD_DIR=./build
VERSION?=latest

# Windows环境变量
ifeq ($(OS),Windows_NT)
    EXE_EXT=.exe
    RM=del /Q
    MKDIR=mkdir
    PATH_SEP=\\
else
    EXE_EXT=
    RM=rm -f
    MKDIR=mkdir -p
    PATH_SEP=/
endif

# 帮助信息
help: ## 显示帮助信息
	@echo "BeeCount Cloud - 可用命令:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "构建命令:"
	@echo "  build              - 构建当前平台（简单构建）"
	@echo "  build-windows       - 构建Windows平台（动态库分离）"
	@echo "  build-linux        - 构建Linux平台（动态库分离）"
	@echo "  build-darwin       - 构建macOS平台（动态库分离）"
	@echo ""

# 简单构建（不分离动态库）
build: ## 构建项目（简单构建，不分离动态库）
	@echo "构建 $(BINARY_NAME)..."
	@$(MKDIR) $(BUILD_DIR) 2>/dev/null || true
	@go build -o $(BUILD_DIR)/$(BINARY_NAME)$(EXE_EXT) $(MAIN_FILE)
	@echo "构建完成: $(BUILD_DIR)/$(BINARY_NAME)$(EXE_EXT)"

# Windows构建（动态库分离）
build-windows: ## 构建Windows平台（动态库分离）
	@echo "构建Windows平台（动态库分离）..."
	@if exist scripts\build_windows.bat (
		@scripts\build_windows.bat
	) else (
		@echo "错误: scripts\build_windows.bat 不存在"
		@exit /b 1
	)

# Linux构建（动态库分离）
build-linux: ## 构建Linux平台（动态库分离）
	@echo "构建Linux平台（动态库分离）..."
	@if [ -f scripts/build_linux.sh ]; then \
		chmod +x scripts/build_linux.sh; \
		./scripts/build_linux.sh; \
	else \
		echo "错误: scripts/build_linux.sh 不存在"; \
		exit 1; \
	fi

# macOS构建（动态库分离）
build-darwin: ## 构建macOS平台（动态库分离）
	@echo "构建macOS平台（动态库分离）..."
	@$(MKDIR) $(BUILD_DIR) 2>/dev/null || true
	@$(MKDIR) lib/darwin/amd64 2>/dev/null || true
	@echo "整理依赖..."
	@go mod tidy
	@go mod download
	@echo "构建应用程序..."
	@CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build \
		-o $(BUILD_DIR)/$(BINARY_NAME) \
		-ldflags="-w -s -linkmode external" \
		-buildmode=exe \
		$(MAIN_FILE)
	@echo "构建成功: $(BUILD_DIR)/$(BINARY_NAME)"
	@echo "复制依赖的动态库..."
	@if [ -f $(BUILD_DIR)/$(BINARY_NAME) ]; then \
		otool -L $(BUILD_DIR)/$(BINARY_NAME) | grep -v "/usr/lib" | grep -v "/System" | awk '{print $$1}' | while read lib; do \
			if [ -f "$$lib" ]; then \
				cp -v "$$lib" lib/darwin/amd64/ 2>/dev/null || true; \
			fi; \
		done; \
	fi

# 运行项目
run: ## 运行项目
	@echo "运行 $(BINARY_NAME)..."
	@go run $(MAIN_FILE)

# 运行测试
test: ## 运行测试
	@echo "运行测试..."
	@go test -v ./...

# 生成测试覆盖率
test-coverage: ## 生成测试覆盖率报告
	@echo "生成测试覆盖率..."
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "覆盖率报告已生成: coverage.html"

# 清理构建文件
clean: ## 清理构建文件
	@echo "清理构建文件..."
	@$(RM) $(BUILD_DIR)/*$(EXE_EXT) 2>/dev/null || true
	@$(RM) coverage.out coverage.html 2>/dev/null || true
	@echo "清理完成"

# 清理所有构建产物（包括lib目录）
clean-all: ## 清理所有构建产物（包括lib目录）
	@echo "清理所有构建产物..."
	@$(RM) -rf $(BUILD_DIR) 2>/dev/null || true
	@$(RM) -rf lib 2>/dev/null || true
	@$(RM) coverage.out coverage.html 2>/dev/null || true
	@echo "清理完成"

# 整理依赖
tidy: ## 整理依赖
	@echo "整理依赖..."
	@go mod tidy
	@go mod verify
	@echo "依赖整理完成"

# 下载依赖
deps: ## 下载依赖
	@echo "下载依赖..."
	@go mod download
	@echo "依赖下载完成"

# 格式化代码
fmt: ## 格式化代码
	@echo "格式化代码..."
	@go fmt ./...
	@echo "代码格式化完成"

# 代码检查
lint: ## 代码检查
	@echo "代码检查..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint 未安装，跳过代码检查"; \
	fi

# 生成Swagger文档
swagger: ## 生成Swagger文档
	@echo "生成Swagger文档..."
	@if command -v swag >/dev/null 2>&1; then \
		swag init -g $(MAIN_FILE) -o docs/swagger --parseDependency; \
		echo "Swagger文档生成完成"; \
	else \
		echo "swag 未安装，请先运行: go install github.com/swaggo/swag/cmd/swag@latest"; \
	fi

# 下载ReDoc
redoc: ## 下载ReDoc静态文件
	@echo "下载ReDoc静态文件..."
	@mkdir -p web/redoc
	@curl -o web/redoc/index.html https://raw.githubusercontent.com/Redocly/redoc/main/bundles/redoc.standalone.js.html 2>/dev/null || echo "下载失败，请手动创建 web/redoc/index.html"
	@echo "ReDoc静态文件下载完成"

# Docker构建
docker-build: ## 构建Docker镜像
	@echo "构建Docker镜像..."
	@docker build -t $(BINARY_NAME):latest .
	@echo "Docker镜像构建完成"

# Docker运行
docker-run: ## 运行Docker容器
	@echo "运行Docker容器..."
	@docker-compose up -d

# Docker停止
docker-stop: ## 停止Docker容器
	@echo "停止Docker容器..."
	@docker-compose down

# Windows环境检查
check-windows: ## 检查Windows环境配置
	@echo "检查Windows环境配置..."
	@echo "CGO_ENABLED: $$(go env CGO_ENABLED)"
	@echo "GOOS: $$(go env GOOS)"
	@echo "GOARCH: $$(go env GOARCH)"
	@echo "GCC版本:"
	@gcc --version 2>/dev/null || echo "GCC 未找到，请安装MSYS2"
	@echo ""
	@echo "检查完成"

# Linux环境检查
check-linux: ## 检查Linux环境配置
	@echo "检查Linux环境配置..."
	@echo "CGO_ENABLED: $$(go env CGO_ENABLED)"
	@echo "GOOS: $$(go env GOOS)"
	@echo "GOARCH: $$(go env GOARCH)"
	@echo "GCC版本:"
	@gcc --version 2>/dev/null || echo "GCC 未找到，请运行: sudo apt-get install build-essential"
	@echo "SQLite开发库:"
	@pkg-config --cflags sqlite3 2>/dev/null || echo "SQLite开发库未找到，请运行: sudo apt-get install libsqlite3-dev"
	@echo ""
	@echo "检查完成"

# Windows环境切换到MINGW64
switch-mingw64: ## 切换到MINGW64环境（仅Windows）
ifeq ($(OS),Windows_NT)
	@echo "切换到MINGW64环境..."
	@echo "请手动将以下路径添加到系统PATH环境变量:"
	@echo "C:\msys64\mingw64\bin"
	@echo "或者使用配置脚本: scripts\setup_windows.bat mingw64"
	@echo "然后重启终端"
else
	@echo "此命令仅在Windows上可用"
endif

# Windows环境切换到UCRT64
switch-ucrt64: ## 切换到UCRT64环境（仅Windows）
ifeq ($(OS),Windows_NT)
	@echo "切换到UCRT64环境..."
	@echo "请手动将以下路径添加到系统PATH环境变量:"
	@echo "C:\msys64\ucrt64\bin"
	@echo "或者使用配置脚本: scripts\setup_windows.bat ucrt64"
	@echo "然后重启终端"
else
	@echo "此命令仅在Windows上可用"
endif

# 完整构建流程（检查+整理+构建）
all: check-$(OS) tidy build ## 完整构建流程
	@echo "完整构建流程完成"

# 打包部署文件
package: ## 打包部署文件
	@echo "打包部署文件..."
	@if [ "$(OS)" = "Windows_NT" ]; then \
		echo "Windows平台，请使用 build-windows.bat 进行打包"; \
	else \
		echo "Linux/macOS平台，请使用 build_linux.sh 进行打包"; \
	fi
