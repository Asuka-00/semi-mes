package middleware

import (
	"semi-mes/server/internal/config"
	"semi-mes/server/pkg/utils"
	"strings"

	"github.com/gin-gonic/gin"
)

func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.Error(c, 40001, "error.auth.missing_token")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			utils.Error(c, 40001, "error.auth.invalid_token_format")
			c.Abort()
			return
		}

		tokenString := parts[1]
		claims, err := utils.ParseToken(tokenString, config.GlobalConfig.JWT.Secret)
		if err != nil {
			utils.Error(c, 40001, "error.auth.invalid_token")
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Next()
	}
}
