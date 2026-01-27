package internal

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/fishdivinity/BeeCount-Cloud/common/proto/auth"
	"github.com/fishdivinity/BeeCount-Cloud/common/proto/common"
	"github.com/golang-jwt/jwt/v4"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// JWTConfig JWT配置
type JWTConfig struct {
	Secret               string
	ExpireHours          int
	RotationIntervalDays int
}

// BannedIP 封禁IP信息
type BannedIP struct {
	Reason     string
	ExpireTime int64
}

// AuthService 认证服务实现
type AuthService struct {
	auth.UnimplementedAuthServiceServer
	common.UnimplementedHealthCheckServiceServer

	// JWT配置
	jwtConfig JWTConfig
	mu        sync.RWMutex

	// 封禁IP列表
	bannedIPs map[string]BannedIP
	banMu     sync.RWMutex
}

// NewAuthService 创建认证服务实例
func NewAuthService() *AuthService {
	return &AuthService{
		bannedIPs: make(map[string]BannedIP),
	}
}

// ConfigureJWT 配置JWT
func (s *AuthService) ConfigureJWT(config JWTConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.jwtConfig = config
	return nil
}

// GenerateToken 生成JWT令牌
func (s *AuthService) GenerateToken(ctx context.Context, req *auth.GenerateTokenRequest) (*auth.GenerateTokenResponse, error) {
	s.mu.RLock()
	config := s.jwtConfig
	s.mu.RUnlock()

	// 设置过期时间
	expireTime := time.Now().Add(time.Duration(req.ExpireHours) * time.Hour)
	if req.ExpireHours == 0 {
		expireTime = time.Now().Add(time.Duration(config.ExpireHours) * time.Hour)
	}

	// 创建JWT声明
	claims := jwt.MapClaims{
		"user_id":  req.UserId,
		"username": req.Username,
		"exp":      expireTime.Unix(),
		"iat":      time.Now().Unix(),
	}

	// 添加自定义声明
	for k, v := range req.Claims {
		claims[k] = v
	}

	// 创建令牌
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 签名令牌
	tokenString, err := token.SignedString([]byte(config.Secret))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to sign token: %v", err)
	}

	// 生成刷新令牌（简单实现，实际可更复杂）
	refreshToken := fmt.Sprintf("refresh_%s_%d", tokenString[:10], time.Now().Unix())

	return &auth.GenerateTokenResponse{
		Token:        tokenString,
		RefreshToken: refreshToken,
		ExpireAt:     expireTime.Unix(),
	}, nil
}

// ValidateToken 验证JWT令牌
func (s *AuthService) ValidateToken(ctx context.Context, req *auth.ValidateTokenRequest) (*auth.ValidateTokenResponse, error) {
	s.mu.RLock()
	config := s.jwtConfig
	s.mu.RUnlock()

	// 解析令牌
	token, err := jwt.Parse(req.Token, func(token *jwt.Token) (interface{}, error) {
		// 验证签名方法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(config.Secret), nil
	})

	if err != nil {
		return &auth.ValidateTokenResponse{
			Valid: false,
		}, nil
	}

	// 验证令牌有效性
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// 提取声明
		userID, _ := claims["user_id"].(string)
		username, _ := claims["username"].(string)
		exp, _ := claims["exp"].(float64)

		// 提取自定义声明
		customClaims := make(map[string]string)
		for k, v := range claims {
			if k != "user_id" && k != "username" && k != "exp" && k != "iat" {
				if strV, ok := v.(string); ok {
					customClaims[k] = strV
				}
			}
		}

		return &auth.ValidateTokenResponse{
			Valid:    true,
			UserId:   userID,
			Username: username,
			Claims:   customClaims,
			ExpireAt: int64(exp),
		}, nil
	}

	return &auth.ValidateTokenResponse{
		Valid: false,
	}, nil
}

// BanIP 封禁IP
func (s *AuthService) BanIP(ctx context.Context, req *auth.BanIPRequest) (*common.Response, error) {
	s.banMu.Lock()
	defer s.banMu.Unlock()

	// 计算过期时间
	expireTime := time.Now().Add(time.Duration(req.DurationSeconds) * time.Second).Unix()

	// 添加到封禁列表
	s.bannedIPs[req.Ip] = BannedIP{
		Reason:     req.Reason,
		ExpireTime: expireTime,
	}

	return &common.Response{
		Success: true,
		Message: fmt.Sprintf("IP %s banned for %d seconds", req.Ip, req.DurationSeconds),
		Code:    200,
	}, nil
}

// UnbanIP 解封IP
func (s *AuthService) UnbanIP(ctx context.Context, req *auth.UnbanIPRequest) (*common.Response, error) {
	s.banMu.Lock()
	defer s.banMu.Unlock()

	// 从封禁列表中移除
	delete(s.bannedIPs, req.Ip)

	return &common.Response{
		Success: true,
		Message: fmt.Sprintf("IP %s unbanned", req.Ip),
		Code:    200,
	}, nil
}

// CheckIP 检查IP是否被封禁
func (s *AuthService) CheckIP(ctx context.Context, req *auth.CheckIPRequest) (*auth.CheckIPResponse, error) {
	s.banMu.RLock()
	bannedIP, exists := s.bannedIPs[req.Ip]
	s.banMu.RUnlock()

	// 检查IP是否被封禁
	if !exists {
		return &auth.CheckIPResponse{
			IsBanned: false,
		}, nil
	}

	// 检查是否过期
	currentTime := time.Now().Unix()
	if currentTime > bannedIP.ExpireTime {
		// 过期，移除封禁
		s.banMu.Lock()
		delete(s.bannedIPs, req.Ip)
		s.banMu.Unlock()

		return &auth.CheckIPResponse{
			IsBanned: false,
		}, nil
	}

	return &auth.CheckIPResponse{
		IsBanned:    true,
		Reason:      bannedIP.Reason,
		BanExpireAt: bannedIP.ExpireTime,
	}, nil
}

// Check 健康检查
func (s *AuthService) Check(ctx context.Context, req *common.HealthCheckRequest) (*common.HealthCheckResponse, error) {
	return &common.HealthCheckResponse{
		Status: common.HealthCheckResponse_SERVING,
	}, nil
}

// Watch 健康检查监听
func (s *AuthService) Watch(req *common.HealthCheckRequest, stream common.HealthCheckService_WatchServer) error {
	// 实现健康检查监听逻辑
	return status.Errorf(codes.Unimplemented, "method Watch not implemented")
}
