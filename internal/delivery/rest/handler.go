package rest

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	appaudit "github.com/fastygo/cms/internal/application/audit"
	appcontent "github.com/fastygo/cms/internal/application/content"
	appcontenttype "github.com/fastygo/cms/internal/application/contenttype"
	appmedia "github.com/fastygo/cms/internal/application/media"
	appmenus "github.com/fastygo/cms/internal/application/menus"
	appmeta "github.com/fastygo/cms/internal/application/meta"
	appsettings "github.com/fastygo/cms/internal/application/settings"
	apptaxonomy "github.com/fastygo/cms/internal/application/taxonomy"
	appusers "github.com/fastygo/cms/internal/application/users"
	domainaudit "github.com/fastygo/cms/internal/domain/audit"
	domainauthz "github.com/fastygo/cms/internal/domain/authz"
	domaincontent "github.com/fastygo/cms/internal/domain/content"
	domaincontenttype "github.com/fastygo/cms/internal/domain/contenttype"
	domainmedia "github.com/fastygo/cms/internal/domain/media"
	domainmenus "github.com/fastygo/cms/internal/domain/menus"
	domaintaxonomy "github.com/fastygo/cms/internal/domain/taxonomy"
	domainusers "github.com/fastygo/cms/internal/domain/users"
	platformplugins "github.com/fastygo/cms/internal/platform/plugins"
	"github.com/fastygo/framework/pkg/web"
)

type Services struct {
	Content      appcontent.Service
	ContentTypes appcontenttype.Service
	Taxonomy     apptaxonomy.Service
	Media        appmedia.Service
	Users        appusers.Service
	Settings     appsettings.Service
	Menus        appmenus.Service
	Audit        appaudit.Service
}

type Handler struct {
	services Services
	auth     Authenticator
	registry *platformplugins.Registry
	meta     *appmeta.Registry
}

func NewHandler(services Services, auth Authenticator) Handler {
	return NewHandlerWithOptions(services, auth, nil, nil)
}

func NewHandlerWithRegistry(services Services, auth Authenticator, registry *platformplugins.Registry) Handler {
	return NewHandlerWithOptions(services, auth, registry, nil)
}

func NewHandlerWithOptions(services Services, auth Authenticator, registry *platformplugins.Registry, metaRegistry *appmeta.Registry) Handler {
	return Handler{services: services, auth: auth, registry: registry, meta: metaRegistry}
}

func (h Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /go-json", h.rootDiscovery)
	mux.HandleFunc("GET /go-json/go/v2/", h.namespaceDiscovery)
	mux.HandleFunc("GET /go-json/go/v2/posts", h.listContent(domaincontent.KindPost))
	mux.HandleFunc("POST /go-json/go/v2/posts", h.createContent(domaincontent.KindPost))
	mux.HandleFunc("GET /go-json/go/v2/posts/{id}", h.getContentByID)
	mux.HandleFunc("PATCH /go-json/go/v2/posts/{id}", h.updateContent)
	mux.HandleFunc("DELETE /go-json/go/v2/posts/{id}", h.trashContent)
	mux.HandleFunc("GET /go-json/go/v2/posts/by-slug/{slug}", h.getContentBySlug(domaincontent.KindPost))
	mux.HandleFunc("GET /go-json/go/v2/pages", h.listContent(domaincontent.KindPage))
	mux.HandleFunc("POST /go-json/go/v2/pages", h.createContent(domaincontent.KindPage))
	mux.HandleFunc("GET /go-json/go/v2/pages/{id}", h.getContentByID)
	mux.HandleFunc("PATCH /go-json/go/v2/pages/{id}", h.updateContent)
	mux.HandleFunc("DELETE /go-json/go/v2/pages/{id}", h.trashContent)
	mux.HandleFunc("GET /go-json/go/v2/pages/by-slug/{slug}", h.getContentBySlug(domaincontent.KindPage))
	mux.HandleFunc("GET /go-json/go/v2/content-types", h.listContentTypes)
	mux.HandleFunc("POST /go-json/go/v2/content-types", h.registerContentType)
	mux.HandleFunc("GET /go-json/go/v2/media", h.listMedia)
	mux.HandleFunc("POST /go-json/go/v2/media", h.saveMedia)
	mux.HandleFunc("GET /go-json/go/v2/media/{id}", h.getMedia)
	mux.HandleFunc("PATCH /go-json/go/v2/media/{id}", h.saveMedia)
	mux.HandleFunc("POST /go-json/go/v2/media/{id}/featured/{content_id}", h.attachFeaturedMedia)
	mux.HandleFunc("GET /go-json/go/v2/taxonomies", h.listTaxonomies)
	mux.HandleFunc("POST /go-json/go/v2/taxonomies", h.registerTaxonomy)
	mux.HandleFunc("GET /go-json/go/v2/taxonomies/{type}", h.listTerms)
	mux.HandleFunc("GET /go-json/go/v2/taxonomies/{type}/{id}", h.getTerm)
	mux.HandleFunc("POST /go-json/go/v2/taxonomies/{type}/terms", h.createTerm)
	mux.HandleFunc("POST /go-json/go/v2/content/{id}/terms", h.assignTerms)
	mux.HandleFunc("GET /go-json/go/v2/menus", h.listMenus)
	mux.HandleFunc("GET /go-json/go/v2/menus/{location}", h.getMenu)
	mux.HandleFunc("GET /go-json/go/v2/settings", h.listSettings)
	mux.HandleFunc("GET /go-json/go/v2/authors/{id}", h.getAuthor)
	mux.HandleFunc("GET /go-json/go/v2/search", h.search)
}

