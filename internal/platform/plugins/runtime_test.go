package plugins

import (
	"context"
	"errors"
	"testing"

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
