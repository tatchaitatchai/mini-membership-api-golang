package models

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/shopspring/decimal"
)

// OrderStatus represents valid order statuses
type OrderStatus string

const (
	OrderStatusOpen      OrderStatus = "OPEN"
	OrderStatusPaid      OrderStatus = "PAID"
	OrderStatusCancelled OrderStatus = "CANCELLED"
	OrderStatusVoid      OrderStatus = "VOID"
)

// Order represents the orders table
type Order struct {
	ID            int64           `json:"id" db:"id"`
	StoreID       int64           `json:"store_id" db:"store_id"`
	BranchID      int64           `json:"branch_id" db:"branch_id"`
	ShiftID       sql.NullInt64   `json:"shift_id" db:"shift_id"`
	CustomerID    sql.NullInt64   `json:"customer_id" db:"customer_id"`
	StaffID       int64           `json:"staff_id" db:"staff_id"`
	Subtotal      decimal.Decimal `json:"subtotal" db:"subtotal"`
	DiscountTotal decimal.Decimal `json:"discount_total" db:"discount_total"`
	TotalPrice    decimal.Decimal `json:"total_price" db:"total_price"`
	ChangeAmount  decimal.Decimal `json:"change_amount" db:"change_amount"`
	Status        OrderStatus     `json:"status" db:"status"`
	CancelledBy   sql.NullInt64   `json:"cancelled_by" db:"cancelled_by"`
	CancelReason  sql.NullString  `json:"cancel_reason" db:"cancel_reason"`
	CancelledAt   sql.NullTime    `json:"cancelled_at" db:"cancelled_at"`
	CreatedAt     time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at" db:"updated_at"`
}

// OrderItem represents the order_items table
type OrderItem struct {
	ID             int64           `json:"id" db:"id"`
	OrderID        int64           `json:"order_id" db:"order_id"`
	ProductID      int64           `json:"product_id" db:"product_id"`
	Quantity       int             `json:"quantity" db:"quantity"`
	Price          decimal.Decimal `json:"price" db:"price"`
	FromStockCount sql.NullInt32   `json:"from_stock_count" db:"from_stock_count"`
	ToStockCount   sql.NullInt32   `json:"to_stock_count" db:"to_stock_count"`
	CreatedAt      time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at" db:"updated_at"`
}

// OrderPromotion represents the order_promotions table
type OrderPromotion struct {
	ID             int64           `json:"id" db:"id"`
	OrderID        int64           `json:"order_id" db:"order_id"`
	PromotionID    int64           `json:"promotion_id" db:"promotion_id"`
	DiscountAmount decimal.Decimal `json:"discount_amount" db:"discount_amount"`
	Metadata       json.RawMessage `json:"metadata" db:"metadata"`
	CreatedAt      time.Time       `json:"created_at" db:"created_at"`
}