func (h Handler) rootDiscovery(w http.ResponseWriter, _ *http.Request) {
	_ = web.WriteJSON(w, http.StatusOK, Discovery{
		Name:           "GoCMS",
		Version:        "2",
		Routes:         map[string]any{"go/v2": "/go-json/go/v2/"},
		Authentication: []string{"browser_session", "app_token", "dev_bearer"},
		Links:          map[string]any{"self": "/go-json"},
	})
}

func (h Handler) namespaceDiscovery(w http.ResponseWriter, _ *http.Request) {
	routes := map[string]any{
		"posts":        "/go-json/go/v2/posts",
		"pages":        "/go-json/go/v2/pages",
		"media":        "/go-json/go/v2/media",
		"taxonomies":   "/go-json/go/v2/taxonomies",
		"menus":        "/go-json/go/v2/menus",
		"settings":     "/go-json/go/v2/settings",
		"search":       "/go-json/go/v2/search",
		"contentTypes": "/go-json/go/v2/content-types",
	}
	_ = web.WriteJSON(w, http.StatusOK, Discovery{
		Name:           "GoCMS",
		Version:        "2",
		Routes:         routes,
		Authentication: []string{"browser_session", "app_token", "dev_bearer"},
		Links:          map[string]any{"self": "/go-json/go/v2/"},
	})
}

func (h Handler) listContent(kind domaincontent.Kind) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, authenticated := h.auth.Principal(r)
		includePrivate := authenticated && principal.Has(domainauthz.CapabilityContentReadPrivate)
		query, err := parseContentQuery(r, kind, !includePrivate)
		if err != nil {
			WriteError(w, http.StatusBadRequest, "validation_error", err.Error(), nil)
			return
		}
		result, err := h.services.Content.List(r.Context(), query)
		if err != nil {
			WriteError(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
			return
		}
		data := make([]ContentDTO, 0, len(result.Items))
		for _, entry := range result.Items {
			projected, err := h.projectContent(r, principal, authenticated, entry, includePrivate)
			if err != nil {
				WriteError(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
				return
			}
			data = append(data, projected)
		}
		_ = web.WriteJSON(w, http.StatusOK, ListEnvelope{
			Data:       data,
			Pagination: Pagination{Page: result.Page, PerPage: result.PerPage, Total: result.Total, TotalPages: result.TotalPages},
		})
	}
}

