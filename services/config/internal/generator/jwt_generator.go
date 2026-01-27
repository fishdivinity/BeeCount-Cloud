package generator

import (
	"fmt"
	"github.com/fishdivinity/BeeCount-Cloud/services/config/internal/model"
)

// GenerateJWTConfig 生成JWT配置内容
func GenerateJWTConfig(cfg *model.JWTConfig) string {
	return `# JWT Configuration
jwt:
  secret: ` + cfg.Secret + ` # JWT secret, automatically generated on first start
  expire_hours: ` + fmt.Sprintf("%d", cfg.ExpireHours) + ` # JWT expiration time (hours)
  rotation_interval_days: ` + fmt.Sprintf("%d", cfg.RotationIntervalDays) + ` # Secret rotation interval (days)
  last_rotation_date: "` + cfg.LastRotationDate + `" # Last secret rotation date
  previous_secret: "` + cfg.PreviousSecret + `" # Old secret for compatibility with currently used JWT tokens
`
}
