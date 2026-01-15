# BeeCount Cloud 管理员操作指南

本文档面向 BeeCount Cloud 的管理员，介绍管理员的权限和操作。

## 默认管理员账户

系统首次运行时会自动创建默认管理员账户：

- **用户名**: `beecount`
- **密码**: `beecount_admin_2024`

**重要：** 生产环境必须立即修改默认密码！

## 修改管理员密码

### 方法1：通过配置文件

编辑 `config.yaml`：

```yaml
server:
  admin_account:
    username: beecount
    password: your_new_secure_password
```

重启服务后生效。

### 方法2：通过环境变量

```bash
ADMIN_PASSWORD=your_new_secure_password
```

### 方法3：通过数据库（高级）

```bash
# 连接到数据库
sqlite3 data/beecount.db

# 更新密码（需要先生成bcrypt哈希）
UPDATE users SET password_hash = '$2a$10$...' WHERE username = 'beecount';
```

## 管理员权限

管理员账户具有以下特殊权限：

1. **用户管理**
   - 查看所有用户
   - 禁用/启用用户账户
   - 重置用户密码

2. **系统配置**
   - 修改系统配置
   - 切换数据库/存储
   - 测试连接

3. **数据管理**
   - 数据备份/恢复
   - 数据迁移
   - 日志查看

## 用户注册控制

默认情况下，用户注册功能是关闭的。这是出于安全考虑，因为当前的验证方式风险较高，容易被攻击。

### 开放用户注册

编辑 `config.yaml`：

```yaml
server:
  allow_registration: true
```

重启服务后生效。

**安全建议：**
- 生产环境建议保持注册关闭
- 如需开放注册，建议配置邮箱验证
- 使用强密码策略
- 启用速率限制防止暴力破解

### 手动创建用户

管理员可以通过API手动创建用户：

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "newuser",
    "email": "newuser@example.com",
    "password": "secure_password"
  }'
```

## 数据库和存储测试

### 测试数据库连接

在Docker容器内或开发环境命令行测试数据库连接：

```bash
# 进入Docker容器
docker exec -it beecount-cloud sh

# 测试MySQL连接
./server test-connection --type=mysql --host=localhost --port=3306 --username=root --password=password --database=beecount

# 测试PostgreSQL连接
./server test-connection --type=postgres --host=localhost --port=5432 --username=postgres --password=password --database=beecount
```

### 测试存储连接

```bash
# 进入Docker容器
docker exec -it beecount-cloud sh

# 测试S3连接
./server test-connection --type=s3 --region=us-east-1 --bucket=my-bucket --access-key=xxx --secret-key=xxx
```

**说明：**
- 这些测试命令需要在容器内部执行
- 测试命令会验证配置的正确性
- 不会修改任何数据，只进行连接测试

## 数据库和存储切换

### 通过配置文件切换

修改 `config.yaml` 中的数据库或存储配置，然后重启服务：

```yaml
database:
  type: mysql  # 从sqlite切换到mysql
  active: mysql
  mysql:
    host: localhost
    port: 3306
    username: root
    password: password
    database: beecount

storage:
  type: s3  # 从local切换到s3
  active: s3
  s3:
    region: us-east-1
    bucket: my-bucket
    access_key_id: xxx
    secret_access_key: xxx
```

### 通过环境变量切换

```bash
# 切换到MySQL
DATABASE_TYPE=mysql

# 切换到S3
STORAGE_TYPE=s3
```

**重要：**
- 切换数据库/存储是高风险操作
- 建议先备份数据
- 切换前先测试连接
- 生产环境建议在维护窗口进行

## JWT密钥管理

### 查看当前密钥

JWT密钥存储在 `config.yaml` 中：

```yaml
jwt:
  secret: current_secret_here
  previous_secret: old_secret_here
  last_rotation_date: 2024-01-15T10:30:00Z
```

### 手动轮换密钥

编辑 `config.yaml`，删除 `secret` 和 `previous_secret`，重启服务会自动生成新密钥。

### 调整轮换间隔

```yaml
jwt:
  rotation_interval_days: 7  # 默认7天
```

## 日志管理

### 查看日志

日志文件位于 `./logs/` 目录：

```bash
# 查看最新日志
tail -f logs/app.log

# 查看错误日志
grep "error" logs/app.log
```

### 日志清理

日志会自动轮转和清理，配置如下：

```yaml
log:
  file:
    max_size: 100            # 单个文件最大100MB
    max_backups: 3           # 保留3个备份
    max_age: 28              # 保留28天
    max_total_size_gb: 10     # 总大小限制10GB
```

## 系统监控

### 健康检查

```bash
curl http://localhost:8080/health
```

响应示例：

```json
{
  "status": "ok",
  "time": "2024-01-15T10:30:00Z"
}
```

### 查看系统状态

管理员可以查看详细的系统状态，包括：

- 数据库连接状态
- 存储连接状态
- 活跃用户数
- 磁盘使用情况
- 内存使用情况

## 安全建议

1. **定期修改密码**
   - 建议每90天修改一次管理员密码

2. **启用双因素认证**（未来功能）
   - 计划支持2FA增强安全性

3. **限制访问IP**
   - 使用防火墙限制管理端口的访问

4. **审计日志**
   - 定期检查审计日志，发现异常活动

5. **备份策略**
   - 每日备份数据库
   - 每周备份上传文件
   - 保留至少30天的备份

## 故障排查

### 管理员无法登录

1. 检查密码是否正确
2. 检查JWT密钥是否被重置
3. 检查数据库连接是否正常
4. 查看日志文件获取详细错误信息

### 数据库连接失败

1. 检查数据库服务是否运行
2. 检查网络连接是否正常
3. 检查数据库配置是否正确
4. 使用测试连接功能验证配置

### 存储连接失败

1. 检查S3服务是否可访问
2. 检查访问密钥是否有效
3. 检查存储桶是否存在
4. 检查网络连接是否正常

## 联系支持

如遇到无法解决的问题，请联系：

- **Email**: fishdivinity@foxmail.com
- **GitHub Issues**: https://github.com/fishdivinity/beecount-cloud/issues