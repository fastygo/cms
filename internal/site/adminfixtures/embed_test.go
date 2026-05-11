package adminfixtures

import (
	"testing"

	"github.com/fastygo/cms/internal/platform/locales"
)

func TestLoadAdminFixtures(t *testing.T) {
	for _, loc := range Locales {
		t.Run(loc, func(t *testing.T) {
			bundle, err := Load(loc)
			if err != nil {
				t.Fatalf("Load(%q): %v", loc, err)
			}
			if bundle.Admin.Meta.Title == "" {
				t.Fatalf("admin.meta.title is empty for locale %q", loc)
			}
			if len(bundle.Admin.Navigation) == 0 {
				t.Fatalf("admin.navigation is empty for locale %q", loc)
			}
			if len(bundle.Admin.Dashboard.Cards) == 0 {
				t.Fatalf("admin.dashboard.cards is empty for locale %q", loc)
			}
			if _, ok := bundle.Admin.Screen("dashboard"); !ok {
				t.Fatalf("admin.screens.dashboard missing for locale %q", loc)
			}
		})
	}
}

func TestMustLoadFallsBackToDefaultLocale(t *testing.T) {
	fixture := MustLoad("does-not-exist")
	if fixture.Meta.Title == "" {
		t.Fatalf("fallback fixture meta title is empty")
	}
	if fixture.Meta.Lang != locales.Default {
		t.Fatalf("fallback fixture lang = %q, want %q", fixture.Meta.Lang, locales.Default)
	}
}
