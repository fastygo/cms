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
	"github.com/fastygo/cms/internal/platform/cmspanel"
	"github.com/fastygo/cms/internal/platform/runtimeprofile"
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
		"/static/css/app.css":        "/static/css/app.testhash123.css",
		"/static/js/admin-editor.js": "/static/js/admin-editor.testhash123.js",
		"/static/js/theme.js":        "/static/js/theme.testhash123.js",
		"/static/js/ui8kit.js":       "/static/js/ui8kit.testhash123.js",
		"/static/js/snapshots.js":    "/static/js/snapshots.testhash123.js",
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
	for _, expected := range []string{
		`data-gocms-editor-provider="tiptap-basic"`,
		`data-gocms-editor-field="content"`,
		`gocms-richtext-surface-host`,
		`gocms-editor-details`,
		`/static/js/admin-editor`,
	} {
		if !strings.Contains(newPage.Body.String(), expected) {
			t.Fatalf("expected content editor page to contain %q", expected)
		}
	}
	if strings.Contains(newPage.Body.String(), `contenteditable="true"`) {
		t.Fatalf("editor host must not render contenteditable; TipTap owns the ProseMirror editable node")
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

func TestAdminContentEditorShowsGeneratedMetadataFields(t *testing.T) {
	mux, closeFn := newAdminMux(t)
	defer closeFn()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/go-admin/posts/new", nil)
	req.Header.Set("Authorization", "Bearer admin-token")
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected editor page, got %d: %s", rec.Code, rec.Body.String())
	}
	for _, expected := range []string{
		`name="meta__seo_title"`,
		`name="meta__seo_description"`,
		`name="meta__seo_canonical_url"`,
		`name="meta__seo_noindex"`,
		`name="custom_meta_key"`,
		`name="custom_meta_value"`,
	} {
		if !strings.Contains(rec.Body.String(), expected) {
			t.Fatalf("expected editor to contain %q", expected)
		}
	}
	if strings.Contains(rec.Body.String(), `name="meta_key"`) {
		t.Fatalf("legacy generic metadata field should not be rendered")
	}
}

func TestAdminScreensRenderSinglePageDescription(t *testing.T) {
	mux, closeFn := newAdminMux(t)
	defer closeFn()

	cases := []struct {
		path        string
		description string
	}{
		{path: "/go-admin", description: "Manage GoCMS content, taxonomies, media, users, and headless delivery."},
		{path: "/go-admin/posts", description: "Create, edit, publish, schedule, trash, and restore content."},
		{path: "/go-admin/posts/new", description: "Create a draft and choose publish state."},
		{path: "/go-admin/pages", description: "Create, edit, publish, schedule, trash, and restore content."},
		{path: "/go-admin/content-types", description: "Manage built-in and custom content types."},
		{path: "/go-admin/taxonomies", description: "Manage taxonomy definitions and terms."},
		{path: "/go-admin/taxonomies/category/terms", description: "Manage taxonomy terms."},
		{path: "/go-admin/media", description: "Manage media metadata and featured media references."},
		{path: "/go-admin/menus", description: "Manage navigation menus."},
		{path: "/go-admin/users", description: "Manage users and account state."},
		{path: "/go-admin/authors", description: "Review public author projections."},
		{path: "/go-admin/capabilities", description: "Review capability groups enforced server-side."},
		{path: "/go-admin/settings", description: "Configure public site settings."},
		{path: "/go-admin/themes", description: "Inspect the active built-in theme contract and available template roles."},
		{path: "/go-admin/permalinks", description: "Configure public post and page routes."},
		{path: "/go-admin/headless", description: "Inspect API delivery mode and upcoming plugin state."},
		{path: "/go-admin/runtime", description: "Inspect the resolved preset, bootstrap provider, active plugins, and provider switch rules."},
	}

	for _, tc := range cases {
		t.Run(tc.path, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			req.Header.Set("Authorization", "Bearer admin-token")
			mux.ServeHTTP(rec, req)
			if rec.Code != http.StatusOK {
				t.Fatalf("expected %s to render, got %d: %s", tc.path, rec.Code, rec.Body.String())
			}
			if count := strings.Count(rec.Body.String(), tc.description); count != 1 {
				t.Fatalf("expected page description %q once on %s, got %d", tc.description, tc.path, count)
			}
		})
	}
}

