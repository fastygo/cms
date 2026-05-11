package cms

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	domaincontent "github.com/fastygo/cms/internal/domain/content"
	domainsettings "github.com/fastygo/cms/internal/domain/settings"
	platformplugins "github.com/fastygo/cms/internal/platform/plugins"
	platformthemes "github.com/fastygo/cms/internal/platform/themes"
	frameworkapp "github.com/fastygo/framework/pkg/app"
)

func TestFullProfileExposesPublicRoutesAndPublishedContent(t *testing.T) {
	mux, closeFn := newPublicMux(t, "full", nil)
	defer closeFn()

	home := requestPublic(mux, http.MethodGet, "/", "", "")
	if home.Code != http.StatusOK {
		t.Fatalf("home status = %d body = %s", home.Code, home.Body.String())
	}
	body := home.Body.String()
	for _, expected := range []string{
		`data-gocms-public-screen="home"`,
		`data-gocms-public-header="gocms-default"`,
		`data-gocms-menu-location="header"`,
		`data-gocms-menu-location="footer"`,
		"GoCMS Fixture",
		"Published Post",
		"News",
		"Author Jane",
	} {
		if !strings.Contains(body, expected) {
			t.Fatalf("expected home page to contain %q", expected)
		}
	}
	for _, unexpected := range []string{"Draft Post", "Scheduled Post"} {
		if strings.Contains(body, unexpected) {
			t.Fatalf("unexpected private content leak %q", unexpected)
		}
	}

	page := requestPublic(mux, http.MethodGet, "/about/", "", "")
	if page.Code != http.StatusOK || !strings.Contains(page.Body.String(), "About page fixture") {
		t.Fatalf("page response = %d body = %s", page.Code, page.Body.String())
	}

	post := requestPublic(mux, http.MethodGet, "/published-post/", "", "")
	if post.Code != http.StatusOK || !strings.Contains(post.Body.String(), "Public fixture content") {
		t.Fatalf("post response = %d body = %s", post.Code, post.Body.String())
	}
	for _, expected := range []string{
		"/media/cover.jpg",
		"Jane Editor",
		`data-gocms-breadcrumbs="public"`,
		`property="og:title"`,
		`rel="canonical"`,
	} {
		if !strings.Contains(post.Body.String(), expected) {
			t.Fatalf("expected post page to contain %q", expected)
		}
	}

	blog := requestPublic(mux, http.MethodGet, "/blog/", "", "")
	if blog.Code != http.StatusOK || !strings.Contains(blog.Body.String(), `data-gocms-public-screen="blog"`) {
		t.Fatalf("blog response = %d body = %s", blog.Code, blog.Body.String())
	}

	search := requestPublic(mux, http.MethodGet, "/search?q=Published", "", "")
	if search.Code != http.StatusOK || !strings.Contains(search.Body.String(), "Published Post") {
		t.Fatalf("search response = %d body = %s", search.Code, search.Body.String())
	}

	taxonomy := requestPublic(mux, http.MethodGet, "/category/news/", "", "")
	if taxonomy.Code != http.StatusOK || !strings.Contains(taxonomy.Body.String(), "Published Post") {
		t.Fatalf("taxonomy response = %d body = %s", taxonomy.Code, taxonomy.Body.String())
	}

	author := requestPublic(mux, http.MethodGet, "/author/jane/", "", "")
	if author.Code != http.StatusOK || !strings.Contains(author.Body.String(), "Published Post") {
		t.Fatalf("author response = %d body = %s", author.Code, author.Body.String())
	}

	draft := requestPublic(mux, http.MethodGet, "/draft-post/", "", "")
	if draft.Code != http.StatusNotFound {
		t.Fatalf("draft should not render publicly, got %d", draft.Code)
	}

	scheduled := requestPublic(mux, http.MethodGet, "/scheduled-post/", "", "")
	if scheduled.Code != http.StatusNotFound {
		t.Fatalf("scheduled should not render publicly, got %d", scheduled.Code)
	}
}

