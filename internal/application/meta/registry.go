package meta

import (
	"fmt"
	"net/url"
	"slices"
	"strconv"
	"strings"

	"github.com/fastygo/cms/internal/domain/authz"
	domaincontent "github.com/fastygo/cms/internal/domain/content"
	domainmeta "github.com/fastygo/cms/internal/domain/meta"
)

const formFieldPrefix = "meta__"

type Registry struct {
	items map[string]domainmeta.Definition
	order []string
}

func NewRegistry(definitions ...domainmeta.Definition) (*Registry, error) {
	registry := &Registry{items: map[string]domainmeta.Definition{}}
	if err := registry.Add(definitions...); err != nil {
		return nil, err
	}
	return registry, nil
}

func (r *Registry) Add(definitions ...domainmeta.Definition) error {
	if r == nil {
		return nil
	}
	for _, definition := range definitions {
		if err := definition.Validate(); err != nil {
			return err
		}
		if _, exists := r.items[definition.Key]; exists {
			return fmt.Errorf("meta definition %q is already registered", definition.Key)
		}
		r.items[definition.Key] = definition
		r.order = append(r.order, definition.Key)
	}
	return nil
}

func (r *Registry) Definitions(kind domaincontent.Kind) []domainmeta.Definition {
	if r == nil {
		return nil
	}
	result := make([]domainmeta.Definition, 0, len(r.order))
	for _, key := range r.order {
		definition := r.items[key]
		if definition.AppliesTo(kind) {
			result = append(result, definition)
		}
	}
	return result
}

func (r *Registry) Definition(kind domaincontent.Kind, key string) (domainmeta.Definition, bool) {
	if r == nil {
		return domainmeta.Definition{}, false
	}
	definition, ok := r.items[key]
	if !ok || !definition.AppliesTo(kind) {
		return domainmeta.Definition{}, false
	}
	return definition, true
}

func (r *Registry) Normalize(principal authz.Principal, kind domaincontent.Kind, metadata domaincontent.Metadata) (domaincontent.Metadata, error) {
	if r == nil {
		return cloneMetadata(metadata), nil
	}
	result := cloneMetadata(metadata)
	if result == nil {
		result = domaincontent.Metadata{}
	}
	for _, definition := range r.Definitions(kind) {
		value, ok := result[definition.Key]
		if !ok {
			if definition.Default == nil {
				continue
			}
			result[definition.Key] = domaincontent.MetaValue{Value: definition.Default, Public: definition.Public}
			continue
		}
		if definition.Capability != "" && !principal.Has(definition.Capability) {
			return nil, fmt.Errorf("capability %q is required for meta %q", definition.Capability, definition.Key)
		}
		normalized, omit, err := normalizeValue(definition, value.Value)
		if err != nil {
			return nil, fmt.Errorf("meta %q %w", definition.Key, err)
		}
		if omit {
			delete(result, definition.Key)
			continue
		}
		result[definition.Key] = domaincontent.MetaValue{Value: normalized, Public: definition.Public}
	}
	return result, nil
}

