package sync

import (
	"fmt"
	"os"
	"time"

	"github.com/fishdivinity/BeeCount-Cloud/services/config/internal/generator"
	"github.com/fishdivinity/BeeCount-Cloud/services/config/internal/model"
)

// SyncConfig 同步配置
// 根据配置来源执行相应的同步操作
func SyncConfig(cfg *model.Config, source model.ConfigSource, configPath string) error {
	// 先执行内部同步，更新内部状态
	if err := InternalSyncHandler(cfg, source); err != nil {
		return fmt.Errorf("internal sync failed: %w", err)
	}

	// 再执行外部同步，与外部系统（环境变量、配置文件）同步
	if err := ExternalSyncHandler(cfg, source, configPath); err != nil {
		return fmt.Errorf("external sync failed: %w", err)
	}

	return nil
}

// generateConfigContent 生成配置文件内容
// 复用generator包的功能
func generateConfigContent(cfg *model.Config) string {
	// 创建一个临时文件路径
	tempPath := "./temp_config.yaml"

	// 使用generator包生成配置文件
	if err := generator.GenerateDefaultConfig(tempPath); err != nil {
		return "" // 返回空字符串表示生成失败
	}

	// 读取临时文件内容
	content, err := os.ReadFile(tempPath)
	if err != nil {
		return ""
	}

	// 删除临时文件
	os.Remove(tempPath)

	return string(content)
}

// CheckConfigIntegrity 检查配置完整性
// 如果配置项缺失，使用默认值补充
func CheckConfigIntegrity(cfg *model.Config) *model.Config {
	// 创建默认配置
	defaultCfg := &model.Config{
		Server: model.ServerConfig{
			Port:         8080,
			Mode:         "release",
			ReadTimeout:  60 * time.Second,
			WriteTimeout: 60 * time.Second,
			Docs: model.DocsConfig{
				Enabled: false,
			},
			AdminAccount: model.AdminConfig{
				Username: "beecount",
				Password: "beecount_admin_2026",
			},
		},
		Database: model.DatabaseConfig{
			Active: "sqlite",
			SQLite: model.SQLiteConfig{
				Path: "./data/beecount.db",
			},
			MySQL: model.MySQLConfig{
				Host:      "localhost",
				Port:      3306,
				Username:  "root",
				Password:  "password",
				Database:  "beecount",
				Charset:   "utf8mb4",
				ParseTime: true,
				Loc:       "Local",
			},
			Postgres: model.PostgresConfig{
				Host:     "localhost",
				Port:     5432,
				Username: "postgres",
				Password: "password",
				Database: "beecount",
				SSLMode:  "disable",
				Timezone: "UTC",
			},
			Pool: model.PoolConfig{
				MaxIdleConns:    10,
				MaxOpenConns:    100,
				ConnMaxLifetime: 1 * time.Hour,
				ConnMaxIdleTime: 30 * time.Minute,
			},
		},
		Storage: model.StorageConfig{
			Active:           "local",
			MaxFileSize:      5242880,
			AllowedFileTypes: []string{"image/jpeg", "image/png", "image/gif", "image/webp"},
			Local: model.LocalConfig{
				Path:      "./data/uploads",
				URLPrefix: "/uploads",
			},
			S3: model.S3Config{
				Region:          "us-east-1",
				Bucket:          "beecount-uploads",
				AccessKeyID:     "your-access-key",
				SecretAccessKey: "your-secret-key",
				Endpoint:        "https://s3.amazonaws.com",
			},
		},
		JWT: model.JWTConfig{
			Secret:               "", // 保留现有密钥
			ExpireHours:          24,
			RotationIntervalDays: 7,
			LastRotationDate:     "",
			PreviousSecret:       "",
		},
		Log: model.LogConfig{
			Level:  "warn",
			Format: "console",
			Output: "file",
			File: model.FileConfig{
				Path:           "./logs/app.log",
				MaxSize:        10,
				MaxBackups:     100,
				MaxAge:         28,
				Compress:       true,
				MaxTotalSizeGB: 1,
			},
		},
		CORS: model.CORSConfig{
			AllowedOrigins:   []string{"*"},
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"*"},
			ExposedHeaders:   []string{"Content-Length"},
			AllowCredentials: true,
			MaxAge:           "12h",
		},
		Cache: model.CacheConfig{
			Active: "memory",
			Memory: model.MemoryCacheConfig{
				MaxSize: 10000,
			},
			Redis: model.RedisCacheConfig{
				Host:     "localhost",
				Port:     6379,
				Password: "",
				DB:       0,
			},
			Options: make(map[string]interface{}),
		},
	}

	// 如果现有配置的JWT密钥为空，使用默认配置的密钥
	if cfg.JWT.Secret == "" {
		// 生成随机密钥
		secret, err := generator.GenerateRandomSecret()
		if err != nil {
			// 生成失败，使用默认值
			secret = "default_secret"
		}
		cfg.JWT.Secret = secret
	}

	// 合并配置，使用现有配置覆盖默认配置，缺失的配置项使用默认配置
	mergeConfigs(cfg, defaultCfg)

	return cfg
}

