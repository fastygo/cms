package content_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	appcontent "github.com/fastygo/cms/internal/application/content"
	appcontenttype "github.com/fastygo/cms/internal/application/contenttype"
	"github.com/fastygo/cms/internal/domain/authz"
	domaincontent "github.com/fastygo/cms/internal/domain/content"
	domaincontenttype "github.com/fastygo/cms/internal/domain/contenttype"
)

func TestContentWorkflows(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 5, 2, 10, 0, 0, 0, time.UTC)
	contentRepo := newMemoryContentRepo()
	typeRepo := newMemoryTypeRepo()
	typeService := appcontenttype.NewService(typeRepo)
	if err := typeService.InstallBuiltIns(ctx); err != nil {
		t.Fatal(err)
	}
	customType := domaincontenttype.Type{
		ID:          domaincontent.Kind("product"),
		Label:       "Products",
		Public:      true,
		RESTVisible: true,
		Supports:    domaincontenttype.Supports{Title: true, Editor: true},
	}
	if err := typeService.Register(ctx, customType); err != nil {
		t.Fatal(err)
	}
	service := appcontent.NewService(contentRepo, typeRepo, func() time.Time { return now })
	editor := authz.NewPrincipal("author-1",
		authz.CapabilityContentCreate,
		authz.CapabilityContentPublish,
		authz.CapabilityContentSchedule,
		authz.CapabilityContentDelete,
		authz.CapabilityContentRestore,
		authz.CapabilityContentEditOwn,
	)

	post, err := service.CreateDraft(ctx, editor, appcontent.CreateDraftCommand{
		Kind:     domaincontent.KindPost,
		Title:    domaincontent.LocalizedText{"en": "Hello"},
		Slug:     domaincontent.LocalizedText{"en": "hello"},
		AuthorID: "author-1",
		Metadata: domaincontent.Metadata{
			"public":  {Value: "shown", Public: true},
			"private": {Value: "hidden", Public: false},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if post.Status != domaincontent.StatusDraft || post.Kind != domaincontent.KindPost {
		t.Fatalf("unexpected draft: %+v", post)
	}
	if private := post.Metadata.Public()["private"]; private.Value != nil {
		t.Fatalf("private metadata leaked: %+v", private)
	}

	page, err := service.CreateDraft(ctx, editor, appcontent.CreateDraftCommand{
		Kind:  domaincontent.KindPage,
		Title: domaincontent.LocalizedText{"en": "About"},
		Slug:  domaincontent.LocalizedText{"en": "about"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if page.Kind != domaincontent.KindPage {
		t.Fatalf("expected page kind, got %q", page.Kind)
	}

	custom, err := service.CreateDraft(ctx, editor, appcontent.CreateDraftCommand{
		Kind:  domaincontent.Kind("product"),
		Title: domaincontent.LocalizedText{"en": "Camera"},
		Slug:  domaincontent.LocalizedText{"en": "camera"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if custom.Kind != domaincontent.Kind("product") {
		t.Fatalf("expected custom kind, got %q", custom.Kind)
	}

	published, err := service.Publish(ctx, editor, post.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !published.IsPublicAt(now) {
		t.Fatal("published content should be public at current time")
	}

	future := now.Add(24 * time.Hour)
	scheduled, err := service.Schedule(ctx, editor, page.ID, future)
	if err != nil {
		t.Fatal(err)
	}
	if scheduled.IsPublicAt(now) {
		t.Fatal("scheduled content should not be public before publish time")
	}

	public, err := service.List(ctx, domaincontent.Query{PublicOnly: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(public.Items) != 1 || public.Items[0].ID != post.ID {
		t.Fatalf("expected only published post in public list, got %+v", public)
	}

	trashed, err := service.Trash(ctx, editor, post.ID)
	if err != nil {
		t.Fatal(err)
	}
	if trashed.Status != domaincontent.StatusTrashed || trashed.DeletedAt == nil {
		t.Fatalf("expected trashed entry, got %+v", trashed)
	}
	restored, err := service.Restore(ctx, editor, post.ID)
	if err != nil {
		t.Fatal(err)
	}
	if restored.Status != domaincontent.StatusDraft || restored.DeletedAt != nil {
		t.Fatalf("expected restored draft, got %+v", restored)
	}
}

func TestContentRequiresCapabilitiesAndRegisteredTypes(t *testing.T) {
	ctx := context.Background()
	contentRepo := newMemoryContentRepo()
	typeRepo := newMemoryTypeRepo()
	service := appcontent.NewService(contentRepo, typeRepo, time.Now)

	_, err := service.CreateDraft(ctx, authz.NewPrincipal("author-1"), appcontent.CreateDraftCommand{Kind: domaincontent.KindPost})
	if err == nil {
		t.Fatal("expected create draft to require capability")
	}

	creator := authz.NewPrincipal("author-1", authz.CapabilityContentCreate)
	_, err = service.CreateDraft(ctx, creator, appcontent.CreateDraftCommand{Kind: domaincontent.KindPost})
	if err == nil {
		t.Fatal("expected unregistered content type to be rejected")
	}
}

type memoryContentRepo struct {
	next    int
	entries map[domaincontent.ID]domaincontent.Entry
}

func newMemoryContentRepo() *memoryContentRepo {
	return &memoryContentRepo{entries: make(map[domaincontent.ID]domaincontent.Entry)}
}

func (r *memoryContentRepo) NextID(context.Context) (domaincontent.ID, error) {
	r.next++
	return domaincontent.ID(fmt.Sprintf("content-%d", r.next)), nil
}

func (r *memoryContentRepo) Save(_ context.Context, entry domaincontent.Entry) error {
	r.entries[entry.ID] = entry
	return nil
}

func (r *memoryContentRepo) Get(_ context.Context, id domaincontent.ID) (domaincontent.Entry, error) {
	entry, ok := r.entries[id]
	if !ok {
		return domaincontent.Entry{}, fmt.Errorf("content %q not found", id)
	}
	return entry, nil
}

func (r *memoryContentRepo) List(_ context.Context, query domaincontent.Query) (domaincontent.ListResult, error) {
	var entries []domaincontent.Entry
	for _, entry := range r.entries {
		if query.PublicOnly && !entry.IsPublicAt(query.PublishedAt) {
			continue
		}
		entries = append(entries, entry)
	}
	return domaincontent.ListResult{Items: entries, Total: len(entries), Page: 1, PerPage: len(entries), TotalPages: 1}, nil
}

type memoryTypeRepo struct {
	types map[domaincontent.Kind]domaincontenttype.Type
}

func newMemoryTypeRepo() *memoryTypeRepo {
	return &memoryTypeRepo{types: make(map[domaincontent.Kind]domaincontenttype.Type)}
}

func (r *memoryTypeRepo) SaveContentType(_ context.Context, contentType domaincontenttype.Type) error {
	r.types[contentType.ID] = contentType
	return nil
}

func (r *memoryTypeRepo) GetContentType(_ context.Context, kind domaincontent.Kind) (domaincontenttype.Type, bool, error) {
	contentType, ok := r.types[kind]
	return contentType, ok, nil
}

func (r *memoryTypeRepo) ListContentTypes(context.Context) ([]domaincontenttype.Type, error) {
	types := make([]domaincontenttype.Type, 0, len(r.types))
	for _, contentType := range r.types {
		types = append(types, contentType)
	}
	return types, nil
}