func TestAdminRegistersCoreRoutesFromCMSPanelDescriptors(t *testing.T) {
	mux, closeFn := newAdminMux(t)
	defer closeFn()

	for _, page := range cmspanel.AdminPages() {
		for _, route := range page.Routes {
			if !strings.HasPrefix(route.Pattern, "GET ") {
				continue
			}
			t.Run(string(page.ID), func(t *testing.T) {
				path := strings.TrimPrefix(route.Pattern, "GET ")
				path = strings.ReplaceAll(path, "{type}", "category")
				rec := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodGet, path, nil)
				req.Header.Set("Authorization", "Bearer admin-token")
				mux.ServeHTTP(rec, req)
				if rec.Code != http.StatusOK {
					t.Fatalf("expected %s to render from cmspanel descriptor, got %d: %s", path, rec.Code, rec.Body.String())
				}
			})
		}
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

func TestAdminMediaMetadataFormAndValidation(t *testing.T) {
	mux, closeFn := newAdminMux(t)
	defer closeFn()

	page := httptest.NewRecorder()
	pageReq := httptest.NewRequest(http.MethodGet, "/go-admin/media", nil)
	pageReq.Header.Set("Authorization", "Bearer admin-token")
	mux.ServeHTTP(page, pageReq)
	if page.Code != http.StatusOK {
		t.Fatalf("expected media page, got %d: %s", page.Code, page.Body.String())
	}
	token := extractToken(t, page.Body.String())
	for _, expected := range []string{
		`name="provider"`,
		`name="provider_key"`,
		`name="provider_url"`,
		`name="provider_checksum"`,
		`name="provider_etag"`,
	} {
		if !strings.Contains(page.Body.String(), expected) {
			t.Fatalf("expected media form to contain %q", expected)
		}
	}

	invalid := url.Values{
		"action_token": {token},
		"id":           {"media-invalid"},
		"filename":     {"cover.txt"},
		"mime_type":    {"text/plain"},
		"public_url":   {"https://cdn.example.test/cover.txt"},
	}
	invalidRec := httptest.NewRecorder()
	invalidReq := httptest.NewRequest(http.MethodPost, "/go-admin/media", strings.NewReader(invalid.Encode()))
	invalidReq.Header.Set("Authorization", "Bearer admin-token")
	invalidReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	mux.ServeHTTP(invalidRec, invalidReq)
	if invalidRec.Code != http.StatusBadRequest {
		t.Fatalf("expected invalid media metadata to be rejected, got %d: %s", invalidRec.Code, invalidRec.Body.String())
	}

	valid := url.Values{
		"action_token":      {token},
		"id":                {"media-remote"},
		"filename":          {"cover.webp"},
		"mime_type":         {"image/webp"},
		"size_bytes":        {"2048"},
		"width":             {"1024"},
		"height":            {"512"},
		"public_url":        {"https://cdn.example.test/cover.webp"},
		"alt_text":          {"Remote cover"},
		"caption":           {"Stored by provider metadata only."},
		"provider":          {"s3"},
		"provider_key":      {"media/originals/cover.webp"},
		"provider_url":      {"https://bucket.example.test/cover.webp"},
		"provider_checksum": {"sha256:test"},
		"provider_etag":     {"etag-1"},
	}
	validRec := httptest.NewRecorder()
	validReq := httptest.NewRequest(http.MethodPost, "/go-admin/media", strings.NewReader(valid.Encode()))
	validReq.Header.Set("Authorization", "Bearer admin-token")
	validReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	mux.ServeHTTP(validRec, validReq)
	if validRec.Code != http.StatusSeeOther {
		t.Fatalf("expected valid media metadata redirect, got %d: %s", validRec.Code, validRec.Body.String())
	}
}

func TestAdminContentListControlsAndScreenPreferences(t *testing.T) {
	mux, closeFn := newAdminMux(t)
	defer closeFn()

	list := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/go-admin/posts?search=Published&filter_status=published&columns=title,status", nil)
	req.Header.Set("Authorization", "Bearer admin-token")
	mux.ServeHTTP(list, req)
	if list.Code != http.StatusOK {
		t.Fatalf("expected posts list, got %d: %s", list.Code, list.Body.String())
	}
	body := list.Body.String()
	for _, expected := range []string{
		`name="search"`,
		`name="filter_status"`,
		`name="sort"`,
		`name="bulk_action"`,
		`edit=content-post-published`,
		`value="title,status"`,
	} {
		if !strings.Contains(body, expected) {
			t.Fatalf("expected posts list to contain %q", expected)
		}
	}
	if !strings.Contains(body, "Published Post") {
		t.Fatalf("expected filtered list to include published post")
	}
	if strings.Contains(body, "Draft Post") {
		t.Fatalf("draft post should be filtered out by published status")
	}

	token := extractTokenForAction(t, body, "content-screen-options")
	savePrefs := httptest.NewRecorder()
	form := url.Values{
		"action_token": {token},
		"per_page":     {"100"},
		"columns":      {"title,status"},
		"return_to":    {"/go-admin/posts"},
	}
	saveReq := httptest.NewRequest(http.MethodPost, "/go-admin/preferences/posts", strings.NewReader(form.Encode()))
	saveReq.Header.Set("Authorization", "Bearer admin-token")
	saveReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	mux.ServeHTTP(savePrefs, saveReq)
	if savePrefs.Code != http.StatusSeeOther {
		t.Fatalf("expected screen preference redirect, got %d: %s", savePrefs.Code, savePrefs.Body.String())
	}

	updated := httptest.NewRecorder()
	updatedReq := httptest.NewRequest(http.MethodGet, "/go-admin/posts", nil)
	updatedReq.Header.Set("Authorization", "Bearer admin-token")
	mux.ServeHTTP(updated, updatedReq)
	if !strings.Contains(updated.Body.String(), `value="title,status"`) {
		t.Fatalf("expected saved columns preference to be rendered")
	}
}

func TestAdminUsersQuickEditAndBulkStatus(t *testing.T) {
	mux, closeFn := newAdminMux(t)
	defer closeFn()

	page := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/go-admin/users?edit=author-1", nil)
	req.Header.Set("Authorization", "Bearer admin-token")
	mux.ServeHTTP(page, req)
	if page.Code != http.StatusOK {
		t.Fatalf("expected users page, got %d: %s", page.Code, page.Body.String())
	}
	body := page.Body.String()
	for _, expected := range []string{
		`name="display_name"`,
		`name="status"`,
		`status:suspended`,
		`edit=author-1`,
	} {
		if !strings.Contains(body, expected) {
			t.Fatalf("expected users page to contain %q", expected)
		}
	}
	token := extractToken(t, body)

	quickEdit := httptest.NewRecorder()
	quickForm := url.Values{
		"action_token": {token},
		"id":           {"author-1"},
		"login":        {"jane"},
		"display_name": {"Jane Editor"},
		"email":        {"jane@example.test"},
		"status":       {"suspended"},
		"return_to":    {"/go-admin/users"},
	}
	quickReq := httptest.NewRequest(http.MethodPost, "/go-admin/users", strings.NewReader(quickForm.Encode()))
	quickReq.Header.Set("Authorization", "Bearer admin-token")
	quickReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	mux.ServeHTTP(quickEdit, quickReq)
	if quickEdit.Code != http.StatusSeeOther {
		t.Fatalf("expected quick edit redirect, got %d: %s", quickEdit.Code, quickEdit.Body.String())
	}

	suspended := httptest.NewRecorder()
	suspendedReq := httptest.NewRequest(http.MethodGet, "/go-admin/users?filter_status=suspended", nil)
	suspendedReq.Header.Set("Authorization", "Bearer admin-token")
	mux.ServeHTTP(suspended, suspendedReq)
	if !strings.Contains(suspended.Body.String(), "Jane Editor") || !strings.Contains(suspended.Body.String(), "suspended") {
		t.Fatalf("expected quick edited user to be suspended: %s", suspended.Body.String())
	}

	bulk := httptest.NewRecorder()
	bulkForm := url.Values{
		"action_token": {token},
		"bulk_action":  {"status:active"},
		"selected_id":  {"author-1"},
		"return_to":    {"/go-admin/users"},
	}
	bulkReq := httptest.NewRequest(http.MethodPost, "/go-admin/users", strings.NewReader(bulkForm.Encode()))
	bulkReq.Header.Set("Authorization", "Bearer admin-token")
	bulkReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	mux.ServeHTTP(bulk, bulkReq)
	if bulk.Code != http.StatusSeeOther {
		t.Fatalf("expected bulk update redirect, got %d: %s", bulk.Code, bulk.Body.String())
	}

	active := httptest.NewRecorder()
	activeReq := httptest.NewRequest(http.MethodGet, "/go-admin/users?filter_status=active", nil)
	activeReq.Header.Set("Authorization", "Bearer admin-token")
	mux.ServeHTTP(active, activeReq)
	if !strings.Contains(active.Body.String(), "Jane Editor") || !strings.Contains(active.Body.String(), "active") {
		t.Fatalf("expected bulk restored user to active: %s", active.Body.String())
	}
}

func TestAdminUserSecurityActionsAndRuntimeDiagnostics(t *testing.T) {
	mux, closeFn := newAdminMux(t)
	defer closeFn()

	page := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/go-admin/users?edit=admin", nil)
	req.Header.Set("Authorization", "Bearer admin-token")
	mux.ServeHTTP(page, req)
	if page.Code != http.StatusOK {
		t.Fatalf("expected users security page, got %d: %s", page.Code, page.Body.String())
	}
	token := extractToken(t, page.Body.String())

	createAppToken := httptest.NewRecorder()
	appTokenForm := url.Values{
		"action_token":           {token},
		"id":                     {"admin"},
		"security_action":        {"create_app_token"},
		"app_token_name":         {"CLI"},
		"app_token_ttl_hours":    {"24"},
		"app_token_capabilities": {"content.read_private"},
		"return_to":              {"/go-admin/users"},
	}
	appTokenReq := httptest.NewRequest(http.MethodPost, "/go-admin/users", strings.NewReader(appTokenForm.Encode()))
	appTokenReq.Header.Set("Authorization", "Bearer admin-token")
	appTokenReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	mux.ServeHTTP(createAppToken, appTokenReq)
	if createAppToken.Code != http.StatusOK {
		t.Fatalf("expected app token result page, got %d: %s", createAppToken.Code, createAppToken.Body.String())
	}
	if !strings.Contains(createAppToken.Body.String(), "App token") || !strings.Contains(createAppToken.Body.String(), "copy-now") {
		t.Fatalf("expected app token page to show one-time token: %s", createAppToken.Body.String())
	}

	createRecovery := httptest.NewRecorder()
	recoveryForm := url.Values{
		"action_token":        {token},
		"id":                  {"admin"},
		"security_action":     {"generate_recovery_codes"},
		"recovery_code_count": {"2"},
		"return_to":           {"/go-admin/users"},
	}
	recoveryReq := httptest.NewRequest(http.MethodPost, "/go-admin/users", strings.NewReader(recoveryForm.Encode()))
	recoveryReq.Header.Set("Authorization", "Bearer admin-token")
	recoveryReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	mux.ServeHTTP(createRecovery, recoveryReq)
	if createRecovery.Code != http.StatusOK {
		t.Fatalf("expected recovery result page, got %d: %s", createRecovery.Code, createRecovery.Body.String())
	}
	if !strings.Contains(createRecovery.Body.String(), "Recovery code") {
		t.Fatalf("expected recovery result page to list codes: %s", createRecovery.Body.String())
	}

	runtime := httptest.NewRecorder()
	runtimeReq := httptest.NewRequest(http.MethodGet, "/go-admin/runtime", nil)
	runtimeReq.Header.Set("Authorization", "Bearer admin-token")
	mux.ServeHTTP(runtime, runtimeReq)
	if runtime.Code != http.StatusOK {
		t.Fatalf("expected runtime page, got %d: %s", runtime.Code, runtime.Body.String())
	}
	body := runtime.Body.String()
	if !strings.Contains(body, "Health: Database connectivity") {
		t.Fatalf("expected runtime page to include health rows: %s", body)
	}
	if !strings.Contains(body, "Audit: auth.app_token.create") || !strings.Contains(body, "Audit: auth.recovery_codes.generate") {
		t.Fatalf("expected runtime page to include recent audit rows: %s", body)
	}
}

func TestAdminCoreSectionCreateWorkflows(t *testing.T) {
	mux, closeFn := newAdminMux(t)
	defer closeFn()

	cases := []struct {
		name string
		path string
		form url.Values
	}{
		{
			name: "content-types",
			path: "/go-admin/content-types",
			form: url.Values{"id": {"case-study"}, "label": {"Case studies"}},
		},
		{
			name: "terms",
			path: "/go-admin/taxonomies/category/terms",
			form: url.Values{"id": {"featured"}, "name": {"Featured"}, "slug": {"featured"}},
		},
		{
			name: "media",
			path: "/go-admin/media",
			form: url.Values{"id": {"media-admin-test"}, "filename": {"admin-test.jpg"}, "mime_type": {"image/jpeg"}, "public_url": {"/media/admin-test.jpg"}},
		},
		{
			name: "menus",
			path: "/go-admin/menus",
			form: url.Values{"id": {"secondary"}, "name": {"Secondary"}, "location": {"secondary"}},
		},
		{
			name: "users",
			path: "/go-admin/users",
			form: url.Values{"id": {"editor-2"}, "login": {"editor2"}, "display_name": {"Editor Two"}, "email": {"editor2@example.test"}},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			page := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			req.Header.Set("Authorization", "Bearer admin-token")
			mux.ServeHTTP(page, req)
			if page.Code != http.StatusOK {
				t.Fatalf("expected %s page, got %d: %s", tc.name, page.Code, page.Body.String())
			}

			tc.form.Set("action_token", extractToken(t, page.Body.String()))
			create := httptest.NewRecorder()
			createReq := httptest.NewRequest(http.MethodPost, tc.path, strings.NewReader(tc.form.Encode()))
			createReq.Header.Set("Authorization", "Bearer admin-token")
			createReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			mux.ServeHTTP(create, createReq)
			if create.Code != http.StatusSeeOther {
				t.Fatalf("expected %s redirect, got %d: %s", tc.name, create.Code, create.Body.String())
			}
		})
	}
}

