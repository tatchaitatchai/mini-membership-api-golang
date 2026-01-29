package models

import (
	"database/sql"
	"time"
)

// Customer represents the customers table
type Customer struct {
	ID           int64          `json:"id" db:"id"`
	StoreID      int64          `json:"store_id" db:"store_id"`
	CustomerCode sql.NullString `json:"customer_code" db:"customer_code"`
	FullName     sql.NullString `json:"full_name" db:"full_name"`
	Phone        sql.NullString `json:"phone" db:"phone"`
	PhoneLast4   sql.NullString `json:"phone_last4" db:"phone_last4"`
	Email        sql.NullString `json:"email" db:"email"`
	IsActive     bool           `json:"is_active" db:"is_active"`
	CreatedAt    time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at" db:"updated_at"`
}
