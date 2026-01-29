package service

import (
	"context"
	"database/sql"
	"errors"
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
	GetCurrentShift(ctx context.Context, branchID int64) (*domain.CurrentShiftResponse, error)
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
	branch, err := s.repo.GetBranchByID(ctx, req.BranchID)
	if err != nil {
		return nil, err
	}
	if branch == nil {
		return nil, errors.New("branch not found")
	}
	if branch.StoreID != storeID {
		return nil, errors.New("branch does not belong to this store")
	}

	// Update session with selected branch
	if err := s.repo.UpdateSessionBranch(ctx, sessionToken, req.BranchID); err != nil {
		return nil, err
	}

	return &domain.SelectBranchResponse{
		BranchID:      branch.ID,
		BranchName:    branch.BranchName,
		IsShiftOpened: branch.IsShiftOpened,
	}, nil
}

func (s *shiftService) OpenShift(ctx context.Context, sessionToken string, storeID int64, branchID int64, staffID *int64, req *domain.OpenShiftRequest) (*domain.OpenShiftResponse, error) {
	// Verify branch exists and belongs to store
	branch, err := s.repo.GetBranchByID(ctx, branchID)
	if err != nil {
		return nil, err
	}
	if branch == nil {
		return nil, errors.New("branch not found")
	}
	if branch.StoreID != storeID {
		return nil, errors.New("branch does not belong to this store")
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
	if err := s.repo.UpdateBranchShiftStatus(ctx, branchID, true); err != nil {
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

func (s *shiftService) GetCurrentShift(ctx context.Context, branchID int64) (*domain.CurrentShiftResponse, error) {
	branch, err := s.repo.GetBranchByID(ctx, branchID)
	if err != nil {
		return nil, err
	}
	if branch == nil {
		return nil, errors.New("branch not found")
	}

	shift, err := s.repo.GetActiveShiftByBranch(ctx, branchID)
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
