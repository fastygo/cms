package themes

import (
	"strings"
	"testing"

	domainthemes "github.com/fastygo/cms/internal/domain/themes"
	gocmsdefault "github.com/fastygo/cms/internal/themes/gocmsdefault"
)

func TestValidateManifestRequiresCoreFields(t *testing.T) {
	for name, manifest := range map[string]domainthemes.Manifest{
		"missing id":       {Name: "Theme", Version: "1.0.0", Contract: "0.1"},
		"missing name":     {ID: "demo", Version: "1.0.0", Contract: "0.1"},
		"missing version":  {ID: "demo", Name: "Theme", Contract: "0.1"},
		"missing contract": {ID: "demo", Name: "Theme", Version: "1.0.0"},
	} {
		t.Run(name, func(t *testing.T) {
			if err := domainthemes.ValidateManifest(manifest); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestDefaultRegistryExposesBuiltInTheme(t *testing.T) {
	registry := DefaultRegistry()
	theme, ok := registry.Get(DefaultThemeID)
	if !ok {
		t.Fatalf("expected built-in theme %q", DefaultThemeID)
	}
	manifest := theme.Manifest()
	if got := manifest.Name; got != DefaultThemeLabel {
		t.Fatalf("name = %q", got)
	}
	if _, ok := manifest.Templates["front"]; !ok {
		t.Fatal("expected front template role")
	}
	if _, ok := manifest.Templates["not_found"]; !ok {
		t.Fatal("expected not_found template role")
	}
	if _, ok := manifest.Assets["theme.css"]; !ok {
		t.Fatal("expected theme.css asset")
	}
}

func TestRegistryRejectsDuplicateThemes(t *testing.T) {
	theme := gocmsdefault.New()
	_, err := NewRegistry(theme, theme)
	if err == nil || !strings.Contains(err.Error(), "already registered") {
		t.Fatalf("expected duplicate registration error, got %v", err)
	}
}
