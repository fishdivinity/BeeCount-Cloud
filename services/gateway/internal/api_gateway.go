package internal

import (
	"time"

	"github.com/fishdivinity/BeeCount-Cloud/common/proto/auth"
	"github.com/fishdivinity/BeeCount-Cloud/common/proto/business"
	"github.com/fishdivinity/BeeCount-Cloud/common/proto/config"
	configpb "github.com/fishdivinity/BeeCount-Cloud/common/proto/config"
	"github.com/fishdivinity/BeeCount-Cloud/common/proto/log"
	"github.com/fishdivinity/BeeCount-Cloud/common/proto/storage"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// GRPCClientConfig gRPC客户端配置
type GRPCClientConfig struct {
	AuthServiceAddr     string
	BusinessServiceAddr string
	StorageServiceAddr  string
	ConfigServiceAddr   string
	LogServiceAddr      string
}

// APIGateway API网关实现
type APIGateway struct {
	// gRPC客户端
	authClient     auth.AuthServiceClient
	businessClient business.BusinessServiceClient
	storageClient  storage.StorageServiceClient
	configClient   config.ConfigServiceClient
	logClient      log.LogServiceClient

	// gRPC连接
	authConn     *grpc.ClientConn
	businessConn *grpc.ClientConn
	storageConn  *grpc.ClientConn
	configConn   *grpc.ClientConn
	logConn      *grpc.ClientConn
}

// NewAPIGateway 创建API网关实例
func NewAPIGateway() *APIGateway {
	return &APIGateway{}
}

// ConfigureGRPCClients 配置gRPC客户端
func (g *APIGateway) ConfigureGRPCClients(grpcConfig GRPCClientConfig) error {
	// 创建gRPC连接选项
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	// 连接认证服务
	authConn, err := grpc.Dial(grpcConfig.AuthServiceAddr, opts...)
	if err != nil {
		return err
	}
	g.authConn = authConn
	g.authClient = auth.NewAuthServiceClient(authConn)

	// 连接业务服务
	businessConn, err := grpc.Dial(grpcConfig.BusinessServiceAddr, opts...)
	if err != nil {
		return err
	}
	g.businessConn = businessConn
	g.businessClient = business.NewBusinessServiceClient(businessConn)

	// 连接存储服务
	storageConn, err := grpc.Dial(grpcConfig.StorageServiceAddr, opts...)
	if err != nil {
		return err
	}
	g.storageConn = storageConn
	g.storageClient = storage.NewStorageServiceClient(storageConn)

	// 连接配置服务
	configConn, err := grpc.Dial(grpcConfig.ConfigServiceAddr, opts...)
	if err != nil {
		return err
	}
	g.configConn = configConn
	g.configClient = configpb.NewConfigServiceClient(configConn)

	// 连接日志服务
	logConn, err := grpc.Dial(grpcConfig.LogServiceAddr, opts...)
	if err != nil {
		return err
	}
	g.logConn = logConn
	g.logClient = log.NewLogServiceClient(logConn)

	return nil
}

// SetupRoutes 设置路由
func (g *APIGateway) SetupRoutes(router *gin.Engine) {
	// CORS中间件
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	// API文档（使用redoc）
	router.Static("/docs", "d:/Work/code/BeeCount-Cloud/web")
	router.GET("/swagger.json", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"swagger": "2.0",
			"info": gin.H{
				"title":   "BeeCount Cloud API",
				"version": "1.0",
			},
			"paths": gin.H{},
		})
	})

	// API v1路由组
	v1 := router.Group("/api/v1")
	{
		// 认证相关路由
		auth := v1.Group("/auth")
		{
			auth.POST("/login", g.handleLogin)
			auth.POST("/register", g.handleRegister)
			auth.POST("/refresh", g.handleRefreshToken)
		}

		// 需要认证的路由
		authRequired := v1.Group("/")
		authRequired.Use(g.authMiddleware)
		{
			// 账本相关路由
			ledgers := authRequired.Group("/ledgers")
			{
				ledgers.GET("", g.handleGetLedgers)
				ledgers.POST("", g.handleCreateLedger)
				ledgers.GET("/:id", g.handleGetLedger)
				ledgers.PUT("/:id", g.handleUpdateLedger)
				ledgers.DELETE("/:id", g.handleDeleteLedger)
			}

			// 交易相关路由
			transactions := authRequired.Group("/transactions")
			{
				transactions.GET("", g.handleGetTransactions)
				transactions.POST("", g.handleCreateTransaction)
				transactions.GET("/:id", g.handleGetTransaction)
				transactions.PUT("/:id", g.handleUpdateTransaction)
				transactions.DELETE("/:id", g.handleDeleteTransaction)
			}

			// 同步相关路由
			sync := authRequired.Group("/sync")
			{
				sync.POST("", g.handleSync)
			}

			// 附件相关路由
			attachments := authRequired.Group("/attachments")
			{
				attachments.POST("", g.handleUploadAttachment)
				attachments.GET("/:id", g.handleDownloadAttachment)
				attachments.DELETE("/:id", g.handleDeleteAttachment)
			}
		}
	}
}

// 认证中间件
func (g *APIGateway) authMiddleware(c *gin.Context) {
	// 获取Authorization头
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(401, gin.H{"error": "Authorization header is required"})
		c.Abort()
		return
	}

	// 验证JWT令牌
	// 这里简化实现，实际应调用authClient.ValidateToken
	c.Next()
}

// 处理登录
func (g *APIGateway) handleLogin(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Login endpoint"})
}

// 处理注册
func (g *APIGateway) handleRegister(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Register endpoint"})
}

// 处理刷新令牌
func (g *APIGateway) handleRefreshToken(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Refresh token endpoint"})
}

// 处理获取账本列表
func (g *APIGateway) handleGetLedgers(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Get ledgers endpoint"})
}

// 处理创建账本
func (g *APIGateway) handleCreateLedger(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Create ledger endpoint"})
}

// 处理获取单个账本
func (g *APIGateway) handleGetLedger(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Get ledger endpoint"})
}

// 处理更新账本
func (g *APIGateway) handleUpdateLedger(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Update ledger endpoint"})
}

// 处理删除账本
func (g *APIGateway) handleDeleteLedger(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Delete ledger endpoint"})
}

// 处理获取交易列表
func (g *APIGateway) handleGetTransactions(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Get transactions endpoint"})
}

// 处理创建交易
func (g *APIGateway) handleCreateTransaction(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Create transaction endpoint"})
}

// 处理获取单个交易
func (g *APIGateway) handleGetTransaction(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Get transaction endpoint"})
}

// 处理更新交易
func (g *APIGateway) handleUpdateTransaction(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Update transaction endpoint"})
}

// 处理删除交易
func (g *APIGateway) handleDeleteTransaction(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Delete transaction endpoint"})
}

// 处理同步数据
func (g *APIGateway) handleSync(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Sync endpoint"})
}

// 处理上传附件
func (g *APIGateway) handleUploadAttachment(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Upload attachment endpoint"})
}

// 处理下载附件
func (g *APIGateway) handleDownloadAttachment(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Download attachment endpoint"})
}

// 处理删除附件
func (g *APIGateway) handleDeleteAttachment(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Delete attachment endpoint"})
}
