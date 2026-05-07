package admin

import (
	"context"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	appcontent "github.com/fastygo/cms/internal/application/content"
	appcontenttype "github.com/fastygo/cms/internal/application/contenttype"
	appmedia "github.com/fastygo/cms/internal/application/media"
	appmenus "github.com/fastygo/cms/internal/application/menus"
	appsettings "github.com/fastygo/cms/internal/application/settings"
	apptaxonomy "github.com/fastygo/cms/internal/application/taxonomy"
	appusers "github.com/fastygo/cms/internal/application/users"
	"github.com/fastygo/cms/internal/delivery/rest"
	"github.com/fastygo/cms/internal/domain/authz"
	domaincontent "github.com/fastygo/cms/internal/domain/content"
	domaincontenttype "github.com/fastygo/cms/internal/domain/contenttype"
	domainmedia "github.com/fastygo/cms/internal/domain/media"
	domainmenus "github.com/fastygo/cms/internal/domain/menus"
	domainsettings "github.com/fastygo/cms/internal/domain/settings"
	domaintaxonomy "github.com/fastygo/cms/internal/domain/taxonomy"
	domainthemes "github.com/fastygo/cms/internal/domain/themes"
	domainusers "github.com/fastygo/cms/internal/domain/users"
	"github.com/fastygo/cms/internal/platform/cmspanel"
	"github.com/fastygo/cms/internal/platform/plugins"
	"github.com/fastygo/cms/internal/platform/runtimeprofile"
	platformthemes "github.com/fastygo/cms/internal/platform/themes"
	"github.com/fastygo/cms/internal/site/adminfixtures"
	"github.com/fastygo/cms/internal/site/assets"
	"github.com/fastygo/cms/internal/site/ui/blocks"
	"github.com/fastygo/cms/internal/site/ui/elements"
	"github.com/fastygo/cms/internal/site/views"
	"github.com/fastygo/framework/pkg/app"
	frameworkauth "github.com/fastygo/framework/pkg/auth"
	"github.com/fastygo/framework/pkg/web"
	"github.com/fastygo/framework/pkg/web/locale"
	"github.com/fastygo/framework/pkg/web/view"
	"github.com/fastygo/panel"
	"github.com/fastygo/ui8kit/ui"
)

type Services struct {
	Content      appcontent.Service
	ContentTypes appcontenttype.Service
	Taxonomy     apptaxonomy.Service
	Media        appmedia.Service
	Users        appusers.Service
	Settings     appsettings.Service
	Menus        appmenus.Service
}

type Handler struct {
	services       Services
	auth           rest.Authenticator
	secret         string
	registry       *plugins.Registry
	playgroundAuth bool
	loginPolicy    string
	runtimeInfo    RuntimeInfo
	themeRegistry  *platformthemes.Registry
}

type RuntimeInfo struct {
	Preset             string
	RuntimeProfile     string
	StorageProfile     string
	ContentProvider    string
	SitePackage        string
	ActivePlugins      []string
	BrowserStateless   bool
	PlaygroundAuth     bool
	EnableDevBearer    bool
	LoginPolicy        string
	AdminPolicy        string
	ProviderSwitchRule string
}

type HandlerOptions struct {
	PlaygroundAuth bool
	LoginPolicy    string
	RuntimeInfo    RuntimeInfo
	ThemeRegistry  *platformthemes.Registry
}

type actionToken struct {
	Action string `json:"action"`
	Exp    int64  `json:"exp"`
}

type dashboardCardValue struct {
	Key   string
	Value string
	Href  string
}

const (
	defaultEditorProviderID  = "tiptap-basic"
	editorProviderSettingKey = "admin.editor.provider"
)

func NewHandler(services Services, authenticator rest.Authenticator, secret string, registry *plugins.Registry, playgroundAuth bool) Handler {
	return NewHandlerWithOptions(services, authenticator, secret, registry, HandlerOptions{PlaygroundAuth: playgroundAuth})
}

func NewHandlerWithOptions(services Services, authenticator rest.Authenticator, secret string, registry *plugins.Registry, options HandlerOptions) Handler {
	if registry == nil {
		registry = plugins.NewRegistry()
	}
	if options.LoginPolicy == "" {
		options.LoginPolicy = "fixture"
	}
	handler := Handler{
		services:       services,
		auth:           authenticator,
		secret:         secret,
		registry:       registry,
		playgroundAuth: options.PlaygroundAuth,
		loginPolicy:    options.LoginPolicy,
		runtimeInfo:    options.RuntimeInfo,
		themeRegistry:  options.ThemeRegistry,
	}
	if handler.themeRegistry == nil {
		handler.themeRegistry = platformthemes.DefaultRegistry()
	}
	handler.registerCoreScreens()
	return handler
}

func (h Handler) registerCoreScreens() {
	h.registry.AddEditorProviders(plugins.EditorProviderRegistration{
		ID:          defaultEditorProviderID,
		Label:       "TipTap (Basic)",
		Description: "Built-in TipTap editor with starter formatting extensions and HTML storage.",
		Priority:    0,
	})
	h.registry.AddAssets(plugins.Asset{
		ID:      "admin-editor-tiptap-basic",
		Surface: plugins.SurfaceAdmin,
		Path:    "/static/js/admin-editor.js",
	})
	contentRoutes := []plugins.Route{}
	for _, resource := range cmspanel.ContentResources() {
		h.registry.AddAdminMenu(resource.Navigation)
		contentRoutes = append(contentRoutes, h.contentResourceRoutes(resource)...)
	}
	pageRoutes := []plugins.Route{}
	for _, page := range cmspanel.AdminPages() {
		if page.Navigation.Path != "" {
			h.registry.AddAdminMenu(page.Navigation)
		}
		pageRoutes = append(pageRoutes, h.adminPageRoutes(page)...)
	}
	h.registry.AddRoutes(pageRoutes...)
	h.registry.AddRoutes(contentRoutes...)
}

func (h Handler) coreRoute(pattern string, capability authz.Capability, handler func(http.ResponseWriter, *http.Request, authz.Principal)) plugins.Route {
	return plugins.Route{
		Pattern:          pattern,
		Surface:          plugins.SurfaceAdmin,
		Capability:       capability,
		Protected:        true,
		ProtectedHandler: handler,
	}
}

func (h Handler) contentResourceRoutes(resource cmspanel.ContentResource) []plugins.Route {
	routes := make([]plugins.Route, 0, len(resource.Routes))
	for _, route := range resource.Routes {
		var handler func(http.ResponseWriter, *http.Request, authz.Principal)
		switch route.Role {
		case panel.RouteIndex:
			handler = h.contentList(resource.Kind, string(resource.ID))
		case panel.RouteNew:
			handler = h.contentNew(resource.Kind, string(resource.ID))
		case panel.RouteCreate:
			handler = h.contentCreate(resource.Kind)
		case panel.RouteEdit:
			handler = h.contentEdit(string(resource.ID))
		case panel.RouteUpdate:
			handler = h.contentUpdate
		case panel.RouteDelete:
			handler = h.contentTrash
		default:
			continue
		}
		routes = append(routes, h.coreRoute(route.Pattern, route.Capability, handler))
	}
	return routes
}

func (h Handler) adminPageRoutes(page cmspanel.AdminPage) []plugins.Route {
	routes := make([]plugins.Route, 0, len(page.Routes))
	for _, route := range page.Routes {
		handler := h.adminPageHandler(string(page.ID), route.Role)
		if handler == nil {
			continue
		}
		routes = append(routes, h.coreRoute(route.Pattern, route.Capability, handler))
	}
	return routes
}

func (h Handler) adminPageHandler(pageID string, role cmspanel.AdminRouteRole) func(http.ResponseWriter, *http.Request, authz.Principal) {
	switch pageID {
	case "dashboard":
		if role == cmspanel.AdminRouteIndex {
			return h.dashboard
		}
	case "content-types":
		if role == cmspanel.AdminRouteCreate {
			return h.contentTypeCreate
		}
		return h.contentTypesPage
	case "taxonomies":
		if role == cmspanel.AdminRouteCreate {
			return h.taxonomyCreate
		}
		return h.taxonomiesPage
	case "terms":
		if role == cmspanel.AdminRouteCreate {
			return h.termCreate
		}
		return h.termsPage
	case "media":
		if role == cmspanel.AdminRouteCreate {
			return h.mediaSave
		}
		return h.mediaPage
	case "menus":
		if role == cmspanel.AdminRouteCreate {
			return h.menuSave
		}
		return h.menusPage
	case "users":
		if role == cmspanel.AdminRouteCreate {
			return h.userSave
		}
		return h.usersPage
	case "authors":
		return h.authorsPage
	case "capabilities":
		return h.capabilitiesPage
	case "settings":
		if role == cmspanel.AdminRouteUpdate {
			return h.settingsSave
		}
		return h.settingsPage
	case "themes":
		if role == cmspanel.AdminRouteUpdate {
			return h.themesSave
		}
		return h.themesPage
	case "permalinks":
		if role == cmspanel.AdminRouteUpdate {
			return h.permalinksSave
		}
		return h.permalinksPage
	case "headless":
		return h.headlessPage
	case "runtime":
		return h.runtimePage
	}
	return nil
}

