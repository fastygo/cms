package playground

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	appsnapshot "github.com/fastygo/cms/internal/application/snapshot"
	domaincontent "github.com/fastygo/cms/internal/domain/content"
	domainmedia "github.com/fastygo/cms/internal/domain/media"
	domainsettings "github.com/fastygo/cms/internal/domain/settings"
	domaintaxonomy "github.com/fastygo/cms/internal/domain/taxonomy"
)

type routeContentRecord struct {
	ID            string       `json:"id"`
	Slug          string       `json:"slug"`
	Status        string       `json:"status"`
	Author        string       `json:"author,omitempty"`
	FeaturedMedia string       `json:"featured_media,omitempty"`
	Date          string       `json:"date,omitempty"`
	Modified      string       `json:"modified,omitempty"`
	Type          string       `json:"type,omitempty"`
	Title         renderedText `json:"title"`
	Content       renderedText `json:"content"`
	Excerpt       renderedText `json:"excerpt"`
}

type routeTermRecord struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Slug     string `json:"slug"`
	Taxonomy string `json:"taxonomy"`
}

type routeMediaRecord struct {
	ID        string       `json:"id"`
	SourceURL string       `json:"source_url"`
	MimeType  string       `json:"mime_type"`
	Title     renderedText `json:"title"`
	Caption   renderedText `json:"caption"`
	AltText   string       `json:"alt_text,omitempty"`
}

type renderedText struct {
	Rendered string `json:"rendered"`
}

func SnapshotFromBundle(bundle appsnapshot.Bundle, source Source) (Snapshot, error) {
	posts := make([]routeContentRecord, 0)
	pages := make([]routeContentRecord, 0)
	categories := make([]routeTermRecord, 0)
	tags := make([]routeTermRecord, 0)
	media := make([]routeMediaRecord, 0, len(bundle.Media))
	mediaMetadata := make([]MediaMetadata, 0, len(bundle.Media))
	settings := make([]SnapshotSetting, 0, len(bundle.Settings))

	for _, entry := range bundle.Content {
		record := contentRouteRecord(entry)
		switch entry.Kind {
		case domaincontent.KindPost:
			posts = append(posts, record)
		case domaincontent.KindPage:
			pages = append(pages, record)
		}
	}
	for _, term := range bundle.TaxonomyTerms {
		record := termRouteRecord(term)
		switch term.Type {
		case domaintaxonomy.Type("category"):
			categories = append(categories, record)
		case domaintaxonomy.Type("tag"):
			tags = append(tags, record)
		}
	}
	for _, asset := range bundle.Media {
		media = append(media, mediaRouteRecord(asset))
		mediaMetadata = append(mediaMetadata, mediaMetadataRecord(asset))
	}
	for _, value := range bundle.Settings {
		raw, err := json.Marshal(value.Value)
		if err != nil {
			return Snapshot{}, err
		}
		settings = append(settings, SnapshotSetting{
			Key:    string(value.Key),
			Value:  raw,
			Public: value.Public,
		})
	}

	routes, err := marshalRoutes(map[string]any{
		RoutePosts:      posts,
		RoutePages:      pages,
		RouteCategories: categories,
		RouteTags:       tags,
		RouteMedia:      media,
	})
	if err != nil {
		return Snapshot{}, err
	}

	if strings.TrimSpace(source.Kind) == "" {
		source.Kind = "gocms-snapshot"
	}
	return Snapshot{
		Version:  DefaultSnapshotVersion,
		Source:   source,
		Routes:   routes,
		Settings: settings,
		Local: SnapshotLocal{
			MediaBlobs:    BlobStatusExcluded,
			MediaMetadata: mediaMetadata,
		},
	}, nil
}

