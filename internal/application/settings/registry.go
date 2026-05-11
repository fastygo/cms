package settings

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	domainauthz "github.com/fastygo/cms/internal/domain/authz"
	domainsettings "github.com/fastygo/cms/internal/domain/settings"
	domainthemes "github.com/fastygo/cms/internal/domain/themes"
	platformplugins "github.com/fastygo/cms/internal/platform/plugins"
	platformthemes "github.com/fastygo/cms/internal/platform/themes"
)

type Registry struct {
	items map[domainsettings.Key]domainsettings.Definition
	order []domainsettings.Key
}

func NewRegistry(definitions ...domainsettings.Definition) (*Registry, error) {
	registry := &Registry{items: map[domainsettings.Key]domainsettings.Definition{}}
	if err := registry.Add(definitions...); err != nil {
		return nil, err
	}
	return registry, nil
}

func (r *Registry) Add(definitions ...domainsettings.Definition) error {
	if r == nil {
		return nil
	}
	for _, definition := range definitions {
		if err := definition.Validate(); err != nil {
			return err
		}
		if _, exists := r.items[definition.Key]; exists {
			return fmt.Errorf("setting definition %q is already registered", definition.Key)
		}
		r.items[definition.Key] = definition
		r.order = append(r.order, definition.Key)
	}
	return nil
}

func (r *Registry) Definition(key domainsettings.Key) (domainsettings.Definition, bool) {
	if r == nil {
		return domainsettings.Definition{}, false
	}
	definition, ok := r.items[key]
	return definition, ok
}

func (r *Registry) Definitions() []domainsettings.Definition {
	if r == nil {
		return nil
	}
	result := make([]domainsettings.Definition, 0, len(r.order))
	for _, key := range r.order {
		result = append(result, r.items[key])
	}
	return result
}

func (r *Registry) DefinitionsByGroup(group domainsettings.Group) []domainsettings.Definition {
	result := []domainsettings.Definition{}
	for _, definition := range r.Definitions() {
		if definition.Group == group {
			result = append(result, definition)
		}
	}
	return result
}

func (r *Registry) PublicKeys() []domainsettings.Key {
	keys := []domainsettings.Key{}
	for _, definition := range r.Definitions() {
		if definition.Public {
			keys = append(keys, definition.Key)
		}
	}
	return keys
}

