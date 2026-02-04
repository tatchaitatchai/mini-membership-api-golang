package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/katom-membership/api/internal/domain"
	"github.com/katom-membership/api/internal/repository"
)

type PointsService interface {
	GetCustomerPoints(ctx context.Context, storeID, customerID int64, customerName, customerCode string) (*domain.GetCustomerPointsResponse, error)
	GetRedeemableProducts(ctx context.Context, storeID, branchID int64) (*domain.ListRedeemableProductsResponse, error)
	RedeemPoints(ctx context.Context, storeID, branchID int64, req *domain.RedeemPointsRequest, staffID *int64) (*domain.RedeemPointsResponse, error)
	EarnPointsFromOrder(ctx context.Context, storeID, branchID int64, customerID int64, orderID int64, items []domain.OrderItemForPoints, staffID *int64) (*domain.EarnPointsResponse, error)
	GetPointHistory(ctx context.Context, storeID, customerID int64, page, limit int) (*domain.GetPointHistoryResponse, error)
}

type pointsService struct {
	pointsRepo repository.PointsRepository
	orderRepo  repository.OrderRepository
}

func NewPointsService(pointsRepo repository.PointsRepository, orderRepo repository.OrderRepository) PointsService {
	return &pointsService{
		pointsRepo: pointsRepo,
		orderRepo:  orderRepo,
	}
}

func (s *pointsService) GetCustomerPoints(ctx context.Context, storeID, customerID int64, customerName, customerCode string) (*domain.GetCustomerPointsResponse, error) {
	products, err := s.pointsRepo.GetCustomerProductPoints(ctx, storeID, customerID)
	if err != nil {
		return nil, err
	}

	return &domain.GetCustomerPointsResponse{
		CustomerID:   customerID,
		CustomerName: customerName,
		CustomerCode: customerCode,
		Products:     products,
	}, nil
}

func (s *pointsService) GetRedeemableProducts(ctx context.Context, storeID, branchID int64) (*domain.ListRedeemableProductsResponse, error) {
	products, err := s.pointsRepo.GetRedeemableProducts(ctx, storeID, branchID)
	if err != nil {
		return nil, err
	}

	return &domain.ListRedeemableProductsResponse{
		Products: products,
	}, nil
}

