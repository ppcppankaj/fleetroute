package service

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gpsgo/services/m15-admin-panel/internal/repository"
)

type Service struct {
	repo      *repository.Repository
	jwtSecret string
}

func New(repo *repository.Repository, jwtSecret string) *Service {
	return &Service{repo: repo, jwtSecret: jwtSecret}
}

type LoginResult struct {
	Token  string                `json:"token"`
	Admin  repository.SuperAdmin `json:"admin"`
}

func (s *Service) Login(ctx context.Context, email, password string) (LoginResult, error) {
	admin, err := s.repo.FindSuperAdminByEmail(ctx, email)
	if err != nil {
		return LoginResult{}, errors.New("invalid credentials")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(password)); err != nil {
		return LoginResult{}, errors.New("invalid credentials")
	}
	_ = s.repo.MarkAdminLogin(ctx, admin.ID)

	claims := jwt.MapClaims{
		"sub":  admin.ID,
		"role": admin.Role,
		"iss":  "fleet-admin",
		"exp":  time.Now().Add(8 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return LoginResult{}, err
	}

	admin.PasswordHash = "" // scrub
	return LoginResult{Token: signed, Admin: admin}, nil
}

func (s *Service) ListTickets(ctx context.Context, status string) ([]repository.SupportTicket, error) {
	return s.repo.ListTickets(ctx, status)
}

func (s *Service) CreateTicket(ctx context.Context, t repository.SupportTicket) (repository.SupportTicket, error) {
	return s.repo.CreateTicket(ctx, t)
}
