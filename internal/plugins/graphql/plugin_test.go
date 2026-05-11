package graphqlplugin_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	domaincontent "github.com/fastygo/cms/internal/domain/content"
	domainmeta "github.com/fastygo/cms/internal/domain/meta"
	"github.com/fastygo/cms/internal/infra/features/cms"
	platformplugins "github.com/fastygo/cms/internal/platform/plugins"
)

type graphQLResponse struct {
	Data   map[string]any `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

func TestGraphQLRouteOnlyExistsWhenPluginIsActive(t *testing.T) {
	t.Run("inactive", func(t *testing.T) {
		mux := newMux(t, nil)
		rec := graphQLRequest(mux, `query { posts { pagination { total } } }`, "")
		if rec.Code != http.StatusNotFound {
			t.Fatalf("status = %d body = %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("active", func(t *testing.T) {
		mux := newMux(t, []string{"graphql"})
		rec := graphQLRequest(mux, `query { posts { pagination { total } } }`, "")
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d body = %s", rec.Code, rec.Body.String())
		}
	})
}

func TestGraphQLReadCoverageAndPublicVisibility(t *testing.T) {
	mux := newMux(t, []string{"graphql"})
	rec := graphQLRequest(mux, `
		query {
			posts {
				items {
					id
					status
					metadata
					authorID
					featuredMediaID
					taxonomies {
						taxonomy
						termID
					}
				}
				pagination {
					total
					perPage
				}
			}
			post(slug: "published-post") {
				id
			}
			pages {
				items {
					id
				}
				pagination {
					total
				}
			}
			page(slug: "about") {
				id
			}
			contentTypes {
				id
				graphqlVisible
			}
			taxonomies {
				type
				graphqlVisible
			}
			terms(type: "category") {
				id
				type
			}
			media {
				id
				filename
				metadata
			}
			authors {
				id
				slug
			}
			menus {
				id
				location
			}
			settings {
				key
				value
				public
			}
			search(query: "Published") {
				items {
					id
				}
				pagination {
					total
				}
			}
		}
	`, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body = %s", rec.Code, rec.Body.String())
	}

	response := decodeGraphQL(t, rec)
	if len(response.Errors) > 0 {
		t.Fatalf("unexpected graphql errors: %+v", response.Errors)
	}

	posts := mustMap(t, response.Data["posts"])
	items := mustSlice(t, posts["items"])
	if len(items) != 1 {
		t.Fatalf("expected one public post, got %d", len(items))
	}
	post := mustMap(t, items[0])
	if got := mustString(t, post["status"]); got != "published" {
		t.Fatalf("status = %q", got)
	}
	metadata := mustMap(t, post["metadata"])
	if _, ok := metadata["private_key"]; ok {
		t.Fatalf("private metadata leaked: %+v", metadata)
	}
	if got := mustString(t, post["authorID"]); got != "author-1" {
		t.Fatalf("authorID = %q", got)
	}
	if got := mustString(t, post["featuredMediaID"]); got != "media-cover" {
		t.Fatalf("featuredMediaID = %q", got)
	}
	assignments := mustSlice(t, post["taxonomies"])
	if len(assignments) != 2 {
		t.Fatalf("expected taxonomy assignments, got %+v", assignments)
	}
	pagination := mustMap(t, posts["pagination"])
	if got := mustInt(t, pagination["total"]); got != 1 {
		t.Fatalf("posts total = %d", got)
	}

	if response.Data["post"] == nil {
		t.Fatal("post lookup returned nil")
	}
	if response.Data["page"] == nil {
		t.Fatal("page lookup returned nil")
	}

	contentTypes := mustSlice(t, response.Data["contentTypes"])
	if len(contentTypes) < 2 {
		t.Fatalf("contentTypes = %+v", contentTypes)
	}
	taxonomies := mustSlice(t, response.Data["taxonomies"])
	if len(taxonomies) < 2 {
		t.Fatalf("taxonomies = %+v", taxonomies)
	}
	terms := mustSlice(t, response.Data["terms"])
	if len(terms) == 0 {
		t.Fatal("expected category terms")
	}
	media := mustSlice(t, response.Data["media"])
	if len(media) == 0 {
		t.Fatal("expected media assets")
	}
	authors := mustSlice(t, response.Data["authors"])
	if len(authors) == 0 {
		t.Fatal("expected authors")
	}
	menus := mustSlice(t, response.Data["menus"])
	if len(menus) == 0 {
		t.Fatal("expected menus")
	}
	settings := mustSlice(t, response.Data["settings"])
	keys := map[string]struct{}{}
	for _, item := range settings {
		keys[mustString(t, mustMap(t, item)["key"])] = struct{}{}
	}
	for _, expected := range []string{"site.title", "public.rendering", "theme.brand_name", "theme.home_intro"} {
		if _, ok := keys[expected]; !ok {
			t.Fatalf("expected public setting %q, got %+v", expected, settings)
		}
	}
	if _, ok := keys["site.private_note"]; ok {
		t.Fatalf("private setting leaked: %+v", settings)
	}
	search := mustMap(t, response.Data["search"])
	if got := mustInt(t, mustMap(t, search["pagination"])["total"]); got != 1 {
		t.Fatalf("search total = %d", got)
	}
}

func TestGraphQLAuthVisibilityAndMutationCapabilities(t *testing.T) {
	mux := newMux(t, []string{"graphql"})

	rec := graphQLRequest(mux, `query { post(id: "content-post-draft") { id status } }`, "")
	response := decodeGraphQL(t, rec)
	if len(response.Errors) > 0 {
		t.Fatalf("unexpected anonymous errors: %+v", response.Errors)
	}
	if response.Data["post"] != nil {
		t.Fatalf("draft leaked to anonymous caller: %+v", response.Data["post"])
	}

	rec = graphQLRequest(mux, `query { post(id: "content-post-draft") { id status } }`, "Bearer admin-token")
	response = decodeGraphQL(t, rec)
	if len(response.Errors) > 0 {
		t.Fatalf("unexpected admin errors: %+v", response.Errors)
	}
	privatePost := mustMap(t, response.Data["post"])
	if got := mustString(t, privatePost["status"]); got != "draft" {
		t.Fatalf("status = %q", got)
	}

	rec = graphQLRequest(mux, `mutation { publishContent(id: "content-post-draft") { id } }`, "Bearer viewer-token")
	response = decodeGraphQL(t, rec)
	if len(response.Errors) == 0 {
		t.Fatalf("expected mutation to fail for low privilege principal: %+v", response.Data)
	}

	rec = graphQLRequest(mux, `
		mutation {
			createPost(input: {
				title: { en: "GraphQL Created" }
				slug: { en: "graphql-created" }
				content: { en: "Created from GraphQL" }
				status: "published"
			}) {
				id
				status
				slug
			}
		}
	`, "Bearer admin-token")
	response = decodeGraphQL(t, rec)
	if len(response.Errors) > 0 {
		t.Fatalf("unexpected create mutation errors: %+v", response.Errors)
	}
	created := mustMap(t, response.Data["createPost"])
	if got := mustString(t, created["status"]); got != "published" {
		t.Fatalf("created status = %q", got)
	}
	slug := mustMap(t, created["slug"])
	if got := mustString(t, slug["en"]); got != "graphql-created" {
		t.Fatalf("created slug = %q", got)
	}
}

func TestGraphQLMatchesRESTProjectionForPublishedPost(t *testing.T) {
	mux := newMux(t, []string{"graphql"})

	restRec := request(mux, http.MethodGet, "/go-json/go/v2/posts/by-slug/published-post", "", "")
	if restRec.Code != http.StatusOK {
		t.Fatalf("rest status = %d body = %s", restRec.Code, restRec.Body.String())
	}
	var restEnvelope struct {
		Data map[string]any `json:"data"`
	}
	if err := json.Unmarshal(restRec.Body.Bytes(), &restEnvelope); err != nil {
		t.Fatal(err)
	}

	graphRec := graphQLRequest(mux, `
		query {
			post(slug: "published-post") {
				id
				status
				authorID
				featuredMediaID
				slug
				title
				taxonomies {
					taxonomy
					termID
				}
			}
		}
	`, "")
	if graphRec.Code != http.StatusOK {
		t.Fatalf("graphql status = %d body = %s", graphRec.Code, graphRec.Body.String())
	}
	graph := decodeGraphQL(t, graphRec)
	if len(graph.Errors) > 0 {
		t.Fatalf("unexpected graphql errors: %+v", graph.Errors)
	}
	post := mustMap(t, graph.Data["post"])
	if got, want := mustString(t, post["id"]), mustString(t, restEnvelope.Data["id"]); got != want {
		t.Fatalf("id mismatch: got %q want %q", got, want)
	}
	if got, want := mustString(t, post["status"]), mustString(t, restEnvelope.Data["status"]); got != want {
		t.Fatalf("status mismatch: got %q want %q", got, want)
	}
	if got, want := mustString(t, post["authorID"]), mustString(t, restEnvelope.Data["author_id"]); got != want {
		t.Fatalf("author mismatch: got %q want %q", got, want)
	}
	if got, want := mustString(t, post["featuredMediaID"]), mustString(t, restEnvelope.Data["featured_media_id"]); got != want {
		t.Fatalf("featured media mismatch: got %q want %q", got, want)
	}
	restSlug := mustMap(t, restEnvelope.Data["slug"])
	if got, want := mustString(t, mustMap(t, post["slug"])["en"]), mustString(t, restSlug["en"]); got != want {
		t.Fatalf("slug mismatch: got %q want %q", got, want)
	}
	restTitle := mustMap(t, restEnvelope.Data["title"])
	if got, want := mustString(t, mustMap(t, post["title"])["en"]), mustString(t, restTitle["en"]); got != want {
		t.Fatalf("title mismatch: got %q want %q", got, want)
	}
	if got, want := len(mustSlice(t, post["taxonomies"])), len(mustSlice(t, restEnvelope.Data["taxonomy_ids"])); got != want {
		t.Fatalf("taxonomy count mismatch: got %d want %d", got, want)
	}
}

func TestGraphQLMatchesRESTForRegisteredPublicMetadata(t *testing.T) {
	module, err := cms.NewWithOptions(cms.Options{
		DataSource:      "file:graphql-meta-projection?mode=memory&cache=shared",
		SessionKey:      "test-session-secret-test-session-secret",
		SeedFixtures:    true,
		RuntimeProfile:  "full",
		StorageProfile:  "sqlite",
		ActivePlugins:   []string{"graphql"},
		EnableDevBearer: true,
		LoginPolicy:     "fixture",
		AdminPolicy:     "enabled",
		Preset:          "test",
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

	create := request(mux, http.MethodPost, "/go-json/go/v2/posts", `{"title":{"en":"Projection Post"},"slug":{"en":"projection-post"},"content":{"en":"Body"},"status":"published","metadata":{"seo_title":"Projection Title","secret_token":"top-secret"}}`, "Bearer admin-token")
	if create.Code != http.StatusCreated {
		t.Fatalf("create status = %d body = %s", create.Code, create.Body.String())
	}

	restRec := request(mux, http.MethodGet, "/go-json/go/v2/posts/by-slug/projection-post", "", "")
	if restRec.Code != http.StatusOK {
		t.Fatalf("rest status = %d body = %s", restRec.Code, restRec.Body.String())
	}
	var restEnvelope struct {
		Data struct {
			Metadata map[string]any `json:"metadata"`
		} `json:"data"`
	}
	if err := json.Unmarshal(restRec.Body.Bytes(), &restEnvelope); err != nil {
		t.Fatal(err)
	}

	graphRec := graphQLRequest(mux, `query { post(slug: "projection-post") { metadata } }`, "")
	if graphRec.Code != http.StatusOK {
		t.Fatalf("graphql status = %d body = %s", graphRec.Code, graphRec.Body.String())
	}
	response := decodeGraphQL(t, graphRec)
	if len(response.Errors) > 0 {
		t.Fatalf("unexpected graphql errors: %+v", response.Errors)
	}
	post := mustMap(t, response.Data["post"])
	graphMeta := mustMap(t, post["metadata"])

	if got := restEnvelope.Data.Metadata["seo_title"]; got != "Projection Title" {
		t.Fatalf("rest seo_title = %v", got)
	}
	if got := graphMeta["seo_title"]; got != "Projection Title" {
		t.Fatalf("graphql seo_title = %v", got)
	}
	if _, ok := restEnvelope.Data.Metadata["secret_token"]; ok {
		t.Fatalf("rest leaked private meta: %+v", restEnvelope.Data.Metadata)
	}
	if _, ok := graphMeta["secret_token"]; ok {
		t.Fatalf("graphql leaked private meta: %+v", graphMeta)
	}
}

func TestGraphQLStatusAppearsInAdminSurfaces(t *testing.T) {
	mux := newMux(t, []string{"graphql"})

	rec := request(mux, http.MethodGet, "/go-admin/headless", "", "Bearer admin-token")
	if rec.Code != http.StatusOK {
		t.Fatalf("headless status = %d body = %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "/go-admin/plugins/graphql/status") {
		t.Fatalf("headless page does not include GraphQL status action: %s", rec.Body.String())
	}

	rec = request(mux, http.MethodGet, "/go-admin/plugins/graphql/status", "", "Bearer admin-token")
	if rec.Code != http.StatusOK {
		t.Fatalf("status route = %d body = %s", rec.Code, rec.Body.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if got := mustString(t, payload["plugin"]); got != "graphql" {
		t.Fatalf("plugin = %q", got)
	}
	if got := mustString(t, payload["endpoint"]); got != "/go-graphql" {
		t.Fatalf("endpoint = %q", got)
	}
}

func TestGraphQLProjectionFilterRemainsAuthorizationSafe(t *testing.T) {
	descriptor := graphqlTestDescriptor{
		manifest: platformplugins.Manifest{
			ID:          "graphql-filter",
			Name:        "GraphQL Filter",
			Version:     "1.0.0",
			Contract:    "0.1",
			Description: "Filters GraphQL content projections.",
			Hooks: []platformplugins.HookRegistration{
				{HookID: "graphql.content.filter", HandlerID: "graphql-filter.content", OwnerID: "graphql-filter", Category: platformplugins.HookCategoryFilter},
			},
		},
		register: func(_ context.Context, registry *platformplugins.Registry) error {
			hook := platformplugins.HookRegistration{HookID: "graphql.content.filter", HandlerID: "graphql-filter.content", OwnerID: "graphql-filter", Category: platformplugins.HookCategoryFilter}
			registry.AddHooks(hook)
			registry.AddFilterHandlers(platformplugins.FilterHandlerRegistration{
				Hook: hook,
				Handle: func(_ context.Context, _ platformplugins.HookContext, value any) (any, error) {
					projection := value.(platformplugins.ContentProjection)
					projection.Title["en"] = "[filtered] " + projection.Title["en"]
					projection.Metadata["private_key"] = "leak"
					return projection, nil
				},
			})
			return nil
		},
	}
	mux := newMuxWithDescriptors(t, []string{"graphql", "graphql-filter"}, []platformplugins.Descriptor{descriptor})
	rec := graphQLRequest(mux, `
		query {
			post(slug: "published-post") {
				title
				metadata
			}
		}
	`, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("graphql status = %d body = %s", rec.Code, rec.Body.String())
	}
	response := decodeGraphQL(t, rec)
	if len(response.Errors) > 0 {
		t.Fatalf("unexpected graphql errors: %+v", response.Errors)
	}
	post := mustMap(t, response.Data["post"])
	title := mustMap(t, post["title"])
	if got := mustString(t, title["en"]); got != "[filtered] Published Post" {
		t.Fatalf("filtered title = %q", got)
	}
	metadata := mustMap(t, post["metadata"])
	if _, ok := metadata["private_key"]; ok {
		t.Fatalf("private metadata leaked: %+v", metadata)
	}
}

func newMux(t *testing.T, activePlugins []string) *http.ServeMux {
	return newMuxWithDescriptors(t, activePlugins, nil)
}

func newMuxWithDescriptors(t *testing.T, activePlugins []string, descriptors []platformplugins.Descriptor) *http.ServeMux {
	t.Helper()
	module, err := cms.NewWithOptions(cms.Options{
		DataSource:       fmt.Sprintf("file:%s?mode=memory&cache=shared", sanitizeName(t.Name())),
		SessionKey:       "test-session-secret-test-session-secret",
		SeedFixtures:     true,
		RuntimeProfile:   "full",
		StorageProfile:   "sqlite",
		ActivePlugins:    activePlugins,
		EnableDevBearer:  true,
		LoginPolicy:      "fixture",
		AdminPolicy:      "enabled",
		Preset:           "test",
		ExtraDescriptors: descriptors,
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
	return mux
}

func graphQLRequest(handler http.Handler, query string, authorization string) *httptest.ResponseRecorder {
	body, _ := json.Marshal(map[string]any{"query": query})
	req := httptest.NewRequest(http.MethodPost, "/go-graphql", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "gocms-graphql-test")
	if authorization != "" {
		req.Header.Set("Authorization", authorization)
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

func request(handler http.Handler, method string, path string, body string, authorization string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("User-Agent", "gocms-graphql-test")
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

func decodeGraphQL(t *testing.T, rec *httptest.ResponseRecorder) graphQLResponse {
	t.Helper()
	var response graphQLResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("unmarshal response: %v body = %s", err, rec.Body.String())
	}
	return response
}

func sanitizeName(value string) string {
	replacer := strings.NewReplacer("/", "-", "\\", "-", " ", "-")
	return replacer.Replace(strings.ToLower(value))
}

func mustMap(t *testing.T, value any) map[string]any {
	t.Helper()
	result, ok := value.(map[string]any)
	if !ok {
		t.Fatalf("expected map, got %T", value)
	}
	return result
}

func mustSlice(t *testing.T, value any) []any {
	t.Helper()
	result, ok := value.([]any)
	if !ok {
		t.Fatalf("expected slice, got %T", value)
	}
	return result
}

func mustString(t *testing.T, value any) string {
	t.Helper()
	result, ok := value.(string)
	if !ok {
		t.Fatalf("expected string, got %T", value)
	}
	return result
}

func mustInt(t *testing.T, value any) int {
	t.Helper()
	number, ok := value.(float64)
	if !ok {
		t.Fatalf("expected number, got %T", value)
	}
	return int(number)
}

type graphqlTestDescriptor struct {
	manifest platformplugins.Manifest
	register func(context.Context, *platformplugins.Registry) error
}

func (d graphqlTestDescriptor) Manifest() platformplugins.Manifest {
	return d.manifest
}

func (d graphqlTestDescriptor) Register(ctx context.Context, registry *platformplugins.Registry) error {
	if d.register == nil {
		return nil
	}
	return d.register(ctx, registry)
}
