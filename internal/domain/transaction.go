package domain

import (
	"time"

	"github.com/google/uuid"
)

type TransactionAction string

const (
	ActionEarn   TransactionAction = "EARN"
	ActionRedeem TransactionAction = "REDEEM"
)

type ProductType string

const (
	ProductType1_0Liter ProductType = "1_0_LITER"
	ProductType1_5Liter ProductType = "1_5_LITER"
)

type MemberPointTransaction struct {
	ID          uuid.UUID         `db:"id" json:"id"`
	MemberID    uuid.UUID         `db:"member_id" json:"member_id"`
	StaffUserID uuid.UUID         `db:"staff_user_id" json:"staff_user_id"`
	Action      TransactionAction `db:"action" json:"action"`
	ProductType ProductType       `db:"product_type" json:"product_type"`
	Points      int               `db:"points" json:"points"`
	ReceiptText *string           `db:"receipt_text" json:"receipt_text,omitempty"`
	CreatedAt   time.Time         `db:"created_at" json:"created_at"`
}

type ProductPoint struct {
	ProductType ProductType `json:"product_type" binding:"required,oneof=1_0_LITER 1_5_LITER"`
	Points      int         `json:"points" binding:"required,min=1"`
}

type TransactionCreateRequest struct {
	MemberID    uuid.UUID         `json:"member_id" binding:"required"`
	Action      TransactionAction `json:"action" binding:"required,oneof=EARN REDEEM"`
	Products    []ProductPoint    `json:"products" binding:"required,min=1,dive"`
	ReceiptText *string           `json:"receipt_text,omitempty"`
}

type TransactionCreateResponse struct {
	Transactions []MemberPointTransaction `json:"transactions"`
	TotalPoints  int                      `json:"total_points"`
	Message      string                   `json:"message"`
}

type TransactionListResponse struct {
	Transactions []MemberPointTransaction `json:"transactions"`
	Total        int                      `json:"total"`
	Page         int                      `json:"page"`
	Limit        int                      `json:"limit"`
}
