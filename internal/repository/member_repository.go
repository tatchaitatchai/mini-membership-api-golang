package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/katom-membership/api/internal/domain"
)

type MemberRepository interface {
	Create(ctx context.Context, member *domain.Member) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Member, error)
	Update(ctx context.Context, member *domain.Member) error
	List(ctx context.Context, branch *string, page, limit int) ([]domain.Member, int, error)
	UpdatePoints(ctx context.Context, memberID uuid.UUID, totalPoints, milestoneScore, points1_0Liter, points1_5Liter int) error
}

type memberRepository struct {
	db *sqlx.DB
}

func NewMemberRepository(db *sqlx.DB) MemberRepository {
	return &memberRepository{db: db}
}

func (r *memberRepository) Create(ctx context.Context, member *domain.Member) error {
	query := `
		INSERT INTO members (
			id, old_id, name, last4, total_points, milestone_score,
			points_1_0_liter, points_1_5_liter, branch, status,
			membership_number, registration_receipt_number,
			welcome_bonus_claimed, registered_by_staff, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	`
	_, err := r.db.ExecContext(ctx, query,
		member.ID, member.OldID, member.Name, member.Last4,
		member.TotalPoints, member.MilestoneScore,
		member.Points1_0Liter, member.Points1_5Liter,
		member.Branch, member.Status,
		member.MembershipNumber, member.RegistrationReceiptNumber,
		member.WelcomeBonusClaimed, member.RegisteredByStaff,
		member.CreatedAt, member.UpdatedAt,
	)
	return err
}

func (r *memberRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Member, error) {
	var member domain.Member
	query := `
		SELECT id, old_id, name, last4, total_points, milestone_score,
			   points_1_0_liter, points_1_5_liter, branch, status,
			   membership_number, registration_receipt_number,
			   welcome_bonus_claimed, registered_by_staff, created_at, updated_at
		FROM members WHERE id = $1
	`

	err := r.db.GetContext(ctx, &member, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &member, nil
}

func (r *memberRepository) Update(ctx context.Context, member *domain.Member) error {
	query := `
		UPDATE members 
		SET name = $1, last4 = $2, branch = $3, status = $4, updated_at = $5
		WHERE id = $6
	`
	_, err := r.db.ExecContext(ctx, query,
		member.Name, member.Last4, member.Branch, member.Status,
		member.UpdatedAt, member.ID,
	)
	return err
}

func (r *memberRepository) List(ctx context.Context, branch *string, page, limit int) ([]domain.Member, int, error) {
	var members []domain.Member
	var total int

	offset := (page - 1) * limit

	whereClause := ""
	args := []interface{}{limit, offset}

	if branch != nil && *branch != "" {
		whereClause = "WHERE branch = $3"
		args = append(args, *branch)
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM members %s", whereClause)
	err := r.db.GetContext(ctx, &total, countQuery, args[2:]...)
	if err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`
		SELECT id, old_id, name, last4, total_points, milestone_score,
			   points_1_0_liter, points_1_5_liter, branch, status,
			   membership_number, registration_receipt_number,
			   welcome_bonus_claimed, registered_by_staff, created_at, updated_at
		FROM members
		%s
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, whereClause)

	err = r.db.SelectContext(ctx, &members, query, args...)
	if err != nil {
		return nil, 0, err
	}

	return members, total, nil
}

func (r *memberRepository) UpdatePoints(ctx context.Context, memberID uuid.UUID, totalPoints, milestoneScore, points1_0Liter, points1_5Liter int) error {
	query := `
		UPDATE members 
		SET total_points = $1, 
		    milestone_score = $2,
		    points_1_0_liter = $3,
		    points_1_5_liter = $4,
		    updated_at = NOW()
		WHERE id = $5
	`
	_, err := r.db.ExecContext(ctx, query, totalPoints, milestoneScore, points1_0Liter, points1_5Liter, memberID)
	return err
}
