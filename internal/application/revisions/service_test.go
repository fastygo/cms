package revisions_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	apprevisions "github.com/fastygo/cms/internal/application/revisions"
	"github.com/fastygo/cms/internal/domain/authz"
	domaincontent "github.com/fastygo/cms/internal/domain/content"
	domainrevisions "github.com/fastygo/cms/internal/domain/revisions"
)

func TestRevisionCreateAndRestore(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 5, 2, 10, 0, 0, 0, time.UTC)
	entryRepo := &memoryEntryRepo{entries: map[domaincontent.ID]domaincontent.Entry{
		"content-1": {ID: "content-1", Title: domaincontent.LocalizedText{"en": "Original"}, AuthorID: "author-1"},
	}}
	revisionRepo := &memoryRevisionRepo{revisions: make(map[domainrevisions.ID]domainrevisions.Revision)}
	service := apprevisions.NewService(revisionRepo, entryRepo, func() time.Time { return now })
	editor := authz.NewPrincipal("author-1", authz.CapabilityContentEdit, authz.CapabilityContentRestore)

	revision, err := service.Create(ctx, editor, "content-1", "before update")
	if err != nil {
		t.Fatal(err)
	}
	entry := entryRepo.entries["content-1"]
	entry.Title = domaincontent.LocalizedText{"en": "Changed"}
	entryRepo.entries["content-1"] = entry

	restored, err := service.Restore(ctx, editor, revision.ID)
	if err != nil {
		t.Fatal(err)
	}
	if restored.Title["en"] != "Original" {
		t.Fatalf("expected restored title, got %+v", restored.Title)
	}
}

type memoryRevisionRepo struct {
	next      int
	revisions map[domainrevisions.ID]domainrevisions.Revision
}

func (r *memoryRevisionRepo) NextRevisionID(context.Context) (domainrevisions.ID, error) {
	r.next++
	return domainrevisions.ID(fmt.Sprintf("revision-%d", r.next)), nil
}

func (r *memoryRevisionRepo) SaveRevision(_ context.Context, revision domainrevisions.Revision) error {
	r.revisions[revision.ID] = revision
	return nil
}

func (r *memoryRevisionRepo) GetRevision(_ context.Context, id domainrevisions.ID) (domainrevisions.Revision, error) {
	revision, ok := r.revisions[id]
	if !ok {
		return domainrevisions.Revision{}, fmt.Errorf("revision %q not found", id)
	}
	return revision, nil
}

type memoryEntryRepo struct {
	entries map[domaincontent.ID]domaincontent.Entry
}

func (r *memoryEntryRepo) Get(_ context.Context, id domaincontent.ID) (domaincontent.Entry, error) {
	entry, ok := r.entries[id]
	if !ok {
		return domaincontent.Entry{}, fmt.Errorf("content %q not found", id)
	}
	return entry, nil
}

func (r *memoryEntryRepo) Save(_ context.Context, entry domaincontent.Entry) error {
	r.entries[entry.ID] = entry
	return nil
}
