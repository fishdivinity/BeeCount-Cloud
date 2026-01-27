# BeeCount Cloud 构建脚本（PowerShell）

param(
    [string]$Version = "latest"
)

Write-Host "=== BeeCount Cloud 构建脚本 ==="

# 项目根目录
$ROOT_DIR = (Split-Path -Parent $PSScriptRoot) -replace '\\', '/'

# 构建输出目录
$BUILD_DIR = "$ROOT_DIR/build"
$APP_DIR = "$BUILD_DIR/BeeCount-Cloud"
$BIN_DIR = "$APP_DIR/bin"

# 创建构建目录
if (-not (Test-Path $BUILD_DIR)) {
    New-Item -ItemType Directory -Path $BUILD_DIR -Force | Out-Null
    Write-Host "✅ 构建目录已创建: $BUILD_DIR"
}

if (-not (Test-Path $APP_DIR)) {
    New-Item -ItemType Directory -Path $APP_DIR -Force | Out-Null
    Write-Host "✅ 应用目录已创建: $APP_DIR"
}

if (-not (Test-Path $BIN_DIR)) {
    New-Item -ItemType Directory -Path $BIN_DIR -Force | Out-Null
    Write-Host "✅ 二进制文件目录已创建: $BIN_DIR"
}

# 定义服务列表和可执行文件映射
$SERVICE_MAPPING = @{
    "gateway"      = "gateway"
    "config"       = "config"
    "auth"         = "auth"
    "business"     = "business"
    "storage"      = "storage"
    "log"          = "log"
    "firewall"     = "firewall"
    "beecount"     = "BeeCount-Cloud"
}

# 构建所有服务
foreach ($service in $SERVICE_MAPPING.Keys) {
    $executable = $SERVICE_MAPPING[$service]
    Write-Host ""
    Write-Host "=== 构建服务: $service -> $executable ==="

    # 切换到服务目录
    $service_dir = "$ROOT_DIR/services/$service"
    Set-Location $service_dir

    # 安装依赖
    Write-Host "安装依赖..."
    go mod tidy
    if ($LASTEXITCODE -ne 0) {
        Write-Host "❌ 安装依赖失败"
        exit 1
    }

    # 构建服务
    Write-Host "构建服务..."
    go build -ldflags='-s -w' -o "$BIN_DIR/$executable.exe" ./cmd
    if ($LASTEXITCODE -ne 0) {
        Write-Host "❌ 构建服务失败"
        exit 1
    }

    Write-Host "✅ $service 构建完成，生成 $executable.exe"
}

# 复制资源文件到构建目录
Write-Host ""
Write-Host "=== 复制资源文件 ==="

# 复制web目录
$WEB_SRC = "$ROOT_DIR/web"
$WEB_DST = "$APP_DIR/web"
if (Test-Path $WEB_SRC) {
    if (Test-Path $WEB_DST) {
        Remove-Item -Path $WEB_DST -Recurse -Force
    }
    Copy-Item -Path $WEB_SRC -Destination $WEB_DST -Recurse -Force
    Write-Host "✅ 已复制 web 目录"
}

# 复制config目录
$CONFIG_SRC = "$ROOT_DIR/config"
$CONFIG_DST = "$APP_DIR/config"
if (Test-Path $CONFIG_SRC) {
    if (Test-Path $CONFIG_DST) {
        Remove-Item -Path $CONFIG_DST -Recurse -Force
    }
    Copy-Item -Path $CONFIG_SRC -Destination $CONFIG_DST -Recurse -Force
    Write-Host "✅ 已复制 config 目录"
}

# 复制i18n目录
$I18N_SRC = "$ROOT_DIR/services/beecount/i18n"
$I18N_DST = "$APP_DIR/i18n"
if (Test-Path $I18N_SRC) {
    if (Test-Path $I18N_DST) {
        Remove-Item -Path $I18N_DST -Recurse -Force
    }
    Copy-Item -Path $I18N_SRC -Destination $I18N_DST -Recurse -Force
    Write-Host "✅ 已复制 i18n 目录"
}

# 压缩打包
Write-Host ""
Write-Host "=== 压缩打包 ==="

$ZIP_NAME = "BeeCount-Cloud-$Version.zip"
$ZIP_PATH = "$BUILD_DIR/$ZIP_NAME"

# 检查是否已存在相同名称的压缩包，若存在则删除
if (Test-Path $ZIP_PATH) {
    Remove-Item -Path $ZIP_PATH -Force
    Write-Host "⚠️  已删除同名压缩包: $ZIP_NAME"
}

# 关键：必须确保压缩包内顶层目录是 "BeeCount-Cloud"
# 方法：将 $APP_DIR 移动到临时目录，然后压缩整个临时目录
$TEMP_ZIP_DIR = "$ROOT_DIR/temp_zip_$([Guid]::NewGuid().ToString().Substring(0,8))"

try {
    # 创建临时目录
    New-Item -ItemType Directory -Path $TEMP_ZIP_DIR -Force | Out-Null
    
    # 将整个 $APP_DIR 目录移动到临时目录中
    Move-Item -Path $APP_DIR -Destination $TEMP_ZIP_DIR -Force
    
    # 压缩整个临时目录（这样顶层就是 BeeCount-Cloud 目录）
    Add-Type -AssemblyName System.IO.Compression.FileSystem
    [System.IO.Compression.ZipFile]::CreateFromDirectory($TEMP_ZIP_DIR, $ZIP_PATH)
    Write-Host "✅ 已创建压缩包: $ZIP_PATH"
}
finally {
    # 清理临时目录
    if (Test-Path $TEMP_ZIP_DIR) {
        Remove-Item -Path $TEMP_ZIP_DIR -Recurse -Force -ErrorAction SilentlyContinue
    }
}

Write-Host ""
Write-Host "=== 构建完成 ==="
Write-Host "构建产物: $ZIP_PATH"
Write-Host "压缩包大小: $([math]::Round((Get-Item $ZIP_PATH).Length / 1MB, 2)) MB"

exit 0