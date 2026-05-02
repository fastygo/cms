package media_test

import (
	"context"
	"fmt"
	"testing"

	appmedia "github.com/fastygo/cms/internal/application/media"
	"github.com/fastygo/cms/internal/domain/authz"
	domaincontent "github.com/fastygo/cms/internal/domain/content"
	domainmedia "github.com/fastygo/cms/internal/domain/media"
)

func TestMediaMetadataAndFeaturedAttachment(t *testing.T) {
	ctx := context.Background()
	mediaRepo := &memoryMediaRepo{assets: make(map[domainmedia.ID]domainmedia.Asset)}
	entryRepo := &memoryEntryRepo{entries: map[domaincontent.ID]domaincontent.Entry{
		"content-1": {ID: "content-1", Kind: domaincontent.KindPost},
	}}
	service := appmedia.NewService(mediaRepo, entryRepo)
	editor := authz.NewPrincipal("editor-1", authz.CapabilityMediaEdit)

	asset := domainmedia.Asset{ID: "asset-1", Filename: "cover.jpg", MimeType: "image/jpeg", PublicURL: "/media/cover.jpg"}
	if err := service.SaveMetadata(ctx, editor, asset); err != nil {
		t.Fatal(err)
	}
	entry, err := service.AttachFeatured(ctx, editor, "content-1", "asset-1")
	if err != nil {
		t.Fatal(err)
	}
	if entry.FeaturedMediaID != "asset-1" {
		t.Fatalf("expected featured media asset, got %q", entry.FeaturedMediaID)
	}
}

type memoryMediaRepo struct {
	assets map[domainmedia.ID]domainmedia.Asset
}

func (r *memoryMediaRepo) GetMedia(_ context.Context, id domainmedia.ID) (domainmedia.Asset, bool, error) {
	asset, ok := r.assets[id]
	return asset, ok, nil
}

func (r *memoryMediaRepo) ListMedia(context.Context) ([]domainmedia.Asset, error) {
	assets := make([]domainmedia.Asset, 0, len(r.assets))
	for _, asset := range r.assets {
		assets = append(assets, asset)
	}
	return assets, nil
}

func (r *memoryMediaRepo) SaveMedia(_ context.Context, asset domainmedia.Asset) error {
	r.assets[asset.ID] = asset
	return nil
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
