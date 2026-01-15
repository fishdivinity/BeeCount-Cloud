package middleware

import (
	"net/http"

	"github.com/fishdivinity/BeeCount-Cloud/pkg/logger"
	"github.com/fishdivinity/BeeCount-Cloud/pkg/utils"
	"github.com/gin-gonic/gin"
)

// ErrorHandler 错误处理中间件
// 捕获并处理请求中的错误
func ErrorHandler(log logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			for _, err := range c.Errors {
				log.Error("Unhandled error", logger.Error(err))
			}
			c.JSON(http.StatusInternalServerError, utils.NewInternalServerError("internal server error"))
		}
	}
}

// Recovery 恢复中间件
// 捕获panic并恢复，防止服务崩溃
func Recovery(log logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				log.Error("Panic recovered", logger.Any("panic", r))
				c.JSON(http.StatusInternalServerError, utils.NewInternalServerError("internal server error"))
				c.Abort()
			}
		}()
		c.Next()
	}
}

// CORS 跨域资源共享中间件
// 处理跨域请求
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
