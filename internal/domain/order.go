package domain

import "time"

type ProductInfo struct {
	ID           int64   `json:"id"`
	ProductName  string  `json:"product_name"`
	CategoryName *string `json:"category_name,omitempty"`
	BasePrice    float64 `json:"base_price"`
	ImagePath    *string `json:"image_path,omitempty"`
	OnStock      int     `json:"on_stock"`
}

type ListProductsResponse struct {
	Products []ProductInfo `json:"products"`
}

type CustomerInfo struct {
	ID           int64  `json:"id"`
	CustomerCode string `json:"customer_code"`
	FullName     string `json:"full_name"`
	PhoneLast4   string `json:"phone_last4"`
}

type SearchCustomersResponse struct {
	Customers []CustomerInfo `json:"customers"`
}

type OrderItemRequest struct {
	ProductID int64   `json:"product_id" binding:"required"`
	Quantity  int     `json:"quantity" binding:"required,min=1"`
	Price     float64 `json:"price" binding:"required,gte=0"`
}

type PaymentRequest struct {
	Method string  `json:"method" binding:"required,oneof=CASH TRANSFER QR CARD OTHER"`
	Amount float64 `json:"amount" binding:"required,gte=0"`
}

type CreateOrderRequest struct {
	CustomerID    *int64             `json:"customer_id"`
	Items         []OrderItemRequest `json:"items" binding:"required,min=1,dive"`
	Subtotal      float64            `json:"subtotal" binding:"required,gte=0"`
	DiscountTotal float64            `json:"discount_total" binding:"gte=0"`
	TotalPrice    float64            `json:"total_price" binding:"required,gte=0"`
	Payments      []PaymentRequest   `json:"payments" binding:"required,min=1,dive"`
	ChangeAmount  float64            `json:"change_amount" binding:"gte=0"`
	PromotionID   *int64             `json:"promotion_id"`
}

type CreateOrderResponse struct {
	OrderID      int64     `json:"order_id"`
	Status       string    `json:"status"`
	TotalPrice   float64   `json:"total_price"`
	ChangeAmount float64   `json:"change_amount"`
	CreatedAt    time.Time `json:"created_at"`
}

type OrderInfo struct {
	ID            int64           `json:"id"`
	CustomerID    *int64          `json:"customer_id,omitempty"`
	CustomerName  *string         `json:"customer_name,omitempty"`
	Subtotal      float64         `json:"subtotal"`
	DiscountTotal float64         `json:"discount_total"`
	TotalPrice    float64         `json:"total_price"`
	ChangeAmount  float64         `json:"change_amount"`
	Status        string          `json:"status"`
	CreatedAt     time.Time       `json:"created_at"`
	CreatedBy     string          `json:"created_by"`
	Items         []OrderItemInfo `json:"items,omitempty"`
}

type OrderItemInfo struct {
	ProductID   int64   `json:"product_id"`
	ProductName string  `json:"product_name"`
	Quantity    int     `json:"quantity"`
	Price       float64 `json:"price"`
	Total       float64 `json:"total"`
}

type ListOrdersResponse struct {
	Orders []OrderInfo `json:"orders"`
}
