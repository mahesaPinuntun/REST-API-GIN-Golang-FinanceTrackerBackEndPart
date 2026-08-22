package utils

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func GenerateToken(userID uint) (string, error) {

	secretKey := []byte(os.Getenv("JWT_SECRET"))

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		jwt.MapClaims{
			"user_id": userID,
			"exp": time.Now().
				Add(time.Hour * 24).
				Unix(),
		},
	)

	return token.SignedString(secretKey)
}