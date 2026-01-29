package models

import (
	"database/sql"
	"time"

	"github.com/shopspring/decimal"
)

// PaymentMethod represents valid payment methods
type PaymentMethod string

const (
	PaymentMethodCash     PaymentMethod = "CASH"
	PaymentMethodTransfer PaymentMethod = "TRANSFER"
	PaymentMethodQR       PaymentMethod = "QR"
	PaymentMethodCard     PaymentMethod = "CARD"
	PaymentMethodOther    PaymentMethod = "OTHER"
)

// Payment represents the payments table
type Payment struct {
	ID        int64           `json:"id" db:"id"`
	OrderID   int64           `json:"order_id" db:"order_id"`
	Method    PaymentMethod   `json:"method" db:"method"`
	Amount    decimal.Decimal `json:"amount" db:"amount"`
	PaidAt    time.Time       `json:"paid_at" db:"paid_at"`
	CreatedAt time.Time       `json:"created_at" db:"created_at"`
}

// PaymentAttachment represents the payment_attachments table
type PaymentAttachment struct {
	ID        int64          `json:"id" db:"id"`
	PaymentID int64          `json:"payment_id" db:"payment_id"`
	FilePath  string         `json:"file_path" db:"file_path"`
	FileType  sql.NullString `json:"file_type" db:"file_type"`
	CreatedAt time.Time      `json:"created_at" db:"created_at"`
}
