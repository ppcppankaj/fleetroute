package service

import (
	"context"

	"gpsgo/services/m16-activity-log/internal/repository"
	sharedtypes "gpsgo/shared/types"
)

type Service struct {
	repo *repository.Repository
}

func New(repo *repository.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) LogAlert(ctx context.Context, evt sharedtypes.AlertTriggeredEvent) {
	_ = s.repo.Insert(ctx, repository.ActivityEvent{
		TenantID:  evt.TenantID,
		VehicleID: &evt.VehicleID,
		DriverID:  &evt.DriverID,
		Type:      "ALERT_TRIGGERED",
		Title:     evt.Message,
	})
}

func (s *Service) LogTrip(ctx context.Context, evt sharedtypes.TripStartedEvent) {
	_ = s.repo.Insert(ctx, repository.ActivityEvent{
		TenantID:  evt.TenantID,
		VehicleID: &evt.VehicleID,
		DriverID:  &evt.DriverID,
		Type:      "TRIP_STARTED",
		Title:     "Trip started for vehicle " + evt.VehicleID,
	})
}

func (s *Service) LogTripCompleted(ctx context.Context, evt sharedtypes.TripCompletedEvent) {
	_ = s.repo.Insert(ctx, repository.ActivityEvent{
		TenantID:  evt.TenantID,
		VehicleID: &evt.VehicleID,
		DriverID:  &evt.DriverID,
		Type:      "TRIP_COMPLETED",
		Title:     "Trip completed",
	})
}

func (s *Service) ListEvents(ctx context.Context, tenantID string, limit int) ([]repository.ActivityEvent, error) {
	return s.repo.List(ctx, tenantID, limit)
}
