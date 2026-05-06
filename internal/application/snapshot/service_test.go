package snapshot

import (
	"context"
	"testing"
	"time"

	"github.com/fastygo/cms/internal/runtime/fixtures"
	sqlitestore "github.com/fastygo/cms/internal/storage/sqlite"
)

func TestExportAndImportRoundTrip(t *testing.T) {
	ctx := context.Background()
	source, err := sqlitestore.Open("file:snapshot-source?mode=memory&cache=shared")
	if err != nil {
		t.Fatalf("Open(source) error = %v", err)
	}
	defer source.Close(ctx)
	if err := source.Init(ctx); err != nil {
		t.Fatalf("Init(source) error = %v", err)
	}
	if err := fixtures.Seed(ctx, source); err != nil {
		t.Fatalf("Seed(source) error = %v", err)
	}
	service := NewService(source, func() time.Time { return time.Unix(10, 0).UTC() })
	bundle, err := service.Export(ctx)
	if err != nil {
		t.Fatalf("Export() error = %v", err)
	}
	if bundle.Version != SnapshotVersion {
		t.Fatalf("Version = %q, want %q", bundle.Version, SnapshotVersion)
	}
	if len(bundle.Content) == 0 || len(bundle.ContentTypes) == 0 || len(bundle.Settings) == 0 {
		t.Fatalf("expected exported bundle to contain seeded data")
	}

	target, err := sqlitestore.Open("file:snapshot-target?mode=memory&cache=shared")
	if err != nil {
		t.Fatalf("Open(target) error = %v", err)
	}
	defer target.Close(ctx)
	if err := target.Init(ctx); err != nil {
		t.Fatalf("Init(target) error = %v", err)
	}
	targetService := NewService(target, time.Now)
	if err := targetService.Import(ctx, bundle); err != nil {
		t.Fatalf("Import() error = %v", err)
	}
	exportedAgain, err := targetService.Export(ctx)
	if err != nil {
		t.Fatalf("Export(target) error = %v", err)
	}
	if len(exportedAgain.Content) != len(bundle.Content) {
		t.Fatalf("imported content count = %d, want %d", len(exportedAgain.Content), len(bundle.Content))
	}
}