func DefaultDefinitions() []domainsettings.Definition {
	return []domainsettings.Definition{
		{
			Key:         "site.title",
			Label:       "Site title",
			Owner:       "core",
			Group:       domainsettings.GroupCore,
			Type:        domainsettings.ValueTypeString,
			Public:      true,
			Default:     "GoCMS",
			Description: "Public site title.",
			Capability:  domainauthz.CapabilitySettingsManage,
			Autoload:    domainsettings.AutoloadEager,
			Rules:       []domainsettings.ValidationRule{{Name: "required"}},
			FieldHint:   domainsettings.FieldHintText,
		},
		{
			Key:         "public.rendering",
			Label:       "Public rendering",
			Owner:       "core",
			Group:       domainsettings.GroupPublic,
			Type:        domainsettings.ValueTypeSelect,
			Public:      true,
			Default:     "enabled",
			Description: "Enable or disable public page rendering.",
			Capability:  domainauthz.CapabilitySettingsManage,
			Autoload:    domainsettings.AutoloadEager,
			FieldHint:   domainsettings.FieldHintSelect,
			Options: []domainsettings.Option{
				{Value: "enabled", Label: "Enabled"},
				{Value: "disabled", Label: "Disabled"},
			},
		},
		{
			Key:         domainsettings.Key(platformthemes.ActiveThemeKey),
			Label:       "Active theme",
			Owner:       "core",
			Group:       domainsettings.GroupTheme,
			Type:        domainsettings.ValueTypeString,
			Public:      false,
			Default:     string(platformthemes.DefaultThemeID),
			Description: "Currently active theme.",
			Capability:  domainauthz.CapabilityThemesManage,
			Autoload:    domainsettings.AutoloadEager,
			FieldHint:   domainsettings.FieldHintText,
		},
		{
			Key:         domainsettings.Key(platformthemes.StylePresetKey),
			Label:       "Style preset",
			Owner:       "core",
			Group:       domainsettings.GroupTheme,
			Type:        domainsettings.ValueTypeString,
			Public:      false,
			Default:     "default",
			Description: "Active theme style preset.",
			Capability:  domainauthz.CapabilityThemesManage,
			Autoload:    domainsettings.AutoloadEager,
			FieldHint:   domainsettings.FieldHintText,
		},
		{
			Key:         domainsettings.Key(platformthemes.PreviewThemeKey),
			Label:       "Preview theme",
			Owner:       "core",
			Group:       domainsettings.GroupTheme,
			Type:        domainsettings.ValueTypeString,
			Public:      false,
			Default:     string(platformthemes.DefaultThemeID),
			Description: "Preview theme override used by admin.",
			Capability:  domainauthz.CapabilityThemesManage,
			Autoload:    domainsettings.AutoloadLazy,
			FieldHint:   domainsettings.FieldHintText,
		},
		{
			Key:         "theme.preview_preset",
			Label:       "Preview preset",
			Owner:       "core",
			Group:       domainsettings.GroupTheme,
			Type:        domainsettings.ValueTypeString,
			Public:      false,
			Default:     "default",
			Description: "Preview style preset override used by admin.",
			Capability:  domainauthz.CapabilityThemesManage,
			Autoload:    domainsettings.AutoloadLazy,
			FieldHint:   domainsettings.FieldHintText,
		},
		{
			Key:         "permalinks.post_pattern",
			Label:       "Post permalink pattern",
			Owner:       "core",
			Group:       domainsettings.GroupPublic,
			Type:        domainsettings.ValueTypeString,
			Public:      false,
			Default:     "/%postname%/",
			Description: "Permalink pattern for posts.",
			Capability:  domainauthz.CapabilitySettingsManage,
			Autoload:    domainsettings.AutoloadEager,
			Rules:       []domainsettings.ValidationRule{{Name: "required"}},
			FieldHint:   domainsettings.FieldHintText,
		},
		{
			Key:         "permalinks.page_pattern",
			Label:       "Page permalink pattern",
			Owner:       "core",
			Group:       domainsettings.GroupPublic,
			Type:        domainsettings.ValueTypeString,
			Public:      false,
			Default:     "/{slug}/",
			Description: "Permalink pattern for pages.",
			Capability:  domainauthz.CapabilitySettingsManage,
			Autoload:    domainsettings.AutoloadEager,
			Rules:       []domainsettings.ValidationRule{{Name: "required"}},
			FieldHint:   domainsettings.FieldHintText,
		},
		{
			Key:         "admin.editor.provider",
			Label:       "Admin editor provider",
			Owner:       "core",
			Group:       domainsettings.GroupAdmin,
			Type:        domainsettings.ValueTypeString,
			Public:      false,
			Default:     "tiptap-basic",
			Description: "Preferred editor provider in admin.",
			Capability:  domainauthz.CapabilityControlPanelAccess,
			Autoload:    domainsettings.AutoloadLazy,
			FieldHint:   domainsettings.FieldHintText,
		},
		{
			Key:         "auth.session.idle_ttl_hours",
			Label:       "Session idle TTL (hours)",
			Owner:       "core",
			Group:       domainsettings.GroupOperational,
			Type:        domainsettings.ValueTypeInteger,
			Public:      false,
			Default:     12,
			Description: "Local/admin session idle timeout in hours.",
			Capability:  domainauthz.CapabilitySettingsManage,
			Autoload:    domainsettings.AutoloadLazy,
			FieldHint:   domainsettings.FieldHintNumber,
		},
		{
			Key:         "auth.session.absolute_ttl_hours",
			Label:       "Session absolute TTL (hours)",
			Owner:       "core",
			Group:       domainsettings.GroupOperational,
			Type:        domainsettings.ValueTypeInteger,
			Public:      false,
			Default:     168,
			Description: "Maximum local/admin session lifetime in hours.",
			Capability:  domainauthz.CapabilitySettingsManage,
			Autoload:    domainsettings.AutoloadLazy,
			FieldHint:   domainsettings.FieldHintNumber,
		},
		{
			Key:         "auth.login.max_attempts",
			Label:       "Max login attempts",
			Owner:       "core",
			Group:       domainsettings.GroupOperational,
			Type:        domainsettings.ValueTypeInteger,
			Public:      false,
			Default:     3,
			Description: "Maximum failed login attempts before temporary lockout.",
			Capability:  domainauthz.CapabilitySettingsManage,
			Autoload:    domainsettings.AutoloadLazy,
			FieldHint:   domainsettings.FieldHintNumber,
		},
		{
			Key:         "auth.login.attempt_window_minutes",
			Label:       "Login attempt window (minutes)",
			Owner:       "core",
			Group:       domainsettings.GroupOperational,
			Type:        domainsettings.ValueTypeInteger,
			Public:      false,
			Default:     1440,
			Description: "Rolling window used to count failed login attempts.",
			Capability:  domainauthz.CapabilitySettingsManage,
			Autoload:    domainsettings.AutoloadLazy,
			FieldHint:   domainsettings.FieldHintNumber,
		},
		{
			Key:         "auth.login.lockout_minutes",
			Label:       "Login lockout duration (minutes)",
			Owner:       "core",
			Group:       domainsettings.GroupOperational,
			Type:        domainsettings.ValueTypeInteger,
			Public:      false,
			Default:     1440,
			Description: "Temporary lockout duration after too many failed login attempts.",
			Capability:  domainauthz.CapabilitySettingsManage,
			Autoload:    domainsettings.AutoloadLazy,
			FieldHint:   domainsettings.FieldHintNumber,
		},
		{
			Key:         "auth.reset_token.ttl_minutes",
			Label:       "Reset token TTL (minutes)",
			Owner:       "core",
			Group:       domainsettings.GroupOperational,
			Type:        domainsettings.ValueTypeInteger,
			Public:      false,
			Default:     30,
			Description: "Default admin-issued reset token lifetime.",
			Capability:  domainauthz.CapabilitySettingsManage,
			Autoload:    domainsettings.AutoloadLazy,
			FieldHint:   domainsettings.FieldHintNumber,
		},
	}
}

