package diagnostics

import (
	"context"
	"time"

	domaindiagnostics "github.com/fastygo/cms/internal/domain/diagnostics"
	"github.com/google/uuid"
)

type Repository interface {
	SaveErrorRecord(context.Context, domaindiagnostics.ErrorRecord) error
	ListErrorRecords(context.Context, int) ([]domaindiagnostics.ErrorRecord, error)
}

type Service struct {
	repo Repository
	now  func() time.Time
}

func NewService(repo Repository, now func() time.Time) Service {
	if now == nil {
		now = time.Now
	}
	return Service{repo: repo, now: now}
}

func (s Service) Enabled() bool {
	return s.repo != nil && s.now != nil
}

func (s Service) Record(ctx context.Context, source string, message string, severity string, details map[string]any) error {
	if !s.Enabled() {
		return nil
	}
	return s.repo.SaveErrorRecord(ctx, domaindiagnostics.ErrorRecord{
		ID:         uuid.NewString(),
		Source:     source,
		Message:    message,
		Severity:   severity,
		OccurredAt: s.now().UTC(),
		Details:    details,
	})
}

func (s Service) Recent(ctx context.Context, limit int) ([]domaindiagnostics.ErrorRecord, error) {
	if !s.Enabled() {
		return nil, nil
	}
	if limit <= 0 {
		limit = 50
	}
	return s.repo.ListErrorRecords(ctx, limit)
}