func (h Handler) getContentByID(w http.ResponseWriter, r *http.Request) {
	principal, authenticated := h.auth.Principal(r)
	entry, err := h.services.Content.Get(r.Context(), principal, domaincontent.ID(r.PathValue("id")))
	if err != nil {
		WriteError(w, http.StatusNotFound, "not_found", err.Error(), nil)
		return
	}
	projected, err := h.projectContent(r, principal, authenticated, entry, authenticated && principal.Has(domainauthz.CapabilityContentReadPrivate))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
		return
	}
	_ = web.WriteJSON(w, http.StatusOK, ResourceEnvelope{Data: projected})
}

func (h Handler) getContentBySlug(kind domaincontent.Kind) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, authenticated := h.auth.Principal(r)
		entry, err := h.services.Content.GetBySlug(r.Context(), principal, kind, r.PathValue("slug"), r.URL.Query().Get("locale"))
		if err != nil {
			WriteError(w, http.StatusNotFound, "not_found", err.Error(), nil)
			return
		}
		projected, err := h.projectContent(r, principal, authenticated, entry, authenticated && principal.Has(domainauthz.CapabilityContentReadPrivate))
		if err != nil {
			WriteError(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
			return
		}
		_ = web.WriteJSON(w, http.StatusOK, ResourceEnvelope{Data: projected})
	}
}

func (h Handler) createContent(kind domaincontent.Kind) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := h.requirePrincipal(w, r)
		if !ok {
			return
		}
		var request contentWriteRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			WriteError(w, http.StatusBadRequest, "validation_error", "invalid JSON body", nil)
			return
		}
		entry, err := h.services.Content.CreateDraft(r.Context(), principal, appcontent.CreateDraftCommand{
			Kind:            kind,
			Title:           request.Title,
			Slug:            request.Slug,
			Body:            request.Content,
			Excerpt:         request.Excerpt,
			AuthorID:        request.AuthorID,
			FeaturedMediaID: request.FeaturedMediaID,
			Metadata:        metadataFromRequest(request.Metadata, true),
			Terms:           request.Terms,
		})
		if err == nil {
			entry, err = h.applyRequestedStatus(r, principal, entry.ID, request)
		}
		if err != nil {
			WriteError(w, statusFromError(err), errorCodeFromStatus(statusFromError(err)), err.Error(), nil)
			return
		}
		projected, err := h.projectContent(r, principal, true, entry, true)
		if err != nil {
			WriteError(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
			return
		}
		h.audit(r, principal, "api.content.create", "content", string(entry.ID), domainaudit.StatusSuccess, map[string]any{"kind": entry.Kind, "status": entry.Status})
		_ = web.WriteJSON(w, http.StatusCreated, ResourceEnvelope{Data: projected})
	}
}

func (h Handler) updateContent(w http.ResponseWriter, r *http.Request) {
	principal, ok := h.requirePrincipal(w, r)
	if !ok {
		return
	}
	var request contentWriteRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		WriteError(w, http.StatusBadRequest, "validation_error", "invalid JSON body", nil)
		return
	}
	entry, err := h.services.Content.Update(r.Context(), principal, appcontent.UpdateCommand{
		ID:              domaincontent.ID(r.PathValue("id")),
		Title:           request.Title,
		Slug:            request.Slug,
		Body:            request.Content,
		Excerpt:         request.Excerpt,
		AuthorID:        request.AuthorID,
		FeaturedMediaID: request.FeaturedMediaID,
		Metadata:        metadataFromRequest(request.Metadata, true),
		Terms:           request.Terms,
	})
	if err == nil {
		entry, err = h.applyRequestedStatus(r, principal, entry.ID, request)
	}
	if err != nil {
		WriteError(w, statusFromError(err), errorCodeFromStatus(statusFromError(err)), err.Error(), nil)
		return
	}
	projected, err := h.projectContent(r, principal, true, entry, true)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
		return
	}
	h.audit(r, principal, "api.content.update", "content", string(entry.ID), domainaudit.StatusSuccess, map[string]any{"kind": entry.Kind, "status": entry.Status})
	_ = web.WriteJSON(w, http.StatusOK, ResourceEnvelope{Data: projected})
}

