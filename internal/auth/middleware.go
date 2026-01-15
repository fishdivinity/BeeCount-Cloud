package auth

import (
	"net/http"
	"strings"

	"github.com/fishdivinity/BeeCount-Cloud/pkg/utils"
	"github.com/gin-gonic/gin"
)

const (
	// UserIDKey 用户ID在gin.Context中的键名
	UserIDKey = "user_id"
	// UsernameKey 用户名在gin.Context中的键名
	UsernameKey = "username"
	// EmailKey 邮箱在gin.Context中的键名
	EmailKey = "email"
)

// AuthMiddleware 认证中间件
// 验证请求头中的JWT token，将用户信息存入context
func AuthMiddleware(authService AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, utils.NewUnauthorizedError("missing authorization header"))
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, utils.NewUnauthorizedError("invalid authorization header format"))
			c.Abort()
			return
		}

		tokenString := parts[1]
		claims, err := authService.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, utils.NewUnauthorizedError("invalid or expired token"))
			c.Abort()
			return
		}

		c.Set(UserIDKey, claims.UserID)
		c.Set(UsernameKey, claims.Username)
		c.Set(EmailKey, claims.Email)

		c.Next()
	}
}

// GetUserID 从context中获取用户ID
func GetUserID(c *gin.Context) uint {
	userID, exists := c.Get(UserIDKey)
	if !exists {
		return 0
	}
	return userID.(uint)
}

// GetUsername 从context中获取用户名
func GetUsername(c *gin.Context) string {
	username, exists := c.Get(UsernameKey)
	if !exists {
		return ""
	}
	return username.(string)
}

// GetEmail 从context中获取邮箱
func GetEmail(c *gin.Context) string {
	email, exists := c.Get(EmailKey)
	if !exists {
		return ""
	}
	return email.(string)
}

