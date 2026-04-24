package service

import (
	"context"

	"gpsgo/services/m08-maintenance/internal/repository"
	"gpsgo/services/m08-maintenance/internal/kafka"
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

func (s *Service) CreateTask(ctx context.Context, t repository.MaintenanceTask) (repository.MaintenanceTask, error) {
	return s.repo.CreateTask(ctx, t)
}

func (s *Service) ListTasks(ctx context.Context, tenantID string) ([]repository.MaintenanceTask, error) {
	return s.repo.ListTasks(ctx, tenantID)
}

func (s *Service) CompleteTask(ctx context.Context, id, tenantID string, cost float64, vendor, notes string) error {
	return s.repo.CompleteTask(ctx, id, tenantID, cost, vendor, notes)
}

func (s *Service) CheckOverdueAndEmit(ctx context.Context) error {
	tasks, err := s.repo.ListOverdueTasks(ctx)
	if err != nil {
		return err
	}
	for _, t := range tasks {
		_ = s.producer.Publish(ctx, shared.TopicMaintenanceDue, t.ID, sharedtypes.MaintenanceDueEvent{
			TaskID:    t.ID,
			VehicleID: t.VehicleID,
			TenantID:  t.TenantID,
			Type:      t.Type,
			Title:     t.Title,
			DueAt:     t.DueAt,
		})
	}
	return nil
}
