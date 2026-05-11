package cmspanel

import (
	"testing"

	appmeta "github.com/fastygo/cms/internal/application/meta"
	"github.com/fastygo/cms/internal/domain/authz"
	domaincontent "github.com/fastygo/cms/internal/domain/content"
	domainmeta "github.com/fastygo/cms/internal/domain/meta"
	"github.com/fastygo/panel"
)

func TestContentResourcesDeclarePostsAndPages(t *testing.T) {
	resources := ContentResources()
	if len(resources) != 2 {
		t.Fatalf("resources = %d, want posts and pages", len(resources))
	}

	assertContentResource(t, resources[0], "posts", domaincontent.KindPost, "/go-admin/posts", "file", 1)
	assertContentResource(t, resources[1], "pages", domaincontent.KindPage, "/go-admin/pages", "book", 2)
}

func TestPublicProjectionContractsCoverPublicFrontendSurfaces(t *testing.T) {
	contracts := PublicProjectionContracts()
	if len(contracts) != 5 {
		t.Fatalf("contracts = %d, want posts, pages, media, taxonomies, and menus", len(contracts))
	}

	byID := make(map[string]PublicProjectionContract, len(contracts))
	for _, contract := range contracts {
		byID[contract.ID] = contract
		if contract.REST.CollectionPath == "" || contract.GraphQL.CollectionField == "" {
			t.Fatalf("contract %q is missing REST or GraphQL collection metadata: %+v", contract.ID, contract)
		}
		if len(contract.REST.RequiredFields) == 0 || len(contract.GraphQL.RequiredFields) == 0 {
			t.Fatalf("contract %q is missing required projection fields: %+v", contract.ID, contract)
		}
	}
	if byID["posts"].Kind != domaincontent.KindPost || byID["posts"].ResourceID != "posts" {
		t.Fatalf("posts contract = %+v, want cmspanel posts resource projection", byID["posts"])
	}
	if byID["pages"].Kind != domaincontent.KindPage || byID["pages"].ResourceID != "pages" {
		t.Fatalf("pages contract = %+v, want cmspanel pages resource projection", byID["pages"])
	}
	for _, id := range []string{"media", "taxonomies", "menus"} {
		if byID[id].Implementation == "" {
			t.Fatalf("contract %q should document its current extraction state", id)
		}
	}
}

func TestAdminPagesDescribeCoreCMSScreens(t *testing.T) {
	pages := AdminPages()
	expected := map[string]struct {
		path       string
		capability authz.Capability
		hasForm    bool
		hasTable   bool
		nav        bool
	}{
		"dashboard":     {path: "/go-admin", nav: true},
		"content-types": {path: "/go-admin/content-types", capability: authz.CapabilitySettingsManage, hasForm: true, hasTable: true, nav: true},
		"taxonomies":    {path: "/go-admin/taxonomies", capability: authz.CapabilityTaxonomiesManage, hasForm: true, hasTable: true, nav: true},
		"terms":         {path: "/go-admin/taxonomies/{type}/terms", capability: authz.CapabilityTaxonomiesManage, hasForm: true, hasTable: true},
		"media":         {path: "/go-admin/media", capability: authz.CapabilityMediaUpload, hasForm: true, hasTable: true, nav: true},
		"menus":         {path: "/go-admin/menus", capability: authz.CapabilityMenusManage, hasForm: true, hasTable: true, nav: true},
		"users":         {path: "/go-admin/users", capability: authz.CapabilityUsersManage, hasForm: true, hasTable: true, nav: true},
		"authors":       {path: "/go-admin/authors", capability: authz.CapabilityContentReadPrivate, hasTable: true, nav: true},
		"capabilities":  {path: "/go-admin/capabilities", capability: authz.CapabilityRolesManage, hasTable: true, nav: true},
		"settings":      {path: "/go-admin/settings", capability: authz.CapabilitySettingsManage, hasForm: true, nav: true},
		"themes":        {path: "/go-admin/themes", capability: authz.CapabilityThemesManage, hasForm: true, hasTable: true, nav: true},
		"permalinks":    {path: "/go-admin/permalinks", capability: authz.CapabilitySettingsManage, hasForm: true, nav: true},
		"headless":      {path: "/go-admin/headless", capability: authz.CapabilitySettingsManage, hasTable: true, nav: true},
		"runtime":       {path: "/go-admin/runtime", capability: authz.CapabilitySettingsManage, hasTable: true, nav: true},
	}
	if len(pages) != len(expected) {
		t.Fatalf("pages = %d, want %d core CMS admin pages", len(pages), len(expected))
	}
	for _, page := range pages {
		want, ok := expected[string(page.ID)]
		if !ok {
			t.Fatalf("unexpected page descriptor %q", page.ID)
		}
		if page.Path != want.path || page.Capability != want.capability {
			t.Fatalf("page %q path/capability = %q/%q, want %q/%q", page.ID, page.Path, page.Capability, want.path, want.capability)
		}
		if want.hasForm && len(page.Form.Fields) == 0 {
			t.Fatalf("page %q should describe a form", page.ID)
		}
		if want.hasTable && len(page.Table.Columns) == 0 {
			t.Fatalf("page %q should describe a table", page.ID)
		}
		if want.nav && page.Navigation.Path == "" {
			t.Fatalf("page %q should contribute navigation", page.ID)
		}
		if !want.nav && page.Navigation.Path != "" {
			t.Fatalf("page %q should not contribute top-level navigation", page.ID)
		}
		if len(page.Routes) == 0 {
			t.Fatalf("page %q should describe admin routes", page.ID)
		}
		if err := page.Validate(); err != nil {
			t.Fatalf("page %q validation error = %v", page.ID, err)
		}
	}
}

