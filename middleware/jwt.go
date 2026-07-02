package middleware

import (
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		// Get Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(401, gin.H{
				"error": "Authorization header is required",
			})
			return
		}

		// Check Bearer format
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(401, gin.H{
				"error": "Invalid authorization format",
			})
			return
		}

		// Remove "Bearer "
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Parse JWT
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {

			// Ensure signing method is HMAC
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrTokenSignatureInvalid
			}

			return []byte(os.Getenv("JWT_SECRET")), nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(401, gin.H{
				"error": "Invalid or expired login token",
			})
			return
		}

		// Extract claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(401, gin.H{
				"error": "Invalid token claims",
			})
			return
		}

		// Read claims safely
		userIDFloat, ok := claims["userId"].(float64)
		if !ok {
			c.AbortWithStatusJSON(401, gin.H{
				"error": "Invalid userId claim",
			})
			return
		}

		userEmail, ok := claims["userEmail"].(string)
		if !ok {
			c.AbortWithStatusJSON(401, gin.H{
				"error": "Invalid userEmail claim",
			})
			return
		}

		/*userName, ok := claims["userName"].(string)
		if !ok {
			c.AbortWithStatusJSON(401, gin.H{
				"error": "Invalid userName claim",
			})
			return
		}*/

		// Store into Gin context
		c.Set("userID", uint(userIDFloat))
		c.Set("userEmail", userEmail)
		//c.Set("userName", userName)
		c.Set("userToken",tokenString)
		c.Next()
	}
}