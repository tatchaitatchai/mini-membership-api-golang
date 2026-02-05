package service

import (
	"context"
	"errors"

	"github.com/mini-membership/api/internal/domain"
	"github.com/mini-membership/api/internal/repository"
)

const DefaultLowStockThreshold = 10

type InventoryService interface {
	AdjustStock(ctx context.Context, storeID, branchID, staffID int64, req *domain.AdjustStockRequest) error
	GetMovements(ctx context.Context, storeID, branchID int64, limit, offset int) ([]domain.InventoryMovementResponse, error)
	GetLowStockItems(ctx context.Context, storeID, branchID int64) (*domain.LowStockResponse, error)
}

type inventoryService struct {
	repo repository.InventoryRepository
}

func NewInventoryService(repo repository.InventoryRepository) InventoryService {
	return &inventoryService{repo: repo}
}

func (s *inventoryService) AdjustStock(ctx context.Context, storeID, branchID, staffID int64, req *domain.AdjustStockRequest) error {
	if req.Quantity <= 0 {
		return errors.New("quantity must be positive")
	}

	// For ADJUST type, we reduce stock (negative quantity change)
	quantityChange := -req.Quantity

	return s.repo.AdjustStock(
		ctx,
		storeID,
		branchID,
		req.ProductID,
		quantityChange,
		domain.MovementTypeAdjust,
		&req.Reason,
		&req.Note,
		staffID,
	)
}

func (s *inventoryService) GetMovements(ctx context.Context, storeID, branchID int64, limit, offset int) ([]domain.InventoryMovementResponse, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return s.repo.GetMovementsByBranch(ctx, storeID, branchID, limit, offset)
}

func (s *inventoryService) GetLowStockItems(ctx context.Context, storeID, branchID int64) (*domain.LowStockResponse, error) {
	return s.repo.GetLowStockItems(ctx, storeID, branchID, DefaultLowStockThreshold)
}
