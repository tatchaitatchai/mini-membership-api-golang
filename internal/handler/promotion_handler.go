package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/katom-membership/api/internal/domain"
	"github.com/katom-membership/api/internal/service"
)

type PromotionHandler struct {
	promotionService service.PromotionService
	appAuthService   service.AppAuthService
}

func NewPromotionHandler(promotionService service.PromotionService, appAuthService service.AppAuthService) *PromotionHandler {
	return &PromotionHandler{
		promotionService: promotionService,
		appAuthService:   appAuthService,
	}
}

// GetActivePromotions returns all active promotions for the current branch
func (h *PromotionHandler) GetActivePromotions(c *gin.Context) {
	token := extractBearerToken(c)
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "session token required"})
		return
	}

	sessionInfo, err := h.appAuthService.ValidateSession(c.Request.Context(), token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	if sessionInfo.BranchID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "please select a branch first"})
		return
	}

	promotions, err := h.promotionService.GetActivePromotions(c.Request.Context(), sessionInfo.StoreID, *sessionInfo.BranchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, promotions)
}

// CalculateDiscount calculates the discount for a given promotion and items
func (h *PromotionHandler) CalculateDiscount(c *gin.Context) {
	token := extractBearerToken(c)
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "session token required"})
		return
	}

	sessionInfo, err := h.appAuthService.ValidateSession(c.Request.Context(), token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var req domain.CalculateDiscountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.promotionService.CalculateDiscount(c.Request.Context(), sessionInfo.StoreID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// DetectPromotions detects applicable promotions based on cart items
func (h *PromotionHandler) DetectPromotions(c *gin.Context) {
	token := extractBearerToken(c)
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "session token required"})
		return
	}

	sessionInfo, err := h.appAuthService.ValidateSession(c.Request.Context(), token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	if sessionInfo.BranchID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "please select a branch first"})
		return
	}

	var req domain.DetectPromotionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	detected, err := h.promotionService.DetectApplicablePromotions(c.Request.Context(), sessionInfo.StoreID, *sessionInfo.BranchID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, detected)
}
