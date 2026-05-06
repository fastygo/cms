package blank

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
		ID:          "blank",
		Name:        "Blank",
		Version:     "0.2.0",
		Contract:    "0.2",
		Description: "Minimal fallback native theme with a single index-style renderer.",
		Author:      "GoCMS",
		Templates: map[domainthemes.TemplateRole]string{
			"front":     "index",
			"index":     "index",
			"page":      "index",
			"post":      "index",
			"archive":   "index",
			"taxonomy":  "index",
			"author":    "index",
			"search":    "index",
			"not_found": "index",
			"error":     "index",
		},
		Assets: map[string]domainthemes.Asset{
			"theme.css": {ID: "theme.css", Type: domainthemes.AssetTypeCSS, Path: "/static/themes/blank/theme.css", Load: domainthemes.AssetLocationHead},
		},
		Slots: []domainthemes.Slot{"header", "footer"},
	}
}

func (Theme) Presets() []domainthemes.StylePreset {
	return []domainthemes.StylePreset{
		{
			ID:          "minimal",
			Name:        "Minimal",
			Description: "Minimal preset for the blank starter theme.",
			Stylesheets: []string{"/static/presets/minimal/preset.css"},
			TokenJSON:   "/static/presets/minimal/tokens.json",
		},
	}
}

func (Theme) Render(_ context.Context, request publicrender.RenderRequest) (templ.Component, error) {
	return ThemePage(request.Page), nil
}