func (h Handler) trashContent(w http.ResponseWriter, r *http.Request) {
	principal, ok := h.requirePrincipal(w, r)
	if !ok {
		return
	}
	entry, err := h.services.Content.Trash(r.Context(), principal, domaincontent.ID(r.PathValue("id")))
	if err != nil {
		WriteError(w, statusFromError(err), errorCodeFromStatus(statusFromError(err)), err.Error(), nil)
		return
	}
	projected, err := h.projectContent(r, principal, true, entry, true)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
		return
	}
	h.audit(r, principal, "api.content.trash", "content", string(entry.ID), domainaudit.StatusSuccess, map[string]any{"kind": entry.Kind})
	_ = web.WriteJSON(w, http.StatusOK, ResourceEnvelope{Data: projected})
}

func (h Handler) applyRequestedStatus(r *http.Request, principal domainauthz.Principal, id domaincontent.ID, request contentWriteRequest) (domaincontent.Entry, error) {
	switch request.Status {
	case "", domaincontent.StatusDraft:
		return h.services.Content.Get(r.Context(), principal, id)
	case domaincontent.StatusPublished:
		return h.services.Content.Publish(r.Context(), principal, id)
	case domaincontent.StatusScheduled:
		if request.PublishedAt == nil {
			return domaincontent.Entry{}, errors.New("published_at is required for scheduled content")
		}
		return h.services.Content.Schedule(r.Context(), principal, id, *request.PublishedAt)
	case domaincontent.StatusTrashed:
		return h.services.Content.Trash(r.Context(), principal, id)
	default:
		return domaincontent.Entry{}, errors.New("unsupported content status")
	}
}

func (h Handler) listContentTypes(w http.ResponseWriter, r *http.Request) {
	items, err := h.services.ContentTypes.List(r.Context())
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
		return
	}
	data := make([]ContentTypeDTO, 0, len(items))
	for _, item := range items {
		data = append(data, ContentTypeProjection(item))
	}
	_ = web.WriteJSON(w, http.StatusOK, ResourceEnvelope{Data: data})
}

func (h Handler) registerContentType(w http.ResponseWriter, r *http.Request) {
	principal, ok := h.requirePrincipal(w, r)
	if !ok {
		return
	}
	if !principal.Has(domainauthz.CapabilitySettingsManage) {
		WriteError(w, http.StatusForbidden, "forbidden", "settings management capability is required", nil)
		return
	}
	var request contentTypeRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		WriteError(w, http.StatusBadRequest, "validation_error", "invalid JSON body", nil)
		return
	}
	contentType := domaincontenttype.Type{
		ID:             domaincontent.Kind(request.ID),
		Label:          request.Label,
		Public:         request.Public,
		RESTVisible:    request.RESTVisible,
		GraphQLVisible: request.GraphQLVisible,
		Supports:       request.Supports,
		Archive:        request.Archive,
		Permalink:      request.Permalink,
	}
	if err := h.services.ContentTypes.Register(r.Context(), contentType); err != nil {
		WriteError(w, http.StatusBadRequest, "validation_error", err.Error(), nil)
		return
	}
	h.audit(r, principal, "api.content_type.register", "content_type", string(contentType.ID), domainaudit.StatusSuccess, nil)
	_ = web.WriteJSON(w, http.StatusCreated, ResourceEnvelope{Data: ContentTypeProjection(contentType)})
}

func (h Handler) listTaxonomies(w http.ResponseWriter, r *http.Request) {
	items, err := h.services.Taxonomy.ListDefinitions(r.Context())
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
		return
	}
	data := make([]TaxonomyDTO, 0, len(items))
	for _, item := range items {
		data = append(data, TaxonomyProjection(item))
	}
	_ = web.WriteJSON(w, http.StatusOK, ResourceEnvelope{Data: data})
}

