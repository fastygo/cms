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
	register func(context.Context, *Registry) error
}

func (d testDescriptor) Manifest() Manifest { return d.manifest }

func (d testDescriptor) Register(_ context.Context, registry *Registry) error {
	if d.register != nil {
		return d.register(context.Background(), registry)
	}
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

func TestRegistryDispatchActionUsesPriorityAndStableSamePriorityOrder(t *testing.T) {
	registry := NewRegistry()
	order := make([]string, 0)
	registry.AddActionHandlers(
		ActionHandlerRegistration{
			Hook: HookRegistration{HookID: "content.save", HandlerID: "late", OwnerID: "a", Priority: 20},
			Handle: func(context.Context, HookContext, any) error {
				order = append(order, "late")
				return nil
			},
		},
		ActionHandlerRegistration{
			Hook: HookRegistration{HookID: "content.save", HandlerID: "first", OwnerID: "a", Priority: 10},
			Handle: func(context.Context, HookContext, any) error {
				order = append(order, "first")
				return nil
			},
		},
		ActionHandlerRegistration{
			Hook: HookRegistration{HookID: "content.save", HandlerID: "second", OwnerID: "a", Priority: 10},
			Handle: func(context.Context, HookContext, any) error {
				order = append(order, "second")
				return nil
			},
		},
	)

	if err := registry.DispatchAction(context.Background(), "content.save", HookContext{}, nil); err != nil {
		t.Fatalf("DispatchAction() error = %v", err)
	}
	if got, want := len(order), 3; got != want {
		t.Fatalf("handler count = %d, want %d", got, want)
	}
	if order[0] != "first" || order[1] != "second" || order[2] != "late" {
		t.Fatalf("handler order = %v", order)
	}
}

func TestRegistryApplyFilterUsesPriorityOrder(t *testing.T) {
	registry := NewRegistry()
	registry.AddFilterHandlers(
		FilterHandlerRegistration{
			Hook: HookRegistration{HookID: "render.content.filter", HandlerID: "prefix", OwnerID: "a", Priority: 10},
			Handle: func(_ context.Context, _ HookContext, value any) (any, error) {
				return "prefix:" + value.(string), nil
			},
		},
		FilterHandlerRegistration{
			Hook: HookRegistration{HookID: "render.content.filter", HandlerID: "suffix", OwnerID: "a", Priority: 20},
			Handle: func(_ context.Context, _ HookContext, value any) (any, error) {
				return value.(string) + ":suffix", nil
			},
		},
	)

	filtered, err := FilterValue(context.Background(), registry, "render.content.filter", HookContext{}, "body")
	if err != nil {
		t.Fatalf("FilterValue() error = %v", err)
	}
	if filtered != "prefix:body:suffix" {
		t.Fatalf("filtered value = %q", filtered)
	}
}

func TestRegistryActionErrorPoliciesFailAndCollect(t *testing.T) {
	t.Run("fail", func(t *testing.T) {
		registry := NewRegistry()
		calls := 0
		registry.AddActionHandlers(
			ActionHandlerRegistration{
				Hook: HookRegistration{HookID: "plugin.activate.after", HandlerID: "broken", OwnerID: "a", Priority: 10, ErrorPolicy: HookErrorPolicyFail},
				Handle: func(context.Context, HookContext, any) error {
					calls++
					return errors.New("boom")
				},
			},
			ActionHandlerRegistration{
				Hook: HookRegistration{HookID: "plugin.activate.after", HandlerID: "next", OwnerID: "a", Priority: 20},
				Handle: func(context.Context, HookContext, any) error {
					calls++
					return nil
				},
			},
		)
		if err := registry.DispatchAction(context.Background(), "plugin.activate.after", HookContext{}, nil); err == nil {
			t.Fatalf("expected failure")
		}
		if calls != 1 {
			t.Fatalf("calls = %d, want 1", calls)
		}
	})

	t.Run("collect", func(t *testing.T) {
		registry := NewRegistry()
		calls := 0
		registry.AddActionHandlers(
			ActionHandlerRegistration{
				Hook: HookRegistration{HookID: "plugin.activate.after", HandlerID: "broken", OwnerID: "a", ErrorPolicy: HookErrorPolicyCollect},
				Handle: func(context.Context, HookContext, any) error {
					calls++
					return errors.New("boom")
				},
			},
			ActionHandlerRegistration{
				Hook: HookRegistration{HookID: "plugin.activate.after", HandlerID: "next", OwnerID: "a", Priority: 20},
				Handle: func(context.Context, HookContext, any) error {
					calls++
					return nil
				},
			},
		)
		if err := registry.DispatchAction(context.Background(), "plugin.activate.after", HookContext{}, nil); err == nil {
			t.Fatalf("expected aggregated failure")
		}
		if calls != 2 {
			t.Fatalf("calls = %d, want 2", calls)
		}
	})
}

func TestRuntimeLifecycleHooksAndDeactivateBehavior(t *testing.T) {
	repo := NewInMemoryStateRepository()
	events := make([]string, 0)
	runtime, err := NewRuntime(repo,
		testDescriptor{
			manifest: Manifest{ID: "observer", Name: "Observer", Version: "1.0.0", Contract: "0.1"},
			register: func(_ context.Context, registry *Registry) error {
				registry.AddActionHandlers(
					ActionHandlerRegistration{
						Hook: HookRegistration{HookID: "plugin.activate.after", HandlerID: "observer.activate", OwnerID: "observer"},
						Handle: func(_ context.Context, hookContext HookContext, _ any) error {
							events = append(events, "activate:"+hookContext.Metadata["plugin_id"].(string))
							return nil
						},
					},
					ActionHandlerRegistration{
						Hook: HookRegistration{HookID: "plugin.deactivate.after", HandlerID: "observer.deactivate", OwnerID: "observer"},
						Handle: func(_ context.Context, hookContext HookContext, _ any) error {
							events = append(events, "deactivate:"+hookContext.Metadata["plugin_id"].(string))
							return nil
						},
					},
				)
				return nil
			},
		},
		testDescriptor{
			manifest: Manifest{ID: "target", Name: "Target", Version: "1.0.0", Contract: "0.1"},
			register: func(_ context.Context, registry *Registry) error {
				registry.AddFilterHandlers(FilterHandlerRegistration{
					Hook: HookRegistration{HookID: "render.content.filter", HandlerID: "target.filter", OwnerID: "target"},
					Handle: func(_ context.Context, _ HookContext, value any) (any, error) {
						return value.(string) + ":filtered", nil
					},
				})
				return nil
			},
		},
	)
	if err != nil {
		t.Fatalf("NewRuntime() error = %v", err)
	}

	registry, err := runtime.Activate(context.Background(), []string{"observer", "target"})
	if err != nil {
		t.Fatalf("Activate() error = %v", err)
	}
	filtered, err := FilterValue(context.Background(), registry, "render.content.filter", HookContext{}, "body")
	if err != nil || filtered != "body:filtered" {
		t.Fatalf("active filter = %q err=%v", filtered, err)
	}

	if err := runtime.Deactivate(context.Background(), []string{"target"}); err != nil {
		t.Fatalf("Deactivate() error = %v", err)
	}
	registry, err = runtime.Activate(context.Background(), []string{"observer"})
	if err != nil {
		t.Fatalf("Activate(observer) error = %v", err)
	}
	filtered, err = FilterValue(context.Background(), registry, "render.content.filter", HookContext{}, "body")
	if err != nil {
		t.Fatalf("FilterValue() error = %v", err)
	}
	if filtered != "body" {
		t.Fatalf("deactivated filter should be removed, got %q", filtered)
	}
	if len(events) < 2 || events[0] != "activate:observer" || events[1] != "activate:target" {
		t.Fatalf("unexpected activation events = %v", events)
	}
	if events[len(events)-2] != "deactivate:target" {
		t.Fatalf("unexpected final event sequence = %v", events)
	}
}

func TestRegistryTracksDescriptorSurfaces(t *testing.T) {
	runtime, err := NewRuntime(NewInMemoryStateRepository(), testDescriptor{
		manifest: Manifest{
			ID:           "descriptor-surfaces",
			Name:         "Descriptor Surfaces",
			Version:      "1.0.0",
			Contract:     "0.1",
			Capabilities: []CapabilityDefinition{{ID: "example.manage", Description: "Manage example plugin"}},
			Settings:     []SettingDefinition{{Key: "example.enabled", Type: "boolean", Default: "true", Public: false}},
			Hooks:        []HookRegistration{{HookID: "plugin.activate.after", HandlerID: "descriptor.audit", OwnerID: "descriptor-surfaces", Category: HookCategoryAction}},
			Assets:       []Asset{{ID: "example-admin", Surface: SurfaceAdmin, Path: "/static/js/example.js"}},
		},
		register: func(_ context.Context, registry *Registry) error {
			manifest := testDescriptor{
				manifest: Manifest{
					ID:           "descriptor-surfaces",
					Name:         "Descriptor Surfaces",
					Version:      "1.0.0",
					Contract:     "0.1",
					Capabilities: []CapabilityDefinition{{ID: "example.manage", Description: "Manage example plugin"}},
					Settings:     []SettingDefinition{{Key: "example.enabled", Type: "boolean", Default: "true", Public: false}},
					Hooks:        []HookRegistration{{HookID: "plugin.activate.after", HandlerID: "descriptor.audit", OwnerID: "descriptor-surfaces", Category: HookCategoryAction}},
					Assets:       []Asset{{ID: "example-admin", Surface: SurfaceAdmin, Path: "/static/js/example.js"}},
				},
			}.manifest
			registry.AddCapabilities(manifest.Capabilities...)
			registry.AddSettings(manifest.Settings...)
			registry.AddHooks(manifest.Hooks...)
			registry.AddActionHandlers(ActionHandlerRegistration{
				Hook:   manifest.Hooks[0],
				Handle: func(context.Context, HookContext, any) error { return nil },
			})
			registry.AddAssets(manifest.Assets...)
			registry.AddRoutes(Route{Pattern: "GET /go-admin/example", Surface: SurfaceAdmin, Protected: true})
			registry.AddEditorProviders(EditorProviderRegistration{ID: "descriptor-editor", Label: "Descriptor Editor", Priority: 10})
			registry.AddScreenActions(ScreenActionRegistration{
				ScreenID: "settings",
				Build: func(adminfixtures.AdminFixture) []elements.Action {
					return []elements.Action{{Label: "Descriptor action", Href: "/go-admin/example", Enabled: true}}
				},
			})
			return nil
		},
	})
	if err != nil {
		t.Fatalf("NewRuntime() error = %v", err)
	}

	registry, err := runtime.Activate(context.Background(), []string{"descriptor-surfaces"})
	if err != nil {
		t.Fatalf("Activate() error = %v", err)
	}
	if len(registry.Capabilities()) != 1 || registry.Capabilities()[0].ID != "example.manage" {
		t.Fatalf("capabilities = %+v", registry.Capabilities())
	}
	if len(registry.Settings()) != 1 || registry.Settings()[0].Key != "example.enabled" {
		t.Fatalf("settings = %+v", registry.Settings())
	}
	if len(registry.Hooks()) != 1 || registry.Hooks()[0].HookID != "plugin.activate.after" {
		t.Fatalf("hooks = %+v", registry.Hooks())
	}
	if len(registry.AssetsForSurface(SurfaceAdmin)) != 1 {
		t.Fatalf("assets = %+v", registry.AssetsForSurface(SurfaceAdmin))
	}
	if len(registry.RoutesForSurface(SurfaceAdmin)) != 1 {
		t.Fatalf("routes = %+v", registry.RoutesForSurface(SurfaceAdmin))
	}
	if len(registry.ScreenActions("settings", adminfixtures.MustLoad("en"))) != 1 {
		t.Fatalf("screen actions = %+v", registry.ScreenActions("settings", adminfixtures.MustLoad("en")))
	}
	if provider, ok := registry.ResolveEditorProvider("descriptor-editor"); !ok || provider.ID != "descriptor-editor" {
		t.Fatalf("editor provider = %+v ok=%v", provider, ok)
	}
}
