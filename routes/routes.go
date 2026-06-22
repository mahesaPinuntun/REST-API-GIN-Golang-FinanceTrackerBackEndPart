package routes

import (
	"finance-tracker/controllers"
	"finance-tracker/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {

	r.POST("/register",
		controllers.Register)

	r.POST("/login",
		controllers.Login)

	api := r.Group("/api")
	api.Use(middleware.AuthMiddleware())

	{
		api.POST(
			"/transactions",
			controllers.CreateTransaction,
		)

		api.GET(
			"/transactions",
			controllers.GetTransactions,
		)

		api.GET(
			"/dashboard",
			controllers.Dashboard,
		)
	}
}
