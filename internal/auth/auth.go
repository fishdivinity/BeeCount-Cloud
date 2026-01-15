package auth

import (
	"errors"
	"time"

	"github.com/fishdivinity/BeeCount-Cloud/internal/config"
	"github.com/fishdivinity/BeeCount-Cloud/internal/models"
	jwtlib "github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Claims JWT声明结构体
// 包含用户信息和JWT标准声明
type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	jwtlib.RegisteredClaims
}

// AuthService 认证服务接口
// 定义认证相关的方法，包括token生成、验证和密码处理
type AuthService interface {
	GenerateToken(user *models.User) (string, error)
	ValidateToken(tokenString string) (*Claims, error)
	HashPassword(password string) (string, error)
	CheckPassword(password, hash string) error
}

// JWTAuthService JWT认证服务实现
type JWTAuthService struct {
	currentSecret  string
	previousSecret string
	expireHours   int
}

// NewJWTAuthService 创建JWT认证服务实例
func NewJWTAuthService(cfg *config.JWTConfig) AuthService {
	return &JWTAuthService{
		currentSecret:  cfg.Secret,
		previousSecret: cfg.PreviousSecret,
		expireHours:   cfg.ExpireHours,
	}
}

// NewJWTAuthServiceWithSecrets 使用指定密钥创建JWT认证服务实例
func NewJWTAuthServiceWithSecrets(current, previous string, expireHours int) AuthService {
	return &JWTAuthService{
		currentSecret:  current,
		previousSecret: previous,
		expireHours:   expireHours,
	}
}

// GenerateToken 生成JWT token
// 根据用户信息生成带有过期时间的JWT token
func (s *JWTAuthService) GenerateToken(user *models.User) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:   user.ID,
		Username: user.Username,
		Email:    user.Email,
		RegisteredClaims: jwtlib.RegisteredClaims{
			ExpiresAt: jwtlib.NewNumericDate(now.Add(time.Hour * time.Duration(s.expireHours))),
			IssuedAt:  jwtlib.NewNumericDate(now),
			NotBefore: jwtlib.NewNumericDate(now),
		},
	}

	token := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.currentSecret))
}

// ValidateToken 验证JWT token
// 解析并验证token的有效性，支持当前和之前的密钥
func (s *JWTAuthService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwtlib.ParseWithClaims(tokenString, &Claims{}, func(token *jwtlib.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwtlib.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(s.currentSecret), nil
	})

	if err == nil {
		if claims, ok := token.Claims.(*Claims); ok && token.Valid {
			return claims, nil
		}
	}

	if s.previousSecret != "" {
		token, err = jwtlib.ParseWithClaims(tokenString, &Claims{}, func(token *jwtlib.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwtlib.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return []byte(s.previousSecret), nil
		})

		if err == nil {
			if claims, ok := token.Claims.(*Claims); ok && token.Valid {
				return claims, nil
			}
		}
	}

	return nil, errors.New("invalid token")
}

// HashPassword 加密密码
// 使用bcrypt算法加密密码
func (s *JWTAuthService) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPassword 验证密码
// 比较明文密码和加密后的哈希值
func (s *JWTAuthService) CheckPassword(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

