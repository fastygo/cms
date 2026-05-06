package preset

import (
	"strings"

	"github.com/fastygo/cms/internal/platform/runtimeprofile"
)

type Name string

const (
	PresetOfflineJSONSQL Name = "offline-json-sql"
	PresetSSHFixtures    Name = "ssh-fixtures"
	PresetFull           Name = "full"
	PresetHeadless       Name = "headless"
	PresetPlayground     Name = "playground"
)

// DefaultPreset may be overridden at build time with:
// -ldflags "-X github.com/fastygo/cms/internal/platform/preset.DefaultPreset=playground"
var DefaultPreset = string(PresetFull)

type Plan struct {
	Name             string
	RuntimeProfile   string
	StorageProfile   string
	AppBind          string
	DataSource       string
	ActivePlugins    []string
	SitePackageDir   string
	PlaygroundAuth   bool
	BrowserStateless bool
	EnableDevBearer  bool
	LoginPolicy      string
	AdminPolicy      string
}

type Options struct {
	Preset          string
	RuntimeProfile  string
	StorageProfile  string
	AppBind         string
	DataSource      string
	PluginSet       string
	SitePackageDir  string
	EnableDevBearer string
	LoginPolicy     string
	AdminPolicy     string
}

func Resolve(options Options) Plan {
	plan := defaults(normalizePreset(options.Preset))
	if runtimeprofile.IsRuntimeProfile(options.RuntimeProfile) {
		plan.RuntimeProfile = options.RuntimeProfile
	}
	if runtimeprofile.IsStorageProfile(options.StorageProfile) {
		plan.StorageProfile = options.StorageProfile
	}
	if strings.TrimSpace(options.AppBind) != "" {
		plan.AppBind = strings.TrimSpace(options.AppBind)
	}
	if value := strings.TrimSpace(options.DataSource); value != "" && value != "fixture" {
		plan.DataSource = value
	}
	if value := normalizePlugins(options.PluginSet); len(value) > 0 {
		plan.ActivePlugins = value
	}
	if value, ok := parseBoolOverride(options.EnableDevBearer); ok {
		plan.EnableDevBearer = value
	}
	if value := strings.TrimSpace(strings.ToLower(options.LoginPolicy)); value != "" {
		plan.LoginPolicy = value
	}
	if value := strings.TrimSpace(strings.ToLower(options.AdminPolicy)); value != "" {
		plan.AdminPolicy = value
	}
	if value := strings.TrimSpace(options.SitePackageDir); value != "" {
		plan.SitePackageDir = value
	}
	if plan.StorageProfile == string(runtimeprofile.StorageProfileBrowserIndexedDB) {
		plan.BrowserStateless = true
	}
	if plan.RuntimeProfile == string(runtimeprofile.RuntimeProfilePlayground) {
		plan.PlaygroundAuth = true
		plan.BrowserStateless = true
	}
	if plan.DataSource == "" {
		plan.DataSource = defaultDataSource(plan)
	}
	return plan
}

func normalizePreset(raw string) string {
	value := strings.TrimSpace(strings.ToLower(raw))
	if value == "" {
		value = strings.TrimSpace(strings.ToLower(DefaultPreset))
	}
	switch Name(value) {
	case PresetOfflineJSONSQL, PresetSSHFixtures, PresetFull, PresetHeadless, PresetPlayground:
		return value
	default:
		return string(PresetFull)
	}
}

func defaults(name string) Plan {
	switch Name(name) {
	case PresetOfflineJSONSQL:
		return Plan{
			Name:            name,
			RuntimeProfile:  string(runtimeprofile.RuntimeProfileAdmin),
			StorageProfile:  string(runtimeprofile.StorageProfileSQLite),
			AppBind:         "127.0.0.1:8080",
			ActivePlugins:   []string{"json-import-export"},
			EnableDevBearer: true,
			LoginPolicy:     "fixture",
			AdminPolicy:     "enabled",
		}
	case PresetSSHFixtures:
		return Plan{
			Name:            name,
			RuntimeProfile:  string(runtimeprofile.RuntimeProfileAdmin),
			StorageProfile:  string(runtimeprofile.StorageProfileSQLite),
			AppBind:         "127.0.0.1:8080",
			ActivePlugins:   []string{"json-import-export"},
			EnableDevBearer: false,
			LoginPolicy:     "fixture",
			AdminPolicy:     "operator",
		}
	case PresetHeadless:
		return Plan{
			Name:            name,
			RuntimeProfile:  string(runtimeprofile.RuntimeProfileHeadless),
			StorageProfile:  string(runtimeprofile.StorageProfileSQLite),
			AppBind:         "127.0.0.1:8080",
			ActivePlugins:   nil,
			EnableDevBearer: true,
			LoginPolicy:     "disabled",
			AdminPolicy:     "disabled",
		}
	case PresetPlayground:
		return Plan{
			Name:             name,
			RuntimeProfile:   string(runtimeprofile.RuntimeProfilePlayground),
			StorageProfile:   string(runtimeprofile.StorageProfileBrowserIndexedDB),
			AppBind:          "127.0.0.1:8080",
			ActivePlugins:    []string{"playground"},
			PlaygroundAuth:   true,
			BrowserStateless: true,
			EnableDevBearer:  false,
			LoginPolicy:      "playground",
			AdminPolicy:      "enabled",
		}
	default:
		return Plan{
			Name:            string(PresetFull),
			RuntimeProfile:  string(runtimeprofile.RuntimeProfileFull),
			StorageProfile:  string(runtimeprofile.StorageProfileSQLite),
			AppBind:         "127.0.0.1:8080",
			ActivePlugins:   []string{"json-import-export"},
			EnableDevBearer: true,
			LoginPolicy:     "fixture",
			AdminPolicy:     "enabled",
		}
	}
}

func normalizePlugins(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	result := make([]string, 0, len(parts))
	seen := make(map[string]struct{}, len(parts))
	for _, part := range parts {
		id := strings.TrimSpace(strings.ToLower(part))
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result
}

func defaultDataSource(plan Plan) string {
	switch plan.StorageProfile {
	case string(runtimeprofile.StorageProfileMemory):
		return "file:gocms-memory?mode=memory&cache=shared"
	case string(runtimeprofile.StorageProfileBrowserIndexedDB):
		return "file:gocms-playground?mode=memory&cache=shared"
	case string(runtimeprofile.StorageProfileJSONFixtures):
		return "file:gocms-json-fixtures?mode=memory&cache=shared"
	default:
		switch plan.Name {
		case string(PresetOfflineJSONSQL):
			return "file:gocms-offline.db"
		case string(PresetSSHFixtures):
			return "file:gocms-ssh-fixtures.db"
		case string(PresetHeadless):
			return "file:gocms-headless.db"
		case string(PresetPlayground):
			return "file:gocms-playground?mode=memory&cache=shared"
		default:
			return "file:gocms.db"
		}
	}
}

func parseBoolOverride(raw string) (bool, bool) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "1", "true", "yes", "on":
		return true, true
	case "0", "false", "no", "off":
		return false, true
	default:
		return false, false
	}
}
