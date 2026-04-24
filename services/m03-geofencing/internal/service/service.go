package service

import (
	"context"

	"gpsgo/services/m03-geofencing/internal/geom"
	"gpsgo/services/m03-geofencing/internal/repository"
	"gpsgo/services/m03-geofencing/internal/kafka"
	sharedtypes "gpsgo/shared/types"
	shared "gpsgo/shared/kafka"
	"encoding/json"
	"time"
)

type Service struct {
	repo     *repository.Repository
	producer *kafka.Producer
}

func New(repo *repository.Repository, producer *kafka.Producer) *Service {
	return &Service{repo: repo, producer: producer}
}

func (s *Service) EvaluateLocation(ctx context.Context, evt sharedtypes.LocationUpdatedEvent) {
	geofences, err := s.repo.ListGeofencesByTenant(ctx, evt.TenantID)
	if err != nil {
		return
	}

	pt := geom.Point{Lat: evt.Lat, Lng: evt.Lng}

	for _, g := range geofences {
		isInside := false
		
		if g.Type == "POLYGON" {
			var poly geom.Polygon
			b, _ := json.Marshal(g.Polygon)
			_ = json.Unmarshal(b, &poly)
			isInside = poly.Contains(pt)
		} else if g.Type == "CIRCLE" {
			// simplified
		}

		currentState, err := s.repo.GetVehicleState(ctx, evt.VehicleID, g.ID)
		if err != nil { currentState = "OUTSIDE" } // assume outside if not found

		if isInside && currentState != "INSIDE" {
			// ENTRY breach
			s.repo.UpdateVehicleState(ctx, evt.VehicleID, g.ID, evt.TenantID, "INSIDE")
			s.emitBreach(ctx, evt, g, "ENTRY")
		} else if !isInside && currentState == "INSIDE" {
			// EXIT breach
			s.repo.UpdateVehicleState(ctx, evt.VehicleID, g.ID, evt.TenantID, "OUTSIDE")
			s.emitBreach(ctx, evt, g, "EXIT")
		}
	}
}

func (s *Service) emitBreach(ctx context.Context, loc sharedtypes.LocationUpdatedEvent, g repository.Geofence, evtType string) {
	_ = s.producer.Publish(ctx, shared.TopicGeofenceBreach, loc.VehicleID, sharedtypes.GeofenceBreachEvent{
		ZoneID:    g.ID,
		ZoneName:  g.Name,
		VehicleID: loc.VehicleID,
		TenantID:  loc.TenantID,
		EventType: evtType,
		Lat:       loc.Lat,
		Lng:       loc.Lng,
		Speed:     loc.Speed,
		Timestamp: time.Now(),
	})
}

func (s *Service) CreateGeofence(ctx context.Context, g repository.Geofence) (repository.Geofence, error) {
	return s.repo.CreateGeofence(ctx, g)
}

func (s *Service) ListGeofences(ctx context.Context, tenantID string) ([]repository.Geofence, error) {
	return s.repo.ListGeofencesByTenant(ctx, tenantID)
}
