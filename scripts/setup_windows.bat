@echo off
REM BeeCount Cloud Windows环境配置脚本
REM 用于切换MSYS2环境（MINGW64或UCRT64）

setlocal enabledelayedexpansion

if "%1"=="" (
    echo 用法: setup_windows.bat [mingw64^|ucrt64] [msys2_path]
    echo.
    echo 示例:
    echo   setup_windows.bat mingw64                    - 切换到MINGW64环境（默认路径）
    echo   setup_windows.bat ucrt64                      - 切换到UCRT64环境（默认路径）
    echo   setup_windows.bat mingw64 D:\msys64           - 切换到MINGW64环境（自定义路径）
    echo   setup_windows.bat ucrt64 E:\devtools\msys64   - 切换到UCRT64环境（自定义路径）
    echo.
    goto :eof
)

set "TARGET_ENV=%1"

REM 如果提供了第二个参数，使用自定义路径，否则使用默认路径
if "%2"=="" (
    set "MSYS2_PATH=C:\msys64"
) else (
    set "MSYS2_PATH=%2"
)

echo 检查MSYS2路径: %MSYS2_PATH%

if not exist "%MSYS2_PATH%" (
    echo 错误: MSYS2未安装在 %MSYS2_PATH%
    echo.
    echo 请从 https://www.msys2.org/ 下载并安装MSYS2
    echo.
    echo 如果MSYS2安装在其他位置，请提供完整路径作为第二个参数:
    echo   setup_windows.bat mingw64 D:\你的\msys2\路径
    exit /b 1
)

if /i "%TARGET_ENV%"=="mingw64" (
    set "BIN_PATH=%MSYS2_PATH%\mingw64\bin"
    echo 切换到MINGW64环境...
) else if /i "%TARGET_ENV%"=="ucrt64" (
    set "BIN_PATH=%MSYS2_PATH%\ucrt64\bin"
    echo 切换到UCRT64环境...
) else (
    echo 错误: 不支持的环境 '%TARGET_ENV%'
    echo 支持的环境: mingw64, ucrt64
    exit /b 1
)

if not exist "%BIN_PATH%" (
    echo 错误: 路径不存在 %BIN_PATH%
    echo 请确认MSYS2安装正确
    exit /b 1
)

echo.
echo 正在配置环境变量...
echo.

REM 检查是否需要管理员权限
net session >nul 2>&1
if %errorLevel% == 0 (
    echo 以管理员权限运行，将修改系统PATH环境变量

    REM 获取当前系统PATH
    for /f "tokens=2*" %%A in ('reg query "HKLM\SYSTEM\CurrentControlSet\Control\Session Manager\Environment" /v Path 2^>nul') do set "SYSTEM_PATH=%%B"

    REM 检查是否已经包含目标路径
    echo !SYSTEM_PATH! | findstr /C:"%BIN_PATH%" >nul
    if %errorLevel% == 0 (
        echo 目标路径已在系统PATH中
    ) else (
        echo 添加 %BIN_PATH% 到系统PATH...
        setx PATH "!SYSTEM_PATH!;%BIN_PATH%" /M >nul
        echo 系统PATH已更新
    )

    REM 获取当前用户PATH
    for /f "tokens=2*" %%A in ('reg query "HKCU\Environment" /v Path 2^>nul') do set "USER_PATH=%%B"

    REM 检查是否已经包含目标路径
    echo !USER_PATH! | findstr /C:"%BIN_PATH%" >nul
    if %errorLevel% == 0 (
        echo 目标路径已在用户PATH中
    ) else (
        echo 添加 %BIN_PATH% 到用户PATH...
        setx PATH "!USER_PATH!;%BIN_PATH%" >nul
        echo 用户PATH已更新
    )
) else (
    echo 普通用户权限，将修改用户PATH环境变量

    REM 获取当前用户PATH
    for /f "tokens=2*" %%A in ('reg query "HKCU\Environment" /v Path 2^>nul') do set "USER_PATH=%%B"

    REM 检查是否已经包含目标路径
    echo !USER_PATH! | findstr /C:"%BIN_PATH%" >nul
    if %errorLevel% == 0 (
        echo 目标路径已在用户PATH中
    ) else (
        echo 添加 %BIN_PATH% 到用户PATH...
        setx PATH "!USER_PATH!;%BIN_PATH%" >nul
        echo 用户PATH已更新
    )
)

echo.
echo ========================================
echo 配置完成！
echo ========================================
echo.
echo 配置信息:
echo   MSYS2路径: %MSYS2_PATH%
echo   环境类型: %TARGET_ENV%
echo   添加路径: %BIN_PATH%
echo.
echo 重要提示:
echo 1. 请关闭所有终端窗口
echo 2. 重新打开终端以使环境变量生效
echo 3. 验证配置: gcc --version
echo 4. 验证CGO: go env CGO_ENABLED
echo.

endlocal
