package service

import (
	"context"
	"errors"

	"github.com/mini-membership/api/internal/domain"
	"github.com/mini-membership/api/internal/repository"
)

type StockTransferService interface {
	CreateTransfer(ctx context.Context, storeID, branchID, staffID int64, req *domain.CreateStockTransferRequest) (*domain.StockTransferResponse, error)
	WithdrawGoods(ctx context.Context, storeID, branchID, staffID int64, req *domain.WithdrawGoodsRequest) (*domain.StockTransferResponse, error)
	GetTransferByID(ctx context.Context, storeID, transferID int64) (*domain.StockTransferResponse, error)
	GetTransfersByBranch(ctx context.Context, storeID, branchID int64, limit, offset int) (*domain.StockTransferListResponse, error)
	GetPendingTransfers(ctx context.Context, storeID, branchID int64) ([]domain.StockTransferResponse, error)
	ReceiveTransfer(ctx context.Context, storeID, branchID, transferID, staffID int64, items []domain.UpdateStockTransferItemInput) error
	CancelTransfer(ctx context.Context, storeID, transferID int64) error
}

type stockTransferService struct {
	repo repository.StockTransferRepository
}

func NewStockTransferService(repo repository.StockTransferRepository) StockTransferService {
	return &stockTransferService{repo: repo}
}

func (s *stockTransferService) CreateTransfer(ctx context.Context, storeID, branchID, staffID int64, req *domain.CreateStockTransferRequest) (*domain.StockTransferResponse, error) {
	if len(req.Items) == 0 {
		return nil, errors.New("at least one item is required")
	}

	// Set from_branch_id to current branch if not specified
	if req.FromBranchID == nil {
		req.FromBranchID = &branchID
	}

	transfer, err := s.repo.Create(ctx, storeID, req, staffID)
	if err != nil {
		return nil, err
	}

	return s.repo.GetByID(ctx, storeID, transfer.ID)
}

func (s *stockTransferService) WithdrawGoods(ctx context.Context, storeID, branchID, staffID int64, req *domain.WithdrawGoodsRequest) (*domain.StockTransferResponse, error) {
	if len(req.Items) == 0 {
		return nil, errors.New("at least one item is required")
	}

	// Withdraw = Request goods FROM central/other branch TO current branch
	// from_branch_id = NULL means from central warehouse
	// to_branch_id = current branch (requesting branch)
	transferReq := &domain.CreateStockTransferRequest{
		FromBranchID: nil, // From central (NULL)
		ToBranchID:   branchID,
		Note:         req.Note,
		Items:        make([]domain.CreateStockTransferItemInput, len(req.Items)),
	}

	for i, item := range req.Items {
		transferReq.Items[i] = domain.CreateStockTransferItemInput{
			ProductID: item.ProductID,
			SendCount: item.Quantity,
		}
	}

	// Create transfer request with status CREATED
	// Stock is NOT reduced - central will process and send goods
	transfer, err := s.repo.Create(ctx, storeID, transferReq, staffID)
	if err != nil {
		return nil, err
	}

	return s.repo.GetByID(ctx, storeID, transfer.ID)
}

func (s *stockTransferService) GetTransferByID(ctx context.Context, storeID, transferID int64) (*domain.StockTransferResponse, error) {
	return s.repo.GetByID(ctx, storeID, transferID)
}

func (s *stockTransferService) GetTransfersByBranch(ctx context.Context, storeID, branchID int64, limit, offset int) (*domain.StockTransferListResponse, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return s.repo.GetByBranch(ctx, storeID, branchID, limit, offset)
}

func (s *stockTransferService) GetPendingTransfers(ctx context.Context, storeID, branchID int64) ([]domain.StockTransferResponse, error) {
	return s.repo.GetPendingTransfers(ctx, storeID, branchID)
}

func (s *stockTransferService) ReceiveTransfer(ctx context.Context, storeID, branchID, transferID, staffID int64, items []domain.UpdateStockTransferItemInput) error {
	// Get current transfer
	transfer, err := s.repo.GetByID(ctx, storeID, transferID)
	if err != nil {
		return err
	}
	if transfer == nil {
		return errors.New("transfer not found")
	}

	if transfer.Status != domain.StockTransferStatusSent {
		return errors.New("transfer is not in SENT status")
	}

	if len(items) == 0 {
		return errors.New("at least one item is required")
	}

	// Receive goods and add stock in a single transaction
	return s.repo.ReceiveAndAddStock(ctx, storeID, branchID, transferID, items, staffID)
}

func (s *stockTransferService) CancelTransfer(ctx context.Context, storeID, transferID int64) error {
	transfer, err := s.repo.GetByID(ctx, storeID, transferID)
	if err != nil {
		return err
	}
	if transfer == nil {
		return errors.New("transfer not found")
	}

	if transfer.Status == domain.StockTransferStatusReceived {
		return errors.New("cannot cancel received transfer")
	}

	return s.repo.UpdateStatus(ctx, storeID, transferID, domain.StockTransferStatusCancelled, nil)
}
