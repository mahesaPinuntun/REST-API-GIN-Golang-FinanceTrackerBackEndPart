package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware() gin.HandlerFunc {

	return func(c *gin.Context) {

		authHeader := c.GetHeader("Authorization")

		if authHeader == "" {
			c.AbortWithStatusJSON(
				401,
				gin.H{
					"error": "token required",
				},
			)
			return
		}

		tokenString := strings.TrimPrefix(
			authHeader,
			"Bearer ",
		)

		token, err := jwt.Parse(
			tokenString,
			func(token *jwt.Token) (interface{}, error) {
				return []byte("secret"), nil
			},
		)

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(
				401,
				gin.H{"error": "invalid token"},
			)
			return
		}

		c.Next()
	}
}
