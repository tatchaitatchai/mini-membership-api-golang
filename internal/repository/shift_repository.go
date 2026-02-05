package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/mini-membership/api/pkg/models"
	"github.com/shopspring/decimal"
)

type ShiftRepository interface {
	GetBranchesByStoreID(ctx context.Context, storeID int64) ([]models.Branch, error)
	GetBranchByID(ctx context.Context, storeID, branchID int64) (*models.Branch, error)
	UpdateSessionBranch(ctx context.Context, sessionToken string, storeID, branchID int64) error
	GetActiveShiftByBranch(ctx context.Context, storeID, branchID int64) (*models.Shift, error)
	CreateShift(ctx context.Context, shift *models.Shift) error
	UpdateBranchShiftStatus(ctx context.Context, storeID, branchID int64, isOpened bool) error
	CloseShiftTx(ctx context.Context, storeID, branchID, shiftID int64, endCash, expectedCash decimal.Decimal, staffID *int64, note string, stockCounts []StockCountItem) error
	GetShiftSalesSummary(ctx context.Context, storeID, shiftID int64) (totalSales decimal.Decimal, orderCount int, err error)
	GetShiftCashSales(ctx context.Context, storeID, shiftID int64) (decimal.Decimal, error)
	GetShiftCashMovements(ctx context.Context, storeID, shiftID int64) (totalOut decimal.Decimal, totalIn decimal.Decimal, err error)
	GetStaffNameByID(ctx context.Context, storeID, staffID int64) (string, error)
	GetShiftCancelledOrdersSummary(ctx context.Context, storeID, shiftID int64) (cancelledTotal decimal.Decimal, cancelledCount int, err error)
}