func TestPublicThemePreviewAndActivationAffectRender(t *testing.T) {
	module := newModuleForPublicTests(t, "full", nil)
	t.Cleanup(func() {
		_ = module.Close(t.Context())
	})
	mux := http.NewServeMux()
	module.Routes(mux)

	preview := requestPublic(mux, http.MethodGet, "/?preview_theme=blank&preview_preset=minimal", "", "")
	if preview.Code != http.StatusOK || !strings.Contains(preview.Body.String(), `data-gocms-theme="blank"`) {
		t.Fatalf("preview response = %d body = %s", preview.Code, preview.Body.String())
	}

	if err := module.store.SaveSetting(t.Context(), domainsettings.Value{Key: domainsettings.Key(platformthemes.ActiveThemeKey), Value: "blank", Public: false}); err != nil {
		t.Fatal(err)
	}
	if err := module.store.SaveSetting(t.Context(), domainsettings.Value{Key: domainsettings.Key(platformthemes.StylePresetKey), Value: "minimal", Public: false}); err != nil {
		t.Fatal(err)
	}

	activated := requestPublic(mux, http.MethodGet, "/", "", "")
	if activated.Code != http.StatusOK || !strings.Contains(activated.Body.String(), `data-gocms-theme="blank"`) {
		t.Fatalf("activated response = %d body = %s", activated.Code, activated.Body.String())
	}
}

func TestPublicContentBodyRendersHTMLInsideProse(t *testing.T) {
	module := newModuleForPublicTests(t, "full", nil)
	t.Cleanup(func() {
		_ = module.Close(t.Context())
	})

	entry, err := module.store.Get(t.Context(), domaincontent.ID("content-post-published"))
	if err != nil {
		t.Fatal(err)
	}
	entry.Body = domaincontent.LocalizedText{
		"en": "<p>Public <strong>fixture</strong> <em>content</em></p><blockquote><p>Rendered quote</p></blockquote>",
	}
	if err := module.store.Save(t.Context(), entry); err != nil {
		t.Fatal(err)
	}

	mux := http.NewServeMux()
	module.Routes(mux)

	post := requestPublic(mux, http.MethodGet, "/published-post/", "", "")
	if post.Code != http.StatusOK {
		t.Fatalf("post response = %d body = %s", post.Code, post.Body.String())
	}
	for _, expected := range []string{
		`prose max-w-none`,
		`<strong>fixture</strong>`,
		`<em>content</em>`,
		`<blockquote><p>Rendered quote</p></blockquote>`,
	} {
		if !strings.Contains(post.Body.String(), expected) {
			t.Fatalf("expected rendered post body to contain %q", expected)
		}
	}
	for _, unexpected := range []string{
		`&lt;strong&gt;fixture&lt;/strong&gt;`,
		`&lt;blockquote&gt;`,
	} {
		if strings.Contains(post.Body.String(), unexpected) {
			t.Fatalf("expected public body HTML to render, got escaped fragment %q", unexpected)
		}
	}
}

func TestPublicRenderFilterAppliesAtSafeOutputBoundary(t *testing.T) {
	descriptor := moduleTestDescriptor{
		manifest: platformplugins.Manifest{
			ID:          "render-filter",
			Name:        "Render Filter",
			Version:     "1.0.0",
			Contract:    "0.1",
			Description: "Adds a public render marker.",
			Hooks: []platformplugins.HookRegistration{
				{HookID: "render.content.filter", HandlerID: "render-filter.public", OwnerID: "render-filter", Category: platformplugins.HookCategoryFilter},
			},
		},
		register: func(_ context.Context, registry *platformplugins.Registry) error {
			manifest := platformplugins.Manifest{
				Hooks: []platformplugins.HookRegistration{
					{HookID: "render.content.filter", HandlerID: "render-filter.public", OwnerID: "render-filter", Category: platformplugins.HookCategoryFilter},
				},
			}
			registry.AddHooks(manifest.Hooks...)
			registry.AddFilterHandlers(platformplugins.FilterHandlerRegistration{
				Hook: manifest.Hooks[0],
				Handle: func(_ context.Context, _ platformplugins.HookContext, value any) (any, error) {
					return `<div data-render-filter="enabled"></div>` + value.(string), nil
				},
			})
			return nil
		},
	}

	module := newModuleForPublicTestsWithDescriptors(t, "full", []string{"render-filter"}, []platformplugins.Descriptor{descriptor})
	t.Cleanup(func() {
		_ = module.Close(t.Context())
	})

	mux := http.NewServeMux()
	module.Routes(mux)
	post := requestPublic(mux, http.MethodGet, "/published-post/", "", "")
	if post.Code != http.StatusOK {
		t.Fatalf("post response = %d body = %s", post.Code, post.Body.String())
	}
	if !strings.Contains(post.Body.String(), `data-render-filter="enabled"`) {
		t.Fatalf("filtered post body = %s", post.Body.String())
	}
	for _, unexpected := range []string{"Draft Post", "Scheduled Post"} {
		if strings.Contains(post.Body.String(), unexpected) {
			t.Fatalf("unexpected private content leak %q", unexpected)
		}
	}
}

