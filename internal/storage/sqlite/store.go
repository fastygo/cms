package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	domaincontent "github.com/fastygo/cms/internal/domain/content"
	domaincontenttype "github.com/fastygo/cms/internal/domain/contenttype"
	domainmedia "github.com/fastygo/cms/internal/domain/media"
	domainmenus "github.com/fastygo/cms/internal/domain/menus"
	domainpreview "github.com/fastygo/cms/internal/domain/preview"
	domainrevisions "github.com/fastygo/cms/internal/domain/revisions"
	domainsettings "github.com/fastygo/cms/internal/domain/settings"
	domaintaxonomy "github.com/fastygo/cms/internal/domain/taxonomy"
	domainusers "github.com/fastygo/cms/internal/domain/users"
	_ "modernc.org/sqlite"
)

type Store struct {
	db *sql.DB
}

func Open(dataSource string) (*Store, error) {
	if strings.TrimSpace(dataSource) == "" || dataSource == "fixture" {
		dataSource = "file:gocms.db"
	}
	db, err := sql.Open("sqlite", dataSource)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	return &Store{db: db}, nil
}

func (s *Store) Init(ctx context.Context) error {
	statements := []string{
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
	}
	for _, statement := range statements {
		if _, err := s.db.ExecContext(ctx, statement); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) Close(context.Context) error {
	return s.db.Close()
}

func (s *Store) HealthCheck(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

func (s *Store) NextID(ctx context.Context) (domaincontent.ID, error) {
	id, err := s.nextID(ctx, "content_entries", "id", "content-")
	return domaincontent.ID(id), err
}

func (s *Store) Save(ctx context.Context, entry domaincontent.Entry) error {
	payload, err := encode(entry)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO content_entries (id, kind, status, author_id, entry_json, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			kind = excluded.kind,
			status = excluded.status,
			author_id = excluded.author_id,
			entry_json = excluded.entry_json,
			updated_at = excluded.updated_at
	`, string(entry.ID), string(entry.Kind), string(entry.Status), entry.AuthorID, payload, entry.UpdatedAt.Format(time.RFC3339Nano))
	return err
}

func (s *Store) Get(ctx context.Context, id domaincontent.ID) (domaincontent.Entry, error) {
	var payload string
	if err := s.db.QueryRowContext(ctx, `SELECT entry_json FROM content_entries WHERE id = ?`, string(id)).Scan(&payload); err != nil {
		if err == sql.ErrNoRows {
			return domaincontent.Entry{}, fmt.Errorf("content %q not found", id)
		}
		return domaincontent.Entry{}, err
	}
	var entry domaincontent.Entry
	return entry, decode(payload, &entry)
}

func (s *Store) List(ctx context.Context, query domaincontent.Query) (domaincontent.ListResult, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT entry_json FROM content_entries`)
	if err != nil {
		return domaincontent.ListResult{}, err
	}
	defer rows.Close()

	var entries []domaincontent.Entry
	for rows.Next() {
		var payload string
		if err := rows.Scan(&payload); err != nil {
			return domaincontent.ListResult{}, err
		}
		var entry domaincontent.Entry
		if err := decode(payload, &entry); err != nil {
			return domaincontent.ListResult{}, err
		}
		if contentMatches(entry, query) {
			entries = append(entries, entry)
		}
	}
	if err := rows.Err(); err != nil {
		return domaincontent.ListResult{}, err
	}

	sortContent(entries, query)
	total := len(entries)
	page, perPage := normalizePagination(query.Page, query.PerPage)
	start := (page - 1) * perPage
	if start > total {
		start = total
	}
	end := start + perPage
	if end > total {
		end = total
	}
	totalPages := 0
	if total > 0 {
		totalPages = (total + perPage - 1) / perPage
	}
	return domaincontent.ListResult{
		Items:      entries[start:end],
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	}, nil
}

func (s *Store) SaveType(ctx context.Context, contentType domaincontenttype.Type) error {
	return s.SaveContentType(ctx, contentType)
}

func (s *Store) SaveContentType(ctx context.Context, contentType domaincontenttype.Type) error {
	payload, err := encode(contentType)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `INSERT INTO content_types (id, type_json) VALUES (?, ?) ON CONFLICT(id) DO UPDATE SET type_json = excluded.type_json`, string(contentType.ID), payload)
	return err
}

func (s *Store) GetContentType(ctx context.Context, kind domaincontent.Kind) (domaincontenttype.Type, bool, error) {
	return s.GetType(ctx, kind)
}

func (s *Store) GetType(ctx context.Context, kind domaincontent.Kind) (domaincontenttype.Type, bool, error) {
	var payload string
	if err := s.db.QueryRowContext(ctx, `SELECT type_json FROM content_types WHERE id = ?`, string(kind)).Scan(&payload); err != nil {
		if err == sql.ErrNoRows {
			return domaincontenttype.Type{}, false, nil
		}
		return domaincontenttype.Type{}, false, err
	}
	var contentType domaincontenttype.Type
	return contentType, true, decode(payload, &contentType)
}

func (s *Store) ListContentTypes(ctx context.Context) ([]domaincontenttype.Type, error) {
	return s.ListTypes(ctx)
}

func (s *Store) ListTypes(ctx context.Context) ([]domaincontenttype.Type, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT type_json FROM content_types ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []domaincontenttype.Type
	for rows.Next() {
		var payload string
		if err := rows.Scan(&payload); err != nil {
			return nil, err
		}
		var contentType domaincontenttype.Type
		if err := decode(payload, &contentType); err != nil {
			return nil, err
		}
		result = append(result, contentType)
	}
	return result, rows.Err()
}

func (s *Store) SaveDefinition(ctx context.Context, definition domaintaxonomy.Definition) error {
	payload, err := encode(definition)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `INSERT INTO taxonomy_definitions (type, definition_json) VALUES (?, ?) ON CONFLICT(type) DO UPDATE SET definition_json = excluded.definition_json`, string(definition.Type), payload)
	return err
}

