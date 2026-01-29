package models

import (
	"database/sql"
	"time"

	"github.com/shopspring/decimal"
)

// Shift represents the shifts table
type Shift struct {
	ID              int64                  `json:"id" db:"id"`
	StoreID         int64                  `json:"store_id" db:"store_id"`
	BranchID        int64                  `json:"branch_id" db:"branch_id"`
	StartMoneyInbox decimal.Decimal        `json:"start_money_inbox" db:"start_money_inbox"`
	EndMoneyInbox   decimal.NullDecimal    `json:"end_money_inbox" db:"end_money_inbox"`
	StartedAt       time.Time              `json:"started_at" db:"started_at"`
	EndedAt         sql.NullTime           `json:"ended_at" db:"ended_at"`
	IsActiveShift   bool                   `json:"is_active_shift" db:"is_active_shift"`
	OpenedBy        sql.NullInt64          `json:"opened_by" db:"opened_by"`
	ClosedBy        sql.NullInt64          `json:"closed_by" db:"closed_by"`
	CreatedAt       time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at" db:"updated_at"`
}
