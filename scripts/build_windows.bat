@echo off
REM BeeCount Cloud Windows构建脚本
REM 支持动态库分离，避免生成过大的单个可执行文件

setlocal enabledelayedexpansion

set "BINARY_NAME=beecount-cloud"
set "CMD_DIR=cmd\server"
set "MAIN_FILE=%CMD_DIR%\main.go"
set "BUILD_DIR=build"
set "LIB_DIR=lib\windows\amd64"
set "VERSION=%VERSION:latest%"

echo ========================================
echo BeeCount Cloud Windows构建脚本
echo ========================================
echo.

REM 检查CGO是否启用
for /f "delims=" %%i in ('go env CGO_ENABLED') do set CGO_ENABLED=%%i
if not "%CGO_ENABLED%"=="1" (
    echo 警告: CGO未启用，正在启用...
    go env -w CGO_ENABLED=1
)

REM 检查必要的工具
echo 检查构建环境...
where gcc >nul 2>&1
if %errorLevel% neq 0 (
    echo 错误: gcc未安装，请安装MSYS2
    echo 运行: setup_windows.bat mingw64
    exit /b 1
)

where go >nul 2>&1
if %errorLevel% neq 0 (
    echo 错误: go未安装
    exit /b 1
)

REM 创建构建目录
echo 创建构建目录...
if not exist "%BUILD_DIR%" mkdir "%BUILD_DIR%"
if not exist "%LIB_DIR%" mkdir "%LIB_DIR%"

REM 整理依赖
echo 整理依赖...
go mod tidy
go mod download

REM 构建应用程序（启用CGO，使用动态链接）
echo 构建应用程序...
set CGO_ENABLED=1
set GOOS=windows
set GOARCH=amd64
go build -o "%BUILD_DIR%\%BINARY_NAME%.exe" -ldflags="-w -s" -buildmode=exe "%MAIN_FILE%"

if %errorLevel% neq 0 (
    echo 构建失败
    exit /b 1
)

echo 构建成功: %BUILD_DIR%\%BINARY_NAME%.exe

REM 复制依赖的动态库（如果存在）
echo 复制依赖的动态库...

REM 查找并复制SQLite库
if exist "C:\msys64\mingw64\bin\libsqlite3-0.dll" (
    copy /Y "C:\msys64\mingw64\bin\libsqlite3-0.dll" "%LIB_DIR%\" >nul
    echo 已复制: libsqlite3-0.dll
)

REM 查找并复制其他依赖的DLL
REM 使用objdump查找依赖的DLL
where objdump >nul 2>&1
if %errorLevel% equ 0 (
    for /f "delims=" %%i in ('objdump -p "%BUILD_DIR%\%BINARY_NAME%.exe" ^| findstr /i "DLL Name:"') do (
        set "dll=%%i"
        set "dll=!dll:~10!"
        
        REM 检查是否是系统DLL
        echo !dll! | findstr /i "kernel32.dll user32.dll gdi32.dll ntdll.dll msvcrt.dll" >nul
        if %errorLevel% neq 0 (
            REM 尝试在MSYS2目录中找到DLL
            if exist "C:\msys64\mingw64\bin\!dll!" (
                copy /Y "C:\msys64\mingw64\bin\!dll!" "%LIB_DIR%\" >nul 2>&1
                if %errorLevel% equ 0 (
                    echo 已复制: !dll!
                )
            )
        )
    )
)

REM 创建部署包
echo 创建部署包...
set "DEPLOY_DIR=%BUILD_DIR%\deploy"
if not exist "%DEPLOY_DIR%" mkdir "%DEPLOY_DIR%"

REM 复制可执行文件
copy /Y "%BUILD_DIR%\%BINARY_NAME%.exe" "%DEPLOY_DIR%\" >nul

REM 复制动态库
if exist "%LIB_DIR%\*.dll" (
    xcopy /Y /I "%LIB_DIR%" "%DEPLOY_DIR%\lib\" >nul
)

REM 复制配置文件
copy /Y config.yaml "%DEPLOY_DIR%\" >nul

REM 复制ReDoc静态文件
if exist "web\redoc\index.html" (
    xcopy /Y /I /E "web\redoc" "%DEPLOY_DIR%\web\redoc\" >nul
    echo 已复制: ReDoc静态文件
)

REM 创建启动脚本
echo @echo off > "%DEPLOY_DIR%\start.bat"
echo REM BeeCount Cloud启动脚本 >> "%DEPLOY_DIR%\start.bat"
echo. >> "%DEPLOY_DIR%\start.bat"
echo REM 设置库路径（将lib目录添加到PATH） >> "%DEPLOY_DIR%\start.bat"
echo set "PATH=%%~dp0lib;%%PATH%%" >> "%DEPLOY_DIR%\start.bat"
echo. >> "%DEPLOY_DIR%\start.bat"
echo REM 启动服务 >> "%DEPLOY_DIR%\start.bat"
echo "%%~dp0%BINARY_NAME%.exe" >> "%DEPLOY_DIR%\start.bat"

