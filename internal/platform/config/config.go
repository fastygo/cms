package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fastygo/cms/internal/platform/locales"
	"github.com/fastygo/cms/internal/platform/preset"
	"github.com/fastygo/cms/internal/platform/runtimeprofile"
	"github.com/fastygo/framework/pkg/app"
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
	Framework         app.Config
	SeedFixtures      bool
	RuntimeProfile    string
	StorageProfile    string
	DeploymentProfile string
	Preset            string
	ActivePlugins     []string
	SitePackageDir    string
	PlaygroundAuth    bool
	BrowserStateless  bool
	EnableDevBearer   bool
	LoginPolicy       string
	AdminPolicy       string
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

	plan := preset.Resolve(preset.Options{
		Preset:            os.Getenv("GOCMS_PRESET"),
		RuntimeProfile:    os.Getenv("GOCMS_RUNTIME_PROFILE"),
		StorageProfile:    os.Getenv("GOCMS_STORAGE_PROFILE"),
		DeploymentProfile: os.Getenv("GOCMS_DEPLOYMENT_PROFILE"),
		AppBind:           os.Getenv("APP_BIND"),
		DataSource:        frameworkConfig.DataSource,
		PluginSet:         os.Getenv("GOCMS_PLUGIN_SET"),
		SitePackageDir:    os.Getenv("GOCMS_SITE_PACKAGE_DIR"),
		EnableDevBearer:   os.Getenv("GOCMS_ENABLE_DEV_BEARER"),
		LoginPolicy:       os.Getenv("GOCMS_LOGIN_POLICY"),
		AdminPolicy:       os.Getenv("GOCMS_ADMIN_POLICY"),
	})
	frameworkConfig.AppBind = plan.AppBind
	frameworkConfig.DataSource = plan.DataSource

	return Config{
		Framework:         frameworkConfig,
		SeedFixtures:      parseBool(os.Getenv("GOCMS_SEED_FIXTURES"), true),
		RuntimeProfile:    plan.RuntimeProfile,
		StorageProfile:    plan.StorageProfile,
		DeploymentProfile: plan.DeploymentProfile,
		Preset:            plan.Name,
		ActivePlugins:     plan.ActivePlugins,
		SitePackageDir:    plan.SitePackageDir,
		PlaygroundAuth:    plan.PlaygroundAuth,
		BrowserStateless:  plan.BrowserStateless,
		EnableDevBearer:   plan.EnableDevBearer,
		LoginPolicy:       plan.LoginPolicy,
		AdminPolicy:       plan.AdminPolicy,
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
	if cfg.DefaultLocale == "" {
		cfg.DefaultLocale = locales.Default
	}
	if len(cfg.AvailableLocales) == 0 {
		cfg.AvailableLocales = locales.Supported()
	} else {
		cfg.AvailableLocales = normalizeFrameworkLocales(cfg.DefaultLocale, cfg.AvailableLocales)
	}
	if !containsLocale(cfg.AvailableLocales, cfg.DefaultLocale) {
		cfg.AvailableLocales = append([]string{cfg.DefaultLocale}, cfg.AvailableLocales...)
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
	if !containsLocale(cfg.AvailableLocales, cfg.DefaultLocale) {
		return fmt.Errorf("default locale %q must be included in available locales", cfg.DefaultLocale)
	}
	return nil
}

func normalizeFrameworkLocales(defaultLocale string, raw []string) []string {
	def := locales.NormalizeOrDefault(defaultLocale)
	out := make([]string, 0, len(raw)+1)
	seen := map[string]struct{}{}
	add := func(code string) {
		code = locales.Normalize(code)
		if !locales.IsSupported(code) {
			return
		}
		if _, ok := seen[code]; ok {
			return
		}
		seen[code] = struct{}{}
		out = append(out, code)
	}
	add(def)
	for _, v := range raw {
		add(v)
	}
	if len(out) == 0 {
		return locales.Supported()
	}
	return out
}

func containsLocale(list []string, target string) bool {
	t := locales.Normalize(target)
	for _, v := range list {
		if locales.Normalize(v) == t {
			return true
		}
	}
	return false
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

func normalizeDeploymentProfile(raw string) string {
	if raw == "" {
		return string(runtimeprofile.DefaultDeploymentProfile)
	}
	if err := runtimeprofile.ValidateDeploymentProfile(raw); err == nil {
		return raw
	}
	return string(runtimeprofile.DefaultDeploymentProfile)
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
