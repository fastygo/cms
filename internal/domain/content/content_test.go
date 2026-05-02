package content_test

import (
	"testing"
	"time"

	"github.com/fastygo/cms/internal/domain/content"
)

func TestValidateStatusAndKind(t *testing.T) {
	if err := content.ValidateKind(content.KindPost); err != nil {
		t.Fatal(err)
	}
	if err := content.ValidateKind(""); err == nil {
		t.Fatal("expected empty kind to be invalid")
	}
	if err := content.ValidateStatus(content.StatusPublished); err != nil {
		t.Fatal(err)
	}
	if err := content.ValidateStatus("unknown"); err == nil {
		t.Fatal("expected unknown status to be invalid")
	}
}

func TestPublicVisibilityRules(t *testing.T) {
	now := time.Date(2026, 5, 2, 10, 0, 0, 0, time.UTC)
	future := now.Add(time.Hour)
	entry := content.Entry{Status: content.StatusPublished, Visibility: content.VisibilityPublic, PublishedAt: &future}
	if entry.IsPublicAt(now) {
		t.Fatal("future published entry should not be public yet")
	}
	entry.PublishedAt = &now
	if !entry.IsPublicAt(now) {
		t.Fatal("published entry should be public")
	}
	entry.Visibility = content.VisibilityPrivate
	if entry.IsPublicAt(now) {
		t.Fatal("private entry should not be public")
	}
}
