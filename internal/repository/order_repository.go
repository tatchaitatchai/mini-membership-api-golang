package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/shopspring/decimal"
)

type OrderRepository interface {
	GetProductsByBranch(ctx context.Context, storeID, branchID int64) ([]BranchProductInfo, error)
	SearchCustomersByLast4(ctx context.Context, storeID int64, last4 string) ([]CustomerSearchResult, error)
	CreateOrderTx(ctx context.Context, order *OrderCreate) (*OrderResult, error)
	GetOrdersByShift(ctx context.Context, storeID, branchID, shiftID int64) ([]OrderWithItems, error)
	GetOrderByID(ctx context.Context, storeID, orderID int64) (*OrderWithItems, error)
}

type BranchProductInfo struct {
	ProductID    int64           `db:"product_id"`
	ProductName  string          `db:"product_name"`
	CategoryName sql.NullString  `db:"category_name"`
	BasePrice    decimal.Decimal `db:"base_price"`
	ImagePath    sql.NullString  `db:"image_path"`
	OnStock      int             `db:"on_stock"`
}

type CustomerSearchResult struct {
	ID           int64          `db:"id"`
	CustomerCode sql.NullString `db:"customer_code"`
	FullName     sql.NullString `db:"full_name"`
	PhoneLast4   sql.NullString `db:"phone_last4"`
}

type OrderCreate struct {
	StoreID       int64
	BranchID      int64
	ShiftID       int64
	StaffID       int64
	CustomerID    *int64
	Items         []OrderItemCreate
	Subtotal      decimal.Decimal
	DiscountTotal decimal.Decimal
	TotalPrice    decimal.Decimal
	ChangeAmount  decimal.Decimal
	Payments      []PaymentCreate
	PromotionID   *int64
}

type OrderItemCreate struct {
	ProductID int64
	Quantity  int
	Price     decimal.Decimal
}

type PaymentCreate struct {
	Method string
	Amount decimal.Decimal
}

type OrderResult struct {
	OrderID   int64
	Status    string
	CreatedAt time.Time
}

type OrderWithItems struct {
	ID            int64           `db:"id"`
	CustomerID    sql.NullInt64   `db:"customer_id"`
	CustomerName  sql.NullString  `db:"customer_name"`
	StaffName     sql.NullString  `db:"staff_name"`
	Subtotal      decimal.Decimal `db:"subtotal"`
	DiscountTotal decimal.Decimal `db:"discount_total"`
	TotalPrice    decimal.Decimal `db:"total_price"`
	ChangeAmount  decimal.Decimal `db:"change_amount"`
	Status        string          `db:"status"`
	CreatedAt     time.Time       `db:"created_at"`
	Items         []OrderItemResult
}

type OrderItemResult struct {
	ProductID   int64           `db:"product_id"`
	ProductName string          `db:"product_name"`
	Quantity    int             `db:"quantity"`
	Price       decimal.Decimal `db:"price"`
}

type orderRepository struct {
	db *sqlx.DB
}

func NewOrderRepository(db *sqlx.DB) OrderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) GetProductsByBranch(ctx context.Context, storeID, branchID int64) ([]BranchProductInfo, error) {
	var products []BranchProductInfo
	query := `
		SELECT 
			p.id as product_id,
			p.product_name,
			c.category_name,
			p.base_price,
			p.image_path,
			COALESCE(bp.on_stock, 0) as on_stock
		FROM products p
		LEFT JOIN categories c ON c.id = p.category_id AND c.store_id = p.store_id
		LEFT JOIN branch_products bp ON bp.product_id = p.id AND bp.branch_id = $2 AND bp.store_id = p.store_id
		WHERE p.store_id = $1 AND p.is_active = true
		ORDER BY c.category_name NULLS LAST, p.product_name
	`
	err := r.db.SelectContext(ctx, &products, query, storeID, branchID)
	if err != nil {
		return nil, err
	}
	return products, nil
}