func TestAdminThemesAndPermalinksAreCapabilityGated(t *testing.T) {
	mux, closeFn := newAdminMux(t)
	defer closeFn()

	viewerThemes := httptest.NewRecorder()
	viewerThemesReq := httptest.NewRequest(http.MethodGet, "/go-admin/themes", nil)
	viewerThemesReq.Header.Set("Authorization", "Bearer viewer-token")
	mux.ServeHTTP(viewerThemes, viewerThemesReq)
	if viewerThemes.Code != http.StatusForbidden {
		t.Fatalf("expected viewer themes forbidden, got %d", viewerThemes.Code)
	}

	adminThemes := httptest.NewRecorder()
	adminThemesReq := httptest.NewRequest(http.MethodGet, "/go-admin/themes", nil)
	adminThemesReq.Header.Set("Authorization", "Bearer admin-token")
	mux.ServeHTTP(adminThemes, adminThemesReq)
	if adminThemes.Code != http.StatusOK {
		t.Fatalf("expected admin themes page, got %d: %s", adminThemes.Code, adminThemes.Body.String())
	}
	body := adminThemes.Body.String()
	for _, expected := range []string{`data-gocms-screen="themes"`, "GoCMS Default", "front", "active", "bold-tech", "preview_theme"} {
		if !strings.Contains(body, expected) {
			t.Fatalf("expected themes page to contain %q", expected)
		}
	}
	themesToken := extractToken(t, body)

	saveThemes := httptest.NewRecorder()
	saveThemesForm := url.Values{
		"action_token":         {themesToken},
		"theme_active":         {"blank"},
		"theme_style_preset":   {"minimal"},
		"theme_preview":        {"gocms-default"},
		"theme_preview_preset": {"bold-tech"},
	}
	saveThemesReq := httptest.NewRequest(http.MethodPost, "/go-admin/themes", strings.NewReader(saveThemesForm.Encode()))
	saveThemesReq.Header.Set("Authorization", "Bearer admin-token")
	saveThemesReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	mux.ServeHTTP(saveThemes, saveThemesReq)
	if saveThemes.Code != http.StatusSeeOther {
		t.Fatalf("expected themes redirect, got %d: %s", saveThemes.Code, saveThemes.Body.String())
	}

	updatedThemes := httptest.NewRecorder()
	updatedThemesReq := httptest.NewRequest(http.MethodGet, "/go-admin/themes", nil)
	updatedThemesReq.Header.Set("Authorization", "Bearer admin-token")
	mux.ServeHTTP(updatedThemes, updatedThemesReq)
	for _, expected := range []string{"blank", "minimal", "bold-tech"} {
		if !strings.Contains(updatedThemes.Body.String(), expected) {
			t.Fatalf("expected updated themes page to contain %q", expected)
		}
	}

	viewerPermalinks := httptest.NewRecorder()
	viewerPermalinksReq := httptest.NewRequest(http.MethodGet, "/go-admin/permalinks", nil)
	viewerPermalinksReq.Header.Set("Authorization", "Bearer viewer-token")
	mux.ServeHTTP(viewerPermalinks, viewerPermalinksReq)
	if viewerPermalinks.Code != http.StatusForbidden {
		t.Fatalf("expected viewer permalinks forbidden, got %d", viewerPermalinks.Code)
	}

	adminPermalinks := httptest.NewRecorder()
	adminPermalinksReq := httptest.NewRequest(http.MethodGet, "/go-admin/permalinks", nil)
	adminPermalinksReq.Header.Set("Authorization", "Bearer admin-token")
	mux.ServeHTTP(adminPermalinks, adminPermalinksReq)
	if adminPermalinks.Code != http.StatusOK {
		t.Fatalf("expected admin permalinks page, got %d: %s", adminPermalinks.Code, adminPermalinks.Body.String())
	}
	token := extractToken(t, adminPermalinks.Body.String())

	save := httptest.NewRecorder()
	form := url.Values{
		"action_token": {token},
		"post_pattern": {"/archives/%id%/"},
		"page_pattern": {"/pages/{slug}/"},
	}
	saveReq := httptest.NewRequest(http.MethodPost, "/go-admin/permalinks", strings.NewReader(form.Encode()))
	saveReq.Header.Set("Authorization", "Bearer admin-token")
	saveReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	mux.ServeHTTP(save, saveReq)
	if save.Code != http.StatusSeeOther {
		t.Fatalf("expected permalinks redirect, got %d: %s", save.Code, save.Body.String())
	}

	updated := httptest.NewRecorder()
	updatedReq := httptest.NewRequest(http.MethodGet, "/go-admin/permalinks", nil)
	updatedReq.Header.Set("Authorization", "Bearer admin-token")
	mux.ServeHTTP(updated, updatedReq)
	for _, expected := range []string{"/archives/%id%/", "/pages/{slug}/"} {
		if !strings.Contains(updated.Body.String(), expected) {
			t.Fatalf("expected permalinks page to contain %q", expected)
		}
	}
}

