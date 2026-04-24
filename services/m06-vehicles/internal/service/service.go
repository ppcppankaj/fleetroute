package service

import (
	"context"
	"errors"

	"gpsgo/services/m06-vehicles/internal/repository"
	"gpsgo/services/m06-vehicles/internal/kafka"
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

func (s *Service) CreateVehicle(ctx context.Context, v repository.Vehicle) (repository.Vehicle, error) {
	v.Status = "ACTIVE"
	created, err := s.repo.CreateVehicle(ctx, v)
	if err != nil {
		return repository.Vehicle{}, err
	}
	
	_ = s.producer.Publish(ctx, shared.TopicVehicleCreated, created.ID, sharedtypes.VehicleCreatedEvent{
		VehicleID:   created.ID,
		TenantID:    created.TenantID,
		PlateNumber: created.PlateNumber,
		Make:        created.Make,
		Model:       created.Model,
		Year:        created.Year,
		FuelType:    created.FuelType,
		CreatedAt:   created.CreatedAt,
	})

	return created, nil
}

func (s *Service) GetVehicle(ctx context.Context, id, tenantID string) (repository.Vehicle, error) {
	return s.repo.GetVehicle(ctx, id, tenantID)
}

func (s *Service) ListVehicles(ctx context.Context, tenantID string) ([]repository.Vehicle, error) {
	return s.repo.ListVehicles(ctx, tenantID)
}

func (s *Service) DeleteVehicle(ctx context.Context, id, tenantID string) error {
	return s.repo.DeleteVehicle(ctx, id, tenantID)
}

func (s *Service) UpdateVehicleOdometer(ctx context.Context, vehicleID string, distance float64) error {
	if distance <= 0 {
		return errors.New("invalid distance")
	}
	return s.repo.UpdateVehicleOdometer(ctx, vehicleID, distance)
}
