package models

import "gorm.io/gorm"

type User struct {
	gorm.Model

	// ── existing columns ──────────────────────────
	Name     string `json:"name"`
	Email    string `json:"email" gorm:"unique"`
	Password string `json:"-"`

	// ── salary fields (fixed typo + exported bool) ─
	SalaryAmount    float64 `json:"salaryAmount"`    
	SalaryCurrency  string  `json:"salaryCurrency"  gorm:"default:'USD'"`
	SalaryFrequency string  `json:"salaryFrequency" gorm:"default:'monthly'"`

	// ── new columns ───────────────────────────────
	IsEmailConfirmed bool `json:"isEmailConfirmed" gorm:"default:false"` 
}

// PendingUpdate stores requested profile changes before email confirmation
type PendingUpdate struct {
	gorm.Model
 
	UserID    uint      `json:"user_id"`
	UserEmail string    `json:"user_email"`
	Token     string    `json:"token" gorm:"unique"`
	ExpiresAt time.Time `json:"expires_at"`
 
	// Fields that can be updated — empty string means no change
	NewName            string  `json:"new_name"`
	NewEmail           string  `json:"new_email"`
	NewSalaryAmount    float64 `json:"new_salary_amount"`
	NewSalaryCurrency  string  `json:"new_salary_currency"`
	NewSalaryFrequency string  `json:"new_salary_frequency"`
}