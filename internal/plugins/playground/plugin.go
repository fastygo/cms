package playground

import (
	"context"

	"github.com/fastygo/cms/internal/platform/plugins"
	"github.com/fastygo/cms/internal/site/adminfixtures"
	"github.com/fastygo/cms/internal/site/ui/elements"
)

type Plugin struct{}

func New() Plugin {
	return Plugin{}
}

func (Plugin) Manifest() plugins.Manifest {
	return plugins.Manifest{
		ID:          "playground",
		Name:        "Playground",
		Version:     "0.1.0",
		Contract:    "0.1",
		Description: "Browser-local playground import/export and source bootstrap UX.",
		Assets: []plugins.Asset{
			{ID: "playground-admin", Surface: plugins.SurfaceAdmin, Path: "/static/js/playground.js"},
		},
		Hooks: []plugins.HookRegistration{
			{HookID: "plugin.activate.after", HandlerID: "playground.audit", OwnerID: "playground", Category: "action", Priority: 100},
		},
	}
}

func (Plugin) Register(_ context.Context, registry *plugins.Registry) error {
	manifest := Plugin{}.Manifest()
	registry.AddAssets(manifest.Assets...)
	registry.AddHooks(manifest.Hooks...)
	registry.AddScreenActions(
		plugins.ScreenActionRegistration{ScreenID: "settings", Build: playgroundActions},
		plugins.ScreenActionRegistration{ScreenID: "headless", Build: playgroundActions},
	)
	return nil
}

func playgroundActions(fixture adminfixtures.AdminFixture) []elements.Action {
	return []elements.Action{
		{Label: fixture.Label("action_import_source", "Import from compatibility REST source"), Href: "?playground=import-source", Style: "outline", Enabled: true},
		{Label: fixture.Label("action_import_json", "Import JSON from device"), Href: "?playground=import-json", Style: "outline", Enabled: true},
		{Label: fixture.Label("action_export_json", "Export JSON to device"), Href: "?playground=export-json", Style: "outline", Enabled: true},
		{Label: fixture.Label("action_refresh", "Refresh from source"), Href: "?playground=refresh-source", Style: "outline", Enabled: true},
		{Label: fixture.Label("action_reset", "Reset local playground storage"), Href: "?playground=reset-storage", Style: "outline", Enabled: true},
	}
}
