package repository

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/mini-membership/api/internal/domain"
	"github.com/shopspring/decimal"
)

type PromotionRepository interface {
	GetActivePromotions(ctx context.Context, storeID, branchID int64) ([]domain.PromotionResponse, error)
	GetPromotionByID(ctx context.Context, storeID, promotionID int64) (*domain.PromotionResponse, error)
	GetPromotionProducts(ctx context.Context, promotionID int64) ([]domain.PromotionProduct, error)
}

type promotionRepository struct {
	db *sqlx.DB
}

func NewPromotionRepository(db *sqlx.DB) PromotionRepository {
	return &promotionRepository{db: db}
}

type promotionRow struct {
	ID                    int64           `db:"id"`
	PromotionName         string          `db:"promotion_name"`
	IsActive              bool            `db:"is_active"`
	StartsAt              sql.NullTime    `db:"starts_at"`
	EndsAt                sql.NullTime    `db:"ends_at"`
	TypeID                int64           `db:"type_id"`
	TypeName              string          `db:"type_name"`
	TypeDetail            sql.NullString  `db:"type_detail"`
	PercentDiscount       decimal.NullDecimal `db:"percent_discount"`
	BahtDiscount          decimal.NullDecimal `db:"baht_discount"`
	TotalPriceSetDiscount decimal.NullDecimal `db:"total_price_set_discount"`
	OldPriceSet           decimal.NullDecimal `db:"old_price_set"`
	CountConditionProduct sql.NullInt64   `db:"count_condition_product"`
	ProductCount          int             `db:"product_count"`
}

func (r *promotionRepository) GetActivePromotions(ctx context.Context, storeID, branchID int64) ([]domain.PromotionResponse, error) {
	query := `
		SELECT 
			p.id,
			p.promotion_name,
			p.is_active,
			p.starts_at,
			p.ends_at,
			pt.id as type_id,
			pt.name as type_name,
			pt.detail as type_detail,
			pc.percent_discount,
			pc.baht_discount,
			pc.total_price_set_discount,
			pc.old_price_set,
			pc.count_condition_product,
			COALESCE((SELECT COUNT(*) FROM promotion_products pp WHERE pp.promotion_id = p.id), 0) as product_count
		FROM promotions p
		JOIN promotion_types pt ON pt.id = p.promotion_type_id
		LEFT JOIN promotion_configs pc ON pc.promotion_id = p.id
		WHERE p.store_id = $1 
			AND p.is_active = true
			AND (p.starts_at IS NULL OR p.starts_at <= NOW())
			AND (p.ends_at IS NULL OR p.ends_at >= NOW())
			AND EXISTS (
				SELECT 1 FROM promotion_type_branches ptb 
				WHERE ptb.promotion_type_id = pt.id AND ptb.branch_id = $2
			)
		ORDER BY p.id
	`

	var rows []promotionRow
	if err := r.db.SelectContext(ctx, &rows, query, storeID, branchID); err != nil {
		return nil, err
	}

	promotions := make([]domain.PromotionResponse, 0, len(rows))
	for _, row := range rows {
		promo := domain.PromotionResponse{
			ID:            row.ID,
			PromotionName: row.PromotionName,
			IsActive:      row.IsActive,
			IsBillLevel:   row.ProductCount == 0,
			PromotionType: domain.PromotionTypeInfo{
				ID:   row.TypeID,
				Name: row.TypeName,
			},
			Config: domain.PromotionConfig{},
		}

		if row.TypeDetail.Valid {
			promo.PromotionType.Detail = row.TypeDetail.String
		}
		if row.StartsAt.Valid {
			promo.StartsAt = &row.StartsAt.Time
		}
		if row.EndsAt.Valid {
			promo.EndsAt = &row.EndsAt.Time
		}
		if row.PercentDiscount.Valid {
			v, _ := row.PercentDiscount.Decimal.Float64()
			promo.Config.PercentDiscount = &v
		}
		if row.BahtDiscount.Valid {
			v, _ := row.BahtDiscount.Decimal.Float64()
			promo.Config.BahtDiscount = &v
		}
		if row.TotalPriceSetDiscount.Valid {
			v, _ := row.TotalPriceSetDiscount.Decimal.Float64()
			promo.Config.TotalPriceSetDiscount = &v
		}
		if row.OldPriceSet.Valid {
			v, _ := row.OldPriceSet.Decimal.Float64()
			promo.Config.OldPriceSet = &v
		}
		if row.CountConditionProduct.Valid {
			v := int(row.CountConditionProduct.Int64)
			promo.Config.CountConditionProduct = &v
		}

		// Get products for this promotion
		products, err := r.GetPromotionProducts(ctx, row.ID)
		if err != nil {
			return nil, err
		}
		promo.Products = products

		promotions = append(promotions, promo)
	}

	return promotions, nil
}

