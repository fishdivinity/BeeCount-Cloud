# BeeCount Cloud API 对接文档

## 概述

BeeCount Cloud API 为 BeeCount 应用提供云端同步服务，支持多账户、数据同步和图片存储。

## 在线 API 文档

**推荐使用 ReDoc 查看最新的 API 文档：**

- **本地开发**: http://localhost:8080/docs
- **生产环境**: https://[你自己的服务器地址]/docs

ReDoc 提供现代化的 API 文档界面，展示完整的 API 接口说明、请求/响应示例，并自动保持与代码同步。

## 基础信息

- **Base URL**: `https://[你自己的服务器地址]/api/v1`
- **认证方式**: JWT Bearer Token
- **数据格式**: JSON
- **字符编码**: UTF-8

## 认证流程

### 1. 用户注册

```http
POST /auth/register
Content-Type: application/json

{
  "username": "john_doe",
  "email": "john@example.com",
  "password": "password123"
}
```

**响应**:
```json
{
  "id": 1,
  "username": "john_doe",
  "email": "john@example.com",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

### 2. 用户登录

```http
POST /auth/login
Content-Type: application/json

{
  "email": "john@example.com",
  "password": "password123"
}
```

**响应**:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 1,
    "username": "john_doe",
    "email": "john@example.com",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

### 3. 使用Token

所有需要认证的API请求都需要在Header中携带Token：

```http
Authorization: Bearer {token}
```

## API端点

### 账本管理

#### 创建账本
```http
POST /ledgers
Authorization: Bearer {token}
Content-Type: application/json

{
  "name": "My Ledger",
  "currency": "CNY",
  "type": "personal"
}
```

#### 获取账本列表
```http
GET /ledgers
Authorization: Bearer {token}
```

#### 获取单个账本
```http
GET /ledgers/{id}
Authorization: Bearer {token}
```

#### 更新账本
```http
PUT /ledgers/{id}
Authorization: Bearer {token}
Content-Type: application/json

{
  "name": "Updated Ledger",
  "currency": "USD"
}
```

#### 删除账本
```http
DELETE /ledgers/{id}
Authorization: Bearer {token}
```

### 交易管理

#### 创建交易
```http
POST /ledgers/{ledger_id}/transactions
Authorization: Bearer {token}
Content-Type: application/json

{
  "type": "expense",
  "amount": 100.50,
  "category_id": 1,
  "account_id": 1,
  "happened_at": "2024-01-01T12:00:00Z",
  "note": "Lunch"
}
```

#### 获取交易列表
```http
GET /ledgers/{ledger_id}/transactions?limit=20&offset=0
Authorization: Bearer {token}
```

#### 获取单个交易
```http
GET /ledgers/{ledger_id}/transactions/{id}
Authorization: Bearer {token}
```

#### 更新交易
```http
PUT /ledgers/{ledger_id}/transactions/{id}
Authorization: Bearer {token}
Content-Type: application/json

{
  "amount": 150.00,
  "note": "Updated note"
}
```

#### 删除交易
```http
DELETE /ledgers/{ledger_id}/transactions/{id}
Authorization: Bearer {token}
```

### 同步管理

#### 上传账本数据
```http
POST /ledgers/{ledger_id}/sync/upload
Authorization: Bearer {token}
```

**响应**: 返回JSON格式的账本数据（与BeeCount导出格式兼容）

#### 下载账本数据
```http
GET /ledgers/{ledger_id}/sync/download
Authorization: Bearer {token}
```

**响应**: 返回JSON格式的账本数据

#### 获取同步状态
```http
GET /ledgers/{ledger_id}/sync/status
Authorization: Bearer {token}
```

**响应**:
```json
{
  "ledger_id": 1,
  "ledger_name": "My Ledger",
  "currency": "CNY",
  "count": 100,
  "last_updated": "2024-01-01T00:00:00Z",
  "fingerprint": "abc123..."
}
```

### 附件管理

#### 上传附件
```http
POST /ledgers/{ledger_id}/transactions/{transaction_id}/attachments
Authorization: Bearer {token}
Content-Type: multipart/form-data

file: (binary)
```

**响应**:
```json
{
  "id": 1,
  "file_name": "attachments/1/1_photo.jpg",
  "url": "https://api.beecount.com/uploads/attachments/1/1_photo.jpg"
}
```

#### 获取附件
```http
GET /ledgers/{ledger_id}/transactions/{transaction_id}/attachments/{id}
Authorization: Bearer {token}
```

#### 删除附件
```http
DELETE /ledgers/{ledger_id}/transactions/{transaction_id}/attachments/{id}
Authorization: Bearer {token}
```

## 数据格式

### 账本导出格式

```json
{
  "version": 5,
  "exported_at": "2024-01-01T00:00:00Z",
  "ledger_id": 1,
  "ledger_name": "My Ledger",
  "currency": "CNY",
  "count": 100,
  "accounts": [
    {
      "name": "Cash",
      "type": "cash",
      "currency": "CNY",
      "initial_balance": 1000.00
    }
  ],
  "categories": [
    {
      "name": "Food",
      "kind": "expense",
      "level": 1,
      "sort_order": 0,
      "icon": "restaurant",
      "icon_type": "material"
    }
  ],
  "tags": [
    {
      "name": "Business",
      "color": "#FF5722"
    }
  ],
  "items": [
    {
      "type": "expense",
      "amount": 100.50,
      "category_name": "Food",
      "category_kind": "expense",
      "happened_at": "2024-01-01T12:00:00Z",
      "note": "Lunch",
      "account_name": "Cash",
      "tags": "Business",
      "attachments": [
        {
          "file_name": "photo.jpg",
          "original_name": "photo.jpg",
          "file_size": 102400,
          "width": 1920,
          "height": 1080,
          "sort_order": 0
        }
      ]
    }
  ]
}
```

## 错误处理

所有API错误都遵循以下格式：

```json
{
  "error": "error_type",
  "message": "detailed error message",
  "code": 400
}
```

### 常见错误码

- `400` - 请求参数错误
- `401` - 未授权（Token无效或过期）
- `403` - 禁止访问（权限不足）
- `404` - 资源不存在
- `409` - 资源冲突（如重复注册）
- `500` - 服务器内部错误

## 限流策略

- 每个IP每分钟最多100次请求
- 超过限制将返回 `429 Too Many Requests`

## 最佳实践

1. **Token管理**
   - Token有效期为24小时
   - 建议在客户端缓存Token，过期前刷新
   - 不要在URL中传递Token

2. **数据同步**
   - 建议定期同步（如每5分钟）
   - 使用指纹判断是否需要更新
   - 处理网络错误，实现重试机制

3. **图片上传**
   - 单个文件最大10MB
   - 支持格式：JPEG、PNG、GIF、WebP
   - 建议在上传前压缩图片

4. **错误处理**
   - 实现统一的错误处理机制
   - 向用户展示友好的错误信息
   - 记录错误日志用于调试

## 测试环境

- **测试环境**: `https://staging-api.beecount.com/api/v1`
- **API文档**: `https://staging-api.beecount.com/docs`

## 联系方式

如有问题，请联系：
- **Email**: support@beecount.com
- **GitHub Issues**: https://github.com/beecount/beecount-cloud/issues