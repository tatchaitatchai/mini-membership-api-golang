package service

import (
	"context"
	"errors"

	"github.com/katom-membership/api/internal/domain"
	"github.com/katom-membership/api/internal/repository"
	"github.com/shopspring/decimal"
)

type OrderService interface {
	ListProducts(ctx context.Context, storeID, branchID int64) (*domain.ListProductsResponse, error)
	SearchCustomers(ctx context.Context, storeID int64, last4 string) (*domain.SearchCustomersResponse, error)
	CreateOrder(ctx context.Context, storeID, branchID, shiftID, staffID int64, req *domain.CreateOrderRequest) (*domain.CreateOrderResponse, error)
	GetOrdersByShift(ctx context.Context, storeID, branchID, shiftID int64) (*domain.ListOrdersResponse, error)
	GetOrderByID(ctx context.Context, storeID, orderID int64) (*domain.OrderInfo, error)
	CancelOrder(ctx context.Context, storeID, orderID int64, reason string, cancelledBy *int64) (*domain.CancelOrderResponse, error)
}

type orderService struct {
	repo repository.OrderRepository
}

func NewOrderService(repo repository.OrderRepository) OrderService {
	return &orderService{repo: repo}
}

func (s *orderService) ListProducts(ctx context.Context, storeID, branchID int64) (*domain.ListProductsResponse, error) {
	products, err := s.repo.GetProductsByBranch(ctx, storeID, branchID)
	if err != nil {
		return nil, err
	}

	result := make([]domain.ProductInfo, len(products))
	for i, p := range products {
		price, _ := p.BasePrice.Float64()
		result[i] = domain.ProductInfo{
			ID:          p.ProductID,
			ProductName: p.ProductName,
			BasePrice:   price,
			OnStock:     p.OnStock,
		}
		if p.CategoryName.Valid {
			result[i].CategoryName = &p.CategoryName.String
		}
		if p.ImagePath.Valid {
			result[i].ImagePath = &p.ImagePath.String
		}
	}

	return &domain.ListProductsResponse{Products: result}, nil
}

func (s *orderService) SearchCustomers(ctx context.Context, storeID int64, last4 string) (*domain.SearchCustomersResponse, error) {
	if len(last4) != 4 {
		return nil, errors.New("last4 must be exactly 4 characters")
	}

	customers, err := s.repo.SearchCustomersByLast4(ctx, storeID, last4)
	if err != nil {
		return nil, err
	}

	result := make([]domain.CustomerInfo, len(customers))
	for i, c := range customers {
		result[i] = domain.CustomerInfo{
			ID: c.ID,
		}
		if c.CustomerCode.Valid {
			result[i].CustomerCode = c.CustomerCode.String
		}
		if c.FullName.Valid {
			result[i].FullName = c.FullName.String
		}
		if c.PhoneLast4.Valid {
			result[i].PhoneLast4 = c.PhoneLast4.String
		}
	}

	return &domain.SearchCustomersResponse{Customers: result}, nil
}

func (s *orderService) CreateOrder(ctx context.Context, storeID, branchID, shiftID, staffID int64, req *domain.CreateOrderRequest) (*domain.CreateOrderResponse, error) {
	if len(req.Items) == 0 {
		return nil, errors.New("order must have at least one item")
	}

	if len(req.Payments) == 0 {
		return nil, errors.New("order must have at least one payment")
	}

	// Validate total payment amount
	var totalPayment decimal.Decimal
	for _, p := range req.Payments {
		totalPayment = totalPayment.Add(decimal.NewFromFloat(p.Amount))
	}
	totalPrice := decimal.NewFromFloat(req.TotalPrice)
	if totalPayment.LessThan(totalPrice) {
		return nil, errors.New("payment amount is less than total price")
	}

	// Build order create struct
	items := make([]repository.OrderItemCreate, len(req.Items))
	for i, item := range req.Items {
		items[i] = repository.OrderItemCreate{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     decimal.NewFromFloat(item.Price),
		}
	}

	payments := make([]repository.PaymentCreate, len(req.Payments))
	for i, p := range req.Payments {
		payments[i] = repository.PaymentCreate{
			Method: p.Method,
			Amount: decimal.NewFromFloat(p.Amount),
		}
	}

	order := &repository.OrderCreate{
		StoreID:       storeID,
		BranchID:      branchID,
		ShiftID:       shiftID,
		StaffID:       staffID,
		CustomerID:    req.CustomerID,
		Items:         items,
		Subtotal:      decimal.NewFromFloat(req.Subtotal),
		DiscountTotal: decimal.NewFromFloat(req.DiscountTotal),
		TotalPrice:    totalPrice,
		ChangeAmount:  decimal.NewFromFloat(req.ChangeAmount),
		Payments:      payments,
		PromotionID:   req.PromotionID,
	}

	result, err := s.repo.CreateOrderTx(ctx, order)
	if err != nil {
		return nil, err
	}

	return &domain.CreateOrderResponse{
		OrderID:      result.OrderID,
		Status:       result.Status,
		TotalPrice:   req.TotalPrice,
		ChangeAmount: req.ChangeAmount,
		CreatedAt:    result.CreatedAt,
	}, nil
}

