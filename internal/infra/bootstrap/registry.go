package bootstrap

import (
	"context"
	"fmt"

	appaudit "github.com/fastygo/cms/internal/application/audit"
	appauthn "github.com/fastygo/cms/internal/application/authn"
	appcontent "github.com/fastygo/cms/internal/application/content"
	appcontenttype "github.com/fastygo/cms/internal/application/contenttype"
	appdiagnostics "github.com/fastygo/cms/internal/application/diagnostics"
	appmedia "github.com/fastygo/cms/internal/application/media"
	appmenus "github.com/fastygo/cms/internal/application/menus"
	appsettings "github.com/fastygo/cms/internal/application/settings"
	"github.com/fastygo/cms/internal/application/snapshot"
	apptaxonomy "github.com/fastygo/cms/internal/application/taxonomy"
	appusers "github.com/fastygo/cms/internal/application/users"
	"github.com/fastygo/cms/internal/platform/plugins"
	"github.com/fastygo/cms/internal/platform/runtimeprofile"
	"github.com/fastygo/cms/internal/runtime/fixtures"
	"github.com/fastygo/cms/internal/sitepackage/jsondir"
	sqlitestore "github.com/fastygo/cms/internal/storage/sqlite"
)

type Store interface {
	appcontent.Repository
	appcontent.TypeRegistry
	appcontenttype.Repository
	appauthn.Repository
	appaudit.Repository
	appdiagnostics.Repository
	apptaxonomy.Repository
	apptaxonomy.EntryRepository
	appmedia.Repository
	appmedia.EntryRepository
	appusers.Repository
	appsettings.Repository
	appmenus.Repository
	fixtures.Store
	snapshot.Repository
	Init(context.Context) error
	Close(context.Context) error
	HealthCheck(context.Context) error
}

type ProviderPlan struct {
	StorageProfile string
	DataSource     string
	SitePackageDir string
}

type Runtime struct {
	Store                Store
	PluginState          plugins.StateRepository
	SitePackage          jsondir.Provider
	ContentProvider      string
	ProviderCapabilities ProviderCapabilities
	StorageProfile       string
	DataSource           string
	SitePackageDir       string
}

type ProviderCapabilities struct {
	Durable             bool
	Ephemeral           bool
	BrowserLocal        bool
	Transitional        bool
	UsesSQLiteShim      bool
	SupportsHealthCheck bool
	SupportsSnapshots   bool
	RequiresMigrations  bool
	BlobStorageBoundary bool
	Notes               []string
}

type Registry struct{}

func NewRegistry() Registry {
	return Registry{}
}

func (Registry) Resolve(plan ProviderPlan) (Runtime, error) {
	store, providerName, err := openStore(plan.StorageProfile, plan.DataSource)
	if err != nil {
		return Runtime{}, err
	}
	return Runtime{
		Store:                store,
		PluginState:          plugins.NewInMemoryStateRepository(),
		SitePackage:          jsondir.Provider{Dir: plan.SitePackageDir},
		ContentProvider:      providerName,
		ProviderCapabilities: providerCapabilities(plan.StorageProfile),
		StorageProfile:       plan.StorageProfile,
		DataSource:           plan.DataSource,
		SitePackageDir:       plan.SitePackageDir,
	}, nil
}

func providerCapabilities(storageProfile string) ProviderCapabilities {
	switch storageProfile {
	case string(runtimeprofile.StorageProfileBrowserIndexedDB):
		return ProviderCapabilities{
			Ephemeral:           true,
			BrowserLocal:        true,
			Transitional:        true,
			UsesSQLiteShim:      true,
			SupportsHealthCheck: true,
			SupportsSnapshots:   true,
			BlobStorageBoundary: true,
			Notes: []string{
				"browser-indexeddb currently resolves to an in-memory SQLite shim",
				"future browser-local storage must use an IndexedDB/WASM provider boundary",
			},
		}
	case string(runtimeprofile.StorageProfileMemory), string(runtimeprofile.StorageProfileJSONFixtures):
		return ProviderCapabilities{
			Ephemeral:           true,
			SupportsHealthCheck: true,
			SupportsSnapshots:   true,
		}
	case string(runtimeprofile.StorageProfileMySQL), string(runtimeprofile.StorageProfilePostgres), string(runtimeprofile.StorageProfileBbolt):
		return ProviderCapabilities{
			Durable:             true,
			RequiresMigrations:  true,
			BlobStorageBoundary: true,
			Notes: []string{
				"provider is declared but not implemented yet",
			},
		}
	default:
		return ProviderCapabilities{
			Durable:             true,
			SupportsHealthCheck: true,
			SupportsSnapshots:   true,
			RequiresMigrations:  true,
			BlobStorageBoundary: true,
		}
	}
}

func openStore(storageProfile string, dataSource string) (Store, string, error) {
	switch storageProfile {
	case "", string(runtimeprofile.StorageProfileSQLite):
		store, err := sqlitestore.Open(dataSource)
		return store, string(runtimeprofile.StorageProfileSQLite), err
	case string(runtimeprofile.StorageProfileMemory):
		store, err := sqlitestore.Open("file:gocms-memory?mode=memory&cache=shared")
		return store, string(runtimeprofile.StorageProfileMemory), err
	case string(runtimeprofile.StorageProfileBrowserIndexedDB):
		store, err := sqlitestore.Open("file:gocms-playground?mode=memory&cache=shared")
		return store, string(runtimeprofile.StorageProfileBrowserIndexedDB), err
	case string(runtimeprofile.StorageProfileJSONFixtures):
		store, err := sqlitestore.Open("file:gocms-json-fixtures?mode=memory&cache=shared")
		return store, string(runtimeprofile.StorageProfileJSONFixtures), err
	case string(runtimeprofile.StorageProfileBbolt):
		return nil, string(runtimeprofile.StorageProfileBbolt), fmt.Errorf("bootstrap provider %q is declared but not implemented yet", runtimeprofile.StorageProfileBbolt)
	case string(runtimeprofile.StorageProfileMySQL):
		return nil, string(runtimeprofile.StorageProfileMySQL), fmt.Errorf("bootstrap provider %q is declared but not implemented yet", runtimeprofile.StorageProfileMySQL)
	case string(runtimeprofile.StorageProfilePostgres):
		return nil, string(runtimeprofile.StorageProfilePostgres), fmt.Errorf("bootstrap provider %q is declared but not implemented yet", runtimeprofile.StorageProfilePostgres)
	default:
		return nil, storageProfile, fmt.Errorf("bootstrap provider %q is not supported", storageProfile)
	}
}
