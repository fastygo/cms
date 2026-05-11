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
	if plan.DeploymentProfile != "browser" {
		t.Fatalf("DeploymentProfile = %q, want browser", plan.DeploymentProfile)
	}
	if !plan.PlaygroundAuth || !plan.BrowserStateless {
		t.Fatalf("playground preset should enable demo auth and browser-stateless mode")
	}
	if len(plan.ActivePlugins) != 1 {
		t.Fatalf("ActivePlugins = %v, want playground defaults", plan.ActivePlugins)
	}
	if plan.EnableDevBearer {
		t.Fatalf("playground preset must keep dev bearer disabled")
	}
	if plan.LoginPolicy != "playground" || plan.AdminPolicy != "enabled" {
		t.Fatalf("unexpected playground policies: login=%q admin=%q", plan.LoginPolicy, plan.AdminPolicy)
	}
}

func TestResolveAllowsOverrides(t *testing.T) {
	plan := Resolve(Options{
		Preset:            string(PresetOfflineJSONSQL),
		RuntimeProfile:    "admin",
		StorageProfile:    "memory",
		DeploymentProfile: "serverless",
		AppBind:           "127.0.0.1:9090",
		DataSource:        "file:custom.db",
		PluginSet:         "playground,json-import-export,playground",
		SitePackageDir:    "/tmp/site-package",
		EnableDevBearer:   "false",
		LoginPolicy:       "disabled",
		AdminPolicy:       "operator",
	})
	if plan.StorageProfile != "memory" {
		t.Fatalf("StorageProfile = %q, want memory", plan.StorageProfile)
	}
	if plan.DeploymentProfile != "serverless" {
		t.Fatalf("DeploymentProfile = %q, want serverless", plan.DeploymentProfile)
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
	if plan.EnableDevBearer {
		t.Fatalf("EnableDevBearer = true, want false")
	}
	if plan.LoginPolicy != "disabled" || plan.AdminPolicy != "operator" {
		t.Fatalf("unexpected policy overrides: login=%q admin=%q", plan.LoginPolicy, plan.AdminPolicy)
	}
}

func TestResolveHardensDemoAuthForProductionDeployments(t *testing.T) {
	plan := Resolve(Options{
		Preset:            string(PresetFull),
		DeploymentProfile: "serverless",
	})
	if plan.EnableDevBearer {
		t.Fatalf("serverless deployment must disable dev bearer auth")
	}
	if plan.LoginPolicy != "local" {
		t.Fatalf("LoginPolicy = %q, want local", plan.LoginPolicy)
	}

	plan = Resolve(Options{
		Preset:            string(PresetFull),
		DeploymentProfile: "container",
		LoginPolicy:       "disabled",
	})
	if plan.LoginPolicy != "disabled" {
		t.Fatalf("explicit non-demo login policy should be preserved, got %q", plan.LoginPolicy)
	}
}