func (h Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /go-login", h.loginPage)
	mux.HandleFunc("POST /go-login", h.loginSubmit)
	mux.HandleFunc("POST /go-logout", h.logoutSubmit)
	for _, route := range h.registry.RoutesForSurface(plugins.SurfaceAdmin) {
		if route.Protected {
			mux.HandleFunc(route.Pattern, h.protectCapability(route.Capability, route.ProtectedHandler))
			continue
		}
		mux.HandleFunc(route.Pattern, route.Handler)
	}
}

func (h Handler) NavItems() []app.NavItem {
	return h.navigation(adminfixtures.MustLoad("en"), authz.Root())
}

func (h Handler) NavItemsFromBundle(bundle adminfixtures.AdminFixture) []app.NavItem {
	return h.navigation(bundle, authz.Root())
}

func (h Handler) navigation(bundle adminfixtures.AdminFixture, principal authz.Principal) []app.NavItem {
	menu := h.registry.AdminMenuItems(principal)
	items := make([]app.NavItem, 0, len(menu))
	for _, item := range menu {
		label := item.Label
		if screen, ok := bundle.Screen(item.ID); ok && screen.Title != "" {
			label = screen.Title
		}
		items = append(items, app.NavItem{Label: label, Path: item.Path, Icon: item.Icon, Order: item.Order})
	}
	return items
}

func (h Handler) fixture(r *http.Request) adminfixtures.AdminFixture {
	if r == nil {
		return adminfixtures.MustLoad("en")
	}
	return adminfixtures.MustLoad(locale.From(r.Context()))
}

func (h Handler) loginPage(w http.ResponseWriter, r *http.Request) {
	fixture := h.fixture(r)
	data := views.LoginPageData{
		Title:         fixture.Login.Title,
		Subtitle:      fixture.Login.Subtitle,
		Lang:          fixture.Meta.Lang,
		ReturnTo:      returnTo(r),
		ActionToken:   h.token("login"),
		EmailLabel:    fixture.Label("field_email", "Email"),
		PasswordLabel: fixture.Label("field_password", "Password"),
		SubmitLabel:   fixture.Label("action_sign_in", "Sign in"),
	}
	_ = web.Render(r.Context(), w, views.LoginPage(data))
}

func (h Handler) loginSubmit(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data.", http.StatusBadRequest)
		return
	}
	if !h.validToken(r.PostForm.Get("action_token"), "login") {
		http.Error(w, "Invalid action token.", http.StatusForbidden)
		return
	}
	fixture := h.fixture(r)
	session, ok := h.sessionForLogin(r.PostForm.Get("email"), r.PostForm.Get("password"))
	if !ok {
		data := views.LoginPageData{
			Title:         fixture.Login.Title,
			Subtitle:      fixture.Login.Subtitle,
			Error:         fixture.Login.ErrorInvalidCredentials,
			Lang:          fixture.Meta.Lang,
			ReturnTo:      returnTo(r),
			ActionToken:   h.token("login"),
			EmailLabel:    fixture.Label("field_email", "Email"),
			PasswordLabel: fixture.Label("field_password", "Password"),
			SubmitLabel:   fixture.Label("action_sign_in", "Sign in"),
		}
		_ = web.Render(r.Context(), w, views.LoginPage(data))
		return
	}
	if err := h.auth.Session.Issue(w, session); err != nil {
		http.Error(w, "Unable to issue session.", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, safeReturnTo(r.PostForm.Get("return_to")), http.StatusSeeOther)
}

func (h Handler) logoutSubmit(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err == nil && h.validToken(r.PostForm.Get("action_token"), "logout") {
		h.auth.Session.Clear(w)
		http.Redirect(w, r, "/go-login", http.StatusSeeOther)
		return
	}
	http.Error(w, "Invalid action token.", http.StatusForbidden)
}

func (h Handler) protect(next func(http.ResponseWriter, *http.Request, authz.Principal)) http.HandlerFunc {
	return h.protectCapability("", next)
}

func (h Handler) protectCapability(capability authz.Capability, next func(http.ResponseWriter, *http.Request, authz.Principal)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := h.auth.Principal(r)
		if !ok || !principal.Has(authz.CapabilityControlPanelAccess) {
			http.Redirect(w, r, "/go-login?return_to="+r.URL.Path, http.StatusSeeOther)
			return
		}
		if capability != "" && !principal.Has(capability) {
			http.Error(w, "Forbidden.", http.StatusForbidden)
			return
		}
		next(w, r, principal)
	}
}

func (h Handler) sessionForLogin(email string, password string) (rest.SessionData, bool) {
	return fixtureSession(email, password, h.loginPolicy, h.playgroundAuth)
}

func (h Handler) dashboard(w http.ResponseWriter, r *http.Request, principal authz.Principal) {
	fixture := h.fixture(r)
	postCount := h.contentCount(r.Context(), domaincontent.KindPost)
	pageCount := h.contentCount(r.Context(), domaincontent.KindPage)
	taxonomies, _ := h.services.Taxonomy.ListDefinitions(r.Context())
	media, _ := h.services.Media.List(r.Context())
	data := views.DashboardData{
		Layout:      h.layout(r, principal, fixture.Dashboard.Title, "/go-admin"),
		Title:       fixture.Dashboard.Title,
		Description: fixture.Dashboard.Description,
		Cards: h.dashboardCards(fixture, []dashboardCardValue{
			{Key: "Posts", Value: strconv.Itoa(postCount), Href: "/go-admin/posts"},
			{Key: "Pages", Value: strconv.Itoa(pageCount), Href: "/go-admin/pages"},
			{Key: "Taxonomies", Value: strconv.Itoa(len(taxonomies)), Href: "/go-admin/taxonomies"},
			{Key: "Media", Value: strconv.Itoa(len(media)), Href: "/go-admin/media"},
		}),
	}
	_ = web.Render(r.Context(), w, views.DashboardPage(data))
}

func (h Handler) contentList(kind domaincontent.Kind, screen string) func(http.ResponseWriter, *http.Request, authz.Principal) {
	return func(w http.ResponseWriter, r *http.Request, principal authz.Principal) {
		resource, _ := cmspanel.ResourceByID(screen)
		result, _ := h.services.Content.List(r.Context(), domaincontent.Query{Kinds: []domaincontent.Kind{kind}, Page: 1, PerPage: 50, SortBy: domaincontent.SortUpdatedAt, SortDesc: true})
		fixture := h.fixture(r)
		screenFixture, _ := fixture.Screen(screen)
		screenTitle := fallbackValue(screenFixture.Title, fallbackValue(resource.Label, titleFor(screen)))
		screenDescription := fallbackValue(screenFixture.Description, fallbackValue(resource.Description, "Create, edit, publish, schedule, trash, and restore content."))
		basePath := fallbackValue(resource.BasePath, "/go-admin/"+screen)
		rows := make([]blocks.ContentRow, 0, len(result.Items))
		for _, entry := range result.Items {
			rows = append(rows, blocks.ContentRow{
				ID:      string(entry.ID),
				Title:   entry.Title.Value("en", "en"),
				Slug:    entry.Slug.Value("en", "en"),
				Status:  string(entry.Status),
				Author:  entry.AuthorID,
				EditURL: basePath + "/" + string(entry.ID) + "/edit",
			})
		}
		data := views.ContentListPageData{
			Layout: h.layout(r, principal, screenTitle, basePath),
			Screen: screen,
			Table: blocks.ContentTableData{
				Title:       screenTitle,
				Description: screenDescription,
				Rows:        rows,
				Actions:     h.panelActions(fixture, principal, resource.Actions),
				Pagination: elements.PaginationData{
					Page: 1, TotalPages: 1, BaseHref: basePath,
					PreviousLabel: fixture.Label("action_previous", "Previous"),
					NextLabel:     fixture.Label("action_next", "Next"),
				},
				Headers:   h.contentTableHeadersFromSchema(fixture, resource.Table),
				EditLabel: fixture.Label("action_edit", "Edit"),
			},
		}
		_ = web.Render(r.Context(), w, views.ContentListPage(data))
	}
}

