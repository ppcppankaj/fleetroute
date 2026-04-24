package service

import (
	"context"
	"errors"

	"gpsgo/services/m02-routes-trips/internal/repository"
	"gpsgo/services/m02-routes-trips/internal/kafka"
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

// HandleLocationUpdate uses ignition state to auto-start/stop trips
func (s *Service) HandleLocationUpdate(ctx context.Context, evt sharedtypes.LocationUpdatedEvent) {
	activeTrip, err := s.repo.GetActiveTrip(ctx, evt.VehicleID, evt.TenantID)
	
	if evt.Ignition {
		if err != nil {
			// No active trip → start one
			trip := repository.Trip{
				TenantID:  evt.TenantID,
				VehicleID: evt.VehicleID,
				StartLat:  &evt.Lat,
				StartLng:  &evt.Lng,
			}
			started, err := s.repo.StartTrip(ctx, trip)
			if err != nil { return }
			_ = s.producer.Publish(ctx, shared.TopicTripStarted, started.ID, sharedtypes.TripStartedEvent{
				TripID:    started.ID,
				VehicleID: started.VehicleID,
				TenantID:  started.TenantID,
				StartLat:  evt.Lat,
				StartLng:  evt.Lng,
				StartTime: started.StartedAt,
			})
		}
	} else {
		if err == nil && activeTrip.ID != "" {
			// Ignition off → end trip
			duration := int(evt.Timestamp.Sub(activeTrip.StartedAt).Seconds())
			_ = s.repo.EndTrip(ctx, activeTrip.ID, evt.Lat, evt.Lng, activeTrip.DistanceKM, activeTrip.FuelUsed, duration)
			_ = s.producer.Publish(ctx, shared.TopicTripCompleted, activeTrip.ID, sharedtypes.TripCompletedEvent{
				TripID:     activeTrip.ID,
				VehicleID:  activeTrip.VehicleID,
				TenantID:   activeTrip.TenantID,
				DistanceKM: activeTrip.DistanceKM,
				StartTime:  activeTrip.StartedAt,
				EndTime:    evt.Timestamp,
				FuelUsed:   activeTrip.FuelUsed,
				EndLat:     evt.Lat,
				EndLng:     evt.Lng,
			})
		}
	}
}

func (s *Service) ListTrips(ctx context.Context, tenantID string) ([]repository.Trip, error) {
	return s.repo.ListTrips(ctx, tenantID, 100)
}

func (s *Service) CreateRoute(ctx context.Context, ro repository.Route) (repository.Route, error) {
	if ro.Name == "" {
		return repository.Route{}, errors.New("route name required")
	}
	return s.repo.CreateRoute(ctx, ro)
}

func (s *Service) ListRoutes(ctx context.Context, tenantID string) ([]repository.Route, error) {
	return s.repo.ListRoutes(ctx, tenantID)
}
