package service

import (
	"context"

	"gpsgo/services/m12-devices/internal/repository"
	"gpsgo/services/m12-devices/internal/kafka"
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

func (s *Service) CreateDevice(ctx context.Context, d repository.Device) (repository.Device, error) {
	d.Status = "ACTIVE"
	created, err := s.repo.CreateDevice(ctx, d)
	if err != nil {
		return repository.Device{}, err
	}
	
	_ = s.producer.Publish(ctx, shared.TopicDeviceProvisioned, created.ID, sharedtypes.DeviceProvisionedEvent{
		DeviceID:  created.ID,
		IMEI:      created.IMEI,
		TenantID:  *created.TenantID,
		Model:     created.Model,
		CreatedAt: created.CreatedAt,
	})

	return created, nil
}

func (s *Service) GetDevice(ctx context.Context, id, tenantID string) (repository.Device, error) {
	return s.repo.GetDevice(ctx, id, tenantID)
}

func (s *Service) ListDevices(ctx context.Context, tenantID string) ([]repository.Device, error) {
	return s.repo.ListDevices(ctx, tenantID)
}

func (s *Service) DeleteDevice(ctx context.Context, id, tenantID string) error {
	return s.repo.DeleteDevice(ctx, id, tenantID)
}

func (s *Service) UpdateDeviceStatusFromOfflineEvent(ctx context.Context, deviceID string) error {
	return s.repo.UpdateDeviceStatus(ctx, deviceID, "OFFLINE")
}
