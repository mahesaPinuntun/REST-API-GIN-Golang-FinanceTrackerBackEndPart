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
		Name     string `json:"name"`
		Email    string `json:"email" gorm:"unique"`
		Password string `json:"-"`
		SalaryAmmount  float64 `json:"salaryAmmount"`
		SalaryCurrency string  `json:"salaryCurrency"`
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
		Name:     req.Name,
		Email:    req.Email,
		Password: hash,
	}

	if err := config.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError,
			gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated,
		gin.H{
			"message": "User created",
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

	c.JSON(200, gin.H{
		"token": token,
	})
}