// mergeConfigs 合并配置
// 将默认配置中的值合并到当前配置中，当前配置中的值优先
func mergeConfigs(cfg, defaultCfg *model.Config) {
	// 合并 Server 配置
	if cfg.Server.Port == 0 {
		cfg.Server.Port = defaultCfg.Server.Port
	}
	if cfg.Server.Mode == "" {
		cfg.Server.Mode = defaultCfg.Server.Mode
	}
	if cfg.Server.ReadTimeout == 0 {
		cfg.Server.ReadTimeout = defaultCfg.Server.ReadTimeout
	}
	if cfg.Server.WriteTimeout == 0 {
		cfg.Server.WriteTimeout = defaultCfg.Server.WriteTimeout
	}
	if !cfg.Server.Docs.Enabled {
		cfg.Server.Docs.Enabled = defaultCfg.Server.Docs.Enabled
	}
	if cfg.Server.AdminAccount.Username == "" {
		cfg.Server.AdminAccount.Username = defaultCfg.Server.AdminAccount.Username
	}
	if cfg.Server.AdminAccount.Password == "" {
		cfg.Server.AdminAccount.Password = defaultCfg.Server.AdminAccount.Password
	}

	// 合并 Database 配置
	if cfg.Database.Active == "" {
		cfg.Database.Active = defaultCfg.Database.Active
	}
	if cfg.Database.SQLite.Path == "" {
		cfg.Database.SQLite.Path = defaultCfg.Database.SQLite.Path
	}
	if cfg.Database.MySQL.Host == "" {
		cfg.Database.MySQL.Host = defaultCfg.Database.MySQL.Host
	}
	if cfg.Database.MySQL.Port == 0 {
		cfg.Database.MySQL.Port = defaultCfg.Database.MySQL.Port
	}
	if cfg.Database.MySQL.Username == "" {
		cfg.Database.MySQL.Username = defaultCfg.Database.MySQL.Username
	}
	if cfg.Database.MySQL.Password == "" {
		cfg.Database.MySQL.Password = defaultCfg.Database.MySQL.Password
	}
	if cfg.Database.MySQL.Database == "" {
		cfg.Database.MySQL.Database = defaultCfg.Database.MySQL.Database
	}
	if cfg.Database.MySQL.Charset == "" {
		cfg.Database.MySQL.Charset = defaultCfg.Database.MySQL.Charset
	}
	if !cfg.Database.MySQL.ParseTime {
		cfg.Database.MySQL.ParseTime = defaultCfg.Database.MySQL.ParseTime
	}
	if cfg.Database.MySQL.Loc == "" {
		cfg.Database.MySQL.Loc = defaultCfg.Database.MySQL.Loc
	}
	if cfg.Database.Postgres.Host == "" {
		cfg.Database.Postgres.Host = defaultCfg.Database.Postgres.Host
	}
	if cfg.Database.Postgres.Port == 0 {
		cfg.Database.Postgres.Port = defaultCfg.Database.Postgres.Port
	}
	if cfg.Database.Postgres.Username == "" {
		cfg.Database.Postgres.Username = defaultCfg.Database.Postgres.Username
	}
	if cfg.Database.Postgres.Password == "" {
		cfg.Database.Postgres.Password = defaultCfg.Database.Postgres.Password
	}
	if cfg.Database.Postgres.Database == "" {
		cfg.Database.Postgres.Database = defaultCfg.Database.Postgres.Database
	}
	if cfg.Database.Postgres.SSLMode == "" {
		cfg.Database.Postgres.SSLMode = defaultCfg.Database.Postgres.SSLMode
	}
	if cfg.Database.Postgres.Timezone == "" {
		cfg.Database.Postgres.Timezone = defaultCfg.Database.Postgres.Timezone
	}
	if cfg.Database.Pool.MaxIdleConns == 0 {
		cfg.Database.Pool.MaxIdleConns = defaultCfg.Database.Pool.MaxIdleConns
	}
	if cfg.Database.Pool.MaxOpenConns == 0 {
		cfg.Database.Pool.MaxOpenConns = defaultCfg.Database.Pool.MaxOpenConns
	}
	if cfg.Database.Pool.ConnMaxLifetime == 0 {
		cfg.Database.Pool.ConnMaxLifetime = defaultCfg.Database.Pool.ConnMaxLifetime
	}
	if cfg.Database.Pool.ConnMaxIdleTime == 0 {
		cfg.Database.Pool.ConnMaxIdleTime = defaultCfg.Database.Pool.ConnMaxIdleTime
	}

	// 合并 Storage 配置
	if cfg.Storage.Active == "" {
		cfg.Storage.Active = defaultCfg.Storage.Active
	}
	if cfg.Storage.MaxFileSize == 0 {
		cfg.Storage.MaxFileSize = defaultCfg.Storage.MaxFileSize
	}
	if len(cfg.Storage.AllowedFileTypes) == 0 {
		cfg.Storage.AllowedFileTypes = defaultCfg.Storage.AllowedFileTypes
	}
	if cfg.Storage.Local.Path == "" {
		cfg.Storage.Local.Path = defaultCfg.Storage.Local.Path
	}
	if cfg.Storage.Local.URLPrefix == "" {
		cfg.Storage.Local.URLPrefix = defaultCfg.Storage.Local.URLPrefix
	}
	if cfg.Storage.S3.Region == "" {
		cfg.Storage.S3.Region = defaultCfg.Storage.S3.Region
	}
	if cfg.Storage.S3.Bucket == "" {
		cfg.Storage.S3.Bucket = defaultCfg.Storage.S3.Bucket
	}
	if cfg.Storage.S3.AccessKeyID == "" {
		cfg.Storage.S3.AccessKeyID = defaultCfg.Storage.S3.AccessKeyID
	}
	if cfg.Storage.S3.SecretAccessKey == "" {
		cfg.Storage.S3.SecretAccessKey = defaultCfg.Storage.S3.SecretAccessKey
	}
	if cfg.Storage.S3.Endpoint == "" {
		cfg.Storage.S3.Endpoint = defaultCfg.Storage.S3.Endpoint
	}

	// 合并 JWT 配置
	if cfg.JWT.Secret == "" {
		cfg.JWT.Secret = defaultCfg.JWT.Secret
	}
	if cfg.JWT.ExpireHours == 0 {
		cfg.JWT.ExpireHours = defaultCfg.JWT.ExpireHours
	}
	if cfg.JWT.RotationIntervalDays == 0 {
		cfg.JWT.RotationIntervalDays = defaultCfg.JWT.RotationIntervalDays
	}
	if cfg.JWT.LastRotationDate == "" {
		cfg.JWT.LastRotationDate = defaultCfg.JWT.LastRotationDate
	}
	if cfg.JWT.PreviousSecret == "" {
		cfg.JWT.PreviousSecret = defaultCfg.JWT.PreviousSecret
	}

	// 合并 Log 配置
	if cfg.Log.Level == "" {
		cfg.Log.Level = defaultCfg.Log.Level
	}
	if cfg.Log.Format == "" {
		cfg.Log.Format = defaultCfg.Log.Format
	}
	if cfg.Log.Output == "" {
		cfg.Log.Output = defaultCfg.Log.Output
	}
	if cfg.Log.File.Path == "" {
		cfg.Log.File.Path = defaultCfg.Log.File.Path
	}
	if cfg.Log.File.MaxSize == 0 {
		cfg.Log.File.MaxSize = defaultCfg.Log.File.MaxSize
	}
	if cfg.Log.File.MaxBackups == 0 {
		cfg.Log.File.MaxBackups = defaultCfg.Log.File.MaxBackups
	}
	if cfg.Log.File.MaxAge == 0 {
		cfg.Log.File.MaxAge = defaultCfg.Log.File.MaxAge
	}
	if !cfg.Log.File.Compress {
		cfg.Log.File.Compress = defaultCfg.Log.File.Compress
	}
	if cfg.Log.File.MaxTotalSizeGB == 0 {
		cfg.Log.File.MaxTotalSizeGB = defaultCfg.Log.File.MaxTotalSizeGB
	}

	// 合并 CORS 配置
	if len(cfg.CORS.AllowedOrigins) == 0 {
		cfg.CORS.AllowedOrigins = defaultCfg.CORS.AllowedOrigins
	}
	if len(cfg.CORS.AllowedMethods) == 0 {
		cfg.CORS.AllowedMethods = defaultCfg.CORS.AllowedMethods
	}
	if len(cfg.CORS.AllowedHeaders) == 0 {
		cfg.CORS.AllowedHeaders = defaultCfg.CORS.AllowedHeaders
	}
	if len(cfg.CORS.ExposedHeaders) == 0 {
		cfg.CORS.ExposedHeaders = defaultCfg.CORS.ExposedHeaders
	}
	if !cfg.CORS.AllowCredentials {
		cfg.CORS.AllowCredentials = defaultCfg.CORS.AllowCredentials
	}
	if cfg.CORS.MaxAge == "" {
		cfg.CORS.MaxAge = defaultCfg.CORS.MaxAge
	}

	// 合并 Cache 配置
	if cfg.Cache.Active == "" {
		cfg.Cache.Active = defaultCfg.Cache.Active
	}
	if cfg.Cache.Memory.MaxSize == 0 {
		cfg.Cache.Memory.MaxSize = defaultCfg.Cache.Memory.MaxSize
	}
	if cfg.Cache.Redis.Host == "" {
		cfg.Cache.Redis.Host = defaultCfg.Cache.Redis.Host
	}
	if cfg.Cache.Redis.Port == 0 {
		cfg.Cache.Redis.Port = defaultCfg.Cache.Redis.Port
	}
	if cfg.Cache.Redis.Password == "" {
		cfg.Cache.Redis.Password = defaultCfg.Cache.Redis.Password
	}
	if cfg.Cache.Redis.DB == 0 {
		cfg.Cache.Redis.DB = defaultCfg.Cache.Redis.DB
	}
	if len(cfg.Cache.Options) == 0 {
		cfg.Cache.Options = defaultCfg.Cache.Options
	}
}