func (h Handler) contentNew(kind domaincontent.Kind, screen string) func(http.ResponseWriter, *http.Request, authz.Principal) {
	return func(w http.ResponseWriter, r *http.Request, principal authz.Principal) {
		if !principal.Has(authz.CapabilityContentCreate) {
			http.Error(w, "Forbidden.", http.StatusForbidden)
			return
		}
		fixture := h.fixture(r)
		screenFixture, _ := fixture.Screen(screen)
		resource, _ := cmspanel.ResourceByID(screen)
		singularName := fallbackValue(screenFixture.Singular, fallbackValue(resource.Singular, singular(screen)))
		description := fallbackValue(screenFixture.FormDescription, h.labelFromFixture(fixture, "action_content_create", "Create a draft and choose publish state."))
		basePath := fallbackValue(resource.BasePath, "/go-admin/"+screen)
		data := views.ContentEditPageData{
			Layout: h.layout(r, principal, h.labelFromFixture(fixture, "action_new", "New")+" "+singularName, basePath),
			Screen: screen + "-edit",
			Editor: h.contentEditor(r.Context(), fixture, "New "+singularName, description, basePath, h.token("content-write"), domaincontent.Entry{Kind: kind, Status: domaincontent.StatusDraft}),
		}
		_ = web.Render(r.Context(), w, views.ContentEditPage(data))
	}
}

func (h Handler) contentEdit(screen string) func(http.ResponseWriter, *http.Request, authz.Principal) {
	return func(w http.ResponseWriter, r *http.Request, principal authz.Principal) {
		entry, err := h.services.Content.Get(r.Context(), principal, domaincontent.ID(r.PathValue("id")))
		if err != nil {
			http.NotFound(w, r)
			return
		}
		fixture := h.fixture(r)
		screenFixture, _ := fixture.Screen(screen)
		description := fallbackValue(screenFixture.FormDescription, h.labelFromFixture(fixture, "action_content_update", "Update content fields and publication state."))
		resource, _ := cmspanel.ResourceByID(screen)
		basePath := fallbackValue(resource.BasePath, "/go-admin/"+screen)
		data := views.ContentEditPageData{
			Layout: h.layout(r, principal, h.labelFromFixture(fixture, "action_edit", "Edit")+" "+entry.Title.Value("en", "en"), basePath),
			Screen: screen + "-edit",
			Editor: h.contentEditor(r.Context(), fixture, h.labelFromFixture(fixture, "action_edit", "Edit")+" "+entry.Title.Value("en", "en"), description, basePath+"/"+string(entry.ID), h.token("content-write"), entry),
		}
		_ = web.Render(r.Context(), w, views.ContentEditPage(data))
	}
}

func (h Handler) contentCreate(kind domaincontent.Kind) func(http.ResponseWriter, *http.Request, authz.Principal) {
	return func(w http.ResponseWriter, r *http.Request, principal authz.Principal) {
		if !principal.Has(authz.CapabilityContentCreate) {
			http.Error(w, "Forbidden.", http.StatusForbidden)
			return
		}
		if err := r.ParseForm(); err != nil || !h.validToken(r.PostForm.Get("action_token"), "content-write") {
			http.Error(w, "Invalid content submission.", http.StatusBadRequest)
			return
		}
		entry, err := h.services.Content.CreateDraft(r.Context(), principal, appcontent.CreateDraftCommand{
			Kind: kind, Title: localized(r.PostForm.Get("title")), Slug: localized(r.PostForm.Get("slug")),
			Body: localized(r.PostForm.Get("content")), Excerpt: localized(r.PostForm.Get("excerpt")), AuthorID: r.PostForm.Get("author_id"),
			FeaturedMediaID: r.PostForm.Get("featured_media_id"), Template: r.PostForm.Get("template"), Metadata: formMetadata(r), Terms: formTerms(r),
		})
		if err == nil {
			_, err = h.applyStatus(r, principal, entry.ID)
		}
		if err != nil {
			http.Error(w, err.Error(), statusFromError(err))
			return
		}
		http.Redirect(w, r, contentListPath(kind), http.StatusSeeOther)
	}
}

func (h Handler) contentUpdate(w http.ResponseWriter, r *http.Request, principal authz.Principal) {
	if !principal.Has(authz.CapabilityContentEdit) && !principal.Has(authz.CapabilityContentEditOwn) {
		http.Error(w, "Forbidden.", http.StatusForbidden)
		return
	}
	if err := r.ParseForm(); err != nil || !h.validToken(r.PostForm.Get("action_token"), "content-write") {
		http.Error(w, "Invalid content submission.", http.StatusBadRequest)
		return
	}
	entry, err := h.services.Content.Update(r.Context(), principal, appcontent.UpdateCommand{
		ID: domaincontent.ID(r.PathValue("id")), Title: localized(r.PostForm.Get("title")), Slug: localized(r.PostForm.Get("slug")),
		Body: localized(r.PostForm.Get("content")), Excerpt: localized(r.PostForm.Get("excerpt")), AuthorID: r.PostForm.Get("author_id"),
		FeaturedMediaID: r.PostForm.Get("featured_media_id"), Template: r.PostForm.Get("template"), Metadata: formMetadata(r), Terms: formTerms(r),
	})
	if err == nil {
		entry, err = h.applyStatus(r, principal, entry.ID)
	}
	if err != nil {
		http.Error(w, err.Error(), statusFromError(err))
		return
	}
	http.Redirect(w, r, contentListPath(entry.Kind), http.StatusSeeOther)
}

func (h Handler) contentTrash(w http.ResponseWriter, r *http.Request, principal authz.Principal) {
	if !principal.Has(authz.CapabilityContentDelete) {
		http.Error(w, "Forbidden.", http.StatusForbidden)
		return
	}
	if err := r.ParseForm(); err != nil || !h.validToken(r.PostForm.Get("action_token"), "content-trash") {
		http.Error(w, "Invalid action token.", http.StatusForbidden)
		return
	}
	entry, err := h.services.Content.Trash(r.Context(), principal, domaincontent.ID(r.PathValue("id")))
	if err != nil {
		http.Error(w, err.Error(), statusFromError(err))
		return
	}
	http.Redirect(w, r, contentListPath(entry.Kind), http.StatusSeeOther)
}

func (h Handler) contentTypesPage(w http.ResponseWriter, r *http.Request, principal authz.Principal) {
	fixture := h.fixture(r)
	items, _ := h.services.ContentTypes.List(r.Context())
	rows := make([]blocks.SimpleListRow, 0, len(items))
	for _, item := range items {
		rows = append(rows, blocks.SimpleListRow{Label: string(item.ID), Description: item.Label, Status: visibleStatus(item.RESTVisible), ActionURL: ""})
	}
	h.renderSimple(w, r, principal, "content-types", fixture, rows, "content-types", h.simpleFields(r, "content-types"), "/go-admin/content-types")
}

func (h Handler) contentTypeCreate(w http.ResponseWriter, r *http.Request, principal authz.Principal) {
	if !principal.Has(authz.CapabilitySettingsManage) {
		http.Error(w, "Forbidden.", http.StatusForbidden)
		return
	}
	_ = r.ParseForm()
	if !h.validToken(r.PostForm.Get("action_token"), "admin-write") {
		http.Error(w, "Invalid action token.", http.StatusForbidden)
		return
	}
	contentType := domaincontenttype.Type{ID: domaincontent.Kind(r.PostForm.Get("id")), Label: r.PostForm.Get("label"), Public: true, RESTVisible: true, GraphQLVisible: true, Supports: domaincontenttype.Supports{Title: true, Editor: true, Excerpt: true, FeaturedMedia: true, Revisions: true, Taxonomies: true, CustomFields: true}, Archive: true}
	if err := h.services.ContentTypes.Register(r.Context(), contentType); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, "/go-admin/content-types", http.StatusSeeOther)
}

func (h Handler) taxonomiesPage(w http.ResponseWriter, r *http.Request, principal authz.Principal) {
	fixture := h.fixture(r)
	items, _ := h.services.Taxonomy.ListDefinitions(r.Context())
	rows := make([]blocks.SimpleListRow, 0, len(items))
	for _, item := range items {
		rows = append(rows, blocks.SimpleListRow{Label: string(item.Type), Description: item.Label, Status: string(item.Mode), ActionURL: "/go-admin/taxonomies/" + string(item.Type) + "/terms"})
	}
	h.renderSimple(w, r, principal, "taxonomies", fixture, rows, "taxonomies", h.simpleFields(r, "taxonomies"), "/go-admin/taxonomies")
}

