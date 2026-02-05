package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mini-membership/api/internal/domain"
	"github.com/mini-membership/api/internal/repository"
)

type TransactionService interface {
	Create(ctx context.Context, req *domain.TransactionCreateRequest, staffUserID uuid.UUID, staffBranch string) (*domain.TransactionCreateResponse, error)
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

func (s *transactionService) Create(ctx context.Context, req *domain.TransactionCreateRequest, staffUserID uuid.UUID, staffBranch string) (*domain.TransactionCreateResponse, error) {
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

	newTotalPoints := member.TotalPoints
	newMilestoneScore := member.MilestoneScore
	newPoints1_0Liter := member.Points1_0Liter
	newPoints1_5Liter := member.Points1_5Liter

	var transactions []domain.MemberPointTransaction
	totalPointsInRequest := 0

	if req.Action == domain.ActionRedeem {
		redeem1_0Liter := 0
		redeem1_5Liter := 0

		for _, product := range req.Products {
			switch product.ProductType {
			case domain.ProductType1_0Liter:
				redeem1_0Liter += product.Points
			case domain.ProductType1_5Liter:
				redeem1_5Liter += product.Points
			}
		}

		if redeem1_0Liter > 0 && member.Points1_0Liter < redeem1_0Liter {
			return nil, fmt.Errorf("insufficient 1.0L points: member has %d points from 1.0L products but trying to redeem %d points", member.Points1_0Liter, redeem1_0Liter)
		}

		if redeem1_5Liter > 0 && member.Points1_5Liter < redeem1_5Liter {
			return nil, fmt.Errorf("insufficient 1.5L points: member has %d points from 1.5L products but trying to redeem %d points", member.Points1_5Liter, redeem1_5Liter)
		}
	}

	for _, product := range req.Products {
		transaction := &domain.MemberPointTransaction{
			ID:          uuid.New(),
			MemberID:    req.MemberID,
			StaffUserID: staffUserID,
			Action:      req.Action,
			ProductType: product.ProductType,
			Points:      product.Points,
			ReceiptText: req.ReceiptText,
			CreatedAt:   time.Now(),
		}

		err = s.transactionRepo.Create(ctx, transaction)
		if err != nil {
			return nil, err
		}

		transactions = append(transactions, *transaction)
		totalPointsInRequest += product.Points

		switch req.Action {
		case domain.ActionEarn:
			newTotalPoints += product.Points
			newMilestoneScore += product.Points

			switch product.ProductType {
			case domain.ProductType1_0Liter:
				newPoints1_0Liter += product.Points
			case domain.ProductType1_5Liter:
				newPoints1_5Liter += product.Points
			}
		case domain.ActionRedeem:
			newTotalPoints -= product.Points

			switch product.ProductType {
			case domain.ProductType1_0Liter:
				newPoints1_0Liter -= product.Points
			case domain.ProductType1_5Liter:
				newPoints1_5Liter -= product.Points
			}
		}
	}

	if newTotalPoints < 0 {
		newTotalPoints = 0
	}

	err = s.memberRepo.UpdatePoints(ctx, req.MemberID, newTotalPoints, newMilestoneScore, newPoints1_0Liter, newPoints1_5Liter)
	if err != nil {
		return nil, err
	}

	message := fmt.Sprintf("Successfully processed %d product(s) with %d total points", len(req.Products), totalPointsInRequest)

	return &domain.TransactionCreateResponse{
		Transactions: transactions,
		TotalPoints:  totalPointsInRequest,
		Message:      message,
	}, nil
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
