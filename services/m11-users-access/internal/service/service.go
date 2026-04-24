package service

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"time"

	"gpsgo/services/m11-users-access/internal/config"
	"gpsgo/services/m11-users-access/internal/repository"

	jwt "github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	privateKey *rsa.PrivateKey
	repo       *repository.Repository
}

type Tokens struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    int64  `json:"expiresIn"`
}

func New(cfg config.Config, repo *repository.Repository) *Service {
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(cfg.JWTPrivateKeyPEM)
	if err != nil {
		panic(err)
	}
	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(cfg.JWTPublicKeyPEM)
	if err != nil {
		panic(err)
	}
	_ = publicKey
	return &Service{privateKey: privateKey, repo: repo}
}

func (s *Service) Login(ctx context.Context, email, password, ip, userAgent string) (Tokens, error) {
	user, err := s.repo.FindUserByEmail(ctx, email)
	if err != nil {
		return Tokens{}, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return Tokens{}, errors.New("invalid credentials")
	}
	accessToken, err := s.signToken(user.ID, user.TenantID, user.RoleName, 15*time.Minute)
	if err != nil {
		return Tokens{}, err
	}
	refreshToken, err := randomToken(48)
	if err != nil {
		return Tokens{}, err
	}
	if err := s.repo.UpsertSession(ctx, user.ID, refreshToken, ip, userAgent, time.Now().UTC().Add(7*24*time.Hour)); err != nil {
		return Tokens{}, err
	}
	_ = s.repo.MarkLogin(ctx, user.ID)
	return Tokens{AccessToken: accessToken, RefreshToken: refreshToken, ExpiresIn: 900}, nil
}

func (s *Service) Refresh(ctx context.Context, refreshToken string) (Tokens, error) {
	userID, tenantID, role, err := s.repo.FindSessionByRefreshToken(ctx, refreshToken)
	if err != nil {
		return Tokens{}, errors.New("invalid refresh token")
	}
	accessToken, err := s.signToken(userID, tenantID, role, 15*time.Minute)
	if err != nil {
		return Tokens{}, err
	}
	newRefresh, err := randomToken(48)
	if err != nil {
		return Tokens{}, err
	}
	if err := s.repo.UpsertSession(ctx, userID, newRefresh, "", "", time.Now().UTC().Add(7*24*time.Hour)); err != nil {
		return Tokens{}, err
	}
	return Tokens{AccessToken: accessToken, RefreshToken: newRefresh, ExpiresIn: 900}, nil
}

func (s *Service) signToken(userID, tenantID, role string, ttl time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"sub":      userID,
		"userId":   userID,
		"tenantId": tenantID,
		"role":     role,
		"exp":      time.Now().UTC().Add(ttl).Unix(),
		"iat":      time.Now().UTC().Unix(),
	}
	tkn := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return tkn.SignedString(s.privateKey)
}

func randomToken(length int) (string, error) {
	buf := make([]byte, length)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	out := make([]byte, length)
	for i := range buf {
		out[i] = alphabet[int(buf[i])%len(alphabet)]
	}
	return string(out), nil
}
