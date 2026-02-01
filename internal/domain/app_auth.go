package domain

import "time"

// AppLoginRequest for store login via email and password
type AppLoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// AppLoginResponse returns session token and store info
type AppLoginResponse struct {
	SessionToken string    `json:"session_token"`
	StoreID      int64     `json:"store_id"`
	BranchID     *int64    `json:"branch_id,omitempty"`
	StoreName    string    `json:"store_name"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// AppPinVerifyRequest for staff PIN verification
type AppPinVerifyRequest struct {
	Pin string `json:"pin" binding:"required,min=4,max=6"`
}

// AppPinVerifyResponse returns staff info after PIN verification
type AppPinVerifyResponse struct {
	StaffID   int64  `json:"staff_id"`
	StaffName string `json:"staff_name"`
	IsManager bool   `json:"is_manager"`
}

// AppRegisterRequest for new business registration
type AppRegisterRequest struct {
	Email        string `json:"email" binding:"required,email"`
	Password     string `json:"password" binding:"required,min=6"`
	BusinessName string `json:"business_name" binding:"required"`
}

// AppRegisterResponse returns new store info
type AppRegisterResponse struct {
	StoreID   int64  `json:"store_id"`
	StoreName string `json:"store_name"`
	Message   string `json:"message"`
}

// AppSessionInfo for session validation response
type AppSessionInfo struct {
	StoreID    int64     `json:"store_id"`
	StoreName  string    `json:"store_name"`
	BranchID   *int64    `json:"branch_id,omitempty"`
	BranchName *string   `json:"branch_name,omitempty"`
	StaffID    *int64    `json:"staff_id,omitempty"`
	StaffName  *string   `json:"staff_name,omitempty"`
	ExpiresAt  time.Time `json:"expires_at"`
}
