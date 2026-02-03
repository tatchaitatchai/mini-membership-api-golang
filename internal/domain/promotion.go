package domain

import "time"

// PromotionType represents the type of promotion
type PromotionTypeInfo struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Detail string `json:"detail,omitempty"`
}

// PromotionConfig represents the configuration for a promotion
type PromotionConfig struct {
	PercentDiscount       *float64 `json:"percent_discount,omitempty"`
	BahtDiscount          *float64 `json:"baht_discount,omitempty"`
	TotalPriceSetDiscount *float64 `json:"total_price_set_discount,omitempty"`
	OldPriceSet           *float64 `json:"old_price_set,omitempty"`
	CountConditionProduct *int     `json:"count_condition_product,omitempty"`
}

// PromotionProduct represents a product linked to a promotion
type PromotionProduct struct {
	ProductID   int64   `json:"product_id"`
	ProductName string  `json:"product_name"`
	BasePrice   float64 `json:"base_price"`
}

// PromotionResponse represents a promotion with all its details
type PromotionResponse struct {
	ID             int64              `json:"id"`
	PromotionName  string             `json:"promotion_name"`
	PromotionType  PromotionTypeInfo  `json:"promotion_type"`
	Config         PromotionConfig    `json:"config"`
	Products       []PromotionProduct `json:"products"`
	IsBillLevel    bool               `json:"is_bill_level"`
	IsActive       bool               `json:"is_active"`
	StartsAt       *time.Time         `json:"starts_at,omitempty"`
	EndsAt         *time.Time         `json:"ends_at,omitempty"`
}

// CalculateDiscountRequest represents a request to calculate discount
type CalculateDiscountRequest struct {
	PromotionID int64                    `json:"promotion_id"`
	Items       []CalculateDiscountItem  `json:"items"`
	Subtotal    float64                  `json:"subtotal"`
}

type CalculateDiscountItem struct {
	ProductID int64   `json:"product_id"`
	Quantity  int     `json:"quantity"`
	UnitPrice float64 `json:"unit_price"`
}

// CalculateDiscountResponse represents the calculated discount
type CalculateDiscountResponse struct {
	PromotionID     int64   `json:"promotion_id"`
	PromotionName   string  `json:"promotion_name"`
	OriginalTotal   float64 `json:"original_total"`
	DiscountAmount  float64 `json:"discount_amount"`
	FinalTotal      float64 `json:"final_total"`
	IsApplicable    bool    `json:"is_applicable"`
	Message         string  `json:"message,omitempty"`
}

// DetectPromotionsRequest represents a request to detect applicable promotions
type DetectPromotionsRequest struct {
	Items []CalculateDiscountItem `json:"items"`
}

// DetectedPromotion represents a promotion that can be applied
type DetectedPromotion struct {
	PromotionID    int64   `json:"promotion_id"`
	PromotionName  string  `json:"promotion_name"`
	TypeName       string  `json:"type_name"`
	DiscountAmount float64 `json:"discount_amount"`
	FinalTotal     float64 `json:"final_total"`
	IsAutoApplied  bool    `json:"is_auto_applied"`
}
