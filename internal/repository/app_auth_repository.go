package repository

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/katom-membership/api/pkg/models"
)

type AppAuthRepository interface {
	GetStaffByEmail(ctx context.Context, email string) (*models.StaffAccount, error)
	GetStoreByEmail(ctx context.Context, email string) (*models.Store, error)
	CreateSession(ctx context.Context, session *models.AppSession) error
	GetSessionByToken(ctx context.Context, token string) (*models.AppSession, error)
	UpdateSessionLastSeen(ctx context.Context, token string) error
	RevokeSession(ctx context.Context, token string) error
	GetStoreByID(ctx context.Context, storeID int64) (*models.Store, error)
	GetBranchByID(ctx context.Context, branchID int64) (*models.Branch, error)
	GetStaffByID(ctx context.Context, staffID int64) (*models.StaffAccount, error)
	GetStaffByPinAndStore(ctx context.Context, pinHash string, storeID int64) (*models.StaffAccount, error)
	CreateStore(ctx context.Context, store *models.Store) (int64, error)
	CreateStaffAccount(ctx context.Context, staff *models.StaffAccount) (int64, error)
	UpdateSessionStaff(ctx context.Context, token string, staffID int64) error
}

type appAuthRepository struct {
	db *sqlx.DB
}

func NewAppAuthRepository(db *sqlx.DB) AppAuthRepository {
	return &appAuthRepository{db: db}
}

func (r *appAuthRepository) GetStaffByEmail(ctx context.Context, email string) (*models.StaffAccount, error) {
	var staff models.StaffAccount
	query := `
		SELECT id, store_id, branch_id, email, password_hash, pin_hash, is_active, is_store_master, is_working, created_at, updated_at
		FROM staff_accounts
		WHERE email = $1 AND is_store_master = true AND is_active = true
		LIMIT 1
	`
	err := r.db.GetContext(ctx, &staff, query, email)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &staff, nil
}

func (r *appAuthRepository) GetStoreByEmail(ctx context.Context, email string) (*models.Store, error) {
	var store models.Store
	query := `
		SELECT s.id, s.store_name, s.is_active, s.created_at, s.updated_at
		FROM stores s
		JOIN staff_accounts sa ON sa.store_id = s.id
		WHERE sa.email = $1 AND sa.is_store_master = true AND s.is_active = true
		LIMIT 1
	`
	err := r.db.GetContext(ctx, &store, query, email)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &store, nil
}

func (r *appAuthRepository) CreateSession(ctx context.Context, session *models.AppSession) error {
	query := `
		INSERT INTO app_sessions (store_id, branch_id, staff_id, session_token, created_at, last_seen_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`
	return r.db.QueryRowContext(
		ctx, query,
		session.StoreID,
		session.BranchID,
		session.StaffID,
		session.SessionToken,
		session.CreatedAt,
		session.LastSeenAt,
	).Scan(&session.ID)
}

func (r *appAuthRepository) GetSessionByToken(ctx context.Context, token string) (*models.AppSession, error) {
	var session models.AppSession
	query := `
		SELECT id, store_id, branch_id, staff_id, session_token, created_at, last_seen_at, revoked_at
		FROM app_sessions
		WHERE session_token = $1 AND revoked_at IS NULL
	`
	err := r.db.GetContext(ctx, &session, query, token)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *appAuthRepository) UpdateSessionLastSeen(ctx context.Context, token string) error {
	query := `UPDATE app_sessions SET last_seen_at = $1 WHERE session_token = $2`
	_, err := r.db.ExecContext(ctx, query, time.Now(), token)
	return err
}

func (r *appAuthRepository) RevokeSession(ctx context.Context, token string) error {
	query := `UPDATE app_sessions SET revoked_at = $1 WHERE session_token = $2`
	_, err := r.db.ExecContext(ctx, query, time.Now(), token)
	return err
}

func (r *appAuthRepository) GetStoreByID(ctx context.Context, storeID int64) (*models.Store, error) {
	var store models.Store
	query := `SELECT id, store_name, is_active, created_at, updated_at FROM stores WHERE id = $1`
	err := r.db.GetContext(ctx, &store, query, storeID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &store, nil
}

func (r *appAuthRepository) GetBranchByID(ctx context.Context, branchID int64) (*models.Branch, error) {
	var branch models.Branch
	query := `SELECT id, store_id, branch_name, is_shift_opened, shift_opened_at, shift_closed_at, is_active, created_at, updated_at FROM branches WHERE id = $1`
	err := r.db.GetContext(ctx, &branch, query, branchID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &branch, nil
}

func (r *appAuthRepository) GetStaffByID(ctx context.Context, staffID int64) (*models.StaffAccount, error) {
	var staff models.StaffAccount
	query := `SELECT id, store_id, branch_id, email, password_hash, pin_hash, is_active, is_store_master, is_working, created_at, updated_at FROM staff_accounts WHERE id = $1`
	err := r.db.GetContext(ctx, &staff, query, staffID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &staff, nil
}

func (r *appAuthRepository) GetStaffByPinAndStore(ctx context.Context, pinHash string, storeID int64) (*models.StaffAccount, error) {
	var staff models.StaffAccount
	query := `
		SELECT id, store_id, branch_id, email, password_hash, pin_hash, is_active, is_store_master, is_working, created_at, updated_at 
		FROM staff_accounts 
		WHERE pin_hash = $1 AND store_id = $2 AND is_active = true
	`
	err := r.db.GetContext(ctx, &staff, query, pinHash, storeID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &staff, nil
}

func (r *appAuthRepository) CreateStore(ctx context.Context, store *models.Store) (int64, error) {
	query := `
		INSERT INTO stores (store_name, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`
	var id int64
	err := r.db.QueryRowContext(ctx, query, store.StoreName, store.IsActive, store.CreatedAt, store.UpdatedAt).Scan(&id)
	return id, err
}

func (r *appAuthRepository) CreateStaffAccount(ctx context.Context, staff *models.StaffAccount) (int64, error) {
	query := `
		INSERT INTO staff_accounts (store_id, branch_id, email, password_hash, pin_hash, is_active, is_store_master, is_working, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id
	`
	var id int64
	err := r.db.QueryRowContext(
		ctx, query,
		staff.StoreID,
		staff.BranchID,
		staff.Email,
		staff.PasswordHash,
		staff.PinHash,
		staff.IsActive,
		staff.IsStoreMaster,
		staff.IsWorking,
		staff.CreatedAt,
		staff.UpdatedAt,
	).Scan(&id)
	return id, err
}

func (r *appAuthRepository) UpdateSessionStaff(ctx context.Context, token string, staffID int64) error {
	query := `UPDATE app_sessions SET staff_id = $1, last_seen_at = $2 WHERE session_token = $3`
	_, err := r.db.ExecContext(ctx, query, staffID, time.Now(), token)
	return err
}

// GenerateSessionToken creates a secure random token
func GenerateSessionToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
