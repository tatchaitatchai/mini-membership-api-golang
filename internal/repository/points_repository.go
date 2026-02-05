package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/mini-membership/api/internal/domain"
)

type PointsRepository interface {
	GetCustomerProductPoints(ctx context.Context, storeID, customerID int64) ([]domain.CustomerProductPointsInfo, error)
	GetProductPoints(ctx context.Context, storeID, customerID, productID int64) (*domain.CustomerProductPoints, error)
	CreateOrUpdateProductPoints(ctx context.Context, storeID, customerID, productID int64, pointsToAdd int) error
	DeductProductPoints(ctx context.Context, storeID, customerID, productID int64, pointsToDeduct int) error
	CreatePointTransaction(ctx context.Context, tx *domain.PointTransaction) error
	CreateRedemption(ctx context.Context, redemption *domain.PointRedemption) (int64, error)
	GetRedeemableProducts(ctx context.Context, storeID, branchID int64) ([]domain.RedeemableProduct, error)
	GetProductPointsToRedeem(ctx context.Context, productID int64) (*int, error)
	GetPointHistory(ctx context.Context, storeID, customerID int64, limit, offset int) ([]domain.PointHistoryItem, int, error)
}

type pointsRepository struct {
	db *sqlx.DB
}

func NewPointsRepository(db *sqlx.DB) PointsRepository {
	return &pointsRepository{db: db}
}

func (r *pointsRepository) GetCustomerProductPoints(ctx context.Context, storeID, customerID int64) ([]domain.CustomerProductPointsInfo, error) {
	query := `
		SELECT 
			cpp.product_id,
			p.product_name,
			c.category_name,
			p.image_path,
			cpp.points,
			cpp.total_points,
			COALESCE(p.points_to_redeem, 0) as points_to_redeem,
			CASE WHEN cpp.points >= COALESCE(p.points_to_redeem, 0) AND p.points_to_redeem > 0 THEN true ELSE false END as can_redeem
		FROM customer_product_points cpp
		JOIN products p ON cpp.product_id = p.id
		LEFT JOIN categories c ON p.category_id = c.id
		WHERE cpp.store_id = $1 AND cpp.customer_id = $2 AND cpp.points > 0
		ORDER BY cpp.points DESC, p.product_name ASC
	`
	var results []domain.CustomerProductPointsInfo
	err := r.db.SelectContext(ctx, &results, query, storeID, customerID)
	if err != nil {
		return nil, err
	}
	return results, nil
}

func (r *pointsRepository) GetProductPoints(ctx context.Context, storeID, customerID, productID int64) (*domain.CustomerProductPoints, error) {
	var points domain.CustomerProductPoints
	query := `
		SELECT id, store_id, customer_id, product_id, points, total_points, created_at, updated_at
		FROM customer_product_points
		WHERE store_id = $1 AND customer_id = $2 AND product_id = $3
	`
	err := r.db.GetContext(ctx, &points, query, storeID, customerID, productID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &points, nil
}

func (r *pointsRepository) CreateOrUpdateProductPoints(ctx context.Context, storeID, customerID, productID int64, pointsToAdd int) error {
	query := `
		INSERT INTO customer_product_points (store_id, customer_id, product_id, points, total_points)
		VALUES ($1, $2, $3, $4, $4)
		ON CONFLICT (store_id, customer_id, product_id)
		DO UPDATE SET 
			points = customer_product_points.points + $4,
			total_points = customer_product_points.total_points + $4,
			updated_at = NOW()
	`
	_, err := r.db.ExecContext(ctx, query, storeID, customerID, productID, pointsToAdd)
	return err
}

func (r *pointsRepository) DeductProductPoints(ctx context.Context, storeID, customerID, productID int64, pointsToDeduct int) error {
	query := `
		UPDATE customer_product_points
		SET points = points - $4, updated_at = NOW()
		WHERE store_id = $1 AND customer_id = $2 AND product_id = $3 AND points >= $4
	`
	result, err := r.db.ExecContext(ctx, query, storeID, customerID, productID, pointsToDeduct)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("insufficient points or customer/product not found")
	}
	return nil
}

