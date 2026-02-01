package service

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/katom-membership/api/internal/domain"
	"github.com/katom-membership/api/internal/repository"
	"github.com/katom-membership/api/pkg/models"
	"golang.org/x/crypto/bcrypt"
)

type AppAuthService interface {
	LoginStore(ctx context.Context, req *domain.AppLoginRequest) (*domain.AppLoginResponse, error)
	ValidateSession(ctx context.Context, token string) (*domain.AppSessionInfo, error)
	VerifyPin(ctx context.Context, token string, req *domain.AppPinVerifyRequest) (*domain.AppPinVerifyResponse, error)
	RegisterBusiness(ctx context.Context, req *domain.AppRegisterRequest) (*domain.AppRegisterResponse, error)
	Logout(ctx context.Context, token string) error
}

type appAuthService struct {
	repo              repository.AppAuthRepository
	sessionExpiration time.Duration
}

func NewAppAuthService(repo repository.AppAuthRepository, sessionExpiration time.Duration) AppAuthService {
	return &appAuthService{
		repo:              repo,
		sessionExpiration: sessionExpiration,
	}
}

func (s *appAuthService) LoginStore(ctx context.Context, req *domain.AppLoginRequest) (*domain.AppLoginResponse, error) {
	staff, err := s.repo.GetStaffByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if staff == nil {
		return nil, errors.New("invalid email or password")
	}

	// Verify password
	if !staff.PasswordHash.Valid || staff.PasswordHash.String == "" {
		return nil, errors.New("invalid email or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(staff.PasswordHash.String), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid email or password")
	}

	store, err := s.repo.GetStoreByID(ctx, staff.StoreID)
	if err != nil {
		return nil, err
	}
	if store == nil {
		return nil, errors.New("store not found")
	}

	token, err := repository.GenerateSessionToken()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	session := &models.AppSession{
		StoreID:      store.ID,
		SessionToken: token,
		CreatedAt:    now,
		LastSeenAt:   now,
	}

	if staff.BranchID.Valid {
		session.BranchID = staff.BranchID
	}

	if err := s.repo.CreateSession(ctx, session); err != nil {
		return nil, err
	}

	resp := &domain.AppLoginResponse{
		SessionToken: token,
		StoreID:      store.ID,
		StoreName:    store.StoreName,
		ExpiresAt:    now.Add(s.sessionExpiration),
	}

	if staff.BranchID.Valid {
		resp.BranchID = &staff.BranchID.Int64
	}

	return resp, nil
}

func (s *appAuthService) ValidateSession(ctx context.Context, token string) (*domain.AppSessionInfo, error) {
	session, err := s.repo.GetSessionByToken(ctx, token)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, errors.New("invalid session")
	}

	// Check if session is expired
	if time.Since(session.LastSeenAt) > s.sessionExpiration {
		_ = s.repo.RevokeSession(ctx, token)
		return nil, errors.New("session expired")
	}

	// Update last seen
	_ = s.repo.UpdateSessionLastSeen(ctx, token)

	store, err := s.repo.GetStoreByID(ctx, session.StoreID)
	if err != nil {
		return nil, err
	}
	if store == nil {
		return nil, errors.New("store not found")
	}

	info := &domain.AppSessionInfo{
		StoreID:   store.ID,
		StoreName: store.StoreName,
		ExpiresAt: session.LastSeenAt.Add(s.sessionExpiration),
	}

	// Add branch info if available
	if session.BranchID.Valid {
		branch, err := s.repo.GetBranchByID(ctx, session.StoreID, session.BranchID.Int64)
		if err == nil && branch != nil {
			info.BranchID = &branch.ID
			info.BranchName = &branch.BranchName
		}
	}

	// Add staff info if available
	if session.StaffID.Valid {
		staff, err := s.repo.GetStaffByID(ctx, session.StoreID, session.StaffID.Int64)
		if err == nil && staff != nil {
			info.StaffID = &staff.ID
			if staff.Email.Valid {
				info.StaffName = &staff.Email.String
			}
		}
	}

	return info, nil
}

func (s *appAuthService) VerifyPin(ctx context.Context, token string, req *domain.AppPinVerifyRequest) (*domain.AppPinVerifyResponse, error) {
	session, err := s.repo.GetSessionByToken(ctx, token)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, errors.New("invalid session")
	}

	fmt.Println("Pin :: ", req.Pin)

	// Hash the PIN for comparison
	pinHash := hashPin(req.Pin)

	fmt.Println("Pin Hash :: ", pinHash)

	staff, err := s.repo.GetStaffByPinAndStore(ctx, pinHash, session.StoreID)
	if err != nil {
		return nil, err
	}
	if staff == nil {
		return nil, errors.New("invalid PIN")
	}

	// Update session with staff ID
	if err := s.repo.UpdateSessionStaff(ctx, token, staff.ID); err != nil {
		return nil, err
	}

	staffName := "Staff"
	if staff.Email.Valid {
		staffName = staff.Email.String
	}

	return &domain.AppPinVerifyResponse{
		StaffID:   staff.ID,
		StaffName: staffName,
		IsManager: staff.IsStoreMaster,
	}, nil
}

func (s *appAuthService) RegisterBusiness(ctx context.Context, req *domain.AppRegisterRequest) (*domain.AppRegisterResponse, error) {
	now := time.Now()

	store := &models.Store{
		StoreName: req.BusinessName,
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	storeID, err := s.repo.CreateStore(ctx, store)
	if err != nil {
		return nil, err
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Create master staff account
	staff := &models.StaffAccount{
		StoreID:       storeID,
		Email:         sql.NullString{String: req.Email, Valid: true},
		PasswordHash:  sql.NullString{String: string(hashedPassword), Valid: true},
		PinHash:       sql.NullString{String: hashPin("1234"), Valid: true}, // Default PIN
		IsActive:      true,
		IsStoreMaster: true,
		IsWorking:     false,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	_, err = s.repo.CreateStaffAccount(ctx, staff)
	if err != nil {
		return nil, err
	}

	return &domain.AppRegisterResponse{
		StoreID:   storeID,
		StoreName: req.BusinessName,
		Message:   "Business registered successfully. Default PIN is 1234.",
	}, nil
}

func (s *appAuthService) Logout(ctx context.Context, token string) error {
	return s.repo.RevokeSession(ctx, token)
}

// hashPin creates a SHA256 hash of the PIN
func hashPin(pin string) string {
	hash := sha256.Sum256([]byte(pin))
	return hex.EncodeToString(hash[:])
}
