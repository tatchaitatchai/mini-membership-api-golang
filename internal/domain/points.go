package domain

import "time"

type CustomerProductPoints struct {
	ID          int64     `db:"id" json:"id"`
	StoreID     int64     `db:"store_id" json:"store_id"`
	CustomerID  int64     `db:"customer_id" json:"customer_id"`
	ProductID   int64     `db:"product_id" json:"product_id"`
	Points      int       `db:"points" json:"points"`
	TotalPoints int       `db:"total_points" json:"total_points"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}

type PointTransaction struct {
	ID              int64     `db:"id" json:"id"`
	StoreID         int64     `db:"store_id" json:"store_id"`
	BranchID        int64     `db:"branch_id" json:"branch_id"`
	CustomerID      int64     `db:"customer_id" json:"customer_id"`
	TransactionType string    `db:"transaction_type" json:"transaction_type"`
	PointsChange    int       `db:"points_change" json:"points_change"`
	ReferenceTable  *string   `db:"reference_table" json:"reference_table,omitempty"`
	ReferenceID     *int64    `db:"reference_id" json:"reference_id,omitempty"`
	ProductID       *int64    `db:"product_id" json:"product_id,omitempty"`
	Note            *string   `db:"note" json:"note,omitempty"`
	StaffID         *int64    `db:"staff_id" json:"staff_id,omitempty"`
	CreatedAt       time.Time `db:"created_at" json:"created_at"`
}

type PointRedemption struct {
	ID         int64     `db:"id" json:"id"`
	StoreID    int64     `db:"store_id" json:"store_id"`
	BranchID   int64     `db:"branch_id" json:"branch_id"`
	CustomerID int64     `db:"customer_id" json:"customer_id"`
	ProductID  int64     `db:"product_id" json:"product_id"`
	PointsUsed int       `db:"points_used" json:"points_used"`
	Quantity   int       `db:"quantity" json:"quantity"`
	Status     string    `db:"status" json:"status"`
	StaffID    *int64    `db:"staff_id" json:"staff_id,omitempty"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time `db:"updated_at" json:"updated_at"`
}

// API Request/Response types

type GetCustomerPointsResponse struct {
	CustomerID   int64                       `json:"customer_id"`
	CustomerName string                      `json:"customer_name"`
	CustomerCode string                      `json:"customer_code"`
	Products     []CustomerProductPointsInfo `json:"products"`
}

// CustomerProductPointsInfo shows points for a specific product
type CustomerProductPointsInfo struct {
	ProductID      int64   `db:"product_id" json:"product_id"`
	ProductName    string  `db:"product_name" json:"product_name"`
	CategoryName   *string `db:"category_name" json:"category_name,omitempty"`
	ImagePath      *string `db:"image_path" json:"image_path,omitempty"`
	Points         int     `db:"points" json:"points"`
	TotalPoints    int     `db:"total_points" json:"total_points"`
	PointsToRedeem int     `db:"points_to_redeem" json:"points_to_redeem"`
	CanRedeem      bool    `db:"can_redeem" json:"can_redeem"`
}

type RedeemableProduct struct {
	ID             int64   `db:"id" json:"id"`
	ProductName    string  `db:"product_name" json:"product_name"`
	CategoryName   *string `db:"category_name" json:"category_name,omitempty"`
	ImagePath      *string `db:"image_path" json:"image_path,omitempty"`
	PointsToRedeem int     `db:"points_to_redeem" json:"points_to_redeem"`
	OnStock        int     `db:"on_stock" json:"on_stock"`
}

type ListRedeemableProductsResponse struct {
	Products []RedeemableProduct `json:"products"`
}

type RedeemPointsRequest struct {
	CustomerID int64 `json:"customer_id" binding:"required"`
	ProductID  int64 `json:"product_id" binding:"required"`
	Quantity   int   `json:"quantity" binding:"required,min=1"`
}

type RedeemPointsResponse struct {
	RedemptionID    int64  `json:"redemption_id"`
	PointsUsed      int    `json:"points_used"`
	RemainingPoints int    `json:"remaining_points"`
	ProductName     string `json:"product_name"`
	Quantity        int    `json:"quantity"`
	Message         string `json:"message"`
}

type OrderItemForPoints struct {
	ProductID int64 `json:"product_id"`
	Quantity  int   `json:"quantity"`
}

type EarnPointsResponse struct {
	PointsEarned    int `json:"points_earned"`
	TotalPoints     int `json:"total_points"`
	RemainingPoints int `json:"remaining_points"`
}

type PointHistoryItem struct {
	ID              int64     `json:"id"`
	TransactionType string    `json:"transaction_type"`
	PointsChange    int       `json:"points_change"`
	ProductName     *string   `json:"product_name,omitempty"`
	Note            *string   `json:"note,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}

type GetPointHistoryResponse struct {
	CustomerID int64              `json:"customer_id"`
	History    []PointHistoryItem `json:"history"`
	Total      int                `json:"total"`
}
