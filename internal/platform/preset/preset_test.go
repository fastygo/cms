package preset

import "testing"

func TestResolveUsesPresetDefaults(t *testing.T) {
	plan := Resolve(Options{Preset: string(PresetPlayground)})
	if plan.RuntimeProfile != "playground" {
		t.Fatalf("RuntimeProfile = %q, want playground", plan.RuntimeProfile)
	}
	if plan.StorageProfile != "browser-indexeddb" {
		t.Fatalf("StorageProfile = %q, want browser-indexeddb", plan.StorageProfile)
	}
	if !plan.PlaygroundAuth || !plan.BrowserStateless {
		t.Fatalf("playground preset should enable demo auth and browser-stateless mode")
	}
	if len(plan.ActivePlugins) != 1 {
		t.Fatalf("ActivePlugins = %v, want playground defaults", plan.ActivePlugins)
	}
}

func TestResolveAllowsOverrides(t *testing.T) {
	plan := Resolve(Options{
		Preset:         string(PresetOfflineJSONSQL),
		RuntimeProfile: "admin",
		StorageProfile: "memory",
		AppBind:        "127.0.0.1:9090",
		DataSource:     "file:custom.db",
		PluginSet:      "playground,json-import-export,playground",
		SitePackageDir: "/tmp/site-package",
	})
	if plan.StorageProfile != "memory" {
		t.Fatalf("StorageProfile = %q, want memory", plan.StorageProfile)
	}
	if plan.AppBind != "127.0.0.1:9090" {
		t.Fatalf("AppBind = %q, want 127.0.0.1:9090", plan.AppBind)
	}
	if plan.DataSource != "file:custom.db" {
		t.Fatalf("DataSource = %q, want file:custom.db", plan.DataSource)
	}
	if len(plan.ActivePlugins) != 2 {
		t.Fatalf("ActivePlugins = %v, want unique override plugins", plan.ActivePlugins)
	}
	if plan.SitePackageDir != "/tmp/site-package" {
		t.Fatalf("SitePackageDir = %q, want /tmp/site-package", plan.SitePackageDir)
	}
}
