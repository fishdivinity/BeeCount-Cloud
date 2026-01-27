package model

import (
	"time"
)

// Config 应用配置
// 包含所有配置项
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Storage  StorageConfig  `mapstructure:"storage"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	Log      LogConfig      `mapstructure:"log"`
	CORS     CORSConfig     `mapstructure:"cors"`
	Cache    CacheConfig    `mapstructure:"cache"`
}

// CacheConfig 缓存配置
type CacheConfig struct {
	Active  string                 `mapstructure:"active"`
	Memory  MemoryCacheConfig      `mapstructure:"memory"`
	Redis   RedisCacheConfig       `mapstructure:"redis"`
	Options map[string]interface{} `mapstructure:"options"`
}

// MemoryCacheConfig 内存缓存配置
type MemoryCacheConfig struct {
	MaxSize int `mapstructure:"max_size"`
}

// RedisCacheConfig Redis缓存配置
type RedisCacheConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port              int           `mapstructure:"port"`
	Mode              string        `mapstructure:"mode"`
	ReadTimeout       time.Duration `mapstructure:"read_timeout"`
	WriteTimeout      time.Duration `mapstructure:"write_timeout"`
	Docs              DocsConfig    `mapstructure:"docs"`
	AdminAccount      AdminConfig   `mapstructure:"admin_account"`
}

// AdminConfig 管理员账户配置
type AdminConfig struct {
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	SQLite   SQLiteConfig   `mapstructure:"sqlite"`
	MySQL    MySQLConfig    `mapstructure:"mysql"`
	Postgres PostgresConfig `mapstructure:"postgres"`
	Active   string         `mapstructure:"active"`
	Pool     PoolConfig     `mapstructure:"pool"`
}

// PoolConfig 数据库连接池配置
type PoolConfig struct {
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration `mapstructure:"conn_max_idle_time"`
}

// SQLiteConfig SQLite配置
type SQLiteConfig struct {
	Path string `mapstructure:"path"`
}

// MySQLConfig MySQL配置
type MySQLConfig struct {
	Host      string `mapstructure:"host"`
	Port      int    `mapstructure:"port"`
	Username  string `mapstructure:"username"`
	Password  string `mapstructure:"password"`
	Database  string `mapstructure:"database"`
	Charset   string `mapstructure:"charset"`
	ParseTime bool   `mapstructure:"parse_time"`
	Loc       string `mapstructure:"loc"`
}

// PostgresConfig PostgreSQL配置
type PostgresConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Database string `mapstructure:"database"`
	SSLMode  string `mapstructure:"sslmode"`
	Timezone string `mapstructure:"timezone"`
}

// StorageConfig 存储配置
type StorageConfig struct {
	Local            LocalConfig `mapstructure:"local"`
	S3               S3Config    `mapstructure:"s3"`
	MaxFileSize      int64       `mapstructure:"max_file_size"`
	AllowedFileTypes []string    `mapstructure:"allowed_file_types"`
	Active           string      `mapstructure:"active"`
}

// LocalConfig 本地存储配置
type LocalConfig struct {
	Path      string `mapstructure:"path"`
	URLPrefix string `mapstructure:"url_prefix"`
}

// S3Config S3存储配置
type S3Config struct {
	Region          string `mapstructure:"region"`
	Bucket          string `mapstructure:"bucket"`
	AccessKeyID     string `mapstructure:"access_key_id"`
	SecretAccessKey string `mapstructure:"secret_access_key"`
	Endpoint        string `mapstructure:"endpoint"`
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret               string `mapstructure:"secret"`
	ExpireHours          int    `mapstructure:"expire_hours"`
	RotationIntervalDays int    `mapstructure:"rotation_interval_days"`
	LastRotationDate     string `mapstructure:"last_rotation_date"`
	PreviousSecret       string `mapstructure:"previous_secret"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level  string     `mapstructure:"level"`
	Format string     `mapstructure:"format"`
	Output string     `mapstructure:"output"`
	File   FileConfig `mapstructure:"file"`
}

// FileConfig 文件日志配置
type FileConfig struct {
	Path           string `mapstructure:"path"`
	MaxSize        int    `mapstructure:"max_size"`
	MaxBackups     int    `mapstructure:"max_backups"`
	MaxAge         int    `mapstructure:"max_age"`
	Compress       bool   `mapstructure:"compress"`
	MaxTotalSizeGB int    `mapstructure:"max_total_size_gb"`
}

// CORSConfig CORS配置
type CORSConfig struct {
	AllowedOrigins   []string `mapstructure:"allowed_origins"`
	AllowedMethods   []string `mapstructure:"allowed_methods"`
	AllowedHeaders   []string `mapstructure:"allowed_headers"`
	ExposedHeaders   []string `mapstructure:"exposed_headers"`
	AllowCredentials bool     `mapstructure:"allow_credentials"`
	MaxAge           string   `mapstructure:"max_age"`
}

// DocsConfig API文档配置
type DocsConfig struct {
	Enabled bool `mapstructure:"enabled"`
}
