package domain

import "time"

// StockTransferStatus represents the status of a stock transfer
type StockTransferStatus string

const (
	StockTransferStatusCreated   StockTransferStatus = "CREATED"
	StockTransferStatusSent      StockTransferStatus = "SENT"
	StockTransferStatusReceived  StockTransferStatus = "RECEIVED"
	StockTransferStatusCancelled StockTransferStatus = "CANCELLED"
)

// StockTransfer represents a stock transfer between branches
type StockTransfer struct {
	ID           int64               `json:"id"`
	StoreID      int64               `json:"store_id"`
	FromBranchID *int64              `json:"from_branch_id,omitempty"`
	ToBranchID   int64               `json:"to_branch_id"`
	Status       StockTransferStatus `json:"status"`
	SentBy       *int64              `json:"sent_by,omitempty"`
	ReceivedBy   *int64              `json:"received_by,omitempty"`
	SentAt       *time.Time          `json:"sent_at,omitempty"`
	ReceivedAt   *time.Time          `json:"received_at,omitempty"`
	Note         *string             `json:"note,omitempty"`
	CreatedAt    time.Time           `json:"created_at"`
	UpdatedAt    time.Time           `json:"updated_at"`
}

// StockTransferItem represents an item in a stock transfer
type StockTransferItem struct {
	ID              int64     `json:"id"`
	StockTransferID int64     `json:"stock_transfer_id"`
	ProductID       int64     `json:"product_id"`
	SendCount       int       `json:"send_count"`
	ReceiveCount    int       `json:"receive_count"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// StockTransferResponse represents a stock transfer with details for API response
type StockTransferResponse struct {
	ID             int64                       `json:"id"`
	FromBranchID   *int64                      `json:"from_branch_id,omitempty"`
	FromBranchName *string                     `json:"from_branch_name,omitempty"`
	ToBranchID     int64                       `json:"to_branch_id"`
	ToBranchName   string                      `json:"to_branch_name"`
	Status         StockTransferStatus         `json:"status"`
	SentByName     *string                     `json:"sent_by_name,omitempty"`
	ReceivedByName *string                     `json:"received_by_name,omitempty"`
	SentAt         *time.Time                  `json:"sent_at,omitempty"`
	ReceivedAt     *time.Time                  `json:"received_at,omitempty"`
	Note           *string                     `json:"note,omitempty"`
	Items          []StockTransferItemResponse `json:"items"`
	CreatedAt      time.Time                   `json:"created_at"`
}

// StockTransferItemResponse represents an item with product details
type StockTransferItemResponse struct {
	ID           int64  `json:"id"`
	ProductID    int64  `json:"product_id"`
	ProductName  string `json:"product_name"`
	SendCount    int    `json:"send_count"`
	ReceiveCount int    `json:"receive_count"`
}

// CreateStockTransferRequest represents a request to create a stock transfer
type CreateStockTransferRequest struct {
	FromBranchID *int64                         `json:"from_branch_id,omitempty"`
	ToBranchID   int64                          `json:"to_branch_id" binding:"required"`
	Note         *string                        `json:"note,omitempty"`
	Items        []CreateStockTransferItemInput `json:"items" binding:"required,min=1"`
}

// CreateStockTransferItemInput represents an item input for creating a transfer
type CreateStockTransferItemInput struct {
	ProductID int64 `json:"product_id" binding:"required"`
	SendCount int   `json:"send_count" binding:"required,min=1"`
}

// UpdateStockTransferRequest represents a request to update transfer status
type UpdateStockTransferRequest struct {
	Status StockTransferStatus            `json:"status,omitempty"`
	Items  []UpdateStockTransferItemInput `json:"items,omitempty"`
}

// UpdateStockTransferItemInput represents an item input for receiving
type UpdateStockTransferItemInput struct {
	ProductID    int64 `json:"product_id" binding:"required"`
	ReceiveCount int   `json:"receive_count" binding:"required,min=0"`
}

// WithdrawGoodsRequest represents a request to withdraw goods (simplified transfer)
type WithdrawGoodsRequest struct {
	Items []WithdrawItemInput `json:"items" binding:"required,min=1"`
	Note  *string             `json:"note,omitempty"`
}

// WithdrawItemInput represents an item to withdraw
type WithdrawItemInput struct {
	ProductID int64 `json:"product_id" binding:"required"`
	Quantity  int   `json:"quantity" binding:"required,min=1"`
}

// StockTransferListResponse represents paginated list of transfers
type StockTransferListResponse struct {
	Transfers []StockTransferResponse `json:"transfers"`
	Total     int                     `json:"total"`
}
