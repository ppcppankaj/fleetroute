package service

import (
	"context"
	"fmt"
	"time"

	"gpsgo/services/m05-reports/internal/repository"
)

type Service struct {
	repo *repository.Repository
}

func New(repo *repository.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateDefinition(ctx context.Context, d repository.ReportDefinition) (repository.ReportDefinition, error) {
	return s.repo.CreateDefinition(ctx, d)
}

func (s *Service) ListDefinitions(ctx context.Context, tenantID string) ([]repository.ReportDefinition, error) {
	return s.repo.ListDefinitions(ctx, tenantID)
}

func (s *Service) ListRuns(ctx context.Context, tenantID string) ([]repository.ReportRun, error) {
	return s.repo.ListRuns(ctx, tenantID)
}

// RunReport creates a run record and asynchronously generates the report.
// In production this would call a MinIO-backed CSV/PDF generator.
func (s *Service) RunReport(ctx context.Context, defID, tenantID string) (repository.ReportRun, error) {
	run, err := s.repo.CreateRun(ctx, defID, tenantID)
	if err != nil {
		return repository.ReportRun{}, err
	}

	// Async generation (simplified — just marks done with a placeholder URL)
	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		// TODO: Build actual CSV/PDF and upload to MinIO
		// For now, mark with a dummy S3 URL
		fileURL := fmt.Sprintf("s3://reports/%s/%s.csv", tenantID, run.ID)
		_ = s.repo.CompleteRun(bgCtx, run.ID, fileURL)
	}()

	return run, nil
}
