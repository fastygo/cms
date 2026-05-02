package admin_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/fastygo/cms/internal/infra/features/cms"
)

func TestAdminAuthFlow(t *testing.T) {
	mux, closeFn := newAdminMux(t)
	defer closeFn()

	blocked := httptest.NewRecorder()
	mux.ServeHTTP(blocked, httptest.NewRequest(http.MethodGet, "/go-admin", nil))
	if blocked.Code != http.StatusSeeOther {
		t.Fatalf("expected unauthenticated redirect, got %d", blocked.Code)
	}

	loginPage := httptest.NewRecorder()
	mux.ServeHTTP(loginPage, httptest.NewRequest(http.MethodGet, "/go-login", nil))
	token := extractToken(t, loginPage.Body.String())
	if !strings.Contains(loginPage.Body.String(), `data-gocms-screen="login"`) {
		t.Fatalf("expected login screen marker")
	}
	if !strings.Contains(strings.ToLower(loginPage.Body.String()), "<!doctype html>") {
		t.Fatalf("expected login page to be a complete HTML document")
	}
	if !strings.Contains(loginPage.Body.String(), "<style>") {
		t.Fatalf("expected login page to include self-contained inline styles")
	}
	if strings.Contains(loginPage.Body.String(), `rel="stylesheet"`) {
		t.Fatalf("login page must not depend on external stylesheets")
	}

	form := url.Values{
		"action_token": {token},
		"email":        {"admin@example.test"},
		"password":     {"admin"},
		"return_to":    {"/go-admin"},
	}
	login := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/go-login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	mux.ServeHTTP(login, req)
	if login.Code != http.StatusSeeOther {
		t.Fatalf("expected login redirect, got %d: %s", login.Code, login.Body.String())
	}

	dashboard := httptest.NewRecorder()
	dashReq := httptest.NewRequest(http.MethodGet, "/go-admin", nil)
	for _, cookie := range login.Result().Cookies() {
		dashReq.AddCookie(cookie)
	}
	mux.ServeHTTP(dashboard, dashReq)
	if dashboard.Code != http.StatusOK {
		t.Fatalf("expected dashboard, got %d: %s", dashboard.Code, dashboard.Body.String())
	}
	if !strings.Contains(dashboard.Body.String(), `data-gocms-screen="dashboard"`) {
		t.Fatalf("expected dashboard screen marker")
	}
}

func TestAdminUsesVersionedAssetsWhenManifestExists(t *testing.T) {
	mux, closeFn := newAdminMux(t)
	defer closeFn()

	manifest := map[string]string{
		"/static/css/app.css":  "/static/css/app.testhash123.css",
		"/static/js/theme.js":  "/static/js/theme.testhash123.js",
		"/static/js/ui8kit.js": "/static/js/ui8kit.testhash123.js",
	}
	writeManifest(t, manifest)

	dashboard := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/go-admin", nil)
	req.Header.Set("Authorization", "Bearer admin-token")
	mux.ServeHTTP(dashboard, req)

	if dashboard.Code != http.StatusOK {
		t.Fatalf("expected dashboard, got %d: %s", dashboard.Code, dashboard.Body.String())
	}
	for _, path := range manifest {
		if !strings.Contains(dashboard.Body.String(), path) {
			t.Fatalf("expected dashboard to reference versioned asset %q", path)
		}
	}
}