func (h Handler) taxonomyCreate(w http.ResponseWriter, r *http.Request, principal authz.Principal) {
	if !principal.Has(authz.CapabilityTaxonomiesManage) {
		http.Error(w, "Forbidden.", http.StatusForbidden)
		return
	}
	_ = r.ParseForm()
	if !h.validToken(r.PostForm.Get("action_token"), "admin-write") {
		http.Error(w, "Invalid action token.", http.StatusForbidden)
		return
	}
	definition := domaintaxonomy.Definition{Type: domaintaxonomy.Type(r.PostForm.Get("type")), Label: r.PostForm.Get("label"), Mode: domaintaxonomy.Mode(defaultValue(r.PostForm.Get("mode"), "flat")), AssignedToKinds: []domaincontent.Kind{domaincontent.KindPost}, Public: true, RESTVisible: true, GraphQLVisible: true}
	if err := h.services.Taxonomy.Register(r.Context(), principal, definition); err != nil {
		http.Error(w, err.Error(), statusFromError(err))
		return
	}
	http.Redirect(w, r, "/go-admin/taxonomies", http.StatusSeeOther)
}

func (h Handler) termsPage(w http.ResponseWriter, r *http.Request, principal authz.Principal) {
	fixture := h.fixture(r)
	taxonomyType := domaintaxonomy.Type(r.PathValue("type"))
	items, _ := h.services.Taxonomy.ListTerms(r.Context(), taxonomyType)
	rows := make([]blocks.SimpleListRow, 0, len(items))
	for _, item := range items {
		rows = append(rows, blocks.SimpleListRow{Label: string(item.ID), Description: item.Name.Value("en", "en"), Status: string(item.Type)})
	}
	h.renderSimple(w, r, principal, "terms", fixture, rows, "terms", h.simpleFields(r, "terms"), "/go-admin/taxonomies/"+string(taxonomyType)+"/terms")
}

func (h Handler) termCreate(w http.ResponseWriter, r *http.Request, principal authz.Principal) {
	if !principal.Has(authz.CapabilityTaxonomiesManage) {
		http.Error(w, "Forbidden.", http.StatusForbidden)
		return
	}
	_ = r.ParseForm()
	if !h.validToken(r.PostForm.Get("action_token"), "admin-write") {
		http.Error(w, "Invalid action token.", http.StatusForbidden)
		return
	}
	term := domaintaxonomy.Term{ID: domaintaxonomy.TermID(r.PostForm.Get("id")), Type: domaintaxonomy.Type(r.PathValue("type")), Name: localized(r.PostForm.Get("name")), Slug: localized(r.PostForm.Get("slug"))}
	if err := h.services.Taxonomy.CreateTerm(r.Context(), principal, term); err != nil {
		http.Error(w, err.Error(), statusFromError(err))
		return
	}
	http.Redirect(w, r, "/go-admin/taxonomies/"+r.PathValue("type")+"/terms", http.StatusSeeOther)
}

func (h Handler) mediaPage(w http.ResponseWriter, r *http.Request, principal authz.Principal) {
	fixture := h.fixture(r)
	items, _ := h.services.Media.List(r.Context())
	rows := make([]blocks.SimpleListRow, 0, len(items))
	for _, item := range items {
		rows = append(rows, blocks.SimpleListRow{Label: string(item.ID), Description: item.Filename, Status: item.MimeType})
	}
	h.renderSimple(w, r, principal, "media", fixture, rows, "media", h.simpleFields(r, "media"), "/go-admin/media")
}

func (h Handler) mediaSave(w http.ResponseWriter, r *http.Request, principal authz.Principal) {
	if !principal.Has(authz.CapabilityMediaUpload) && !principal.Has(authz.CapabilityMediaEdit) {
		http.Error(w, "Forbidden.", http.StatusForbidden)
		return
	}
	_ = r.ParseForm()
	if !h.validToken(r.PostForm.Get("action_token"), "admin-write") {
		http.Error(w, "Invalid action token.", http.StatusForbidden)
		return
	}
	asset := domainmedia.Asset{ID: domainmedia.ID(r.PostForm.Get("id")), Filename: r.PostForm.Get("filename"), MimeType: r.PostForm.Get("mime_type"), PublicURL: r.PostForm.Get("public_url"), CreatedAt: time.Now(), UpdatedAt: time.Now()}
	if err := h.services.Media.SaveMetadata(r.Context(), principal, asset); err != nil {
		http.Error(w, err.Error(), statusFromError(err))
		return
	}
	http.Redirect(w, r, "/go-admin/media", http.StatusSeeOther)
}

func (h Handler) menusPage(w http.ResponseWriter, r *http.Request, principal authz.Principal) {
	fixture := h.fixture(r)
	items, _ := h.services.Menus.List(r.Context())
	rows := make([]blocks.SimpleListRow, 0, len(items))
	for _, item := range items {
		rows = append(rows, blocks.SimpleListRow{Label: string(item.ID), Description: item.Name, Status: string(item.Location)})
	}
	h.renderSimple(w, r, principal, "menus", fixture, rows, "menus", h.simpleFields(r, "menus"), "/go-admin/menus")
}

func (h Handler) menuSave(w http.ResponseWriter, r *http.Request, principal authz.Principal) {
	if !principal.Has(authz.CapabilityMenusManage) {
		http.Error(w, "Forbidden.", http.StatusForbidden)
		return
	}
	_ = r.ParseForm()
	if !h.validToken(r.PostForm.Get("action_token"), "admin-write") {
		http.Error(w, "Invalid action token.", http.StatusForbidden)
		return
	}
	menu := domainmenus.Menu{ID: domainmenus.ID(r.PostForm.Get("id")), Name: r.PostForm.Get("name"), Location: domainmenus.Location(r.PostForm.Get("location"))}
	if err := h.services.Menus.Save(r.Context(), principal, menu); err != nil {
		http.Error(w, err.Error(), statusFromError(err))
		return
	}
	http.Redirect(w, r, "/go-admin/menus", http.StatusSeeOther)
}

func (h Handler) usersPage(w http.ResponseWriter, r *http.Request, principal authz.Principal) {
	fixture := h.fixture(r)
	items, _ := h.services.Users.List(r.Context())
	rows := make([]blocks.SimpleListRow, 0, len(items))
	for _, item := range items {
		rows = append(rows, blocks.SimpleListRow{Label: string(item.ID), Description: item.DisplayName, Status: string(item.Status)})
	}
	h.renderSimple(w, r, principal, "users", fixture, rows, "users", h.simpleFields(r, "users"), "/go-admin/users")
}

func (h Handler) userSave(w http.ResponseWriter, r *http.Request, principal authz.Principal) {
	if !principal.Has(authz.CapabilityUsersManage) {
		http.Error(w, "Forbidden.", http.StatusForbidden)
		return
	}
	_ = r.ParseForm()
	if !h.validToken(r.PostForm.Get("action_token"), "admin-write") {
		http.Error(w, "Invalid action token.", http.StatusForbidden)
		return
	}
	user := domainusers.User{ID: domainusers.ID(r.PostForm.Get("id")), Login: r.PostForm.Get("login"), DisplayName: r.PostForm.Get("display_name"), Email: r.PostForm.Get("email"), Status: domainusers.StatusActive}
	if err := h.services.Users.Save(r.Context(), user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, "/go-admin/users", http.StatusSeeOther)
}

func (h Handler) authorsPage(w http.ResponseWriter, r *http.Request, principal authz.Principal) {
	fixture := h.fixture(r)
	items, _ := h.services.Users.List(r.Context())
	rows := make([]blocks.SimpleListRow, 0, len(items))
	for _, item := range items {
		author := item.PublicAuthor()
		rows = append(rows, blocks.SimpleListRow{Label: string(author.ID), Description: author.DisplayName, Status: author.Slug})
	}
	h.renderSimple(w, r, principal, "authors", fixture, rows, "authors", nil, "")
}

func (h Handler) capabilitiesPage(w http.ResponseWriter, r *http.Request, principal authz.Principal) {
	fixture := h.fixture(r)
	rows := []blocks.SimpleListRow{
		{Label: fixture.Label("capability_content", "Content"), Description: fixture.Label("capability_content_description", "Create, edit, publish, schedule, trash, restore"), Status: "core"},
		{Label: fixture.Label("capability_taxonomies", "Taxonomies"), Description: fixture.Label("capability_taxonomies_description", "Manage and assign terms"), Status: "core"},
		{Label: fixture.Label("capability_settings", "Settings"), Description: fixture.Label("capability_settings_description", "Manage private and public settings"), Status: "restricted"},
		{Label: fixture.Label("capability_users", "Users"), Description: fixture.Label("capability_users_description", "Manage accounts and roles"), Status: "restricted"},
	}
	h.renderSimple(w, r, principal, "capabilities", fixture, rows, "capabilities", nil, "")
}

