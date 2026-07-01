package models

import "gorm.io/gorm"

// ActivityType — 14 canonical types from activity-types.md
type ActivityType string

const (
	ActivityBuy         ActivityType = "BUY"
	ActivitySell        ActivityType = "SELL"
	ActivitySplit       ActivityType = "SPLIT"
	ActivityDeposit     ActivityType = "DEPOSIT"
	ActivityWithdrawal  ActivityType = "WITHDRAWAL"
	ActivityTransferIn  ActivityType = "TRANSFER_IN"
	ActivityTransferOut ActivityType = "TRANSFER_OUT"
	ActivityDividend    ActivityType = "DIVIDEND"
	ActivityInterest    ActivityType = "INTEREST"
	ActivityCredit      ActivityType = "CREDIT"
	ActivityFee         ActivityType = "FEE"
	ActivityTax         ActivityType = "TAX"
	ActivityAdjustment  ActivityType = "ADJUSTMENT"
	ActivityUnknown     ActivityType = "UNKNOWN"
)

// ActivityStatus — controls whether activity affects calculations
type ActivityStatus string

const (
	StatusPosted  ActivityStatus = "POSTED"
	StatusPending ActivityStatus = "PENDING"
	StatusDraft   ActivityStatus = "DRAFT"
	StatusVoid    ActivityStatus = "VOID"
)

type Transaction struct {
	gorm.Model
	
	UserID      uint    `json:"userId"`
	UserEmail   string  `json:"userEmail"`
	Amount      float64 `json:"amount" gorm:"default:0"`
	Category    string  `json:"category"`
	Description string  `json:"description" gorm:"default:'no description'"`
	Currency    string  `json:"currency" gorm:"default:'USD'"`
	Asset       string  `json:"asset"`
	Type        string  `json:"type"`
	Status      string  `json:"status"`
}