func TestPublicArchivesExposePaginationWithoutPrivateLeaks(t *testing.T) {
	module := newModuleForPublicTests(t, "full", nil)
	t.Cleanup(func() {
		_ = module.Close(t.Context())
	})
	seedAdditionalPublishedPosts(t, module, 12)
	mux := http.NewServeMux()
	module.Routes(mux)

	pageOne := requestPublic(mux, http.MethodGet, "/blog/", "", "")
	if pageOne.Code != http.StatusOK {
		t.Fatalf("page one status = %d body = %s", pageOne.Code, pageOne.Body.String())
	}
	for _, expected := range []string{"Extra Published 12", "Extra Published 11"} {
		if !strings.Contains(pageOne.Body.String(), expected) {
			t.Fatalf("expected page one to contain %q", expected)
		}
	}
	for _, unexpected := range []string{"Draft Post", "Scheduled Post"} {
		if strings.Contains(pageOne.Body.String(), unexpected) {
			t.Fatalf("unexpected archive leak %q", unexpected)
		}
	}

	pageTwo := requestPublic(mux, http.MethodGet, "/blog/?page=2", "", "")
	if pageTwo.Code != http.StatusOK || !strings.Contains(pageTwo.Body.String(), "Extra Published 02") || !strings.Contains(pageTwo.Body.String(), "Published Post") {
		t.Fatalf("page two status = %d body = %s", pageTwo.Code, pageTwo.Body.String())
	}
}

func TestHeadlessProfileDoesNotExposePublicRenderer(t *testing.T) {
	mux, closeFn := newPublicMux(t, "headless", nil)
	defer closeFn()

	rec := requestPublic(mux, http.MethodGet, "/", "", "")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("headless public route = %d, want 404", rec.Code)
	}
}

func TestPublicCatchAllDoesNotSwallowSystemRoutes(t *testing.T) {
	module := newModuleForPublicTests(t, "full", []string{"graphql"})
	t.Cleanup(func() {
		_ = module.Close(t.Context())
	})

	mux := http.NewServeMux()
	module.Routes(mux)

	admin := requestPublic(mux, http.MethodGet, "/go-admin", "", "")
	if admin.Code != http.StatusSeeOther {
		t.Fatalf("/go-admin status = %d body = %s", admin.Code, admin.Body.String())
	}

	rest := requestPublic(mux, http.MethodGet, "/go-json/go/v2/posts/by-slug/published-post", "", "")
	if rest.Code != http.StatusOK {
		t.Fatalf("/go-json status = %d body = %s", rest.Code, rest.Body.String())
	}

	graphqlBody, _ := json.Marshal(map[string]any{"query": `query { post(slug: "published-post") { id } }`})
	graphql := requestPublic(mux, http.MethodPost, "/go-graphql", string(graphqlBody), "")
	if graphql.Code != http.StatusOK {
		t.Fatalf("/go-graphql status = %d body = %s", graphql.Code, graphql.Body.String())
	}

	staticDir := t.TempDir()
	staticFile := filepath.Join(staticDir, "sample.css")
	if err := os.WriteFile(staticFile, []byte("body{}"), 0o644); err != nil {
		t.Fatal(err)
	}

	app := frameworkapp.New(frameworkapp.Config{
		AppBind:   "127.0.0.1:0",
		StaticDir: staticDir,
	}).WithFeature(module).Build()

	static := httptest.NewRecorder()
	staticReq := httptest.NewRequest(http.MethodGet, "/static/sample.css", nil)
	staticReq.Header.Set("User-Agent", "gocms-public-test")
	app.ServeHTTP(static, staticReq)
	if static.Code != http.StatusOK {
		t.Fatalf("/static/sample.css status = %d body = %s", static.Code, static.Body.String())
	}
}