func (h Handler) settingsPage(w http.ResponseWriter, r *http.Request, principal authz.Principal) {
	fixture := h.fixture(r)
	siteTitle := h.settingValue(r.Context(), "site.title", "GoCMS")
	publicRendering := h.settingValue(r.Context(), "public.rendering", h.defaultPublicRendering())
	form := h.contentEditorFromFixture(fixture, "settings", "/go-admin/settings", h.token("settings-write"), []blocks.FieldData{
		{ID: "site_title", Name: "site_title", Label: fixture.Label("field_site_title", "Site title"), Value: siteTitle, Required: true},
		{ID: "public_rendering", Name: "public_rendering", Label: fixture.Label("field_public_rendering", "Public rendering"), Value: publicRendering},
	})
	screen, _ := fixture.Screen("settings")
	title := fallbackValue(screen.Title, "Settings")
	description := fallbackValue(screen.Description, "Configure public site settings.")
	form.Actions = h.screenActions("settings", fixture)
	form.Title = title
	form.Description = description
	_ = web.Render(r.Context(), w, views.SettingsPage(views.SettingsPageData{
		Layout:      h.layout(r, principal, title, "/go-admin/settings"),
		Screen:      "settings",
		Form:        form,
		Title:       title,
		Description: description,
	}))
}

func (h Handler) settingsSave(w http.ResponseWriter, r *http.Request, principal authz.Principal) {
	if err := r.ParseForm(); err != nil || !h.validToken(r.PostForm.Get("action_token"), "settings-write") {
		http.Error(w, "Invalid settings submission.", http.StatusBadRequest)
		return
	}
	for _, value := range []domainsettings.Value{
		{Key: "site.title", Value: r.PostForm.Get("site_title"), Public: true},
		{Key: "public.rendering", Value: r.PostForm.Get("public_rendering"), Public: true},
	} {
		if err := h.services.Settings.Save(r.Context(), principal, value); err != nil {
			http.Error(w, err.Error(), statusFromError(err))
			return
		}
	}
	http.Redirect(w, r, "/go-admin/settings", http.StatusSeeOther)
}

func (h Handler) themesPage(w http.ResponseWriter, r *http.Request, principal authz.Principal) {
	fixture := h.fixture(r)
	activeID := h.settingValue(r.Context(), platformthemes.ActiveThemeKey, string(platformthemes.DefaultThemeID))
	activePreset := h.themeRegistry.ResolvePreset(activeID, h.settingValue(r.Context(), platformthemes.StylePresetKey, "default")).ID
	previewTheme := h.settingValue(r.Context(), platformthemes.PreviewThemeKey, activeID)
	previewPreset := h.settingValue(r.Context(), "theme.preview_preset", activePreset)
	items := h.themeRegistry.List()
	rows := make([]blocks.SimpleListRow, 0, len(items))
	for _, item := range items {
		status := "inactive"
		if string(item.ID) == activeID {
			status = "active"
		}
		presets := h.themeRegistry.ListPresets(string(item.ID))
		previewURL := "/?preview_theme=" + string(item.ID)
		if len(presets) > 0 {
			previewURL += "&preview_preset=" + presets[0].ID
		}
		rows = append(rows, blocks.SimpleListRow{
			Label:       item.Name,
			Description: item.Version + " | contract " + item.Contract + " | roles " + strings.Join(themeRoles(item), ", ") + " | presets " + strings.Join(themePresetIDs(presets), ", ") + " | preview " + previewURL,
			Status:      status,
		})
	}
	form := h.contentEditorFromFixture(fixture, "themes", "/go-admin/themes", h.token("admin-write"), []blocks.FieldData{
		{ID: "theme_active", Name: "theme_active", Label: fixture.Label("field_theme_active", "Active theme"), Value: activeID, Required: true},
		{ID: "theme_style_preset", Name: "theme_style_preset", Label: fixture.Label("field_theme_style_preset", "Style preset"), Value: activePreset, Required: true},
		{ID: "theme_preview", Name: "theme_preview", Label: fixture.Label("field_theme_preview", "Preview theme"), Value: previewTheme},
		{ID: "theme_preview_preset", Name: "theme_preview_preset", Label: fixture.Label("field_theme_preview_preset", "Preview preset"), Value: previewPreset},
	})
	updateFieldValue(form.Fields, "theme_active", activeID)
	updateFieldValue(form.Fields, "theme_style_preset", activePreset)
	updateFieldValue(form.Fields, "theme_preview", previewTheme)
	updateFieldValue(form.Fields, "theme_preview_preset", previewPreset)
	screen, _ := fixture.Screen("themes")
	title := fallbackValue(screen.Title, "Themes")
	description := fallbackValue(screen.Description, "Inspect installed themes, choose the active theme, and select a style preset.")
	form.Title = title
	form.Description = description
	h.renderSimple(w, r, principal, "themes", fixture, rows, "themes", form.Fields, "/go-admin/themes")
}

func (h Handler) themesSave(w http.ResponseWriter, r *http.Request, principal authz.Principal) {
	if err := r.ParseForm(); err != nil || !h.validToken(r.PostForm.Get("action_token"), "admin-write") {
		http.Error(w, "Invalid theme submission.", http.StatusBadRequest)
		return
	}
	activeTheme := h.themeRegistry.ResolveActive(r.PostForm.Get("theme_active")).Manifest().ID
	activePreset := h.themeRegistry.ResolvePreset(string(activeTheme), r.PostForm.Get("theme_style_preset")).ID
	previewTheme := h.themeRegistry.ResolveActive(r.PostForm.Get("theme_preview")).Manifest().ID
	previewPreset := h.themeRegistry.ResolvePreset(string(previewTheme), r.PostForm.Get("theme_preview_preset")).ID
	for _, value := range []domainsettings.Value{
		{Key: domainsettings.Key(platformthemes.ActiveThemeKey), Value: string(activeTheme), Public: false},
		{Key: domainsettings.Key(platformthemes.StylePresetKey), Value: activePreset, Public: false},
		{Key: domainsettings.Key(platformthemes.PreviewThemeKey), Value: string(previewTheme), Public: false},
		{Key: "theme.preview_preset", Value: previewPreset, Public: false},
	} {
		if err := h.services.Settings.Save(r.Context(), principal, value); err != nil {
			http.Error(w, err.Error(), statusFromError(err))
			return
		}
	}
	http.Redirect(w, r, "/go-admin/themes", http.StatusSeeOther)
}

func (h Handler) permalinksPage(w http.ResponseWriter, r *http.Request, principal authz.Principal) {
	fixture := h.fixture(r)
	postPattern := h.settingValue(r.Context(), "permalinks.post_pattern", "/%postname%/")
	pagePattern := h.settingValue(r.Context(), "permalinks.page_pattern", "/{slug}/")
	form := h.contentEditorFromFixture(fixture, "permalinks", "/go-admin/permalinks", h.token("permalinks-write"), []blocks.FieldData{
		{ID: "post_pattern", Name: "post_pattern", Label: fixture.Label("field_post_pattern", "Post permalink pattern"), Value: postPattern, Required: true, Placeholder: "/%postname%/"},
		{ID: "page_pattern", Name: "page_pattern", Label: fixture.Label("field_page_pattern", "Page permalink pattern"), Value: pagePattern, Required: true, Placeholder: "/{slug}/"},
	})
	updateFieldValue(form.Fields, "post_pattern", postPattern)
	updateFieldValue(form.Fields, "page_pattern", pagePattern)
	screen, _ := fixture.Screen("permalinks")
	title := fallbackValue(screen.Title, "Permalinks")
	description := fallbackValue(screen.Description, "Configure public routes for posts and pages.")
	form.Actions = h.screenActions("permalinks", fixture)
	form.Title = title
	form.Description = description
	_ = web.Render(r.Context(), w, views.SettingsPage(views.SettingsPageData{
		Layout:      h.layout(r, principal, title, "/go-admin/permalinks"),
		Screen:      "permalinks",
		Form:        form,
		Title:       title,
		Description: description,
	}))
}