func ThemeDefinitions(registry *platformthemes.Registry) []domainsettings.Definition {
	if registry == nil {
		return nil
	}
	result := []domainsettings.Definition{}
	for _, manifest := range registry.List() {
		for _, setting := range manifest.Settings {
			result = append(result, themeDefinition(manifest.ID, setting))
		}
	}
	return result
}

func ScreenPreferenceDefinitions(screens []string, defaults map[string]int) []domainsettings.Definition {
	result := make([]domainsettings.Definition, 0, len(screens)*2)
	for _, screen := range screens {
		perPageDefault := defaults[screen]
		if perPageDefault <= 0 {
			perPageDefault = 25
		}
		result = append(result,
			domainsettings.Definition{
				Key:         domainsettings.Key("admin.screen." + screen + ".per_page"),
				Label:       screenLabel(screen) + " per page",
				Owner:       "core",
				Group:       domainsettings.GroupAdmin,
				Type:        domainsettings.ValueTypeInteger,
				Public:      false,
				Default:     perPageDefault,
				Description: "Preferred rows per page for the " + screenLabel(screen) + " screen.",
				Capability:  domainauthz.CapabilityControlPanelAccess,
				Autoload:    domainsettings.AutoloadLazy,
				FieldHint:   domainsettings.FieldHintNumber,
			},
			domainsettings.Definition{
				Key:         domainsettings.Key("admin.screen." + screen + ".columns"),
				Label:       screenLabel(screen) + " visible columns",
				Owner:       "core",
				Group:       domainsettings.GroupAdmin,
				Type:        domainsettings.ValueTypeString,
				Public:      false,
				Default:     "",
				Description: "Comma-separated visible columns for the " + screenLabel(screen) + " screen.",
				Capability:  domainauthz.CapabilityControlPanelAccess,
				Autoload:    domainsettings.AutoloadLazy,
				FieldHint:   domainsettings.FieldHintText,
			},
		)
	}
	return result
}

func PluginDefinitions(settings []platformplugins.SettingDefinition) []domainsettings.Definition {
	result := make([]domainsettings.Definition, 0, len(settings))
	for _, setting := range settings {
		result = append(result, pluginDefinition(setting))
	}
	return result
}

