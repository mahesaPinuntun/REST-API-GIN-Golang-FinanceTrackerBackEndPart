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

// timeNowPlusHours is a helper used across controllers
func timeNowPlusHours(h time.Duration) time.Time {
	return time.Now().Add(h * time.Hour)
}

// generateToken creates a secure random 32-byte hex token
func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// sendConfirmationEmail sends an email via Resend API
func sendConfirmationEmail(toEmail, toName, token string) error {
	appURL := os.Getenv("APP_URL")
	confirmURL := fmt.Sprintf("%s/api/auth/confirm?token=%s", appURL, token)

	payload := map[string]interface{}{
		"from":    "Finance Tracker <onboarding@resend.dev>",
		"to":      []string{toEmail},
		"subject": "Confirm your email address",
		"html": fmt.Sprintf(`
<!DOCTYPE html>
<html>
<body style="font-family:-apple-system,sans-serif;background:#f4f4f4;padding:40px 0;">
  <div style="max-width:480px;margin:0 auto;background:#fff;border-radius:12px;padding:40px;">
    <h2 style="margin:0 0 8px;">Hi %s 👋</h2>
    <p style="color:#666;margin:0 0 24px;">
      Thanks for signing up for Finance Tracker!
      Please confirm your email address to unlock all features.
    </p>
    <a href="%s" style="
      display:inline-block;padding:12px 28px;
      background:#000;color:#fff;
      text-decoration:none;border-radius:8px;
      font-weight:600;font-size:15px;
    ">Confirm Email</a>
    <p style="color:#999;font-size:13px;margin:24px 0 0;">
      This link expires in <strong>24 hours</strong>.<br>
      If you didn't create an account, you can safely ignore this email.
    </p>
    <hr style="border:none;border-top:1px solid #eee;margin:24px 0;">
    <p style="color:#bbb;font-size:12px;margin:0;">
      Or copy this link: %s
    </p>
  </div>
</body>
</html>
		`, toName, confirmURL, confirmURL),
	}

	body, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", "https://api.resend.com/emails", bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+os.Getenv("RESEND_API_KEY"))

	client := &http.Client{Timeout: 10 * time.Second}
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
// Resends confirmation email to logged-in user
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
	config.DB.Where("user_email = ?", user.Email).Delete(&models.EmailToken{})

	token, err := generateToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	emailToken := models.EmailToken{
		UserEmail: user.Email,
		Token:     token,
		ExpiresAt: timeNowPlusHours(24),
	}

	if err := config.DB.Create(&emailToken).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save token"})
		return
	}

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
func ConfirmEmail(c *gin.Context) {
	token := c.Query("token")

	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "token is required"})
		return
	}

	var emailToken models.EmailToken
	if err := config.DB.Where("token = ?", token).First(&emailToken).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid or expired token"})
		return
	}

	if time.Now().After(emailToken.ExpiresAt) {
		config.DB.Delete(&emailToken)
		c.JSON(http.StatusBadRequest, gin.H{"error": "token has expired, please request a new one"})
		return
	}

	// Update using email instead of ID
	if err := config.DB.Model(&models.User{}).
		Where("email = ?", emailToken.UserEmail).
		Update("is_email_confirmed", true).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to confirm email"})
		return
	}

	config.DB.Delete(&emailToken)

	c.JSON(http.StatusOK, gin.H{
		"message": "Email confirmed successfully. You now have full access.",
	})
}