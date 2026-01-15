# BeeCount Cloud

BeeCount的云端服务，提供多账户支持、数据同步和图片存储功能。

<div align="center">

![GitHub stars](https://img.shields.io/github/stars/fishdivinity/BeeCount-Cloud?style=social)
![License](https://img.shields.io/badge/license-AGPL--3.0%20%7C%20Commercial-orange.svg)
![Platform](https://img.shields.io/badge/platform-Linux%20%7C%20Windows-lightgrey.svg)
![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)
![Maintained](https://img.shields.io/badge/Maintained%3F-yes-green.svg)

**BeeCount的云端服务，提供多账户支持、数据同步和图片存储功能**

**核心优势：支持多数据库、JWT认证、RESTful API、对象存储和Docker部署**

<br/>

[📚 部署指南](docs/DEPLOYMENT.md) | [🚀 开发说明](docs/BUILD.md) | [💡 API对接](docs/API_INTEGRATION_GUIDE.md) | [🔄 更新日志](docs/CHANGELOG.md)

</div>

## 功能特性

- 多数据库支持 (SQLite3/MySQL/PostgreSQL)
- JWT认证授权（自动密钥轮换）
- RESTful API
- 对象存储支持 (本地/S3/阿里云/腾讯云/MinIO)
- 文件上传断点续传
- ReDoc API文档
- Docker部署
- Kubernetes部署
- GitHub Actions CI/CD
- 多平台Docker镜像构建 (amd64/arm64)
- 自动配置文件生成
- 管理员账户管理
- 日志自动轮转和清理

## 快速开始

### 环境要求

- Go 1.21+ (推荐使用 go1.25.5 windows/amd64)
- CGO (已启用)
- MSYS2 (Windows环境，推荐MINGW64)
- Docker (可选)
- Docker Compose (可选)

### Windows环境配置

本项目使用SQLite数据库，需要CGO支持。请确保已安装MSYS2并配置正确的环境：

#### 1. 安装MSYS2

下载并安装MSYS2: https://www.msys2.org/

#### 2. 配置MSYS2环境

项目提供了自动配置脚本，可以快速切换MSYS2环境：

```bash
# 使用默认路径 C:\msys64
scripts\setup_windows.bat mingw64
scripts\setup_windows.bat ucrt64

# 使用自定义路径（例如安装在D盘）
scripts\setup_windows.bat mingw64 D:\msys64
scripts\setup_windows.bat ucrt64 E:\devtools\msys64
```

脚本会自动检测MSYS2路径并配置环境变量，配置完成后请重启终端。

#### 3. 验证CGO环境

```bash
# 验证CGO是否启用
go env CGO_ENABLED
# 应该输出: 1

# 验证GCC是否可用
gcc --version
```

### Ubuntu环境配置

```bash
# 安装必要的依赖（如果尚未安装）
sudo apt-get update
sudo apt-get install -y build-essential gcc sqlite3 libsqlite3-dev
```

### 使用Docker Compose

```bash
git clone https://github.com/fishdivinity/BeeCount-Cloud.git
cd BeeCount-Cloud
docker-compose up -d
```

服务将在 `http://localhost:8080` 启动。

**注意：** Docker环境默认关闭API文档，如需开启请设置环境变量 `SERVER_DOCS_ENABLED=true`

### 本地开发

详细的本地开发指南请参考：
- [构建文档](docs/BUILD.md) - 环境配置和构建指南
- [开发文档](docs/DEVELOPMENT.md) - 开发指南和项目结构

快速启动：

```bash
# 克隆项目
git clone https://github.com/fishdivinity/BeeCount-Cloud.git
cd BeeCount-Cloud

# 下载依赖
go mod download

# 运行服务（会自动创建必要的目录）
go run cmd/server/main.go
```

**首次运行说明：**
- 程序会自动创建 `./data/` 目录用于存储SQLite数据库
- 程序会自动创建 `./data/uploads/` 目录用于存储上传的文件
- 程序会自动生成 `config.yaml` 配置文件（如果不存在）
- 程序会自动创建默认管理员账户（用户名：beecount，密码：beecount_admin_2024）

**默认管理员账户：**
- 用户名：`beecount`
- 密码：`beecount_admin_2024`
- **重要：** 生产环境必须立即修改默认密码！

访问 http://localhost:8080/docs 查看API文档（使用 ReDoc）

## 配置

配置文件位于 `config.yaml`，支持环境变量覆盖。

主要配置项：

- **server**: 服务器配置（端口、模式等）
- **database**: 数据库配置（支持SQLite/MySQL/PostgreSQL）
- **jwt**: JWT认证配置
- **storage**: 存储配置（支持本地/S3）
- **log**: 日志配置
- **cors**: CORS配置
- **docs**: API文档配置（默认本地开启，Docker环境默认关闭）

**环境变量覆盖示例：**

```bash
# Docker环境关闭API文档
SERVER_DOCS_ENABLED=false

# 修改服务器端口
SERVER_PORT=9000

# 修改日志级别
LOG_LEVEL=debug
```

## API文档

启动服务后访问 `/docs` 查看完整的API文档（使用 ReDoc）。

详细的API对接文档请参考：[API对接文档](docs/API_INTEGRATION_GUIDE.md)

## 文档

- [部署文档](docs/DEPLOYMENT.md) - Docker、Kubernetes部署和CI/CD配置
- [构建文档](docs/BUILD.md) - 本地构建和部署指南
- [开发文档](docs/DEVELOPMENT.md) - 开发指南和项目结构
- [API对接文档](docs/API_INTEGRATION_GUIDE.md) - API对接详细文档
- [更新日志](docs/CHANGELOG.md) - 版本更新记录

## 测试

```bash
# 运行所有测试
make test

# 生成覆盖率报告
make test-coverage
```

## 贡献

欢迎贡献代码！请参考[开发文档](docs/DEVELOPMENT.md)了解贡献指南。

## 许可证

本项目采用双重许可证：

### GNU Affero General Public License v3.0 (AGPL-3.0)

适用于大多数用户，允许自由使用、修改和分发，但要求：
- 如果分发修改后的版本，必须开放源代码
- 如果通过网络提供服务(SaaS)，必须向用户提供源代码
- 保留所有版权声明和许可证信息

完整的许可证文本请查看 [LICENSE.AGPL](LICENSE.AGPL) 文件。

### 商业许可证

如果您需要：
- 不希望公开您的修改
- 需要私有部署而不开源
- 需要集成到专有软件中
- 需要正式的技术支持和保证

请购买商业许可证。详细信息请查看 [LICENSE.md](LICENSE.md) 或 [COMMERCIAL_LICENSE.md](COMMERCIAL_LICENSE.md)。

## 联系方式

- **Email**: fishdivinity@foxmail.com
- **GitHub**: https://github.com/fishdivinity/beecount-cloud