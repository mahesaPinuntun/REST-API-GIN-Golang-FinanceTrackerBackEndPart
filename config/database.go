package config

import (
	"fmt"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB() {

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASS"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_SSLMODE"),
	)

	db, err := gorm.Open(
		postgres.Open(dsn),
		&gorm.Config{},
	)

	if err != nil {
		panic(err)
	}

	DB = db
	fmt.Println("HOST =", os.Getenv("DB_HOST"))
	fmt.Println("USER =", os.Getenv("DB_USER"))
	fmt.Println("DB   =", os.Getenv("DB_NAME"))
	fmt.Println("PORT =", os.Getenv("DB_PORT"))
	fmt.Println("HOST:", os.Getenv("DB_HOST"))
	fmt.Println("DB:", os.Getenv("DB_NAME"))
	fmt.Println("SSL:", os.Getenv("DB_SSLMODE"))s
}