package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/katom-membership/api/internal/domain"
	"github.com/katom-membership/api/internal/repository"
	"github.com/katom-membership/api/pkg/models"
	"github.com/shopspring/decimal"
)

type ShiftService interface {
	ListBranches(ctx context.Context, storeID int64) (*domain.ListBranchesResponse, error)
	SelectBranch(ctx context.Context, sessionToken string, storeID int64, req *domain.SelectBranchRequest) (*domain.SelectBranchResponse, error)
	OpenShift(ctx context.Context, sessionToken string, storeID int64, branchID int64, staffID *int64, req *domain.OpenShiftRequest) (*domain.OpenShiftResponse, error)
	GetCurrentShift(ctx context.Context, storeID, branchID int64) (*domain.CurrentShiftResponse, error)
	CloseShift(ctx context.Context, storeID, branchID int64, staffID *int64, req *domain.CloseShiftRequest) (*domain.CloseShiftResponse, error)
	GetShiftSummary(ctx context.Context, storeID, branchID int64) (*domain.ShiftSummaryResponse, error)
}

type shiftService struct {
	repo repository.ShiftRepository
}

func NewShiftService(repo repository.ShiftRepository) ShiftService {
	return &shiftService{repo: repo}
}

func (s *shiftService) ListBranches(ctx context.Context, storeID int64) (*domain.ListBranchesResponse, error) {
	branches, err := s.repo.GetBranchesByStoreID(ctx, storeID)
	if err != nil {
		return nil, err
	}

	result := make([]domain.BranchInfo, len(branches))
	for i, b := range branches {
		result[i] = domain.BranchInfo{
			ID:            b.ID,
			BranchName:    b.BranchName,
			IsShiftOpened: b.IsShiftOpened,
		}
	}

	return &domain.ListBranchesResponse{Branches: result}, nil
}

func (s *shiftService) SelectBranch(ctx context.Context, sessionToken string, storeID int64, req *domain.SelectBranchRequest) (*domain.SelectBranchResponse, error) {
	branch, err := s.repo.GetBranchByID(ctx, storeID, req.BranchID)
	if err != nil {
		return nil, err
	}
	if branch == nil {
		return nil, errors.New("branch not found")
	}

	// Update session with selected branch
	if err := s.repo.UpdateSessionBranch(ctx, sessionToken, storeID, req.BranchID); err != nil {
		return nil, err
	}

	fmt.Println("Branch selected: ", branch)
	fmt.Println("IsShiftOpened: ", branch.IsShiftOpened)

	return &domain.SelectBranchResponse{
		BranchID:      branch.ID,
		BranchName:    branch.BranchName,
		IsShiftOpened: branch.IsShiftOpened,
	}, nil
}

