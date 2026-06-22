package controllers

import (
	"net/http"

	"finance-tracker/config"
	"finance-tracker/models"

	"github.com/gin-gonic/gin"
)

func CreateTransaction(c *gin.Context) {

	var trx models.Transaction

	if err := c.ShouldBindJSON(&trx); err != nil {

		c.JSON(http.StatusBadRequest,
			gin.H{"error": err.Error()})

		return
	}

	config.DB.Create(&trx)

	c.JSON(http.StatusCreated, trx)
}
func GetTransactions(c *gin.Context) {

	var transactions []models.Transaction

	config.DB.Find(&transactions)

	c.JSON(200, transactions)
}
func Dashboard(c *gin.Context) {

	var income float64
	var expense float64

	config.DB.
		Model(&models.Transaction{}).
		Where("type = ?", "income").
		Select("COALESCE(SUM(amount),0)").
		Scan(&income)

	config.DB.
		Model(&models.Transaction{}).
		Where("type = ?", "expense").
		Select("COALESCE(SUM(amount),0)").
		Scan(&expense)

	c.JSON(200, gin.H{
		"income":  income,
		"expense": expense,
		"balance": income - expense,
	})
}
