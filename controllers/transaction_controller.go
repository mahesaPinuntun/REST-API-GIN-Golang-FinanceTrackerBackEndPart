package controllers
import (
	"net/http"
	"finance-tracker/config"
	"finance-tracker/models"
	"github.com/gin-gonic/gin"
)
func CreateTransaction(c *gin.Context) {
	userID := c.MustGet("userID").(uint)
	userEmail := c.MustGet("userEmail").(string)
	userToken := c.MustGet("userToken").(string)

	var trx models.Transaction

	if err := c.ShouldBindJSON(&trx); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Request Body is not satisfied",
		})
		return
	}

	// Check user exists
	var user models.User
	if err := config.DB.Where("email = ?", userEmail).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "User not found",
		})
		return
	}

	// Check session exists for this email + token
	var session models.Sessions
	if err := config.DB.
		Where("email = ? AND token = ?", userEmail, userToken).
		First(&session).Error; err != nil {

		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}

	switch trx.Type {
	case "income", "expense":
		//if input negative amount, return error
		if trx.Amount < 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"Bad request": "Amount cannot be assigned by negative value",
			})
			return
		}
	}
	// Assign ownership
	trx.UserID = userID
	trx.UserEmail = userEmail

	// Save transaction
	if err := config.DB.Create(&trx).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create transaction",
		})
		return
	}

	c.JSON(http.StatusCreated, trx)
}
func GetTransactions(c *gin.Context) {
	userID := c.MustGet("userID").(uint)
	userEmail := c.MustGet("userEmail").(string)
	userToken := c.MustGet("userToken").(string)

	// Verify session still exists
	var session models.Sessions
	if err := config.DB.
		Where("email = ? AND token = ?", userEmail, userToken).
		First(&session).Error; err != nil {

		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}

	var transactions []models.Transaction

	// Fetch only this user's transactions
	if err := config.DB.
		Where("user_id = ? AND user_email = ?", userID, userEmail).
		Find(&transactions).Error; err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve transactions",
		})
		return
	}
	var session models.Sessions
	if err := config.DB.
		Where("email = ? AND token = ?", userEmail, userToken).
		First(&session).Error; err != nil {

		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}
	
	c.JSON(http.StatusOK, transactions)
}
func Dashboard(c *gin.Context) {
	c.JSON(200, "call success unfortunately this feature is under construction")
	/*
	userID := c.MustGet("userID").(uint)
	var income float64
	var expense float64
	config.DB.
		Model(&models.Transaction{}).
		Where("userEmail = ? AND type = ?", userID, "income").
		Select("COALESCE(SUM(amount),0)").
		Scan(&income)
	config.DB.
		Model(&models.Transaction{}).
		Where("userEmail = ? AND type = ?", userID, "expense").
		Select("COALESCE(SUM(amount),0)").
		Scan(&expense)
	c.JSON(200, gin.H{
		"income":  income,
		"expense": expense,
		"balance": income - expense,
	})*/

}