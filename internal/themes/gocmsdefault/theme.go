package gocmsdefault

import (
	"context"

	"github.com/a-h/templ"
	"github.com/fastygo/cms/internal/application/publicrender"
	domainthemes "github.com/fastygo/cms/internal/domain/themes"
)

type Theme struct{}

func New() Theme {
	return Theme{}
}

func (Theme) Manifest() domainthemes.Manifest {
	return domainthemes.Manifest{
		ID:          "gocms-default",
		Name:        "GoCMS Default",
		Version:     "0.2.0",
		Contract:    "0.2",
		Description: "Starter native GoCMS theme with blog, archives, author pages, breadcrumbs, SEO, and footer support.",
		Author:      "GoCMS",
		Templates: map[domainthemes.TemplateRole]string{
			"front":     "front",
			"index":     "index",
			"page":      "page",
			"post":      "post",
			"archive":   "archive",
			"taxonomy":  "taxonomy",
			"author":    "author",
			"search":    "search",
			"not_found": "404",
			"error":     "404",
		},
		Assets: map[string]domainthemes.Asset{
			"theme.css": {ID: "theme.css", Type: domainthemes.AssetTypeCSS, Path: "/static/themes/gocms-default/theme.css", Load: domainthemes.AssetLocationHead},
		},
		Slots: []domainthemes.Slot{"header", "footer", "before_content", "after_content"},
		Settings: []domainthemes.SettingDefinition{
			{Key: "theme.brand_name", Label: "Brand name", Type: "string", Default: "GoCMS", Public: true, Validation: "required"},
			// Empty default: public site resolves copy from publicfixtures by locale until an explicit value is saved.
			{Key: "theme.home_intro", Label: "Home intro", Type: "string", Default: "", Public: true},
		},
	}
}

func (Theme) Presets() []domainthemes.StylePreset {
	return []domainthemes.StylePreset{
		{
			ID:           "default",
			Name:         "Default",
			Description:  "Neutral GoCMS preset.",
			Stylesheets:  []string{"/static/presets/default/preset.css"},
			TokenJSON:    "/static/presets/default/tokens.json",
			PreviewClass: "preset-default",
		},
		{
			ID:           "bold-tech",
			Name:         "Bold Tech",
			Description:  "BrandOSS-style bold technology preset.",
			Stylesheets:  []string{"/static/presets/bold-tech/preset.css"},
			TokenJSON:    "/static/presets/bold-tech/tokens.json",
			PreviewClass: "preset-bold-tech",
		},
	}
}

func (Theme) Render(_ context.Context, request publicrender.RenderRequest) (templ.Component, error) {
	return ThemePage(request.Page), nil
}
