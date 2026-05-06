package plugins

import (
	"context"
	"errors"
	"testing"

	"github.com/fastygo/cms/internal/domain/authz"
	"github.com/fastygo/cms/internal/site/adminfixtures"
	"github.com/fastygo/cms/internal/site/ui/elements"
)

type testDescriptor struct {
	manifest Manifest
	err      error
}

func (d testDescriptor) Manifest() Manifest { return d.manifest }

func (d testDescriptor) Register(_ context.Context, registry *Registry) error {
	if d.err != nil {
		return d.err
	}
	registry.AddEditorProviders(EditorProviderRegistration{
		ID:          "plugin-editor",
		Label:       "Plugin Editor",
		Description: "Plugin-provided editor",
		Priority:    50,
	})
	registry.AddScreenActions(ScreenActionRegistration{
		ScreenID: "settings",
		Build: func(adminfixtures.AdminFixture) []elements.Action {
			return []elements.Action{{Label: "Plugin action", Href: "/go-admin/plugin", Enabled: true}}
		},
	})
	registry.AddAssets(Asset{ID: "asset", Surface: SurfaceAdmin, Path: "/static/js/plugin.js"})
	return nil
}

func TestRuntimeActivateRegistersCompiledPlugins(t *testing.T) {
	runtime, err := NewRuntime(NewInMemoryStateRepository(), testDescriptor{
		manifest: Manifest{ID: "example-plugin", Name: "Example", Version: "1.0.0", Contract: "0.1"},
	})
	if err != nil {
		t.Fatalf("NewRuntime() error = %v", err)
	}
	registry, err := runtime.Activate(context.Background(), []string{"example-plugin"})
	if err != nil {
		t.Fatalf("Activate() error = %v", err)
	}
	if len(registry.ScreenActions("settings", adminfixtures.MustLoad("en"))) != 1 {
		t.Fatalf("expected plugin screen action to be registered")
	}
	if len(registry.AssetsForSurface(SurfaceAdmin)) != 1 {
		t.Fatalf("expected plugin asset to be registered")
	}
	if provider, ok := registry.ResolveEditorProvider("plugin-editor"); !ok || provider.ID != "plugin-editor" {
		t.Fatalf("expected plugin editor provider to be registered, got %v %v", provider, ok)
	}
}

func TestRuntimeActivateFailsSafely(t *testing.T) {
	repo := NewInMemoryStateRepository()
	runtime, err := NewRuntime(repo, testDescriptor{
		manifest: Manifest{ID: "broken-plugin", Name: "Broken", Version: "1.0.0", Contract: "0.1"},
		err:      errors.New("boom"),
	})
	if err != nil {
		t.Fatalf("NewRuntime() error = %v", err)
	}
	if _, err := runtime.Activate(context.Background(), []string{"broken-plugin"}); err == nil {
		t.Fatalf("expected activation failure")
	}
	states, err := repo.Load(context.Background())
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if states["broken-plugin"].State != StateFailed {
		t.Fatalf("plugin state = %q, want failed", states["broken-plugin"].State)
	}
}

func TestRuntimeDeactivateMarksPluginInactive(t *testing.T) {
	repo := NewInMemoryStateRepository()
	runtime, err := NewRuntime(repo, testDescriptor{
		manifest: Manifest{ID: "example-plugin", Name: "Example", Version: "1.0.0", Contract: "0.1"},
	})
	if err != nil {
		t.Fatalf("NewRuntime() error = %v", err)
	}
	if err := runtime.Deactivate(context.Background(), []string{"example-plugin"}); err != nil {
		t.Fatalf("Deactivate() error = %v", err)
	}
	states, err := repo.Load(context.Background())
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if states["example-plugin"].State != StateInactive {
		t.Fatalf("plugin state = %q, want inactive", states["example-plugin"].State)
	}
}

func TestRegistryKeepsRoutesAndCapabilityFilteredMenuTogether(t *testing.T) {
	registry := NewRegistry()
	registry.AddRoutes(Route{Pattern: "GET /go-admin/example", Surface: SurfaceAdmin, Protected: true})
	registry.AddAdminMenu(AdminMenuItem{
		ID:         "example",
		Label:      "Example",
		Path:       "/go-admin/example",
		Order:      10,
		Capability: authz.CapabilitySettingsManage,
	})

	if len(registry.RoutesForSurface(SurfaceAdmin)) != 1 {
		t.Fatalf("expected admin route to be registered")
	}
	if got := registry.AdminMenuItems(authz.NewPrincipal("viewer", authz.CapabilityControlPanelAccess)); len(got) != 0 {
		t.Fatalf("viewer menu = %v, want filtered out item", got)
	}
	rootItems := registry.AdminMenuItems(authz.Root())
	if len(rootItems) != 1 || rootItems[0].ID != "example" {
		t.Fatalf("root menu = %v, want example item", rootItems)
	}
}

func TestRegistryResolveEditorProviderFallsBackToPriorityOrder(t *testing.T) {
	registry := NewRegistry()
	registry.AddEditorProviders(
		EditorProviderRegistration{ID: "fallback", Label: "Fallback", Priority: 20},
		EditorProviderRegistration{ID: "primary", Label: "Primary", Priority: 10},
	)

	resolved, ok := registry.ResolveEditorProvider("missing")
	if !ok {
		t.Fatalf("expected fallback editor provider")
	}
	if resolved.ID != "primary" {
		t.Fatalf("resolved editor provider = %q, want primary", resolved.ID)
	}
}
