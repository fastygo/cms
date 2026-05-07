package cmspanel

import (
	"testing"

	"github.com/fastygo/cms/internal/domain/authz"
	domaincontent "github.com/fastygo/cms/internal/domain/content"
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
	if len(resource.Form.Fields) != 10 {
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

func fieldByID(fields []panel.Field, id string) panel.Field {
	for _, field := range fields {
		if field.ID == id {
			return field
		}
	}
	return panel.Field{}
}