func NormalizeValue(definition domainsettings.Definition, value any) (any, error) {
	switch definition.Type {
	case domainsettings.ValueTypeString, domainsettings.ValueTypeText:
		normalized := strings.TrimSpace(stringValue(value))
		if err := validateRules(definition, normalized); err != nil {
			return nil, err
		}
		if normalized == "" && definition.Default != nil {
			return definition.Default, nil
		}
		return normalized, nil
	case domainsettings.ValueTypeBoolean:
		switch typed := value.(type) {
		case bool:
			return typed, nil
		default:
			return domainsettings.ParseBoolean(stringValue(value))
		}
	case domainsettings.ValueTypeInteger:
		switch typed := value.(type) {
		case int:
			return typed, nil
		case int64:
			return int(typed), nil
		case float64:
			if typed != float64(int(typed)) {
				return nil, fmt.Errorf("must be an integer")
			}
			return int(typed), nil
		default:
			trimmed := strings.TrimSpace(stringValue(value))
			if trimmed == "" && definition.Default != nil {
				return definition.Default, nil
			}
			return domainsettings.ParseInteger(trimmed)
		}
	case domainsettings.ValueTypeSelect:
		normalized := strings.TrimSpace(stringValue(value))
		if normalized == "" && definition.Default != nil {
			normalized = stringValue(definition.Default)
		}
		if err := validateRules(definition, normalized); err != nil {
			return nil, err
		}
		if len(definition.Options) > 0 {
			valid := false
			for _, option := range definition.Options {
				if option.Value == normalized {
					valid = true
					break
				}
			}
			if !valid {
				return nil, fmt.Errorf("must be one of the allowed options")
			}
		}
		return normalized, nil
	default:
		if value == nil && definition.Default != nil {
			return definition.Default, nil
		}
		return value, nil
	}
}

func MergeStoredWithDefinitions(registry *Registry, values []domainsettings.Value, publicOnly bool) []domainsettings.Value {
	if registry == nil {
		if publicOnly {
			result := []domainsettings.Value{}
			for _, value := range values {
				if value.Public {
					result = append(result, value)
				}
			}
			return result
		}
		return append([]domainsettings.Value(nil), values...)
	}
	byKey := map[domainsettings.Key]domainsettings.Value{}
	for _, value := range values {
		byKey[value.Key] = value
	}
	activeThemeOwner := activeThemeSettingOwner(registry, byKey)
	result := []domainsettings.Value{}
	for _, definition := range registry.Definitions() {
		if publicOnly && !definition.Public {
			continue
		}
		if publicOnly && definition.Group == domainsettings.GroupTheme && strings.HasPrefix(definition.Owner, "theme:") && definition.Owner != activeThemeOwner {
			continue
		}
		value, ok := byKey[definition.Key]
		if !ok {
			if definition.Default == nil {
				continue
			}
			result = append(result, domainsettings.Value{Key: definition.Key, Value: definition.Default, Public: definition.Public})
			continue
		}
		value.Public = definition.Public
		result = append(result, value)
	}
	return result
}

