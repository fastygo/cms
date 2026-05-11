package sqlite

import (
	"context"
	"testing"
)

func TestInitAppliesMigrationsIdempotently(t *testing.T) {
	store, err := Open("file:migrations-test?mode=memory&cache=shared")
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer store.Close(context.Background())

	if err := store.Init(context.Background()); err != nil {
		t.Fatalf("Init() error = %v", err)
	}
	if err := store.Init(context.Background()); err != nil {
		t.Fatalf("Init() second call error = %v", err)
	}
	if err := store.MigrationStatus(context.Background()); err != nil {
		t.Fatalf("MigrationStatus() error = %v", err)
	}

	var count int
	if err := store.db.QueryRowContext(context.Background(), `SELECT COUNT(*) FROM schema_migrations`).Scan(&count); err != nil {
		t.Fatalf("count schema_migrations error = %v", err)
	}
	if count != len(sqliteMigrations()) {
		t.Fatalf("schema_migrations count = %d, want %d", count, len(sqliteMigrations()))
	}
}
