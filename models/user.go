package models

import "gorm.io/gorm"

type User struct {
	gorm.Model

	Name     string `json:"name"`
	Email    string `json:"email" gorm:"unique"`
	Password string `json:"-"`
	SalaryAmmount  float64 `json:"salaryAmmount"`
	SalaryCurrency string  `json:"salaryCurrency"`
	SalaryFrequency string  `json:"salaryFrequency"`
}
