package audit

import (
	"context"
	"strings"
	"time"

	domainaudit "github.com/fastygo/cms/internal/domain/audit"
	"github.com/google/uuid"
)

type Repository interface {
	SaveAuditEvent(context.Context, domainaudit.Event) error
	ListAuditEvents(context.Context, int) ([]domainaudit.Event, error)
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

func (s Service) Record(ctx context.Context, event domainaudit.Event) error {
	if !s.Enabled() {
		return nil
	}
	if strings.TrimSpace(event.ID) == "" {
		event.ID = uuid.NewString()
	}
	if event.OccurredAt.IsZero() {
		event.OccurredAt = s.now().UTC()
	}
	event.Details = redactDetails(event.Details)
	return s.repo.SaveAuditEvent(ctx, event)
}

func (s Service) Recent(ctx context.Context, limit int) ([]domainaudit.Event, error) {
	if !s.Enabled() {
		return nil, nil
	}
	if limit <= 0 {
		limit = 50
	}
	return s.repo.ListAuditEvents(ctx, limit)
}

func redactDetails(details map[string]any) map[string]any {
	if len(details) == 0 {
		return nil
	}
	result := map[string]any{}
	for key, value := range details {
		lower := strings.ToLower(strings.TrimSpace(key))
		switch {
		case strings.Contains(lower, "password"),
			strings.Contains(lower, "token"),
			strings.Contains(lower, "secret"),
			strings.Contains(lower, "code"),
			strings.Contains(lower, "hash"):
			result[key] = "[redacted]"
		default:
			result[key] = value
		}
	}
	return result
}
