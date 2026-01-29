package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/katom-membership/api/pkg/models"
	"github.com/shopspring/decimal"
)

type ShiftRepository interface {
	GetBranchesByStoreID(ctx context.Context, storeID int64) ([]models.Branch, error)
	GetBranchByID(ctx context.Context, branchID int64) (*models.Branch, error)
	UpdateSessionBranch(ctx context.Context, sessionToken string, branchID int64) error
	GetActiveShiftByBranch(ctx context.Context, branchID int64) (*models.Shift, error)
	CreateShift(ctx context.Context, shift *models.Shift) error
	UpdateBranchShiftStatus(ctx context.Context, branchID int64, isOpened bool) error
}

type shiftRepository struct {
	db *sqlx.DB
}

func NewShiftRepository(db *sqlx.DB) ShiftRepository {
	return &shiftRepository{db: db}
}

func (r *shiftRepository) GetBranchesByStoreID(ctx context.Context, storeID int64) ([]models.Branch, error) {
	var branches []models.Branch
	query := `
		SELECT id, store_id, branch_name, is_shift_opened, shift_opened_at, shift_closed_at, is_active, created_at, updated_at
		FROM branches
		WHERE store_id = $1 AND is_active = true
		ORDER BY branch_name
	`
	err := r.db.SelectContext(ctx, &branches, query, storeID)
	if err != nil {
		return nil, err
	}
	return branches, nil
}

func (r *shiftRepository) GetBranchByID(ctx context.Context, branchID int64) (*models.Branch, error) {
	var branch models.Branch
	query := `
		SELECT id, store_id, branch_name, is_shift_opened, shift_opened_at, shift_closed_at, is_active, created_at, updated_at
		FROM branches
		WHERE id = $1
	`
	err := r.db.GetContext(ctx, &branch, query, branchID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &branch, nil
}

func (r *shiftRepository) UpdateSessionBranch(ctx context.Context, sessionToken string, branchID int64) error {
	query := `UPDATE app_sessions SET branch_id = $1, last_seen_at = $2 WHERE session_token = $3`
	_, err := r.db.ExecContext(ctx, query, branchID, time.Now(), sessionToken)
	return err
}

func (r *shiftRepository) GetActiveShiftByBranch(ctx context.Context, branchID int64) (*models.Shift, error) {
	var shift models.Shift
	query := `
		SELECT id, store_id, branch_id, start_money_inbox, end_money_inbox, started_at, ended_at, is_active_shift, opened_by, closed_by, created_at, updated_at
		FROM shifts
		WHERE branch_id = $1 AND is_active_shift = true
		ORDER BY started_at DESC
		LIMIT 1
	`
	err := r.db.GetContext(ctx, &shift, query, branchID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &shift, nil
}

func (r *shiftRepository) CreateShift(ctx context.Context, shift *models.Shift) error {
	query := `
		INSERT INTO shifts (store_id, branch_id, start_money_inbox, started_at, is_active_shift, opened_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`
	return r.db.QueryRowContext(
		ctx, query,
		shift.StoreID,
		shift.BranchID,
		shift.StartMoneyInbox,
		shift.StartedAt,
		shift.IsActiveShift,
		shift.OpenedBy,
		shift.CreatedAt,
		shift.UpdatedAt,
	).Scan(&shift.ID)
}

func (r *shiftRepository) UpdateBranchShiftStatus(ctx context.Context, branchID int64, isOpened bool) error {
	var query string
	if isOpened {
		query = `UPDATE branches SET is_shift_opened = true, shift_opened_at = $1 WHERE id = $2`
	} else {
		query = `UPDATE branches SET is_shift_opened = false, shift_closed_at = $1 WHERE id = $2`
	}
	_, err := r.db.ExecContext(ctx, query, time.Now(), branchID)
	return err
}

// OpenShiftTx opens a shift within a transaction
func (r *shiftRepository) OpenShiftTx(ctx context.Context, storeID, branchID int64, startingCash decimal.Decimal, staffID *int64) (*models.Shift, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	now := time.Now()
	shift := &models.Shift{
		StoreID:         storeID,
		BranchID:        branchID,
		StartMoneyInbox: startingCash,
		StartedAt:       now,
		IsActiveShift:   true,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if staffID != nil {
		shift.OpenedBy = sql.NullInt64{Int64: *staffID, Valid: true}
	}

	query := `
		INSERT INTO shifts (store_id, branch_id, start_money_inbox, started_at, is_active_shift, opened_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`
	err = tx.QueryRowContext(
		ctx, query,
		shift.StoreID,
		shift.BranchID,
		shift.StartMoneyInbox,
		shift.StartedAt,
		shift.IsActiveShift,
		shift.OpenedBy,
		shift.CreatedAt,
		shift.UpdatedAt,
	).Scan(&shift.ID)
	if err != nil {
		return nil, err
	}

	// Update branch shift status
	_, err = tx.ExecContext(ctx, `UPDATE branches SET is_shift_opened = true, shift_opened_at = $1 WHERE id = $2`, now, branchID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return shift, nil
}
