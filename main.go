package main

import (
	"embed"
	"finance-tracker/config"
	"finance-tracker/models"
	"finance-tracker/routes"
	"io/fs"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

//go:embed public/*
var staticFiles embed.FS

func main() {
	godotenv.Load(".env") // optional — ignored in production

	config.ConnectDB()

	config.DB.AutoMigrate( //migrate db structure
		&models.User{},
		&models.Transaction{},
		&models.EmailToken{},
		&models.Sessions{},
	)

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// Serve embedded static files
	publicFS, _ := fs.Sub(staticFiles, "public")
	r.StaticFS("/static", http.FS(publicFS))

	// Serve index.html at root
	r.GET("/", func(c *gin.Context) {
		data, err := staticFiles.ReadFile("public/index.html")
		if err != nil {
			c.String(401,"index html is Not found", http.StatusNotFound)
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", data)
	})

	// Serve favicon
	r.GET("/favicon.ico", func(c *gin.Context) {
		data, err := staticFiles.ReadFile("public/favicon.ico")
		if err != nil {
			c.Status(http.StatusNotFound)
			return
		}
		c.Data(http.StatusOK, "image/x-icon", data)
	})

	// API + auth routes
	routes.SetupRoutes(r)

	port := os.Getenv("PORT")

	if port == "" {
		port = "3000"
	}

	r.Run(":" + port)
}