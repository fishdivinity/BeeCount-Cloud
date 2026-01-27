package generator

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/fishdivinity/BeeCount-Cloud/services/config/internal/model"
)

// GenerateDefaultConfig 生成默认配置文件
// 在指定路径生成默认的config.yaml文件
func GenerateDefaultConfig(configPath string) error {
	// 生成随机密钥
	secret, err := GenerateRandomSecret()
	if err != nil {
		return fmt.Errorf("failed to generate secret: %w", err)
	}

	// 生成默认配置
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
			Secret:               secret,
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

	// 生成配置文件内容
	configContent := generateConfigContent(defaultCfg)

	// 获取目录路径
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GenerateSplitConfigFiles 生成拆分的配置文件
// 为每种配置类型生成独立的YAML文件
func GenerateSplitConfigFiles(configDir string) error {
	// 生成随机密钥
	secret, err := GenerateRandomSecret()
	if err != nil {
		return fmt.Errorf("failed to generate secret: %w", err)
	}

	// 生成默认配置
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
		},
		JWT: model.JWTConfig{
			Secret:               secret,
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
	}

	// 确保配置目录存在
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// 生成各配置文件
	configFiles := map[string]string{
		"server.yaml":   GenerateServerConfig(&defaultCfg.Server),
		"database.yaml": GenerateDatabaseConfig(&defaultCfg.Database),
		"storage.yaml":  GenerateStorageConfig(&defaultCfg.Storage),
		"jwt.yaml":      GenerateJWTConfig(&defaultCfg.JWT),
		"log.yaml":      GenerateLogConfig(&defaultCfg.Log),
		"cors.yaml":     GenerateCORSConfig(&defaultCfg.CORS),
	}

	// 写入配置文件
	for filename, content := range configFiles {
		filePath := filepath.Join(configDir, filename)
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write %s: %w", filename, err)
		}
	}

	return nil
}