func TestAdminContentWorkflowAndCapabilityChecks(t *testing.T) {
	mux, closeFn := newAdminMux(t)
	defer closeFn()

	viewerNew := httptest.NewRecorder()
	viewerReq := httptest.NewRequest(http.MethodGet, "/go-admin/posts/new", nil)
	viewerReq.Header.Set("Authorization", "Bearer viewer-token")
	mux.ServeHTTP(viewerNew, viewerReq)
	if viewerNew.Code != http.StatusForbidden {
		t.Fatalf("expected viewer create screen forbidden, got %d", viewerNew.Code)
	}

	newPage := httptest.NewRecorder()
	adminReq := httptest.NewRequest(http.MethodGet, "/go-admin/posts/new", nil)
	adminReq.Header.Set("Authorization", "Bearer admin-token")
	mux.ServeHTTP(newPage, adminReq)
	if newPage.Code != http.StatusOK {
		t.Fatalf("expected new post page, got %d: %s", newPage.Code, newPage.Body.String())
	}
	token := extractToken(t, newPage.Body.String())

	form := url.Values{
		"action_token": {token},
		"title":        {"Admin Test Post"},
		"slug":         {"admin-test-post"},
		"content":      {"Created from admin workflow test."},
		"excerpt":      {"Admin excerpt"},
		"author_id":    {"author-1"},
		"status":       {"published"},
	}
	create := httptest.NewRecorder()
	createReq := httptest.NewRequest(http.MethodPost, "/go-admin/posts", strings.NewReader(form.Encode()))
	createReq.Header.Set("Authorization", "Bearer admin-token")
	createReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	mux.ServeHTTP(create, createReq)
	if create.Code != http.StatusSeeOther {
		t.Fatalf("expected create redirect, got %d: %s", create.Code, create.Body.String())
	}

	list := httptest.NewRecorder()
	listReq := httptest.NewRequest(http.MethodGet, "/go-admin/posts", nil)
	listReq.Header.Set("Authorization", "Bearer admin-token")
	mux.ServeHTTP(list, listReq)
	if !strings.Contains(list.Body.String(), "Admin Test Post") {
		t.Fatalf("expected created post in admin list")
	}

	viewerCreate := httptest.NewRecorder()
	viewerPost := httptest.NewRequest(http.MethodPost, "/go-admin/posts", strings.NewReader(form.Encode()))
	viewerPost.Header.Set("Authorization", "Bearer viewer-token")
	viewerPost.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	mux.ServeHTTP(viewerCreate, viewerPost)
	if viewerCreate.Code != http.StatusForbidden {
		t.Fatalf("expected viewer direct POST forbidden, got %d", viewerCreate.Code)
	}
}

func TestAdminTaxonomyAndSettingsWorkflows(t *testing.T) {
	mux, closeFn := newAdminMux(t)
	defer closeFn()

	taxonomies := httptest.NewRecorder()
	taxReq := httptest.NewRequest(http.MethodGet, "/go-admin/taxonomies", nil)
	taxReq.Header.Set("Authorization", "Bearer admin-token")
	mux.ServeHTTP(taxonomies, taxReq)
	token := extractToken(t, taxonomies.Body.String())

	form := url.Values{"action_token": {token}, "type": {"topic"}, "label": {"Topics"}, "mode": {"flat"}}
	create := httptest.NewRecorder()
	createReq := httptest.NewRequest(http.MethodPost, "/go-admin/taxonomies", strings.NewReader(form.Encode()))
	createReq.Header.Set("Authorization", "Bearer admin-token")
	createReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	mux.ServeHTTP(create, createReq)
	if create.Code != http.StatusSeeOther {
		body, _ := io.ReadAll(create.Result().Body)
		t.Fatalf("expected taxonomy redirect, got %d: %s", create.Code, string(body))
	}

	settings := httptest.NewRecorder()
	settingsReq := httptest.NewRequest(http.MethodGet, "/go-admin/settings", nil)
	settingsReq.Header.Set("Authorization", "Bearer admin-token")
	mux.ServeHTTP(settings, settingsReq)
	settingsToken := extractToken(t, settings.Body.String())

	save := httptest.NewRecorder()
	saveForm := url.Values{"action_token": {settingsToken}, "site_title": {"GoCMS Test"}, "public_rendering": {"disabled"}}
	saveReq := httptest.NewRequest(http.MethodPost, "/go-admin/settings", strings.NewReader(saveForm.Encode()))
	saveReq.Header.Set("Authorization", "Bearer admin-token")
	saveReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	mux.ServeHTTP(save, saveReq)
	if save.Code != http.StatusSeeOther {
		t.Fatalf("expected settings redirect, got %d: %s", save.Code, save.Body.String())
	}
}

func newAdminMux(t *testing.T) (*http.ServeMux, func()) {
	t.Helper()
	module, err := cms.New("file:"+strings.ReplaceAll(t.Name(), "/", "-")+"?mode=memory&cache=shared", "admin-test-session-secret", true)
	if err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	module.Routes(mux)
	return mux, func() {
		_ = module.Close(t.Context())
	}
}

func extractToken(t *testing.T, body string) string {
	t.Helper()
	matches := regexp.MustCompile(`name="action_token"[^>]*value="([^"]+)"`).FindAllStringSubmatch(body, -1)
	if len(matches) == 0 {
		t.Fatalf("action token not found in body: %s", body)
	}
	return matches[len(matches)-1][1]
}

func writeManifest(t *testing.T, manifest map[string]string) {
	t.Helper()
	const path = "../../../web/static/asset-manifest.json"
	previous, readErr := os.ReadFile(path)
	t.Cleanup(func() {
		if readErr == nil {
			_ = os.WriteFile(path, previous, 0o644)
			return
		}
		_ = os.Remove(path)
	})
	payload, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, payload, 0o644); err != nil {
		t.Fatal(err)
	}
}
