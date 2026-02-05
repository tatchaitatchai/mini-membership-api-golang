package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mini-membership/api/internal/domain"
	"github.com/mini-membership/api/internal/service"
)

type InventoryHandler struct {
	inventoryService service.InventoryService
	appAuthService   service.AppAuthService
}

func NewInventoryHandler(inventoryService service.InventoryService, appAuthService service.AppAuthService) *InventoryHandler {
	return &InventoryHandler{
		inventoryService: inventoryService,
		appAuthService:   appAuthService,
	}
}

// AdjustStock handles stock adjustment requests
func (h *InventoryHandler) AdjustStock(c *gin.Context) {
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

	if sessionInfo.BranchID == nil || sessionInfo.StaffID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "please select a branch first"})
		return
	}

	var req domain.AdjustStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.inventoryService.AdjustStock(c.Request.Context(), sessionInfo.StoreID, *sessionInfo.BranchID, *sessionInfo.StaffID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "stock adjusted successfully"})
}

// GetMovements returns inventory movement history
func (h *InventoryHandler) GetMovements(c *gin.Context) {
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

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	movements, err := h.inventoryService.GetMovements(c.Request.Context(), sessionInfo.StoreID, *sessionInfo.BranchID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, movements)
}

// GetLowStockItems returns products with low stock
func (h *InventoryHandler) GetLowStockItems(c *gin.Context) {
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

	response, err := h.inventoryService.GetLowStockItems(c.Request.Context(), sessionInfo.StoreID, *sessionInfo.BranchID)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}
