package rest

import (
	"context"
	"net/http"
	"time"

	appmeta "github.com/fastygo/cms/internal/application/meta"
	domaincontent "github.com/fastygo/cms/internal/domain/content"
	domaincontenttype "github.com/fastygo/cms/internal/domain/contenttype"
	domainmedia "github.com/fastygo/cms/internal/domain/media"
	domainmenus "github.com/fastygo/cms/internal/domain/menus"
	domainsettings "github.com/fastygo/cms/internal/domain/settings"
	domaintaxonomy "github.com/fastygo/cms/internal/domain/taxonomy"
	domainusers "github.com/fastygo/cms/internal/domain/users"
	"github.com/fastygo/cms/internal/platform/plugins"
	"github.com/fastygo/framework/pkg/web"
)

type ResourceEnvelope struct {
	Data  any            `json:"data"`
	Links map[string]any `json:"links,omitempty"`
	Meta  map[string]any `json:"meta,omitempty"`
}

type ListEnvelope struct {
	Data       any            `json:"data"`
	Pagination Pagination     `json:"pagination"`
	Links      map[string]any `json:"links,omitempty"`
	Meta       map[string]any `json:"meta,omitempty"`
}

type Pagination struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

type ErrorEnvelope struct {
	Error APIError `json:"error"`
}

type APIError struct {
	Code      string         `json:"code"`
	Message   string         `json:"message"`
	Status    int            `json:"status"`
	Details   map[string]any `json:"details,omitempty"`
	RequestID string         `json:"request_id,omitempty"`
}

type Discovery struct {
	Name           string         `json:"name"`
	Version        string         `json:"version"`
	Routes         map[string]any `json:"routes"`
	Authentication []string       `json:"authentication"`
	Links          map[string]any `json:"links"`
}

