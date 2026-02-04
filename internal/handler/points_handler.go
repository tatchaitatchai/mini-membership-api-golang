package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/katom-membership/api/internal/domain"
	"github.com/katom-membership/api/internal/service"
)

type PointsHandler struct {
	pointsService  service.PointsService
	appAuthService service.AppAuthService
}

func NewPointsHandler(pointsService service.PointsService, appAuthService service.AppAuthService) *PointsHandler {
	return &PointsHandler{
		pointsService:  pointsService,
		appAuthService: appAuthService,
	}
}

func (h *PointsHandler) GetCustomerPoints(c *gin.Context) {
	token := extractBearerToken(c)
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "session token required"})
		return
	}

	session, err := h.appAuthService.ValidateSession(c.Request.Context(), token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid session"})
		return
	}

	customerID, err := strconv.ParseInt(c.Param("customer_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid customer_id"})
		return
	}

	customerName := c.Query("name")
	customerCode := c.Query("code")

	result, err := h.pointsService.GetCustomerPoints(c.Request.Context(), session.StoreID, customerID, customerName, customerCode)
	if err != nil {

		fmt.Println("err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *PointsHandler) GetRedeemableProducts(c *gin.Context) {
	token := extractBearerToken(c)
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "session token required"})
		return
	}

	session, err := h.appAuthService.ValidateSession(c.Request.Context(), token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid session"})
		return
	}

	if session.BranchID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "branch not selected"})
		return
	}

	result, err := h.pointsService.GetRedeemableProducts(c.Request.Context(), session.StoreID, *session.BranchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *PointsHandler) RedeemPoints(c *gin.Context) {
	token := extractBearerToken(c)
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "session token required"})
		return
	}

	session, err := h.appAuthService.ValidateSession(c.Request.Context(), token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid session"})
		return
	}

	if session.BranchID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "branch not selected"})
		return
	}

	var req domain.RedeemPointsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.pointsService.RedeemPoints(c.Request.Context(), session.StoreID, *session.BranchID, &req, session.StaffID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *PointsHandler) GetPointHistory(c *gin.Context) {
	token := extractBearerToken(c)
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "session token required"})
		return
	}

	session, err := h.appAuthService.ValidateSession(c.Request.Context(), token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid session"})
		return
	}

	customerID, err := strconv.ParseInt(c.Param("customer_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid customer_id"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	result, err := h.pointsService.GetPointHistory(c.Request.Context(), session.StoreID, customerID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}
