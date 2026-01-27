# BeeCount Cloud Proto文件生成脚本（PowerShell）

Write-Host "=== BeeCount Cloud Proto文件生成脚本 ==="

# 记录启动脚本时的目录
$ORIGINAL_DIR = Get-Location

# 项目根目录
$ROOT_DIR = (Split-Path -Parent $PSScriptRoot) -replace '\\', '/'

# 定义安装命令常量
$INSTALL_CMD_PROTO_GEN_GO = "go install google.golang.org/protobuf/cmd/protoc-gen-go@latest"
$INSTALL_CMD_PROTO_GEN_GO_GRPC = "go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest"

# 检查protoc是否可用
if (-not (Get-Command protoc -ErrorAction SilentlyContinue)) {
    Write-Host "protoc未安装或未添加到环境变量，请先安装protoc和相关插件"
    Write-Host "安装命令："
    Write-Host "  $INSTALL_CMD_PROTO_GEN_GO"
    Write-Host "  $INSTALL_CMD_PROTO_GEN_GO_GRPC"
    Set-Location $ORIGINAL_DIR
    exit 1
}

# 检查protoc-gen-go是否可用
if (-not (Get-Command protoc-gen-go -ErrorAction SilentlyContinue)) {
    Write-Host "protoc-gen-go未安装，请先安装"
    Write-Host "安装命令：$INSTALL_CMD_PROTO_GEN_GO"
    Set-Location $ORIGINAL_DIR
    exit 1
}

# 检查protoc-gen-go-grpc是否可用
if (-not (Get-Command protoc-gen-go-grpc -ErrorAction SilentlyContinue)) {
    Write-Host "protoc-gen-go-grpc未安装，请先安装"
    Write-Host "安装命令：$INSTALL_CMD_PROTO_GEN_GO_GRPC"
    Set-Location $ORIGINAL_DIR
    exit 1
}

Write-Host "所有依赖工具已就绪"

# 切换到proto文件目录
Set-Location "$ROOT_DIR/common/proto"

Write-Host ""
Write-Host "开始生成proto文件"

# 生成common.proto文件
Write-Host "生成 common/common.proto..."
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative --proto_path=. common/common.proto
if ($LASTEXITCODE -ne 0) {
    Write-Host "生成 common/common.proto 失败"
    Set-Location $ORIGINAL_DIR
    exit 1
}
Write-Host "common/common.proto 生成完成"

# 生成auth.proto文件
Write-Host "生成 auth/auth.proto..."
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative --proto_path=. auth/auth.proto
if ($LASTEXITCODE -ne 0) {
    Write-Host "生成 auth/auth.proto 失败"
    Set-Location $ORIGINAL_DIR
    exit 1
}
Write-Host "auth/auth.proto 生成完成"

# 生成business.proto文件
Write-Host "生成 business/business.proto..."
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative --proto_path=. business/business.proto
if ($LASTEXITCODE -ne 0) {
    Write-Host "生成 business/business.proto 失败"
    Set-Location $ORIGINAL_DIR
    exit 1
}
Write-Host "business/business.proto 生成完成"

# 生成config.proto文件
Write-Host "生成 config/config.proto..."
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative --proto_path=. config/config.proto
if ($LASTEXITCODE -ne 0) {
    Write-Host "生成 config/config.proto 失败"
    Set-Location $ORIGINAL_DIR
    exit 1
}
Write-Host "config/config.proto 生成完成"

# 生成firewall.proto文件
Write-Host "生成 firewall/firewall.proto..."
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative --proto_path=. firewall/firewall.proto
if ($LASTEXITCODE -ne 0) {
    Write-Host "生成 firewall/firewall.proto 失败"
    Set-Location $ORIGINAL_DIR
    exit 1
}
Write-Host "firewall/firewall.proto 生成完成"

# 生成log.proto文件
Write-Host "生成 log/log.proto..."
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative --proto_path=. log/log.proto
if ($LASTEXITCODE -ne 0) {
    Write-Host "生成 log/log.proto 失败"
    Set-Location $ORIGINAL_DIR
    exit 1
}
Write-Host "log/log.proto 生成完成"

# 生成storage.proto文件
Write-Host "生成 storage/storage.proto..."
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative --proto_path=. storage/storage.proto
if ($LASTEXITCODE -ne 0) {
    Write-Host "生成 storage/storage.proto 失败"
    Set-Location $ORIGINAL_DIR
    exit 1
}
Write-Host "storage/storage.proto 生成完成"

Write-Host ""
Write-Host "所有proto文件生成完成"

# 回到原始目录
Set-Location $ORIGINAL_DIR

exit 0