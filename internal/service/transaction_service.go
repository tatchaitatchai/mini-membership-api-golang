package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/katom-membership/api/internal/domain"
	"github.com/katom-membership/api/internal/repository"
)

type TransactionService interface {
	Create(ctx context.Context, req *domain.TransactionCreateRequest, staffUserID uuid.UUID, staffBranch string) (*domain.MemberPointTransaction, error)
	ListByMember(ctx context.Context, memberID uuid.UUID, staffBranch string, page, limit int) (*domain.TransactionListResponse, error)
	ListByBranch(ctx context.Context, staffBranch string, page, limit int) (*domain.TransactionListResponse, error)
}

type transactionService struct {
	transactionRepo repository.TransactionRepository
	memberRepo      repository.MemberRepository
}

func NewTransactionService(transactionRepo repository.TransactionRepository, memberRepo repository.MemberRepository) TransactionService {
	return &transactionService{
		transactionRepo: transactionRepo,
		memberRepo:      memberRepo,
	}
}

func (s *transactionService) Create(ctx context.Context, req *domain.TransactionCreateRequest, staffUserID uuid.UUID, staffBranch string) (*domain.MemberPointTransaction, error) {
	member, err := s.memberRepo.GetByID(ctx, req.MemberID)
	if err != nil {
		return nil, err
	}
	if member == nil {
		return nil, errors.New("member not found")
	}

	if member.Branch != nil && *member.Branch != staffBranch {
		return nil, errors.New("unauthorized: member belongs to different branch")
	}

	transaction := &domain.MemberPointTransaction{
		ID:          uuid.New(),
		MemberID:    req.MemberID,
		StaffUserID: staffUserID,
		Action:      req.Action,
		ProductType: req.ProductType,
		Points:      req.Points,
		ReceiptText: req.ReceiptText,
		CreatedAt:   time.Now(),
	}

	err = s.transactionRepo.Create(ctx, transaction)
	if err != nil {
		return nil, err
	}

	newTotalPoints := member.TotalPoints
	newMilestoneScore := member.MilestoneScore
	newPoints1_0Liter := member.Points1_0Liter
	newPoints1_5Liter := member.Points1_5Liter

	switch req.Action {
	case domain.ActionAdd:
		newTotalPoints += req.Points
		newMilestoneScore += req.Points

		switch req.ProductType {
		case domain.ProductType1_0Liter:
			newPoints1_0Liter += req.Points
		case domain.ProductType1_5Liter:
			newPoints1_5Liter += req.Points
		}
	case domain.ActionDeduct, domain.ActionRedeem:
		newTotalPoints -= req.Points
		if newTotalPoints < 0 {
			newTotalPoints = 0
		}
	case domain.ActionAdjust:
		newTotalPoints = req.Points
	}

	err = s.memberRepo.UpdatePoints(ctx, req.MemberID, newTotalPoints, newMilestoneScore, newPoints1_0Liter, newPoints1_5Liter)
	if err != nil {
		return nil, err
	}

	return transaction, nil
}

func (s *transactionService) ListByMember(ctx context.Context, memberID uuid.UUID, staffBranch string, page, limit int) (*domain.TransactionListResponse, error) {
	member, err := s.memberRepo.GetByID(ctx, memberID)
	if err != nil {
		return nil, err
	}
	if member == nil {
		return nil, errors.New("member not found")
	}

	if member.Branch != nil && *member.Branch != staffBranch {
		return nil, errors.New("unauthorized: member belongs to different branch")
	}

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	transactions, total, err := s.transactionRepo.ListByMember(ctx, memberID, page, limit)
	if err != nil {
		return nil, err
	}

	return &domain.TransactionListResponse{
		Transactions: transactions,
		Total:        total,
		Page:         page,
		Limit:        limit,
	}, nil
}

func (s *transactionService) ListByBranch(ctx context.Context, staffBranch string, page, limit int) (*domain.TransactionListResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	transactions, total, err := s.transactionRepo.ListByBranch(ctx, staffBranch, page, limit)
	if err != nil {
		return nil, err
	}

	return &domain.TransactionListResponse{
		Transactions: transactions,
		Total:        total,
		Page:         page,
		Limit:        limit,
	}, nil
}
