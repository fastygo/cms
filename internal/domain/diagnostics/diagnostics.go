package diagnostics

import "time"

type ErrorRecord struct {
	ID         string         `json:"id"`
	Source     string         `json:"source"`
	Message    string         `json:"message"`
	Severity   string         `json:"severity"`
	OccurredAt time.Time      `json:"occurred_at"`
	Details    map[string]any `json:"details,omitempty"`
}
