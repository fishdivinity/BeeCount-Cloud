package loader

import (
	"fmt"
	"os"
	"time"

	"github.com/fishdivinity/BeeCount-Cloud/services/config/internal/model"
	"github.com/spf13/viper"
)

// NewViper 创建并配置Viper实例
func NewViper(configPath string) *viper.Viper {
	v := viper.New()
	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")

	// 绑定环境变量
	BindServerEnv(v)
	BindDatabaseEnv(v)
	BindStorageEnv(v)
	BindJWTEnv(v)
	BindLogEnv(v)
	BindCORSEnv(v)

	return v
}

// LoadConfig 加载配置文件
// 从指定路径加载配置文件并解析为Config结构
func LoadConfig(configPath string) (*model.Config, *model.ExplicitlySetFields, error) {
	v := NewViper(configPath)

	configExists := true
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		configExists = false
	}

	var cfg model.Config
	if !configExists {
		// 配置文件不存在，直接返回错误，让调用者处理
		return nil, nil, os.ErrNotExist
	}

	// 读取配置文件
	if err := v.ReadInConfig(); err != nil {
		return nil, nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// 加载各类型配置
	serverCfg, err := LoadServerConfig(v)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load server config: %w", err)
	}

	dbCfg, err := LoadDatabaseConfig(v)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load database config: %w", err)
	}

	storageCfg, err := LoadStorageConfig(v)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load storage config: %w", err)
	}

	jwtCfg, err := LoadJWTConfig(v)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load jwt config: %w", err)
	}

	logCfg, err := LoadLogConfig(v)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load log config: %w", err)
	}

	corsCfg, err := LoadCORSConfig(v)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load cors config: %w", err)
	}

	// 构建完整配置
	cfg = model.Config{
		Server:   *serverCfg,
		Database: *dbCfg,
		Storage:  *storageCfg,
		JWT:      *jwtCfg,
		Log:      *logCfg,
		CORS:     *corsCfg,
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

	// 检查配置文件中显式设置的字段
	explicitlySet := &model.ExplicitlySetFields{
		ServerDocsEnabled:      v.IsSet("server.docs.enabled"),
		DatabaseMySQLParseTime: v.IsSet("database.mysql.parse_time"),
		LogFileCompress:        v.IsSet("log.file.compress"),
		CORSAllowCredentials:   v.IsSet("cors.allow_credentials"),
	}

	// 检查是否存在环境变量
	hasEnvVars := checkForEnvVars()

	if hasEnvVars {
		// 存在环境变量，使用环境变量覆盖配置文件内容
		fmt.Println("[INFO] Environment variables found, overriding configuration file...")
	} else {
		// 不存在环境变量，使用配置文件内容，并生成对应的环境变量
		fmt.Println("[INFO] No environment variables found, using configuration file and generating environment variables...")
		// 将配置同步到环境变量
		// SyncConfigToEnv(&cfg) - 这个函数将在sync包中实现
	}

	// 生成默认配置，用于补充缺失的配置项
	defaultCfg := getDefaultConfig()

	// 合并配置，使用默认配置补充缺失的配置项
	mergeConfigs(&cfg, defaultCfg, explicitlySet)

	return &cfg, explicitlySet, nil
}

// getDefaultConfig 获取默认配置
func getDefaultConfig() *model.Config {
	return &model.Config{
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
			Secret:               "",
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
}

// mergeConfigs 合并两个配置结构体
// 将默认配置中的值合并到当前配置中，当前配置中的值优先
// 对于当前配置中未显式设置的字段，使用默认配置中的值
func mergeConfigs(cfg, defaultCfg *model.Config, explicitlySet *model.ExplicitlySetFields) {
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
	// 合并 Server.Docs 配置
	if !explicitlySet.ServerDocsEnabled {
		cfg.Server.Docs.Enabled = defaultCfg.Server.Docs.Enabled
	}
	// 合并 Server.AdminAccount 配置
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
	if !explicitlySet.DatabaseMySQLParseTime {
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
	if !explicitlySet.LogFileCompress {
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
	if !explicitlySet.CORSAllowCredentials {
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

// checkForEnvVars 检查是否存在环境变量
func checkForEnvVars() bool {
	// 检查关键环境变量是否存在
	return os.Getenv("DATABASE_TYPE") != "" ||
		os.Getenv("STORAGE_TYPE") != "" ||
		os.Getenv("ADMIN_PASSWORD") != "" ||
		os.Getenv("SERVER_PORT") != ""
}
