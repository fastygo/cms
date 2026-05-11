// Package locales defines the built-in locale set for GoCMS (public + admin fixtures,
// LocalizedText fallbacks, and framework defaults). Add new languages here and in
// embedded fixture folders under internal/site/*fixtures/*/.
package locales

import (
	"context"
	"strings"

	"github.com/fastygo/framework/pkg/web/locale"
)

const (
	// Default is the primary site locale (HTML lang, default negotiation).
	Default = "ru"
)

var supported = []string{"ru", "en"}

// Supported returns a copy of all built-in locale codes (order matters for fallbacks).
func Supported() []string {
	out := make([]string, len(supported))
	copy(out, supported)
	return out
}

// IsSupported reports whether code is a known built-in locale.
func IsSupported(code string) bool {
	n := Normalize(code)
	if n == "" {
		return false
	}
	for _, s := range supported {
		if s == n {
			return true
		}
	}
	return false
}

// Normalize returns a lowercased two-letter tag or empty if input is empty.
func Normalize(code string) string {
	code = strings.ToLower(strings.TrimSpace(code))
	if code == "" {
		return ""
	}
	if len(code) > 2 {
		code = code[:2]
	}
	return code
}

// NormalizeOrDefault returns Normalize(code) when supported, otherwise Default.
func NormalizeOrDefault(code string) string {
	if n := Normalize(code); IsSupported(n) {
		return n
	}
	return Default
}

// FromContext returns the active request locale from framework context, normalized
// to a supported built-in locale, or Default when missing/unsupported.
func FromContext(ctx context.Context) string {
	if ctx == nil {
		return Default
	}
	return NormalizeOrDefault(locale.From(ctx))
}

// ContentFallback picks the other supported locale to use as LocalizedText.Value
// second argument (primary is usually FromContext).
func ContentFallback(active string) string {
	active = NormalizeOrDefault(active)
	for _, s := range supported {
		if s != active {
			return s
		}
	}
	return Default
}

// DefaultForI18n is the locale passed to framework i18n loaders as the built-in
// fallback when a requested bundle file is missing.
func DefaultForI18n() string {
	return Default
}
