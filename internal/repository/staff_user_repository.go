package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/katom-membership/api/internal/domain"
)

type StaffUserRepository interface {
	Create(ctx context.Context, staffUser *domain.StaffUser) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.StaffUser, error)
	GetByEmail(ctx context.Context, email string) (*domain.StaffUser, error)
	Update(ctx context.Context, staffUser *domain.StaffUser) error
}

type staffUserRepository struct {
	db *sqlx.DB
}

func NewStaffUserRepository(db *sqlx.DB) StaffUserRepository {
	return &staffUserRepository{db: db}
}

func (r *staffUserRepository) Create(ctx context.Context, staffUser *domain.StaffUser) error {
	query := `
		INSERT INTO staff_users (id, email, password_hash, branch, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.ExecContext(ctx, query,
		staffUser.ID,
		staffUser.Email,
		staffUser.PasswordHash,
		staffUser.Branch,
		staffUser.CreatedAt,
		staffUser.UpdatedAt,
	)
	return err
}

func (r *staffUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.StaffUser, error) {
	var staffUser domain.StaffUser
	query := `SELECT id, email, password_hash, branch, created_at, updated_at FROM staff_users WHERE id = $1`

	err := r.db.GetContext(ctx, &staffUser, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &staffUser, nil
}

func (r *staffUserRepository) GetByEmail(ctx context.Context, email string) (*domain.StaffUser, error) {
	var staffUser domain.StaffUser
	query := `SELECT id, email, password_hash, branch, created_at, updated_at FROM staff_users WHERE email = $1`

	err := r.db.GetContext(ctx, &staffUser, query, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &staffUser, nil
}

func (r *staffUserRepository) Update(ctx context.Context, staffUser *domain.StaffUser) error {
	query := `
		UPDATE staff_users 
		SET email = $1, password_hash = $2, branch = $3, updated_at = $4
		WHERE id = $5
	`
	_, err := r.db.ExecContext(ctx, query,
		staffUser.Email,
		staffUser.PasswordHash,
		staffUser.Branch,
		staffUser.UpdatedAt,
		staffUser.ID,
	)
	return err
}