func TestAdminLoadsPluginActionsAndSnapshotExport(t *testing.T) {
	mux, closeFn := newAdminMux(t)
	defer closeFn()

	settings := httptest.NewRecorder()
	settingsReq := httptest.NewRequest(http.MethodGet, "/go-admin/settings", nil)
	settingsReq.Header.Set("Authorization", "Bearer admin-token")
	mux.ServeHTTP(settings, settingsReq)
	if settings.Code != http.StatusOK {
		t.Fatalf("expected settings page, got %d: %s", settings.Code, settings.Body.String())
	}
	if !strings.Contains(settings.Body.String(), "/go-admin/plugins/json-import-export/export") {
		t.Fatalf("expected settings page to expose json export action")
	}
	if !strings.Contains(settings.Body.String(), "/static/js/snapshots") {
		t.Fatalf("expected settings page to load snapshots plugin asset")
	}

	export := httptest.NewRecorder()
	exportReq := httptest.NewRequest(http.MethodGet, "/go-admin/plugins/json-import-export/export", nil)
	exportReq.Header.Set("Authorization", "Bearer admin-token")
	mux.ServeHTTP(export, exportReq)
	if export.Code != http.StatusOK {
		t.Fatalf("expected export route, got %d: %s", export.Code, export.Body.String())
	}
	if contentType := export.Header().Get("Content-Type"); !strings.Contains(contentType, "application/json") {
		t.Fatalf("Content-Type = %q, want application/json", contentType)
	}
}

