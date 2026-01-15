# lib目录说明

此目录用于存放编译生成的动态库和静态库文件。

## 目录结构

```
lib/
├── windows/          # Windows平台库文件
│   ├── amd64/       # 64位Windows
│   │   ├── *.dll   # 动态链接库
│   │   └── *.lib  # 静态库
│   └── 386/        # 32位Windows（可选）
│       ├── *.dll
│       └── *.lib
├── linux/            # Linux平台库文件
│   ├── amd64/       # 64位Linux
│   │   ├── *.so    # 动态链接库
│   │   └── *.a    # 静态库
│   └── 386/        # 32位Linux（可选）
│       ├── *.so
│       └── *.a
└── darwin/           # macOS平台库文件（可选）
    ├── amd64/       # 64位macOS
    │   ├── *.dylib # 动态链接库
    │   └── *.a    # 静态库
    └── arm64/       # ARM64 macOS
        ├── *.dylib
        └── *.a
```

## 构建说明

使用项目提供的构建脚本会自动将库文件输出到对应目录：

- Windows: `make build-windows` 或 `build_windows.bat`
- Linux: `make build-linux` 或 `./build_linux.sh`
- macOS: `make build-darwin` 或 `./build_darwin.sh`

## 注意事项

1. 此目录下的文件由构建脚本自动生成
2. 不要手动修改或删除这些文件
3. 提交代码时，此目录会被.gitignore忽略
4. 部署时需要将对应的库文件一同部署