func BundleFromSnapshot(snapshot Snapshot, now func() time.Time) (appsnapshot.Bundle, error) {
	if now == nil {
		now = time.Now
	}
	var (
		posts      []routeContentRecord
		pages      []routeContentRecord
		categories []routeTermRecord
		tags       []routeTermRecord
		mediaRows  []routeMediaRecord
	)
	if err := decodeRoute(snapshot.Routes, RoutePosts, &posts); err != nil {
		return appsnapshot.Bundle{}, err
	}
	if err := decodeRoute(snapshot.Routes, RoutePages, &pages); err != nil {
		return appsnapshot.Bundle{}, err
	}
	if err := decodeRoute(snapshot.Routes, RouteCategories, &categories); err != nil {
		return appsnapshot.Bundle{}, err
	}
	if err := decodeRoute(snapshot.Routes, RouteTags, &tags); err != nil {
		return appsnapshot.Bundle{}, err
	}
	if err := decodeRoute(snapshot.Routes, RouteMedia, &mediaRows); err != nil {
		return appsnapshot.Bundle{}, err
	}

	content := make([]domaincontent.Entry, 0, len(posts)+len(pages))
	for _, record := range posts {
		content = append(content, contentFromRouteRecord(record, domaincontent.KindPost, now()))
	}
	for _, record := range pages {
		content = append(content, contentFromRouteRecord(record, domaincontent.KindPage, now()))
	}

	mediaMetadata := make(map[string]MediaMetadata, len(snapshot.Local.MediaMetadata))
	for _, item := range snapshot.Local.MediaMetadata {
		mediaMetadata[item.ID] = item
	}
	assets := make([]domainmedia.Asset, 0, len(mediaRows))
	for _, record := range mediaRows {
		assets = append(assets, mediaFromRouteRecord(record, mediaMetadata[record.ID], now()))
	}

	settings := make([]domainsettings.Value, 0, len(snapshot.Settings))
	for _, item := range snapshot.Settings {
		var value any
		if len(item.Value) > 0 {
			if err := json.Unmarshal(item.Value, &value); err != nil {
				return appsnapshot.Bundle{}, fmt.Errorf("decode playground setting %q: %w", item.Key, err)
			}
		}
		settings = append(settings, domainsettings.Value{
			Key:    domainsettings.Key(item.Key),
			Value:  value,
			Public: item.Public,
		})
	}

	taxonomyDefinitions := make([]domaintaxonomy.Definition, 0, 2)
	taxonomyTerms := make([]domaintaxonomy.Term, 0, len(categories)+len(tags))
	if len(categories) > 0 {
		taxonomyDefinitions = append(taxonomyDefinitions, taxonomyDefinition("category", "Categories"))
		for _, item := range categories {
			taxonomyTerms = append(taxonomyTerms, taxonomyTermFromRoute(item, "category"))
		}
	}
	if len(tags) > 0 {
		taxonomyDefinitions = append(taxonomyDefinitions, taxonomyDefinition("tag", "Tags"))
		for _, item := range tags {
			taxonomyTerms = append(taxonomyTerms, taxonomyTermFromRoute(item, "tag"))
		}
	}

	return appsnapshot.Bundle{
		Version:             appsnapshot.SnapshotVersion,
		ExportedAt:          now().UTC(),
		Content:             content,
		TaxonomyDefinitions: taxonomyDefinitions,
		TaxonomyTerms:       taxonomyTerms,
		Media:               assets,
		Settings:            settings,
	}, nil
}

func marshalRoutes(values map[string]any) (map[string]json.RawMessage, error) {
	routes := map[string]json.RawMessage{}
	for _, key := range []string{RoutePosts, RoutePages, RouteCategories, RouteTags, RouteMedia} {
		payload, err := json.Marshal(values[key])
		if err != nil {
			return nil, err
		}
		routes[key] = payload
	}
	return routes, nil
}

func decodeRoute(routes map[string]json.RawMessage, key string, target any) error {
	payload, ok := routes[key]
	if !ok || len(payload) == 0 {
		payload = json.RawMessage(`[]`)
	}
	if err := json.Unmarshal(payload, target); err != nil {
		return fmt.Errorf("decode playground route %q: %w", key, err)
	}
	return nil
}

func contentRouteRecord(entry domaincontent.Entry) routeContentRecord {
	return routeContentRecord{
		ID:            string(entry.ID),
		Slug:          entry.Slug.Value("en", "en"),
		Status:        string(entry.Status),
		Author:        entry.AuthorID,
		FeaturedMedia: entry.FeaturedMediaID,
		Date:          entry.CreatedAt.UTC().Format(time.RFC3339),
		Modified:      entry.UpdatedAt.UTC().Format(time.RFC3339),
		Type:          string(entry.Kind),
		Title:         renderedText{Rendered: entry.Title.Value("en", "en")},
		Content:       renderedText{Rendered: entry.Body.Value("en", "en")},
		Excerpt:       renderedText{Rendered: entry.Excerpt.Value("en", "en")},
	}
}

