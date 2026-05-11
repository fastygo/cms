package meta_test

import (
	"testing"

	appmeta "github.com/fastygo/cms/internal/application/meta"
	"github.com/fastygo/cms/internal/domain/authz"
	domaincontent "github.com/fastygo/cms/internal/domain/content"
	domainmeta "github.com/fastygo/cms/internal/domain/meta"
)

func TestRegistryNormalizesRegisteredDefinitionsAndPreservesFlexibleMetadata(t *testing.T) {
	registry, err := appmeta.NewRegistry(
		domainmeta.Definition{
			Key:      "seo_title",
			Label:    "SEO title",
			Owner:    "test",
			Scope:    domainmeta.ScopeContent,
			Kinds:    []domaincontent.Kind{domaincontent.KindPost},
			Type:     domainmeta.ValueTypeString,
			Public:   true,
			Required: true,
		},
		domainmeta.Definition{
			Key:         "secret_token",
			Label:       "Secret token",
			Owner:       "test",
			Scope:       domainmeta.ScopeContent,
			Kinds:       []domaincontent.Kind{domaincontent.KindPost},
			Type:        domainmeta.ValueTypeString,
			Public:      false,
			Capability:  authz.CapabilitySettingsManage,
			Description: "Private secret",
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	normalized, err := registry.Normalize(
		authz.NewPrincipal("admin", authz.CapabilitySettingsManage),
		domaincontent.KindPost,
		domaincontent.Metadata{
			"seo_title":    {Value: "  Search Title  ", Public: false},
			"secret_token": {Value: "top-secret", Public: true},
			"custom_hint":  {Value: "keep-me", Public: true},
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if got := normalized["seo_title"]; got.Value != "Search Title" || !got.Public {
		t.Fatalf("seo_title = %+v", got)
	}
	if got := normalized["secret_token"]; got.Value != "top-secret" || got.Public {
		t.Fatalf("secret_token = %+v", got)
	}
	if got := normalized["custom_hint"]; got.Value != "keep-me" || !got.Public {
		t.Fatalf("custom_hint = %+v", got)
	}

	publicMetadata := registry.PublicMetadata(domaincontent.KindPost, normalized, false)
	if _, ok := publicMetadata["secret_token"]; ok {
		t.Fatalf("private metadata leaked: %+v", publicMetadata)
	}
	if _, ok := publicMetadata["custom_hint"]; !ok {
		t.Fatalf("custom metadata should remain visible when explicitly public: %+v", publicMetadata)
	}
}

func TestRegistryEnforcesCapabilitiesForRegisteredMetadata(t *testing.T) {
	registry, err := appmeta.NewRegistry(domainmeta.Definition{
		Key:        "secret_token",
		Label:      "Secret token",
		Owner:      "test",
		Scope:      domainmeta.ScopeContent,
		Kinds:      []domaincontent.Kind{domaincontent.KindPost},
		Type:       domainmeta.ValueTypeString,
		Public:     false,
		Capability: authz.CapabilitySettingsManage,
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = registry.Normalize(
		authz.NewPrincipal("editor"),
		domaincontent.KindPost,
		domaincontent.Metadata{"secret_token": {Value: "nope", Public: false}},
	)
	if err == nil {
		t.Fatal("expected capability validation error")
	}
}