func (h Handler) permalinksSave(w http.ResponseWriter, r *http.Request, principal authz.Principal) {
	if err := r.ParseForm(); err != nil || !h.validToken(r.PostForm.Get("action_token"), "permalinks-write") {
		http.Error(w, "Invalid permalink submission.", http.StatusBadRequest)
		return
	}
	for _, value := range []domainsettings.Value{
		{Key: "permalinks.post_pattern", Value: r.PostForm.Get("post_pattern"), Public: false},
		{Key: "permalinks.page_pattern", Value: r.PostForm.Get("page_pattern"), Public: false},
	} {
		if err := h.services.Settings.Save(r.Context(), principal, value); err != nil {
			http.Error(w, err.Error(), statusFromError(err))
			return
		}
	}
	http.Redirect(w, r, "/go-admin/permalinks", http.StatusSeeOther)
}

func (h Handler) headlessPage(w http.ResponseWriter, r *http.Request, principal authz.Principal) {
	fixture := h.fixture(r)
	graphqlStatus := "available"
	graphqlDescription := fixture.Label("headless_graphql_description", "Available through the graphql plugin when activated.")
	if hasActivePlugin(h.runtimeInfo.ActivePlugins, "graphql") {
		graphqlStatus = "enabled"
		graphqlDescription = fixture.Label("headless_graphql_enabled_description", "/go-graphql is enabled through the graphql plugin.")
	}
	rows := []blocks.SimpleListRow{
		{Label: fixture.Label("headless_rest", "REST"), Description: fixture.Label("headless_rest_description", "/go-json/go/v2/ is enabled"), Status: "enabled"},
		{Label: fixture.Label("headless_graphql", "GraphQL"), Description: graphqlDescription, Status: graphqlStatus},
		{Label: fixture.Label("headless_rendering", "Public rendering"), Description: fixture.Label("headless_rendering_description", "Can remain disabled for headless mode"), Status: "disabled"},
	}
	h.renderSimple(w, r, principal, "headless", fixture, rows, "headless-settings", nil, "")
}

func (h Handler) runtimePage(w http.ResponseWriter, r *http.Request, principal authz.Principal) {
	fixture := h.fixture(r)
	info := h.runtimeInfo
	switchRule := fallbackValue(info.ProviderSwitchRule, "Durable provider switches require export, migration, restart or redeploy, and import through the JSON/site-package handoff. They are not runtime toggles.")
	rows := []blocks.SimpleListRow{
		{Label: fixture.Label("runtime_preset", "Preset"), Description: valueOrUnset(info.Preset), Status: "resolved"},
		{Label: fixture.Label("runtime_profile", "Runtime profile"), Description: valueOrUnset(info.RuntimeProfile), Status: "resolved"},
		{Label: fixture.Label("runtime_storage", "Storage profile"), Description: valueOrUnset(info.StorageProfile), Status: "resolved"},
		{Label: fixture.Label("runtime_content_provider", "Content provider"), Description: valueOrUnset(info.ContentProvider), Status: "resolved"},
		{Label: fixture.Label("runtime_site_package", "Site package"), Description: valueOrUnset(info.SitePackage), Status: enabledStatus(info.SitePackage != "")},
		{Label: fixture.Label("runtime_active_plugins", "Active plugins"), Description: strings.Join(info.ActivePlugins, ", "), Status: visibleStatus(len(info.ActivePlugins) > 0)},
		{Label: fixture.Label("runtime_browser_stateless", "Browser stateless"), Description: boolDescription(info.BrowserStateless), Status: visibleStatus(info.BrowserStateless)},
		{Label: fixture.Label("runtime_playground_auth", "Playground auth"), Description: boolDescription(info.PlaygroundAuth), Status: visibleStatus(info.PlaygroundAuth)},
		{Label: fixture.Label("runtime_dev_bearer", "Dev bearer auth"), Description: boolDescription(info.EnableDevBearer), Status: visibleStatus(info.EnableDevBearer)},
		{Label: fixture.Label("runtime_login_policy", "Login policy"), Description: valueOrUnset(info.LoginPolicy), Status: "resolved"},
		{Label: fixture.Label("runtime_admin_policy", "Admin policy"), Description: valueOrUnset(info.AdminPolicy), Status: "resolved"},
		{Label: fixture.Label("runtime_provider_switch_rule", "Provider switch rule"), Description: switchRule, Status: "restart-required"},
	}
	h.renderSimple(w, r, principal, "runtime", fixture, rows, "runtime-status", nil, "")
}

func (h Handler) renderSimple(w http.ResponseWriter, r *http.Request, principal authz.Principal, screen string, bundle adminfixtures.AdminFixture, rows []blocks.SimpleListRow, marker string, fields []blocks.FieldData, formAction string) {
	screenData, _ := bundle.Screen(screen)
	title := fallbackValue(screenData.Title, titleFor(screen))
	description := fallbackValue(screenData.Description, "Manage admin content.")
	actions := []elements.Action{}
	if screen == "headless" || screen == "settings" || screen == "themes" {
		actions = h.screenActions(screen, bundle)
	}
	data := views.SimpleListPageData{
		Layout: h.layout(r, principal, title, "/go-admin/"+screen),
		Screen: screen,
		List: blocks.SimpleListData{
			Title:       title,
			Description: description,
			Marker:      marker,
			Rows:        rows,
			Actions:     actions,
			FormAction:  formAction, Token: h.token("admin-write"), Fields: fields,
			Headers:   h.simpleListHeaders(bundle),
			OpenLabel: bundle.Label("action_open", "Open"),
			SaveLabel: bundle.Label("action_save", "Save"),
		},
	}
	_ = web.Render(r.Context(), w, views.SimpleListPage(data))
}

func (h Handler) layout(r *http.Request, principal authz.Principal, title string, active string) views.LayoutData {
	resolvedAssets := assets.Resolve()
	fixture := h.fixture(r)
	language := view.BuildLanguageToggleFromContext(r.Context(),
		view.WithLabel(fixture.Language.Label),
		view.WithCurrentLabel(fixture.Language.CurrentLabel),
		view.WithNextLocale(fixture.Language.NextLocale),
		view.WithNextLabel(fixture.Language.NextLabel),
		view.WithLocaleLabels(fixture.Language.LocaleLabels),
	)
	return views.LayoutData{
		Title:    title,
		Lang:     fixture.Meta.Lang,
		Brand:    fallbackValue(fixture.Meta.BrandName, "GoCMS"),
		Active:   active,
		NavItems: h.navigation(fixture, principal),
		Account: elements.AccountActionsData{
			Email:        principal.ID,
			Token:        h.token("logout"),
			SignOutLabel: fixture.Label("action_sign_out", "Sign out"),
		},
		Theme: view.ThemeToggleData{
			Label:              fixture.Theme.Label,
			SwitchToDarkLabel:  fixture.Theme.SwitchToDarkLabel,
			SwitchToLightLabel: fixture.Theme.SwitchToLightLabel,
		},
		Language: language,
		Assets: views.AssetPaths{
			CSS:              resolvedAssets.CSS,
			ThemeJS:          resolvedAssets.ThemeJS,
			AppJS:            resolvedAssets.AppJS,
			ExtraHeadScripts: h.extraAdminScripts(),
		},
	}
}

func (h Handler) contentEditor(ctx context.Context, bundle adminfixtures.AdminFixture, title string, description string, action string, token string, entry domaincontent.Entry) blocks.ContentEditorData {
	editorProvider := h.activeEditorProvider(ctx)
	resource, _ := cmspanel.ResourceByKind(entry.Kind)
	fields := h.formFieldsFromFixture(bundle, "content-editor", contentFormFields(resource.Form, entry, editorProvider))

	updateFieldValue(fields, "title", entry.Title.Value("en", "en"))
	updateFieldValue(fields, "slug", entry.Slug.Value("en", "en"))
	updateFieldValue(fields, "content", entry.Body.Value("en", "en"))
	updateFieldValue(fields, "excerpt", entry.Excerpt.Value("en", "en"))
	updateFieldValue(fields, "author_id", defaultValue(entry.AuthorID, "author-1"))
	updateFieldValue(fields, "featured_media_id", entry.FeaturedMediaID)
	updateFieldValue(fields, "template", entry.Template)
	updateFieldValue(fields, "terms", formatTerms(entry.Terms))
	return blocks.ContentEditorData{
		Title: title, Description: description, Action: action, Token: token, Status: string(defaultStatus(entry.Status)),
		Fields:        fields,
		Actions:       []elements.Action{{Label: h.labelFromFixture(bundle, "action_back", "Back"), Href: contentListPath(entry.Kind), Enabled: true}},
		PublishTitle:  bundle.Label("panel_publish", "Publish"),
		StatusLabel:   bundle.Label("field_status", "Status"),
		SaveLabel:     bundle.Label("action_save", "Save"),
		StatusOptions: h.statusOptions(bundle),
	}
}

