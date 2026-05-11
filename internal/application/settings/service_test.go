package settings

import (
	"context"
	"testing"

	domainauthz "github.com/fastygo/cms/internal/domain/authz"
	domainsettings "github.com/fastygo/cms/internal/domain/settings"
)

type memoryRepo struct {
	items map[domainsettings.Key]domainsettings.Value
}

func (r *memoryRepo) SaveSetting(_ context.Context, value domainsettings.Value) error {
	if r.items == nil {
		r.items = map[domainsettings.Key]domainsettings.Value{}
	}
	r.items[value.Key] = value
	return nil
}

func (r *memoryRepo) GetSetting(_ context.Context, key domainsettings.Key) (domainsettings.Value, bool, error) {
	value, ok := r.items[key]
	return value, ok, nil
}

func (r *memoryRepo) ListPublicSettings(_ context.Context) ([]domainsettings.Value, error) {
	result := []domainsettings.Value{}
	for _, value := range r.items {
		if value.Public {
			result = append(result, value)
		}
	}
	return result, nil
}

func (r *memoryRepo) ListSettings(_ context.Context) ([]domainsettings.Value, error) {
	result := make([]domainsettings.Value, 0, len(r.items))
	for _, value := range r.items {
		result = append(result, value)
	}
	return result, nil
}

func TestServiceDefaultsCapabilitiesAndPublicProjection(t *testing.T) {
	repo := &memoryRepo{items: map[domainsettings.Key]domainsettings.Value{
		"theme.active": {Key: "theme.active", Value: "company", Public: false},
	}}
	registry, err := NewRegistry(
		domainsettings.Definition{Key: "site.title", Label: "Site title", Owner: "core", Group: domainsettings.GroupCore, Type: domainsettings.ValueTypeString, Public: true, Default: "GoCMS", Capability: domainauthz.CapabilitySettingsManage, Autoload: domainsettings.AutoloadEager},
		domainsettings.Definition{Key: "theme.active", Label: "Active theme", Owner: "core", Group: domainsettings.GroupTheme, Type: domainsettings.ValueTypeString, Public: false, Default: "gocms-default", Capability: domainauthz.CapabilityThemesManage, Autoload: domainsettings.AutoloadEager},
		domainsettings.Definition{Key: "theme.gocms-default.brand_name", Label: "Brand name", Owner: "theme:gocms-default", Group: domainsettings.GroupTheme, Type: domainsettings.ValueTypeString, Public: true, Default: "GoCMS", Capability: domainauthz.CapabilityThemesManage, Autoload: domainsettings.AutoloadLazy},
		domainsettings.Definition{Key: "theme.company.cta_label", Label: "CTA label", Owner: "theme:company", Group: domainsettings.GroupTheme, Type: domainsettings.ValueTypeString, Public: true, Default: "Explore", Capability: domainauthz.CapabilityThemesManage, Autoload: domainsettings.AutoloadLazy},
	)
	if err != nil {
		t.Fatal(err)
	}
	service := NewService(repo, WithRegistry(registry))

	value, ok, err := service.Get(t.Context(), "site.title")
	if err != nil || !ok {
		t.Fatalf("expected default site.title, ok=%v err=%v", ok, err)
	}
	if value.Value != "GoCMS" || !value.Public {
		t.Fatalf("unexpected default value: %+v", value)
	}

	if err := service.Save(t.Context(), domainauthz.NewPrincipal("viewer", domainauthz.CapabilityControlPanelAccess), domainsettings.Value{
		Key:   "site.title",
		Value: "Blocked",
	}); err == nil {
		t.Fatal("expected capability check to reject viewer save")
	}

	if err := service.Save(t.Context(), domainauthz.Root(), domainsettings.Value{
		Key:   "site.title",
		Value: "Typed Title",
	}); err != nil {
		t.Fatal(err)
	}

	publicValues, err := service.Public(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	keys := map[domainsettings.Key]domainsettings.Value{}
	for _, item := range publicValues {
		keys[item.Key] = item
	}
	if got := keys["site.title"].Value; got != "Typed Title" {
		t.Fatalf("site.title = %v", got)
	}
	if _, ok := keys["theme.gocms-default.brand_name"]; ok {
		t.Fatalf("inactive theme setting leaked: %+v", publicValues)
	}
	if got := keys["theme.company.cta_label"].Value; got != "Explore" {
		t.Fatalf("active theme default missing: %+v", publicValues)
	}
}