func (r *promotionRepository) GetPromotionByID(ctx context.Context, storeID, promotionID int64) (*domain.PromotionResponse, error) {
	query := `
		SELECT 
			p.id,
			p.promotion_name,
			p.is_active,
			p.starts_at,
			p.ends_at,
			pt.id as type_id,
			pt.name as type_name,
			pt.detail as type_detail,
			pc.percent_discount,
			pc.baht_discount,
			pc.total_price_set_discount,
			pc.old_price_set,
			pc.count_condition_product,
			COALESCE((SELECT COUNT(*) FROM promotion_products pp WHERE pp.promotion_id = p.id), 0) as product_count
		FROM promotions p
		JOIN promotion_types pt ON pt.id = p.promotion_type_id
		LEFT JOIN promotion_configs pc ON pc.promotion_id = p.id
		WHERE p.store_id = $1 AND p.id = $2
	`

	var row promotionRow
	if err := r.db.GetContext(ctx, &row, query, storeID, promotionID); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	promo := &domain.PromotionResponse{
		ID:            row.ID,
		PromotionName: row.PromotionName,
		IsActive:      row.IsActive,
		IsBillLevel:   row.ProductCount == 0,
		PromotionType: domain.PromotionTypeInfo{
			ID:   row.TypeID,
			Name: row.TypeName,
		},
		Config: domain.PromotionConfig{},
	}

	if row.TypeDetail.Valid {
		promo.PromotionType.Detail = row.TypeDetail.String
	}
	if row.StartsAt.Valid {
		promo.StartsAt = &row.StartsAt.Time
	}
	if row.EndsAt.Valid {
		promo.EndsAt = &row.EndsAt.Time
	}
	if row.PercentDiscount.Valid {
		v, _ := row.PercentDiscount.Decimal.Float64()
		promo.Config.PercentDiscount = &v
	}
	if row.BahtDiscount.Valid {
		v, _ := row.BahtDiscount.Decimal.Float64()
		promo.Config.BahtDiscount = &v
	}
	if row.TotalPriceSetDiscount.Valid {
		v, _ := row.TotalPriceSetDiscount.Decimal.Float64()
		promo.Config.TotalPriceSetDiscount = &v
	}
	if row.OldPriceSet.Valid {
		v, _ := row.OldPriceSet.Decimal.Float64()
		promo.Config.OldPriceSet = &v
	}
	if row.CountConditionProduct.Valid {
		v := int(row.CountConditionProduct.Int64)
		promo.Config.CountConditionProduct = &v
	}

	products, err := r.GetPromotionProducts(ctx, row.ID)
	if err != nil {
		return nil, err
	}
	promo.Products = products

	return promo, nil
}

func (r *promotionRepository) GetPromotionProducts(ctx context.Context, promotionID int64) ([]domain.PromotionProduct, error) {
	query := `
		SELECT 
			pp.product_id,
			p.product_name,
			p.base_price
		FROM promotion_products pp
		JOIN products p ON p.id = pp.product_id
		WHERE pp.promotion_id = $1
		ORDER BY p.product_name
	`

	type productRow struct {
		ProductID   int64           `db:"product_id"`
		ProductName string          `db:"product_name"`
		BasePrice   decimal.Decimal `db:"base_price"`
	}

	var rows []productRow
	if err := r.db.SelectContext(ctx, &rows, query, promotionID); err != nil {
		return nil, err
	}

	products := make([]domain.PromotionProduct, 0, len(rows))
	for _, row := range rows {
		price, _ := row.BasePrice.Float64()
		products = append(products, domain.PromotionProduct{
			ProductID:   row.ProductID,
			ProductName: row.ProductName,
			BasePrice:   price,
		})
	}

	return products, nil
}
