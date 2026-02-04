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
	shiftService   service.ShiftService
}

func NewOrderHandler(orderService service.OrderService, appAuthService service.AppAuthService, shiftService service.ShiftService) *OrderHandler {
	return &OrderHandler{
		orderService:   orderService,
		appAuthService: appAuthService,
		shiftService:   shiftService,
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
	currentShift, err := h.shiftService.GetCurrentShift(c.Request.Context(), sessionInfo.StoreID, *sessionInfo.BranchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !currentShift.HasActiveShift || currentShift.Shift == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no active shift, please open a shift first"})
		return
	}
	shiftID := currentShift.Shift.ID

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

	currentShift, err := h.shiftService.GetCurrentShift(c.Request.Context(), sessionInfo.StoreID, *sessionInfo.BranchID)
	if err != nil || currentShift == nil || !currentShift.HasActiveShift || currentShift.Shift == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no active shift"})
		return
	}

	resp, err := h.orderService.GetOrdersByShift(c.Request.Context(), sessionInfo.StoreID, *sessionInfo.BranchID, currentShift.Shift.ID)
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

func (h *OrderHandler) CancelOrder(c *gin.Context) {
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

	var req domain.CancelOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.orderService.CancelOrder(c.Request.Context(), sessionInfo.StoreID, orderID, req.Reason, sessionInfo.StaffID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}