func (h Handler) registerTaxonomy(w http.ResponseWriter, r *http.Request) {
	principal, ok := h.requirePrincipal(w, r)
	if !ok {
		return
	}
	var request taxonomyRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		WriteError(w, http.StatusBadRequest, "validation_error", "invalid JSON body", nil)
		return
	}
	definition := domaintaxonomy.Definition{
		Type:            domaintaxonomy.Type(request.Type),
		Label:           request.Label,
		Mode:            domaintaxonomy.Mode(request.Mode),
		AssignedToKinds: request.AssignedToKinds,
		Public:          request.Public,
		RESTVisible:     request.RESTVisible,
		GraphQLVisible:  request.GraphQLVisible,
	}
	if err := h.services.Taxonomy.Register(r.Context(), principal, definition); err != nil {
		WriteError(w, statusFromError(err), errorCodeFromStatus(statusFromError(err)), err.Error(), nil)
		return
	}
	h.audit(r, principal, "api.taxonomy.register", "taxonomy", string(definition.Type), domainaudit.StatusSuccess, nil)
	_ = web.WriteJSON(w, http.StatusCreated, ResourceEnvelope{Data: TaxonomyProjection(definition)})
}

func (h Handler) listTerms(w http.ResponseWriter, r *http.Request) {
	terms, err := h.services.Taxonomy.ListTerms(r.Context(), domaintaxonomy.Type(r.PathValue("type")))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
		return
	}
	data := make([]TermDTO, 0, len(terms))
	for _, term := range terms {
		data = append(data, TermProjection(term))
	}
	_ = web.WriteJSON(w, http.StatusOK, ResourceEnvelope{Data: data})
}

func (h Handler) getTerm(w http.ResponseWriter, r *http.Request) {
	term, ok, err := h.services.Taxonomy.GetTerm(r.Context(), domaintaxonomy.TermID(r.PathValue("id")))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
		return
	}
	if !ok || string(term.Type) != r.PathValue("type") {
		WriteError(w, http.StatusNotFound, "not_found", "term not found", nil)
		return
	}
	_ = web.WriteJSON(w, http.StatusOK, ResourceEnvelope{Data: TermProjection(term)})
}

func (h Handler) createTerm(w http.ResponseWriter, r *http.Request) {
	principal, ok := h.requirePrincipal(w, r)
	if !ok {
		return
	}
	var request termRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		WriteError(w, http.StatusBadRequest, "validation_error", "invalid JSON body", nil)
		return
	}
	term := domaintaxonomy.Term{
		ID:          domaintaxonomy.TermID(request.ID),
		Type:        domaintaxonomy.Type(r.PathValue("type")),
		Name:        request.Name,
		Slug:        request.Slug,
		Description: request.Description,
		ParentID:    domaintaxonomy.TermID(request.ParentID),
	}
	if err := h.services.Taxonomy.CreateTerm(r.Context(), principal, term); err != nil {
		WriteError(w, statusFromError(err), errorCodeFromStatus(statusFromError(err)), err.Error(), nil)
		return
	}
	h.audit(r, principal, "api.term.create", "term", string(term.ID), domainaudit.StatusSuccess, map[string]any{"taxonomy": term.Type})
	_ = web.WriteJSON(w, http.StatusCreated, ResourceEnvelope{Data: TermProjection(term)})
}

func (h Handler) assignTerms(w http.ResponseWriter, r *http.Request) {
	principal, ok := h.requirePrincipal(w, r)
	if !ok {
		return
	}
	var request assignTermsRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		WriteError(w, http.StatusBadRequest, "validation_error", "invalid JSON body", nil)
		return
	}
	entry, err := h.services.Taxonomy.AssignTerms(r.Context(), principal, domaincontent.ID(r.PathValue("id")), request.Terms)
	if err != nil {
		WriteError(w, statusFromError(err), errorCodeFromStatus(statusFromError(err)), err.Error(), nil)
		return
	}
	projected, err := h.projectContent(r, principal, true, entry, true)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
		return
	}
	h.audit(r, principal, "api.terms.assign", "content", string(entry.ID), domainaudit.StatusSuccess, map[string]any{"terms": len(request.Terms)})
	_ = web.WriteJSON(w, http.StatusOK, ResourceEnvelope{Data: projected})
}

func (h Handler) listMedia(w http.ResponseWriter, r *http.Request) {
	items, err := h.services.Media.List(r.Context())
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
		return
	}
	data := make([]MediaDTO, 0, len(items))
	for _, item := range items {
		data = append(data, MediaProjection(item))
	}
	_ = web.WriteJSON(w, http.StatusOK, ResourceEnvelope{Data: data})
}

