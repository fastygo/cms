package sqlite_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/fastygo/cms/internal/domain/content"
	"github.com/fastygo/cms/internal/runtime/fixtures"
	sqlitestore "github.com/fastygo/cms/internal/storage/sqlite"
)

func TestStorePersistsAndFiltersSeedContent(t *testing.T) {
	ctx := context.Background()
	path := "file:" + filepath.ToSlash(filepath.Join(t.TempDir(), "gocms.db"))

	store, err := sqlitestore.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Init(ctx); err != nil {
		t.Fatal(err)
	}
	if err := fixtures.Seed(ctx, store); err != nil {
		t.Fatal(err)
	}
	if err := store.Close(ctx); err != nil {
		t.Fatal(err)
	}

	store, err = sqlitestore.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := store.Close(ctx); err != nil {
			t.Fatal(err)
		}
	}()
	if err := store.Init(ctx); err != nil {
		t.Fatal(err)
	}

	now := time.Date(2026, 5, 2, 10, 0, 0, 0, time.UTC)
	result, err := store.List(ctx, content.Query{
		Kinds:       []content.Kind{content.KindPost},
		PublicOnly:  true,
		PublishedAt: now,
		Page:        1,
		PerPage:     10,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Total != 1 || result.Items[0].ID != "content-post-published" {
		t.Fatalf("expected only published seeded post, got %+v", result)
	}

	result, err = store.List(ctx, content.Query{
		Slug:        "published-post",
		Locale:      "en",
		PublicOnly:  true,
		PublishedAt: now,
		Page:        1,
		PerPage:     10,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Total != 1 {
		t.Fatalf("expected slug filter result, got %+v", result)
	}

	result, err = store.List(ctx, content.Query{
		Taxonomy:    "category",
		TermID:      "term-news",
		PublicOnly:  true,
		PublishedAt: now,
		Page:        1,
		PerPage:     10,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Total != 1 {
		t.Fatalf("expected taxonomy filter result, got %+v", result)
	}

	result, err = store.List(ctx, content.Query{
		Search:      "fixture",
		PublicOnly:  true,
		PublishedAt: now,
		Page:        1,
		PerPage:     1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.PerPage != 1 || result.TotalPages == 0 {
		t.Fatalf("expected paginated search result, got %+v", result)
	}
}
