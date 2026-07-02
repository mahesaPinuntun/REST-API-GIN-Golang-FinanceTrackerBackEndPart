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