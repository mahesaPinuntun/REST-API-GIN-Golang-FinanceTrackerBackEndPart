package routes

import (
	"finance-tracker/controllers"
	"finance-tracker/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {

	r.POST("/register", controllers.Register)
	r.POST("/login", controllers.Login)
	
	// Email confirmation — public (token is the auth)
	r.GET("/api/auth/confirm", controllers.ConfirmEmail)
	api := r.Group("/api")
	api.Use(middleware.AuthMiddleware())

	{
		// Transactions
		api.POST("/transactions", controllers.CreateTransaction)
		api.GET("/transactions", controllers.GetTransactions)

		// Dashboard
		api.GET("/dashboard", controllers.Dashboard)
		api.GET("/dashboard/convert", controllers.GetDashboardInCurrency)

		// Currency
		api.GET("/currency/convert", controllers.ConvertCurrency)
		api.GET("/currency/supported", controllers.GetSupportedCurrencies)
	}
}