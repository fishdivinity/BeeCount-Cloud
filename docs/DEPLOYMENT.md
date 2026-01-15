# BeeCount Cloud 部署文档

## 目录

- [快速开始](#快速开始)
- [本地部署](#本地部署)
- [Docker部署](#docker部署)
- [Kubernetes部署](#kubernetes部署)
- [CI/CD](#cicd)
- [Nginx配置](#nginx配置)
- [环境变量](#环境变量)
- [监控与日志](#监控与日志)
- [备份与恢复](#备份与恢复)
- [故障排查](#故障排查)

## 快速开始

### 前置要求

- Docker 20.10+
- Docker Compose 2.0+
- 或 Kubernetes 1.20+

### 使用Docker Compose快速启动

```bash
git clone https://github.com/fishdivinity/BeeCount-Cloud.git
cd BeeCount-Cloud
docker-compose up -d
```

服务将在 `http://localhost:8080` 启动。

## 本地部署

### Windows环境

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

### Linux环境

```bash
# 克隆项目
git clone https://github.com/fishdivinity/BeeCount-Cloud.git
cd BeeCount-Cloud

# 安装依赖
sudo apt-get update
sudo apt-get install -y build-essential gcc sqlite3 libsqlite3-dev

# 下载依赖
go mod download

# 启用CGO
go env -w CGO_ENABLED=1

# 运行服务
go run cmd/server/main.go
```

## Docker部署

### 构建镜像

```bash
docker build -t beecount-cloud:latest .
```

### 运行容器

```bash
docker run -d \
  --name beecount-cloud \
  -p 8080:8080 \
  -v $(pwd)/data:/app/data \
  -v $(pwd)/logs:/app/logs \
  -v $(pwd)/uploads:/app/uploads \
  -v $(pwd)/config.yaml:/app/config.yaml \
  beecount-cloud:latest
```

### 使用Docker Compose

```bash
docker-compose up -d
```

查看日志：

```bash
docker-compose logs -f
```

停止服务：

```bash
docker-compose down
```

## Kubernetes部署

### 部署到Kubernetes

```bash
kubectl apply -f deployments/k8s.yaml
```

### 查看部署状态

```bash
kubectl get pods -l app=beecount-cloud
kubectl get services beecount-cloud
```

### 扩容

```bash
kubectl scale deployment beecount-cloud --replicas=5
```

### 查看日志

```bash
kubectl logs -f deployment/beecount-cloud
```

## CI/CD

项目使用GitHub Actions进行自动化构建和部署。

### 工作流文件

项目包含两个GitHub Actions工作流：

1. **docker-image.yml**: 仅推送到GitHub Container Registry
2. **docker-hub.yml**: 同时推送到Docker Hub和GitHub Container Registry

### 支持的平台

- **linux/amd64**: 标准x86_64架构
- **linux/arm64**: ARM64架构（树莓派、Apple Silicon等）

### 触发条件

- Push到main、master、develop分支
- 创建v*标签（如v1.0.0）
- Pull Request到main、master、develop分支
- 手动触发

### GitHub Container Registry

默认使用GitHub Container Registry，无需额外配置。

**镜像地址：**
```bash
ghcr.io/fishdivinity/beecount-cloud:latest
ghcr.io/fishdivinity/beecount-cloud:v1.0.0
ghcr.io/fishdivinity/beecount-cloud:main
ghcr.io/fishdivinity/beecount-cloud:develop
```

**使用镜像：**
```bash
# 拉取镜像
docker pull ghcr.io/fishdivinity/beecount-cloud:latest

# 运行容器
docker run -d -p 8080:8080 \
  -v $(pwd)/data:/app/data \
  -v $(pwd)/logs:/app/logs \
  -v $(pwd)/uploads:/app/uploads \
  ghcr.io/fishdivinity/beecount-cloud:latest
```

### Docker Hub

支持同时推送到Docker Hub，需要配置认证信息。

**配置步骤：**

1. 在GitHub仓库中添加Secrets：
   - `DOCKER_USERNAME`: Docker Hub用户名
   - `DOCKER_PASSWORD`: Docker Hub访问令牌（不是密码）

2. 创建Docker Hub访问令牌：
   - 访问 https://hub.docker.com/settings/security
   - 点击"New Access Token"
   - 选择"Read & Write"权限
   - 复制生成的令牌

**镜像地址：**
```bash
docker.io/fishdivinity/beecount-cloud:latest
docker.io/fishdivinity/beecount-cloud:v1.0.0
```

**使用镜像：**
```bash
# 拉取镜像
docker pull fishdivinity/beecount-cloud:latest

# 运行容器
docker run -d -p 8080:8080 \
  -v $(pwd)/data:/app/data \
  -v $(pwd)/logs:/app/logs \
  -v $(pwd)/uploads:/app/uploads \
  fishdivinity/beecount-cloud:latest
```

### 手动触发构建

可以在GitHub Actions页面手动触发构建：

1. 访问仓库的"Actions"标签页
2. 选择"Docker Image CI"或"Docker Image CI (Docker Hub)"工作流
3. 点击"Run workflow"按钮
4. 选择分支（可选）
5. 点击"Run workflow"开始构建

### 查看构建状态

```bash
# 查看GitHub Actions运行状态
# 访问: https://github.com/fishdivinity/BeeCount-Cloud/actions

# 查看构建的镜像
# 访问: https://github.com/fishdivinity/BeeCount-Cloud/packages
```

## Nginx配置

### 基础配置

```nginx
server {
    listen 80;
    server_name beecount.example.com;

    location / {
        proxy_pass http://beecount-cloud:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        proxy_redirect off;
        proxy_buffering off;
    }

    location /uploads {
        proxy_pass http://beecount-cloud:8080;
        proxy_set_header Host $host;
    }
}
```

### SSL配置

```nginx
server {
    listen 443 ssl http2;
    server_name beecount.example.com;

    ssl_certificate /etc/nginx/ssl/beecount.crt;
    ssl_certificate_key /etc/nginx/ssl/beecount.key;

    location / {
        proxy_pass http://beecount-cloud:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        proxy_redirect off;
        proxy_buffering off;
    }
}

server {
    listen 80;
    server_name beecount.example.com;
    return 301 https://$server_name$request_uri;
}
```

## 环境变量

| 变量名 | 说明 | 默认值 |
|---------|------|---------|
| GIN_MODE | 运行模式 (debug/release) | release |
| DB_TYPE | 数据库类型 (sqlite/mysql/postgres) | sqlite |
| DB_PATH | SQLite数据库路径 | ./data/beecount.db |
| DB_HOST | MySQL/PostgreSQL主机 | localhost |
| DB_PORT | MySQL/PostgreSQL端口 | 3306/5432 |
| DB_USER | 数据库用户名 | root |
| DB_PASSWORD | 数据库密码 | - |
| DB_NAME | 数据库名称 | beecount |
| JWT_SECRET | JWT密钥 | - |
| JWT_EXPIRE_HOURS | Token过期时间（小时） | 24 |
| STORAGE_TYPE | 存储类型 (local/s3) | local |
| STORAGE_PATH | 本地存储路径 | ./data/uploads |
| S3_REGION | S3区域 | us-east-1 |
| S3_BUCKET | S3桶名 | - |
| S3_ACCESS_KEY_ID | S3访问密钥 | - |
| S3_SECRET_ACCESS_KEY | S3秘密密钥 | - |
| S3_ENDPOINT | S3端点 | https://s3.amazonaws.com |
| LOG_LEVEL | 日志级别 (debug/info/warn/error) | info |
| LOG_FORMAT | 日志格式 (json/console) | json |
| SERVER_DOCS_ENABLED | 是否启用API文档 | true |

## 监控与日志

### 健康检查

```bash
curl http://localhost:8080/health
```

响应：

```json
{
  "status": "ok"
}
```

### 日志查看

日志文件位于 `/app/logs/app.log`。

查看实时日志：

```bash
tail -f /app/logs/app.log
```

### Prometheus监控（可选）

如需启用Prometheus监控，需在配置中添加：

```yaml
monitoring:
  enabled: true
  port: 9090
```

访问 `http://localhost:9090/metrics` 查看指标。

## 备份与恢复

### 数据库备份

```bash
# SQLite
cp /app/data/beecount.db /backup/beecount_$(date +%Y%m%d).db

# MySQL
mysqldump -u root -p beecount > /backup/beecount_$(date +%Y%m%d).sql

# PostgreSQL
pg_dump -U postgres beecount > /backup/beecount_$(date +%Y%m%d).sql
```

### 数据恢复

```bash
# SQLite
cp /backup/beecount_20240101.db /app/data/beecount.db

# MySQL
mysql -u root -p beecount < /backup/beecount_20240101.sql

# PostgreSQL
psql -U postgres beecount < /backup/beecount_20240101.sql
```

## 故障排查

### 服务无法启动

1. 检查端口占用：
   ```bash
   netstat -tulpn | grep 8080
   ```

2. 检查日志：
   ```bash
   docker logs beecount-cloud
   ```

3. 检查配置文件语法：
   ```bash
   cat config.yaml
   ```

### 数据库连接失败

1. 检查数据库是否运行：
   ```bash
   docker ps | grep mysql
   ```

2. 检查连接信息：
   ```bash
   docker exec -it beecount-cloud sh
   ping db_host
   ```

3. 检查数据库权限：
   ```bash
   mysql -u root -p -e "SHOW GRANTS FOR 'user'@'%';"
   ```

### 存储访问失败

1. 检查存储路径权限：
   ```bash
   ls -la /app/uploads
   ```

2. 检查磁盘空间：
   ```bash
   df -h
   ```

3. 检查S3配置：
   ```bash
   aws s3 ls s3://bucket-name
   ```

## 性能优化

### 数据库优化

1. 创建索引：
   ```sql
   CREATE INDEX idx_user_id ON users(user_id);
   CREATE INDEX idx_ledger_id ON transactions(ledger_id);
   ```

2. 连接池配置：
   ```yaml
   database:
     max_open_conns: 100
     max_idle_conns: 10
     conn_max_lifetime: 3600
   ```

### 缓存配置（可选）

如需启用Redis缓存：

```yaml
cache:
  enabled: true
  type: redis
  redis:
    addr: localhost:6379
    password: ""
    db: 0
```

## 安全建议

1. **修改默认密钥**
   - 修改JWT_SECRET为强密码
   - 修改数据库默认密码

2. **启用HTTPS**
   - 使用Let's Encrypt获取免费SSL证书
   - 强制HTTPS重定向

3. **防火墙配置**
   - 只开放必要端口（80、443）
   - 限制数据库访问IP

4. **定期更新**
   - 及时更新Docker镜像
   - 定期更新依赖包

## 联系支持

如有问题，请联系：
- **Email**: fishdivinity@foxmail.com
- **GitHub Issues**: https://github.com/fishdivinity/BeeCount-Cloud/issues
