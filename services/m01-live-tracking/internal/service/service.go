package service

import (
	"context"

	"gpsgo/services/m01-live-tracking/internal/repository"
	"gpsgo/services/m01-live-tracking/internal/kafka"
	"gpsgo/services/m01-live-tracking/internal/websocket"
	sharedtypes "gpsgo/shared/types"
	shared "gpsgo/shared/kafka"
)

type Service struct {
	repo     *repository.Repository
	producer *kafka.Producer
	hub      *websocket.Hub
}

func New(repo *repository.Repository, producer *kafka.Producer, hub *websocket.Hub) *Service {
	return &Service{repo: repo, producer: producer, hub: hub}
}

func (s *Service) ProcessLocation(ctx context.Context, evt sharedtypes.LocationUpdatedEvent) {
	// 1. Save to DB
	_ = s.repo.InsertBreadcrumb(ctx, evt)
	
	// 2. Broadcast via WS
	s.hub.Broadcast(evt)
	
	// 3. Publish to Kafka for downstream services (M03 Geofencing, M04 Alerts, etc)
	_ = s.producer.Publish(ctx, shared.TopicLocationUpdated, evt.VehicleID, evt)
}

func (s *Service) GetBreadcrumbs(ctx context.Context, vehicleID, tenantID string, limit int) ([]repository.Breadcrumb, error) {
	return s.repo.GetRecentBreadcrumbs(ctx, vehicleID, tenantID, limit)
}
