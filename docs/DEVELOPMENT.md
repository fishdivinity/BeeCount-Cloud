# BeeCount Cloud 开发文档

## 目录

- [开发环境搭建](#开发环境搭建)
- [项目结构](#项目结构)
- [代码规范](#代码规范)
- [API 文档](#api-文档)
- [测试指南](#测试指南)
- [贡献指南](#贡献指南)

## 开发环境搭建

### 前置要求

- Go 1.21+
- Git
- Docker（可选）
- Make（可选）

### 安装依赖

```bash
go mod download
```

### 运行服务

```bash
go run cmd/server/main.go
```

### 热重载（可选）

安装air：

```bash
go install github.com/cosmtrek/air@latest
```

运行：

```bash
air
```

## 项目结构

```
beecount-cloud/
├── cmd/
│   └── server/          # 应用入口
├── internal/             # 私有应用代码
│   ├── api/            # HTTP处理器
│   ├── auth/           # 认证授权
│   ├── config/         # 配置管理
│   ├── middleware/     # 中间件
│   ├── models/         # 数据模型
│   ├── repository/     # 数据访问层
│   ├── service/        # 业务逻辑层
│   └── storage/        # 存储抽象
├── pkg/                 # 公共库
│   ├── database/       # 数据库抽象
│   ├── logger/         # 日志
│   └── utils/          # 工具函数
├── docs/                # 文档
├── deployments/          # 部署配置
└── tests/               # 测试
```

## 代码规范

### 命名约定

- **包名**: 小写，单词间无分隔符
  ```go
  package user
  package auth
  ```

- **接口名**: 动词或名词，以`er`结尾
  ```go
  type UserService interface {}
  type Repository interface {}
  ```

- **结构体名**: 名词，首字母大写
  ```go
  type User struct {}
  type Config struct {}
  ```

- **常量**: 首字母大写，使用下划线分隔
  ```go
  const MaxFileSize = 10 * 1024 * 1024
  const DefaultTimeout = 30 * time.Second
  ```

- **变量/函数**: 驼峰命名
  ```go
  var userID uint
  func getUserByID(id uint) (*User, error) {}
  ```

### 错误处理

使用`pkg/utils`中定义的错误：

```go
if err != nil {
    return utils.WrapError(err, "failed to get user")
}
```

### 日志记录

```go
log.Infof("User %d logged in", userID)
log.Errorf("Failed to create user: %v", err)
```

### 注释规范

- 公共函数必须有注释
- 复杂逻辑必须有注释
- 注释使用中文

```go
// GetUserByID 根据ID获取用户信息
// 参数:
//   - id: 用户ID
// 返回:
//   - user: 用户信息
//   - err: 错误信息
func (s *userService) GetUserByID(id uint) (*models.User, error) {
    // 实现逻辑
}
```

## API 文档

本项目使用 swag 自动生成 API 文档，所有 API 接口都必须添加 Swagger 注释。生成的文档使用 ReDoc 展示。

### 安装 swag

```bash
go install github.com/swaggo/swag/cmd/swag@latest
```

### 生成文档

```bash
# 使用 Makefile（推荐）
make swagger

# 或直接使用命令
swag init -g cmd/server/main.go -o docs/swagger --parseDependency
```

### Swagger 注释规范

本项目使用 Swagger 注释生成 OpenAPI 规范，然后通过 ReDoc 展示。所有 API 接口都必须添加完整的 Swagger 注释。

#### 主注释（main.go）

在 `cmd/server/main.go` 文件顶部添加 API 总体信息：

```go
// @title           BeeCount Cloud API
// @version         1.0
// @description     BeeCount Cloud API 是一个用于管理个人财务账本的云服务 API
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  AGPL 3.0
// @license.url   https://www.gnu.org/licenses/agpl-3.0.en.html

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
```

#### API 接口注释

每个 API 接口都需要添加完整的 Swagger 注释：

```go
// CreateUser 创建用户
// @Summary 创建用户
// @Description 创建新用户账户
// @Tags 用户
// @Accept json
// @Produce json
// @Param request body CreateUserRequest true "用户信息"
// @Success 201 {object} UserResponse "创建成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Router /users [post]
func (h *UserHandler) CreateUser(c *gin.Context) {
    // 实现逻辑
}
```

#### 注释字段说明

- `@Summary`: 简短描述（必填）
- `@Description`: 详细描述
- `@Tags`: 分组标签，用于 API 分类（必填）
- `@Accept`: 接受的内容类型（如 json、multipart/form-data）
- `@Produce`: 返回的内容类型（如 json、application/octet-stream）
- `@Param`: 参数说明
  - `path`: 路径参数
  - `query`: 查询参数
  - `body`: 请求体
  - `formData`: 表单数据
- `@Success`: 成功响应
- `@Failure`: 失败响应
- `@Router`: 路由路径和方法（必填）

#### 参数类型定义

请求和响应的类型定义会自动从代码中提取，无需手动编写：

```go
type CreateUserRequest struct {
    Username string `json:"username" binding:"required,min=3,max=50" example:"johndoe"`
    Email    string `json:"email" binding:"required,email" example:"john@example.com"`
    Password string `json:"password" binding:"required,min=6" example:"password123"`
}

type UserResponse struct {
    ID          uint   `json:"id" example:"1"`
    Username    string `json:"username" example:"johndoe"`
    Email       string `json:"email" example:"john@example.com"`
    CreatedAt   string `json:"created_at" example:"2024-01-01T00:00:00Z"`
}
```

#### 认证接口

需要认证的接口需要添加 `@Security` 注解：

```go
// @Summary 获取当前用户信息
// @Tags 用户
// @Security Bearer
// @Success 200 {object} UserResponse
// @Router /users/me [get]
func (h *UserHandler) GetMe(c *gin.Context) {
    // 实现逻辑
}
```

### 文件上传接口

文件上传接口需要特殊处理：

```go
// @Summary 上传附件
// @Tags 附件
// @Accept multipart/form-data
// @Param file formData file true "附件文件"
// @Success 201 {object} AttachmentResponse
// @Router /attachments [post]
func (h *AttachmentHandler) UploadAttachment(c *gin.Context) {
    // 实现逻辑
}
```

### 常见问题

**Q: 修改 API 后文档没有更新？**

A: 需要重新运行 `make swagger` 生成文档。

**Q: 类型定义没有出现在文档中？**

A: 确保类型定义在 `internal/api/types.go` 或相应的 handler 文件中，并且使用了正确的 json 标签。

**Q: 如何禁用 API 文档？**

A: 在 `config.yaml` 中设置 `server.docs.enabled: false`，或设置环境变量 `SERVER_DOCS_ENABLED=false`。

### 最佳实践

1. **及时更新文档**：修改 API 接口后立即更新注释并重新生成文档
2. **使用有意义的标签**：`@Tags` 应该清晰地表示 API 的功能分组
3. **提供示例值**：在类型定义中使用 `example` 标签提供示例值
4. **完整的错误响应**：为所有可能的错误情况添加 `@Failure` 注解
5. **保持一致性**：所有 API 接口的注释风格保持一致

## 测试指南

### 单元测试

```bash
go test ./...
```

### 测试覆盖率

```bash
go test -cover ./...
```

### 生成覆盖率报告

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### 集成测试

```bash
go test -tags=integration ./tests/...
```

### 测试示例

```go
func TestUserService_Register(t *testing.T) {
    mockRepo := &MockUserRepository{}
    authService := auth.NewJWTAuthService(&config.JWTConfig{})
    service := NewUserService(mockRepo, authService)

    user := &models.User{
        Username: "test",
        Email:    "test@example.com",
        PasswordHash: "password",
    }

    err := service.Register(user)
    assert.NoError(t, err)
    assert.NotZero(t, user.ID)
}
```

## 贡献指南

### 提交代码

1. Fork仓库
2. 创建特性分支：
   ```bash
   git checkout -b feature/your-feature
   ```
3. 提交更改：
   ```bash
   git add .
   git commit -m "feat: add user registration"
   ```
4. 推送到分支：
   ```bash
   git push origin feature/your-feature
   ```
5. 创建Pull Request

### 提交信息规范

使用Conventional Commits：

```
<type>(<scope>): <subject>

<body>

<footer>
```

类型：
- `feat`: 新功能
- `fix`: Bug修复
- `docs`: 文档更新
- `style`: 代码格式（不影响功能）
- `refactor`: 重构
- `test`: 测试相关
- `chore`: 构建/工具相关

示例：

```
feat(auth): add JWT token refresh

fix(database): resolve connection pool issue

docs(readme): update deployment instructions
```

### 代码审查

- 确保所有测试通过
- 确保代码覆盖率不低于80%
- 确保遵循代码规范
- 确保添加必要的注释

### 发布流程

1. 更新版本号（遵循语义化版本）
2. 更新CHANGELOG.md
3. 创建Git标签：
   ```bash
   git tag -a v1.0.0 -m "Release version 1.0.0"
   git push origin v1.0.0
   ```
4. 构建Docker镜像：
   ```bash
   docker build -t beecount-cloud:v1.0.0 .
   docker push beecount-cloud:v1.0.0
   ```

## 性能优化

### 数据库查询优化

1. 使用索引
2. 避免N+1查询
3. 使用预加载（Preload）

### 缓存策略

1. 热数据缓存
2. 设置合理的过期时间
3. 使用缓存穿透保护

### 并发控制

1. 使用连接池
2. 限制并发数
3. 实现超时机制

## 安全最佳实践

1. **输入验证**
   - 验证所有用户输入
   - 使用参数绑定
   - 防止SQL注入

2. **认证授权**
   - 使用HTTPS
   - Token定期刷新
   - 权限最小化原则

3. **敏感信息**
   - 不记录密码
   - 不在日志中输出敏感信息
   - 使用环境变量存储密钥

## 常见问题

### Q: 如何添加新的数据库支持？

A: 在`pkg/database`中实现新的Dialector，并在配置中添加相应配置。

### Q: 如何添加新的存储后端？

A: 在`internal/storage`中实现Storage接口，并在factory中注册。

### Q: 如何扩展API？

A: 在`internal/api`中添加新的Handler，并在`main.go`中注册路由。

## 联系方式

- **Email**: dev@beecount.com
- **GitHub**: https://github.com/beecount/beecount-cloud