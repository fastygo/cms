package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"slices"
	"time"
)

type migration struct {
	Version     string
	Description string
	Statements  []string
}

func sqliteMigrations() []migration {
	return []migration{
		{
			Version:     "0001_core_schema",
			Description: "create core content, taxonomy, media, users, settings, menus, revisions, and preview tables",
			Statements: []string{
				`CREATE TABLE IF NOT EXISTS content_entries (id TEXT PRIMARY KEY, kind TEXT NOT NULL, status TEXT NOT NULL, author_id TEXT, entry_json TEXT NOT NULL, updated_at TEXT NOT NULL)`,
				`CREATE TABLE IF NOT EXISTS content_types (id TEXT PRIMARY KEY, type_json TEXT NOT NULL)`,
				`CREATE TABLE IF NOT EXISTS taxonomy_definitions (type TEXT PRIMARY KEY, definition_json TEXT NOT NULL)`,
				`CREATE TABLE IF NOT EXISTS taxonomy_terms (id TEXT PRIMARY KEY, type TEXT NOT NULL, term_json TEXT NOT NULL)`,
				`CREATE TABLE IF NOT EXISTS media_assets (id TEXT PRIMARY KEY, asset_json TEXT NOT NULL)`,
				`CREATE TABLE IF NOT EXISTS users (id TEXT PRIMARY KEY, user_json TEXT NOT NULL)`,
				`CREATE TABLE IF NOT EXISTS settings (key TEXT PRIMARY KEY, setting_json TEXT NOT NULL)`,
				`CREATE TABLE IF NOT EXISTS menus (id TEXT PRIMARY KEY, location TEXT NOT NULL, menu_json TEXT NOT NULL)`,
				`CREATE TABLE IF NOT EXISTS revisions (id TEXT PRIMARY KEY, entry_id TEXT NOT NULL, revision_json TEXT NOT NULL)`,
				`CREATE TABLE IF NOT EXISTS preview_access (token TEXT PRIMARY KEY, access_json TEXT NOT NULL)`,
				`CREATE INDEX IF NOT EXISTS idx_content_kind ON content_entries(kind)`,
				`CREATE INDEX IF NOT EXISTS idx_content_status ON content_entries(status)`,
				`CREATE INDEX IF NOT EXISTS idx_terms_type ON taxonomy_terms(type)`,
				`CREATE INDEX IF NOT EXISTS idx_menus_location ON menus(location)`,
			},
		},
		{
			Version:     "0002_auth_audit",
			Description: "create authn and audit tables",
			Statements: []string{
				`CREATE TABLE IF NOT EXISTS auth_recovery_codes (id TEXT PRIMARY KEY, user_id TEXT NOT NULL, used_at TEXT, recovery_json TEXT NOT NULL)`,
				`CREATE TABLE IF NOT EXISTS auth_reset_tokens (id TEXT PRIMARY KEY, user_id TEXT NOT NULL, expires_at TEXT NOT NULL, reset_json TEXT NOT NULL)`,
				`CREATE TABLE IF NOT EXISTS auth_app_tokens (id TEXT PRIMARY KEY, prefix TEXT NOT NULL UNIQUE, user_id TEXT NOT NULL, expires_at TEXT, revoked_at TEXT, token_json TEXT NOT NULL)`,
				`CREATE TABLE IF NOT EXISTS auth_login_attempts (id TEXT PRIMARY KEY, attempt_key TEXT NOT NULL, created_at TEXT NOT NULL, success INTEGER NOT NULL, attempt_json TEXT NOT NULL)`,
				`CREATE TABLE IF NOT EXISTS audit_events (id TEXT PRIMARY KEY, occurred_at TEXT NOT NULL, actor_id TEXT, action TEXT NOT NULL, resource TEXT NOT NULL, resource_id TEXT, event_json TEXT NOT NULL)`,
				`CREATE INDEX IF NOT EXISTS idx_auth_recovery_user ON auth_recovery_codes(user_id)`,
				`CREATE INDEX IF NOT EXISTS idx_auth_reset_user ON auth_reset_tokens(user_id)`,
				`CREATE INDEX IF NOT EXISTS idx_auth_app_user ON auth_app_tokens(user_id)`,
				`CREATE INDEX IF NOT EXISTS idx_auth_attempts_key_created ON auth_login_attempts(attempt_key, created_at)`,
				`CREATE INDEX IF NOT EXISTS idx_audit_occurred_at ON audit_events(occurred_at)`,
			},
		},
		{
			Version:     "0003_error_logs",
			Description: "create bounded error log table",
			Statements: []string{
				`CREATE TABLE IF NOT EXISTS error_logs (id TEXT PRIMARY KEY, occurred_at TEXT NOT NULL, source TEXT NOT NULL, severity TEXT NOT NULL, error_json TEXT NOT NULL)`,
				`CREATE INDEX IF NOT EXISTS idx_error_logs_occurred_at ON error_logs(occurred_at)`,
			},
		},
	}
}

func (s *Store) applyMigrations(ctx context.Context) error {
	if _, err := s.db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (version TEXT PRIMARY KEY, description TEXT NOT NULL, applied_at TEXT NOT NULL)`); err != nil {
		return err
	}
	applied, err := s.appliedMigrations(ctx)
	if err != nil {
		return err
	}
	for _, migration := range sqliteMigrations() {
		if slices.Contains(applied, migration.Version) {
			continue
		}
		if err := s.applyMigration(ctx, migration); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) appliedMigrations(ctx context.Context) ([]string, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT version FROM schema_migrations ORDER BY version`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []string{}
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		result = append(result, version)
	}
	return result, rows.Err()
}

func (s *Store) applyMigration(ctx context.Context, migration migration) error {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, statement := range migration.Statements {
		if _, err := tx.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("apply migration %s: %w", migration.Version, err)
		}
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO schema_migrations (version, description, applied_at) VALUES (?, ?, ?)`, migration.Version, migration.Description, time.Now().UTC().Format(time.RFC3339Nano)); err != nil {
		return fmt.Errorf("record migration %s: %w", migration.Version, err)
	}
	return tx.Commit()
}

func (s *Store) MigrationStatus(ctx context.Context) error {
	applied, err := s.appliedMigrations(ctx)
	if err != nil {
		return err
	}
	expected := sqliteMigrations()
	for _, migration := range expected {
		if !slices.Contains(applied, migration.Version) {
			return fmt.Errorf("pending migration %s", migration.Version)
		}
	}
	return nil
}
