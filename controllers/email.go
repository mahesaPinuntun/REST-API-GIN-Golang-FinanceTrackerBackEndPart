package controllers

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"finance-tracker/config"
	"finance-tracker/models"

	"github.com/gin-gonic/gin"
)

// generateToken creates a secure random 32-byte hex token
func generateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// sendConfirmationEmail sends an email via Resend API
func sendConfirmationEmail(toEmail, toName, token string) error {
	appURL := os.Getenv("APP_URL") // e.g. https://your-app.vercel.app
	confirmURL := fmt.Sprintf("%s/api/auth/confirm?token=%s", appURL, token)

	payload := map[string]interface{}{
		"from":    "Finance Tracker <noreply@yourdomain.com>",
		"to":      []string{toEmail},
		"subject": "Confirm your email address",
		"html": fmt.Sprintf(`
			<h2>Hi %s,</h2>
			<p>Thanks for signing up! Please confirm your email address by clicking the button below.</p>
			<p>This link expires in <strong>24 hours</strong>.</p>
			<a href="%s" style="
				display:inline-block;
				padding:12px 24px;
				background:#000;
				color:#fff;
				text-decoration:none;
				border-radius:6px;
				font-weight:bold;
			">Confirm Email</a>
			<p>Or copy this link: %s</p>
			<p>If you did not create an account, you can ignore this email.</p>
		`, toName, confirmURL, confirmURL),
	}

	body, _ := json.Marshal(payload)

	req, err := http.NewRequest(
		"POST",
		"https://api.resend.com/emails",
		bytes.NewBuffer(body),
	)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+os.Getenv("RESEND_API_KEY"))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("resend API error: status %d", resp.StatusCode)
	}

	return nil
}

// SendConfirmationEmail godoc
// POST /api/auth/send-confirmation
// Generates a token and sends confirmation email to logged-in user
func SendConfirmationEmail(c *gin.Context) {
	userID := c.MustGet("userID").(uint)

	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	if user.IsEmailConfirmed {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email already confirmed"})
		return
	}

	// Delete any existing unused tokens for this user
	config.DB.Where("user_id = ?", userID).Delete(&models.EmailToken{})

	// Generate new token
	token, err := generateToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	// Save token to DB with 24h expiry
	emailToken := models.EmailToken{
		UserID:    userID,
		Token:     token,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	if err := config.DB.Create(&emailToken).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save token"})
		return
	}

	// Send email
	if err := sendConfirmationEmail(user.Email, user.Name, token); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to send email: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Confirmation email sent to " + user.Email,
	})
}

// ConfirmEmail godoc
// GET /api/auth/confirm?token=xxx
// Validates token and sets is_email_confirmed = true
func ConfirmEmail(c *gin.Context) {
	token := c.Query("token")

	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "token is required"})
		return
	}

	// Find token in DB
	var emailToken models.EmailToken
	if err := config.DB.Where("token = ?", token).First(&emailToken).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid or expired token"})
		return
	}

	// Check expiry
	if time.Now().After(emailToken.ExpiresAt) {
		config.DB.Delete(&emailToken)
		c.JSON(http.StatusBadRequest, gin.H{"error": "token has expired, please request a new one"})
		return
	}

	// Set is_email_confirmed = true on user
	if err := config.DB.Model(&models.User{}).
		Where("id = ?", emailToken.UserID).
		Update("is_email_confirmed", true).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to confirm email"})
		return
	}

	// Delete used token
	config.DB.Delete(&emailToken)

	c.JSON(http.StatusOK, gin.H{
		"message": "Email confirmed successfully",
	})
}