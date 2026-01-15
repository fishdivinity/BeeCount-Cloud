package main

// @title           BeeCount Cloud API
// @version         1.0
// @description     BeeCount Cloud API 是一个用于管理个人财务账本的云服务 API，提供账本、交易、分类、标签等功能。
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  AGPL 3.0
// @license.url   https://www.gnu.org/licenses/agpl-3.0.en.html

// @host      localhost:8080
// @BasePath  /api/v1

// @Tags 认证 账本 交易 同步 附件

// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fishdivinity/BeeCount-Cloud/internal/admin"
	"github.com/fishdivinity/BeeCount-Cloud/internal/api"
	"github.com/fishdivinity/BeeCount-Cloud/internal/auth"
	"github.com/fishdivinity/BeeCount-Cloud/internal/config"
	jwtmanager "github.com/fishdivinity/BeeCount-Cloud/internal/jwt"
	"github.com/fishdivinity/BeeCount-Cloud/internal/middleware"
	"github.com/fishdivinity/BeeCount-Cloud/internal/repository"
	"github.com/fishdivinity/BeeCount-Cloud/internal/service"
	"github.com/fishdivinity/BeeCount-Cloud/internal/storage"
	_ "github.com/fishdivinity/BeeCount-Cloud/docs/swagger"
	"github.com/fishdivinity/BeeCount-Cloud/pkg/database"
	"github.com/fishdivinity/BeeCount-Cloud/pkg/logger"
	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置文件
	cfg, err := config.Load("config.yaml")
	if err != nil {
		panic(fmt.Sprintf("Failed to load config: %v", err))
	}

	// 初始化日志
	log, err := logger.NewLogger(
		cfg.Log.Level,
		cfg.Log.Format,
		cfg.Log.Output,
		logger.FileConfig{
			Path:          cfg.Log.File.Path,
			MaxSize:       cfg.Log.File.MaxSize,
			MaxBackups:    cfg.Log.File.MaxBackups,
			MaxAge:        cfg.Log.File.MaxAge,
			Compress:      cfg.Log.File.Compress,
			MaxTotalSizeGB: cfg.Log.File.MaxTotalSizeGB,
		},
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}

	// 启动日志清理例程
	logger.StartCleanupRoutine(logger.FileConfig{
		Path:          cfg.Log.File.Path,
		MaxTotalSizeGB: cfg.Log.File.MaxTotalSizeGB,
	}, 24)

	log.Info("Starting BeeCount Cloud server...")

	// 初始化JWT密钥管理器并检查轮换
	secretManager := jwtmanager.NewSecretManager("config.yaml", &cfg.JWT)
	if err := secretManager.CheckRotation(); err != nil {
		log.Warn("Failed to check JWT secret rotation", logger.Error(err))
	}

	// 使用最新的密钥初始化认证服务
	currentSecret, previousSecret := secretManager.GetSecrets()
	authService := auth.NewJWTAuthServiceWithSecrets(currentSecret, previousSecret, cfg.JWT.ExpireHours)

	// 初始化数据库
	db, err := database.NewDatabase(&cfg.Database)
	if err != nil {
		log.Fatal("Failed to initialize database", logger.Error(err))
	}
	defer db.Close()

	// 自动迁移数据库表结构
	if err := database.AutoMigrateModels(db); err != nil {
		log.Fatal("Failed to migrate database", logger.Error(err))
	}

	log.Info("Database migrated successfully")

	// 确保管理员账户存在
	adminManager := admin.NewManager(db.GetDB(), authService, cfg.Server.AdminAccount)
	if err := adminManager.EnsureAdmin(); err != nil {
		log.Warn("Failed to ensure admin account", logger.Error(err))
	}

	// 初始化存储
	storageBackend, err := storage.NewStorage(&cfg.Storage)
	if err != nil {
		log.Fatal("Failed to initialize storage", logger.Error(err))
	}

	// 初始化仓储
	repos := repository.NewRepositories(db.GetDB())

	// 初始化服务层
	services := service.NewServices(repos, authService)

	// 设置Gin模式
	gin.SetMode(cfg.Server.Mode)

	// 创建Gin路由
	router := gin.New()

	// 添加中间件
	router.Use(middleware.Recovery(log))
	router.Use(middleware.ErrorHandler(log))
	router.Use(middleware.CORS())

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	// API路由组
	v1 := router.Group("/api/v1")
	{
		// 认证路由
		authHandler := api.NewAuthHandler(services.User, authService, cfg.Server.AllowRegistration)
		v1.POST("/auth/register", authHandler.Register)
		v1.POST("/auth/login", authHandler.Login)
		v1.GET("/users/me", auth.AuthMiddleware(authService), authHandler.Me)

		// 账本路由
		ledgerHandler := api.NewLedgerHandler(services.Ledger)
		ledgerGroup := v1.Group("/ledgers")
		ledgerGroup.Use(auth.AuthMiddleware(authService))
		{
			ledgerGroup.POST("", ledgerHandler.CreateLedger)
			ledgerGroup.GET("", ledgerHandler.GetLedgers)
			ledgerGroup.GET("/:ledger_id", ledgerHandler.GetLedger)
			ledgerGroup.PUT("/:ledger_id", ledgerHandler.UpdateLedger)
			ledgerGroup.DELETE("/:ledger_id", ledgerHandler.DeleteLedger)
		}

		// 交易路由
		txHandler := api.NewTransactionHandler(services.Transaction)
		txGroup := v1.Group("/ledgers/:ledger_id/transactions")
		txGroup.Use(auth.AuthMiddleware(authService))
		{
			txGroup.POST("", txHandler.CreateTransaction)
			txGroup.GET("", txHandler.GetTransactions)
			txGroup.GET("/:id", txHandler.GetTransaction)
			txGroup.PUT("/:id", txHandler.UpdateTransaction)
			txGroup.DELETE("/:id", txHandler.DeleteTransaction)
		}

		// 同步路由
		syncHandler := api.NewSyncHandler(services.Sync)
		syncGroup := v1.Group("/ledgers/:ledger_id/sync")
		syncGroup.Use(auth.AuthMiddleware(authService))
		{
			syncGroup.POST("/upload", syncHandler.UploadLedger)
			syncGroup.GET("/download", syncHandler.DownloadLedger)
			syncGroup.GET("/status", syncHandler.GetSyncStatus)
		}

		// 附件路由
		attachmentHandler := api.NewAttachmentHandler(repos.Attachment, storageBackend, cfg.Storage.MaxFileSize, cfg.Storage.AllowedFileTypes)
		attachmentUploadGroup := v1.Group("/ledgers/:ledger_id/transactions/:transaction_id/attachments")
		attachmentUploadGroup.Use(auth.AuthMiddleware(authService))
		{
			attachmentUploadGroup.POST("", attachmentHandler.UploadAttachment)
		}
		attachmentGroup := v1.Group("/attachments")
		attachmentGroup.Use(auth.AuthMiddleware(authService))
		{
			attachmentGroup.GET("/:id", attachmentHandler.GetAttachment)
			attachmentGroup.DELETE("/:id", attachmentHandler.DeleteAttachment)
		}
	}

	log.Info(fmt.Sprintf("API Docs enabled: %v", cfg.Server.Docs.Enabled))

	if cfg.Server.Docs.Enabled {
		// 使用静态文件服务提供swagger目录（包含生成的swagger.json）
		router.Static("/swagger", "./docs/swagger")
		// 使用静态文件服务提供ReDoc页面
		router.Static("/docs", "./web/redoc")
		log.Info("API documentation enabled at /docs (ReDoc) and /swagger/swagger.json")
	} else {
		log.Info("API documentation disabled")
	}

	// 创建HTTP服务器
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// 启动服务器
	go func() {
		log.Info(fmt.Sprintf("Server is running on port %d", cfg.Server.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start server", logger.Error(err))
		}
	}()

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown", logger.Error(err))
	}

	log.Info("Server exited")
}