func TestAdminShowsRuntimeStatusFromResolvedProviders(t *testing.T) {
	mux, closeFn := newAdminMux(t)
	defer closeFn()

	status := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/go-admin/runtime", nil)
	req.Header.Set("Authorization", "Bearer admin-token")
	mux.ServeHTTP(status, req)
	if status.Code != http.StatusOK {
		t.Fatalf("expected runtime status, got %d: %s", status.Code, status.Body.String())
	}
	body := status.Body.String()
	for _, expected := range []string{
		`data-gocms-screen="runtime"`,
		"Runtime status",
		"Deployment profile",
		"Content provider",
		"json-import-export",
		"Provider switch rule",
		"restart",
	} {
		if !strings.Contains(body, expected) {
			t.Fatalf("expected runtime status body to contain %q", expected)
		}
	}
	if !strings.Contains(body, `/go-admin/runtime`) {
		t.Fatalf("expected registry-driven runtime navigation item")
	}
}

func TestAdminPolicyCanDisableDevBearer(t *testing.T) {
	module, err := cms.NewWithOptions(cms.Options{
		DataSource:      "file:" + strings.ReplaceAll(t.Name(), "/", "-") + "?mode=memory&cache=shared",
		SessionKey:      "admin-test-session-secret",
		SeedFixtures:    true,
		RuntimeProfile:  string(runtimeprofile.RuntimeProfileFull),
		StorageProfile:  string(runtimeprofile.StorageProfileSQLite),
		EnableDevBearer: false,
		LoginPolicy:     "disabled",
		AdminPolicy:     "enabled",
		Preset:          "full",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = module.Close(t.Context())
	}()

	mux := http.NewServeMux()
	module.Routes(mux)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/go-admin", nil)
	req.Header.Set("Authorization", "Bearer admin-token")
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("expected disabled dev bearer to redirect to login, got %d", rec.Code)
	}
}

