package models

import (
	"time"

	"gorm.io/gorm"
)

type EmailToken struct {
	gorm.Model

	UserID    uint      `json:"user_id"`
	Token     string    `json:"token" gorm:"unique"`
	ExpiresAt time.Time `json:"expires_at"`
}