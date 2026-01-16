package domain

import (
	"time"

	"github.com/google/uuid"
)

type TransactionAction string

const (
	ActionAdd    TransactionAction = "add"
	ActionDeduct TransactionAction = "deduct"
	ActionRedeem TransactionAction = "redeem"
	ActionAdjust TransactionAction = "adjust"
)

type ProductType string

const (
	ProductType1_0Liter ProductType = "1.0_liter"
	ProductType1_5Liter ProductType = "1.5_liter"
	ProductTypeOther    ProductType = "other"
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

type TransactionCreateRequest struct {
	MemberID    uuid.UUID         `json:"member_id" binding:"required"`
	Action      TransactionAction `json:"action" binding:"required,oneof=add deduct redeem adjust"`
	ProductType ProductType       `json:"product_type" binding:"required,oneof=1.0_liter 1.5_liter other"`
	Points      int               `json:"points" binding:"required,min=1"`
	ReceiptText *string           `json:"receipt_text,omitempty"`
}

type TransactionListResponse struct {
	Transactions []MemberPointTransaction `json:"transactions"`
	Total        int                      `json:"total"`
	Page         int                      `json:"page"`
	Limit        int                      `json:"limit"`
}
