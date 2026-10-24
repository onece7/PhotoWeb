package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	jwtv5 "github.com/golang-jwt/jwt/v5"

	"novel2comic/backend/internal/service"
	"novel2comic/backend/pkg/jwt"
	"novel2comic/backend/pkg/response"
)

// AuthRequired 认证中间件：验证 JWT Token
func AuthRequired(authService *service.AuthService, jwtManager *jwt.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 提取 Bearer Token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Error(c, http.StatusUnauthorized, response.CodeUnauthorized, "请先登录")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			response.Error(c, http.StatusUnauthorized, response.CodeUnauthorized, "Token 格式错误")
			c.Abort()
			return
		}
		tokenString := parts[1]

		// 检查黑名单
		if authService.IsTokenBlacklisted(c.Request.Context(), tokenString) {
			response.Error(c, http.StatusUnauthorized, response.CodeTokenExpired, "Token 已失效")
			c.Abort()
			return
		}

		// 解析 Token
		claims, err := jwtManager.ParseAccessToken(tokenString)
		if err != nil {
			if errors.Is(err, jwtv5.ErrTokenExpired) {
				response.Error(c, http.StatusUnauthorized, response.CodeTokenExpired, "Token 已过期")
			} else {
				response.Error(c, http.StatusUnauthorized, response.CodeUnauthorized, "Token 无效")
			}
			c.Abort()
			return
		}

		// 将用户信息存入 context
		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		c.Next()
	}
}
