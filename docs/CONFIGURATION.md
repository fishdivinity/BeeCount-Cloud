# BeeCount Cloud 配置说明

本文档详细说明 BeeCount Cloud 的所有配置项。

## 配置文件位置

默认配置文件为 `config.yaml`，位于项目根目录。

如果配置文件不存在，程序会自动生成默认配置。

## 配置项说明

### Server 服务器配置

```yaml
server:
  port: 8080                    # 服务器端口
  mode: debug                    # 运行模式：debug, release, test
  read_timeout: 60s              # 读取超时时间
  write_timeout: 60s             # 写入超时时间
  docs:
    enabled: true               # 是否启用API文档
  allow_registration: false        # 是否允许用户注册
  admin_account:
    username: beecount         # 管理员用户名
    password: beecount_admin_2024  # 管理员密码
```

**说明：**
- `mode`: 生产环境建议使用 `release`
- `allow_registration`: 默认关闭，需要管理员手动创建用户
- 管理员账户会在首次运行时自动创建

### Database 数据库配置

```yaml
database:
  type: sqlite                   # 数据库类型：sqlite, mysql, postgres
  active: sqlite                 # 当前激活的数据库类型
  sqlite:
    path: ./data/beecount.db   # SQLite数据库文件路径
  mysql:
    host: localhost             # MySQL主机地址
    port: 3306                 # MySQL端口
    username: root              # MySQL用户名
    password: password          # MySQL密码
    database: beecount          # MySQL数据库名
    charset: utf8mb4           # 字符集
    parse_time: true           # 解析时间
    loc: Local                # 时区
  postgres:
    host: localhost             # PostgreSQL主机地址
    port: 5432                 # PostgreSQL端口
    username: postgres          # PostgreSQL用户名
    password: password          # PostgreSQL密码
    database: beecount          # PostgreSQL数据库名
    sslmode: disable           # SSL模式
    timezone: UTC             # 时区
```

**说明：**
- `type` 和 `active` 用于配置多个数据库，但只激活一个
- SQLite 适合小型部署，MySQL/PostgreSQL 适合生产环境

### JWT 认证配置

```yaml
jwt:
  secret: your-secret-key-change-in-production  # JWT密钥（自动生成）
  expire_hours: 24                          # Token过期时间（小时）
  rotation_interval_days: 7                  # 密钥轮换间隔（天）
  last_rotation_date: ""                    # 上次轮换日期（自动维护）
  previous_secret: ""                        # 之前的密钥（自动维护）
```

**说明：**
- `secret` 首次运行时自动生成，无需手动设置
- 密钥会自动轮换，旧的密钥会保留一段时间以确保平滑过渡
- `rotation_interval_days` 默认为7天

### Storage 存储配置

```yaml
storage:
  type: local                    # 存储类型：local, s3
  active: local                 # 当前激活的存储类型
  max_file_size: 10485760       # 最大文件大小（字节，10MB）
  allowed_file_types:           # 允许的文件类型
    - image/jpeg
    - image/png
    - image/gif
    - image/webp
  local:
    path: ./data/uploads       # 本地存储路径
    url_prefix: /uploads       # URL前缀
  s3:
    region: us-east-1          # S3区域
    bucket: beecount-uploads   # S3存储桶
    access_key_id: your-access-key    # AWS访问密钥ID
    secret_access_key: your-secret-key  # AWS秘密访问密钥
    endpoint: https://s3.amazonaws.com  # S3端点
```

**说明：**
- `max_file_size` 默认为10MB，可根据需要调整
- `allowed_file_types` 支持自定义文件类型
- S3配置支持AWS S3、阿里云OSS、腾讯云COS、MinIO等

### Log 日志配置

```yaml
log:
  level: info                    # 日志级别：debug, info, warn, error
  format: json                  # 日志格式：json, console
  output: stdout                # 输出目标：stdout, file
  file:
    path: ./logs/app.log    # 日志文件路径
    max_size: 100            # 单个日志文件最大大小（MB）
    max_backups: 3           # 保留的备份文件数量
    max_age: 28              # 日志文件保留天数
    compress: true            # 是否压缩旧日志
    max_total_size_gb: 10     # 日志总大小限制（GB）
```

**说明：**
- `max_total_size_gb` 用于限制日志总大小，超过后会自动清理旧日志
- 日志会自动轮转，避免单个文件过大

### CORS 跨域配置

```yaml
cors:
  allowed_origins:             # 允许的源
    - "*"
  allowed_methods:             # 允许的HTTP方法
    - GET
    - POST
    - PUT
    - DELETE
    - OPTIONS
  allowed_headers:             # 允许的请求头
    - "*"
  exposed_headers:             # 暴露的响应头
    - Content-Length
  allow_credentials: true      # 是否允许携带凭证
  max_age: 12h               # 预检请求缓存时间
```

## 环境变量

支持以下环境变量（仅关键配置）：

- `CONFIG_PATH`: 配置文件路径（默认：config.yaml）
- `DATABASE_TYPE`: 数据库类型（覆盖配置文件）
- `STORAGE_TYPE`: 存储类型（覆盖配置文件）
- `ADMIN_PASSWORD`: 管理员密码（覆盖配置文件）

**示例：**

```bash
# 使用MySQL数据库
DATABASE_TYPE=mysql

# 使用S3存储
STORAGE_TYPE=s3

# 设置管理员密码
ADMIN_PASSWORD=my_secure_password
```

## 配置文件自动生成

如果 `config.yaml` 不存在，程序会自动生成默认配置，包括：

- 随机生成的JWT密钥
- 默认的服务器配置
- SQLite数据库配置
- 本地存储配置
- 管理员账户配置

## 安全建议

1. **生产环境必须修改默认密码**
2. **使用强JWT密钥**（自动生成已足够安全）
3. **关闭API文档**（生产环境）
4. **使用MySQL/PostgreSQL**（生产环境）
5. **配置HTTPS**（生产环境）
6. **定期备份数据库**
7. **限制CORS源**（生产环境）

## 配置验证

程序启动时会验证配置的有效性，如果配置错误会输出详细的错误信息。

## 配置热更新

当前版本不支持配置热更新，修改配置后需要重启服务。