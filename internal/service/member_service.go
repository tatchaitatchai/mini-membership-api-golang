package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/katom-membership/api/internal/domain"
	"github.com/katom-membership/api/internal/repository"
)

type MemberService interface {
	Create(ctx context.Context, req *domain.MemberCreateRequest, staffBranch, registeredByStaff string) (*domain.Member, error)
	GetByID(ctx context.Context, memberID uuid.UUID, staffBranch string) (*domain.Member, error)
	Update(ctx context.Context, memberID uuid.UUID, req *domain.MemberUpdateRequest, staffBranch string) (*domain.Member, error)
	List(ctx context.Context, staffBranch string, page, limit int) (*domain.MemberListResponse, error)
}

type memberService struct {
	memberRepo repository.MemberRepository
}

func NewMemberService(memberRepo repository.MemberRepository) MemberService {
	return &memberService{
		memberRepo: memberRepo,
	}
}

func (s *memberService) Create(ctx context.Context, req *domain.MemberCreateRequest, staffBranch, registeredByStaff string) (*domain.Member, error) {
	member := &domain.Member{
		ID:                        uuid.New(),
		Name:                      req.Name,
		Last4:                     req.Last4,
		TotalPoints:               0,
		MilestoneScore:            0,
		Points1_0Liter:            0,
		Points1_5Liter:            0,
		Branch:                    &staffBranch,
		Status:                    "active",
		MembershipNumber:          req.MembershipNumber,
		RegistrationReceiptNumber: req.RegistrationReceiptNumber,
		WelcomeBonusClaimed:       false,
		RegisteredByStaff:         &registeredByStaff,
		CreatedAt:                 time.Now(),
		UpdatedAt:                 time.Now(),
	}

	if req.Branch != nil {
		member.Branch = req.Branch
	}

	err := s.memberRepo.Create(ctx, member)
	if err != nil {
		return nil, err
	}

	return member, nil
}

func (s *memberService) GetByID(ctx context.Context, memberID uuid.UUID, staffBranch string) (*domain.Member, error) {
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

	return member, nil
}

func (s *memberService) Update(ctx context.Context, memberID uuid.UUID, req *domain.MemberUpdateRequest, staffBranch string) (*domain.Member, error) {
	member, err := s.GetByID(ctx, memberID, staffBranch)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		member.Name = *req.Name
	}
	if req.Last4 != nil {
		member.Last4 = req.Last4
	}
	if req.Branch != nil {
		member.Branch = req.Branch
	}
	if req.Status != nil {
		member.Status = *req.Status
	}

	member.UpdatedAt = time.Now()

	err = s.memberRepo.Update(ctx, member)
	if err != nil {
		return nil, err
	}

	return member, nil
}

func (s *memberService) List(ctx context.Context, staffBranch string, page, limit int) (*domain.MemberListResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	members, total, err := s.memberRepo.List(ctx, &staffBranch, page, limit)
	if err != nil {
		return nil, err
	}

	return &domain.MemberListResponse{
		Members: members,
		Total:   total,
		Page:    page,
		Limit:   limit,
	}, nil
}
