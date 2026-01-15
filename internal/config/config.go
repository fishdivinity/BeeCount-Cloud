package config

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"time"

	"github.com/spf13/viper"
)

// Config 应用配置
// 包含所有配置项
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	Storage  StorageConfig  `mapstructure:"storage"`
	Log      LogConfig      `mapstructure:"log"`
	CORS     CORSConfig     `mapstructure:"cors"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
		Port            int           `mapstructure:"port"`
		Mode            string        `mapstructure:"mode"`
		ReadTimeout     time.Duration `mapstructure:"read_timeout"`
		WriteTimeout    time.Duration `mapstructure:"write_timeout"`
		Docs            DocsConfig    `mapstructure:"docs"`
		AllowRegistration bool         `mapstructure:"allow_registration"`
		AdminAccount    AdminConfig   `mapstructure:"admin_account"`
	}

// AdminConfig 管理员账户配置
type AdminConfig struct {
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Type     string           `mapstructure:"type"`
	SQLite   SQLiteConfig     `mapstructure:"sqlite"`
	MySQL    MySQLConfig      `mapstructure:"mysql"`
	Postgres PostgresConfig   `mapstructure:"postgres"`
	Active   string          `mapstructure:"active"`
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

// JWTConfig JWT配置
type JWTConfig struct {
	Secret              string `mapstructure:"secret"`
	ExpireHours         int    `mapstructure:"expire_hours"`
	RotationIntervalDays int    `mapstructure:"rotation_interval_days"`
	LastRotationDate    string `mapstructure:"last_rotation_date"`
	PreviousSecret      string `mapstructure:"previous_secret"`
}

// StorageConfig 存储配置
type StorageConfig struct {
	Type             string      `mapstructure:"type"`
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

// LogConfig 日志配置
type LogConfig struct {
	Level  string     `mapstructure:"level"`
	Format string     `mapstructure:"format"`
	Output string     `mapstructure:"output"`
	File   FileConfig `mapstructure:"file"`
}

// FileConfig 文件日志配置
type FileConfig struct {
	Path          string `mapstructure:"path"`
	MaxSize       int    `mapstructure:"max_size"`
	MaxBackups    int    `mapstructure:"max_backups"`
	MaxAge        int    `mapstructure:"max_age"`
	Compress      bool   `mapstructure:"compress"`
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

// Load 加载配置文件
// 从指定路径加载配置文件并解析为Config结构
// 如果配置文件不存在，则自动生成默认配置
func Load(configPath string) (*Config, error) {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	viper.BindEnv("config_path", "CONFIG_PATH")
	viper.BindEnv("database_type", "DATABASE_TYPE")
	viper.BindEnv("storage_type", "STORAGE_TYPE")
	viper.BindEnv("admin_password", "ADMIN_PASSWORD")

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		if err := GenerateDefaultConfig(configPath); err != nil {
			return nil, fmt.Errorf("failed to generate default config: %w", err)
		}
	}

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}

// GenerateDefaultConfig 生成默认配置文件
// 在指定路径生成默认的config.yaml文件
func GenerateDefaultConfig(configPath string) error {
	defaultConfig := `# BeeCount Cloud Configuration

server:
  port: 8080
  mode: debug # debug, release, test
  read_timeout: 60s
  write_timeout: 60s
  docs:
    enabled: true
  allow_registration: false
  admin_account:
    username: beecount
    password: beecount_admin_2024

database:
  type: sqlite # sqlite, mysql, postgres
  sqlite:
    path: ./data/beecount.db
  mysql:
    host: localhost
    port: 3306
    username: root
    password: password
    database: beecount
    charset: utf8mb4
    parse_time: true
    loc: Local
  postgres:
    host: localhost
    port: 5432
    username: postgres
    password: password
    database: beecount
    sslmode: disable
    timezone: UTC

jwt:
  secret: %s
  expire_hours: 24
  rotation_interval_days: 7
  last_rotation_date: ""
  previous_secret: ""

storage:
  type: local # local, s3
  max_file_size: 10485760 # 10MB in bytes
  allowed_file_types:
    - image/jpeg
    - image/png
    - image/gif
    - image/webp
  local:
    path: ./data/uploads
    url_prefix: /uploads
  s3:
    region: us-east-1
    bucket: beecount-uploads
    access_key_id: your-access-key
    secret_access_key: your-secret-key
    endpoint: https://s3.amazonaws.com

log:
  level: info # debug, info, warn, error
  format: json # json, console
  output: stdout # stdout, file
  file:
    path: ./logs/app.log
    max_size: 100 # MB
    max_backups: 3
    max_age: 28 # days
    compress: true
    max_total_size_gb: 10

cors:
  allowed_origins:
    - "*"
  allowed_methods:
    - GET
    - POST
    - PUT
    - DELETE
    - OPTIONS
  allowed_headers:
    - "*"
  exposed_headers:
    - Content-Length
  allow_credentials: true
  max_age: 12h
`

	secret, err := generateRandomSecret()
	if err != nil {
		return fmt.Errorf("failed to generate secret: %w", err)
	}

	configContent := fmt.Sprintf(defaultConfig, secret)

	dir := configPath[:len(configPath)-len("/config.yaml")]
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// generateRandomSecret 生成随机密钥
func generateRandomSecret() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