type StockCountItem struct {
	ProductID   int64
	ActualStock int
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

func (r *shiftRepository) GetBranchByID(ctx context.Context, storeID, branchID int64) (*models.Branch, error) {
	var branch models.Branch
	query := `
		SELECT id, store_id, branch_name, is_shift_opened, shift_opened_at, shift_closed_at, is_active, created_at, updated_at
		FROM branches
		WHERE id = $1 AND store_id = $2
	`
	err := r.db.GetContext(ctx, &branch, query, branchID, storeID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &branch, nil
}

func (r *shiftRepository) UpdateSessionBranch(ctx context.Context, sessionToken string, storeID, branchID int64) error {
	query := `UPDATE app_sessions SET branch_id = $1, last_seen_at = $2 WHERE session_token = $3 AND store_id = $4`
	_, err := r.db.ExecContext(ctx, query, branchID, time.Now(), sessionToken, storeID)
	return err
}

func (r *shiftRepository) GetActiveShiftByBranch(ctx context.Context, storeID, branchID int64) (*models.Shift, error) {
	var shift models.Shift
	query := `
		SELECT id, store_id, branch_id, start_money_inbox, end_money_inbox, started_at, ended_at, is_active_shift, opened_by, closed_by, created_at, updated_at
		FROM shifts
		WHERE branch_id = $1 AND store_id = $2 AND is_active_shift = true
		ORDER BY started_at DESC
		LIMIT 1
	`
	err := r.db.GetContext(ctx, &shift, query, branchID, storeID)
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

func (r *shiftRepository) UpdateBranchShiftStatus(ctx context.Context, storeID, branchID int64, isOpened bool) error {
	var query string
	if isOpened {
		query = `UPDATE branches SET is_shift_opened = true, shift_opened_at = $1 WHERE id = $2 AND store_id = $3`
	} else {
		query = `UPDATE branches SET is_shift_opened = false, shift_closed_at = $1 WHERE id = $2 AND store_id = $3`
	}
	_, err := r.db.ExecContext(ctx, query, time.Now(), branchID, storeID)
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
	_, err = tx.ExecContext(ctx, `UPDATE branches SET is_shift_opened = true, shift_opened_at = $1 WHERE id = $2 AND store_id = $3`, now, branchID, storeID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return shift, nil
}

// CloseShiftTx closes a shift within a transaction
func (r *shiftRepository) CloseShiftTx(ctx context.Context, storeID, branchID, shiftID int64, endCash, expectedCash decimal.Decimal, staffID *int64, note string, stockCounts []StockCountItem) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	now := time.Now()

	// Update shift record
	updateQuery := `
		UPDATE shifts 
		SET end_money_inbox = $1, ended_at = $2, is_active_shift = false, closed_by = $3, updated_at = $4, closing_cash_expected = $8
		WHERE id = $5 AND store_id = $6 AND branch_id = $7
	`
	var closedBy sql.NullInt64
	if staffID != nil {
		closedBy = sql.NullInt64{Int64: *staffID, Valid: true}
	}
	_, err = tx.ExecContext(ctx, updateQuery, endCash, now, closedBy, now, shiftID, storeID, branchID, expectedCash)
	if err != nil {
		return err
	}

	// Update branch shift status
	_, err = tx.ExecContext(ctx, `UPDATE branches SET is_shift_opened = false, shift_closed_at = $1 WHERE id = $2 AND store_id = $3`, now, branchID, storeID)
	if err != nil {
		return err
	}

	// Save stock counts if provided
	if len(stockCounts) > 0 {
		// Create shift_stock_counts record
		var stockCountID int64
		stockCountQuery := `
			INSERT INTO shift_stock_counts (store_id, branch_id, shift_id, counted_by, counted_at, note, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $5, $5)
			RETURNING id
		`
		err = tx.QueryRowContext(ctx, stockCountQuery, storeID, branchID, shiftID, closedBy, now, note).Scan(&stockCountID)
		if err != nil {
			return err
		}

		// Insert individual stock count items
		for _, item := range stockCounts {
			// Get expected stock from branch_products
			var expectedStock int
			expectedQuery := `SELECT COALESCE(on_stock, 0) FROM branch_products WHERE store_id = $1 AND branch_id = $2 AND product_id = $3`
			err = tx.QueryRowContext(ctx, expectedQuery, storeID, branchID, item.ProductID).Scan(&expectedStock)
			if err == sql.ErrNoRows {
				expectedStock = 0
			} else if err != nil {
				return err
			}

			difference := item.ActualStock - expectedStock

			itemQuery := `
				INSERT INTO shift_stock_count_items (shift_stock_count_id, product_id, expected_stock, actual_stock, difference, created_at)
				VALUES ($1, $2, $3, $4, $5, $6)
			`
			_, err = tx.ExecContext(ctx, itemQuery, stockCountID, item.ProductID, expectedStock, item.ActualStock, difference, now)
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

// GetShiftSalesSummary returns total sales (all methods) and order count for a shift
func (r *shiftRepository) GetShiftSalesSummary(ctx context.Context, storeID, shiftID int64) (totalSales decimal.Decimal, orderCount int, err error) {
	query := `
		SELECT COALESCE(SUM(total_price), 0) as total_sales, COUNT(*) as order_count
		FROM orders
		WHERE store_id = $1 AND shift_id = $2 AND status = 'PAID'
	`
	var result struct {
		TotalSales decimal.Decimal `db:"total_sales"`
		OrderCount int             `db:"order_count"`
	}
	err = r.db.GetContext(ctx, &result, query, storeID, shiftID)
	if err != nil {
		return decimal.Zero, 0, err
	}
	return result.TotalSales, result.OrderCount, nil
}

// GetShiftCashSales returns net CASH in drawer (cash received - change given)
func (r *shiftRepository) GetShiftCashSales(ctx context.Context, storeID, shiftID int64) (decimal.Decimal, error) {
	query := `
		SELECT COALESCE(SUM(p.amount - o.change_amount), 0) as net_cash
		FROM payments p
		JOIN orders o ON o.id = p.order_id
		WHERE o.store_id = $1 AND o.shift_id = $2 AND o.status = 'PAID' AND p.method = 'CASH'
	`
	var netCash decimal.Decimal
	err := r.db.GetContext(ctx, &netCash, query, storeID, shiftID)
	if err != nil {
		return decimal.Zero, err
	}
	return netCash, nil
}

// GetShiftCashMovements returns total cash out and cash in for a shift
func (r *shiftRepository) GetShiftCashMovements(ctx context.Context, storeID, shiftID int64) (totalOut decimal.Decimal, totalIn decimal.Decimal, err error) {
	query := `
		SELECT 
			COALESCE(SUM(CASE WHEN direction = 'OUT' THEN amount ELSE 0 END), 0) as total_out,
			COALESCE(SUM(CASE WHEN direction = 'IN' THEN amount ELSE 0 END), 0) as total_in
		FROM shift_cash_movements
		WHERE store_id = $1 AND shift_id = $2
	`
	var result struct {
		TotalOut decimal.Decimal `db:"total_out"`
		TotalIn  decimal.Decimal `db:"total_in"`
	}
	err = r.db.GetContext(ctx, &result, query, storeID, shiftID)
	if err != nil {
		return decimal.Zero, decimal.Zero, err
	}
	return result.TotalOut, result.TotalIn, nil
}

// GetStaffNameByID returns staff display name by ID
func (r *shiftRepository) GetStaffNameByID(ctx context.Context, storeID, staffID int64) (string, error) {
	var name string
	query := `SELECT COALESCE(display_name, email) FROM staff_accounts WHERE id = $1 AND store_id = $2`
	err := r.db.GetContext(ctx, &name, query, staffID, storeID)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return name, err
}

// GetShiftCancelledOrdersSummary returns total and count of cancelled orders for a shift
func (r *shiftRepository) GetShiftCancelledOrdersSummary(ctx context.Context, storeID, shiftID int64) (cancelledTotal decimal.Decimal, cancelledCount int, err error) {
	query := `
		SELECT COALESCE(SUM(total_price), 0) as cancelled_total, COUNT(*) as cancelled_count
		FROM orders
		WHERE store_id = $1 AND shift_id = $2 AND status = 'CANCELLED'
	`
	var result struct {
		CancelledTotal decimal.Decimal `db:"cancelled_total"`
		CancelledCount int             `db:"cancelled_count"`
	}
	err = r.db.GetContext(ctx, &result, query, storeID, shiftID)
	if err != nil {
		return decimal.Zero, 0, err
	}
	return result.CancelledTotal, result.CancelledCount, nil
}
