package models

import "gorm.io/gorm"

type Transaction struct {
	gorm.Model

	UserID      uint    `json:"user_id"`
	Title       string  `json:"title"`
	Amount      float64 `json:"amount"`
	Type        string  `json:"type"`
	Category    string  `json:"category"`
	Description string  `json:"description"`
}
