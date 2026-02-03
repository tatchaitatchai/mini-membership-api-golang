package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/katom-membership/api/internal/domain"
	"github.com/katom-membership/api/internal/service"
)

type StockTransferHandler struct {
	stockTransferService service.StockTransferService
	appAuthService       service.AppAuthService
}

func NewStockTransferHandler(stockTransferService service.StockTransferService, appAuthService service.AppAuthService) *StockTransferHandler {
	return &StockTransferHandler{
		stockTransferService: stockTransferService,
		appAuthService:       appAuthService,
	}
}

// CreateTransfer creates a new stock transfer between branches
func (h *StockTransferHandler) CreateTransfer(c *gin.Context) {
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

	var req domain.CreateStockTransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	transfer, err := h.stockTransferService.CreateTransfer(c.Request.Context(), sessionInfo.StoreID, *sessionInfo.BranchID, *sessionInfo.StaffID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, transfer)
}

// WithdrawGoods withdraws goods from current branch (simplified transfer)
func (h *StockTransferHandler) WithdrawGoods(c *gin.Context) {
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

	var req domain.WithdrawGoodsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	transfer, err := h.stockTransferService.WithdrawGoods(c.Request.Context(), sessionInfo.StoreID, *sessionInfo.BranchID, *sessionInfo.StaffID, &req)
	if err != nil {

		fmt.Println("error :: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	fmt.Println("transfer :: ", transfer)

	c.JSON(http.StatusCreated, transfer)
}

// GetTransfer gets a specific stock transfer by ID
func (h *StockTransferHandler) GetTransfer(c *gin.Context) {
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

	transferID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid transfer ID"})
		return
	}

	transfer, err := h.stockTransferService.GetTransferByID(c.Request.Context(), sessionInfo.StoreID, transferID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if transfer == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "transfer not found"})
		return
	}

	c.JSON(http.StatusOK, transfer)
}

// GetTransfers gets all stock transfers for current branch
func (h *StockTransferHandler) GetTransfers(c *gin.Context) {
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

	transfers, err := h.stockTransferService.GetTransfersByBranch(c.Request.Context(), sessionInfo.StoreID, *sessionInfo.BranchID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, transfers)
}

// ReceiveTransfer marks a transfer as received and adds stock
func (h *StockTransferHandler) ReceiveTransfer(c *gin.Context) {
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

	transferID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid transfer ID"})
		return
	}

	var req domain.UpdateStockTransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.stockTransferService.ReceiveTransfer(c.Request.Context(), sessionInfo.StoreID, *sessionInfo.BranchID, transferID, *sessionInfo.StaffID, req.Items)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "transfer received successfully"})
}

// GetPendingTransfers gets transfers waiting to be received
func (h *StockTransferHandler) GetPendingTransfers(c *gin.Context) {
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

	transfers, err := h.stockTransferService.GetPendingTransfers(c.Request.Context(), sessionInfo.StoreID, *sessionInfo.BranchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, transfers)
}

// CancelTransfer cancels a stock transfer
func (h *StockTransferHandler) CancelTransfer(c *gin.Context) {
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

	transferID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid transfer ID"})
		return
	}

	err = h.stockTransferService.CancelTransfer(c.Request.Context(), sessionInfo.StoreID, transferID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "transfer cancelled successfully"})
}
