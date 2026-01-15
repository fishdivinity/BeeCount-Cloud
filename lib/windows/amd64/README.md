# Windows平台库文件（64位）

此目录用于存放Windows 64位平台的动态库和静态库文件。

## 文件类型

- `*.dll` - 动态链接库（Dynamic Link Library）
- `*.lib` - 静态库（Static Library）

## 构建命令

```bash
# 使用Makefile
make build-windows

# 或使用批处理脚本
build_windows.bat
```

## 部署说明

部署到Windows服务器时，需要将此目录下的所有DLL文件复制到服务器上，并确保它们与可执行文件在同一目录或系统PATH中。