func visibleColumns(value domainsettings.Value) []string {
	parts := strings.Split(strings.TrimSpace(stringValue(value.Value)), ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}

func VisibleColumns(value domainsettings.Value) []string {
	return visibleColumns(value)
}

func IntValue(value domainsettings.Value, fallback int) int {
	switch typed := value.Value.(type) {
	case int:
		return typed
	case int64:
		return int(typed)
	case float64:
		return int(typed)
	case string:
		parsed, err := strconv.Atoi(strings.TrimSpace(typed))
		if err == nil {
			return parsed
		}
	}
	return fallback
}

func stringValue(value any) string {
	switch typed := value.(type) {
	case nil:
		return ""
	case string:
		return typed
	default:
		return fmt.Sprint(typed)
	}
}

func themeDefinition(themeID domainthemes.ThemeID, setting domainthemes.SettingDefinition) domainsettings.Definition {
	definition := domainsettings.Definition{
		Key:         domainsettings.Key(setting.Key),
		Label:       fallbackLabel(setting.Label, labelFromKey(setting.Key)),
		Owner:       "theme:" + string(themeID),
		Group:       domainsettings.GroupTheme,
		Type:        domainsettings.ValueType(setting.Type),
		Public:      setting.Public,
		Default:     setting.Default,
		Description: setting.Description,
		Capability:  domainauthz.CapabilityThemesManage,
		Autoload:    domainsettings.AutoloadLazy,
		FieldHint:   fieldHint(domainsettings.ValueType(setting.Type)),
		Rules:       parseValidation(setting.Validation),
	}
	if definition.Type == "" {
		definition.Type = domainsettings.ValueTypeString
	}
	return definition
}

func pluginDefinition(setting platformplugins.SettingDefinition) domainsettings.Definition {
	definition := domainsettings.Definition{
		Key:         domainsettings.Key(setting.Key),
		Label:       labelFromKey(setting.Key),
		Owner:       "plugin:" + ownerFromKey(setting.Key),
		Group:       domainsettings.GroupPlugin,
		Type:        domainsettings.ValueType(setting.Type),
		Public:      setting.Public,
		Default:     parsePluginDefault(setting.Type, setting.Default),
		Description: "Plugin-managed setting.",
		Capability:  setting.Capability,
		Autoload:    domainsettings.AutoloadLazy,
		FieldHint:   fieldHint(domainsettings.ValueType(setting.Type)),
	}
	if definition.Type == "" {
		definition.Type = domainsettings.ValueTypeString
	}
	if definition.Capability == "" {
		definition.Capability = domainauthz.CapabilitySettingsManage
	}
	return definition
}

func parsePluginDefault(typ string, value string) any {
	switch domainsettings.ValueType(typ) {
	case domainsettings.ValueTypeBoolean:
		parsed, err := domainsettings.ParseBoolean(value)
		if err == nil {
			return parsed
		}
	case domainsettings.ValueTypeInteger:
		parsed, err := domainsettings.ParseInteger(value)
		if err == nil {
			return parsed
		}
	}
	return value
}

func validateRules(definition domainsettings.Definition, value string) error {
	for _, rule := range definition.Rules {
		switch strings.TrimSpace(strings.ToLower(rule.Name)) {
		case "required":
			if strings.TrimSpace(value) == "" {
				return fmt.Errorf("is required")
			}
		case "one_of":
			allowed := strings.Split(rule.Arg, "|")
			if !slices.Contains(allowed, value) {
				return fmt.Errorf("must be one of %s", strings.Join(allowed, ", "))
			}
		}
	}
	return nil
}

func parseValidation(value string) []domainsettings.ValidationRule {
	value = strings.TrimSpace(strings.ToLower(value))
	if value == "" {
		return nil
	}
	if value == "required" {
		return []domainsettings.ValidationRule{{Name: "required"}}
	}
	return nil
}

func fieldHint(valueType domainsettings.ValueType) domainsettings.FieldHint {
	switch valueType {
	case domainsettings.ValueTypeBoolean:
		return domainsettings.FieldHintCheckbox
	case domainsettings.ValueTypeInteger:
		return domainsettings.FieldHintNumber
	case domainsettings.ValueTypeText:
		return domainsettings.FieldHintTextarea
	case domainsettings.ValueTypeSelect:
		return domainsettings.FieldHintSelect
	default:
		return domainsettings.FieldHintText
	}
}

func ownerFromKey(key string) string {
	parts := strings.SplitN(key, ".", 2)
	return parts[0]
}

func labelFromKey(key string) string {
	parts := strings.Split(key, ".")
	for i, part := range parts {
		parts[i] = strings.Title(strings.ReplaceAll(part, "_", " "))
	}
	return strings.Join(parts, " ")
}

func screenLabel(screen string) string {
	return strings.Title(strings.ReplaceAll(screen, "-", " "))
}

func fallbackLabel(value string, fallback string) string {
	if strings.TrimSpace(value) != "" {
		return value
	}
	return fallback
}

func activeThemeSettingOwner(registry *Registry, values map[domainsettings.Key]domainsettings.Value) string {
	const activeThemeKey = domainsettings.Key(platformthemes.ActiveThemeKey)
	if value, ok := values[activeThemeKey]; ok {
		return "theme:" + strings.TrimSpace(stringValue(value.Value))
	}
	if registry == nil {
		return ""
	}
	definition, ok := registry.Definition(activeThemeKey)
	if !ok {
		return ""
	}
	return "theme:" + strings.TrimSpace(stringValue(definition.Default))
}