func (h Handler) contentEditorFromFixture(bundle adminfixtures.AdminFixture, screen string, action string, token string, fallbackFields []blocks.FieldData) blocks.ContentEditorData {
	return blocks.ContentEditorData{
		Action:    action,
		Token:     token,
		Fields:    h.formFieldsFromFixture(bundle, screen, fallbackFields),
		SaveLabel: bundle.Label("action_save", "Save"),
	}
}

func (h Handler) applyStatus(r *http.Request, principal authz.Principal, id domaincontent.ID) (domaincontent.Entry, error) {
	switch domaincontent.Status(r.PostForm.Get("status")) {
	case domaincontent.StatusPublished:
		return h.services.Content.Publish(r.Context(), principal, id)
	case domaincontent.StatusScheduled:
		return h.services.Content.Schedule(r.Context(), principal, id, time.Now().Add(24*time.Hour))
	case domaincontent.StatusTrashed:
		return h.services.Content.Trash(r.Context(), principal, id)
	default:
		return h.services.Content.Get(r.Context(), principal, id)
	}
}

func (h Handler) contentCount(ctx context.Context, kind domaincontent.Kind) int {
	result, err := h.services.Content.List(ctx, domaincontent.Query{Kinds: []domaincontent.Kind{kind}, Page: 1, PerPage: 1})
	if err != nil {
		return 0
	}
	return result.Total
}

func (h Handler) dashboardCards(bundle adminfixtures.AdminFixture, values []dashboardCardValue) []blocks.StatCard {
	result := make([]blocks.StatCard, 0, len(values))
	for i, value := range values {
		label := value.Key
		if i < len(bundle.Dashboard.Cards) {
			label = fallbackValue(bundle.Dashboard.Cards[i].Label, value.Key)
		}
		result = append(result, blocks.StatCard{
			Label:       label,
			Value:       value.Value,
			Href:        value.Href,
			ActionLabel: bundle.Label("action_open", "Open"),
		})
	}
	return result
}

func (h Handler) token(action string) string {
	token, err := frameworkauth.SignedEncode(actionToken{Action: action, Exp: time.Now().Add(2 * time.Hour).Unix()}, h.secret)
	if err != nil {
		return ""
	}
	return token
}

func (h Handler) validToken(raw string, action string) bool {
	var token actionToken
	if err := frameworkauth.SignedDecode(raw, h.secret, &token); err != nil {
		return false
	}
	return token.Action == action && time.Now().Unix() <= token.Exp
}

func fixtureSession(email string, password string, loginPolicy string, playgroundAuth bool) (rest.SessionData, bool) {
	principals := rest.DevBearerPrincipals()
	switch strings.TrimSpace(strings.ToLower(loginPolicy)) {
	case "playground":
		if email == "admin" && password == "admin" {
			return sessionFromPrincipal(principals["admin-token"]), true
		}
		return rest.SessionData{}, false
	case "disabled":
		return rest.SessionData{}, false
	}
	if playgroundAuth {
		if email == "admin" && password == "admin" {
			return sessionFromPrincipal(principals["admin-token"]), true
		}
		return rest.SessionData{}, false
	}
	switch {
	case email == "admin@example.test" && password == "admin":
		return sessionFromPrincipal(principals["admin-token"]), true
	case email == "editor@example.test" && password == "editor":
		return sessionFromPrincipal(principals["editor-token"]), true
	case email == "viewer@example.test" && password == "viewer":
		return sessionFromPrincipal(principals["viewer-token"]), true
	default:
		return rest.SessionData{}, false
	}
}

func sessionFromPrincipal(principal authz.Principal) rest.SessionData {
	capabilities := make([]string, 0, len(principal.Capabilities))
	for capability := range principal.Capabilities {
		capabilities = append(capabilities, string(capability))
	}
	return rest.SessionData{UserID: principal.ID, Capabilities: capabilities}
}

func returnTo(r *http.Request) string {
	value := r.URL.Query().Get("return_to")
	if value == "" {
		return "/go-admin"
	}
	return safeReturnTo(value)
}

func safeReturnTo(value string) string {
	if strings.HasPrefix(value, "/go-admin") {
		return value
	}
	return "/go-admin"
}

func localized(value string) domaincontent.LocalizedText {
	return domaincontent.LocalizedText{"en": strings.TrimSpace(value)}
}

func formMetadata(r *http.Request) domaincontent.Metadata {
	key := strings.TrimSpace(r.PostForm.Get("meta_key"))
	if key == "" {
		return nil
	}
	return domaincontent.Metadata{
		key: domaincontent.MetaValue{Value: r.PostForm.Get("meta_value"), Public: true},
	}
}

func formTerms(r *http.Request) []domaincontent.TermRef {
	raw := strings.TrimSpace(r.PostForm.Get("terms"))
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	terms := make([]domaincontent.TermRef, 0, len(parts))
	for _, part := range parts {
		pair := strings.SplitN(strings.TrimSpace(part), ":", 2)
		if len(pair) != 2 || strings.TrimSpace(pair[0]) == "" || strings.TrimSpace(pair[1]) == "" {
			continue
		}
		terms = append(terms, domaincontent.TermRef{Taxonomy: strings.TrimSpace(pair[0]), TermID: strings.TrimSpace(pair[1])})
	}
	return terms
}

func formatTerms(terms []domaincontent.TermRef) string {
	values := make([]string, 0, len(terms))
	for _, term := range terms {
		values = append(values, term.Taxonomy+":"+term.TermID)
	}
	return strings.Join(values, ",")
}

func defaultStatus(status domaincontent.Status) domaincontent.Status {
	if status == "" {
		return domaincontent.StatusDraft
	}
	return status
}

