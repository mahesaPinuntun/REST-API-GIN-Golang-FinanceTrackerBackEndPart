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
		c.JSON(http.StatusBadRequest,
			gin.H{"error": err.Error()})
		return
	}

	hash, err := utils.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			gin.H{"error": "Failed to hash password"})
		return
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
		c.JSON(http.StatusInternalServerError,
			gin.H{"error": err.Error()})
		return
	}

	// Auto-send confirmation email after registration
	token, err := generateToken()
	if err == nil {
		emailToken := models.EmailToken{
			UserEmail: user.Email,
			Token:     token,
			ExpiresAt: timeNowPlusHours(24),
		}
		if err := config.DB.Create(&emailToken).Error; err == nil {
			sendConfirmationEmail(user.Email, user.Name, token)
		}
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":            "User created. Please check your email to confirm your account.",
		"is_email_confirmed": false,
	})
}

func Login(c *gin.Context) {

	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}


	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	var user models.User

	result := config.DB.
		Where("email = ?", req.Email).
		First(&user)

	if result.Error != nil {
		c.JSON(401, gin.H{"error": "User not found"})
		return
	}

	if !utils.CheckPasswordHash(req.Password, user.Password) {
		c.JSON(401, gin.H{"error": "Invalid credential"})
		return
	}

	token, err := utils.GenerateToken(user.ID)
	if err != nil {
		c.JSON(500, gin.H{"error": "Token error"})
		return
	}

	
/*
	if !user.IsEmailConfirmed {
		response["warning"] = "Your email is unverified. open through registered email to click verify button."
	}*/
	//note login session in db
	//UserEmail string    `json:"user_email"`
	//Token     string    `json:"token" gorm:"unique"`
	//ExpiresAt time.Time `json:"expires_at"`
	session := models.Sessions{
		UserEmail:          req.Email,
		Token:            token,
		ExpiresAt:        timeNowPlusHours(240), // Set session expiration to 7 days
	
	}
	config.DB.Create(&session)
         

	response := gin.H{
		"token":              token,
		"is_email_confirmed": user.IsEmailConfirmed,
		"user": gin.H{
			"userId":    user.ID,
			"userName":  user.Name,
			"userEmail": user.Email,
			"userToken": token,
		},
	}   
	c.JSON(200, response)
}