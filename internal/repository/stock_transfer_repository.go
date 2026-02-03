package repository

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/katom-membership/api/internal/domain"
)

type StockTransferRepository interface {
	Create(ctx context.Context, storeID int64, req *domain.CreateStockTransferRequest, sentBy int64) (*domain.StockTransfer, error)
	GetByID(ctx context.Context, storeID, transferID int64) (*domain.StockTransferResponse, error)
	GetByBranch(ctx context.Context, storeID, branchID int64, limit, offset int) (*domain.StockTransferListResponse, error)
	GetPendingTransfers(ctx context.Context, storeID, branchID int64) ([]domain.StockTransferResponse, error)
	UpdateStatus(ctx context.Context, storeID, transferID int64, status domain.StockTransferStatus, receivedBy *int64) error
	UpdateReceiveCounts(ctx context.Context, transferID int64, items []domain.UpdateStockTransferItemInput) error
	ReceiveAndAddStock(ctx context.Context, storeID, branchID, transferID int64, items []domain.UpdateStockTransferItemInput, receivedBy int64) error
	GetTransferItems(ctx context.Context, transferID int64) ([]domain.StockTransferItemResponse, error)
}

type stockTransferRepository struct {
	db *sqlx.DB
}

func NewStockTransferRepository(db *sqlx.DB) StockTransferRepository {
	return &stockTransferRepository{db: db}
}

