package cms

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	appaudit "github.com/fastygo/cms/internal/application/audit"
	appauthn "github.com/fastygo/cms/internal/application/authn"
	appcontent "github.com/fastygo/cms/internal/application/content"
	appcontenttype "github.com/fastygo/cms/internal/application/contenttype"
	appdiagnostics "github.com/fastygo/cms/internal/application/diagnostics"
	apphealth "github.com/fastygo/cms/internal/application/health"
	appmedia "github.com/fastygo/cms/internal/application/media"
	appmenus "github.com/fastygo/cms/internal/application/menus"
	appmeta "github.com/fastygo/cms/internal/application/meta"
	appsettings "github.com/fastygo/cms/internal/application/settings"
	appsnapshot "github.com/fastygo/cms/internal/application/snapshot"
	apptaxonomy "github.com/fastygo/cms/internal/application/taxonomy"
	appusers "github.com/fastygo/cms/internal/application/users"
	"github.com/fastygo/cms/internal/delivery/admin"
	"github.com/fastygo/cms/internal/delivery/publicsite"
	"github.com/fastygo/cms/internal/delivery/rest"
	"github.com/fastygo/cms/internal/domain/authz"
	domainmeta "github.com/fastygo/cms/internal/domain/meta"
	"github.com/fastygo/cms/internal/infra/bootstrap"
	"github.com/fastygo/cms/internal/platform/cmspanel"
	platformplugins "github.com/fastygo/cms/internal/platform/plugins"
	"github.com/fastygo/cms/internal/platform/preset"
	"github.com/fastygo/cms/internal/platform/runtimeprofile"
	platformthemes "github.com/fastygo/cms/internal/platform/themes"
	graphqlplugin "github.com/fastygo/cms/internal/plugins/graphql"
	jsonplugin "github.com/fastygo/cms/internal/plugins/jsonimportexport"
	playgroundplugin "github.com/fastygo/cms/internal/plugins/playground"
	"github.com/fastygo/cms/internal/runtime/fixtures"
	"github.com/fastygo/framework/pkg/app"
)

type Module struct {
	store          bootstrap.Store
	handler        rest.Handler
	adminHandler   admin.Handler
	publicHandler  publicsite.Handler
	pluginRegistry *platformplugins.Registry
	contentTypes   appcontenttype.Service
	seedFixtures   bool
	runtimeProfile string
	adminPolicy    string
	runtimeInfo    admin.RuntimeInfo
}

type Options struct {
	DataSource        string
	SessionKey        string
	SeedFixtures      bool
	RuntimeProfile    string
	StorageProfile    string
	DeploymentProfile string
	ActivePlugins     []string
	SitePackageDir    string
	PlaygroundAuth    bool
	BrowserStateless  bool
	EnableDevBearer   bool
	LoginPolicy       string
	AdminPolicy       string
	Preset            string
	ExtraDescriptors  []platformplugins.Descriptor
	ExtraMeta         []domainmeta.Definition
}

