package service

import (
	"context"

	"gpsgo/services/m07-drivers/internal/repository"
	"gpsgo/services/m07-drivers/internal/kafka"
)

type Service struct {
	repo     *repository.Repository
	producer *kafka.Producer
}

func New(repo *repository.Repository, producer *kafka.Producer) *Service {
	return &Service{repo: repo, producer: producer}
}

func (s *Service) CreateDriver(ctx context.Context, d repository.Driver) (repository.Driver, error) {
	return s.repo.CreateDriver(ctx, d)
}

func (s *Service) GetDriver(ctx context.Context, id, tenantID string) (repository.Driver, error) {
	return s.repo.GetDriver(ctx, id, tenantID)
}

func (s *Service) ListDrivers(ctx context.Context, tenantID string) ([]repository.Driver, error) {
	return s.repo.ListDrivers(ctx, tenantID)
}

func (s *Service) DeleteDriver(ctx context.Context, id, tenantID string) error {
	return s.repo.DeleteDriver(ctx, id, tenantID)
}

func (s *Service) UpdateDriverBehaviorScore(ctx context.Context, driverID string, points float64) error {
	return s.repo.UpdateDriverBehaviorScore(ctx, driverID, points)
}
