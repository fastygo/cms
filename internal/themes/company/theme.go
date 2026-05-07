package company

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
		ID:          "company",
		Name:        "Company",
		Version:     "0.1.0",
		Contract:    "0.2",
		Description: "Project/company native theme proving compiled public frontend packages can stay independent from CMS admin internals.",
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
			"theme.css": {ID: "theme.css", Type: domainthemes.AssetTypeCSS, Path: "/static/themes/company/theme.css", Load: domainthemes.AssetLocationHead},
		},
		Slots: []domainthemes.Slot{"header", "footer", "hero", "before_content", "after_content"},
		Settings: []domainthemes.SettingDefinition{
			{Key: "theme.company.cta_label", Label: "CTA label", Type: "string", Default: "Explore insights", Public: true},
			{Key: "theme.company.industry", Label: "Industry label", Type: "string", Default: "Digital operations", Public: true},
		},
	}
}

func (Theme) Presets() []domainthemes.StylePreset {
	return []domainthemes.StylePreset{
		{
			ID:           "company-default",
			Name:         "Company Default",
			Description:  "Clean project/company preset for public frontend validation.",
			Stylesheets:  []string{"/static/presets/default/preset.css"},
			TokenJSON:    "/static/presets/default/tokens.json",
			PreviewClass: "preset-company-default",
		},
		{
			ID:           "company-bold-tech",
			Name:         "Company Bold Tech",
			Description:  "Company layout using the BrandOSS-style bold technology token preset.",
			Stylesheets:  []string{"/static/presets/bold-tech/preset.css"},
			TokenJSON:    "/static/presets/bold-tech/tokens.json",
			PreviewClass: "preset-company-bold-tech",
		},
	}
}

func (Theme) Render(_ context.Context, request publicrender.RenderRequest) (templ.Component, error) {
	return ThemePage(request.Page), nil
}
