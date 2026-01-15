package jwt

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"time"

	"github.com/fishdivinity/BeeCount-Cloud/internal/config"
	"gopkg.in/yaml.v3"
)

// SecretManager JWT密钥管理器
type SecretManager struct {
	configPath string
	config     *config.JWTConfig
}

// NewSecretManager 创建JWT密钥管理器
func NewSecretManager(configPath string, cfg *config.JWTConfig) *SecretManager {
	return &SecretManager{
		configPath: configPath,
		config:     cfg,
	}
}

// CheckRotation 检查是否需要轮换密钥
func (m *SecretManager) CheckRotation() error {
	if m.config.Secret == "" {
		newSecret, err := generateRandomSecret()
		if err != nil {
			return fmt.Errorf("failed to generate initial secret: %w", err)
		}
		m.config.Secret = newSecret
		m.config.LastRotationDate = time.Now().Format(time.RFC3339)
		return m.saveConfig()
	}

	if m.config.RotationIntervalDays <= 0 {
		m.config.RotationIntervalDays = 7
	}

	if m.config.LastRotationDate == "" {
		m.config.LastRotationDate = time.Now().Format(time.RFC3339)
		return m.saveConfig()
	}

	lastRotation, err := time.Parse(time.RFC3339, m.config.LastRotationDate)
	if err != nil {
		return fmt.Errorf("failed to parse last rotation date: %w", err)
	}

	daysSinceRotation := int(time.Since(lastRotation).Hours() / 24)

	if daysSinceRotation >= m.config.RotationIntervalDays {
		return m.RotateSecret()
	}

	return nil
}

// RotateSecret 轮换JWT密钥
func (m *SecretManager) RotateSecret() error {
	newSecret, err := generateRandomSecret()
	if err != nil {
		return fmt.Errorf("failed to generate new secret: %w", err)
	}

	m.config.PreviousSecret = m.config.Secret
	m.config.Secret = newSecret
	m.config.LastRotationDate = time.Now().Format(time.RFC3339)

	return m.saveConfig()
}

// GetSecrets 获取当前和之前的密钥
func (m *SecretManager) GetSecrets() (current, previous string) {
	return m.config.Secret, m.config.PreviousSecret
}

// saveConfig 保存配置到文件
func (m *SecretManager) saveConfig() error {
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg map[string]interface{}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if jwtCfg, ok := cfg["jwt"].(map[string]interface{}); ok {
		jwtCfg["secret"] = m.config.Secret
		jwtCfg["previous_secret"] = m.config.PreviousSecret
		jwtCfg["last_rotation_date"] = m.config.LastRotationDate
	}

	newData, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(m.configPath, newData, 0644); err != nil {
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