type ContentDTO struct {
	ID              string            `json:"id"`
	Kind            string            `json:"kind"`
	Status          string            `json:"status"`
	Slug            map[string]string `json:"slug"`
	Title           map[string]string `json:"title"`
	Content         map[string]string `json:"content"`
	Excerpt         map[string]string `json:"excerpt"`
	AuthorID        string            `json:"author_id"`
	FeaturedMediaID string            `json:"featured_media_id,omitempty"`
	TaxonomyIDs     []string          `json:"taxonomy_ids"`
	Metadata        map[string]any    `json:"metadata,omitempty"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
	PublishedAt     *time.Time        `json:"published_at,omitempty"`
	Links           map[string]string `json:"links"`
}

type ContentTypeDTO struct {
	ID             string                     `json:"id"`
	Label          string                     `json:"label"`
	Public         bool                       `json:"public"`
	RESTVisible    bool                       `json:"rest_visible"`
	GraphQLVisible bool                       `json:"graphql_visible"`
	Supports       domaincontenttype.Supports `json:"supports"`
	Archive        bool                       `json:"archive"`
	Permalink      string                     `json:"permalink"`
}

type TaxonomyDTO struct {
	Type            string   `json:"type"`
	Label           string   `json:"label"`
	Mode            string   `json:"mode"`
	AssignedToKinds []string `json:"assigned_to_kinds"`
	Public          bool     `json:"public"`
	RESTVisible     bool     `json:"rest_visible"`
	GraphQLVisible  bool     `json:"graphql_visible"`
}

type TermDTO struct {
	ID          string            `json:"id"`
	Type        string            `json:"type"`
	Name        map[string]string `json:"name"`
	Slug        map[string]string `json:"slug"`
	Description map[string]string `json:"description"`
	ParentID    string            `json:"parent_id,omitempty"`
}

type MediaDTO struct {
	ID          string                `json:"id"`
	Filename    string                `json:"filename"`
	MimeType    string                `json:"mime_type"`
	SizeBytes   int64                 `json:"size_bytes"`
	Width       int                   `json:"width,omitempty"`
	Height      int                   `json:"height,omitempty"`
	AltText     string                `json:"alt_text,omitempty"`
	Caption     string                `json:"caption,omitempty"`
	PublicURL   string                `json:"public_url"`
	Provider    string                `json:"provider,omitempty"`
	ProviderURL string                `json:"provider_url,omitempty"`
	PublicMeta  map[string]any        `json:"metadata,omitempty"`
	Variants    []domainmedia.Variant `json:"variants,omitempty"`
}

type AuthorDTO struct {
	ID          string `json:"id"`
	Slug        string `json:"slug"`
	DisplayName string `json:"display_name"`
	Bio         string `json:"bio,omitempty"`
	AvatarURL   string `json:"avatar_url,omitempty"`
	WebsiteURL  string `json:"website_url,omitempty"`
}

type SettingDTO struct {
	Key   string `json:"key"`
	Value any    `json:"value"`
}

type MenuDTO struct {
	ID       string        `json:"id"`
	Name     string        `json:"name"`
	Location string        `json:"location"`
	Items    []MenuItemDTO `json:"items"`
}

type MenuItemDTO struct {
	ID       string        `json:"id"`
	Label    string        `json:"label"`
	URL      string        `json:"url"`
	Kind     string        `json:"kind,omitempty"`
	TargetID string        `json:"target_id,omitempty"`
	Children []MenuItemDTO `json:"children,omitempty"`
}

func WriteError(w http.ResponseWriter, status int, code string, message string, details map[string]any) {
	_ = web.WriteJSON(w, status, ErrorEnvelope{Error: APIError{
		Code:    code,
		Message: message,
		Status:  status,
		Details: details,
	}})
}

func ContentProjection(entry domaincontent.Entry, includePrivate bool) ContentDTO {
	projected, err := projectContent(context.Background(), nil, nil, entry, includePrivate, plugins.HookContext{})
	if err != nil {
		return baseContentProjection(entry, metadataValues(nil, entry, includePrivate))
	}
	return projected
}

func projectContent(ctx context.Context, registry *plugins.Registry, metaRegistry *appmeta.Registry, entry domaincontent.Entry, includePrivate bool, hookContext plugins.HookContext) (ContentDTO, error) {
	dto := baseContentProjection(entry, metadataValues(metaRegistry, entry, includePrivate))
	metadata, err := filterMetadata(ctx, registry, dto.Metadata, hookContext, metadataValues(metaRegistry, entry, includePrivate))
	if err != nil {
		return ContentDTO{}, err
	}
	dto.Metadata = metadata
	filtered, err := plugins.FilterValue(ctx, registry, "rest.content.filter", hookContext, dto)
	if err != nil {
		return ContentDTO{}, err
	}
	filtered.Metadata = sanitizeContentMetadata(filtered.Metadata, metaRegistry, entry, includePrivate)
	return filtered, nil
}

func baseContentProjection(entry domaincontent.Entry, values domaincontent.Metadata) ContentDTO {
	metadata := make(map[string]any)
	for key, value := range values {
		metadata[key] = value.Value
	}
	taxonomyIDs := make([]string, 0, len(entry.Terms))
	for _, ref := range entry.Terms {
		taxonomyIDs = append(taxonomyIDs, ref.Taxonomy+":"+ref.TermID)
	}
	return ContentDTO{
		ID:              string(entry.ID),
		Kind:            string(entry.Kind),
		Status:          string(entry.Status),
		Slug:            localizedMap(entry.Slug),
		Title:           localizedMap(entry.Title),
		Content:         localizedMap(entry.Body),
		Excerpt:         localizedMap(entry.Excerpt),
		AuthorID:        entry.AuthorID,
		FeaturedMediaID: entry.FeaturedMediaID,
		TaxonomyIDs:     taxonomyIDs,
		Metadata:        metadata,
		CreatedAt:       entry.CreatedAt,
		UpdatedAt:       entry.UpdatedAt,
		PublishedAt:     entry.PublishedAt,
		Links: map[string]string{
			"self": "/go-json/go/v2/" + string(entry.Kind) + "s/" + string(entry.ID),
		},
	}
}

func sanitizeContentMetadata(filtered map[string]any, metaRegistry *appmeta.Registry, entry domaincontent.Entry, includePrivate bool) map[string]any {
	if len(filtered) == 0 {
		return map[string]any{}
	}
	allowed := metadataValues(metaRegistry, entry, includePrivate)
	result := make(map[string]any, len(filtered))
	for key := range allowed {
		value, ok := filtered[key]
		if !ok {
			continue
		}
		result[key] = value
	}
	return result
}

func metadataValues(metaRegistry *appmeta.Registry, entry domaincontent.Entry, includePrivate bool) domaincontent.Metadata {
	if metaRegistry == nil {
		if includePrivate {
			return entry.Metadata
		}
		return entry.Metadata.Public()
	}
	return metaRegistry.PublicMetadata(entry.Kind, entry.Metadata, includePrivate)
}

func filterMetadata(ctx context.Context, registry *plugins.Registry, metadata map[string]any, hookContext plugins.HookContext, allowed domaincontent.Metadata) (map[string]any, error) {
	filtered, err := plugins.FilterValue(ctx, registry, "content.metadata.public.filter", hookContext, metadata)
	if err != nil {
		return nil, err
	}
	result := make(map[string]any, len(filtered))
	for key := range allowed {
		value, ok := filtered[key]
		if !ok {
			continue
		}
		result[key] = value
	}
	return result, nil
}

func ContentTypeProjection(contentType domaincontenttype.Type) ContentTypeDTO {
	return ContentTypeDTO{
		ID:             string(contentType.ID),
		Label:          contentType.Label,
		Public:         contentType.Public,
		RESTVisible:    contentType.RESTVisible,
		GraphQLVisible: contentType.GraphQLVisible,
		Supports:       contentType.Supports,
		Archive:        contentType.Archive,
		Permalink:      contentType.Permalink,
	}
}

func TaxonomyProjection(definition domaintaxonomy.Definition) TaxonomyDTO {
	kinds := make([]string, 0, len(definition.AssignedToKinds))
	for _, kind := range definition.AssignedToKinds {
		kinds = append(kinds, string(kind))
	}
	return TaxonomyDTO{
		Type:            string(definition.Type),
		Label:           definition.Label,
		Mode:            string(definition.Mode),
		AssignedToKinds: kinds,
		Public:          definition.Public,
		RESTVisible:     definition.RESTVisible,
		GraphQLVisible:  definition.GraphQLVisible,
	}
}

func TermProjection(term domaintaxonomy.Term) TermDTO {
	return TermDTO{
		ID:          string(term.ID),
		Type:        string(term.Type),
		Name:        localizedMap(term.Name),
		Slug:        localizedMap(term.Slug),
		Description: localizedMap(term.Description),
		ParentID:    string(term.ParentID),
	}
}

func MediaProjection(asset domainmedia.Asset) MediaDTO {
	return MediaDTO{
		ID:          string(asset.ID),
		Filename:    asset.Filename,
		MimeType:    asset.MimeType,
		SizeBytes:   asset.SizeBytes,
		Width:       asset.Width,
		Height:      asset.Height,
		AltText:     asset.AltText,
		Caption:     asset.Caption,
		PublicURL:   asset.PublicURL,
		Provider:    asset.ProviderRef.Provider,
		ProviderURL: asset.ProviderRef.URL,
		PublicMeta:  asset.PublicMeta,
		Variants:    asset.Variants,
	}
}

func AuthorProjection(author domainusers.AuthorProfile) AuthorDTO {
	return AuthorDTO{
		ID:          string(author.ID),
		Slug:        author.Slug,
		DisplayName: author.DisplayName,
		Bio:         author.Bio,
		AvatarURL:   author.AvatarURL,
		WebsiteURL:  author.WebsiteURL,
	}
}

func SettingProjection(value domainsettings.Value) SettingDTO {
	return SettingDTO{Key: string(value.Key), Value: value.Value}
}

func MenuProjection(menu domainmenus.Menu) MenuDTO {
	return MenuDTO{
		ID:       string(menu.ID),
		Name:     menu.Name,
		Location: string(menu.Location),
		Items:    menuItems(menu.Items),
	}
}

func menuItems(items []domainmenus.Item) []MenuItemDTO {
	result := make([]MenuItemDTO, 0, len(items))
	for _, item := range items {
		result = append(result, MenuItemDTO{
			ID:       string(item.ID),
			Label:    item.Label,
			URL:      item.URL,
			Kind:     item.Kind,
			TargetID: item.TargetID,
			Children: menuItems(item.Children),
		})
	}
	return result
}

func localizedMap(values domaincontent.LocalizedText) map[string]string {
	result := make(map[string]string, len(values))
	for key, value := range values {
		result[key] = value
	}
	return result
}
