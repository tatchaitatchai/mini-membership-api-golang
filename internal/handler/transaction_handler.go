package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mini-membership/api/internal/domain"
	"github.com/mini-membership/api/internal/middleware"
	"github.com/mini-membership/api/internal/service"
)

type TransactionHandler struct {
	transactionService service.TransactionService
}

func NewTransactionHandler(transactionService service.TransactionService) *TransactionHandler {
	return &TransactionHandler{
		transactionService: transactionService,
	}
}

func (h *TransactionHandler) Create(c *gin.Context) {
	var req domain.TransactionCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	claims := middleware.GetClaims(c)

	transaction, err := h.transactionService.Create(
		c.Request.Context(),
		&req,
		claims.UserID,
		claims.Branch,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, transaction)
}

func (h *TransactionHandler) ListByMember(c *gin.Context) {
	memberID, err := uuid.Parse(c.Param("member_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid member id"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	claims := middleware.GetClaims(c)

	resp, err := h.transactionService.ListByMember(c.Request.Context(), memberID, claims.Branch, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *TransactionHandler) ListByBranch(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	claims := middleware.GetClaims(c)

	resp, err := h.transactionService.ListByBranch(c.Request.Context(), claims.Branch, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}