func New(dataSource string, sessionKey string, seedFixtures bool) (*Module, error) {
	plan := preset.Resolve(preset.Options{})
	return NewWithOptions(Options{
		DataSource:        dataSource,
		SessionKey:        sessionKey,
		SeedFixtures:      seedFixtures,
		RuntimeProfile:    plan.RuntimeProfile,
		StorageProfile:    plan.StorageProfile,
		DeploymentProfile: plan.DeploymentProfile,
		ActivePlugins:     plan.ActivePlugins,
		SitePackageDir:    plan.SitePackageDir,
		PlaygroundAuth:    plan.PlaygroundAuth,
		BrowserStateless:  plan.BrowserStateless,
		EnableDevBearer:   plan.EnableDevBearer,
		LoginPolicy:       plan.LoginPolicy,
		AdminPolicy:       plan.AdminPolicy,
		Preset:            plan.Name,
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
			DeploymentProfile:  options.DeploymentProfile,
			ContentProvider:    bootstrapRuntime.ContentProvider,
			SitePackage:        bootstrapRuntime.SitePackageDir,
			ActivePlugins:      append([]string(nil), options.ActivePlugins...),
			BrowserStateless:   options.BrowserStateless,
			PlaygroundAuth:     options.PlaygroundAuth,
			EnableDevBearer:    options.EnableDevBearer,
			LoginPolicy:        defaultLoginPolicy(options),
			AdminPolicy:        defaultString(options.AdminPolicy, "enabled"),
			ProviderSwitchRule: "Export JSON/site package first, change the bootstrap provider through deployment config, restart or redeploy, then import into the new provider.",
		},
	}
	metaRegistry, err := appmeta.NewRegistry(appmeta.DefaultContentDefinitions()...)
	if err != nil {
		_ = bootstrapRuntime.Store.Close(context.Background())
		return nil, err
	}
	if err := metaRegistry.Add(options.ExtraMeta...); err != nil {
		_ = bootstrapRuntime.Store.Close(context.Background())
		return nil, err
	}
	themeRegistry := platformthemes.DefaultRegistry()
	hookGetter := func() *platformplugins.Registry {
		return module.pluginRegistry
	}
	contentService := appcontent.NewService(
		bootstrapRuntime.Store,
		bootstrapRuntime.Store,
		time.Now,
		appcontent.WithMetadataRegistry(metaRegistry),
		appcontent.WithHookRegistry(hookGetter),
	)
	settingsRegistry, err := appsettings.NewRegistry(appsettings.DefaultDefinitions()...)
	if err != nil {
		_ = bootstrapRuntime.Store.Close(context.Background())
		return nil, err
	}
	if err := settingsRegistry.Add(appsettings.ThemeDefinitions(themeRegistry)...); err != nil {
		_ = bootstrapRuntime.Store.Close(context.Background())
		return nil, err
	}
	if err := settingsRegistry.Add(appsettings.ScreenPreferenceDefinitions(adminPreferenceScreens(), adminPreferenceDefaults())...); err != nil {
		_ = bootstrapRuntime.Store.Close(context.Background())
		return nil, err
	}
	settingsService := appsettings.NewService(
		bootstrapRuntime.Store,
		appsettings.WithRegistry(settingsRegistry),
		appsettings.WithHookRegistry(hookGetter),
	)
	authnService := appauthn.NewService(bootstrapRuntime.Store, bootstrapRuntime.Store)
	auditService := appaudit.NewService(bootstrapRuntime.Store, time.Now)
	diagnosticsService := appdiagnostics.NewService(bootstrapRuntime.Store, time.Now)
	healthService := apphealth.NewService(time.Now,
		apphealth.Check{ID: "database", Label: "Database connectivity", Description: "Checks the active store provider health endpoint.", Run: bootstrapRuntime.Store.HealthCheck},
		apphealth.Check{ID: "migrations", Label: "Schema migrations", Description: "Checks that all declared schema migrations are applied.", Run: func(ctx context.Context) error {
			type migrationStatusChecker interface {
				MigrationStatus(context.Context) error
			}
			checker, ok := bootstrapRuntime.Store.(migrationStatusChecker)
			if !ok {
				return nil
			}
			return checker.MigrationStatus(ctx)
		}},
		apphealth.Check{ID: "snapshot", Label: "Snapshot capability", Description: "Verifies that snapshot export/import workflows are available.", Run: func(context.Context) error {
			if !bootstrapRuntime.ProviderCapabilities.SupportsSnapshots {
				return fmt.Errorf("snapshot workflows are disabled for storage profile %q", bootstrapRuntime.StorageProfile)
			}
			return nil
		}},
		apphealth.Check{ID: "authn", Label: "Authentication store", Description: "Verifies that local auth state is available.", Run: func(context.Context) error {
			if !authnService.Enabled() {
				return fmt.Errorf("local authn service is not configured")
			}
			return nil
		}},
		apphealth.Check{ID: "audit", Label: "Audit log store", Description: "Verifies that audit event storage is available.", Run: func(context.Context) error {
			if !auditService.Enabled() {
				return fmt.Errorf("audit service is not configured")
			}
			return nil
		}},
		apphealth.Check{ID: "error_logs", Label: "Error log store", Description: "Verifies that bounded local error logging is available.", Run: func(context.Context) error {
			if !diagnosticsService.Enabled() {
				return fmt.Errorf("diagnostics service is not configured")
			}
			return nil
		}},
	)
	services := rest.Services{
		Content:      contentService,
		ContentTypes: module.contentTypes,
		Taxonomy:     apptaxonomy.NewService(bootstrapRuntime.Store, bootstrapRuntime.Store),
		Media:        appmedia.NewService(bootstrapRuntime.Store, bootstrapRuntime.Store),
		Users:        appusers.NewService(bootstrapRuntime.Store),
		Settings:     settingsService,
		Menus:        appmenus.NewService(bootstrapRuntime.Store),
		Audit:        auditService,
	}
	bearerTokens := map[string]authz.Principal(nil)
	if options.EnableDevBearer {
		bearerTokens = rest.DevBearerPrincipals()
	}
	authenticator := rest.NewAuthenticatorWithOptions(options.SessionKey, bearerTokens, rest.AuthenticatorOptions{
		SessionSecure:      secureSessionCookies(options.DeploymentProfile),
		AbsoluteSessionTTL: 7 * 24 * time.Hour,
		BearerResolver:     authnService,
	})
	snapshotService := appsnapshot.NewService(bootstrapRuntime.Store, time.Now).WithProviderProfile(bootstrapRuntime.StorageProfile)
	graphqlDescriptor, err := graphqlplugin.New(graphqlplugin.Services{
		Content:      services.Content,
		ContentTypes: services.ContentTypes,
		Taxonomy:     services.Taxonomy,
		Media:        services.Media,
		Users:        services.Users,
		Settings:     services.Settings,
		Menus:        services.Menus,
	}, authenticator, metaRegistry)
	if err != nil {
		_ = bootstrapRuntime.Store.Close(context.Background())
		return nil, err
	}
	descriptors := []platformplugins.Descriptor{
		graphqlDescriptor,
		jsonplugin.New(snapshotService, bootstrapRuntime.SitePackage),
		playgroundplugin.New(),
	}
	descriptors = append(descriptors, options.ExtraDescriptors...)
	pluginRuntime, err := platformplugins.NewRuntime(bootstrapRuntime.PluginState, descriptors...)
	if err != nil {
		_ = bootstrapRuntime.Store.Close(context.Background())
		return nil, err
	}
	module.pluginRegistry, err = pluginRuntime.Activate(context.Background(), options.ActivePlugins)
	if err != nil {
		_ = bootstrapRuntime.Store.Close(context.Background())
		return nil, err
	}
	if err := settingsRegistry.Add(appsettings.PluginDefinitions(module.pluginRegistry.Settings())...); err != nil {
		_ = bootstrapRuntime.Store.Close(context.Background())
		return nil, err
	}
	module.handler = rest.NewHandlerWithOptions(services, authenticator, module.pluginRegistry, metaRegistry)
	module.adminHandler = admin.NewHandlerWithOptions(admin.Services{
		Content:      services.Content,
		ContentTypes: services.ContentTypes,
		Taxonomy:     services.Taxonomy,
		Media:        services.Media,
		Users:        services.Users,
		Settings:     services.Settings,
		Menus:        services.Menus,
		Authn:        authnService,
		Audit:        auditService,
	}, authenticator, options.SessionKey, module.pluginRegistry, admin.HandlerOptions{
		PlaygroundAuth: options.PlaygroundAuth,
		LoginPolicy:    defaultString(options.LoginPolicy, defaultLoginPolicy(options)),
		RuntimeInfo:    module.runtimeInfo,
		Settings:       settingsRegistry,
		ThemeRegistry:  themeRegistry,
		MetaRegistry:   metaRegistry,
		Health:         healthService,
		Diagnostics:    diagnosticsService,
	})
	module.publicHandler = publicsite.NewWithRegistry(publicsite.Services{
		Content:  services.Content,
		Media:    services.Media,
		Menus:    services.Menus,
		Settings: services.Settings,
		Taxonomy: services.Taxonomy,
		Users:    services.Users,
	}, themeRegistry, module.pluginRegistry)
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
		if m.exposesPublic() {
			m.publicHandler.Register(mux)
		}
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
	if m.exposesPublic() {
		m.publicHandler.Register(mux)
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

func (m *Module) exposesPublic() bool {
	return m.runtimeProfile == string(runtimeprofile.RuntimeProfileFull) ||
		m.runtimeProfile == string(runtimeprofile.RuntimeProfilePlayground)
}

func defaultString(value string, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func defaultLoginPolicy(options Options) string {
	if strings.TrimSpace(strings.ToLower(options.LoginPolicy)) != "" {
		return strings.TrimSpace(strings.ToLower(options.LoginPolicy))
	}
	if options.PlaygroundAuth || options.RuntimeProfile == string(runtimeprofile.RuntimeProfilePlayground) {
		return "playground"
	}
	return "local"
}

func secureSessionCookies(deploymentProfile string) bool {
	switch strings.TrimSpace(strings.ToLower(deploymentProfile)) {
	case string(runtimeprofile.DeploymentProfileContainer), string(runtimeprofile.DeploymentProfileServerless), string(runtimeprofile.DeploymentProfileSSH):
		return true
	default:
		return false
	}
}

func adminPreferenceScreens() []string {
	result := make([]string, 0, len(cmspanel.ContentResources())+len(cmspanel.AdminPages()))
	for _, resource := range cmspanel.ContentResources() {
		result = append(result, string(resource.ID))
	}
	for _, page := range cmspanel.AdminPages() {
		result = append(result, string(page.ID))
	}
	return result
}

func adminPreferenceDefaults() map[string]int {
	defaults := map[string]int{}
	for _, resource := range cmspanel.ContentResources() {
		defaults[string(resource.ID)] = firstPerPage(resource.Table.PerPage)
	}
	for _, page := range cmspanel.AdminPages() {
		defaults[string(page.ID)] = firstPerPage(page.Table.PerPage)
	}
	return defaults
}

func firstPerPage(values []int) int {
	if len(values) > 0 && values[0] > 0 {
		return values[0]
	}
	return 25
}
