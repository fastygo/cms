package rest_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	restpkg "github.com/fastygo/cms/internal/delivery/rest"
	domaincontent "github.com/fastygo/cms/internal/domain/content"
	domainmeta "github.com/fastygo/cms/internal/domain/meta"
	"github.com/fastygo/cms/internal/infra/features/cms"
	platformplugins "github.com/fastygo/cms/internal/platform/plugins"
)

func TestRESTDiscoveryPublicReadsAndWrites(t *testing.T) {
	module, err := cms.New("file:rest-test?mode=memory&cache=shared", "test-session-secret-test-session-secret", true)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := module.Close(t.Context()); err != nil {
			t.Fatal(err)
		}
	}()
	mux := http.NewServeMux()
	module.Routes(mux)

	rec := request(mux, http.MethodGet, "/go-json", "", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("discovery status = %d body = %s", rec.Code, rec.Body.String())
	}

	rec = request(mux, http.MethodGet, "/go-json/go/v2/posts?per_page=1", "", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("posts status = %d body = %s", rec.Code, rec.Body.String())
	}
	var list struct {
		Data       []map[string]any `json:"data"`
		Pagination struct {
			Page       int `json:"page"`
			PerPage    int `json:"per_page"`
			Total      int `json:"total"`
			TotalPages int `json:"total_pages"`
		} `json:"pagination"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &list); err != nil {
		t.Fatal(err)
	}
	if len(list.Data) != 1 || list.Pagination.PerPage != 1 || list.Pagination.Total != 1 {
		t.Fatalf("unexpected public list: %+v", list)
	}
	metadata, _ := list.Data[0]["metadata"].(map[string]any)
	if _, ok := metadata["private_key"]; ok {
		t.Fatalf("private metadata leaked: %+v", metadata)
	}

	rec = request(mux, http.MethodGet, "/go-json/go/v2/posts/by-slug/published-post", "", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("slug detail status = %d body = %s", rec.Code, rec.Body.String())
	}

	body := `{"title":{"en":"Created Through REST"},"slug":{"en":"created-through-rest"},"content":{"en":"Body"},"status":"published"}`
	rec = request(mux, http.MethodPost, "/go-json/go/v2/posts", body, "Bearer admin-token")
	if rec.Code != http.StatusCreated {
		t.Fatalf("create status = %d body = %s", rec.Code, rec.Body.String())
	}

	rec = request(mux, http.MethodPost, "/go-json/go/v2/posts", body, "Bearer viewer-token")
	if rec.Code != http.StatusForbidden {
		t.Fatalf("low privilege create status = %d body = %s", rec.Code, rec.Body.String())
	}

	taxonomyBody := `{"type":"genre","label":"Genres","mode":"flat","assigned_to_kinds":["post"],"public":true,"rest_visible":true}`
	rec = request(mux, http.MethodPost, "/go-json/go/v2/taxonomies", taxonomyBody, "Bearer admin-token")
	if rec.Code != http.StatusCreated {
		t.Fatalf("taxonomy create status = %d body = %s", rec.Code, rec.Body.String())
	}
	termBody := `{"id":"term-review","name":{"en":"Review"},"slug":{"en":"review"}}`
	rec = request(mux, http.MethodPost, "/go-json/go/v2/taxonomies/genre/terms", termBody, "Bearer admin-token")
	if rec.Code != http.StatusCreated {
		t.Fatalf("term create status = %d body = %s", rec.Code, rec.Body.String())
	}

	rec = request(mux, http.MethodGet, "/go-json/go/v2/settings", "", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("settings status = %d body = %s", rec.Code, rec.Body.String())
	}
	if bytes.Contains(rec.Body.Bytes(), []byte("site.private_note")) {
		t.Fatalf("private setting leaked: %s", rec.Body.String())
	}

	rec = request(mux, http.MethodGet, "/go-json/go/v2/posts?page=not-a-number", "", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("invalid pagination status = %d body = %s", rec.Code, rec.Body.String())
	}
	var errorEnvelope struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
			Status  int    `json:"status"`
		} `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &errorEnvelope); err != nil {
		t.Fatal(err)
	}
	if errorEnvelope.Error.Code != "validation_error" || errorEnvelope.Error.Status != http.StatusBadRequest || errorEnvelope.Error.Message == "" {
		t.Fatalf("unexpected error envelope: %+v", errorEnvelope)
	}
}

