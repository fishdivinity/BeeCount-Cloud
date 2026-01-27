package generator

import (
	"fmt"
	"github.com/fishdivinity/BeeCount-Cloud/services/config/internal/model"
)

// GenerateLogConfig 生成日志配置内容
func GenerateLogConfig(cfg *model.LogConfig) string {
	return `# Log Configuration
log:
  level: ` + cfg.Level + ` # Log level
  format: ` + cfg.Format + ` # Log format
  output: ` + cfg.Output + ` # Log output
  file: # File output configuration
    path: ` + cfg.File.Path + ` # Log file path
    max_size: ` + fmt.Sprintf("%d", cfg.File.MaxSize) + ` # Maximum size of a single log file (MB)
    max_backups: ` + fmt.Sprintf("%d", cfg.File.MaxBackups) + ` # Number of log files to keep
    max_age: ` + fmt.Sprintf("%d", cfg.File.MaxAge) + ` # Log file retention time (days)
    compress: ` + fmt.Sprintf("%t", cfg.File.Compress) + ` # Whether to compress log files
    max_total_size_gb: ` + fmt.Sprintf("%d", cfg.File.MaxTotalSizeGB) + ` # Maximum total size of log files (GB)
`
}