func defaultValue(value string, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func hasActivePlugin(plugins []string, id string) bool {
	for _, pluginID := range plugins {
		if pluginID == id {
			return true
		}
	}
	return false
}

func contentListPath(kind domaincontent.Kind) string {
	if kind == domaincontent.KindPage {
		return "/go-admin/pages"
	}
	return "/go-admin/posts"
}

func titleFor(screen string) string {
	return strings.Title(screen)
}

func singular(screen string) string {
	return strings.TrimSuffix(screen, "s")
}

func visibleStatus(value bool) string {
	if value {
		return "enabled"
	}
	return "disabled"
}

func enabledStatus(value bool) string {
	if value {
		return "enabled"
	}
	return "disabled"
}

func boolDescription(value bool) string {
	if value {
		return "enabled"
	}
	return "disabled"
}

func valueOrUnset(value string) string {
	if strings.TrimSpace(value) == "" {
		return "not configured"
	}
	return value
}

func (h Handler) settingValue(ctx context.Context, key string, fallback string) string {
	value, ok, err := h.services.Settings.Get(ctx, domainsettings.Key(key))
	if err != nil || !ok {
		return fallback
	}
	if raw, ok := value.Value.(string); ok && strings.TrimSpace(raw) != "" {
		return raw
	}
	return fallback
}

func (h Handler) defaultPublicRendering() string {
	if strings.EqualFold(h.runtimeInfo.RuntimeProfile, string(runtimeprofile.RuntimeProfileFull)) {
		return "enabled"
	}
	return "disabled"
}

func (h Handler) simpleFields(r *http.Request, screen string) []blocks.FieldData {
	bundle := h.fixture(r)
	return h.formFieldsFromFixture(bundle, screen, nil)
}

func (h Handler) formFieldsFromFixture(bundle adminfixtures.AdminFixture, key string, fallback []blocks.FieldData) []blocks.FieldData {
	form, ok := bundle.Form(key)
	if !ok || len(form.Fields) == 0 {
		return fallback
	}
	return mergeFixtureFields(h.formFields(form.Fields), fallback)
}

func (h Handler) formFields(fields []adminfixtures.FieldFixture) []blocks.FieldData {
	result := make([]blocks.FieldData, 0, len(fields))
	for _, field := range fields {
		result = append(result, blocks.FieldData{
			ID:          field.ID,
			Name:        field.Name,
			Label:       field.Label,
			Value:       field.Value,
			Type:        field.Type,
			Component:   field.Component,
			Placeholder: field.Placeholder,
			Required:    field.Required,
			Rows:        field.Rows,
		})
	}
	return result
}

func contentFormFields(schema panel.FormSchema, entry domaincontent.Entry, editorProvider plugins.EditorProviderRegistration) []blocks.FieldData {
	fields := make([]blocks.FieldData, 0, len(schema.Fields))
	for _, field := range schema.Fields {
		fields = append(fields, contentFieldData(field, entry, editorProvider))
	}
	return fields
}

func contentFieldData(field panel.Field, entry domaincontent.Entry, editorProvider plugins.EditorProviderRegistration) blocks.FieldData {
	result := blocks.FieldData{
		ID:          field.ID,
		Name:        field.ID,
		Label:       field.Label,
		Value:       contentFieldValue(field.ID, entry),
		Placeholder: field.Placeholder,
		Required:    field.Required,
		Hint:        field.Description,
	}
	switch field.Type {
	case panel.FieldRichText:
		result.Component = "richtext"
		result.Rows = 12
		result.Editor = &blocks.EditorData{ProviderID: editorProvider.ID}
	case panel.FieldTextarea:
		result.Component = "textarea"
		result.Rows = 3
	case panel.FieldNumber:
		result.Type = "number"
	case panel.FieldBoolean:
		result.Type = "checkbox"
	case panel.FieldHidden:
		result.Type = "hidden"
	case panel.FieldSelect:
		result.Component = "select"
		result.Options = panelOptions(field.Options)
	}
	return result
}

func contentFieldValue(id string, entry domaincontent.Entry) string {
	switch id {
	case "title":
		return entry.Title.Value("en", "en")
	case "slug":
		return entry.Slug.Value("en", "en")
	case "content":
		return entry.Body.Value("en", "en")
	case "excerpt":
		return entry.Excerpt.Value("en", "en")
	case "author_id":
		return defaultValue(entry.AuthorID, "author-1")
	case "featured_media_id":
		return entry.FeaturedMediaID
	case "template":
		return entry.Template
	case "terms":
		return formatTerms(entry.Terms)
	default:
		return ""
	}
}

func panelOptions(options []panel.Option) []ui.FieldOption {
	result := make([]ui.FieldOption, 0, len(options))
	for _, option := range options {
		result = append(result, ui.FieldOption{Value: option.Value, Label: option.Label})
	}
	return result
}

func mergeFixtureFields(fields []blocks.FieldData, fallback []blocks.FieldData) []blocks.FieldData {
	if len(fallback) == 0 {
		return fields
	}
	index := make(map[string]blocks.FieldData, len(fallback))
	for _, field := range fallback {
		if field.ID != "" {
			index[field.ID] = field
		}
		if field.Name != "" {
			index[field.Name] = field
		}
	}
	result := make([]blocks.FieldData, 0, len(fields))
	for _, field := range fields {
		override, ok := index[field.ID]
		if !ok && field.Name != "" {
			override, ok = index[field.Name]
		}
		if !ok {
			result = append(result, field)
			continue
		}
		if override.Name != "" {
			field.Name = override.Name
		}
		if override.Value != "" {
			field.Value = override.Value
		}
		if override.Type != "" {
			field.Type = override.Type
		}
		if override.Component != "" {
			field.Component = override.Component
		}
		if override.Placeholder != "" {
			field.Placeholder = override.Placeholder
		}
		field.Required = field.Required || override.Required
		if override.Rows > 0 {
			field.Rows = override.Rows
		}
		if len(override.Options) > 0 {
			field.Options = override.Options
		}
		if override.Hint != "" {
			field.Hint = override.Hint
		}
		if override.Editor != nil {
			field.Editor = override.Editor
		}
		result = append(result, field)
	}
	return result
}

func (h Handler) screenActions(screen string, fixture adminfixtures.AdminFixture) []elements.Action {
	return h.registry.ScreenActions(screen, fixture)
}

func (h Handler) extraAdminScripts() []string {
	assetsList := h.registry.AssetsForSurface(plugins.SurfaceAdmin)
	result := make([]string, 0, len(assetsList))
	for _, asset := range assetsList {
		result = append(result, assets.ResolvePath(asset.Path))
	}
	return result
}

func (h Handler) panelActions(bundle adminfixtures.AdminFixture, principal authz.Principal, actions []panel.Action[authz.Capability]) []elements.Action {
	result := make([]elements.Action, 0, len(actions))
	for _, action := range actions {
		enabled := action.Capability == "" || principal.Has(action.Capability)
		result = append(result, elements.Action{
			Label:   h.panelActionLabel(bundle, action),
			Href:    action.URL,
			Style:   panelActionStyle(action.Style),
			Enabled: enabled,
		})
	}
	return result
}

func (h Handler) panelActionLabel(bundle adminfixtures.AdminFixture, action panel.Action[authz.Capability]) string {
	switch action.ID {
	case "create":
		return h.labelFromFixture(bundle, "action_create", action.Label)
	case "edit":
		return h.labelFromFixture(bundle, "action_edit", action.Label)
	default:
		return action.Label
	}
}

func panelActionStyle(style panel.ActionStyle) string {
	switch style {
	case panel.ActionLink:
		return "link"
	case panel.ActionBadge:
		return "badge"
	default:
		return ""
	}
}

func (h Handler) contentTableHeadersFromSchema(bundle adminfixtures.AdminFixture, schema panel.TableSchema[authz.Capability]) blocks.ContentTableHeaders {
	return blocks.ContentTableHeaders{
		Title:   bundle.Label("table_title", panelColumnLabel(schema, "title", "Title")),
		Slug:    bundle.Label("table_slug", panelColumnLabel(schema, "slug", "Slug")),
		Status:  bundle.Label("table_status", panelColumnLabel(schema, "status", "Status")),
		Author:  bundle.Label("table_author", panelColumnLabel(schema, "author", "Author")),
		Actions: bundle.Label("table_actions", "Actions"),
	}
}

func panelColumnLabel(schema panel.TableSchema[authz.Capability], id string, fallback string) string {
	for _, column := range schema.Columns {
		if column.ID == id && strings.TrimSpace(column.Label) != "" {
			return column.Label
		}
	}
	return fallback
}

func (h Handler) contentTableHeaders(bundle adminfixtures.AdminFixture) blocks.ContentTableHeaders {
	return blocks.ContentTableHeaders{
		Title:   bundle.Label("table_title", "Title"),
		Slug:    bundle.Label("table_slug", "Slug"),
		Status:  bundle.Label("table_status", "Status"),
		Author:  bundle.Label("table_author", "Author"),
		Actions: bundle.Label("table_actions", "Actions"),
	}
}

func (h Handler) simpleListHeaders(bundle adminfixtures.AdminFixture) blocks.SimpleListHeaders {
	return blocks.SimpleListHeaders{
		Name:        bundle.Label("table_name", "Name"),
		Description: bundle.Label("table_description", "Description"),
		Status:      bundle.Label("table_status", "Status"),
		Actions:     bundle.Label("table_actions", "Actions"),
	}
}

func (h Handler) statusOptions(bundle adminfixtures.AdminFixture) []ui.FieldOption {
	return []ui.FieldOption{
		{Value: "draft", Label: bundle.Label("status_draft", "Draft")},
		{Value: "published", Label: bundle.Label("status_published", "Published")},
		{Value: "scheduled", Label: bundle.Label("status_scheduled", "Scheduled")},
		{Value: "trashed", Label: bundle.Label("status_trashed", "Trashed")},
	}
}

func (h Handler) labelFromFixture(bundle adminfixtures.AdminFixture, key string, fallback string) string {
	return bundle.Label(key, fallback)
}

func fallbackValue(value string, fallback string) string {
	if strings.TrimSpace(value) != "" {
		return value
	}
	return fallback
}

func updateFieldValue(fields []blocks.FieldData, id string, value string) {
	for i := range fields {
		if fields[i].ID == id {
			fields[i].Value = value
			return
		}
	}
}

func (h Handler) activeEditorProvider(ctx context.Context) plugins.EditorProviderRegistration {
	configured := h.settingValue(ctx, editorProviderSettingKey, defaultEditorProviderID)
	provider, ok := h.registry.ResolveEditorProvider(configured)
	if ok {
		return provider
	}
	return plugins.EditorProviderRegistration{
		ID:          defaultEditorProviderID,
		Label:       "TipTap (Basic)",
		Description: "Built-in TipTap editor with starter formatting extensions and HTML storage.",
		Priority:    0,
	}
}

func themeRoles(manifest domainthemes.Manifest) []string {
	roles := make([]string, 0, len(manifest.Templates))
	for role := range manifest.Templates {
		roles = append(roles, string(role))
	}
	slices.Sort(roles)
	return roles
}

func themePresetIDs(items []domainthemes.StylePreset) []string {
	ids := make([]string, 0, len(items))
	for _, item := range items {
		ids = append(ids, item.ID)
	}
	slices.Sort(ids)
	return ids
}

func statusFromError(err error) int {
	if err == nil {
		return http.StatusOK
	}
	if strings.Contains(strings.ToLower(err.Error()), "capability") {
		return http.StatusForbidden
	}
	return http.StatusBadRequest
}

func _format(_ string, args ...any) string {
	return fmt.Sprint(args...)
}
