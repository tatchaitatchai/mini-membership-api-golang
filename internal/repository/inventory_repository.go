package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/mini-membership/api/internal/domain"
)

type InventoryRepository interface {
	CreateMovement(ctx context.Context, movement *domain.InventoryMovement) (*domain.InventoryMovement, error)
	GetMovementsByBranch(ctx context.Context, storeID, branchID int64, limit, offset int) ([]domain.InventoryMovementResponse, error)
	GetLowStockItems(ctx context.Context, storeID, branchID int64, threshold int) (*domain.LowStockResponse, error)
	GetBranchProductStock(ctx context.Context, branchID, productID int64) (int, error)
	UpdateBranchProductStock(ctx context.Context, storeID, branchID, productID int64, newStock int) error
	AdjustStock(ctx context.Context, storeID, branchID, productID int64, quantityChange int, movementType domain.MovementType, reason, note *string, changedBy int64) error
}

type inventoryRepository struct {
	db *sqlx.DB
}

func NewInventoryRepository(db *sqlx.DB) InventoryRepository {
	return &inventoryRepository{db: db}
}

func (r *inventoryRepository) CreateMovement(ctx context.Context, movement *domain.InventoryMovement) (*domain.InventoryMovement, error) {
	query := `
		INSERT INTO inventory_movements (
			store_id, branch_id, product_id, movement_type, quantity_change,
			from_stock_count, to_stock_count, reason, note, changed_by,
			reference_table, reference_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, created_at
	`
	err := r.db.QueryRowContext(ctx, query,
		movement.StoreID, movement.BranchID, movement.ProductID, movement.MovementType, movement.QuantityChange,
		movement.FromStockCount, movement.ToStockCount, movement.Reason, movement.Note, movement.ChangedBy,
		movement.ReferenceTable, movement.ReferenceID,
	).Scan(&movement.ID, &movement.CreatedAt)
	if err != nil {
		return nil, err
	}
	return movement, nil
}

func (r *inventoryRepository) GetMovementsByBranch(ctx context.Context, storeID, branchID int64, limit, offset int) ([]domain.InventoryMovementResponse, error) {
	query := `
		SELECT 
			im.id, im.product_id, p.product_name, im.movement_type, im.quantity_change,
			im.from_stock_count, im.to_stock_count, im.reason, im.note, s.email as changed_by_name, im.created_at
		FROM inventory_movements im
		JOIN products p ON im.product_id = p.id
		LEFT JOIN staff_accounts s ON im.changed_by = s.id
		WHERE im.store_id = $1 AND im.branch_id = $2
		ORDER BY im.created_at DESC
		LIMIT $3 OFFSET $4
	`
	rows, err := r.db.QueryContext(ctx, query, storeID, branchID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var movements []domain.InventoryMovementResponse
	for rows.Next() {
		var m domain.InventoryMovementResponse
		var changedByName sql.NullString
		err := rows.Scan(
			&m.ID, &m.ProductID, &m.ProductName, &m.MovementType, &m.QuantityChange,
			&m.FromStockCount, &m.ToStockCount, &m.Reason, &m.Note, &changedByName, &m.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		if changedByName.Valid {
			m.ChangedByName = &changedByName.String
		}
		movements = append(movements, m)
	}
	return movements, nil
}

func (r *inventoryRepository) GetLowStockItems(ctx context.Context, storeID, branchID int64, threshold int) (*domain.LowStockResponse, error) {
	query := `
		SELECT 
			bp.product_id, p.product_name, c.category_name, bp.on_stock, bp.reorder_level, p.base_price
		FROM branch_products bp
		JOIN products p ON bp.product_id = p.id
		JOIN categories c ON p.category_id = c.id
		WHERE bp.store_id = $1 AND bp.branch_id = $2 
			AND bp.is_active = true
			AND (bp.on_stock <= bp.reorder_level OR bp.on_stock <= $3)
		ORDER BY bp.on_stock ASC, p.product_name ASC
	`
	rows, err := r.db.QueryContext(ctx, query, storeID, branchID, threshold)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer rows.Close()

	var items []domain.LowStockItem
	for rows.Next() {
		var item domain.LowStockItem
		err := rows.Scan(&item.ProductID, &item.ProductName, &item.CategoryName, &item.OnStock, &item.ReorderLevel, &item.Price)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	fmt.Println(items)

	return &domain.LowStockResponse{
		Items:      items,
		TotalCount: len(items),
	}, nil
}

func (r *inventoryRepository) GetBranchProductStock(ctx context.Context, branchID, productID int64) (int, error) {
	var stock int
	err := r.db.QueryRowContext(ctx, `
		SELECT COALESCE(on_stock, 0) FROM branch_products WHERE branch_id = $1 AND product_id = $2
	`, branchID, productID).Scan(&stock)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return stock, err
}

func (r *inventoryRepository) UpdateBranchProductStock(ctx context.Context, storeID, branchID, productID int64, newStock int) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO branch_products (store_id, branch_id, product_id, on_stock)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (branch_id, product_id) 
		DO UPDATE SET on_stock = $4, updated_at = NOW()
	`, storeID, branchID, productID, newStock)
	return err
}

func (r *inventoryRepository) AdjustStock(ctx context.Context, storeID, branchID, productID int64, quantityChange int, movementType domain.MovementType, reason, note *string, changedBy int64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Get current stock
	var currentStock int
	err = tx.QueryRowContext(ctx, `
		SELECT COALESCE(on_stock, 0) FROM branch_products WHERE branch_id = $1 AND product_id = $2
	`, branchID, productID).Scan(&currentStock)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	newStock := currentStock + quantityChange

	// Update branch_products
	_, err = tx.ExecContext(ctx, `
		INSERT INTO branch_products (store_id, branch_id, product_id, on_stock)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (branch_id, product_id) 
		DO UPDATE SET on_stock = $4, updated_at = NOW()
	`, storeID, branchID, productID, newStock)
	if err != nil {
		return err
	}

	// Create inventory movement record
	_, err = tx.ExecContext(ctx, `
		INSERT INTO inventory_movements (
			store_id, branch_id, product_id, movement_type, quantity_change,
			from_stock_count, to_stock_count, reason, note, changed_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, storeID, branchID, productID, movementType, quantityChange, currentStock, newStock, reason, note, changedBy)
	if err != nil {
		return err
	}

	return tx.Commit()
}
