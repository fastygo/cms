package audit

import "time"

type Status string

const (
	StatusSuccess Status = "success"
	StatusFailure Status = "failure"
)

type Event struct {
	ID         string         `json:"id"`
	ActorID    string         `json:"actor_id,omitempty"`
	Action     string         `json:"action"`
	Resource   string         `json:"resource"`
	ResourceID string         `json:"resource_id,omitempty"`
	Status     Status         `json:"status"`
	OccurredAt time.Time      `json:"occurred_at"`
	RemoteAddr string         `json:"remote_addr,omitempty"`
	UserAgent  string         `json:"user_agent,omitempty"`
	Details    map[string]any `json:"details,omitempty"`
}
