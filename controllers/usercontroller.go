package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"finance-tracker/config"
	"finance-tracker/models"
	"finance-tracker/utils"

	"github.com/gin-gonic/gin"
)

const updateEmailTemplate = `<!DOCTYPE html>
<html>
<body style="font-family:-apple-system,sans-serif;background:#f4f4f4;padding:40px 0;">
  <div style="max-width:480px;margin:0 auto;background:#fff;border-radius:12px;padding:40px;">
    <h2 style="margin:0 0 8px;">Confirm your account changes</h2>
    <p style="color:#666;margin:0 0 8px;">Hi %s,</p>
    <p style="color:#666;margin:0 0 24px;">
      You requested the following changes to your account:
    </p>
    <div style="background:#f9f9f9;border-radius:8px;padding:16px;margin-bottom:24px;">
      %s
    </div>
    <p style="color:#666;margin:0 0 24px;">
      Click the button below to confirm these changes.
      If you did not request this, you can safely ignore this email.
    </p>
    <a href="%s" style="
      display:inline-block;padding:12px 28px;
      background:#000;color:#fff;
      text-decoration:none;border-radius:8px;
      font-weight:600;font-size:15px;
    ">Confirm Changes</a>
    <p style="color:#999;font-size:13px;margin:24px 0 0;">
      This link expires in <strong>1 hour</strong>.<br>
    </p>
    <hr style="border:none;border-top:1px solid #eee;margin:24px 0;">
    <p style="color:#bbb;font-size:12px;margin:0;">
      Or copy this link: %s
    </p>
  </div>
</body>
</html>`

// sendUpdateConfirmationEmail sends confirmation email for profile changes
func sendUpdateConfirmationEmail(toEmail, toName, changesSummary, token string) error {
	appURL := os.Getenv("APP_URL")
	confirmURL := fmt.Sprintf("%s/api/user/confirm-update?token=%s", appURL, token)

	brevoFrom := os.Getenv("BREVO_FROM")
	brevoKey := os.Getenv("BREVO_API_KEY")

	if brevoFrom == "" {
		return fmt.Errorf("BREVO_FROM env var is not set")
	}
	if brevoKey == "" {
		return fmt.Errorf("BREVO_API_KEY env var is not set")
	}

	payload := map[string]interface{}{
		"sender": map[string]string{
			"name":  "Finance Tracker",
			"email": brevoFrom,
		},
		"to": []map[string]string{
			{
				"email": toEmail,
				"name":  toName,
			},
		},
		"subject":     "Confirm your account changes",
		"htmlContent": fmt.Sprintf(updateEmailTemplate, toName, changesSummary, confirmURL, confirmURL),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", "https://api.brevo.com/v3/smtp/email", bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("accept", "application/json")
	req.Header.Set("api-key", brevoKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("brevo API error: status %d", resp.StatusCode)
	}

	return nil
}

// GetProfile godoc
// GET /api/user/profile
// Returns current user profile
func GetProfile(c *gin.Context) {
	userID := c.MustGet("userID").(uint)

	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":               user.ID,
		"name":             user.Name,
		"email":            user.Email,
		"salaryAmount":     user.SalaryAmount,
		"salaryCurrency":   user.SalaryCurrency,
		"salaryFrequency":  user.SalaryFrequency,
		"isEmailConfirmed": user.IsEmailConfirmed,
		"createdAt":        user.CreatedAt,
	})
}