func (s *Store) GetDefinition(ctx context.Context, taxonomyType domaintaxonomy.Type) (domaintaxonomy.Definition, bool, error) {
	var payload string
	if err := s.db.QueryRowContext(ctx, `SELECT definition_json FROM taxonomy_definitions WHERE type = ?`, string(taxonomyType)).Scan(&payload); err != nil {
		if err == sql.ErrNoRows {
			return domaintaxonomy.Definition{}, false, nil
		}
		return domaintaxonomy.Definition{}, false, err
	}
	var definition domaintaxonomy.Definition
	return definition, true, decode(payload, &definition)
}

func (s *Store) ListDefinitions(ctx context.Context) ([]domaintaxonomy.Definition, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT definition_json FROM taxonomy_definitions ORDER BY type`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []domaintaxonomy.Definition
	for rows.Next() {
		var payload string
		if err := rows.Scan(&payload); err != nil {
			return nil, err
		}
		var definition domaintaxonomy.Definition
		if err := decode(payload, &definition); err != nil {
			return nil, err
		}
		result = append(result, definition)
	}
	return result, rows.Err()
}

func (s *Store) SaveTerm(ctx context.Context, term domaintaxonomy.Term) error {
	payload, err := encode(term)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `INSERT INTO taxonomy_terms (id, type, term_json) VALUES (?, ?, ?) ON CONFLICT(id) DO UPDATE SET type = excluded.type, term_json = excluded.term_json`, string(term.ID), string(term.Type), payload)
	return err
}

func (s *Store) GetTerm(ctx context.Context, id domaintaxonomy.TermID) (domaintaxonomy.Term, bool, error) {
	var payload string
	if err := s.db.QueryRowContext(ctx, `SELECT term_json FROM taxonomy_terms WHERE id = ?`, string(id)).Scan(&payload); err != nil {
		if err == sql.ErrNoRows {
			return domaintaxonomy.Term{}, false, nil
		}
		return domaintaxonomy.Term{}, false, err
	}
	var term domaintaxonomy.Term
	return term, true, decode(payload, &term)
}

func (s *Store) ListTerms(ctx context.Context, taxonomyType domaintaxonomy.Type) ([]domaintaxonomy.Term, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT term_json FROM taxonomy_terms WHERE type = ? ORDER BY id`, string(taxonomyType))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []domaintaxonomy.Term
	for rows.Next() {
		var payload string
		if err := rows.Scan(&payload); err != nil {
			return nil, err
		}
		var term domaintaxonomy.Term
		if err := decode(payload, &term); err != nil {
			return nil, err
		}
		result = append(result, term)
	}
	return result, rows.Err()
}

