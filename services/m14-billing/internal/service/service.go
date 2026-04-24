package service

import (
	"context"

	"gpsgo/services/m14-billing/internal/repository"
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

func (s *Service) GetSubscription(ctx context.Context, tenantID string) (repository.Subscription, error) {
	return s.repo.GetSubscription(ctx, tenantID)
}

func (s *Service) UpsertSubscription(ctx context.Context, sub repository.Subscription) error {
	err := s.repo.UpsertSubscription(ctx, sub)
	if err != nil {
		return err
	}
	_ = s.producer.Publish(ctx, shared.TopicSubscriptionUpdate, sub.TenantID, sharedtypes.SubscriptionUpdatedEvent{
		TenantID: sub.TenantID,
		PlanID:   sub.PlanID,
		Status:   sub.Status,
	})
	return nil
}

func (s *Service) CreateInvoice(ctx context.Context, inv repository.Invoice) (repository.Invoice, error) {
	created, err := s.repo.CreateInvoice(ctx, inv)
	if err != nil {
		return repository.Invoice{}, err
	}
	_ = s.producer.Publish(ctx, shared.TopicInvoiceCreated, created.ID, sharedtypes.InvoiceCreatedEvent{
		InvoiceID: created.ID,
		TenantID:  created.TenantID,
		Amount:    created.Amount,
		Currency:  created.Currency,
		CreatedAt: created.CreatedAt,
	})
	return created, nil
}

func (s *Service) ListInvoices(ctx context.Context, tenantID string) ([]repository.Invoice, error) {
	return s.repo.ListInvoices(ctx, tenantID)
}
