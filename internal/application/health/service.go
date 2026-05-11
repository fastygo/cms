package health

import (
	"context"
	"time"
)

type Check struct {
	ID          string
	Label       string
	Description string
	Run         func(context.Context) error
}

type Result struct {
	ID          string
	Label       string
	Description string
	Status      string
	Error       string
	CheckedAt   time.Time
}

type Service struct {
	checks []Check
	now    func() time.Time
}

func NewService(now func() time.Time, checks ...Check) Service {
	if now == nil {
		now = time.Now
	}
	return Service{checks: append([]Check(nil), checks...), now: now}
}

func (s Service) Results(ctx context.Context) []Result {
	results := make([]Result, 0, len(s.checks))
	for _, check := range s.checks {
		result := Result{
			ID:          check.ID,
			Label:       check.Label,
			Description: check.Description,
			Status:      "ok",
			CheckedAt:   s.now().UTC(),
		}
		if check.Run != nil {
			if err := check.Run(ctx); err != nil {
				result.Status = "error"
				result.Error = err.Error()
			}
		}
		results = append(results, result)
	}
	return results
}
