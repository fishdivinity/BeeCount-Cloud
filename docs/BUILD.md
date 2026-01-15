# BeeCount Cloud 构建和部署指南

本文档详细说明了BeeCount Cloud项目的构建和部署流程。

## 目录

- [环境要求](#环境要求)
- [本地开发](#本地开发)
- [构建方式](#构建方式)
- [部署方式](#部署方式)
- [Go模块替换策略](#go模块替换策略)
- [常见问题](#常见问题)

## 环境要求

### Windows环境

- Go 1.21+
- MSYS2（用于CGO编译）
- GCC（通过MSYS2提供）
- SQLite开发库（如果使用SQLite数据库）

**安装MSYS2：**
1. 下载并安装MSYS2: https://www.msys2.org/
2. 运行配置脚本：`scripts\setup_windows.bat mingw64` 或 `scripts\setup_windows.bat ucrt64`
3. 如果MSYS2安装在其他位置，提供完整路径：`scripts\setup_windows.bat mingw64 D:\你的\msys2\路径`

### Ubuntu环境

- Go 1.21+
- GCC
- SQLite开发库

**安装依赖：**
```bash
sudo apt-get update
sudo apt-get install -y build-essential gcc sqlite3 libsqlite3-dev
```

### Docker环境

- Docker 20.10+
- Docker Compose 2.0+

## Go 环境配置

### GOPROXY 和 GOPRIVATE 排查

在国内环境下，Go 模块下载可能会遇到网络问题。以下是常见的排查步骤和解决方案。

#### 问题排查步骤

1. **检查当前 GOPROXY 设置**
   ```bash
   go env GOPROXY
   ```

2. **检查当前 GOPRIVATE 设置**
   ```bash
   go env GOPRIVATE
   ```

3. **如果 GOPRIVATE 包含 `golang.org`，需要清空**
   ```bash
   go env -u GOPRIVATE
   ```

   **重要说明**：新安装的 Go 可能会默认设置 `GOPRIVATE=golang.org`，这会导致 `golang.org/x/*` 相关包不使用代理，从而无法下载。清空该设置后，所有包都会使用 GOPROXY。

4. **设置 GOPROXY**
   ```bash
   go env -w GOPROXY=https://goproxy.cn,direct
   ```

#### 常用 GOPROXY 列表

以下是国内常用的 Go 模块代理服务，任选其一即可：

- `https://goproxy.cn,direct` - 七牛云提供
- `https://goproxy.io,direct` - 全球加速
- `https://mirrors.aliyun.com/goproxy/,direct` - 阿里云提供
- `https://go.proxy.io.cn,direct` - 中国本土加速

#### 验证配置

配置完成后，验证是否生效：

```bash
# 测试下载依赖
go mod download

# 查看当前配置
go env GOPROXY
go env GOPRIVATE
```

#### 常见问题

**Q: 设置了 GOPROXY 但仍然无法下载依赖？**

A: 检查 GOPRIVATE 设置，如果包含了 `golang.org`，请清空：
```bash
go env -u GOPRIVATE
```

**Q: 如何临时使用代理？**

A: 可以在命令前设置环境变量：
```bash
GOPROXY=https://goproxy.cn,direct go mod download
```

**Q: 如何恢复默认设置？**

A: 清除所有自定义设置：
```bash
go env -u GOPROXY
go env -u GOPRIVATE
```

## 本地开发

### Windows开发

```bash
# 克隆项目
git clone https://github.com/fishdivinity/BeeCount-Cloud.git
cd BeeCount-Cloud

# 配置MSYS2环境
scripts\setup_windows.bat mingw64

# 下载依赖
go mod download

# 启用CGO
go env -w CGO_ENABLED=1

# 运行服务
go run cmd/server/main.go
```

### Ubuntu开发

```bash
# 克隆项目
git clone https://github.com/fishdivinity/BeeCount-Cloud.git
cd BeeCount-Cloud

# 下载依赖
go mod download

# 启用CGO
go env -w CGO_ENABLED=1

# 运行服务
go run cmd/server/main.go
```

## 构建方式

项目提供三种构建方式：

### 1. 简单构建（不分离动态库）

适用于快速开发和测试，生成单个可执行文件。

```bash
# Windows/Linux/macOS通用
make build
```

**特点：**
- 生成单个可执行文件
- 文件较大（包含所有依赖）
- 部署简单，只需复制一个文件

**缺点：**
- 文件体积大
- 不便于增量更新

### 2. 动态库分离构建（推荐用于生产）

适用于生产环境，将动态库分离出来以减小可执行文件大小。

#### Windows构建

```bash
# 使用Makefile
make build-windows

# 或使用批处理脚本
scripts\build_windows.bat

# 指定版本号
VERSION=v1.0.0 scripts\build_windows.bat
```

**构建产物：**
- `build/beecount-cloud.exe` - 主程序（约10-20MB）
- `lib/windows/amd64/*.dll` - 动态库文件（libsqlite3-0.dll等）
- `build/deploy/` - 部署包目录
- `build/beecount-cloud-v1.0.0-windows-amd64.zip` - 部署压缩包

**部署步骤：**
1. 解压部署包到服务器
2. 双击运行 `start.bat` 或手动运行：
   ```cmd
   set PATH=%~dp0lib;%PATH%
   beecount-cloud.exe
   ```

#### Linux构建

```bash
# 使用Makefile
make build-linux

# 或使用Shell脚本
./scripts/build_linux.sh

# 指定版本号
VERSION=v1.0.0 ./scripts/build_linux.sh
```

**构建产物：**
- `build/beecount-cloud` - 主程序（约10-20MB）
- `lib/linux/amd64/*.so` - 动态库文件（libsqlite3.so.0等）
- `build/deploy/` - 部署包目录
- `build/beecount-cloud-v1.0.0-linux-amd64.tar.gz` - 部署压缩包

**部署步骤：**
1. 解压部署包到服务器
2. 复制动态库到系统目录：
   ```bash
   sudo cp lib/linux/amd64/*.so /usr/local/lib/
   sudo ldconfig
   ```
3. 运行 `./start.sh` 或手动运行：
   ```bash
   export LD_LIBRARY_PATH=/path/to/lib:$LD_LIBRARY_PATH
   ./beecount-cloud
   ```

#### macOS构建

```bash
# 使用Makefile
make build-darwin
```

**特点：**
- 可执行文件较小
- 动态库可单独更新
- 便于维护和升级

### 3. Docker构建

适用于容器化部署。

```bash
# 构建Docker镜像
make docker-build

# 或直接使用docker命令
docker build -t beecount-cloud:latest .
```

**特点：**
- 包含所有运行时依赖
- 跨平台一致性
- 易于扩展和管理

## 部署方式

### 1. 本地部署

使用动态库分离构建的部署包。

#### Windows服务器

```bash
# 1. 上传部署包
scp build/beecount-cloud-v1.0.0-windows-amd64.zip user@server:/path/to/deploy/

# 2. 解压
unzip beecount-cloud-v1.0.0-windows-amd64.zip

# 3. 修改配置
notepad config.yaml

# 4. 启动服务
start.bat

# 5. 或安装为Windows服务（可选）
# 使用NSSM或sc命令
```

#### Linux服务器

```bash
# 1. 上传部署包
scp build/beecount-cloud-v1.0.0-linux-amd64.tar.gz user@server:/path/to/deploy/

# 2. 解压
tar -xzf beecount-cloud-v1.0.0-linux-amd64.tar.gz

# 3. 安装动态库
sudo cp lib/linux/amd64/*.so /usr/local/lib/
sudo ldconfig

# 4. 修改配置
vim config.yaml

# 5. 启动服务
./start.sh

# 6. 或配置为systemd服务（可选）
sudo vim /etc/systemd/system/beecount-cloud.service
```

**systemd服务示例：**
```ini
[Unit]
Description=BeeCount Cloud Service
After=network.target

[Service]
Type=simple
User=beecount
WorkingDirectory=/opt/beecount-cloud
ExecStart=/opt/beecount-cloud/start.sh
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

### 2. Docker部署

#### 使用Docker Compose

```bash
# 启动服务
docker-compose up -d

# 查看日志
docker-compose logs -f

# 停止服务
docker-compose down

# 重新构建
docker-compose up -d --build
```

**特点：**
- 一键部署
- 自动管理依赖
- 易于扩展和迁移

### 3. Kubernetes部署

```bash
# 部署到K8s集群
kubectl apply -f deployments/k8s.yaml

# 查看状态
kubectl get pods -l app=beecount-cloud

# 查看日志
kubectl logs -l app=beecount-cloud -f

# 扩容
kubectl scale deployment beecount-cloud --replicas=5
```

**特点：**
- 自动扩缩容
- 滚动更新
- 高可用性

## 常见问题

### 1. CGO编译错误

**问题：** `cgo: C compiler not found: exec: "gcc": executable file not found`

**解决方案：**

Windows：
```bash
# 安装MSYS2
scripts\setup_windows.bat mingw64
```

Linux：
```bash
sudo apt-get install build-essential gcc
```

### 2. SQLite库未找到

**问题：** `cannot load libsqlite3.so: shared object file not found`

**解决方案：**

Linux：
```bash
# 安装SQLite开发库
sudo apt-get install libsqlite3-dev

# 复制动态库
sudo cp lib/linux/amd64/*.so /usr/local/lib/
sudo ldconfig
```

Windows：
```bash
# 确保lib目录与可执行文件在同一目录
# 或将lib目录添加到PATH
set PATH=%~dp0lib;%PATH%
```

### 3. 端口被占用

**问题：** `bind: address already in use`

**解决方案：**

1. 修改 `config.yaml` 中的端口
2. 或停止占用端口的进程：
   ```bash
   # Linux
   sudo lsof -i :8080
   sudo kill -9 <PID>
   
   # Windows
   netstat -ano | findstr :8080
   taskkill /PID <PID> /F
   ```

### 4. 数据库连接失败

**问题：** `failed to connect to database`

**解决方案：**

1. 检查 `config.yaml` 中的数据库配置
2. 确保数据库目录存在且有写权限
3. 如果使用MySQL/PostgreSQL，确保服务已启动：
   ```bash
   # MySQL
   sudo systemctl status mysql
   
   # PostgreSQL
   sudo systemctl status postgresql
   ```

### 5. 权限错误

**问题：** `permission denied`

**解决方案：**

Linux：
```bash
# 设置正确的权限
chmod +x beecount-cloud
chmod +x start.sh

# 如果需要root权限运行
sudo ./start.sh
```

Windows：
```bash
# 以管理员身份运行
# 右键 -> 以管理员身份运行
```

## 性能优化建议

### 1. 数据库优化

- 使用生产级数据库（MySQL/PostgreSQL）而非SQLite
- 配置适当的连接池大小
- 定期备份数据库

### 2. 日志优化

- 生产环境使用 `file` 输出而非 `stdout`
- 配置日志轮转（max_size, max_backups, max_age）
- 使用合适的日志级别（info/warn/error）

### 3. 存储优化

- 使用对象存储（S3/OSS/COS）而非本地存储
- 配置CDN加速静态资源访问
- 定期清理过期文件

### 4. 缓存优化

- 使用Redis缓存热点数据
- 配置适当的缓存过期时间
- 监控缓存命中率

## 监控和维护

### 健康检查

服务提供健康检查端点：`/health`

```bash
# 检查服务状态
curl http://localhost:8080/health
```

### 日志查看

```bash
# 查看应用日志
tail -f logs/app.log

# 查看Docker日志
docker-compose logs -f beecount-cloud

# 查看K8s日志
kubectl logs -l app=beecount-cloud -f
```

### 性能监控

建议集成以下监控工具：
- Prometheus + Grafana
- ELK Stack（Elasticsearch, Logstash, Kibana）
- Sentry（错误追踪）

## 安全建议

1. **修改默认配置**
   - 更改JWT密钥
   - 修改数据库密码
   - 配置CORS白名单

2. **启用HTTPS**
   - 使用Nginx/Caddy反向代理
   - 配置SSL证书

3. **定期更新**
   - 更新依赖包
   - 修复安全漏洞
   - 升级Go版本

4. **备份策略**
   - 定期备份数据库
   - 备份配置文件
   - 备份上传的文件

## 技术支持

如有问题，请联系：
- Email: fishdivinity@foxmail.com
- GitHub: https://github.com/fishdivinity/BeeCount-Cloud/issues
