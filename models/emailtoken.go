package models

import (
	"time"

	"gorm.io/gorm"
)

type EmailToken struct {
	gorm.Model

	UserEmail string    `json:"user_email"`
	Token     string    `json:"token" gorm:"unique"`
	ExpiresAt time.Time `json:"expires_at"`
}