package controllers

import (
	"net/http"

	"finance-tracker/config"
	"finance-tracker/models"
	"finance-tracker/utils"

	"github.com/gin-gonic/gin"
)

func Register(c *gin.Context) {

	var req struct {
		Name            string  `json:"name"`
		Email           string  `json:"email"`
		Password        string  `json:"password"`
		SalaryAmount    float64 `json:"salaryAmount"`
		SalaryCurrency  string  `json:"salaryCurrency"`
		SalaryFrequency string  `json:"salaryFrequency"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate required fields
	if req.Name == "" || req.Email == "" || req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name, email and password are required"})
		return
	}

	if len(req.Password) < 8 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "password must be at least 8 characters"})
		return
	}

	// Check if email already exists
	var existing models.User
	if err := config.DB.Where("email = ?", req.Email).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "email already registered"})
		return
	}

	hash, err := utils.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	// Set default currency if not provided
	if req.SalaryCurrency == "" {
		req.SalaryCurrency = "IDR"
	}
	if req.SalaryFrequency == "" {
		req.SalaryFrequency = "monthly"
	}

	user := models.User{
		Name:             req.Name,
		Email:            req.Email,
		Password:         hash,
		SalaryAmount:     req.SalaryAmount,
		SalaryCurrency:   req.SalaryCurrency,
		SalaryFrequency:  req.SalaryFrequency,
		IsEmailConfirmed: false,
	}

	if err := config.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Auto-send confirmation email — errors are silent so registration still succeeds
	token, err := generateToken()
	if err == nil {
		emailToken := models.EmailToken{
			UserEmail: user.Email,
			Token:     token,
			ExpiresAt: timeNowPlusHours(24),
		}
		if err := config.DB.Create(&emailToken).Error; err == nil {
			go sendConfirmationEmail(user.Email, user.Name, token) // run in goroutine so it doesn't block response
		}
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":            "User created. Please check your email to confirm your account.",
		"is_email_confirmed": false,
		"user": gin.H{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
		},
	})
}

func Login(c *gin.Context) {

	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Email == "" || req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email and password are required"})
		return
	}

	var user models.User
	if err := config.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}

	if !utils.CheckPasswordHash(req.Password, user.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credential"})
		return
	}

	token, err := utils.GenerateToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	response := gin.H{
		"token":              token,
		"is_email_confirmed": user.IsEmailConfirmed,
		"user": gin.H{
			"userId":    user.ID,
			"userName":  user.Name,
			"userEmail": user.Email,
		},
	}

	if !user.IsEmailConfirmed {
		response["warning"] = "Your email is not confirmed. Please check your email."
	}

	c.JSON(http.StatusOK, response)
}