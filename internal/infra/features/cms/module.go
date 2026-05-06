package cms

import (
	"context"
	"net/http"
	"time"

	appcontent "github.com/fastygo/cms/internal/application/content"
	appcontenttype "github.com/fastygo/cms/internal/application/contenttype"
	appmedia "github.com/fastygo/cms/internal/application/media"
	appmenus "github.com/fastygo/cms/internal/application/menus"
	appsettings "github.com/fastygo/cms/internal/application/settings"
	apptaxonomy "github.com/fastygo/cms/internal/application/taxonomy"
	appusers "github.com/fastygo/cms/internal/application/users"
	"github.com/fastygo/cms/internal/delivery/admin"
	"github.com/fastygo/cms/internal/delivery/rest"
	"github.com/fastygo/cms/internal/platform/runtimeprofile"
	"github.com/fastygo/cms/internal/runtime/fixtures"
	sqlitestore "github.com/fastygo/cms/internal/storage/sqlite"
	"github.com/fastygo/framework/pkg/app"
)

type Module struct {
	store        *sqlitestore.Store
	handler      rest.Handler
	adminHandler admin.Handler
	contentTypes appcontenttype.Service
	seedFixtures bool
}

type Options struct {
	DataSource     string
	SessionKey     string
	SeedFixtures   bool
	RuntimeProfile string
	StorageProfile string
}

func New(dataSource string, sessionKey string, seedFixtures bool) (*Module, error) {
	return NewWithOptions(Options{
		DataSource:   dataSource,
		SessionKey:   sessionKey,
		SeedFixtures: seedFixtures,
	})
}

func NewWithOptions(options Options) (*Module, error) {
	dataSource := options.DataSource
	seedFixtures := options.SeedFixtures
	if isBrowserLocalProfile(options) {
		dataSource = "file:gocms-playground?mode=memory&cache=shared"
		seedFixtures = false
	}
	store, err := sqlitestore.Open(dataSource)
	if err != nil {
		return nil, err
	}
	module := &Module{
		store:        store,
		contentTypes: appcontenttype.NewService(store),
		seedFixtures: seedFixtures,
	}
	services := rest.Services{
		Content:      appcontent.NewService(store, store, time.Now),
		ContentTypes: module.contentTypes,
		Taxonomy:     apptaxonomy.NewService(store, store),
		Media:        appmedia.NewService(store, store),
		Users:        appusers.NewService(store),
		Settings:     appsettings.NewService(store),
		Menus:        appmenus.NewService(store),
	}
	authenticator := rest.NewAuthenticator(options.SessionKey, rest.DevBearerPrincipals())
	module.handler = rest.NewHandler(services, authenticator)
	module.adminHandler = admin.NewHandler(admin.Services{
		Content:      services.Content,
		ContentTypes: services.ContentTypes,
		Taxonomy:     services.Taxonomy,
		Media:        services.Media,
		Users:        services.Users,
		Settings:     services.Settings,
		Menus:        services.Menus,
	}, authenticator, options.SessionKey)
	if err := module.Init(context.Background()); err != nil {
		_ = store.Close(context.Background())
		return nil, err
	}
	return module, nil
}

func isBrowserLocalProfile(options Options) bool {
	return options.RuntimeProfile == string(runtimeprofile.RuntimeProfilePlayground) ||
		options.StorageProfile == string(runtimeprofile.StorageProfileBrowserIndexedDB)
}

func (m *Module) ID() string {
	return "cms"
}

func (m *Module) Routes(mux *http.ServeMux) {
	m.handler.Register(mux)
	m.adminHandler.Register(mux)
}

func (m *Module) NavItems() []app.NavItem {
	return m.adminHandler.NavItems()
}

func (m *Module) Init(ctx context.Context) error {
	if err := m.store.Init(ctx); err != nil {
		return err
	}
	if err := m.contentTypes.InstallBuiltIns(ctx); err != nil {
		return err
	}
	if m.seedFixtures {
		return fixtures.Seed(ctx, m.store)
	}
	return nil
}

func (m *Module) Close(ctx context.Context) error {
	return m.store.Close(ctx)
}

func (m *Module) HealthCheck(ctx context.Context) error {
	return m.store.HealthCheck(ctx)
}