// RequestUpdateProfile godoc
// PUT /api/user/profile
// Stores pending changes and sends confirmation email
func RequestUpdateProfile(c *gin.Context) {
	userID := c.MustGet("userID").(uint)
	userEmail, _ := c.Get("userEmail")

	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	var req struct {
		Name            string  `json:"name"`
		Email           string  `json:"email"`
		SalaryAmount    float64 `json:"salaryAmount"`
		SalaryCurrency  string  `json:"salaryCurrency"`
		SalaryFrequency string  `json:"salaryFrequency"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if new email already taken by another user
	if req.Email != "" && req.Email != user.Email {
		var existing models.User
		if err := config.DB.Where("email = ?", req.Email).First(&existing).Error; err == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "email already in use"})
			return
		}
	}

	// Build summary of changes for email
	changesSummary := ""
	hasChanges := false

	if req.Name != "" && req.Name != user.Name {
		changesSummary += fmt.Sprintf("<p>• <strong>Name:</strong> %s → %s</p>", user.Name, req.Name)
		hasChanges = true
	}
	if req.Email != "" && req.Email != user.Email {
		changesSummary += fmt.Sprintf("<p>• <strong>Email:</strong> %s → %s</p>", user.Email, req.Email)
		hasChanges = true
	}
	if req.SalaryAmount > 0 && req.SalaryAmount != user.SalaryAmount {
		changesSummary += fmt.Sprintf("<p>• <strong>Salary Amount:</strong> %.2f → %.2f</p>", user.SalaryAmount, req.SalaryAmount)
		hasChanges = true
	}
	if req.SalaryCurrency != "" && req.SalaryCurrency != user.SalaryCurrency {
		changesSummary += fmt.Sprintf("<p>• <strong>Salary Currency:</strong> %s → %s</p>", user.SalaryCurrency, req.SalaryCurrency)
		hasChanges = true
	}
	if req.SalaryFrequency != "" && req.SalaryFrequency != user.SalaryFrequency {
		changesSummary += fmt.Sprintf("<p>• <strong>Salary Frequency:</strong> %s → %s</p>", user.SalaryFrequency, req.SalaryFrequency)
		hasChanges = true
	}

	if !hasChanges {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no changes detected"})
		return
	}

	// Delete any existing pending updates for this user
	config.DB.Where("user_id = ?", userID).Delete(&models.PendingUpdate{})

	// Generate token
	token, err := generateToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	// Determine email to send confirmation to
	emailTo := user.Email
	if email, ok := userEmail.(string); ok && email != "" {
		emailTo = email
	}

	// Save pending update
	pending := models.PendingUpdate{
		UserID:             userID,
		UserEmail:          emailTo,
		Token:              token,
		ExpiresAt:          time.Now().Add(1 * time.Hour),
		NewName:            req.Name,
		NewEmail:           req.Email,
		NewSalaryAmount:    req.SalaryAmount,
		NewSalaryCurrency:  req.SalaryCurrency,
		NewSalaryFrequency: req.SalaryFrequency,
	}

	if err := config.DB.Create(&pending).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save pending update"})
		return
	}

	// Send confirmation email
	if err := sendUpdateConfirmationEmail(emailTo, user.Name, changesSummary, token); err != nil {
		config.DB.Delete(&pending)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to send confirmation email: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Confirmation email sent. Please check your email to apply the changes.",
		"changes": changesSummary,
	})
}

// ConfirmUpdateProfile godoc
// GET /api/user/confirm-update?token=xxx
// Applies pending profile changes after email confirmation
func ConfirmUpdateProfile(c *gin.Context) {
	token := c.Query("token")

	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "token is required"})
		return
	}

	// Find pending update
	var pending models.PendingUpdate
	if err := config.DB.Where("token = ?", token).First(&pending).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid or expired token"})
		return
	}

	// Check expiry
	if time.Now().After(pending.ExpiresAt) {
		config.DB.Delete(&pending)
		c.JSON(http.StatusBadRequest, gin.H{"error": "token has expired, please request changes again"})
		return
	}

	// Find user
	var user models.User
	if err := config.DB.First(&user, pending.UserID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	// Apply only non-empty changes
	if pending.NewName != "" {
		user.Name = pending.NewName
	}
	if pending.NewEmail != "" {
		user.Email = pending.NewEmail
		user.IsEmailConfirmed = false // require re-confirmation if email changes
	}
	if pending.NewSalaryAmount > 0 {
		user.SalaryAmount = pending.NewSalaryAmount
	}
	if pending.NewSalaryCurrency != "" {
		user.SalaryCurrency = pending.NewSalaryCurrency
	}
	if pending.NewSalaryFrequency != "" {
		user.SalaryFrequency = pending.NewSalaryFrequency
	}

	if err := config.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to apply changes"})
		return
	}

	// Delete used token
	config.DB.Delete(&pending)

	// If email changed, send new confirmation email
	emailNote := ""
	if pending.NewEmail != "" && pending.NewEmail != pending.UserEmail {
		emailNote = " Your email has changed — please confirm your new email address."

		newToken, err := generateToken()
		if err == nil {
			newEmailToken := models.EmailToken{
				UserEmail: user.Email,
				Token:     newToken,
				ExpiresAt: timeNowPlusHours(24),
			}
			if err := config.DB.Create(&newEmailToken).Error; err == nil {
				go sendConfirmationEmail(user.Email, user.Name, newToken)
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Account updated successfully." + emailNote,
		"user": gin.H{
			"id":              user.ID,
			"name":            user.Name,
			"email":           user.Email,
			"isEmailConfirmed": user.IsEmailConfirmed,
		},
	})
}

// ChangePassword godoc
// PUT /api/user/change-password
// Changes password — requires current password verification
func ChangePassword(c *gin.Context) {
	userID := c.MustGet("userID").(uint)

	var req struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.CurrentPassword == "" || req.NewPassword == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "current_password and new_password are required"})
		return
	}

	if len(req.NewPassword) < 8 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "new password must be at least 8 characters"})
		return
	}

	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	if !utils.CheckPasswordHash(req.CurrentPassword, user.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "current password is incorrect"})
		return
	}

	hash, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	if err := config.DB.Model(&user).Update("password", hash).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Password changed successfully. Please login again.",
	})
}