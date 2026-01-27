package generator

import (
	"fmt"

	"github.com/fishdivinity/BeeCount-Cloud/services/config/internal/model"
)

// GenerateStorageConfig 生成存储配置内容
func GenerateStorageConfig(cfg *model.StorageConfig) string {
	return `# Storage Configuration
storage:
  active: ` + cfg.Active + ` # Active storage type: local, s3
  max_file_size: ` + fmt.Sprintf("%d", cfg.MaxFileSize) + ` # Maximum upload file size (5MB)
  allowed_file_types: # Allowed upload file types
    ` + generateAllowedFileTypes(cfg.AllowedFileTypes) + `
  local: # Local storage configuration
    path: ` + cfg.Local.Path + ` # Local storage path
    url_prefix: ` + cfg.Local.URLPrefix + ` # Access prefix
  s3: # S3 storage configuration
    region: ` + cfg.S3.Region + ` # S3 region
    bucket: ` + cfg.S3.Bucket + ` # S3 bucket name
    access_key_id: ` + cfg.S3.AccessKeyID + ` # S3 access key ID
    secret_access_key: ` + cfg.S3.SecretAccessKey + ` # S3 secret access key
    endpoint: ` + cfg.S3.Endpoint + ` # S3 endpoint
`
}
