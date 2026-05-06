package cms

import (
	"net/http"
	"net/http/httptest"
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
