package graphqlplugin

import (
	"context"
	"net/http"
	"strings"

	appcontent "github.com/fastygo/cms/internal/application/content"
	appcontenttype "github.com/fastygo/cms/internal/application/contenttype"
	appmedia "github.com/fastygo/cms/internal/application/media"
	appmenus "github.com/fastygo/cms/internal/application/menus"
	appmeta "github.com/fastygo/cms/internal/application/meta"
	appsettings "github.com/fastygo/cms/internal/application/settings"
	apptaxonomy "github.com/fastygo/cms/internal/application/taxonomy"
	appusers "github.com/fastygo/cms/internal/application/users"
	"github.com/fastygo/cms/internal/delivery/rest"
	"github.com/fastygo/cms/internal/domain/authz"
	"github.com/fastygo/cms/internal/platform/plugins"
	"github.com/fastygo/cms/internal/site/adminfixtures"
	"github.com/fastygo/cms/internal/site/ui/elements"
)

const endpointPath = "/go-graphql"

type Services struct {
	Content      appcontent.Service
	ContentTypes appcontenttype.Service
	Taxonomy     apptaxonomy.Service
	Media        appmedia.Service
	Users        appusers.Service
	Settings     appsettings.Service
	Menus        appmenus.Service
}

type Settings struct {
	PublicIntrospection bool
	AuthPolicy          string
	MaxDepth            int
	MaxParallelism      int
	MaxQueryLength      int
	CachePolicy         string
	CORSAllowOrigin     string
}

type Plugin struct {
	handler  Handler
	settings Settings
	resolver *rootResolver
}

func New(services Services, authenticator rest.Authenticator, metaRegistry *appmeta.Registry) (Plugin, error) {
	settings := defaultSettings()
	resolver := &rootResolver{services: services, metaRegistry: metaRegistry}
	handler, err := NewHandler(resolver, authenticator, settings)
	if err != nil {
		return Plugin{}, err
	}
	return Plugin{handler: handler, settings: settings, resolver: resolver}, nil
}

func defaultSettings() Settings {
	return Settings{
		PublicIntrospection: false,
		AuthPolicy:          "public-reads",
		MaxDepth:            12,
		MaxParallelism:      16,
		MaxQueryLength:      32768,
		CachePolicy:         "no-store",
		CORSAllowOrigin:     "",
	}
}

func (p Plugin) Manifest() plugins.Manifest {
	return plugins.Manifest{
		ID:          "graphql",
		Name:        "GraphQL",
		Version:     "0.1.0",
		Contract:    "0.1",
		Description: "Provides the GoCMS GraphQL endpoint over the same application services used by REST and admin.",
		Capabilities: []plugins.CapabilityDefinition{
			{ID: "graphql.manage", Description: "Inspect and manage the GraphQL plugin configuration and endpoint behavior."},
		},
		Settings: []plugins.SettingDefinition{
			{Key: "graphql.endpoint_enabled", Type: "boolean", Default: "true", Public: false, Capability: authz.CapabilitySettingsManage},
			{Key: "graphql.public_introspection", Type: "boolean", Default: "false", Public: false, Capability: authz.CapabilitySettingsManage},
			{Key: "graphql.auth_policy", Type: "string", Default: "public-reads", Public: false, Capability: authz.CapabilitySettingsManage},
			{Key: "graphql.max_depth", Type: "integer", Default: "12", Public: false, Capability: authz.CapabilitySettingsManage},
			{Key: "graphql.max_parallelism", Type: "integer", Default: "16", Public: false, Capability: authz.CapabilitySettingsManage},
			{Key: "graphql.max_query_length", Type: "integer", Default: "32768", Public: false, Capability: authz.CapabilitySettingsManage},
			{Key: "graphql.cache_policy", Type: "string", Default: "no-store", Public: false, Capability: authz.CapabilitySettingsManage},
			{Key: "graphql.cors_allow_origin", Type: "string", Default: "", Public: false, Capability: authz.CapabilitySettingsManage},
		},
		Hooks: []plugins.HookRegistration{
			{HookID: "runtime.status.read", HandlerID: "graphql.status", OwnerID: "graphql", Category: "action", Priority: 100},
		},
	}
}

