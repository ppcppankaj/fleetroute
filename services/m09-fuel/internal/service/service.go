package service

import (
	"context"

	"gpsgo/services/m09-fuel/internal/repository"
	"gpsgo/services/m09-fuel/internal/kafka"
	sharedtypes "gpsgo/shared/types"
	shared "gpsgo/shared/kafka"
)

type Service struct {
	repo     *repository.Repository
	producer *kafka.Producer
}

func New(repo *repository.Repository, producer *kafka.Producer) *Service {
	return &Service{repo: repo, producer: producer}
}

func (s *Service) CreateFuelLog(ctx context.Context, f repository.FuelLog) (repository.FuelLog, error) {
	created, err := s.repo.CreateFuelLog(ctx, f)
	if err != nil {
		return repository.FuelLog{}, err
	}
	_ = s.producer.Publish(ctx, shared.TopicFuelLogged, created.ID, sharedtypes.FuelLoggedEvent{
		LogID:     created.ID,
		VehicleID: created.VehicleID,
		TenantID:  created.TenantID,
		Liters:    created.Liters,
		TotalCost: created.TotalCost,
		LoggedAt:  created.LoggedAt,
	})
	return created, nil
}

func (s *Service) ListFuelLogs(ctx context.Context, tenantID string) ([]repository.FuelLog, error) {
	return s.repo.ListFuelLogs(ctx, tenantID, 100)
}

// ProcessTripCompleted auto-logs fuel used from trip event
func (s *Service) ProcessTripCompleted(ctx context.Context, evt sharedtypes.TripCompletedEvent) {
	if evt.FuelUsed <= 0 {
		return
	}
	f := repository.FuelLog{
		TenantID:  evt.TenantID,
		VehicleID: evt.VehicleID,
		TripID:    &evt.TripID,
		Liters:    evt.FuelUsed,
		TotalCost: 0, // auto-logged from trip, no cost data
	}
	if evt.DriverID != "" {
		f.DriverID = &evt.DriverID
	}
	_, _ = s.repo.CreateFuelLog(ctx, f)
}