func (s *Store) SaveMedia(ctx context.Context, asset domainmedia.Asset) error {
	return s.SaveAsset(ctx, asset)
}

func (s *Store) SaveAsset(ctx context.Context, asset domainmedia.Asset) error {
	payload, err := encode(asset)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `INSERT INTO media_assets (id, asset_json) VALUES (?, ?) ON CONFLICT(id) DO UPDATE SET asset_json = excluded.asset_json`, string(asset.ID), payload)
	return err
}

func (s *Store) GetAsset(ctx context.Context, id domainmedia.ID) (domainmedia.Asset, bool, error) {
	return s.GetMedia(ctx, id)
}

func (s *Store) GetMedia(ctx context.Context, id domainmedia.ID) (domainmedia.Asset, bool, error) {
	var payload string
	if err := s.db.QueryRowContext(ctx, `SELECT asset_json FROM media_assets WHERE id = ?`, string(id)).Scan(&payload); err != nil {
		if err == sql.ErrNoRows {
			return domainmedia.Asset{}, false, nil
		}
		return domainmedia.Asset{}, false, err
	}
	var asset domainmedia.Asset
	return asset, true, decode(payload, &asset)
}

func (s *Store) ListMedia(ctx context.Context) ([]domainmedia.Asset, error) {
	return s.ListAssets(ctx)
}

func (s *Store) ListAssets(ctx context.Context) ([]domainmedia.Asset, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT asset_json FROM media_assets ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []domainmedia.Asset
	for rows.Next() {
		var payload string
		if err := rows.Scan(&payload); err != nil {
			return nil, err
		}
		var asset domainmedia.Asset
		if err := decode(payload, &asset); err != nil {
			return nil, err
		}
		result = append(result, asset)
	}
	return result, rows.Err()
}

func (s *Store) SaveUser(ctx context.Context, user domainusers.User) error {
	payload, err := encode(user)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `INSERT INTO users (id, user_json) VALUES (?, ?) ON CONFLICT(id) DO UPDATE SET user_json = excluded.user_json`, string(user.ID), payload)
	return err
}

func (s *Store) GetUser(ctx context.Context, id domainusers.ID) (domainusers.User, bool, error) {
	var payload string
	if err := s.db.QueryRowContext(ctx, `SELECT user_json FROM users WHERE id = ?`, string(id)).Scan(&payload); err != nil {
		if err == sql.ErrNoRows {
			return domainusers.User{}, false, nil
		}
		return domainusers.User{}, false, err
	}
	var user domainusers.User
	return user, true, decode(payload, &user)
}

func (s *Store) ListUsers(ctx context.Context) ([]domainusers.User, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT user_json FROM users ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []domainusers.User
	for rows.Next() {
		var payload string
		if err := rows.Scan(&payload); err != nil {
			return nil, err
		}
		var user domainusers.User
		if err := decode(payload, &user); err != nil {
			return nil, err
		}
		result = append(result, user)
	}
	return result, rows.Err()
}

func (s *Store) SaveSetting(ctx context.Context, value domainsettings.Value) error {
	payload, err := encode(value)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `INSERT INTO settings (key, setting_json) VALUES (?, ?) ON CONFLICT(key) DO UPDATE SET setting_json = excluded.setting_json`, string(value.Key), payload)
	return err
}

func (s *Store) GetSetting(ctx context.Context, key domainsettings.Key) (domainsettings.Value, bool, error) {
	var payload string
	if err := s.db.QueryRowContext(ctx, `SELECT setting_json FROM settings WHERE key = ?`, string(key)).Scan(&payload); err != nil {
		if err == sql.ErrNoRows {
			return domainsettings.Value{}, false, nil
		}
		return domainsettings.Value{}, false, err
	}
	var value domainsettings.Value
	return value, true, decode(payload, &value)
}