func (r *stockTransferRepository) Create(ctx context.Context, storeID int64, req *domain.CreateStockTransferRequest, requestedBy int64) (*domain.StockTransfer, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Create stock transfer request (status: CREATED - waiting for central to process)
	var transfer domain.StockTransfer
	err = tx.QueryRowContext(ctx, `
		INSERT INTO stock_transfers (store_id, from_branch_id, to_branch_id, status, note)
		VALUES ($1, $2, $3, 'CREATED', $4)
		RETURNING id, store_id, from_branch_id, to_branch_id, status, sent_by, sent_at, note, created_at, updated_at
	`, storeID, req.FromBranchID, req.ToBranchID, req.Note).Scan(
		&transfer.ID, &transfer.StoreID, &transfer.FromBranchID, &transfer.ToBranchID,
		&transfer.Status, &transfer.SentBy, &transfer.SentAt, &transfer.Note,
		&transfer.CreatedAt, &transfer.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	// Create transfer items (requested quantities)
	for _, item := range req.Items {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO stock_transfer_items (stock_transfer_id, product_id, send_count)
			VALUES ($1, $2, $3)
		`, transfer.ID, item.ProductID, item.SendCount)
		if err != nil {
			return nil, err
		}
		// Note: Stock is NOT reduced here - it will be reduced when central sends the goods
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return &transfer, nil
}

func (r *stockTransferRepository) GetByID(ctx context.Context, storeID, transferID int64) (*domain.StockTransferResponse, error) {
	var resp domain.StockTransferResponse
	var sentByName, receivedByName, fromBranchName sql.NullString

	err := r.db.QueryRowContext(ctx, `
		SELECT 
			st.id, st.from_branch_id, fb.branch_name, st.to_branch_id, tb.branch_name,
			st.status, ss.email, rs.email, st.sent_at, st.received_at, st.note, st.created_at
		FROM stock_transfers st
		LEFT JOIN branches fb ON st.from_branch_id = fb.id
		JOIN branches tb ON st.to_branch_id = tb.id
		LEFT JOIN staff_accounts ss ON st.sent_by = ss.id
		LEFT JOIN staff_accounts rs ON st.received_by = rs.id
		WHERE st.id = $1 AND st.store_id = $2
	`, transferID, storeID).Scan(
		&resp.ID, &resp.FromBranchID, &fromBranchName, &resp.ToBranchID, &resp.ToBranchName,
		&resp.Status, &sentByName, &receivedByName, &resp.SentAt, &resp.ReceivedAt, &resp.Note, &resp.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if fromBranchName.Valid {
		resp.FromBranchName = &fromBranchName.String
	}
	if sentByName.Valid {
		resp.SentByName = &sentByName.String
	}
	if receivedByName.Valid {
		resp.ReceivedByName = &receivedByName.String
	}

	// Get items
	items, err := r.GetTransferItems(ctx, transferID)
	if err != nil {
		return nil, err
	}
	resp.Items = items

	return &resp, nil
}

func (r *stockTransferRepository) GetByBranch(ctx context.Context, storeID, branchID int64, limit, offset int) (*domain.StockTransferListResponse, error) {
	// Get total count
	var total int
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM stock_transfers
		WHERE store_id = $1 AND (from_branch_id = $2 OR to_branch_id = $2)
	`, storeID, branchID).Scan(&total)
	if err != nil {
		return nil, err
	}

	// Get transfers
	rows, err := r.db.QueryContext(ctx, `
		SELECT 
			st.id, st.from_branch_id, fb.branch_name, st.to_branch_id, tb.branch_name,
			st.status, ss.email, rs.email, st.sent_at, st.received_at, st.note, st.created_at
		FROM stock_transfers st
		LEFT JOIN branches fb ON st.from_branch_id = fb.id
		JOIN branches tb ON st.to_branch_id = tb.id
		LEFT JOIN staff_accounts ss ON st.sent_by = ss.id
		LEFT JOIN staff_accounts rs ON st.received_by = rs.id
		WHERE st.store_id = $1 AND (st.from_branch_id = $2 OR st.to_branch_id = $2)
		ORDER BY st.created_at DESC
		LIMIT $3 OFFSET $4
	`, storeID, branchID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transfers []domain.StockTransferResponse
	for rows.Next() {
		var resp domain.StockTransferResponse
		var sentByName, receivedByName, fromBranchName sql.NullString

		err := rows.Scan(
			&resp.ID, &resp.FromBranchID, &fromBranchName, &resp.ToBranchID, &resp.ToBranchName,
			&resp.Status, &sentByName, &receivedByName, &resp.SentAt, &resp.ReceivedAt, &resp.Note, &resp.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		if fromBranchName.Valid {
			resp.FromBranchName = &fromBranchName.String
		}
		if sentByName.Valid {
			resp.SentByName = &sentByName.String
		}
		if receivedByName.Valid {
			resp.ReceivedByName = &receivedByName.String
		}

		// Get items for each transfer
		items, err := r.GetTransferItems(ctx, resp.ID)
		if err != nil {
			return nil, err
		}
		resp.Items = items

		transfers = append(transfers, resp)
	}

	return &domain.StockTransferListResponse{
		Transfers: transfers,
		Total:     total,
	}, nil
}

func (r *stockTransferRepository) UpdateStatus(ctx context.Context, storeID, transferID int64, status domain.StockTransferStatus, receivedBy *int64) error {
	var query string
	var args []interface{}

	if status == domain.StockTransferStatusReceived && receivedBy != nil {
		query = `
			UPDATE stock_transfers 
			SET status = $1, received_by = $2, received_at = NOW(), updated_at = NOW()
			WHERE id = $3 AND store_id = $4
		`
		args = []interface{}{status, *receivedBy, transferID, storeID}
	} else {
		query = `
			UPDATE stock_transfers 
			SET status = $1, updated_at = NOW()
			WHERE id = $2 AND store_id = $3
		`
		args = []interface{}{status, transferID, storeID}
	}

	_, err := r.db.ExecContext(ctx, query, args...)
	return err
}

func (r *stockTransferRepository) UpdateReceiveCounts(ctx context.Context, transferID int64, items []domain.UpdateStockTransferItemInput) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, item := range items {
		_, err = tx.ExecContext(ctx, `
			UPDATE stock_transfer_items 
			SET receive_count = $1, updated_at = NOW()
			WHERE stock_transfer_id = $2 AND product_id = $3
		`, item.ReceiveCount, transferID, item.ProductID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *stockTransferRepository) GetTransferItems(ctx context.Context, transferID int64) ([]domain.StockTransferItemResponse, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT sti.id, sti.product_id, p.product_name, sti.send_count, sti.receive_count
		FROM stock_transfer_items sti
		JOIN products p ON sti.product_id = p.id
		WHERE sti.stock_transfer_id = $1
	`, transferID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.StockTransferItemResponse
	for rows.Next() {
		var item domain.StockTransferItemResponse
		err := rows.Scan(&item.ID, &item.ProductID, &item.ProductName, &item.SendCount, &item.ReceiveCount)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, nil
}

func (r *stockTransferRepository) GetPendingTransfers(ctx context.Context, storeID, branchID int64) ([]domain.StockTransferResponse, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT 
			st.id, st.from_branch_id, fb.branch_name, st.to_branch_id, tb.branch_name,
			st.status, ss.email, rs.email, st.sent_at, st.received_at, st.note, st.created_at
		FROM stock_transfers st
		LEFT JOIN branches fb ON st.from_branch_id = fb.id
		JOIN branches tb ON st.to_branch_id = tb.id
		LEFT JOIN staff_accounts ss ON st.sent_by = ss.id
		LEFT JOIN staff_accounts rs ON st.received_by = rs.id
		WHERE st.store_id = $1 AND st.to_branch_id = $2 AND st.status = 'SENT'
		ORDER BY st.created_at DESC
	`, storeID, branchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transfers []domain.StockTransferResponse
	for rows.Next() {
		var resp domain.StockTransferResponse
		var sentByName, receivedByName, fromBranchName sql.NullString

		err := rows.Scan(
			&resp.ID, &resp.FromBranchID, &fromBranchName, &resp.ToBranchID, &resp.ToBranchName,
			&resp.Status, &sentByName, &receivedByName, &resp.SentAt, &resp.ReceivedAt, &resp.Note, &resp.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		if fromBranchName.Valid {
			resp.FromBranchName = &fromBranchName.String
		}
		if sentByName.Valid {
			resp.SentByName = &sentByName.String
		}
		if receivedByName.Valid {
			resp.ReceivedByName = &receivedByName.String
		}

		items, err := r.GetTransferItems(ctx, resp.ID)
		if err != nil {
			return nil, err
		}
		resp.Items = items

		transfers = append(transfers, resp)
	}

	return transfers, nil
}

func (r *stockTransferRepository) ReceiveAndAddStock(ctx context.Context, storeID, branchID, transferID int64, items []domain.UpdateStockTransferItemInput, receivedBy int64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Update receive counts and add stock for each item
	for _, item := range items {
		// Update receive count
		_, err = tx.ExecContext(ctx, `
			UPDATE stock_transfer_items 
			SET receive_count = $1, updated_at = NOW()
			WHERE stock_transfer_id = $2 AND product_id = $3
		`, item.ReceiveCount, transferID, item.ProductID)
		if err != nil {
			return err
		}

		// Add stock to branch_products (upsert)
		_, err = tx.ExecContext(ctx, `
			INSERT INTO branch_products (store_id, branch_id, product_id, on_stock)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (branch_id, product_id) 
			DO UPDATE SET on_stock = branch_products.on_stock + $4, updated_at = NOW()
		`, storeID, branchID, item.ProductID, item.ReceiveCount)
		if err != nil {
			return err
		}
	}

	// Update transfer status to RECEIVED
	_, err = tx.ExecContext(ctx, `
		UPDATE stock_transfers 
		SET status = 'RECEIVED', received_by = $1, received_at = NOW(), updated_at = NOW()
		WHERE id = $2 AND store_id = $3
	`, receivedBy, transferID, storeID)
	if err != nil {
		return err
	}

	return tx.Commit()
}
