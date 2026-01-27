package generator

import (
	"fmt"
	"github.com/fishdivinity/BeeCount-Cloud/services/config/internal/model"
)

// GenerateServerConfig 生成服务器配置内容
func GenerateServerConfig(cfg *model.ServerConfig) string {
	return `# Server Configuration
server:
  port: ` + fmt.Sprintf("%d", cfg.Port) + `              # Server listening port
  mode: ` + cfg.Mode + `           # Running mode: debug, release, test
  read_timeout: ` + fmt.Sprintf("%v", cfg.ReadTimeout) + `       # Read timeout
  write_timeout: ` + fmt.Sprintf("%v", cfg.WriteTimeout) + `      # Write timeout
  docs:
    enabled: ` + fmt.Sprintf("%t", cfg.Docs.Enabled) + `        # Whether to enable API documentation
  admin_account:
    username: ` + cfg.AdminAccount.Username + `    # Admin username
    password: ` + cfg.AdminAccount.Password + ` # Admin password (please change after first use)
`
}
