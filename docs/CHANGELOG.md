# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- 初始版本发布
- 多数据库支持 (SQLite3/MySQL/PostgreSQL)
- JWT认证授权系统
- RESTful API接口
- 对象存储抽象层 (本地/S3)
- 图片上传和附件管理
- Docker和Kubernetes部署配置
- ReDoc API文档
- 完整的测试覆盖

### Changed
- 使用 ReDoc 替代 Swagger UI 作为 API 文档展示界面
- 移除 `github.com/swaggo/gin-swagger` 依赖，减小容器体积
- 配置项从 `server.swagger.enabled` 改为 `server.docs.enabled`
- 环境变量从 `SERVER_SWAGGER_ENABLED` 改为 `SERVER_DOCS_ENABLED`
- API 文档访问路径从 `/swagger/index.html` 改为 `/docs`
- 保留 `swag init` 用于生成 OpenAPI JSON 文件

### Features
- 用户注册和登录
- 账本管理 (CRUD)
- 交易管理 (CRUD)
- 数据同步 (上传/下载/状态查询)
- 附件管理 (上传/下载/删除)
- 多账户支持
- 数据库自动迁移

### Security
- 密码加密存储 (bcrypt)
- JWT Token认证
- CORS支持
- 输入验证
- SQL注入防护

### Documentation
- API对接文档
- 部署文档
- 开发文档
- Swagger在线文档

### Testing
- 单元测试
- 集成测试
- 测试覆盖率报告

## [1.0.0] - 2024-01-13

### Added
- 初始版本发布
- 完整的BeeCount云端服务实现