func (s *shiftService) OpenShift(ctx context.Context, sessionToken string, storeID int64, branchID int64, staffID *int64, req *domain.OpenShiftRequest) (*domain.OpenShiftResponse, error) {
	// Verify branch exists and belongs to store
	branch, err := s.repo.GetBranchByID(ctx, storeID, branchID)
	if err != nil {
		return nil, err
	}
	if branch == nil {
		return nil, errors.New("branch not found")
	}

	// Check if shift is already open
	if branch.IsShiftOpened {
		return nil, errors.New("shift is already open for this branch")
	}

	// Create shift
	now := time.Now()
	shift := &models.Shift{
		StoreID:         storeID,
		BranchID:        branchID,
		StartMoneyInbox: decimal.NewFromFloat(req.StartingCash),
		StartedAt:       now,
		IsActiveShift:   true,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if staffID != nil {
		shift.OpenedBy = sql.NullInt64{Int64: *staffID, Valid: true}
	}

	if err := s.repo.CreateShift(ctx, shift); err != nil {
		return nil, err
	}

	// Update branch shift status
	if err := s.repo.UpdateBranchShiftStatus(ctx, storeID, branchID, true); err != nil {
		return nil, err
	}

	return &domain.OpenShiftResponse{
		ShiftID:      shift.ID,
		BranchID:     branch.ID,
		BranchName:   branch.BranchName,
		StartingCash: req.StartingCash,
		StartedAt:    shift.StartedAt,
	}, nil
}

func (s *shiftService) GetCurrentShift(ctx context.Context, storeID, branchID int64) (*domain.CurrentShiftResponse, error) {
	branch, err := s.repo.GetBranchByID(ctx, storeID, branchID)
	if err != nil {
		return nil, err
	}
	if branch == nil {
		return nil, errors.New("branch not found")
	}

	shift, err := s.repo.GetActiveShiftByBranch(ctx, storeID, branchID)
	if err != nil {
		return nil, err
	}

	if shift == nil {
		return &domain.CurrentShiftResponse{
			HasActiveShift: false,
			Shift:          nil,
		}, nil
	}

	startingCash, _ := shift.StartMoneyInbox.Float64()

	return &domain.CurrentShiftResponse{
		HasActiveShift: true,
		Shift: &domain.ShiftInfo{
			ID:           shift.ID,
			BranchID:     shift.BranchID,
			BranchName:   branch.BranchName,
			StartingCash: startingCash,
			StartedAt:    shift.StartedAt,
		},
	}, nil
}

func (s *shiftService) GetShiftSummary(ctx context.Context, storeID, branchID int64) (*domain.ShiftSummaryResponse, error) {
	shift, err := s.repo.GetActiveShiftByBranch(ctx, storeID, branchID)
	if err != nil {
		return nil, err
	}
	if shift == nil {
		return nil, errors.New("no active shift found")
	}

	startingCash, _ := shift.StartMoneyInbox.Float64()
	totalSales, orderCount, err := s.repo.GetShiftSalesSummary(ctx, storeID, shift.ID)
	if err != nil {
		return nil, err
	}

	// Get only CASH payments for drawer calculation (excludes transfers, QR, etc.)
	cashSales, err := s.repo.GetShiftCashSales(ctx, storeID, shift.ID)
	if err != nil {
		return nil, err
	}

	// Get cancelled orders summary
	cancelledTotal, cancelledCount, err := s.repo.GetShiftCancelledOrdersSummary(ctx, storeID, shift.ID)
	if err != nil {
		return nil, err
	}

	totalSalesFloat, _ := totalSales.Float64()
	cashSalesFloat, _ := cashSales.Float64()
	cancelledTotalFloat, _ := cancelledTotal.Float64()

	// Expected cash in drawer = starting + cash_received - change_given
	// cashSales = total cash received from customers (before giving change)
	// We need: starting + (cash_received - change_given)
	// But cashSales is just the CASH payment amount, not accounting for change
	// Actually, the order's total_price already accounts for the net (price of items)
	// So expected = starting + net_cash_sales (cash payments only, which equals items sold for cash)
	expectedCash := startingCash + cashSalesFloat

	return &domain.ShiftSummaryResponse{
		ShiftID:        shift.ID,
		StartingCash:   startingCash,
		TotalSales:     totalSalesFloat,
		OrderCount:     orderCount,
		ExpectedCash:   expectedCash,
		CancelledTotal: cancelledTotalFloat,
		CancelledCount: cancelledCount,
	}, nil
}

func (s *shiftService) CloseShift(ctx context.Context, storeID, branchID int64, staffID *int64, req *domain.CloseShiftRequest) (*domain.CloseShiftResponse, error) {
	branch, err := s.repo.GetBranchByID(ctx, storeID, branchID)
	if err != nil {
		return nil, err
	}
	if branch == nil {
		return nil, errors.New("branch not found")
	}

	shift, err := s.repo.GetActiveShiftByBranch(ctx, storeID, branchID)
	if err != nil {
		return nil, err
	}
	if shift == nil {
		return nil, errors.New("no active shift to close")
	}

	// Get sales summary
	totalSales, orderCount, err := s.repo.GetShiftSalesSummary(ctx, storeID, shift.ID)
	if err != nil {
		return nil, err
	}

	// Get only CASH payments for drawer calculation
	cashSales, err := s.repo.GetShiftCashSales(ctx, storeID, shift.ID)
	if err != nil {
		return nil, err
	}

	startingCash, _ := shift.StartMoneyInbox.Float64()
	totalSalesFloat, _ := totalSales.Float64()
	cashSalesFloat, _ := cashSales.Float64()
	expectedCash := startingCash + cashSalesFloat
	cashDifference := req.ActualCash - expectedCash

	// Convert stock counts from request to repository format
	var stockCounts []repository.StockCountItem
	for _, sc := range req.StockCounts {
		stockCounts = append(stockCounts, repository.StockCountItem{
			ProductID:   sc.ProductID,
			ActualStock: sc.ActualStock,
		})
	}

	// Close shift in transaction
	endCash := decimal.NewFromFloat(req.ActualCash)
	expectedCashDecimal := decimal.NewFromFloat(expectedCash)
	if err := s.repo.CloseShiftTx(ctx, storeID, branchID, shift.ID, endCash, expectedCashDecimal, staffID, req.Note, stockCounts); err != nil {
		return nil, err
	}

	// Get staff name for response
	closedBy := ""
	if staffID != nil {
		closedBy, _ = s.repo.GetStaffNameByID(ctx, storeID, *staffID)
	}

	return &domain.CloseShiftResponse{
		ShiftID:        shift.ID,
		BranchID:       branchID,
		BranchName:     branch.BranchName,
		StartingCash:   startingCash,
		ExpectedCash:   expectedCash,
		ActualCash:     req.ActualCash,
		CashDifference: cashDifference,
		TotalSales:     totalSalesFloat,
		OrderCount:     orderCount,
		StartedAt:      shift.StartedAt,
		EndedAt:        time.Now(),
		ClosedBy:       closedBy,
	}, nil
}
