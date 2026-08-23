package utils

import (
	"crypto/sha256"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// hashEmail creates a SHA256 hash of the email
// used as an extra claim in JWT for additional verification
func hashEmail(email string) string {
	h := sha256.New()
	h.Write([]byte(email))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func GenerateToken(userID uint, email string) (string, error) {

	secretKey := []byte(os.Getenv("JWT_SECRET"))

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		jwt.MapClaims{
			"user_id":    userID,
			"email_hash": hashEmail(email), // SHA256 hash of email
			"exp": time.Now().
				Add(time.Hour * 24).
				Unix(),
		},
	)

	return token.SignedString(secretKey)
}

// VerifyEmailHash checks if the email matches the hash stored in JWT
func VerifyEmailHash(email, hash string) bool {
	return hashEmail(email) == hash
}