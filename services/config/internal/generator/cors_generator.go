package generator

import (
	"fmt"
	"github.com/fishdivinity/BeeCount-Cloud/services/config/internal/model"
)

// GenerateCORSConfig 生成CORS配置内容
func GenerateCORSConfig(cfg *model.CORSConfig) string {
	return `# CORS Configuration
cors:
  allowed_origins:
    ` + generateAllowedOrigins(cfg.AllowedOrigins) + `
  allowed_methods:
    ` + generateAllowedMethods(cfg.AllowedMethods) + `
  allowed_headers:
    ` + generateAllowedHeaders(cfg.AllowedHeaders) + `
  exposed_headers:
    ` + generateExposedHeaders(cfg.ExposedHeaders) + `
  allow_credentials: ` + fmt.Sprintf("%t", cfg.AllowCredentials) + ` # Allow credentials
  max_age: ` + cfg.MaxAge + ` # Preflight request cache time
`
}
