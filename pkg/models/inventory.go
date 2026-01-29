package models

import (
	"database/sql"
	"time"
)

// MovementType represents valid inventory movement types
type MovementType string

const (
	MovementTypeSale        MovementType = "SALE"
	MovementTypeCancelSale  MovementType = "CANCEL_SALE"
	MovementTypeReceive     MovementType = "RECEIVE"
	MovementTypeIssue       MovementType = "ISSUE"
	MovementTypeAdjust      MovementType = "ADJUST"
	MovementTypeTransferIn  MovementType = "TRANSFER_IN"
	MovementTypeTransferOut MovementType = "TRANSFER_OUT"
	MovementTypeDamage      MovementType = "DAMAGE"
)

// InventoryMovement represents the inventory_movements table
type InventoryMovement struct {
	ID             int64          `json:"id" db:"id"`
	StoreID        int64          `json:"store_id" db:"store_id"`
	BranchID       int64          `json:"branch_id" db:"branch_id"`
	ProductID      int64          `json:"product_id" db:"product_id"`
	MovementType   MovementType   `json:"movement_type" db:"movement_type"`
	QuantityChange int            `json:"quantity_change" db:"quantity_change"`
	FromStockCount sql.NullInt32  `json:"from_stock_count" db:"from_stock_count"`
	ToStockCount   sql.NullInt32  `json:"to_stock_count" db:"to_stock_count"`
	Reason         sql.NullString `json:"reason" db:"reason"`
	Note           sql.NullString `json:"note" db:"note"`
	ChangedBy      sql.NullInt64  `json:"changed_by" db:"changed_by"`
	ReferenceTable sql.NullString `json:"reference_table" db:"reference_table"`
	ReferenceID    sql.NullInt64  `json:"reference_id" db:"reference_id"`
	CreatedAt      time.Time      `json:"created_at" db:"created_at"`
}

// StockTransferStatus represents valid stock transfer statuses
type StockTransferStatus string

const (
	StockTransferStatusCreated   StockTransferStatus = "CREATED"
	StockTransferStatusSent      StockTransferStatus = "SENT"
	StockTransferStatusReceived  StockTransferStatus = "RECEIVED"
	StockTransferStatusCancelled StockTransferStatus = "CANCELLED"
)

// StockTransfer represents the stock_transfers table
type StockTransfer struct {
	ID           int64               `json:"id" db:"id"`
	StoreID      int64               `json:"store_id" db:"store_id"`
	FromBranchID sql.NullInt64       `json:"from_branch_id" db:"from_branch_id"`
	ToBranchID   int64               `json:"to_branch_id" db:"to_branch_id"`
	Status       StockTransferStatus `json:"status" db:"status"`
	SentBy       sql.NullInt64       `json:"sent_by" db:"sent_by"`
	ReceivedBy   sql.NullInt64       `json:"received_by" db:"received_by"`
	SentAt       sql.NullTime        `json:"sent_at" db:"sent_at"`
	ReceivedAt   sql.NullTime        `json:"received_at" db:"received_at"`
	Note         sql.NullString      `json:"note" db:"note"`
	CreatedAt    time.Time           `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time           `json:"updated_at" db:"updated_at"`
}

// StockTransferItem represents the stock_transfer_items table
type StockTransferItem struct {
	ID              int64     `json:"id" db:"id"`
	StockTransferID int64     `json:"stock_transfer_id" db:"stock_transfer_id"`
	ProductID       int64     `json:"product_id" db:"product_id"`
	SendCount       int       `json:"send_count" db:"send_count"`
	ReceiveCount    int       `json:"receive_count" db:"receive_count"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}
