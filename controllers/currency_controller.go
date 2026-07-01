package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"finance-tracker/config"
	"finance-tracker/models"

	"github.com/gin-gonic/gin"
)

const fxAPIBase = "https://fxapi.app/api" //open source free API for currency exchange rates, no API key required

// fxRate holds the response from fxapi.app single pair endpoint
type fxRate struct {
	Base      string    `json:"base"`
	Target    string    `json:"target"`
	Rate      float64   `json:"rate"`
	Timestamp time.Time `json:"timestamp"`
}

// fxAllRates holds the response from fxapi.app all-rates endpoint
type fxAllRates struct {
	Base      string             `json:"base"`
	Timestamp time.Time          `json:"timestamp"`
	Rates     map[string]float64 `json:"rates"`
}

// getExchangeRate fetches rate between two currencies from fxapi.app
func getExchangeRate(base, target string) (float64, error) {
	base   = strings.ToUpper(base)
	target = strings.ToUpper(target)

	if base == target {
		return 1.0, nil
	}

	url := fmt.Sprintf("%s/%s/%s.json", fxAPIBase, base, target)

	resp, err := http.Get(url)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch exchange rate:","err1") //err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("fxapi returned status ", "API doesnt respon")
	}

	var result fxRate
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("failed to decode exchange rate:","err2" ) //err
	}

	return result.Rate, nil
}

// ConvertCurrency godoc
// GET /api/currency/convert?from=USD&to=IDR&amount=100
// Returns the converted amount using live fxapi.app rates
func ConvertCurrency(c *gin.Context) {
	from   := c.DefaultQuery("from", "USD")
	to     := c.DefaultQuery("to", "USD")
	amount := 1.0

	if a := c.Query("amount"); a != "" {
		fmt.Sscanf(a, "%f", &amount)
	}

	rate, err := getExchangeRate(from, to)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"from":      strings.ToUpper(from),
		"to":        strings.ToUpper(to),
		"rate":      rate,
		"amount":    amount,
		"converted": amount * rate,
	})
}

// GetSupportedCurrencies godoc
// GET /api/currency/supported
// Returns all currencies supported by fxapi.app
func GetSupportedCurrencies(c *gin.Context) {
	resp, err := http.Get(fxAPIBase + "/currencies.json")
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "failed to fetch currencies"})
		return
	}
	defer resp.Body.Close()

	var currencies interface{}
	json.NewDecoder(resp.Body).Decode(&currencies)

	c.JSON(http.StatusOK, currencies)
}

// GetDashboardInCurrency godoc
// GET /api/dashboard/convert?currency=IDR
// Returns the user's income/expense/balance/salary converted to requested currency
func GetDashboardInCurrency(c *gin.Context) {
	userID   := c.MustGet("userID").(uint)
	toCurrency := strings.ToUpper(c.DefaultQuery("currency", "USD"))

	// Fetch user for salary info
	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	// Aggregate POSTED transactions per currency
	type currencySum struct {
		Currency string
		Income   float64
		Expense  float64
	}

	var rows []struct {
		Currency string
		Type     string
		Total    float64
	}

	config.DB.
		Model(&models.Transaction{}).
		Select("currency, type, COALESCE(SUM(amount), 0) as total").
		Where("user_id = ? AND status = 'POSTED'", userID).
		Group("currency, type").
		Scan(&rows)

	// Convert each currency bucket to target currency
	totalIncome  := 0.0
	totalExpense := 0.0

	rateCache := map[string]float64{}

	for _, row := range rows {
		fromCurrency := row.Currency
		if fromCurrency == "" {
			fromCurrency = "USD"
		}

		// Cache rates to avoid duplicate API calls
		cacheKey := fromCurrency + "_" + toCurrency
		rate, ok := rateCache[cacheKey]
		if !ok {
			var err error
			rate, err = getExchangeRate(fromCurrency, toCurrency)
			if err != nil {
				c.JSON(http.StatusBadGateway, gin.H{
					"error": fmt.Sprintf("failed to get rate for %s→%s: %s", fromCurrency, toCurrency, err.Error()),
				})
				return
			}
			rateCache[cacheKey] = rate
		}

		converted := row.Total * rate

		if row.Type == string(models.ActivityDeposit) ||
			row.Type == string(models.ActivityCredit) ||
			row.Type == string(models.ActivityDividend) ||
			row.Type == string(models.ActivityInterest) ||
			row.Type == "income" {
			totalIncome += converted
		} else if row.Type == string(models.ActivityWithdrawal) ||
			row.Type == string(models.ActivityFee) ||
			row.Type == string(models.ActivityTax) ||
			row.Type == "expense" {
			totalExpense += converted
		}
	}

	// Convert salary to target currency
	salaryConverted := 0.0
	if user.SalaryAmount > 0 && user.SalaryCurrency != "" {
		salaryRate, err := getExchangeRate(user.SalaryCurrency, toCurrency)
		if err == nil {
			salaryConverted = user.SalaryAmount * salaryRate
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"currency":         toCurrency,
		"income":           totalIncome,
		"expense":          totalExpense,
		"balance":          totalIncome - totalExpense,
		"salary_amount":    salaryConverted,
		"salary_frequency": user.SalaryFrequency,
		"salary_currency":  user.SalaryCurrency,
	})
}