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

	// ── existing columns ──────────────────────────
	UserID      uint    `json:"user_id"`
	Title       string  `json:"title"`
	Amount      float64 `json:"amount"`
	Type        string  `json:"type"`      // kept as string for backward compat
	Category    string  `json:"category"`
	Description string  `json:"description"`

	// ── new columns ───────────────────────────────
	Currency  string         `json:"currency"   gorm:"default:'USD'"`
	Fee       float64        `json:"fee"        gorm:"default:0"`
	Status    ActivityStatus `json:"status"     gorm:"default:'POSTED'"`
	Subtype   string         `json:"subtype"`
	Metadata  string         `json:"metadata"`  // JSON string for extra context

	// ── asset fields (for BUY, SELL, DIVIDEND, SPLIT) ─
	Asset     string  `json:"asset"`
	Quantity  float64 `json:"quantity"`
	UnitPrice float64 `json:"unit_price"`
}