package main

import (
	"finance-tracker/config"
	"finance-tracker/models"
	"finance-tracker/routes"
	"log"

	"github.com/gin-gonic/gin"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Failed to load .env file")
	}
	config.ConnectDB()

	config.DB.AutoMigrate(
		&models.User{},
		&models.Transaction{},
	)

	r := gin.Default()

	routes.SetupRoutes(r)

	r.Run(":8080")
}