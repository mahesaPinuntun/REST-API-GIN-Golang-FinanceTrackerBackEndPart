package middleware

import (
	"os"
	"strings"

	"finance-tracker/config"
	"finance-tracker/models"
	"finance-tracker/utils"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		authHeader := c.GetHeader("Authorization")

		if authHeader == "" {
			c.AbortWithStatusJSON(401, gin.H{"error": "Authorization header is required"})
			return
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(401, gin.H{"error": "Invalid authorization format"})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrTokenSignatureInvalid
			}
			return []byte(os.Getenv("JWT_SECRET")), nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(401, gin.H{"error": "Invalid or expired token"})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(401, gin.H{"error": "Invalid token claims"})
			return
		}

		// Read user_id
		userIDFloat, ok := claims["user_id"].(float64)
		if !ok {
			c.AbortWithStatusJSON(401, gin.H{"error": "Invalid token: missing user_id"})
			return
		}

		userID := uint(userIDFloat)

		// Read email_hash
		emailHash, hasEmailHash := claims["email_hash"].(string)

		// If token has email_hash, verify it against DB
		if hasEmailHash && emailHash != "" {
			var user models.User
			if err := config.DB.First(&user, userID).Error; err != nil {
				c.AbortWithStatusJSON(401, gin.H{"error": "User not found"})
				return
			}

			if !utils.VerifyEmailHash(user.Email, emailHash) {
				c.AbortWithStatusJSON(401, gin.H{"error": "Invalid token: email mismatch"})
				return
			}

			c.Set("userEmail", user.Email)
		}

		c.Set("userID", userID)
		c.Next()
	}
}