func TestExternalLoginPolicyDoesNotFallbackToFixtureLogin(t *testing.T) {
	module, err := cms.NewWithOptions(cms.Options{
		DataSource:      "file:" + strings.ReplaceAll(t.Name(), "/", "-") + "?mode=memory&cache=shared",
		SessionKey:      "admin-test-session-secret",
		SeedFixtures:    true,
		RuntimeProfile:  string(runtimeprofile.RuntimeProfileFull),
		StorageProfile:  string(runtimeprofile.StorageProfileSQLite),
		EnableDevBearer: false,
		LoginPolicy:     "external",
		AdminPolicy:     "enabled",
		Preset:          "full",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = module.Close(t.Context())
	}()

	mux := http.NewServeMux()
	module.Routes(mux)
	loginPage := httptest.NewRecorder()
	mux.ServeHTTP(loginPage, httptest.NewRequest(http.MethodGet, "/go-login", nil))
	token := extractToken(t, loginPage.Body.String())

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
	if login.Code == http.StatusSeeOther {
		t.Fatalf("external login policy must not accept fixture credentials")
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

func extractTokenForAction(t *testing.T, body string, action string) string {
	t.Helper()
	pattern := regexp.MustCompile(`data-gocms-action="` + regexp.QuoteMeta(action) + `".*?name="action_token"[^>]*value="([^"]+)"`)
	match := pattern.FindStringSubmatch(strings.ReplaceAll(body, "\n", " "))
	if len(match) < 2 {
		t.Fatalf("action token for %q not found in body: %s", action, body)
	}
	return match[1]
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
