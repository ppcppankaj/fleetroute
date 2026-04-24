package service

import (
	"context"

	"gpsgo/services/m17-roadmap/internal/repository"
)

type Service struct {
	repo *repository.Repository
}

func New(repo *repository.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListFeatures(ctx context.Context) ([]repository.Feature, error) {
	return s.repo.ListFeatures(ctx)
}

func (s *Service) CreateFeature(ctx context.Context, f repository.Feature) (repository.Feature, error) {
	if f.Status == "" {
		f.Status = "PLANNED"
	}
	return s.repo.CreateFeature(ctx, f)
}

func (s *Service) CastVote(ctx context.Context, featureID, tenantID, userID string) error {
	return s.repo.CastVote(ctx, featureID, tenantID, userID)
}
