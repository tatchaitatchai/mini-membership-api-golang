package models

import (
	"database/sql"
	"time"
)

// AppSession represents the app_sessions table
type AppSession struct {
	ID           int64         `json:"id" db:"id"`
	StoreID      int64         `json:"store_id" db:"store_id"`
	BranchID     sql.NullInt64 `json:"branch_id" db:"branch_id"`
	StaffID      sql.NullInt64 `json:"staff_id" db:"staff_id"`
	SessionToken string        `json:"session_token" db:"session_token"`
	CreatedAt    time.Time     `json:"created_at" db:"created_at"`
	LastSeenAt   time.Time     `json:"last_seen_at" db:"last_seen_at"`
	RevokedAt    sql.NullTime  `json:"revoked_at" db:"revoked_at"`
}
