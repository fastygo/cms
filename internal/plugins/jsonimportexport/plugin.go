package jsonimportexport

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/fastygo/cms/internal/application/snapshot"
	"github.com/fastygo/cms/internal/domain/authz"
	"github.com/fastygo/cms/internal/platform/plugins"
	"github.com/fastygo/cms/internal/site/adminfixtures"
	"github.com/fastygo/cms/internal/site/ui/elements"
	"github.com/fastygo/cms/internal/sitepackage/jsondir"
)

type Plugin struct {
	service     snapshot.Service
	sitePackage jsondir.Provider
}

func New(service snapshot.Service, sitePackage jsondir.Provider) Plugin {
	return Plugin{service: service, sitePackage: sitePackage}
}

func (p Plugin) Manifest() plugins.Manifest {
	return plugins.Manifest{
		ID:       "json-import-export",
		Name:     "JSON Import Export",
		Version:  "0.1.0",
		Contract: "0.1",
		Description: "Exports and imports GoCMS content snapshots as JSON or site package files.",
		Capabilities: []plugins.CapabilityDefinition{
			{ID: "json-import-export.manage", Description: "Manage JSON snapshot import and export workflows."},
		},
		Settings: []plugins.SettingDefinition{
			{Key: "json-import-export.site_package_dir", Type: "string", Default: "", Public: false, Capability: authz.CapabilitySettingsManage},
		},
		Hooks: []plugins.HookRegistration{
			{HookID: "settings.update.after", HandlerID: "json-import-export.audit", OwnerID: "json-import-export", Category: "action", Priority: 100},
		},
		Assets: []plugins.Asset{
			{ID: "json-import-export-admin", Surface: plugins.SurfaceAdmin, Path: "/static/js/snapshots.js"},
		},
	}
}

func (p Plugin) Register(_ context.Context, registry *plugins.Registry) error {
	manifest := p.Manifest()
	registry.AddCapabilities(manifest.Capabilities...)
	registry.AddSettings(manifest.Settings...)
	registry.AddHooks(manifest.Hooks...)
	registry.AddAssets(manifest.Assets...)
	registry.AddScreenActions(
		plugins.ScreenActionRegistration{ScreenID: "settings", Build: p.actions},
		plugins.ScreenActionRegistration{ScreenID: "headless", Build: p.actions},
	)
	registry.AddRoutes(
		plugins.Route{
			Pattern:          "GET /go-admin/plugins/json-import-export/export",
			Surface:          plugins.SurfaceAdmin,
			Capability:       authz.CapabilitySettingsManage,
			Protected:        true,
			ProtectedHandler: p.exportJSON,
		},
		plugins.Route{
			Pattern:          "POST /go-admin/plugins/json-import-export/import",
			Surface:          plugins.SurfaceAdmin,
			Capability:       authz.CapabilitySettingsManage,
			Protected:        true,
			ProtectedHandler: p.importJSON,
		},
		plugins.Route{
			Pattern:          "POST /go-admin/plugins/json-import-export/export-site-package",
			Surface:          plugins.SurfaceAdmin,
			Capability:       authz.CapabilitySettingsManage,
			Protected:        true,
			ProtectedHandler: p.exportSitePackage,
		},
		plugins.Route{
			Pattern:          "POST /go-admin/plugins/json-import-export/import-site-package",
			Surface:          plugins.SurfaceAdmin,
			Capability:       authz.CapabilitySettingsManage,
			Protected:        true,
			ProtectedHandler: p.importSitePackage,
		},
	)
	return nil
}

func (p Plugin) actions(fixture adminfixtures.AdminFixture) []elements.Action {
	actions := []elements.Action{
		{Label: fixture.Label("action_export_json", "Export JSON to device"), Href: "/go-admin/plugins/json-import-export/export", Style: "outline", Enabled: true},
		{Label: fixture.Label("action_import_json", "Import JSON from device"), Href: "?plugin-action=json-import-export.import", Style: "outline", Enabled: true},
	}
	if p.sitePackage.Enabled() {
		actions = append(actions,
			elements.Action{Label: fixture.Label("action_export_site_package", "Export site package"), Href: "?plugin-action=json-import-export.export-site-package", Style: "outline", Enabled: true},
			elements.Action{Label: fixture.Label("action_import_site_package", "Import site package"), Href: "?plugin-action=json-import-export.import-site-package", Style: "outline", Enabled: true},
		)
	}
	return actions
}

func (p Plugin) exportJSON(w http.ResponseWriter, r *http.Request, _ authz.Principal) {
	bundle, err := p.service.Export(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	payload, err := json.MarshalIndent(bundle, "", "  ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", `attachment; filename="gocms-content-snapshot.json"`)
	_, _ = w.Write(append(payload, '\n'))
}

func (p Plugin) importJSON(w http.ResponseWriter, r *http.Request, _ authz.Principal) {
	if err := r.ParseMultipartForm(8 << 20); err != nil {
		http.Error(w, "Invalid import request.", http.StatusBadRequest)
		return
	}
	file, _, err := r.FormFile("snapshot")
	if err != nil {
		http.Error(w, "Snapshot file is required.", http.StatusBadRequest)
		return
	}
	defer file.Close()
	var bundle snapshot.Bundle
	if err := json.NewDecoder(file).Decode(&bundle); err != nil {
		http.Error(w, "Invalid snapshot payload.", http.StatusBadRequest)
		return
	}
	if err := p.service.Import(r.Context(), bundle); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (p Plugin) exportSitePackage(w http.ResponseWriter, r *http.Request, _ authz.Principal) {
	if !p.sitePackage.Enabled() {
		http.Error(w, "Site package provider is not configured.", http.StatusBadRequest)
		return
	}
	bundle, err := p.service.Export(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := p.sitePackage.Save(bundle); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (p Plugin) importSitePackage(w http.ResponseWriter, r *http.Request, _ authz.Principal) {
	if !p.sitePackage.Enabled() {
		http.Error(w, "Site package provider is not configured.", http.StatusBadRequest)
		return
	}
	bundle, err := p.sitePackage.Load()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := p.service.Import(r.Context(), bundle); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
