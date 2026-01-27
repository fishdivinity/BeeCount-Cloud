# BeeCount Cloud 微服务架构构建脚本

# 项目根目录
ROOT_DIR := $(CURDIR)
# 构建输出目录
BUILD_DIR := $(ROOT_DIR)/build
# 二进制文件目录
BIN_DIR := $(BUILD_DIR)/bin

# 服务列表和对应的可执行文件名
SERVICES := gateway config auth business storage log firewall beecount
EXECUTABLES := gateway config auth business storage log firewall BeeCount-Cloud

# 服务到可执行文件的映射
SERVICE_TO_EXECUTABLE := \
	gateway:gateway \
	config:config \
	auth:auth \
	business:business \
	storage:storage \
	log:log \
	firewall:firewall \
	beecount:BeeCount-Cloud

# 默认目标
all: build

# 构建所有服务
build: $(SERVICES) copy_resources

# 复制资源文件到构建目录
copy_resources:
	@echo "Copying resources to build directory..."
	@mkdir -p $(BUILD_DIR)
	@# 复制web目录
	@cp -r $(ROOT_DIR)/web $(BUILD_DIR)/ 2>/dev/null || xcopy $(ROOT_DIR)\web $(BUILD_DIR)\web /E /I /Y 2>NUL
	@# 复制config目录
	@cp -r $(ROOT_DIR)/config $(BUILD_DIR)/ 2>/dev/null || xcopy $(ROOT_DIR)\config $(BUILD_DIR)\config /E /I /Y 2>NUL
	@# 复制i18n目录
	@cp -r $(ROOT_DIR)/services/beecount/i18n $(BUILD_DIR)/ 2>/dev/null || xcopy $(ROOT_DIR)\services\beecount\i18n $(BUILD_DIR)\i18n /E /I /Y 2>NUL

# 构建单个服务
$(SERVICES):
	@echo "Building $@..."
	@mkdir -p $(BIN_DIR)
	@if [ "$@" = "beecount" ]; then \
		if [ "$(OS)" = "Windows_NT" ] || [ "$$(uname)" = "MINGW32_NT" ] || [ "$$(uname)" = "MINGW64_NT" ] || [ "$$(uname)" = "CYGWIN_NT" ]; then \
			cd $(ROOT_DIR)/services/$@ && go build -ldflags="-s -w" -o $(BIN_DIR)/BeeCount-Cloud.exe ./cmd; \
		else \
			cd $(ROOT_DIR)/services/$@ && go build -ldflags="-s -w" -o $(BIN_DIR)/BeeCount-Cloud ./cmd; \
		fi \
	else \
		if [ "$(OS)" = "Windows_NT" ] || [ "$$(uname)" = "MINGW32_NT" ] || [ "$$(uname)" = "MINGW64_NT" ] || [ "$$(uname)" = "CYGWIN_NT" ]; then \
			cd $(ROOT_DIR)/services/$@ && go build -ldflags="-s -w" -o $(BIN_DIR)/$@.exe ./cmd; \
		else \
			cd $(ROOT_DIR)/services/$@ && go build -ldflags="-s -w" -o $(BIN_DIR)/$@ ./cmd; \
		fi \
	fi

# 清理构建产物
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)

# 启动所有服务
start: build
	@echo "Starting all services..."
	@if [ -f "$(BIN_DIR)/BeeCount-Cloud.exe" ]; then \
		$(BIN_DIR)/BeeCount-Cloud.exe --all; \
	else \
		$(BIN_DIR)/BeeCount-Cloud --all; \
	fi

# 停止所有服务
stop:
	@echo "Stopping all services..."
	@if [ "$(OS)" = "Windows_NT" ] || [ "$$(uname)" = "MINGW32_NT" ] || [ "$$(uname)" = "MINGW64_NT" ] || [ "$$(uname)" = "CYGWIN_NT" ]; then \
		taskkill /f /im BeeCount-Cloud.exe; \
	else \
		pkill -f "BeeCount-Cloud"; \
	fi

# 运行测试
test:
	@echo "Running tests..."
	@for service in $(SERVICES); do \
		cd $(ROOT_DIR)/services/$$service && go test ./...; \
	done

# 格式化代码
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# 检查代码
lint:
	@echo "Linting code..."
	@golangci-lint run

# 安装依赖
mod-tidy:
	@echo "Installing dependencies..."
	@for service in $(SERVICES); do \
		cd $(ROOT_DIR)/services/$$service && go mod tidy; \
	done
	@cd $(ROOT_DIR)/common && go mod tidy

# 显示帮助
help:
	@echo "Usage: make [target]"
	@echo "Targets:"
	@echo "  all        Build all services"
	@echo "  build      Build all services"
	@echo "  $(SERVICES) Build a specific service"
	@echo "  clean      Clean build artifacts"
	@echo "  start      Start all services"
	@echo "  stop       Stop all services"
	@echo "  test       Run tests"
	@echo "  fmt        Format code"
	@echo "  lint       Lint code"
	@echo "  mod-tidy   Install dependencies"
	@echo "  help       Show this help"

.PHONY: all build $(SERVICES) clean start stop test fmt lint mod-tidy help