func (r *orderRepository) SearchCustomersByLast4(ctx context.Context, storeID int64, last4 string) ([]CustomerSearchResult, error) {
	var customers []CustomerSearchResult
	query := `
		SELECT id, customer_code, full_name, phone_last4
		FROM customers
		WHERE store_id = $1 AND phone_last4 = $2 AND is_active = true
		ORDER BY full_name
	`

	err := r.db.SelectContext(ctx, &customers, query, storeID, last4)
	if err != nil {
		return nil, err
	}
	return customers, nil
}

func (r *orderRepository) CreateOrderTx(ctx context.Context, order *OrderCreate) (*OrderResult, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	now := time.Now()

	// 1. Create order
	orderQuery := `
		INSERT INTO orders (store_id, branch_id, shift_id, customer_id, staff_id, subtotal, discount_total, total_price, change_amount, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, 'PAID', $10, $10)
		RETURNING id
	`
	var orderID int64
	err = tx.QueryRowContext(ctx, orderQuery,
		order.StoreID,
		order.BranchID,
		order.ShiftID,
		order.CustomerID,
		order.StaffID,
		order.Subtotal,
		order.DiscountTotal,
		order.TotalPrice,
		order.ChangeAmount,
		now,
	).Scan(&orderID)
	if err != nil {
		return nil, err
	}

	// 2. Create order items and deduct stock
	for _, item := range order.Items {
		// Get current stock
		var currentStock int
		var branchProductID int64
		stockQuery := `SELECT id, on_stock FROM branch_products WHERE store_id = $1 AND branch_id = $2 AND product_id = $3 FOR UPDATE`
		err = tx.QueryRowContext(ctx, stockQuery, order.StoreID, order.BranchID, item.ProductID).Scan(&branchProductID, &currentStock)
		if err == sql.ErrNoRows {
			// Create branch_product if not exists
			insertBPQuery := `
				INSERT INTO branch_products (store_id, branch_id, product_id, on_stock, is_active, created_at, updated_at)
				VALUES ($1, $2, $3, 0, true, $4, $4)
				RETURNING id, on_stock
			`
			err = tx.QueryRowContext(ctx, insertBPQuery, order.StoreID, order.BranchID, item.ProductID, now).Scan(&branchProductID, &currentStock)
			if err != nil {
				return nil, err
			}
		} else if err != nil {
			return nil, err
		}

		newStock := currentStock - item.Quantity
		if newStock < 0 {
			newStock = 0
		}

		// Insert order item
		itemQuery := `
			INSERT INTO order_items (order_id, product_id, quantity, price, from_stock_count, to_stock_count, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $7)
		`
		_, err = tx.ExecContext(ctx, itemQuery, orderID, item.ProductID, item.Quantity, item.Price, currentStock, newStock, now)
		if err != nil {
			return nil, err
		}

		// Update stock
		updateStockQuery := `UPDATE branch_products SET on_stock = $1, updated_at = $2 WHERE id = $3`
		_, err = tx.ExecContext(ctx, updateStockQuery, newStock, now, branchProductID)
		if err != nil {
			return nil, err
		}

		// Log inventory movement
		movementQuery := `
			INSERT INTO inventory_movements (store_id, branch_id, product_id, movement_type, quantity_change, from_stock_count, to_stock_count, changed_by, reference_table, reference_id, created_at)
			VALUES ($1, $2, $3, 'SALE', $4, $5, $6, $7, 'orders', $8, $9)
		`
		_, err = tx.ExecContext(ctx, movementQuery, order.StoreID, order.BranchID, item.ProductID, -item.Quantity, currentStock, newStock, order.StaffID, orderID, now)
		if err != nil {
			return nil, err
		}
	}

	// 3. Create payments
	for _, payment := range order.Payments {
		paymentQuery := `
			INSERT INTO payments (order_id, method, amount, paid_at, created_at)
			VALUES ($1, $2, $3, $4, $4)
		`
		_, err = tx.ExecContext(ctx, paymentQuery, orderID, payment.Method, payment.Amount, now)
		if err != nil {
			return nil, err
		}
	}

	// 4. Create order promotion if provided
	if order.PromotionID != nil {
		promoQuery := `
			INSERT INTO order_promotions (order_id, promotion_id, discount_amount, metadata, created_at)
			VALUES ($1, $2, $3, '{}', $4)
		`
		_, err = tx.ExecContext(ctx, promoQuery, orderID, *order.PromotionID, order.DiscountTotal, now)
		if err != nil {
			return nil, err
		}
	}

	// 5. Record cash movement for change given (เงินทอน)
	if order.ChangeAmount.GreaterThan(decimal.Zero) {
		cashMovementQuery := `
			INSERT INTO shift_cash_movements (store_id, branch_id, shift_id, movement_type, direction, amount, note, created_by_staff_id, created_at)
			VALUES ($1, $2, $3, 'PAID_OUT', 'OUT', $4, $5, $6, $7)
		`
		note := fmt.Sprintf("เงินทอนจาก Order #%d", orderID)
		_, err = tx.ExecContext(ctx, cashMovementQuery, order.StoreID, order.BranchID, order.ShiftID, order.ChangeAmount, note, order.StaffID, now)
		if err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &OrderResult{
		OrderID:   orderID,
		Status:    "PAID",
		CreatedAt: now,
	}, nil
}

func (r *orderRepository) GetOrdersByShift(ctx context.Context, storeID, branchID, shiftID int64) ([]OrderWithItems, error) {
	var orders []OrderWithItems
	query := `
		SELECT 
			o.id, o.customer_id, c.full_name as customer_name, 
			COALESCE(s.email, 'Staff') as staff_name,
			o.subtotal, o.discount_total, o.total_price, o.change_amount, o.status, o.created_at
		FROM orders o
		LEFT JOIN customers c ON c.id = o.customer_id
		LEFT JOIN staff_accounts s ON s.id = o.staff_id
		WHERE o.store_id = $1 AND o.branch_id = $2 AND o.shift_id = $3
		ORDER BY o.created_at DESC
	`
	err := r.db.SelectContext(ctx, &orders, query, storeID, branchID, shiftID)
	if err != nil {
		return nil, err
	}

	// Load items for each order
	for i := range orders {
		items, err := r.getOrderItems(ctx, orders[i].ID)
		if err != nil {
			return nil, err
		}
		orders[i].Items = items
	}

	return orders, nil
}

func (r *orderRepository) GetOrderByID(ctx context.Context, storeID, orderID int64) (*OrderWithItems, error) {
	var order OrderWithItems
	query := `
		SELECT 
			o.id, o.customer_id, c.full_name as customer_name,
			COALESCE(s.email, 'Staff') as staff_name,
			o.subtotal, o.discount_total, o.total_price, o.change_amount, o.status, o.created_at
		FROM orders o
		LEFT JOIN customers c ON c.id = o.customer_id
		LEFT JOIN staff_accounts s ON s.id = o.staff_id
		WHERE o.id = $1 AND o.store_id = $2
	`
	err := r.db.GetContext(ctx, &order, query, orderID, storeID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	items, err := r.getOrderItems(ctx, orderID)
	if err != nil {
		return nil, err
	}
	order.Items = items

	return &order, nil
}

func (r *orderRepository) getOrderItems(ctx context.Context, orderID int64) ([]OrderItemResult, error) {
	var items []OrderItemResult
	query := `
		SELECT oi.product_id, p.product_name, oi.quantity, oi.price
		FROM order_items oi
		JOIN products p ON p.id = oi.product_id
		WHERE oi.order_id = $1
	`
	err := r.db.SelectContext(ctx, &items, query, orderID)
	if err != nil {
		return nil, err
	}
	return items, nil
}
