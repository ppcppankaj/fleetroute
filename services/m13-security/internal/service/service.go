package service

import (
	"context"

	"gpsgo/services/m13-security/internal/repository"
	sharedtypes "gpsgo/shared/types"
)

type Service struct {
	repo *repository.Repository
}

func New(repo *repository.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) RecordUserLogin(ctx context.Context, evt sharedtypes.UserLoginEvent) {
	action := "LOGIN_SUCCESS"
	if !evt.Success {
		action = "LOGIN_FAILED"
	}
	tenantID := evt.TenantID
	userID := evt.UserID
	_ = s.repo.CreateAuditLog(ctx, repository.AuditLog{
		TenantID:  &tenantID,
		UserID:    &userID,
		Action:    action,
		Resource:  "auth",
		IPAddress: &evt.IPAddress,
		UserAgent: &evt.UserAgent,
	})
}

func (s *Service) RecordUserAction(ctx context.Context, evt sharedtypes.UserActionEvent) {
	tenantID := evt.TenantID
	userID := evt.UserID
	_ = s.repo.CreateAuditLog(ctx, repository.AuditLog{
		TenantID:   &tenantID,
		UserID:     &userID,
		Action:     evt.Action,
		Resource:   evt.Resource,
		ResourceID: &evt.ResourceID,
		IPAddress:  &evt.IPAddress,
	})
}

func (s *Service) ListAuditLogs(ctx context.Context, tenantID string) ([]repository.AuditLog, error) {
	return s.repo.ListAuditLogs(ctx, tenantID, 200)
}

func (s *Service) ListIncidents(ctx context.Context, tenantID string) ([]repository.SecurityIncident, error) {
	return s.repo.ListIncidents(ctx, tenantID)
}