func termRouteRecord(term domaintaxonomy.Term) routeTermRecord {
	return routeTermRecord{
		ID:       string(term.ID),
		Name:     term.Name.Value("en", "en"),
		Slug:     term.Slug.Value("en", "en"),
		Taxonomy: string(term.Type),
	}
}

func mediaRouteRecord(asset domainmedia.Asset) routeMediaRecord {
	return routeMediaRecord{
		ID:        string(asset.ID),
		SourceURL: asset.PublicURL,
		MimeType:  asset.MimeType,
		Title:     renderedText{Rendered: asset.Filename},
		Caption:   renderedText{Rendered: asset.Caption},
		AltText:   asset.AltText,
	}
}

func mediaMetadataRecord(asset domainmedia.Asset) MediaMetadata {
	return MediaMetadata{
		ID:         string(asset.ID),
		Filename:   asset.Filename,
		MimeType:   asset.MimeType,
		Width:      asset.Width,
		Height:     asset.Height,
		Size:       asset.SizeBytes,
		Alt:        asset.AltText,
		Caption:    asset.Caption,
		CreatedAt:  asset.CreatedAt.UTC(),
		BlobStatus: BlobStatusExcluded,
	}
}

func contentFromRouteRecord(record routeContentRecord, kind domaincontent.Kind, fallback time.Time) domaincontent.Entry {
	createdAt := parseRouteTime(record.Date, fallback)
	updatedAt := parseRouteTime(record.Modified, createdAt)
	status := domaincontent.Status(strings.TrimSpace(record.Status))
	if status == "" {
		status = domaincontent.StatusDraft
	}
	entry := domaincontent.Entry{
		ID:              domaincontent.ID(record.ID),
		Kind:            kind,
		Status:          status,
		Visibility:      domaincontent.VisibilityPrivate,
		Title:           domaincontent.LocalizedText{"en": record.Title.Rendered},
		Slug:            domaincontent.LocalizedText{"en": record.Slug},
		Body:            domaincontent.LocalizedText{"en": record.Content.Rendered},
		Excerpt:         domaincontent.LocalizedText{"en": record.Excerpt.Rendered},
		AuthorID:        record.Author,
		FeaturedMediaID: record.FeaturedMedia,
		CreatedAt:       createdAt,
		UpdatedAt:       updatedAt,
	}
	if status == domaincontent.StatusPublished {
		entry.Visibility = domaincontent.VisibilityPublic
		publishedAt := updatedAt
		entry.PublishedAt = &publishedAt
	}
	return entry
}

func mediaFromRouteRecord(record routeMediaRecord, metadata MediaMetadata, fallback time.Time) domainmedia.Asset {
	createdAt := metadata.CreatedAt
	if createdAt.IsZero() {
		createdAt = fallback.UTC()
	}
	return domainmedia.Asset{
		ID:        domainmedia.ID(record.ID),
		Filename:  firstNonEmpty(metadata.Filename, record.Title.Rendered),
		MimeType:  firstNonEmpty(metadata.MimeType, record.MimeType),
		SizeBytes: metadata.Size,
		Width:     metadata.Width,
		Height:    metadata.Height,
		AltText:   firstNonEmpty(metadata.Alt, record.AltText),
		Caption:   firstNonEmpty(metadata.Caption, record.Caption.Rendered),
		PublicURL: record.SourceURL,
		CreatedAt: createdAt,
		UpdatedAt: createdAt,
	}
}

func taxonomyDefinition(taxonomy string, label string) domaintaxonomy.Definition {
	return domaintaxonomy.Definition{
		Type:           domaintaxonomy.Type(taxonomy),
		Label:          label,
		Mode:           domaintaxonomy.ModeFlat,
		Public:         true,
		RESTVisible:    true,
		GraphQLVisible: true,
	}
}

func taxonomyTermFromRoute(record routeTermRecord, taxonomy string) domaintaxonomy.Term {
	return domaintaxonomy.Term{
		ID:          domaintaxonomy.TermID(record.ID),
		Type:        domaintaxonomy.Type(taxonomy),
		Name:        domaincontent.LocalizedText{"en": record.Name},
		Slug:        domaincontent.LocalizedText{"en": record.Slug},
		Description: domaincontent.LocalizedText{},
	}
}

func parseRouteTime(value string, fallback time.Time) time.Time {
	if strings.TrimSpace(value) == "" {
		return fallback.UTC()
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return fallback.UTC()
	}
	return parsed.UTC()
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