REM 创建部署说明
echo # BeeCount Cloud 部署说明 > "%DEPLOY_DIR%\DEPLOY.md"
echo. >> "%DEPLOY_DIR%\DEPLOY.md"
echo ## 文件说明 >> "%DEPLOY_DIR%\DEPLOY.md"
echo. >> "%DEPLOY_DIR%\DEPLOY.md"
echo - `%BINARY_NAME%.exe`: 主程序可执行文件 >> "%DEPLOY_DIR%\DEPLOY.md"
echo - `lib\`: 依赖的动态库文件（DLL） >> "%DEPLOY_DIR%\DEPLOY.md"
echo - `config.yaml`: 配置文件 >> "%DEPLOY_DIR%\DEPLOY.md"
echo - `start.bat`: 启动脚本 >> "%DEPLOY_DIR%\DEPLOY.md"
echo - `web\`: ReDoc API文档（可选） >> "%DEPLOY_DIR%\DEPLOY.md"
echo. >> "%DEPLOY_DIR%\DEPLOY.md"
echo ## 部署步骤 >> "%DEPLOY_DIR%\DEPLOY.md"
echo. >> "%DEPLOY_DIR%\DEPLOY.md"
echo 1. 将所有文件上传到Windows服务器 >> "%DEPLOY_DIR%\DEPLOY.md"
echo 2. 根据需要修改 `config.yaml` >> "%DEPLOY_DIR%\DEPLOY.md"
echo 3. 双击运行 `start.bat` 启动服务 >> "%DEPLOY_DIR%\DEPLOY.md"
echo. >> "%DEPLOY_DIR%\DEPLOY.md"
echo ## 访问API文档 >> "%DEPLOY_DIR%\DEPLOY.md"
echo. >> "%DEPLOY_DIR%\DEPLOY.md"
echo 启动服务后，访问以下地址查看API文档： >> "%DEPLOY_DIR%\DEPLOY.md"
echo - ReDoc界面: http://localhost:8080/docs >> "%DEPLOY_DIR%\DEPLOY.md"
echo - OpenAPI JSON: http://localhost:8080/swagger/swagger.json >> "%DEPLOY_DIR%\DEPLOY.md"
echo. >> "%DEPLOY_DIR%\DEPLOY.md"
echo ## 环境要求 >> "%DEPLOY_DIR%\DEPLOY.md"
echo. >> "%DEPLOY_DIR%\DEPLOY.md"
echo - Windows 64位系统 >> "%DEPLOY_DIR%\DEPLOY.md"
echo - MSYS2运行时库（如果使用了SQLite数据库） >> "%DEPLOY_DIR%\DEPLOY.md"
echo. >> "%DEPLOY_DIR%\DEPLOY.md"
echo ## 手动启动 >> "%DEPLOY_DIR%\DEPLOY.md"
echo. >> "%DEPLOY_DIR%\DEPLOY.md"
echo 如果需要手动启动，确保lib目录与可执行文件在同一目录： >> "%DEPLOY_DIR%\DEPLOY.md"
echo. >> "%DEPLOY_DIR%\DEPLOY.md"
echo ```cmd >> "%DEPLOY_DIR%\DEPLOY.md"
echo set PATH=%%~dp0lib;%%PATH%% >> "%DEPLOY_DIR%\DEPLOY.md"
echo %BINARY_NAME%.exe >> "%DEPLOY_DIR%\DEPLOY.md"
echo ``` >> "%DEPLOY_DIR%\DEPLOY.md"
echo. >> "%DEPLOY_DIR%\DEPLOY.md"
echo ## 配置说明 >> "%DEPLOY_DIR%\DEPLOY.md"
echo. >> "%DEPLOY_DIR%\DEPLOY.md"
echo - API文档默认启用，可通过 `config.yaml` 中的 `server.docs.enabled` 关闭 >> "%DEPLOY_DIR%\DEPLOY.md"
echo - 或设置环境变量 `SERVER_DOCS_ENABLED=false` 关闭 >> "%DEPLOY_DIR%\DEPLOY.md"

REM 打包
echo 打包部署文件...
cd "%BUILD_DIR%"
powershell -Command "Compress-Archive -Path deploy -DestinationPath '%BINARY_NAME%-%VERSION%-windows-amd64.zip' -Force"
cd ..

echo.
echo ========================================
echo 构建完成！
echo ========================================
echo.
echo 可执行文件: %BUILD_DIR%\%BINARY_NAME%.exe
echo 动态库目录: %LIB_DIR%
echo 部署包: %BUILD_DIR%\%BINARY_NAME%-%VERSION%-windows-amd64.zip
echo.
echo 部署说明: %DEPLOY_DIR%\DEPLOY.md
echo.

endlocal
