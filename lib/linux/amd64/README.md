# Linux平台库文件（64位）

此目录用于存放Linux 64位平台的动态库和静态库文件。

## 文件类型

- `*.so` - 共享对象（Shared Object，动态链接库）
- `*.a` - 静态库（Static Library）

## 构建命令

```bash
# 使用Makefile
make build-linux

# 或使用Shell脚本
./build_linux.sh
```

## 部署说明

部署到Linux服务器时，需要将此目录下的所有.so文件复制到服务器上，并确保：

1. 将.so文件复制到 `/usr/local/lib/` 或 `/usr/lib/` 目录
2. 运行 `ldconfig` 更新动态链接库缓存
3. 或者将.so文件与可执行文件放在同一目录，并设置 `LD_LIBRARY_PATH`

### 示例部署步骤

```bash
# 复制库文件到系统目录
sudo cp lib/linux/amd64/*.so /usr/local/lib/

# 更新动态链接库缓存
sudo ldconfig

# 或者设置环境变量
export LD_LIBRARY_PATH=/path/to/lib/linux/amd64:$LD_LIBRARY_PATH
```
