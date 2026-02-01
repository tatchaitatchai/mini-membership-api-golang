package handler

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/katom-membership/api/internal/domain"
	"github.com/katom-membership/api/internal/service"
	"golang.org/x/crypto/bcrypt"
)

type AppAuthHandler struct {
	appAuthService service.AppAuthService
}

func NewAppAuthHandler(appAuthService service.AppAuthService) *AppAuthHandler {
	return &AppAuthHandler{
		appAuthService: appAuthService,
	}
}

// LoginStore handles store login via email
func (h *AppAuthHandler) LoginStore(c *gin.Context) {
	var req domain.AppLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.appAuthService.LoginStore(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ValidateSession checks if the current session is valid
func (h *AppAuthHandler) ValidateSession(c *gin.Context) {
	token := extractToken(c)
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "session token required"})
		return
	}

	info, err := h.appAuthService.ValidateSession(c.Request.Context(), token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, info)
}

// VerifyPin verifies staff PIN and associates staff with session
func (h *AppAuthHandler) VerifyPin(c *gin.Context) {
	token := extractToken(c)
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "session token required"})
		return
	}

	var req domain.AppPinVerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.appAuthService.VerifyPin(c.Request.Context(), token, &req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// RegisterBusiness handles new business registration
func (h *AppAuthHandler) RegisterBusiness(c *gin.Context) {
	var req domain.AppRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.appAuthService.RegisterBusiness(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// Logout revokes the current session
func (h *AppAuthHandler) Logout(c *gin.Context) {
	token := extractToken(c)
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "session token required"})
		return
	}

	if err := h.appAuthService.Logout(c.Request.Context(), token); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "logged out successfully"})
}

// extractToken gets the session token from Authorization header
func extractToken(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return ""
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return ""
	}

	return parts[1]
}

// GenerateHash generates both SHA256 (for PIN) and bcrypt (for password) hashes
func (h *AppAuthHandler) GenerateHash(c *gin.Context) {
	var req struct {
		Value string `json:"value" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// SHA256 hash (for PIN)
	sha256Hash := sha256.Sum256([]byte(req.Value))
	sha256Hex := hex.EncodeToString(sha256Hash[:])

	// bcrypt hash (for password)
	bcryptHash, err := bcrypt.GenerateFromPassword([]byte(req.Value), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"value":       req.Value,
		"sha256_hash": sha256Hex,
		"bcrypt_hash": string(bcryptHash),
	})
}
