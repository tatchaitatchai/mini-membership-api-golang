package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/katom-membership/api/internal/domain"
	"github.com/katom-membership/api/internal/service"
)

type OrderHandler struct {
	orderService   service.OrderService
	appAuthService service.AppAuthService
}

func NewOrderHandler(orderService service.OrderService, appAuthService service.AppAuthService) *OrderHandler {
	return &OrderHandler{
		orderService:   orderService,
		appAuthService: appAuthService,
	}
}

func (h *OrderHandler) ListProducts(c *gin.Context) {
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

	resp, err := h.orderService.ListProducts(c.Request.Context(), sessionInfo.StoreID, *sessionInfo.BranchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *OrderHandler) SearchCustomers(c *gin.Context) {
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

	staffID := int64(0)
	branchID := int64(0)
	if sessionInfo.StaffID != nil {
		staffID = *sessionInfo.StaffID
	}
	if sessionInfo.BranchID != nil {
		branchID = *sessionInfo.BranchID
	}
	fmt.Printf("SearchCustomers - StoreID: %d, StaffID: %d, BranchID: %d\n", sessionInfo.StoreID, staffID, branchID)

	last4 := c.Query("last4")
	if len(last4) != 4 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "last4 must be exactly 4 characters"})
		return
	}

	resp, err := h.orderService.SearchCustomers(c.Request.Context(), sessionInfo.StoreID, last4)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *OrderHandler) CreateOrder(c *gin.Context) {
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

	var req domain.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get current shift
	shiftID := int64(0)
	// Note: In production, you should get the active shift from the shift service
	// For now, we'll use a placeholder or get it from session if available

	staffID := int64(0)
	if sessionInfo.StaffID != nil {
		staffID = *sessionInfo.StaffID
	}

	resp, err := h.orderService.CreateOrder(c.Request.Context(), sessionInfo.StoreID, *sessionInfo.BranchID, shiftID, staffID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

func (h *OrderHandler) GetOrdersByShift(c *gin.Context) {
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

	shiftIDStr := c.Query("shift_id")
	if shiftIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "shift_id is required"})
		return
	}

	shiftID, err := strconv.ParseInt(shiftIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid shift_id"})
		return
	}

	resp, err := h.orderService.GetOrdersByShift(c.Request.Context(), sessionInfo.StoreID, *sessionInfo.BranchID, shiftID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *OrderHandler) GetOrderByID(c *gin.Context) {
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

	orderIDStr := c.Param("id")
	orderID, err := strconv.ParseInt(orderIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}

	resp, err := h.orderService.GetOrderByID(c.Request.Context(), sessionInfo.StoreID, orderID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}
