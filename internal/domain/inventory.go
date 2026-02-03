package domain

import "time"

// MovementType represents the type of inventory movement
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

// InventoryMovement represents a stock movement record
type InventoryMovement struct {
	ID             int64        `json:"id" db:"id"`
	StoreID        int64        `json:"store_id" db:"store_id"`
	BranchID       int64        `json:"branch_id" db:"branch_id"`
	ProductID      int64        `json:"product_id" db:"product_id"`
	MovementType   MovementType `json:"movement_type" db:"movement_type"`
	QuantityChange int          `json:"quantity_change" db:"quantity_change"`
	FromStockCount *int         `json:"from_stock_count,omitempty" db:"from_stock_count"`
	ToStockCount   *int         `json:"to_stock_count,omitempty" db:"to_stock_count"`
	Reason         *string      `json:"reason,omitempty" db:"reason"`
	Note           *string      `json:"note,omitempty" db:"note"`
	ChangedBy      *int64       `json:"changed_by,omitempty" db:"changed_by"`
	ReferenceTable *string      `json:"reference_table,omitempty" db:"reference_table"`
	ReferenceID    *int64       `json:"reference_id,omitempty" db:"reference_id"`
	CreatedAt      time.Time    `json:"created_at" db:"created_at"`
}

// InventoryMovementResponse represents the API response for inventory movement
type InventoryMovementResponse struct {
	ID             int64        `json:"id"`
	ProductID      int64        `json:"product_id"`
	ProductName    string       `json:"product_name"`
	MovementType   MovementType `json:"movement_type"`
	QuantityChange int          `json:"quantity_change"`
	FromStockCount *int         `json:"from_stock_count,omitempty"`
	ToStockCount   *int         `json:"to_stock_count,omitempty"`
	Reason         *string      `json:"reason,omitempty"`
	Note           *string      `json:"note,omitempty"`
	ChangedByName  *string      `json:"changed_by_name,omitempty"`
	CreatedAt      time.Time    `json:"created_at"`
}

// AdjustStockRequest represents a request to adjust stock
type AdjustStockRequest struct {
	ProductID int64  `json:"product_id" binding:"required"`
	Quantity  int    `json:"quantity" binding:"required,min=1"`
	Reason    string `json:"reason" binding:"required"`
	Note      string `json:"note,omitempty"`
}

// BranchProduct represents a product's stock in a branch
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

// LowStockItem represents a product with low stock
type LowStockItem struct {
	ProductID    int64  `json:"product_id"`
	ProductName  string `json:"product_name"`
	CategoryName string `json:"category_name"`
	OnStock      int    `json:"on_stock"`
	ReorderLevel int    `json:"reorder_level"`
	Price        int64  `json:"price"`
}

// LowStockResponse represents the API response for low stock items
type LowStockResponse struct {
	Items      []LowStockItem `json:"items"`
	TotalCount int            `json:"total_count"`
}
