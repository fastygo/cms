package cms

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fastygo/cms/internal/platform/runtimeprofile"
)

func TestNewWithOptionsUsesBrowserLocalProfileWithoutDurableDataSource(t *testing.T) {
	module, err := NewWithOptions(Options{
		DataSource:     "file:/path/that/should/not/be/created/gocms.db",
		SessionKey:     "test-session-key",
		SeedFixtures:   true,
		RuntimeProfile: string(runtimeprofile.RuntimeProfilePlayground),
		StorageProfile: string(runtimeprofile.StorageProfileBrowserIndexedDB),
	})
	if err != nil {
		t.Fatalf("NewWithOptions() error = %v", err)
	}
	t.Cleanup(func() {
		_ = module.Close(t.Context())
	})
}

func TestHeadlessProfileDoesNotExposeAdminRoutes(t *testing.T) {
	module, err := NewWithOptions(Options{
		DataSource:     "file:headless-test?mode=memory&cache=shared",
		SessionKey:     "test-session-key",
		SeedFixtures:   true,
		RuntimeProfile: string(runtimeprofile.RuntimeProfileHeadless),
		StorageProfile: string(runtimeprofile.StorageProfileSQLite),
	})
	if err != nil {
		t.Fatalf("NewWithOptions() error = %v", err)
	}
	t.Cleanup(func() {
		_ = module.Close(t.Context())
	})
	mux := http.NewServeMux()
	module.Routes(mux)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/go-admin", nil)
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestPlaygroundProfileExposesAdminAndPublicPreviewWithoutREST(t *testing.T) {
	module, err := NewWithOptions(Options{
		DataSource:       "file:playground-preview-test?mode=memory&cache=shared",
		SessionKey:       "test-session-key",
		SeedFixtures:     false,
		RuntimeProfile:   string(runtimeprofile.RuntimeProfilePlayground),
		StorageProfile:   string(runtimeprofile.StorageProfileBrowserIndexedDB),
		ActivePlugins:    []string{"playground"},
		PlaygroundAuth:   true,
		BrowserStateless: true,
		LoginPolicy:      "playground",
		AdminPolicy:      "enabled",
	})
	if err != nil {
		t.Fatalf("NewWithOptions() error = %v", err)
	}
	t.Cleanup(func() {
		_ = module.Close(t.Context())
	})

	mux := http.NewServeMux()
	module.Routes(mux)

	admin := httptest.NewRecorder()
	adminReq := httptest.NewRequest(http.MethodGet, "/go-admin", nil)
	mux.ServeHTTP(admin, adminReq)
	if admin.Code != http.StatusSeeOther {
		t.Fatalf("admin status = %d, want %d", admin.Code, http.StatusSeeOther)
	}

	public := httptest.NewRecorder()
	publicReq := httptest.NewRequest(http.MethodGet, "/?preview_theme=blank&preview_preset=minimal", nil)
	mux.ServeHTTP(public, publicReq)
	if public.Code != http.StatusOK {
		t.Fatalf("public status = %d body = %s", public.Code, public.Body.String())
	}
	if body := public.Body.String(); body == "" || !strings.Contains(body, `data-gocms-theme="blank"`) {
		t.Fatalf("public preview body = %s", body)
	}

	rest := httptest.NewRecorder()
	restReq := httptest.NewRequest(http.MethodGet, "/go-json/go/v2/posts", nil)
	mux.ServeHTTP(rest, restReq)
	if rest.Code != http.StatusNotFound {
		t.Fatalf("rest status = %d, want %d", rest.Code, http.StatusNotFound)
	}
}
