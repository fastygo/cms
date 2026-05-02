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