func (r *Registry) PublicMetadata(kind domaincontent.Kind, metadata domaincontent.Metadata, includePrivate bool) domaincontent.Metadata {
	if len(metadata) == 0 {
		return nil
	}
	result := domaincontent.Metadata{}
	for key, value := range metadata {
		definition, ok := r.Definition(kind, key)
		if ok {
			if includePrivate || definition.Public {
				value.Public = definition.Public
				result[key] = value
			}
			continue
		}
		if includePrivate || value.Public {
			result[key] = value
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

func FormFieldName(key string) string {
	return formFieldPrefix + key
}

func FormFieldKey(field string) (string, bool) {
	if !strings.HasPrefix(field, formFieldPrefix) {
		return "", false
	}
	return strings.TrimPrefix(field, formFieldPrefix), true
}

func DefaultContentDefinitions() []domainmeta.Definition {
	return []domainmeta.Definition{
		{
			Key:         "seo_title",
			Label:       "SEO title",
			Owner:       "core",
			Scope:       domainmeta.ScopeContent,
			Kinds:       []domaincontent.Kind{domaincontent.KindPost, domaincontent.KindPage},
			Type:        domainmeta.ValueTypeString,
			FieldHint:   domainmeta.FieldHintText,
			Public:      true,
			Description: "Optional title override for search engines and social previews.",
			Placeholder: "SEO title",
			Rules:       []domainmeta.ValidationRule{{Name: "max_length", Arg: "70"}},
		},
		{
			Key:         "seo_description",
			Label:       "SEO description",
			Owner:       "core",
			Scope:       domainmeta.ScopeContent,
			Kinds:       []domaincontent.Kind{domaincontent.KindPost, domaincontent.KindPage},
			Type:        domainmeta.ValueTypeText,
			FieldHint:   domainmeta.FieldHintTextarea,
			Public:      true,
			Description: "Optional summary override for public snippets.",
			Placeholder: "SEO description",
			Rules:       []domainmeta.ValidationRule{{Name: "max_length", Arg: "180"}},
		},
		{
			Key:         "seo_canonical_url",
			Label:       "Canonical URL",
			Owner:       "core",
			Scope:       domainmeta.ScopeContent,
			Kinds:       []domaincontent.Kind{domaincontent.KindPost, domaincontent.KindPage},
			Type:        domainmeta.ValueTypeURL,
			FieldHint:   domainmeta.FieldHintText,
			Public:      true,
			Description: "Optional canonical URL override.",
			Placeholder: "https://example.test/post",
		},
		{
			Key:         "seo_noindex",
			Label:       "Noindex",
			Owner:       "core",
			Scope:       domainmeta.ScopeContent,
			Kinds:       []domaincontent.Kind{domaincontent.KindPost, domaincontent.KindPage},
			Type:        domainmeta.ValueTypeBoolean,
			FieldHint:   domainmeta.FieldHintCheckbox,
			Public:      true,
			Default:     false,
			Description: "Request search engines to avoid indexing this content.",
		},
	}
}

func normalizeValue(definition domainmeta.Definition, value any) (any, bool, error) {
	switch definition.Type {
	case domainmeta.ValueTypeString, domainmeta.ValueTypeText:
		trimmed := strings.TrimSpace(stringValue(value))
		if trimmed == "" {
			if definition.Default != nil {
				return definition.Default, false, nil
			}
			if definition.Required {
				return nil, false, fmt.Errorf("is required")
			}
			return nil, true, nil
		}
		if err := validateRules(definition, trimmed); err != nil {
			return nil, false, err
		}
		return trimmed, false, nil
	case domainmeta.ValueTypeURL:
		trimmed := strings.TrimSpace(stringValue(value))
		if trimmed == "" {
			if definition.Default != nil {
				return definition.Default, false, nil
			}
			if definition.Required {
				return nil, false, fmt.Errorf("is required")
			}
			return nil, true, nil
		}
		if !validURL(trimmed) {
			return nil, false, fmt.Errorf("must be a valid URL")
		}
		return trimmed, false, nil
	case domainmeta.ValueTypeBoolean:
		switch typed := value.(type) {
		case bool:
			return typed, false, nil
		default:
			parsed, err := domainmeta.ParseBoolean(stringValue(value))
			if err != nil {
				return nil, false, err
			}
			return parsed, false, nil
		}
	case domainmeta.ValueTypeInteger:
		switch typed := value.(type) {
		case int:
			return typed, false, nil
		case int64:
			return int(typed), false, nil
		case float64:
			if typed != float64(int(typed)) {
				return nil, false, fmt.Errorf("must be an integer")
			}
			return int(typed), false, nil
		default:
			trimmed := strings.TrimSpace(stringValue(value))
			if trimmed == "" {
				if definition.Default != nil {
					return definition.Default, false, nil
				}
				if definition.Required {
					return nil, false, fmt.Errorf("is required")
				}
				return nil, true, nil
			}
			parsed, err := domainmeta.ParseInteger(trimmed)
			if err != nil {
				return nil, false, err
			}
			return parsed, false, nil
		}
	case domainmeta.ValueTypeJSON:
		if value == nil {
			if definition.Default != nil {
				return definition.Default, false, nil
			}
			if definition.Required {
				return nil, false, fmt.Errorf("is required")
			}
			return nil, true, nil
		}
		return value, false, nil
	default:
		return value, false, nil
	}
}

func validateRules(definition domainmeta.Definition, value string) error {
	if len(definition.Options) > 0 {
		valid := false
		for _, option := range definition.Options {
			if option.Value == value {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("must be one of the allowed options")
		}
	}
	for _, rule := range definition.Rules {
		switch strings.TrimSpace(strings.ToLower(rule.Name)) {
		case "required":
			if strings.TrimSpace(value) == "" {
				return fmt.Errorf("is required")
			}
		case "min_length":
			limit, err := strconv.Atoi(rule.Arg)
			if err != nil {
				return fmt.Errorf("has an invalid min_length rule")
			}
			if len(value) < limit {
				return fmt.Errorf("must be at least %d characters", limit)
			}
		case "max_length":
			limit, err := strconv.Atoi(rule.Arg)
			if err != nil {
				return fmt.Errorf("has an invalid max_length rule")
			}
			if len(value) > limit {
				return fmt.Errorf("must be at most %d characters", limit)
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

func cloneMetadata(metadata domaincontent.Metadata) domaincontent.Metadata {
	if len(metadata) == 0 {
		return nil
	}
	result := make(domaincontent.Metadata, len(metadata))
	for key, value := range metadata {
		result[key] = value
	}
	return result
}

func stringValue(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case fmt.Stringer:
		return typed.String()
	case nil:
		return ""
	default:
		return fmt.Sprint(value)
	}
}

func validURL(value string) bool {
	if strings.HasPrefix(value, "/") {
		return true
	}
	parsed, err := url.Parse(value)
	return err == nil && (parsed.Scheme == "http" || parsed.Scheme == "https") && parsed.Host != ""
}