func (p Plugin) Register(_ context.Context, registry *plugins.Registry) error {
	manifest := p.Manifest()
	if p.resolver != nil {
		p.resolver.registry = registry
	}
	registry.AddCapabilities(manifest.Capabilities...)
	registry.AddSettings(manifest.Settings...)
	registry.AddHooks(manifest.Hooks...)
	registry.AddActionHandlers(plugins.ActionHandlerRegistration{
		Hook:   manifest.Hooks[0],
		Handle: func(context.Context, plugins.HookContext, any) error { return nil },
	})
	registry.AddScreenActions(
		plugins.ScreenActionRegistration{ScreenID: "headless", Build: p.actions},
		plugins.ScreenActionRegistration{ScreenID: "runtime", Build: p.actions},
	)
	registry.AddRoutes(
		plugins.Route{Pattern: "GET " + endpointPath, Surface: plugins.SurfacePublic, Handler: p.handler.Get},
		plugins.Route{Pattern: "POST " + endpointPath, Surface: plugins.SurfacePublic, Handler: p.handler.Post},
		plugins.Route{Pattern: "OPTIONS " + endpointPath, Surface: plugins.SurfacePublic, Handler: p.handler.Options},
		plugins.Route{
			Pattern:          "GET /go-admin/plugins/graphql/status",
			Surface:          plugins.SurfaceAdmin,
			Capability:       authz.CapabilitySettingsManage,
			Protected:        true,
			ProtectedHandler: p.handler.Status,
		},
	)
	return nil
}

func (p Plugin) actions(fixture adminfixtures.AdminFixture) []elements.Action {
	return []elements.Action{
		{
			Label:   fixture.Label("action_graphql_status", "GraphQL status"),
			Href:    "/go-admin/plugins/graphql/status",
			Style:   "outline",
			Enabled: true,
		},
		{
			Label:   fixture.Label("action_graphql_endpoint", "GraphQL endpoint"),
			Href:    endpointPath,
			Style:   "outline",
			Enabled: true,
		},
	}
}

type statusResponse struct {
	Plugin               string   `json:"plugin"`
	Endpoint             string   `json:"endpoint"`
	Active               bool     `json:"active"`
	PublicIntrospection  bool     `json:"public_introspection"`
	AuthPolicy           string   `json:"auth_policy"`
	MaxDepth             int      `json:"max_depth"`
	MaxParallelism       int      `json:"max_parallelism"`
	MaxQueryLength       int      `json:"max_query_length"`
	CachePolicy          string   `json:"cache_policy"`
	CORSAllowOrigin      string   `json:"cors_allow_origin,omitempty"`
	SupportedQueryGroups []string `json:"supported_query_groups"`
	SupportedMutations   []string `json:"supported_mutations"`
}

func (p Plugin) statusPayload() statusResponse {
	return statusResponse{
		Plugin:              "graphql",
		Endpoint:            endpointPath,
		Active:              true,
		PublicIntrospection: p.settings.PublicIntrospection,
		AuthPolicy:          p.settings.AuthPolicy,
		MaxDepth:            p.settings.MaxDepth,
		MaxParallelism:      p.settings.MaxParallelism,
		MaxQueryLength:      p.settings.MaxQueryLength,
		CachePolicy:         p.settings.CachePolicy,
		CORSAllowOrigin:     p.settings.CORSAllowOrigin,
		SupportedQueryGroups: []string{
			"posts", "post", "pages", "page", "contentTypes", "taxonomies", "terms", "media", "authors", "menus", "settings", "search",
		},
		SupportedMutations: []string{
			"createPost", "createPage", "updateContent", "publishContent", "scheduleContent", "trashContent", "restoreContent", "assignTerms", "attachFeaturedMedia", "saveMenu", "saveSetting",
		},
	}
}

func methodAllowed(w http.ResponseWriter, methods ...string) {
	w.Header().Set("Allow", strings.Join(methods, ", "))
	http.Error(w, "Method not allowed.", http.StatusMethodNotAllowed)
}
