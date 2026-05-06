package bootstrap

import (
	"context"
	"testing"
)

func TestResolveUsesConfiguredBootstrapProviders(t *testing.T) {
	runtime, err := NewRegistry().Resolve(ProviderPlan{
		StorageProfile: "sqlite",
		DataSource:     "file:bootstrap-test?mode=memory&cache=shared",
		SitePackageDir: t.TempDir(),
	})
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	t.Cleanup(func() {
		_ = runtime.Store.Close(context.Background())
	})
	if runtime.ContentProvider != "sqlite" {
		t.Fatalf("ContentProvider = %q, want sqlite", runtime.ContentProvider)
	}
	if runtime.StorageProfile != "sqlite" {
		t.Fatalf("StorageProfile = %q, want sqlite", runtime.StorageProfile)
	}
	if runtime.DataSource != "file:bootstrap-test?mode=memory&cache=shared" {
		t.Fatalf("DataSource = %q, want configured data source", runtime.DataSource)
	}
	if runtime.SitePackageDir == "" {
		t.Fatalf("expected site package dir metadata")
	}
	if !runtime.SitePackage.Enabled() {
		t.Fatalf("expected site package provider to be enabled")
	}
	if runtime.PluginState == nil {
		t.Fatalf("expected plugin state repository")
	}
}

func TestResolveRejectsUnimplementedBootstrapProviders(t *testing.T) {
	if _, err := NewRegistry().Resolve(ProviderPlan{StorageProfile: "bbolt"}); err == nil {
		t.Fatalf("expected bbolt provider to be declared but not implemented")
	}
}
