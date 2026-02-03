package domain

import "time"

type BranchInfo struct {
	ID            int64  `json:"id"`
	BranchName    string `json:"branch_name"`
	IsShiftOpened bool   `json:"is_shift_opened"`
}

type ListBranchesResponse struct {
	Branches []BranchInfo `json:"branches"`
}

type SelectBranchRequest struct {
	BranchID int64 `json:"branch_id" binding:"required"`
}

type SelectBranchResponse struct {
	BranchID      int64  `json:"branch_id"`
	BranchName    string `json:"branch_name"`
	IsShiftOpened bool   `json:"is_shift_opened"`
}

type OpenShiftRequest struct {
	StartingCash float64 `json:"starting_cash" binding:"gte=0"`
}

type OpenShiftResponse struct {
	ShiftID      int64     `json:"shift_id"`
	BranchID     int64     `json:"branch_id"`
	BranchName   string    `json:"branch_name"`
	StartingCash float64   `json:"starting_cash"`
	StartedAt    time.Time `json:"started_at"`
}

type ShiftInfo struct {
	ID           int64     `json:"id"`
	BranchID     int64     `json:"branch_id"`
	BranchName   string    `json:"branch_name"`
	StartingCash float64   `json:"starting_cash"`
	StartedAt    time.Time `json:"started_at"`
}

type CurrentShiftResponse struct {
	HasActiveShift bool       `json:"has_active_shift"`
	Shift          *ShiftInfo `json:"shift,omitempty"`
}

type CloseShiftRequest struct {
	ActualCash  float64           `json:"actual_cash" binding:"gte=0"`
	StockCounts []StockCountInput `json:"stock_counts,omitempty"`
	Note        string            `json:"note,omitempty"`
}

type StockCountInput struct {
	ProductID   int64 `json:"product_id" binding:"required"`
	ActualStock int   `json:"actual_stock" binding:"gte=0"`
}

type CloseShiftResponse struct {
	ShiftID        int64     `json:"shift_id"`
	BranchID       int64     `json:"branch_id"`
	BranchName     string    `json:"branch_name"`
	StartingCash   float64   `json:"starting_cash"`
	ExpectedCash   float64   `json:"expected_cash"`
	ActualCash     float64   `json:"actual_cash"`
	CashDifference float64   `json:"cash_difference"`
	TotalSales     float64   `json:"total_sales"`
	OrderCount     int       `json:"order_count"`
	StartedAt      time.Time `json:"started_at"`
	EndedAt        time.Time `json:"ended_at"`
	ClosedBy       string    `json:"closed_by,omitempty"`
}

type ShiftSummaryResponse struct {
	ShiftID      int64   `json:"shift_id"`
	StartingCash float64 `json:"starting_cash"`
	TotalSales   float64 `json:"total_sales"`
	OrderCount   int     `json:"order_count"`
	ExpectedCash float64 `json:"expected_cash"`
}
