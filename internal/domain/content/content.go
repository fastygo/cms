package content

import (
	"fmt"
	"strings"
	"time"
)

type ID string
type Kind string
type Status string
type Visibility string

const (
	KindPost Kind = "post"
	KindPage Kind = "page"
)

const (
	StatusDraft     Status = "draft"
	StatusScheduled Status = "scheduled"
	StatusPublished Status = "published"
	StatusArchived  Status = "archived"
	StatusTrashed   Status = "trashed"
)

const (
	VisibilityPublic  Visibility = "public"
	VisibilityPrivate Visibility = "private"
)

// LocalizedText stores values by locale.
type LocalizedText map[string]string

// Value returns a locale value with fallback support.
func (t LocalizedText) Value(locale string, fallback string) string {
	if value := strings.TrimSpace(t[locale]); value != "" {
		return value
	}
	return strings.TrimSpace(t[fallback])
}

// MetaValue stores extension metadata with public visibility.
type MetaValue struct {
	Value  any
	Public bool
}

type Metadata map[string]MetaValue

// Public returns metadata safe for public output.
func (m Metadata) Public() Metadata {
	out := make(Metadata)
	for key, value := range m {
		if value.Public {
			out[key] = value
		}
	}
	return out
}

type TermRef struct {
	Taxonomy string
	TermID   string
}

type Entry struct {
	ID              ID
	Kind            Kind
	Status          Status
	Visibility      Visibility
	Title           LocalizedText
	Slug            LocalizedText
	Body            LocalizedText
	Excerpt         LocalizedText
	AuthorID        string
	FeaturedMediaID string
	Template        string
	Metadata        Metadata
	Terms           []TermRef
	CreatedAt       time.Time
	UpdatedAt       time.Time
	PublishedAt     *time.Time
	DeletedAt       *time.Time
}

type Query struct {
	Kinds       []Kind
	Statuses    []Status
	AuthorID    string
	Slug        string
	Taxonomy    string
	TermID      string
	Search      string
	Locale      string
	After       *time.Time
	Before      *time.Time
	PublicOnly  bool
	PublishedAt time.Time
	Page        int
	PerPage     int
	SortBy      SortField
	SortDesc    bool
}

type SortField string

const (
	SortCreatedAt   SortField = "created_at"
	SortUpdatedAt   SortField = "updated_at"
	SortPublishedAt SortField = "published_at"
	SortTitle       SortField = "title"
	SortSlug        SortField = "slug"
)

type ListResult struct {
	Items      []Entry
	Total      int
	Page       int
	PerPage    int
	TotalPages int
}

func NormalizeKind(value string) Kind {
	return Kind(strings.ToLower(strings.TrimSpace(value)))
}

func ValidateKind(kind Kind) error {
	if strings.TrimSpace(string(kind)) == "" {
		return fmt.Errorf("content kind is required")
	}
	return nil
}

func ValidateStatus(status Status) error {
	switch status {
	case StatusDraft, StatusScheduled, StatusPublished, StatusArchived, StatusTrashed:
		return nil
	default:
		return fmt.Errorf("unsupported content status %q", status)
	}
}

func (e Entry) IsPublicAt(now time.Time) bool {
	if e.Visibility == VisibilityPrivate {
		return false
	}
	if e.DeletedAt != nil {
		return false
	}
	switch e.Status {
	case StatusPublished:
		return e.PublishedAt == nil || !e.PublishedAt.After(now)
	default:
		return false
	}
}