func assertContentResource(t *testing.T, resource ContentResource, id string, kind domaincontent.Kind, prefix string, icon string, order int) {
	t.Helper()

	if string(resource.ID) != id || resource.Kind != kind {
		t.Fatalf("resource identity = %q/%q, want %q/%q", resource.ID, resource.Kind, id, kind)
	}
	if resource.BasePath != prefix || resource.Label == "" || resource.Singular == "" || resource.Plural == "" {
		t.Fatalf("resource descriptor = %+v, want base path and labels", resource.Resource)
	}
	if resource.Navigation.ID != id || resource.Navigation.Path != prefix || resource.Navigation.Icon != icon || resource.Navigation.Order != order {
		t.Fatalf("resource menu = %+v, want id=%q path=%q icon=%q order=%d", resource.Navigation, id, prefix, icon, order)
	}
	if resource.Navigation.Capability != authz.CapabilityContentReadPrivate {
		t.Fatalf("menu capability = %q, want content read private", resource.Navigation.Capability)
	}
	if len(resource.Table.Columns) != 4 {
		t.Fatalf("table columns = %v, want title/slug/status/author", resource.Table.Columns)
	}
	if len(resource.Form.Fields) != 8 {
		t.Fatalf("form fields = %v, want content editor fields", resource.Form.Fields)
	}
	if field := fieldByID(resource.Form.Fields, "content"); field.Type != panel.FieldRichText {
		t.Fatalf("content field = %+v, want richtext", field)
	}
	if len(resource.Actions) != 1 || resource.Actions[0].ID != "create" || resource.Actions[0].Capability != authz.CapabilityContentCreate {
		t.Fatalf("resource actions = %v, want create action", resource.Actions)
	}
	if err := resource.Validate(); err != nil {
		t.Fatalf("resource validation error = %v", err)
	}

	routes := map[panel.ResourceRouteRole]panel.ResourceRoute[authz.Capability]{}
	for _, route := range resource.Routes {
		routes[route.Role] = route
	}
	expected := map[panel.ResourceRouteRole]struct {
		pattern    string
		capability authz.Capability
	}{
		panel.RouteIndex:  {"GET " + prefix, authz.CapabilityContentReadPrivate},
		panel.RouteNew:    {"GET " + prefix + "/new", authz.CapabilityContentCreate},
		panel.RouteCreate: {"POST " + prefix, authz.CapabilityContentCreate},
		panel.RouteEdit:   {"GET " + prefix + "/{id}/edit", authz.CapabilityContentEdit},
		panel.RouteUpdate: {"POST " + prefix + "/{id}", authz.CapabilityContentEdit},
		panel.RouteDelete: {"POST " + prefix + "/{id}/trash", authz.CapabilityContentDelete},
	}
	if len(routes) != len(expected) {
		t.Fatalf("routes = %v, want %d route roles", routes, len(expected))
	}
	for role, want := range expected {
		got, ok := routes[role]
		if !ok {
			t.Fatalf("missing route role %q", role)
		}
		if got.Pattern != want.pattern || got.Capability != want.capability {
			t.Fatalf("route %q = %+v, want pattern=%q capability=%q", role, got, want.pattern, want.capability)
		}
	}
}

func TestMetadataFieldsGenerateRegisteredAndFallbackInputs(t *testing.T) {
	fields := MetadataFields([]domainmeta.Definition{
		{
			Key:       "seo_title",
			Label:     "SEO title",
			Owner:     "test",
			Scope:     domainmeta.ScopeContent,
			Type:      domainmeta.ValueTypeString,
			FieldHint: domainmeta.FieldHintText,
			Public:    true,
		},
		{
			Key:       "seo_noindex",
			Label:     "Noindex",
			Owner:     "test",
			Scope:     domainmeta.ScopeContent,
			Type:      domainmeta.ValueTypeBoolean,
			FieldHint: domainmeta.FieldHintCheckbox,
			Public:    true,
		},
	})
	if got := fieldByID(fields, appmeta.FormFieldName("seo_title")); got.Type != panel.FieldText {
		t.Fatalf("seo_title field = %+v", got)
	}
	if got := fieldByID(fields, appmeta.FormFieldName("seo_noindex")); got.Type != panel.FieldBoolean {
		t.Fatalf("seo_noindex field = %+v", got)
	}
	if got := fieldByID(fields, "custom_meta_key"); got.Type != panel.FieldText {
		t.Fatalf("custom_meta_key field = %+v", got)
	}
}

func fieldByID(fields []panel.Field, id string) panel.Field {
	for _, field := range fields {
		if field.ID == id {
			return field
		}
	}
	return panel.Field{}
}
