package controllers

import (
	"net/http"

	"finance-tracker/config"
	"finance-tracker/models"

	"github.com/gin-gonic/gin"
)

func CreateTransaction(c *gin.Context) {

	userID := c.MustGet("userID").(uint)

	var trx models.Transaction

	if err := c.ShouldBindJSON(&trx); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Force the transaction to belong to the logged-in user
	trx.UserID = userID

	config.DB.Create(&trx)

	c.JSON(http.StatusCreated, trx)
}

func GetTransactions(c *gin.Context) {

	userID := c.MustGet("userID").(uint)

	var transactions []models.Transaction

	// Only fetch transactions belonging to the logged-in user
	config.DB.Where("user_id = ?", userID).Find(&transactions)

	c.JSON(200, transactions)
}

func Dashboard(c *gin.Context) {

	userID := c.MustGet("userID").(uint)

	var income float64
	var expense float64

	config.DB.
		Model(&models.Transaction{}).
		Where("user_id = ? AND type = ?", userID, "income").
		Select("COALESCE(SUM(amount),0)").
		Scan(&income)

	config.DB.
		Model(&models.Transaction{}).
		Where("user_id = ? AND type = ?", userID, "expense").
		Select("COALESCE(SUM(amount),0)").
		Scan(&expense)

	c.JSON(200, gin.H{
		"income":  income,
		"expense": expense,
		"balance": income - expense,
	})
}