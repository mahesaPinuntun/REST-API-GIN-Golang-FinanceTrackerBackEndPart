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
	userEmail, _ := c.Get("userEmail")

	var trx models.Transaction
	if err := c.ShouldBindJSON(&trx); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Force transaction to belong to logged-in user
	trx.UserID = userID
	if email, ok := userEmail.(string); ok && email != "" {
		trx.UserEmail = email
	}

	if err := config.DB.Create(&trx).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, trx)
}

func GetTransactions(c *gin.Context) {
	userID := c.MustGet("userID").(uint)
	userEmail, _ := c.Get("userEmail")

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

	// Build base query
	baseQuery := config.DB.Model(&models.Transaction{})
	if email, ok := userEmail.(string); ok && email != "" {
		baseQuery = baseQuery.Where("user_id = ? AND user_email = ?", userID, email)
	} else {
		baseQuery = baseQuery.Where("user_id = ?", userID)
	}

	// Count total
	var total int64
	baseQuery.Count(&total)

	// Fetch paginated transactions — newest first
	var transactions []models.Transaction
	if err := baseQuery.
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

func GetTransaction(c *gin.Context) {
	userID := c.MustGet("userID").(uint)
	userEmail, _ := c.Get("userEmail")
	id := c.Param("id")

	var trx models.Transaction

	query := config.DB.Where("id = ? AND user_id = ?", id, userID)
	if email, ok := userEmail.(string); ok && email != "" {
		query = config.DB.Where("id = ? AND user_id = ? AND user_email = ?", id, userID, email)
	}

	if err := query.First(&trx).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "transaction not found"})
		return
	}

	c.JSON(http.StatusOK, trx)
}

func UpdateTransaction(c *gin.Context) {
	userID := c.MustGet("userID").(uint)
	userEmail, _ := c.Get("userEmail")
	id := c.Param("id")

	// Find transaction first — verify ownership
	var trx models.Transaction
	query := config.DB.Where("id = ? AND user_id = ?", id, userID)
	if email, ok := userEmail.(string); ok && email != "" {
		query = config.DB.Where("id = ? AND user_id = ? AND user_email = ?", id, userID, email)
	}

	if err := query.First(&trx).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "transaction not found"})
		return
	}

	// Bind only fields that are sent
	var req struct {
		Amount      *float64 `json:"amount"`
		Category    string   `json:"category"`
		Description string   `json:"description"`
		Currency    string   `json:"currency"`
		Asset       string   `json:"asset"`
		Type        string   `json:"type"`
		Status      string   `json:"status"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Only update fields that are provided
	if req.Amount != nil {
		trx.Amount = *req.Amount
	}
	if req.Category != "" {
		trx.Category = req.Category
	}
	if req.Description != "" {
		trx.Description = req.Description
	}
	if req.Currency != "" {
		trx.Currency = req.Currency
	}
	if req.Asset != "" {
		trx.Asset = req.Asset
	}
	if req.Type != "" {
		trx.Type = req.Type
	}
	if req.Status != "" {
		trx.Status = req.Status
	}

	if err := config.DB.Save(&trx).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, trx)
}

func DeleteTransaction(c *gin.Context) {
	userID := c.MustGet("userID").(uint)
	userEmail, _ := c.Get("userEmail")
	id := c.Param("id")

	// Find transaction first — verify ownership
	var trx models.Transaction
	query := config.DB.Where("id = ? AND user_id = ?", id, userID)
	if email, ok := userEmail.(string); ok && email != "" {
		query = config.DB.Where("id = ? AND user_id = ? AND user_email = ?", id, userID, email)
	}

	if err := query.First(&trx).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "transaction not found"})
		return
	}

	if err := config.DB.Delete(&trx).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "transaction deleted successfully",
		"id":      id,
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