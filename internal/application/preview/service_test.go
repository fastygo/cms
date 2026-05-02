package preview_test

import (
	"context"
	"testing"
	"time"

	apppreview "github.com/fastygo/cms/internal/application/preview"
	"github.com/fastygo/cms/internal/domain/authz"
	domaincontent "github.com/fastygo/cms/internal/domain/content"
	domainpreview "github.com/fastygo/cms/internal/domain/preview"
)

func TestPreviewAccessModel(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 5, 2, 10, 0, 0, 0, time.UTC)
	repo := &memoryPreviewRepo{items: make(map[domainpreview.Token]domainpreview.Access)}
	service := apppreview.NewService(repo, func() time.Time { return now })
	editor := authz.NewPrincipal("editor-1", authz.CapabilityContentReadPrivate)

	access, err := service.Create(ctx, editor, domaincontent.ID("content-1"), time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	validated, ok, err := service.Validate(ctx, access.Token)
	if err != nil {
		t.Fatal(err)
	}
	if !ok || validated.EntryID != "content-1" {
		t.Fatalf("expected valid preview access, got ok=%v access=%+v", ok, validated)
	}
}

type memoryPreviewRepo struct {
	items map[domainpreview.Token]domainpreview.Access
}

func (r *memoryPreviewRepo) SavePreview(_ context.Context, access domainpreview.Access) error {
	r.items[access.Token] = access
	return nil
}

func (r *memoryPreviewRepo) GetPreview(_ context.Context, token domainpreview.Token) (domainpreview.Access, bool, error) {
	access, ok := r.items[token]
	return access, ok, nil
}