func TestRESTProjectionFilterDoesNotLeakPrivateMetadata(t *testing.T) {
	descriptor := restTestDescriptor{
		manifest: platformplugins.Manifest{
			ID:          "rest-filter",
			Name:        "REST Filter",
			Version:     "1.0.0",
			Contract:    "0.1",
			Description: "Filters REST content DTOs.",
			Hooks: []platformplugins.HookRegistration{
				{HookID: "rest.content.filter", HandlerID: "rest-filter.content", OwnerID: "rest-filter", Category: platformplugins.HookCategoryFilter},
			},
		},
		register: func(_ context.Context, registry *platformplugins.Registry) error {
			hook := platformplugins.HookRegistration{HookID: "rest.content.filter", HandlerID: "rest-filter.content", OwnerID: "rest-filter", Category: platformplugins.HookCategoryFilter}
			registry.AddHooks(hook)
			registry.AddFilterHandlers(platformplugins.FilterHandlerRegistration{
				Hook: hook,
				Handle: func(_ context.Context, _ platformplugins.HookContext, value any) (any, error) {
					dto := value.(restpkg.ContentDTO)
					dto.Title["en"] = "[filtered] " + dto.Title["en"]
					dto.Metadata["unexpected_private"] = "leak"
					return dto, nil
				},
			})
			return nil
		},
	}
	module, err := cms.NewWithOptions(cms.Options{
		DataSource:       "file:rest-filter-test?mode=memory&cache=shared",
		SessionKey:       "test-session-secret-test-session-secret",
		SeedFixtures:     true,
		RuntimeProfile:   "full",
		StorageProfile:   "sqlite",
		ActivePlugins:    []string{"rest-filter"},
		EnableDevBearer:  true,
		LoginPolicy:      "fixture",
		AdminPolicy:      "enabled",
		Preset:           "test",
		ExtraDescriptors: []platformplugins.Descriptor{descriptor},
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := module.Close(t.Context()); err != nil {
			t.Fatal(err)
		}
	})
	mux := http.NewServeMux()
	module.Routes(mux)

	rec := request(mux, http.MethodGet, "/go-json/go/v2/posts/by-slug/published-post", "", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body = %s", rec.Code, rec.Body.String())
	}
	var envelope struct {
		Data struct {
			Title    map[string]any `json:"title"`
			Metadata map[string]any `json:"metadata"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	if got := envelope.Data.Title["en"]; got != "[filtered] Published Post" {
		t.Fatalf("filtered title = %v", got)
	}
	if _, ok := envelope.Data.Metadata["private_key"]; ok {
		t.Fatalf("private metadata leaked: %+v", envelope.Data.Metadata)
	}
	if _, ok := envelope.Data.Metadata["unexpected_private"]; ok {
		t.Fatalf("unexpected metadata leaked: %+v", envelope.Data.Metadata)
	}
}

func TestRESTRegisteredPrivateMetaDoesNotLeakFromPublicProjectionFilter(t *testing.T) {
	descriptor := restTestDescriptor{
		manifest: platformplugins.Manifest{
			ID:          "rest-meta-filter",
			Name:        "REST Meta Filter",
			Version:     "1.0.0",
			Contract:    "0.1",
			Description: "Filters public content metadata.",
			Hooks: []platformplugins.HookRegistration{
				{HookID: "content.metadata.public.filter", HandlerID: "rest-meta-filter.public", OwnerID: "rest-meta-filter", Category: platformplugins.HookCategoryFilter},
			},
		},
		register: func(_ context.Context, registry *platformplugins.Registry) error {
			registry.AddFilterHandlers(platformplugins.FilterHandlerRegistration{
				Hook: platformplugins.HookRegistration{
					HookID:    "content.metadata.public.filter",
					HandlerID: "rest-meta-filter.public",
					OwnerID:   "rest-meta-filter",
					Category:  platformplugins.HookCategoryFilter,
				},
				Handle: func(_ context.Context, _ platformplugins.HookContext, value any) (any, error) {
					metadata := value.(map[string]any)
					metadata["secret_token"] = "leak"
					metadata["unexpected_private"] = "nope"
					return metadata, nil
				},
			})
			return nil
		},
	}
	module, err := cms.NewWithOptions(cms.Options{
		DataSource:       "file:rest-meta-registered?mode=memory&cache=shared",
		SessionKey:       "test-session-secret-test-session-secret",
		SeedFixtures:     true,
		RuntimeProfile:   "full",
		StorageProfile:   "sqlite",
		ActivePlugins:    []string{"rest-meta-filter"},
		EnableDevBearer:  true,
		LoginPolicy:      "fixture",
		AdminPolicy:      "enabled",
		Preset:           "test",
		ExtraDescriptors: []platformplugins.Descriptor{descriptor},
		ExtraMeta: []domainmeta.Definition{{
			Key:    "secret_token",
			Label:  "Secret token",
			Owner:  "test",
			Scope:  domainmeta.ScopeContent,
			Kinds:  []domaincontent.Kind{domaincontent.KindPost},
			Type:   domainmeta.ValueTypeString,
			Public: false,
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := module.Close(t.Context()); err != nil {
			t.Fatal(err)
		}
	})
	mux := http.NewServeMux()
	module.Routes(mux)

	body := `{"title":{"en":"Meta Post"},"slug":{"en":"meta-post"},"content":{"en":"Body"},"status":"published","metadata":{"seo_title":"Meta Title","secret_token":"top-secret","custom_public":"keep"}}`
	create := request(mux, http.MethodPost, "/go-json/go/v2/posts", body, "Bearer admin-token")
	if create.Code != http.StatusCreated {
		t.Fatalf("create status = %d body = %s", create.Code, create.Body.String())
	}

	rec := request(mux, http.MethodGet, "/go-json/go/v2/posts/by-slug/meta-post", "", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("detail status = %d body = %s", rec.Code, rec.Body.String())
	}
	var envelope struct {
		Data struct {
			Metadata map[string]any `json:"metadata"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	if got := envelope.Data.Metadata["seo_title"]; got != "Meta Title" {
		t.Fatalf("seo_title = %v", got)
	}
	if got := envelope.Data.Metadata["custom_public"]; got != "keep" {
		t.Fatalf("custom_public = %v", got)
	}
	if _, ok := envelope.Data.Metadata["secret_token"]; ok {
		t.Fatalf("registered private meta leaked: %+v", envelope.Data.Metadata)
	}
	if _, ok := envelope.Data.Metadata["unexpected_private"]; ok {
		t.Fatalf("unexpected metadata leaked: %+v", envelope.Data.Metadata)
	}
}

func TestRESTMediaMetadataValidation(t *testing.T) {
	module, err := cms.NewWithOptions(cms.Options{
		DataSource:      "file:rest-media-metadata?mode=memory&cache=shared",
		SessionKey:      "test-session-secret-test-session-secret",
		SeedFixtures:    true,
		RuntimeProfile:  "full",
		StorageProfile:  "sqlite",
		EnableDevBearer: true,
		LoginPolicy:     "fixture",
		AdminPolicy:     "enabled",
		Preset:          "test",
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := module.Close(t.Context()); err != nil {
			t.Fatal(err)
		}
	})
	mux := http.NewServeMux()
	module.Routes(mux)

	invalid := `{"id":"media-invalid","filename":"cover.txt","mime_type":"text/plain","public_url":"https://cdn.example.test/cover.txt"}`
	invalidRec := request(mux, http.MethodPost, "/go-json/go/v2/media", invalid, "Bearer admin-token")
	if invalidRec.Code != http.StatusBadRequest {
		t.Fatalf("expected invalid media to be rejected, got %d: %s", invalidRec.Code, invalidRec.Body.String())
	}

	valid := `{"id":"media-remote","filename":"cover.webp","mime_type":"image/webp","size_bytes":2048,"width":1024,"height":512,"alt_text":"Remote cover","caption":"Remote provider asset","public_url":"https://cdn.example.test/cover.webp","provider_ref":{"provider":"s3","key":"media/originals/cover.webp","url":"https://bucket.example.test/cover.webp","checksum":"sha256:test","etag":"etag-1"}}`
	rec := request(mux, http.MethodPost, "/go-json/go/v2/media", valid, "Bearer admin-token")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected valid media metadata, got %d: %s", rec.Code, rec.Body.String())
	}
	var envelope struct {
		Data struct {
			ID          string `json:"id"`
			Provider    string `json:"provider"`
			ProviderURL string `json:"provider_url"`
			PublicURL   string `json:"public_url"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	if envelope.Data.Provider != "s3" || envelope.Data.ProviderURL != "https://bucket.example.test/cover.webp" {
		t.Fatalf("provider projection = %+v", envelope.Data)
	}
}

func request(handler http.Handler, method string, path string, body string, authorization string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("User-Agent", "gocms-test")
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if authorization != "" {
		req.Header.Set("Authorization", authorization)
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

type restTestDescriptor struct {
	manifest platformplugins.Manifest
	register func(context.Context, *platformplugins.Registry) error
}

func (d restTestDescriptor) Manifest() platformplugins.Manifest {
	return d.manifest
}

func (d restTestDescriptor) Register(ctx context.Context, registry *platformplugins.Registry) error {
	if d.register == nil {
		return nil
	}
	return d.register(ctx, registry)
}
