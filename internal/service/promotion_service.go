package service

import (
	"context"
	"errors"

	"github.com/mini-membership/api/internal/domain"
	"github.com/mini-membership/api/internal/repository"
)

type PromotionService interface {
	GetActivePromotions(ctx context.Context, storeID, branchID int64) ([]domain.PromotionResponse, error)
	CalculateDiscount(ctx context.Context, storeID int64, req *domain.CalculateDiscountRequest) (*domain.CalculateDiscountResponse, error)
	DetectApplicablePromotions(ctx context.Context, storeID, branchID int64, req *domain.DetectPromotionsRequest) ([]domain.DetectedPromotion, error)
}

type promotionService struct {
	repo repository.PromotionRepository
}

func NewPromotionService(repo repository.PromotionRepository) PromotionService {
	return &promotionService{repo: repo}
}

func (s *promotionService) GetActivePromotions(ctx context.Context, storeID, branchID int64) ([]domain.PromotionResponse, error) {
	return s.repo.GetActivePromotions(ctx, storeID, branchID)
}

func (s *promotionService) CalculateDiscount(ctx context.Context, storeID int64, req *domain.CalculateDiscountRequest) (*domain.CalculateDiscountResponse, error) {
	promo, err := s.repo.GetPromotionByID(ctx, storeID, req.PromotionID)
	if err != nil {
		return nil, err
	}
	if promo == nil {
		return nil, errors.New("promotion not found")
	}

	// Calculate subtotal from items
	subtotal := req.Subtotal
	if subtotal == 0 {
		for _, item := range req.Items {
			subtotal += item.UnitPrice * float64(item.Quantity)
		}
	}

	response := &domain.CalculateDiscountResponse{
		PromotionID:   promo.ID,
		PromotionName: promo.PromotionName,
		OriginalTotal: subtotal,
		IsApplicable:  true,
	}

	// Calculate discount based on promotion type
	discountAmount := s.calculateDiscountAmount(promo, req.Items, subtotal)

	response.DiscountAmount = discountAmount
	response.FinalTotal = subtotal - discountAmount
	if response.FinalTotal < 0 {
		response.FinalTotal = 0
	}

	return response, nil
}

func (s *promotionService) calculateDiscountAmount(promo *domain.PromotionResponse, items []domain.CalculateDiscountItem, subtotal float64) float64 {
	typeName := promo.PromotionType.Name

	switch typeName {
	case "ลดเปอร์เซ็นต์":
		// Simple percent discount
		if promo.Config.PercentDiscount != nil {
			if promo.IsBillLevel {
				// Bill-level: apply to entire subtotal
				return subtotal * (*promo.Config.PercentDiscount / 100)
			}
			// Product-level: apply only to matching products
			return s.calculateProductDiscount(promo, items, *promo.Config.PercentDiscount, true)
		}

	case "ลดบาท":
		// Simple baht discount
		if promo.Config.BahtDiscount != nil {
			if promo.IsBillLevel {
				// Bill-level: apply to entire bill
				return *promo.Config.BahtDiscount
			}
			// Product-level: apply per matching product
			return s.calculateProductDiscount(promo, items, *promo.Config.BahtDiscount, false)
		}

	case "ซื้อเป็นเซ็ต":
		// Set discount - check if all required products are in cart
		if promo.Config.TotalPriceSetDiscount != nil && promo.Config.OldPriceSet != nil {
			if s.hasAllSetProducts(promo, items) {
				return *promo.Config.OldPriceSet - *promo.Config.TotalPriceSetDiscount
			}
		}

	case "ซื้อครบลดเปอร์เซ็นต์":
		// Buy N items get percent off
		if promo.Config.CountConditionProduct != nil && promo.Config.PercentDiscount != nil {
			matchingQty := s.countMatchingProducts(promo, items)
			if matchingQty >= *promo.Config.CountConditionProduct {
				matchingTotal := s.getMatchingProductsTotal(promo, items)
				return matchingTotal * (*promo.Config.PercentDiscount / 100)
			}
		}

	case "ซื้อครบลดบาท":
		// Buy N items get baht off
		if promo.Config.CountConditionProduct != nil && promo.Config.BahtDiscount != nil {
			matchingQty := s.countMatchingProducts(promo, items)
			if matchingQty >= *promo.Config.CountConditionProduct {
				return *promo.Config.BahtDiscount
			}
		}
	}

	return 0
}

func (s *promotionService) calculateProductDiscount(promo *domain.PromotionResponse, items []domain.CalculateDiscountItem, discountValue float64, isPercent bool) float64 {
	productIDs := make(map[int64]bool)
	for _, p := range promo.Products {
		productIDs[p.ProductID] = true
	}

	var totalDiscount float64
	for _, item := range items {
		if productIDs[item.ProductID] {
			itemTotal := item.UnitPrice * float64(item.Quantity)
			if isPercent {
				totalDiscount += itemTotal * (discountValue / 100)
			} else {
				totalDiscount += discountValue * float64(item.Quantity)
			}
		}
	}
	return totalDiscount
}

func (s *promotionService) hasAllSetProducts(promo *domain.PromotionResponse, items []domain.CalculateDiscountItem) bool {
	requiredProducts := make(map[int64]bool)
	for _, p := range promo.Products {
		requiredProducts[p.ProductID] = true
	}

	for _, item := range items {
		if item.Quantity > 0 {
			delete(requiredProducts, item.ProductID)
		}
	}

	return len(requiredProducts) == 0
}

func (s *promotionService) countMatchingProducts(promo *domain.PromotionResponse, items []domain.CalculateDiscountItem) int {
	productIDs := make(map[int64]bool)
	for _, p := range promo.Products {
		productIDs[p.ProductID] = true
	}

	count := 0
	for _, item := range items {
		if productIDs[item.ProductID] {
			count += item.Quantity
		}
	}
	return count
}

func (s *promotionService) getMatchingProductsTotal(promo *domain.PromotionResponse, items []domain.CalculateDiscountItem) float64 {
	productIDs := make(map[int64]bool)
	for _, p := range promo.Products {
		productIDs[p.ProductID] = true
	}

	var total float64
	for _, item := range items {
		if productIDs[item.ProductID] {
			total += item.UnitPrice * float64(item.Quantity)
		}
	}
	return total
}

func (s *promotionService) DetectApplicablePromotions(ctx context.Context, storeID, branchID int64, req *domain.DetectPromotionsRequest) ([]domain.DetectedPromotion, error) {
	promotions, err := s.repo.GetActivePromotions(ctx, storeID, branchID)
	if err != nil {
		return nil, err
	}

	// Calculate subtotal
	var subtotal float64
	for _, item := range req.Items {
		subtotal += item.UnitPrice * float64(item.Quantity)
	}

	var detected []domain.DetectedPromotion
	for _, promo := range promotions {
		discount := s.calculateDiscountAmount(&promo, req.Items, subtotal)
		if discount > 0 {
			detected = append(detected, domain.DetectedPromotion{
				PromotionID:    promo.ID,
				PromotionName:  promo.PromotionName,
				TypeName:       promo.PromotionType.Name,
				DiscountAmount: discount,
				FinalTotal:     subtotal - discount,
				IsAutoApplied:  promo.PromotionType.Name == "ซื้อเป็นเซ็ต",
			})
		}
	}

	return detected, nil
}
