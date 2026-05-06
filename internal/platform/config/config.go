package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fastygo/framework/pkg/app"
	"github.com/fastygo/cms/internal/platform/runtimeprofile"
)

const (
	defaultHealthLivePath  = "/healthz"
	defaultHealthReadyPath = "/readyz"
	defaultSessionKey      = "gocms-development-session-key-change-me"
	defaultStaticDir       = "web/static"
	frameworkStaticDir     = "internal/site/web/static"
)

// Config contains all runtime configuration resolved by the composition root.
type Config struct {
	Framework    app.Config
	SeedFixtures bool
	RuntimeProfile string
	StorageProfile string
}

// Load reads environment configuration and applies GoCMS pass-0 defaults.
func Load() (Config, error) {
	frameworkConfig, err := app.LoadConfig()
	if err != nil {
		return Config{}, err
	}

	applyDefaults(&frameworkConfig)
	if err := validate(frameworkConfig); err != nil {
		return Config{}, err
	}

	return Config{
		Framework:      frameworkConfig,
		SeedFixtures:   parseBool(os.Getenv("GOCMS_SEED_FIXTURES"), true),
		RuntimeProfile: normalizeRuntimeProfile(os.Getenv("GOCMS_RUNTIME_PROFILE")),
		StorageProfile: normalizeStorageProfile(os.Getenv("GOCMS_STORAGE_PROFILE")),
	}, nil
}

func applyDefaults(cfg *app.Config) {
	if cfg.HealthLivePath == "" {
		cfg.HealthLivePath = defaultHealthLivePath
	}
	if cfg.HealthReadyPath == "" {
		cfg.HealthReadyPath = defaultHealthReadyPath
	}
	if cfg.SessionKey == "" {
		cfg.SessionKey = defaultSessionKey
	}
	if cfg.StaticDir == "" || cfg.StaticDir == frameworkStaticDir {
		cfg.StaticDir = resolveDefaultStaticDir()
	}
}

func resolveDefaultStaticDir() string {
	workingDir, err := os.Getwd()
	if err != nil {
		return defaultStaticDir
	}
	for dir := workingDir; ; dir = filepath.Dir(dir) {
		candidate := filepath.Join(dir, defaultStaticDir)
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return defaultStaticDir
		}
	}
}

func validate(cfg app.Config) error {
	if cfg.DefaultLocale == "" {
		return fmt.Errorf("default locale is required")
	}
	if len(cfg.AvailableLocales) == 0 {
		return fmt.Errorf("at least one available locale is required")
	}
	return nil
}

func normalizeRuntimeProfile(raw string) string {
	if raw == "" {
		return string(runtimeprofile.DefaultRuntimeProfile)
	}
	if err := runtimeprofile.ValidateRuntimeProfile(raw); err == nil {
		return raw
	}
	return string(runtimeprofile.DefaultRuntimeProfile)
}

func normalizeStorageProfile(raw string) string {
	if raw == "" {
		return string(runtimeprofile.DefaultStorageProfile)
	}
	if err := runtimeprofile.ValidateStorageProfile(raw); err == nil {
		return raw
	}
	return string(runtimeprofile.DefaultStorageProfile)
}

func parseBool(value string, fallback bool) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return fallback
	}
}