func newPublicMux(t *testing.T, runtimeProfile string, activePlugins []string) (*http.ServeMux, func()) {
	t.Helper()
	module := newModuleForPublicTests(t, runtimeProfile, activePlugins)
	mux := http.NewServeMux()
	module.Routes(mux)
	return mux, func() {
		_ = module.Close(t.Context())
	}
}

func newModuleForPublicTests(t *testing.T, runtimeProfile string, activePlugins []string) *Module {
	t.Helper()
	return newModuleForPublicTestsWithDescriptors(t, runtimeProfile, activePlugins, nil)
}

func newModuleForPublicTestsWithDescriptors(t *testing.T, runtimeProfile string, activePlugins []string, descriptors []platformplugins.Descriptor) *Module {
	t.Helper()
	module, err := NewWithOptions(Options{
		DataSource:       "file:" + sanitizeName(t.Name()) + "?mode=memory&cache=shared",
		SessionKey:       "public-test-session-secret-public-test",
		SeedFixtures:     true,
		RuntimeProfile:   runtimeProfile,
		StorageProfile:   "sqlite",
		ActivePlugins:    activePlugins,
		EnableDevBearer:  true,
		LoginPolicy:      "fixture",
		AdminPolicy:      "enabled",
		Preset:           runtimeProfile,
		ExtraDescriptors: descriptors,
	})
	if err != nil {
		t.Fatal(err)
	}
	return module
}

type moduleTestDescriptor struct {
	manifest platformplugins.Manifest
	register func(context.Context, *platformplugins.Registry) error
}

func (d moduleTestDescriptor) Manifest() platformplugins.Manifest {
	return d.manifest
}

func (d moduleTestDescriptor) Register(ctx context.Context, registry *platformplugins.Registry) error {
	if d.register == nil {
		return nil
	}
	return d.register(ctx, registry)
}

func requestPublic(handler http.Handler, method string, path string, body string, authorization string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
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

func sanitizeName(value string) string {
	replacer := strings.NewReplacer("/", "-", "\\", "-", " ", "-")
	return replacer.Replace(strings.ToLower(value))
}

func seedAdditionalPublishedPosts(t *testing.T, module *Module, count int) {
	t.Helper()
	now := time.Date(2026, 5, 3, 12, 0, 0, 0, time.UTC)
	for i := 1; i <= count; i++ {
		publishedAt := now.Add(time.Duration(i) * time.Hour)
		entry := domaincontent.Entry{
			ID:          domaincontent.ID("content-post-extra-" + twoDigits(i)),
			Kind:        domaincontent.KindPost,
			Status:      domaincontent.StatusPublished,
			Visibility:  domaincontent.VisibilityPublic,
			Title:       domaincontent.LocalizedText{"en": "Extra Published " + twoDigits(i)},
			Slug:        domaincontent.LocalizedText{"en": "extra-published-" + twoDigits(i)},
			Body:        domaincontent.LocalizedText{"en": "Extra body " + twoDigits(i)},
			Excerpt:     domaincontent.LocalizedText{"en": "Extra excerpt " + twoDigits(i)},
			AuthorID:    "author-1",
			CreatedAt:   publishedAt.Add(-time.Hour),
			UpdatedAt:   publishedAt,
			PublishedAt: &publishedAt,
		}
		if err := module.store.Save(t.Context(), entry); err != nil {
			t.Fatal(err)
		}
	}
}

func twoDigits(value int) string {
	if value < 10 {
		return "0" + strconv.Itoa(value)
	}
	return strconv.Itoa(value)
}
