package config

import (
	"path/filepath"
	"testing"
)

func TestLoadAppliesHealthDefaults(t *testing.T) {
	t.Setenv("HEALTH_LIVE_PATH", "")
	t.Setenv("HEALTH_READY_PATH", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Framework.HealthLivePath != defaultHealthLivePath {
		t.Fatalf("HealthLivePath = %q, want %q", cfg.Framework.HealthLivePath, defaultHealthLivePath)
	}
	if cfg.Framework.HealthReadyPath != defaultHealthReadyPath {
		t.Fatalf("HealthReadyPath = %q, want %q", cfg.Framework.HealthReadyPath, defaultHealthReadyPath)
	}
}

func TestLoadUsesGoCMSStaticDir(t *testing.T) {
	t.Setenv("APP_STATIC_DIR", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if filepath.Base(filepath.Dir(cfg.Framework.StaticDir)) != "web" || filepath.Base(cfg.Framework.StaticDir) != "static" {
		t.Fatalf("StaticDir = %q, want path ending in web/static", cfg.Framework.StaticDir)
	}
}

func TestLoadUsesRuntimeAndStorageProfiles(t *testing.T) {
	t.Setenv("GOCMS_PRESET", "")
	t.Setenv("GOCMS_RUNTIME_PROFILE", "playground")
	t.Setenv("GOCMS_STORAGE_PROFILE", "browser-indexeddb")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.RuntimeProfile != "playground" {
		t.Fatalf("RuntimeProfile = %q, want playground", cfg.RuntimeProfile)
	}
	if cfg.StorageProfile != "browser-indexeddb" {
		t.Fatalf("StorageProfile = %q, want browser-indexeddb", cfg.StorageProfile)
	}

	t.Setenv("GOCMS_RUNTIME_PROFILE", "invalid")
	t.Setenv("GOCMS_STORAGE_PROFILE", "invalid")
	cfg, err = Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.RuntimeProfile == "invalid" || cfg.StorageProfile == "invalid" {
		t.Fatalf("expected unknown profiles to fallback, got runtime=%q storage=%q", cfg.RuntimeProfile, cfg.StorageProfile)
	}
}

func TestLoadUsesPresetPlan(t *testing.T) {
	t.Setenv("GOCMS_PRESET", "ssh-fixtures")
	t.Setenv("GOCMS_RUNTIME_PROFILE", "")
	t.Setenv("GOCMS_STORAGE_PROFILE", "")
	t.Setenv("GOCMS_PLUGIN_SET", "")
	t.Setenv("GOCMS_SITE_PACKAGE_DIR", filepath.FromSlash("/tmp/site-package"))

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Preset != "ssh-fixtures" {
		t.Fatalf("Preset = %q, want ssh-fixtures", cfg.Preset)
	}
	if cfg.Framework.AppBind != "127.0.0.1:8080" {
		t.Fatalf("AppBind = %q, want 127.0.0.1:8080", cfg.Framework.AppBind)
	}
	if cfg.SitePackageDir == "" {
		t.Fatalf("SitePackageDir should be resolved from env")
	}
	if len(cfg.ActivePlugins) == 0 {
		t.Fatalf("expected preset plugins to be resolved")
	}
}
