package service

import (
	"context"

	"gpsgo/services/m04-alerts/internal/repository"
	"gpsgo/services/m04-alerts/internal/kafka"
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

// CheckSpeeding is triggered by M01 Location updates
func (s *Service) CheckSpeeding(ctx context.Context, evt sharedtypes.LocationUpdatedEvent) {
	rules, err := s.repo.ListAlertRulesByEvent(ctx, "SPEEDING")
	if err != nil { return }
	
	for _, rule := range rules {
		if rule.TenantID != evt.TenantID { continue }
		
		// rule.Conditions JSON might contain max_speed. Simplified for POC:
		// If speed is over 100
		if evt.Speed > 100 {
			alert, _ := s.repo.CreateAlert(ctx, repository.ActiveAlert{
				TenantID:  evt.TenantID,
				RuleID:    &rule.ID,
				VehicleID: evt.VehicleID,
				Type:      "SPEEDING",
				Severity:  rule.Severity,
				Message:   "Vehicle exceeded speed limit",
			})
			s.emitTrigger(ctx, alert)
		}
	}
}

// CheckGeofence is triggered by M03
func (s *Service) CheckGeofence(ctx context.Context, evt sharedtypes.GeofenceBreachEvent) {
	rules, err := s.repo.ListAlertRulesByEvent(ctx, "GEOFENCE_BREACH")
	if err != nil { return }
	
	for _, rule := range rules {
		if rule.TenantID != evt.TenantID { continue }
		
		alert, _ := s.repo.CreateAlert(ctx, repository.ActiveAlert{
			TenantID:  evt.TenantID,
			RuleID:    &rule.ID,
			VehicleID: evt.VehicleID,
			DriverID:  &evt.DriverID,
			Type:      "GEOFENCE_BREACH",
			Severity:  rule.Severity,
			Message:   "Vehicle " + evt.EventType + " geofence " + evt.ZoneName,
		})
		s.emitTrigger(ctx, alert)
	}
}

func (s *Service) emitTrigger(ctx context.Context, a repository.ActiveAlert) {
	_ = s.producer.Publish(ctx, shared.TopicAlertTriggered, a.VehicleID, sharedtypes.AlertTriggeredEvent{
		AlertID:   a.ID,
		TenantID:  a.TenantID,
		VehicleID: a.VehicleID,
		Type:      a.Type,
		Severity:  a.Severity,
		Message:   a.Message,
		CreatedAt: a.CreatedAt,
	})
}

func (s *Service) ListActiveAlerts(ctx context.Context, tenantID string) ([]repository.ActiveAlert, error) {
	return s.repo.ListActiveAlerts(ctx, tenantID)
}

func (s *Service) ResolveAlert(ctx context.Context, id, tenantID, resolvedBy string) error {
	err := s.repo.ResolveAlert(ctx, id, tenantID, resolvedBy)
	if err == nil {
		_ = s.producer.Publish(ctx, shared.TopicAlertResolved, id, sharedtypes.AlertResolvedEvent{
			AlertID:  id,
			TenantID: tenantID,
		})
	}
	return err
}
