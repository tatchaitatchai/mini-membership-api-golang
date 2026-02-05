package service

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/mini-membership/api/internal/domain"
	"github.com/mini-membership/api/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Login(ctx context.Context, req *domain.LoginRequest) (*domain.LoginResponse, error)
	ValidateToken(tokenString string) (*Claims, error)
	CreateStaffUser(ctx context.Context, email, password, branch string) error
}

type authService struct {
	staffUserRepo repository.StaffUserRepository
	jwtSecret     string
	jwtExpiration time.Duration
}

type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	Branch string    `json:"branch"`
	jwt.RegisteredClaims
}

func NewAuthService(staffUserRepo repository.StaffUserRepository, jwtSecret string, jwtExpiration time.Duration) AuthService {
	return &authService{
		staffUserRepo: staffUserRepo,
		jwtSecret:     jwtSecret,
		jwtExpiration: jwtExpiration,
	}
}

func (s *authService) Login(ctx context.Context, req *domain.LoginRequest) (*domain.LoginResponse, error) {
	staffUser, err := s.staffUserRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if staffUser == nil {
		return nil, errors.New("invalid email or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(staffUser.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid email or password")
	}

	token, err := s.generateToken(staffUser)
	if err != nil {
		return nil, err
	}

	return &domain.LoginResponse{
		Token:     token,
		StaffUser: staffUser,
	}, nil
}

func (s *authService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

func (s *authService) CreateStaffUser(ctx context.Context, email, password, branch string) error {
	existingUser, err := s.staffUserRepo.GetByEmail(ctx, email)
	if err != nil {
		return err
	}
	if existingUser != nil {
		return errors.New("email already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	staffUser := &domain.StaffUser{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: string(hashedPassword),
		Branch:       branch,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	return s.staffUserRepo.Create(ctx, staffUser)
}

func (s *authService) generateToken(staffUser *domain.StaffUser) (string, error) {
	claims := &Claims{
		UserID: staffUser.ID,
		Email:  staffUser.Email,
		Branch: staffUser.Branch,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.jwtExpiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}