func (s *Store) ListPublicSettings(ctx context.Context) ([]domainsettings.Value, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT setting_json FROM settings ORDER BY key`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []domainsettings.Value
	for rows.Next() {
		var payload string
		if err := rows.Scan(&payload); err != nil {
			return nil, err
		}
		var value domainsettings.Value
		if err := decode(payload, &value); err != nil {
			return nil, err
		}
		if value.Public {
			result = append(result, value)
		}
	}
	return result, rows.Err()
}

func (s *Store) SaveMenu(ctx context.Context, menu domainmenus.Menu) error {
	payload, err := encode(menu)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `INSERT INTO menus (id, location, menu_json) VALUES (?, ?, ?) ON CONFLICT(id) DO UPDATE SET location = excluded.location, menu_json = excluded.menu_json`, string(menu.ID), string(menu.Location), payload)
	return err
}

func (s *Store) ListMenus(ctx context.Context) ([]domainmenus.Menu, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT menu_json FROM menus ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []domainmenus.Menu
	for rows.Next() {
		var payload string
		if err := rows.Scan(&payload); err != nil {
			return nil, err
		}
		var menu domainmenus.Menu
		if err := decode(payload, &menu); err != nil {
			return nil, err
		}
		result = append(result, menu)
	}
	return result, rows.Err()
}

func (s *Store) GetMenuByLocation(ctx context.Context, location domainmenus.Location) (domainmenus.Menu, bool, error) {
	var payload string
	if err := s.db.QueryRowContext(ctx, `SELECT menu_json FROM menus WHERE location = ? LIMIT 1`, string(location)).Scan(&payload); err != nil {
		if err == sql.ErrNoRows {
			return domainmenus.Menu{}, false, nil
		}
		return domainmenus.Menu{}, false, err
	}
	var menu domainmenus.Menu
	return menu, true, decode(payload, &menu)
}

func (s *Store) NextRevisionID(ctx context.Context) (domainrevisions.ID, error) {
	id, err := s.nextID(ctx, "revisions", "id", "revision-")
	return domainrevisions.ID(id), err
}

func (s *Store) SaveRevision(ctx context.Context, revision domainrevisions.Revision) error {
	payload, err := encode(revision)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `INSERT INTO revisions (id, entry_id, revision_json) VALUES (?, ?, ?) ON CONFLICT(id) DO UPDATE SET entry_id = excluded.entry_id, revision_json = excluded.revision_json`, string(revision.ID), string(revision.EntryID), payload)
	return err
}

func (s *Store) GetRevision(ctx context.Context, id domainrevisions.ID) (domainrevisions.Revision, error) {
	var payload string
	if err := s.db.QueryRowContext(ctx, `SELECT revision_json FROM revisions WHERE id = ?`, string(id)).Scan(&payload); err != nil {
		if err == sql.ErrNoRows {
			return domainrevisions.Revision{}, fmt.Errorf("revision %q not found", id)
		}
		return domainrevisions.Revision{}, err
	}
	var revision domainrevisions.Revision
	return revision, decode(payload, &revision)
}

func (s *Store) SavePreview(ctx context.Context, access domainpreview.Access) error {
	payload, err := encode(access)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `INSERT INTO preview_access (token, access_json) VALUES (?, ?) ON CONFLICT(token) DO UPDATE SET access_json = excluded.access_json`, string(access.Token), payload)
	return err
}

func (s *Store) GetPreview(ctx context.Context, token domainpreview.Token) (domainpreview.Access, bool, error) {
	var payload string
	if err := s.db.QueryRowContext(ctx, `SELECT access_json FROM preview_access WHERE token = ?`, string(token)).Scan(&payload); err != nil {
		if err == sql.ErrNoRows {
			return domainpreview.Access{}, false, nil
		}
		return domainpreview.Access{}, false, err
	}
	var access domainpreview.Access
	return access, true, decode(payload, &access)
}

func (s *Store) nextID(ctx context.Context, table string, column string, prefix string) (string, error) {
	var count int
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", table)
	if err := s.db.QueryRowContext(ctx, query).Scan(&count); err != nil {
		return "", err
	}
	for i := count + 1; ; i++ {
		id := fmt.Sprintf("%s%d", prefix, i)
		var exists int
		query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s = ?", table, column)
		if err := s.db.QueryRowContext(ctx, query, id).Scan(&exists); err != nil {
			return "", err
		}
		if exists == 0 {
			return id, nil
		}
	}
}

func contentMatches(entry domaincontent.Entry, query domaincontent.Query) bool {
	if query.PublicOnly && !entry.IsPublicAt(query.PublishedAt) {
		return false
	}
	if len(query.Kinds) > 0 && !kindIn(entry.Kind, query.Kinds) {
		return false
	}
	if len(query.Statuses) > 0 && !statusIn(entry.Status, query.Statuses) {
		return false
	}
	if query.AuthorID != "" && entry.AuthorID != query.AuthorID {
		return false
	}
	if query.Slug != "" && !localizedEquals(entry.Slug, query.Locale, query.Slug) {
		return false
	}
	if query.Taxonomy != "" || query.TermID != "" {
		if !hasTerm(entry, query.Taxonomy, query.TermID) {
			return false
		}
	}
	if query.Search != "" && !matchesSearch(entry, query.Search) {
		return false
	}
	point := entry.CreatedAt
	if entry.PublishedAt != nil {
		point = *entry.PublishedAt
	}
	if query.After != nil && point.Before(*query.After) {
		return false
	}
	if query.Before != nil && point.After(*query.Before) {
		return false
	}
	return true
}

func sortContent(entries []domaincontent.Entry, query domaincontent.Query) {
	sort.SliceStable(entries, func(i, j int) bool {
		less := contentLess(entries[i], entries[j], query)
		if query.SortDesc {
			return !less
		}
		return less
	})
}

func contentLess(a domaincontent.Entry, b domaincontent.Entry, query domaincontent.Query) bool {
	switch query.SortBy {
	case domaincontent.SortTitle:
		return strings.ToLower(a.Title.Value(query.Locale, "en")) < strings.ToLower(b.Title.Value(query.Locale, "en"))
	case domaincontent.SortSlug:
		return strings.ToLower(a.Slug.Value(query.Locale, "en")) < strings.ToLower(b.Slug.Value(query.Locale, "en"))
	case domaincontent.SortPublishedAt:
		return timeValue(a.PublishedAt, a.CreatedAt).Before(timeValue(b.PublishedAt, b.CreatedAt))
	case domaincontent.SortCreatedAt:
		return a.CreatedAt.Before(b.CreatedAt)
	default:
		return a.UpdatedAt.Before(b.UpdatedAt)
	}
}

func normalizePagination(page int, perPage int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if perPage <= 0 {
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}
	return page, perPage
}

func kindIn(kind domaincontent.Kind, values []domaincontent.Kind) bool {
	for _, value := range values {
		if kind == value {
			return true
		}
	}
	return false
}

func statusIn(status domaincontent.Status, values []domaincontent.Status) bool {
	for _, value := range values {
		if status == value {
			return true
		}
	}
	return false
}

func localizedEquals(values domaincontent.LocalizedText, locale string, expected string) bool {
	expected = strings.ToLower(strings.TrimSpace(expected))
	if expected == "" {
		return true
	}
	if locale != "" {
		return strings.ToLower(values.Value(locale, "en")) == expected
	}
	for _, value := range values {
		if strings.ToLower(strings.TrimSpace(value)) == expected {
			return true
		}
	}
	return false
}

func hasTerm(entry domaincontent.Entry, taxonomy string, termID string) bool {
	for _, ref := range entry.Terms {
		if taxonomy != "" && ref.Taxonomy != taxonomy {
			continue
		}
		if termID != "" && ref.TermID != termID {
			continue
		}
		return true
	}
	return false
}

func matchesSearch(entry domaincontent.Entry, value string) bool {
	needle := strings.ToLower(strings.TrimSpace(value))
	if needle == "" {
		return true
	}
	for _, values := range []domaincontent.LocalizedText{entry.Title, entry.Body, entry.Excerpt, entry.Slug} {
		for _, text := range values {
			if strings.Contains(strings.ToLower(text), needle) {
				return true
			}
		}
	}
	return false
}

func timeValue(value *time.Time, fallback time.Time) time.Time {
	if value == nil {
		return fallback
	}
	return *value
}

func encode(value any) (string, error) {
	raw, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

func decode(payload string, out any) error {
	return json.Unmarshal([]byte(payload), out)
}
