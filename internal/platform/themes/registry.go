package themes

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/a-h/templ"
	"github.com/fastygo/cms/internal/application/publicrender"
	domainthemes "github.com/fastygo/cms/internal/domain/themes"
	blanktheme "github.com/fastygo/cms/internal/themes/blank"
	gocmsdefault "github.com/fastygo/cms/internal/themes/gocmsdefault"
)

const (
	DefaultThemeID    domainthemes.ThemeID = "gocms-default"
	ActiveThemeKey    string               = "theme.active"
	PreviewThemeKey   string               = "theme.preview"
	StylePresetKey    string               = "theme.style_preset"
	DefaultThemeLabel string               = "GoCMS Default"
)

type Theme interface {
	Manifest() domainthemes.Manifest
	Presets() []domainthemes.StylePreset
	Render(context.Context, publicrender.RenderRequest) (templ.Component, error)
}

type Registry struct {
	items map[domainthemes.ThemeID]Theme
}

func NewRegistry(items ...Theme) (*Registry, error) {
	registry := &Registry{items: make(map[domainthemes.ThemeID]Theme, len(items))}
	for _, item := range items {
		manifest := item.Manifest()
		if err := domainthemes.ValidateManifest(manifest); err != nil {
			return nil, err
		}
		if _, exists := registry.items[manifest.ID]; exists {
			return nil, fmt.Errorf("theme %q is already registered", manifest.ID)
		}
		registry.items[manifest.ID] = item
	}
	if _, ok := registry.items[DefaultThemeID]; !ok {
		return nil, fmt.Errorf("default theme %q must be registered", DefaultThemeID)
	}
	return registry, nil
}

func DefaultRegistry() *Registry {
	registry, err := NewRegistry(gocmsdefault.New(), blanktheme.New())
	if err != nil {
		panic(err)
	}
	return registry
}

func (r *Registry) List() []domainthemes.Manifest {
	if r == nil {
		return nil
	}
	items := make([]domainthemes.Manifest, 0, len(r.items))
	for _, item := range r.items {
		items = append(items, item.Manifest())
	}
	slices.SortFunc(items, func(left domainthemes.Manifest, right domainthemes.Manifest) int {
		return strings.Compare(left.Name, right.Name)
	})
	return items
}

func (r *Registry) Get(id domainthemes.ThemeID) (Theme, bool) {
	if r == nil {
		return nil, false
	}
	item, ok := r.items[id]
	return item, ok
}

func (r *Registry) ResolveActive(id string) Theme {
	if item, ok := r.Get(domainthemes.ThemeID(id)); ok {
		return item
	}
	item, _ := r.Get(DefaultThemeID)
	return item
}

func (r *Registry) Manifest(id string) domainthemes.Manifest {
	return r.ResolveActive(id).Manifest()
}

func (r *Registry) ListPresets(themeID string) []domainthemes.StylePreset {
	theme := r.ResolveActive(themeID)
	presets := append([]domainthemes.StylePreset(nil), theme.Presets()...)
	slices.SortFunc(presets, func(left domainthemes.StylePreset, right domainthemes.StylePreset) int {
		return strings.Compare(left.Name, right.Name)
	})
	return presets
}

func (r *Registry) ResolvePreset(themeID string, presetID string) domainthemes.StylePreset {
	presets := r.ListPresets(themeID)
	for _, preset := range presets {
		if preset.ID == presetID {
			return preset
		}
	}
	if len(presets) > 0 {
		return presets[0]
	}
	return domainthemes.StylePreset{}
}

func (r *Registry) ThemeAssets(themeID string) (stylesheets []string, scripts []string) {
	manifest := r.Manifest(themeID)
	for _, asset := range manifest.Assets {
		switch asset.Type {
		case domainthemes.AssetTypeCSS:
			stylesheets = append(stylesheets, asset.Path)
		case domainthemes.AssetTypeJS:
			scripts = append(scripts, asset.Path)
		}
	}
	slices.Sort(stylesheets)
	slices.Sort(scripts)
	return stylesheets, scripts
}