func (h Handler) getMedia(w http.ResponseWriter, r *http.Request) {
	asset, ok, err := h.services.Media.Get(r.Context(), domainmedia.ID(r.PathValue("id")))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
		return
	}
	if !ok {
		WriteError(w, http.StatusNotFound, "not_found", "media asset not found", nil)
		return
	}
	_ = web.WriteJSON(w, http.StatusOK, ResourceEnvelope{Data: MediaProjection(asset)})
}

func (h Handler) saveMedia(w http.ResponseWriter, r *http.Request) {
	principal, ok := h.requirePrincipal(w, r)
	if !ok {
		return
	}
	var asset domainmedia.Asset
	if err := json.NewDecoder(r.Body).Decode(&asset); err != nil {
		WriteError(w, http.StatusBadRequest, "validation_error", "invalid JSON body", nil)
		return
	}
	if pathID := r.PathValue("id"); pathID != "" {
		asset.ID = domainmedia.ID(pathID)
	}
	if err := h.services.Media.SaveMetadata(r.Context(), principal, asset); err != nil {
		WriteError(w, statusFromError(err), errorCodeFromStatus(statusFromError(err)), err.Error(), nil)
		return
	}
	h.audit(r, principal, "api.media.save", "media", string(asset.ID), domainaudit.StatusSuccess, map[string]any{"mime_type": asset.MimeType})
	_ = web.WriteJSON(w, http.StatusOK, ResourceEnvelope{Data: MediaProjection(asset)})
}

func (h Handler) attachFeaturedMedia(w http.ResponseWriter, r *http.Request) {
	principal, ok := h.requirePrincipal(w, r)
	if !ok {
		return
	}
	entry, err := h.services.Media.AttachFeatured(r.Context(), principal, domaincontent.ID(r.PathValue("content_id")), domainmedia.ID(r.PathValue("id")))
	if err != nil {
		WriteError(w, statusFromError(err), errorCodeFromStatus(statusFromError(err)), err.Error(), nil)
		return
	}
	_ = web.WriteJSON(w, http.StatusOK, ResourceEnvelope{Data: ContentProjection(entry, true)})
}

func (h Handler) listMenus(w http.ResponseWriter, r *http.Request) {
	items, err := h.services.Menus.List(r.Context())
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
		return
	}
	data := make([]MenuDTO, 0, len(items))
	for _, item := range items {
		data = append(data, MenuProjection(item))
	}
	_ = web.WriteJSON(w, http.StatusOK, ResourceEnvelope{Data: data})
}

func (h Handler) getMenu(w http.ResponseWriter, r *http.Request) {
	menu, ok, err := h.services.Menus.ByLocation(r.Context(), domainmenus.Location(r.PathValue("location")))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
		return
	}
	if !ok {
		WriteError(w, http.StatusNotFound, "not_found", "menu not found", nil)
		return
	}
	_ = web.WriteJSON(w, http.StatusOK, ResourceEnvelope{Data: MenuProjection(menu)})
}

func (h Handler) listSettings(w http.ResponseWriter, r *http.Request) {
	items, err := h.services.Settings.Public(r.Context())
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
		return
	}
	data := make([]SettingDTO, 0, len(items))
	for _, item := range items {
		data = append(data, SettingProjection(item))
	}
	_ = web.WriteJSON(w, http.StatusOK, ResourceEnvelope{Data: data})
}

func (h Handler) getAuthor(w http.ResponseWriter, r *http.Request) {
	author, ok, err := h.services.Users.PublicAuthor(r.Context(), domainusers.ID(r.PathValue("id")))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
		return
	}
	if !ok {
		WriteError(w, http.StatusNotFound, "not_found", "author not found", nil)
		return
	}
	dto := AuthorProjection(author)
	if mid := strings.TrimSpace(author.AvatarMediaID); mid != "" {
		if asset, found, gerr := h.services.Media.Get(r.Context(), domainmedia.ID(mid)); gerr == nil && found {
			dto.AvatarURL = asset.PublicURL
		}
	}
	_ = web.WriteJSON(w, http.StatusOK, ResourceEnvelope{Data: dto})
}

