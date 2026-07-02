package utils

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func GenerateToken(userID uint, userEmail string ) (string, error) {

	secretKey := []byte(os.Getenv("JWT_SECRET"))

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		jwt.MapClaims{
			"userId": userID,
			"exp": time.Now().
				Add(time.Hour * 24).
				Unix(),
			"userEmail": userEmail,
		},
	)

	return token.SignedString(secretKey)
}