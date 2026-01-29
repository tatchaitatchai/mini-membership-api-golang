package models

import (
	"database/sql"
	"time"

	"github.com/shopspring/decimal"
)

// Category represents the categories table
type Category struct {
	ID           int64     `json:"id" db:"id"`
	StoreID      int64     `json:"store_id" db:"store_id"`
	CategoryName string    `json:"category_name" db:"category_name"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// Product represents the products table
type Product struct {
	ID          int64           `json:"id" db:"id"`
	StoreID     int64           `json:"store_id" db:"store_id"`
	CategoryID  sql.NullInt64   `json:"category_id" db:"category_id"`
	ProductName string          `json:"product_name" db:"product_name"`
	ImagePath   sql.NullString  `json:"image_path" db:"image_path"`
	IsActive    bool            `json:"is_active" db:"is_active"`
	SKU         sql.NullString  `json:"sku" db:"sku"`
	Barcode     sql.NullString  `json:"barcode" db:"barcode"`
	BasePrice   decimal.Decimal `json:"base_price" db:"base_price"`
	CreatedAt   time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at" db:"updated_at"`
}

// BranchProduct represents the branch_products table
type BranchProduct struct {
	ID           int64     `json:"id" db:"id"`
	StoreID      int64     `json:"store_id" db:"store_id"`
	BranchID     int64     `json:"branch_id" db:"branch_id"`
	ProductID    int64     `json:"product_id" db:"product_id"`
	IsActive     bool      `json:"is_active" db:"is_active"`
	OnStock      int       `json:"on_stock" db:"on_stock"`
	ReorderLevel int       `json:"reorder_level" db:"reorder_level"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}
