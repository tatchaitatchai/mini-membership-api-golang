package models

import (
	"database/sql"
	"time"
)

// Store represents the stores table
type Store struct {
	ID        int64     `json:"id" db:"id"`
	StoreName string    `json:"store_name" db:"store_name"`
	IsActive  bool      `json:"is_active" db:"is_active"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Branch represents the branches table
type Branch struct {
	ID            int64          `json:"id" db:"id"`
	StoreID       int64          `json:"store_id" db:"store_id"`
	BranchName    string         `json:"branch_name" db:"branch_name"`
	IsShiftOpened bool           `json:"is_shift_opened" db:"is_shift_opened"`
	ShiftOpenedAt sql.NullTime   `json:"shift_opened_at" db:"shift_opened_at"`
	ShiftClosedAt sql.NullTime   `json:"shift_closed_at" db:"shift_closed_at"`
	IsActive      bool           `json:"is_active" db:"is_active"`
	CreatedAt     time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at" db:"updated_at"`
}

// StaffAccount represents the staff_accounts table
type StaffAccount struct {
	ID            int64          `json:"id" db:"id"`
	StoreID       int64          `json:"store_id" db:"store_id"`
	BranchID      sql.NullInt64  `json:"branch_id" db:"branch_id"`
	Email         sql.NullString `json:"email" db:"email"`
	PasswordHash  sql.NullString `json:"-" db:"password_hash"`
	PinHash       sql.NullString `json:"-" db:"pin_hash"`
	IsActive      bool           `json:"is_active" db:"is_active"`
	IsStoreMaster bool           `json:"is_store_master" db:"is_store_master"`
	IsWorking     bool           `json:"is_working" db:"is_working"`
	CreatedAt     time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at" db:"updated_at"`
}
