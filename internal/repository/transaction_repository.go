package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/mini-membership/api/internal/domain"
)

type TransactionRepository interface {
	Create(ctx context.Context, tx *domain.MemberPointTransaction) error
	ListByMember(ctx context.Context, memberID uuid.UUID, page, limit int) ([]domain.MemberPointTransaction, int, error)
	ListByBranch(ctx context.Context, branch string, page, limit int) ([]domain.MemberPointTransaction, int, error)
}

type transactionRepository struct {
	db *sqlx.DB
}

func NewTransactionRepository(db *sqlx.DB) TransactionRepository {
	return &transactionRepository{db: db}
}

func (r *transactionRepository) Create(ctx context.Context, tx *domain.MemberPointTransaction) error {
	query := `
		INSERT INTO member_point_transactions (
			id, member_id, staff_user_id, action, product_type, points, receipt_text, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.ExecContext(ctx, query,
		tx.ID, tx.MemberID, tx.StaffUserID, tx.Action,
		tx.ProductType, tx.Points, tx.ReceiptText, tx.CreatedAt,
	)
	return err
}

func (r *transactionRepository) ListByMember(ctx context.Context, memberID uuid.UUID, page, limit int) ([]domain.MemberPointTransaction, int, error) {
	var transactions []domain.MemberPointTransaction
	var total int

	offset := (page - 1) * limit

	countQuery := "SELECT COUNT(*) FROM member_point_transactions WHERE member_id = $1"
	err := r.db.GetContext(ctx, &total, countQuery, memberID)
	if err != nil {
		return nil, 0, err
	}

	query := `
		SELECT id, member_id, staff_user_id, action, product_type, points, receipt_text, created_at
		FROM member_point_transactions
		WHERE member_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	err = r.db.SelectContext(ctx, &transactions, query, memberID, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	return transactions, total, nil
}

func (r *transactionRepository) ListByBranch(ctx context.Context, branch string, page, limit int) ([]domain.MemberPointTransaction, int, error) {
	var transactions []domain.MemberPointTransaction
	var total int

	offset := (page - 1) * limit

	countQuery := `
		SELECT COUNT(*) 
		FROM member_point_transactions mpt
		JOIN members m ON mpt.member_id = m.id
		WHERE m.branch = $1
	`
	err := r.db.GetContext(ctx, &total, countQuery, branch)
	if err != nil {
		return nil, 0, err
	}

	query := `
		SELECT mpt.id, mpt.member_id, mpt.staff_user_id, mpt.action, 
		       mpt.product_type, mpt.points, mpt.receipt_text, mpt.created_at
		FROM member_point_transactions mpt
		JOIN members m ON mpt.member_id = m.id
		WHERE m.branch = $1
		ORDER BY mpt.created_at DESC
		LIMIT $2 OFFSET $3
	`

	err = r.db.SelectContext(ctx, &transactions, query, branch, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	return transactions, total, nil
}
