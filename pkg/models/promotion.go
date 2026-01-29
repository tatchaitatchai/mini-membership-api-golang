package models

import (
	"database/sql"
	"time"

	"github.com/shopspring/decimal"
)

// PromotionType represents the promotion_types table
type PromotionType struct {
	ID        int64          `json:"id" db:"id"`
	StoreID   int64          `json:"store_id" db:"store_id"`
	Name      string         `json:"name" db:"name"`
	Detail    sql.NullString `json:"detail" db:"detail"`
	CreatedAt time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt time.Time      `json:"updated_at" db:"updated_at"`
}

// PromotionTypeBranch represents the promotion_type_branches table
type PromotionTypeBranch struct {
	ID              int64     `json:"id" db:"id"`
	StoreID         int64     `json:"store_id" db:"store_id"`
	BranchID        int64     `json:"branch_id" db:"branch_id"`
	PromotionTypeID int64     `json:"promotion_type_id" db:"promotion_type_id"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

// Promotion represents the promotions table
type Promotion struct {
	ID              int64        `json:"id" db:"id"`
	StoreID         int64        `json:"store_id" db:"store_id"`
	PromotionTypeID int64        `json:"promotion_type_id" db:"promotion_type_id"`
	PromotionName   string       `json:"promotion_name" db:"promotion_name"`
	IsActive        bool         `json:"is_active" db:"is_active"`
	StartsAt        sql.NullTime `json:"starts_at" db:"starts_at"`
	EndsAt          sql.NullTime `json:"ends_at" db:"ends_at"`
	CreatedAt       time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time    `json:"updated_at" db:"updated_at"`
}

// PromotionConfig represents the promotion_configs table
type PromotionConfig struct {
	ID                    int64               `json:"id" db:"id"`
	PromotionID           int64               `json:"promotion_id" db:"promotion_id"`
	PercentDiscount       decimal.NullDecimal `json:"percent_discount" db:"percent_discount"`
	BahtDiscount          decimal.NullDecimal `json:"baht_discount" db:"baht_discount"`
	TotalPriceSetDiscount decimal.NullDecimal `json:"total_price_set_discount" db:"total_price_set_discount"`
	OldPriceSet           decimal.NullDecimal `json:"old_price_set" db:"old_price_set"`
	CountConditionProduct sql.NullInt32       `json:"count_condition_product" db:"count_condition_product"`
	ProductID             sql.NullInt64       `json:"product_id" db:"product_id"`
	CreatedAt             time.Time           `json:"created_at" db:"created_at"`
	UpdatedAt             time.Time           `json:"updated_at" db:"updated_at"`
}

// PromotionProduct represents the promotion_products table
type PromotionProduct struct {
	ID              int64     `json:"id" db:"id"`
	PromotionTypeID int64     `json:"promotion_type_id" db:"promotion_type_id"`
	ProductID       int64     `json:"product_id" db:"product_id"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}
