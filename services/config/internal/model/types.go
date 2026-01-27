package model

// ExplicitlySetFields 记录配置文件中显式设置的字段
type ExplicitlySetFields struct {
	ServerDocsEnabled      bool
	DatabaseMySQLParseTime bool
	LogFileCompress        bool
	CORSAllowCredentials   bool
}

// ConfigChangeEvent 配置变更事件
type ConfigChangeEvent struct {
	Key       string
	OldValue  *Config
	NewValue  *Config
	Version   string
	Timestamp string
}

// ConfigSource 配置来源
type ConfigSource int

const (
	// ConfigSourceFile 配置文件来源
	ConfigSourceFile ConfigSource = iota
	// ConfigSourceEnv 环境变量来源
	ConfigSourceEnv
	// ConfigSourceGRPC gRPC请求来源
	ConfigSourceGRPC
)

// ConfigUpdateRequest 配置更新请求
type ConfigUpdateRequest struct {
	Config *Config
	Source ConfigSource
}

// ConfigItem 配置项
type ConfigItem struct {
	Key   string
	Value interface{}
	Type  string
}
