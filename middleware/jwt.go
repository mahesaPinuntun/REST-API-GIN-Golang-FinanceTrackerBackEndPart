func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		authHeader := c.GetHeader("Authorization")

		if authHeader == "" {
			c.AbortWithStatusJSON(401, gin.H{
				"error": "Authorization header is required",
			})
			return
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(401, gin.H{
				"error": "Invalid authorization format",
			})
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
			c.AbortWithStatusJSON(401, gin.H{
				"error": "Invalid or expired login token",
			})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(401, gin.H{
				"error": "Invalid token claims",
			})
			return
		}

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

		c.Set("userID", uint(userIDFloat))
		c.Set("userEmail", userEmail)
		c.Set("userToken", tokenString)

		c.Next()
	}
}