func (s *orderService) GetOrdersByShift(ctx context.Context, storeID, branchID, shiftID int64) (*domain.ListOrdersResponse, error) {
	orders, err := s.repo.GetOrdersByShift(ctx, storeID, branchID, shiftID)
	if err != nil {
		return nil, err
	}

	result := make([]domain.OrderInfo, len(orders))
	for i, o := range orders {
		subtotal, _ := o.Subtotal.Float64()
		discountTotal, _ := o.DiscountTotal.Float64()
		totalPrice, _ := o.TotalPrice.Float64()
		changeAmount, _ := o.ChangeAmount.Float64()

		result[i] = domain.OrderInfo{
			ID:            o.ID,
			Subtotal:      subtotal,
			DiscountTotal: discountTotal,
			TotalPrice:    totalPrice,
			ChangeAmount:  changeAmount,
			Status:        o.Status,
			CreatedAt:     o.CreatedAt,
		}
		if o.CustomerID.Valid {
			result[i].CustomerID = &o.CustomerID.Int64
		}
		if o.CustomerName.Valid {
			result[i].CustomerName = &o.CustomerName.String
		}
		if o.StaffName.Valid {
			result[i].CreatedBy = o.StaffName.String
		}

		items := make([]domain.OrderItemInfo, len(o.Items))
		for j, item := range o.Items {
			price, _ := item.Price.Float64()
			items[j] = domain.OrderItemInfo{
				ProductID:   item.ProductID,
				ProductName: item.ProductName,
				Quantity:    item.Quantity,
				Price:       price,
				Total:       price * float64(item.Quantity),
			}
		}
		result[i].Items = items
	}

	return &domain.ListOrdersResponse{Orders: result}, nil
}

func (s *orderService) GetOrderByID(ctx context.Context, storeID, orderID int64) (*domain.OrderInfo, error) {
	o, err := s.repo.GetOrderByID(ctx, storeID, orderID)
	if err != nil {
		return nil, err
	}
	if o == nil {
		return nil, errors.New("order not found")
	}

	subtotal, _ := o.Subtotal.Float64()
	discountTotal, _ := o.DiscountTotal.Float64()
	totalPrice, _ := o.TotalPrice.Float64()
	changeAmount, _ := o.ChangeAmount.Float64()

	result := &domain.OrderInfo{
		ID:            o.ID,
		Subtotal:      subtotal,
		DiscountTotal: discountTotal,
		TotalPrice:    totalPrice,
		ChangeAmount:  changeAmount,
		Status:        o.Status,
		CreatedAt:     o.CreatedAt,
	}
	if o.CustomerID.Valid {
		result.CustomerID = &o.CustomerID.Int64
	}
	if o.CustomerName.Valid {
		result.CustomerName = &o.CustomerName.String
	}
	if o.StaffName.Valid {
		result.CreatedBy = o.StaffName.String
	}

	items := make([]domain.OrderItemInfo, len(o.Items))
	for j, item := range o.Items {
		price, _ := item.Price.Float64()
		items[j] = domain.OrderItemInfo{
			ProductID:   item.ProductID,
			ProductName: item.ProductName,
			Quantity:    item.Quantity,
			Price:       price,
			Total:       price * float64(item.Quantity),
		}
	}
	result.Items = items

	return result, nil
}

func (s *orderService) CancelOrder(ctx context.Context, storeID, orderID int64, reason string, cancelledBy *int64) (*domain.CancelOrderResponse, error) {
	err := s.repo.CancelOrder(ctx, storeID, orderID, reason, cancelledBy)
	if err != nil {
		return nil, err
	}
	return &domain.CancelOrderResponse{
		OrderID: orderID,
		Status:  "CANCELLED",
		Message: "Order cancelled successfully",
	}, nil
}