func (r *pointsRepository) CreatePointTransaction(ctx context.Context, tx *domain.PointTransaction) error {
	query := `
		INSERT INTO point_transactions (
			store_id, branch_id, customer_id, transaction_type, points_change,
			reference_table, reference_id, product_id, note, staff_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := r.db.ExecContext(ctx, query,
		tx.StoreID, tx.BranchID, tx.CustomerID, tx.TransactionType, tx.PointsChange,
		tx.ReferenceTable, tx.ReferenceID, tx.ProductID, tx.Note, tx.StaffID,
	)
	return err
}

func (r *pointsRepository) CreateRedemption(ctx context.Context, redemption *domain.PointRedemption) (int64, error) {
	query := `
		INSERT INTO point_redemptions (
			store_id, branch_id, customer_id, product_id, points_used, quantity, staff_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`
	var id int64
	err := r.db.QueryRowContext(ctx, query,
		redemption.StoreID, redemption.BranchID, redemption.CustomerID,
		redemption.ProductID, redemption.PointsUsed, redemption.Quantity, redemption.StaffID,
	).Scan(&id)
	return id, err
}

func (r *pointsRepository) GetRedeemableProducts(ctx context.Context, storeID, branchID int64) ([]domain.RedeemableProduct, error) {
	query := `
		SELECT 
			p.id,
			p.product_name,
			c.category_name,
			p.image_path,
			p.points_to_redeem,
			COALESCE(bp.on_stock, 0) as on_stock
		FROM products p
		LEFT JOIN categories c ON p.category_id = c.id
		LEFT JOIN branch_products bp ON p.id = bp.product_id AND bp.branch_id = $2
		WHERE p.store_id = $1 
			AND p.is_active = true 
			AND p.points_to_redeem IS NOT NULL 
			AND p.points_to_redeem > 0
		ORDER BY p.points_to_redeem ASC, p.product_name ASC
	`
	var products []domain.RedeemableProduct
	err := r.db.SelectContext(ctx, &products, query, storeID, branchID)
	if err != nil {
		return nil, err
	}
	return products, nil
}

func (r *pointsRepository) GetProductPointsToRedeem(ctx context.Context, productID int64) (*int, error) {
	var pointsToRedeem sql.NullInt64
	query := `SELECT points_to_redeem FROM products WHERE id = $1 AND is_active = true`
	err := r.db.GetContext(ctx, &pointsToRedeem, query, productID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("product not found")
		}
		return nil, err
	}
	if !pointsToRedeem.Valid {
		return nil, errors.New("product is not redeemable")
	}
	points := int(pointsToRedeem.Int64)
	return &points, nil
}

func (r *pointsRepository) GetPointHistory(ctx context.Context, storeID, customerID int64, limit, offset int) ([]domain.PointHistoryItem, int, error) {
	var total int
	countQuery := `
		SELECT COUNT(*) FROM point_transactions
		WHERE store_id = $1 AND customer_id = $2
	`
	err := r.db.GetContext(ctx, &total, countQuery, storeID, customerID)
	if err != nil {
		return nil, 0, err
	}

	query := `
		SELECT 
			pt.id,
			pt.transaction_type,
			pt.points_change,
			p.product_name,
			pt.note,
			pt.created_at
		FROM point_transactions pt
		LEFT JOIN products p ON pt.product_id = p.id
		WHERE pt.store_id = $1 AND pt.customer_id = $2
		ORDER BY pt.created_at DESC
		LIMIT $3 OFFSET $4
	`
	var history []domain.PointHistoryItem
	err = r.db.SelectContext(ctx, &history, query, storeID, customerID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	return history, total, nil
}
