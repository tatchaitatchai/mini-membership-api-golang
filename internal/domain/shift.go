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