func (h Handler) search(w http.ResponseWriter, r *http.Request) {
	query, err := parseContentQuery(r, "", true)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "validation_error", err.Error(), nil)
		return
	}
	query.Search = r.URL.Query().Get("search")
	if query.Search == "" {
		query.Search = r.URL.Query().Get("q")
	}
	result, err := h.services.Content.List(r.Context(), query)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
		return
	}
	data := make([]ContentDTO, 0, len(result.Items))
	for _, entry := range result.Items {
		projected, err := h.projectContent(r, domainauthz.Principal{}, false, entry, false)
		if err != nil {
			WriteError(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
			return
		}
		data = append(data, projected)
	}
	_ = web.WriteJSON(w, http.StatusOK, ListEnvelope{
		Data:       data,
		Pagination: Pagination{Page: result.Page, PerPage: result.PerPage, Total: result.Total, TotalPages: result.TotalPages},
	})
}

func (h Handler) requirePrincipal(w http.ResponseWriter, r *http.Request) (domainauthz.Principal, bool) {
	principal, ok := h.auth.Principal(r)
	if !ok {
		WriteError(w, http.StatusUnauthorized, "unauthorized", "authentication is required", nil)
		return domainauthz.Principal{}, false
	}
	return principal, true
}

func (h Handler) audit(r *http.Request, principal domainauthz.Principal, action string, resource string, resourceID string, status domainaudit.Status, details map[string]any) {
	if !h.services.Audit.Enabled() {
		return
	}
	_ = h.services.Audit.Record(r.Context(), domainaudit.Event{
		ActorID:    principal.ID,
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
		Status:     status,
		RemoteAddr: r.RemoteAddr,
		UserAgent:  r.UserAgent(),
		Details:    details,
	})
}

func (h Handler) projectContent(r *http.Request, principal domainauthz.Principal, authenticated bool, entry domaincontent.Entry, includePrivate bool) (ContentDTO, error) {
	return projectContent(r.Context(), h.registry, h.meta, entry, includePrivate, platformplugins.HookContext{
		Surface:       platformplugins.SurfaceREST,
		Path:          r.URL.Path,
		Locale:        r.URL.Query().Get("locale"),
		Principal:     principal,
		Authenticated: authenticated,
		Metadata: map[string]any{
			"resource":        "content",
			"include_private": includePrivate,
			"content_id":      string(entry.ID),
		},
	})
}

func parseContentQuery(r *http.Request, kind domaincontent.Kind, publicOnly bool) (domaincontent.Query, error) {
	values := r.URL.Query()
	page, err := parsePositiveInt(values.Get("page"), 1, "page")
	if err != nil {
		return domaincontent.Query{}, err
	}
	perPage, err := parsePositiveInt(values.Get("per_page"), 20, "per_page")
	if err != nil {
		return domaincontent.Query{}, err
	}
	query := domaincontent.Query{
		PublicOnly: publicOnly,
		Locale:     values.Get("locale"),
		Slug:       values.Get("slug"),
		AuthorID:   values.Get("author"),
		Search:     values.Get("search"),
		Page:       page,
		PerPage:    perPage,
		SortBy:     domaincontent.SortField(values.Get("sort")),
		SortDesc:   strings.ToLower(values.Get("order")) == "desc",
	}
	if kind != "" {
		query.Kinds = []domaincontent.Kind{kind}
	}
	if raw := values.Get("kind"); raw != "" && kind == "" {
		for _, part := range strings.Split(raw, ",") {
			query.Kinds = append(query.Kinds, domaincontent.Kind(strings.TrimSpace(part)))
		}
	}
	if raw := values.Get("status"); raw != "" {
		for _, part := range strings.Split(raw, ",") {
			status := domaincontent.Status(strings.TrimSpace(part))
			if err := domaincontent.ValidateStatus(status); err != nil {
				return domaincontent.Query{}, err
			}
			query.Statuses = append(query.Statuses, status)
		}
	}
	if raw := values.Get("taxonomy"); raw != "" {
		parts := strings.SplitN(raw, ":", 2)
		query.Taxonomy = parts[0]
		if len(parts) == 2 {
			query.TermID = parts[1]
		}
	}
	if raw := values.Get("term"); raw != "" {
		query.TermID = raw
	}
	if raw := values.Get("after"); raw != "" {
		parsed, err := parseTime(raw)
		if err != nil {
			return domaincontent.Query{}, err
		}
		query.After = &parsed
	}
	if raw := values.Get("before"); raw != "" {
		parsed, err := parseTime(raw)
		if err != nil {
			return domaincontent.Query{}, err
		}
		query.Before = &parsed
	}
	return query, nil
}

