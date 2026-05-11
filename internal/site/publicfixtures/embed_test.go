package publicfixtures

import (
	"testing"

	"github.com/fastygo/cms/internal/platform/locales"
)

func TestLoadPublicFixtures(t *testing.T) {
	for _, loc := range locales.Supported() {
		t.Run(loc, func(t *testing.T) {
			pub, err := Load(loc)
			if err != nil {
				t.Fatalf("Load(%q): %v", loc, err)
			}
			if pub.Meta.Lang == "" {
				t.Fatalf("public.meta.lang is empty for locale %q", loc)
			}
			if pub.SiteDefaults.Title == "" {
				t.Fatalf("public.site_defaults.title is empty for locale %q", loc)
			}
			if pub.Routes.BlogTitle == "" {
				t.Fatalf("public.routes.blog_title is empty for locale %q", loc)
			}
		})
	}
}

func TestMustLoadFallsBackToDefaultLocale(t *testing.T) {
	pub := MustLoad("does-not-exist")
	if pub.Routes.BlogTitle == "" {
		t.Fatalf("fallback public routes blog title is empty")
	}
	if pub.Meta.Lang != locales.Default {
		t.Fatalf("fallback public meta.lang = %q, want %q", pub.Meta.Lang, locales.Default)
	}
}
