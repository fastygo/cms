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
	appsnapshot "github.com/fastygo/cms/internal/application/snapshot"
	apptaxonomy "github.com/fastygo/cms/internal/application/taxonomy"
	appusers "github.com/fastygo/cms/internal/application/users"
	"github.com/fastygo/cms/internal/delivery/admin"
	"github.com/fastygo/cms/internal/delivery/rest"
	"github.com/fastygo/cms/internal/domain/authz"
	"github.com/fastygo/cms/internal/infra/bootstrap"
	platformplugins "github.com/fastygo/cms/internal/platform/plugins"
	"github.com/fastygo/cms/internal/platform/preset"
	"github.com/fastygo/cms/internal/platform/runtimeprofile"
	jsonplugin "github.com/fastygo/cms/internal/plugins/jsonimportexport"
	playgroundplugin "github.com/fastygo/cms/internal/plugins/playground"
	"github.com/fastygo/cms/internal/runtime/fixtures"
	"github.com/fastygo/framework/pkg/app"
)

type Module struct {
	store          bootstrap.Store
	handler        rest.Handler
	adminHandler   admin.Handler
	pluginRegistry *platformplugins.Registry
	contentTypes   appcontenttype.Service
	seedFixtures   bool
	runtimeProfile string
	adminPolicy    string
	runtimeInfo    admin.RuntimeInfo
}

type Options struct {
	DataSource       string
	SessionKey       string
	SeedFixtures     bool
	RuntimeProfile   string
	StorageProfile   string
	ActivePlugins    []string
	SitePackageDir   string
	PlaygroundAuth   bool
	BrowserStateless bool
	EnableDevBearer  bool
	LoginPolicy      string
	AdminPolicy      string
	Preset           string
}

func New(dataSource string, sessionKey string, seedFixtures bool) (*Module, error) {
	plan := preset.Resolve(preset.Options{})
	return NewWithOptions(Options{
		DataSource:       dataSource,
		SessionKey:       sessionKey,
		SeedFixtures:     seedFixtures,
		RuntimeProfile:   plan.RuntimeProfile,
		StorageProfile:   plan.StorageProfile,
		ActivePlugins:    plan.ActivePlugins,
		SitePackageDir:   plan.SitePackageDir,
		PlaygroundAuth:   plan.PlaygroundAuth,
		BrowserStateless: plan.BrowserStateless,
		EnableDevBearer:  plan.EnableDevBearer,
		LoginPolicy:      plan.LoginPolicy,
		AdminPolicy:      plan.AdminPolicy,
		Preset:           plan.Name,
	})
}