func parseTime(value string) (time.Time, error) {
	if parsed, err := time.Parse(time.RFC3339, value); err == nil {
		return parsed, nil
	}
	return time.Parse("2006-01-02", value)
}

func parsePositiveInt(value string, fallback int, field string) (int, error) {
	if strings.TrimSpace(value) == "" {
		return fallback, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return 0, fmt.Errorf("%s must be a positive integer", field)
	}
	return parsed, nil
}

func metadataFromRequest(values map[string]any, public bool) domaincontent.Metadata {
	if len(values) == 0 {
		return nil
	}
	metadata := make(domaincontent.Metadata, len(values))
	for key, value := range values {
		metadata[key] = domaincontent.MetaValue{Value: value, Public: public}
	}
	return metadata
}

func statusFromError(err error) int {
	if err == nil {
		return http.StatusOK
	}
	message := strings.ToLower(err.Error())
	switch {
	case strings.Contains(message, "capability"):
		return http.StatusForbidden
	case strings.Contains(message, "not found"):
		return http.StatusNotFound
	case strings.Contains(message, "required"), strings.Contains(message, "unsupported"), strings.Contains(message, "invalid"):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

func errorCodeFromStatus(status int) string {
	switch status {
	case http.StatusUnauthorized:
		return "unauthorized"
	case http.StatusForbidden:
		return "forbidden"
	case http.StatusNotFound:
		return "not_found"
	case http.StatusBadRequest:
		return "validation_error"
	default:
		return "internal_error"
	}
}

type contentWriteRequest struct {
	Status          domaincontent.Status        `json:"status"`
	Title           domaincontent.LocalizedText `json:"title"`
	Slug            domaincontent.LocalizedText `json:"slug"`
	Content         domaincontent.LocalizedText `json:"content"`
	Excerpt         domaincontent.LocalizedText `json:"excerpt"`
	AuthorID        string                      `json:"author_id"`
	FeaturedMediaID string                      `json:"featured_media_id"`
	Metadata        map[string]any              `json:"metadata"`
	Terms           []domaincontent.TermRef     `json:"terms"`
	PublishedAt     *time.Time                  `json:"published_at"`
}

type contentTypeRequest struct {
	ID             string                     `json:"id"`
	Label          string                     `json:"label"`
	Public         bool                       `json:"public"`
	RESTVisible    bool                       `json:"rest_visible"`
	GraphQLVisible bool                       `json:"graphql_visible"`
	Supports       domaincontenttype.Supports `json:"supports"`
	Archive        bool                       `json:"archive"`
	Permalink      string                     `json:"permalink"`
}

type taxonomyRequest struct {
	Type            string               `json:"type"`
	Label           string               `json:"label"`
	Mode            string               `json:"mode"`
	AssignedToKinds []domaincontent.Kind `json:"assigned_to_kinds"`
	Public          bool                 `json:"public"`
	RESTVisible     bool                 `json:"rest_visible"`
	GraphQLVisible  bool                 `json:"graphql_visible"`
}

type termRequest struct {
	ID          string                      `json:"id"`
	Name        domaincontent.LocalizedText `json:"name"`
	Slug        domaincontent.LocalizedText `json:"slug"`
	Description domaincontent.LocalizedText `json:"description"`
	ParentID    string                      `json:"parent_id"`
}

type assignTermsRequest struct {
	Terms []domaincontent.TermRef `json:"terms"`
}
