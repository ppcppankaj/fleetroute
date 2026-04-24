package service

import (
	"context"

	"gpsgo/services/m10-multi-tenant/internal/repository"
	sharedtypes "gpsgo/shared/types"
	shared "gpsgo/shared/kafka"
	"encoding/json"
	"strings"
	"github.com/segmentio/kafka-go"
)

type Producer struct {
	writer *kafka.Writer
}

func NewProducer(brokers string) *Producer {
	return &Producer{
		writer: &kafka.Writer{
			Addr:     kafka.TCP(strings.Split(brokers, ",")...),
			Balancer: &kafka.LeastBytes{},
		},
	}
}

func (p *Producer) Publish(ctx context.Context, topic, key string, payload any) error {
	b, _ := json.Marshal(payload)
	return p.writer.WriteMessages(ctx, kafka.Message{Topic: topic, Key: []byte(key), Value: b})
}

func (p *Producer) Close() error { return p.writer.Close() }

type Service struct {
	repo     *repository.Repository
	producer *Producer
}

func New(repo *repository.Repository, brokers string) *Service {
	return &Service{repo: repo, producer: NewProducer(brokers)}
}

func (s *Service) Close() error { return s.producer.Close() }

func (s *Service) CreateTenant(ctx context.Context, t repository.Tenant) (repository.Tenant, error) {
	if t.Timezone == "" {
		t.Timezone = "UTC"
	}
	if t.MaxVehicles == 0 {
		t.MaxVehicles = 10
	}
	if t.MaxUsers == 0 {
		t.MaxUsers = 5
	}
	created, err := s.repo.CreateTenant(ctx, t)
	if err != nil {
		return repository.Tenant{}, err
	}
	_ = s.producer.Publish(ctx, shared.TopicTenantCreated, created.ID, sharedtypes.TenantCreatedEvent{
		TenantID:  created.ID,
		Name:      created.Name,
		Slug:      created.Slug,
		CreatedAt: created.CreatedAt,
	})
	return created, nil
}

func (s *Service) GetTenant(ctx context.Context, id string) (repository.Tenant, error) {
	return s.repo.GetTenant(ctx, id)
}

func (s *Service) ListTenants(ctx context.Context) ([]repository.Tenant, error) {
	return s.repo.ListTenants(ctx)
}

func (s *Service) SuspendTenant(ctx context.Context, id string) error {
	return s.repo.UpdateTenantStatus(ctx, id, "SUSPENDED")
}

func (s *Service) ActivateTenant(ctx context.Context, id string) error {
	return s.repo.UpdateTenantStatus(ctx, id, "ACTIVE")
}