func NewWithOptions(options Options) (*Module, error) {
	dataSource := options.DataSource
	seedFixtures := options.SeedFixtures
	if isBrowserLocalProfile(options) {
		dataSource = "file:gocms-playground?mode=memory&cache=shared"
		seedFixtures = false
	}
	bootstrapRuntime, err := bootstrap.NewRegistry().Resolve(bootstrap.ProviderPlan{
		StorageProfile: options.StorageProfile,
		DataSource:     dataSource,
		SitePackageDir: options.SitePackageDir,
	})
	if err != nil {
		return nil, err
	}
	module := &Module{
		store:          bootstrapRuntime.Store,
		contentTypes:   appcontenttype.NewService(bootstrapRuntime.Store),
		seedFixtures:   seedFixtures,
		runtimeProfile: options.RuntimeProfile,
		adminPolicy:    defaultString(options.AdminPolicy, "enabled"),
		runtimeInfo: admin.RuntimeInfo{
			Preset:             options.Preset,
			RuntimeProfile:     options.RuntimeProfile,
			StorageProfile:     options.StorageProfile,
			ContentProvider:    bootstrapRuntime.ContentProvider,
			SitePackage:        bootstrapRuntime.SitePackageDir,
			ActivePlugins:      append([]string(nil), options.ActivePlugins...),
			BrowserStateless:   options.BrowserStateless,
			PlaygroundAuth:     options.PlaygroundAuth,
			EnableDevBearer:    options.EnableDevBearer,
			LoginPolicy:        options.LoginPolicy,
			AdminPolicy:        defaultString(options.AdminPolicy, "enabled"),
			ProviderSwitchRule: "Export JSON/site package first, change the bootstrap provider through deployment config, restart or redeploy, then import into the new provider.",
		},
	}
	services := rest.Services{
		Content:      appcontent.NewService(bootstrapRuntime.Store, bootstrapRuntime.Store, time.Now),
		ContentTypes: module.contentTypes,
		Taxonomy:     apptaxonomy.NewService(bootstrapRuntime.Store, bootstrapRuntime.Store),
		Media:        appmedia.NewService(bootstrapRuntime.Store, bootstrapRuntime.Store),
		Users:        appusers.NewService(bootstrapRuntime.Store),
		Settings:     appsettings.NewService(bootstrapRuntime.Store),
		Menus:        appmenus.NewService(bootstrapRuntime.Store),
	}
	bearerTokens := map[string]authz.Principal(nil)
	if options.EnableDevBearer {
		bearerTokens = rest.DevBearerPrincipals()
	}
	authenticator := rest.NewAuthenticatorWithOptions(options.SessionKey, bearerTokens, rest.AuthenticatorOptions{})
	snapshotService := appsnapshot.NewService(bootstrapRuntime.Store, time.Now)
	pluginRuntime, err := platformplugins.NewRuntime(
		bootstrapRuntime.PluginState,
		jsonplugin.New(snapshotService, bootstrapRuntime.SitePackage),
		playgroundplugin.New(),
	)
	if err != nil {
		_ = bootstrapRuntime.Store.Close(context.Background())
		return nil, err
	}
	module.pluginRegistry, err = pluginRuntime.Activate(context.Background(), options.ActivePlugins)
	if err != nil {
		_ = bootstrapRuntime.Store.Close(context.Background())
		return nil, err
	}
	module.handler = rest.NewHandler(services, authenticator)
	module.adminHandler = admin.NewHandlerWithOptions(admin.Services{
		Content:      services.Content,
		ContentTypes: services.ContentTypes,
		Taxonomy:     services.Taxonomy,
		Media:        services.Media,
		Users:        services.Users,
		Settings:     services.Settings,
		Menus:        services.Menus,
	}, authenticator, options.SessionKey, module.pluginRegistry, admin.HandlerOptions{
		PlaygroundAuth: options.PlaygroundAuth,
		LoginPolicy:    defaultString(options.LoginPolicy, "fixture"),
		RuntimeInfo:    module.runtimeInfo,
	})
	if err := module.Init(context.Background()); err != nil {
		_ = bootstrapRuntime.Store.Close(context.Background())
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
	if m.exposesREST() {
		m.handler.Register(mux)
	}
	if m.exposesAdmin() {
		m.adminHandler.Register(mux)
	}
	if m.pluginRegistry == nil {
		return
	}
	if m.exposesREST() {
		for _, route := range m.pluginRegistry.RoutesForSurface(platformplugins.SurfaceREST) {
			mux.HandleFunc(route.Pattern, route.Handler)
		}
	}
	for _, route := range m.pluginRegistry.RoutesForSurface(platformplugins.SurfacePublic) {
		mux.HandleFunc(route.Pattern, route.Handler)
	}
}

func (m *Module) NavItems() []app.NavItem {
	if !m.exposesAdmin() {
		return nil
	}
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

func (m *Module) exposesREST() bool {
	switch m.runtimeProfile {
	case string(runtimeprofile.RuntimeProfilePlayground):
		return false
	default:
		return true
	}
}

func (m *Module) exposesAdmin() bool {
	if m.adminPolicy == "disabled" {
		return false
	}
	switch m.runtimeProfile {
	case string(runtimeprofile.RuntimeProfileHeadless), string(runtimeprofile.RuntimeProfileConformance):
		return false
	default:
		return true
	}
}

func defaultString(value string, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
