package controllers

import (
	"net/http"
	"strconv"

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

	// Force transaction to belong to logged-in user
	trx.UserID = userID

	if err := config.DB.Create(&trx).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, trx)
}

func GetTransactions(c *gin.Context) {
	userID := c.MustGet("userID").(uint)

	// Pagination params
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if err != nil || limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit

	// Count total
	var total int64
	config.DB.Model(&models.Transaction{}).
		Where("user_id = ?", userID).
		Count(&total)

	// Fetch paginated transactions — newest first
	var transactions []models.Transaction
	if err := config.DB.
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&transactions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve transactions"})
		return
	}

	totalPages := int(total) / limit
	if int(total)%limit != 0 {
		totalPages++
	}

	c.JSON(http.StatusOK, gin.H{
		"data":        transactions,
		"page":        page,
		"limit":       limit,
		"total":       total,
		"total_pages": totalPages,
		"has_next":    page < totalPages,
		"has_prev":    page > 1,
	})
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

	c.JSON(http.StatusOK, gin.H{
		"income":  income,
		"expense": expense,
		"balance": income - expense,
	})
}