func (s *pointsService) RedeemPoints(ctx context.Context, storeID, branchID int64, req *domain.RedeemPointsRequest, staffID *int64) (*domain.RedeemPointsResponse, error) {
	// Get product points requirement
	pointsRequired, err := s.pointsRepo.GetProductPointsToRedeem(ctx, req.ProductID)
	if err != nil {
		return nil, err
	}

	totalPointsNeeded := *pointsRequired * req.Quantity

	// Check customer has enough points for THIS PRODUCT
	productPoints, err := s.pointsRepo.GetProductPoints(ctx, storeID, req.CustomerID, req.ProductID)
	if err != nil {
		return nil, err
	}
	if productPoints == nil || productPoints.Points < totalPointsNeeded {
		currentPoints := 0
		if productPoints != nil {
			currentPoints = productPoints.Points
		}
		return nil, fmt.Errorf("แต้มไม่เพียงพอ: ต้องการ %d แต้ม, มี %d แต้ม", totalPointsNeeded, currentPoints)
	}

	// Deduct points for this product
	err = s.pointsRepo.DeductProductPoints(ctx, storeID, req.CustomerID, req.ProductID, totalPointsNeeded)
	if err != nil {
		return nil, err
	}

	// Create redemption record
	redemption := &domain.PointRedemption{
		StoreID:    storeID,
		BranchID:   branchID,
		CustomerID: req.CustomerID,
		ProductID:  req.ProductID,
		PointsUsed: totalPointsNeeded,
		Quantity:   req.Quantity,
		StaffID:    staffID,
	}
	redemptionID, err := s.pointsRepo.CreateRedemption(ctx, redemption)
	if err != nil {
		return nil, err
	}

	// Create point transaction record
	refTable := "point_redemptions"
	pointChange := -totalPointsNeeded
	ptx := &domain.PointTransaction{
		StoreID:         storeID,
		BranchID:        branchID,
		CustomerID:      req.CustomerID,
		TransactionType: "REDEEM",
		PointsChange:    pointChange,
		ReferenceTable:  &refTable,
		ReferenceID:     &redemptionID,
		ProductID:       &req.ProductID,
		StaffID:         staffID,
	}
	err = s.pointsRepo.CreatePointTransaction(ctx, ptx)
	if err != nil {
		fmt.Printf("Failed to create point transaction: %v\n", err)
	}

	// Get product name for response
	products, _ := s.pointsRepo.GetRedeemableProducts(ctx, storeID, branchID)
	productName := "Unknown Product"
	for _, p := range products {
		if p.ID == req.ProductID {
			productName = p.ProductName
			break
		}
	}

	// Get updated points for this product
	updatedPoints, _ := s.pointsRepo.GetProductPoints(ctx, storeID, req.CustomerID, req.ProductID)
	remainingPoints := 0
	if updatedPoints != nil {
		remainingPoints = updatedPoints.Points
	}

	return &domain.RedeemPointsResponse{
		RedemptionID:    redemptionID,
		PointsUsed:      totalPointsNeeded,
		RemainingPoints: remainingPoints,
		ProductName:     productName,
		Quantity:        req.Quantity,
		Message:         fmt.Sprintf("แลกสำเร็จ: %s x%d ใช้ %d แต้ม", productName, req.Quantity, totalPointsNeeded),
	}, nil
}

func (s *pointsService) EarnPointsFromOrder(ctx context.Context, storeID, branchID int64, customerID int64, orderID int64, items []domain.OrderItemForPoints, staffID *int64) (*domain.EarnPointsResponse, error) {
	if customerID <= 0 {
		return nil, errors.New("customer ID is required for earning points")
	}
	if len(items) == 0 {
		return nil, errors.New("no items to earn points from")
	}

	totalPointsEarned := 0
	refTable := "orders"

	// Add points for each product purchased
	for _, item := range items {
		// Each item = 1 point for that product
		pointsToEarn := item.Quantity
		totalPointsEarned += pointsToEarn

		// Add points for this specific product
		err := s.pointsRepo.CreateOrUpdateProductPoints(ctx, storeID, customerID, item.ProductID, pointsToEarn)
		if err != nil {
			fmt.Printf("Failed to add points for product %d: %v\n", item.ProductID, err)
			continue
		}

		// Create point transaction record for each product
		ptx := &domain.PointTransaction{
			StoreID:         storeID,
			BranchID:        branchID,
			CustomerID:      customerID,
			TransactionType: "EARN",
			PointsChange:    pointsToEarn,
			ReferenceTable:  &refTable,
			ReferenceID:     &orderID,
			ProductID:       &item.ProductID,
			StaffID:         staffID,
		}
		err = s.pointsRepo.CreatePointTransaction(ctx, ptx)
		if err != nil {
			fmt.Printf("Failed to create point transaction for product %d: %v\n", item.ProductID, err)
		}
	}

	return &domain.EarnPointsResponse{
		PointsEarned:    totalPointsEarned,
		TotalPoints:     totalPointsEarned,
		RemainingPoints: totalPointsEarned,
	}, nil
}

func (s *pointsService) GetPointHistory(ctx context.Context, storeID, customerID int64, page, limit int) (*domain.GetPointHistoryResponse, error) {
	offset := (page - 1) * limit
	history, total, err := s.pointsRepo.GetPointHistory(ctx, storeID, customerID, limit, offset)
	if err != nil {
		return nil, err
	}

	return &domain.GetPointHistoryResponse{
		CustomerID: customerID,
		History:    history,
		Total:      total,
	}, nil
}