// GenerateSingleConfigFile 生成单个配置文件
func GenerateSingleConfigFile(cfg *model.Config, configPath string) error {
	// 生成配置文件内容
	configContent := generateConfigContent(cfg)

	// 获取目录路径
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// generateConfigContent 生成配置文件内容
// 基于完整配置对象生成配置文件内容
func generateConfigContent(cfg *model.Config) string {
	return `# BeeCount Cloud Configuration File
# This file contains all configuration options for the BeeCount Cloud service

# Server Configuration
server:
  port: ` + fmt.Sprintf("%d", cfg.Server.Port) + `
  mode: ` + cfg.Server.Mode + `
  read_timeout: ` + fmt.Sprintf("%v", cfg.Server.ReadTimeout) + `
  write_timeout: ` + fmt.Sprintf("%v", cfg.Server.WriteTimeout) + `
  docs:
    enabled: ` + fmt.Sprintf("%t", cfg.Server.Docs.Enabled) + `
  admin_account:
    username: ` + cfg.Server.AdminAccount.Username + `
    password: ` + cfg.Server.AdminAccount.Password + `

# Database Configuration
database:
  active: ` + cfg.Database.Active + `
  sqlite:
    path: ` + cfg.Database.SQLite.Path + `
  mysql:
    host: ` + cfg.Database.MySQL.Host + `
    port: ` + fmt.Sprintf("%d", cfg.Database.MySQL.Port) + `
    username: ` + cfg.Database.MySQL.Username + `
    password: ` + cfg.Database.MySQL.Password + `
    database: ` + cfg.Database.MySQL.Database + `
    charset: ` + cfg.Database.MySQL.Charset + `
    parse_time: ` + fmt.Sprintf("%t", cfg.Database.MySQL.ParseTime) + `
    loc: ` + cfg.Database.MySQL.Loc + `
  postgres:
    host: ` + cfg.Database.Postgres.Host + `
    port: ` + fmt.Sprintf("%d", cfg.Database.Postgres.Port) + `
    username: ` + cfg.Database.Postgres.Username + `
    password: ` + cfg.Database.Postgres.Password + `
    database: ` + cfg.Database.Postgres.Database + `
    sslmode: ` + cfg.Database.Postgres.SSLMode + `
    timezone: ` + cfg.Database.Postgres.Timezone + `
  pool:
    max_idle_conns: ` + fmt.Sprintf("%d", cfg.Database.Pool.MaxIdleConns) + `
    max_open_conns: ` + fmt.Sprintf("%d", cfg.Database.Pool.MaxOpenConns) + `
    conn_max_lifetime: ` + fmt.Sprintf("%v", cfg.Database.Pool.ConnMaxLifetime) + `
    conn_max_idle_time: ` + fmt.Sprintf("%v", cfg.Database.Pool.ConnMaxIdleTime) + `

# Storage Configuration
storage:
  active: ` + cfg.Storage.Active + `
  max_file_size: ` + fmt.Sprintf("%d", cfg.Storage.MaxFileSize) + `
  allowed_file_types:
    ` + generateAllowedFileTypes(cfg.Storage.AllowedFileTypes) + `
  local:
    path: ` + cfg.Storage.Local.Path + `
    url_prefix: ` + cfg.Storage.Local.URLPrefix + `
  s3:
    region: ` + cfg.Storage.S3.Region + `
    bucket: ` + cfg.Storage.S3.Bucket + `
    access_key_id: ` + cfg.Storage.S3.AccessKeyID + `
    secret_access_key: ` + cfg.Storage.S3.SecretAccessKey + `
    endpoint: ` + cfg.Storage.S3.Endpoint + `

# JWT Configuration
jwt:
  secret: ` + cfg.JWT.Secret + `
  expire_hours: ` + fmt.Sprintf("%d", cfg.JWT.ExpireHours) + `
  rotation_interval_days: ` + fmt.Sprintf("%d", cfg.JWT.RotationIntervalDays) + `
  last_rotation_date: "` + cfg.JWT.LastRotationDate + `"
  previous_secret: "` + cfg.JWT.PreviousSecret + `"

# Cache Configuration
cache:
  active: ` + cfg.Cache.Active + `
  memory:
    max_size: ` + fmt.Sprintf("%d", cfg.Cache.Memory.MaxSize) + `
  redis:
    host: ` + cfg.Cache.Redis.Host + `
    port: ` + fmt.Sprintf("%d", cfg.Cache.Redis.Port) + `
    password: "` + cfg.Cache.Redis.Password + `"
    db: ` + fmt.Sprintf("%d", cfg.Cache.Redis.DB) + `

# Log Configuration
log:
  level: ` + cfg.Log.Level + `
  format: ` + cfg.Log.Format + `
  output: ` + cfg.Log.Output + `
  file:
    path: ` + cfg.Log.File.Path + `
    max_size: ` + fmt.Sprintf("%d", cfg.Log.File.MaxSize) + `
    max_backups: ` + fmt.Sprintf("%d", cfg.Log.File.MaxBackups) + `
    max_age: ` + fmt.Sprintf("%d", cfg.Log.File.MaxAge) + `
    compress: ` + fmt.Sprintf("%t", cfg.Log.File.Compress) + `
    max_total_size_gb: ` + fmt.Sprintf("%d", cfg.Log.File.MaxTotalSizeGB) + `

# CORS Configuration
cors:
  allowed_origins:
    ` + generateAllowedOrigins(cfg.CORS.AllowedOrigins) + `
  allowed_methods:
    ` + generateAllowedMethods(cfg.CORS.AllowedMethods) + `
  allowed_headers:
    ` + generateAllowedHeaders(cfg.CORS.AllowedHeaders) + `
  exposed_headers:
    ` + generateExposedHeaders(cfg.CORS.ExposedHeaders) + `
  allow_credentials: ` + fmt.Sprintf("%t", cfg.CORS.AllowCredentials) + `
  max_age: ` + cfg.CORS.MaxAge + `
`
}

// generateAllowedFileTypes 生成允许的文件类型配置
func generateAllowedFileTypes(fileTypes []string) string {
	var result string
	for i, fileType := range fileTypes {
		if i > 0 {
			result += "\n    "
		}
		result += fmt.Sprintf("- %s", fileType)
	}
	return result
}

// generateAllowedOrigins 生成允许的源配置
func generateAllowedOrigins(origins []string) string {
	var result string
	for i, origin := range origins {
		if i > 0 {
			result += "\n    "
		}
		// 确保特殊字符被正确引用
		if origin == "*" {
			result += fmt.Sprintf("- \"%s\"", origin)
		} else {
			result += fmt.Sprintf("- %s", origin)
		}
	}
	return result
}

// generateAllowedMethods 生成允许的方法配置
func generateAllowedMethods(methods []string) string {
	var result string
	for i, method := range methods {
		if i > 0 {
			result += "\n    "
		}
		result += fmt.Sprintf("- %s", method)
	}
	return result
}

// generateAllowedHeaders 生成允许的头配置
func generateAllowedHeaders(headers []string) string {
	var result string
	for i, header := range headers {
		if i > 0 {
			result += "\n    "
		}
		// 确保特殊字符被正确引用
		if header == "*" {
			result += fmt.Sprintf("- \"%s\"", header)
		} else {
			result += fmt.Sprintf("- %s", header)
		}
	}
	return result
}

// generateExposedHeaders 生成暴露的头配置
func generateExposedHeaders(headers []string) string {
	var result string
	for i, header := range headers {
		if i > 0 {
			result += "\n    "
		}
		result += fmt.Sprintf("- %s", header)
	}
	return result
}

// GenerateRandomSecret 生成随机密钥
func GenerateRandomSecret() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
