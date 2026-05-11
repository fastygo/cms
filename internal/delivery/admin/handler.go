package admin

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strconv"
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
	apptaxonomy "github.com/fastygo/cms/internal/application/taxonomy"
	appusers "github.com/fastygo/cms/internal/application/users"
	"github.com/fastygo/cms/internal/delivery/admin/listui"
	"github.com/fastygo/cms/internal/delivery/rest"
	domainaudit "github.com/fastygo/cms/internal/domain/audit"
	domainauthn "github.com/fastygo/cms/internal/domain/authn"
	"github.com/fastygo/cms/internal/domain/authz"
	domaincontent "github.com/fastygo/cms/internal/domain/content"
	domaincontenttype "github.com/fastygo/cms/internal/domain/contenttype"
	domainmedia "github.com/fastygo/cms/internal/domain/media"
	domainmenus "github.com/fastygo/cms/internal/domain/menus"
	domainmeta "github.com/fastygo/cms/internal/domain/meta"
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
	Authn        appauthn.Service
	Audit        appaudit.Service
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
	metaRegistry   *appmeta.Registry
	settings       *appsettings.Registry
	authn          appauthn.Service
	health         apphealth.Service
	diagnostics    appdiagnostics.Service
}

type RuntimeInfo struct {
	Preset             string
	RuntimeProfile     string
	StorageProfile     string
	DeploymentProfile  string
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
	Settings       *appsettings.Registry
	ThemeRegistry  *platformthemes.Registry
	MetaRegistry   *appmeta.Registry
	Health         apphealth.Service
	Diagnostics    appdiagnostics.Service
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
		options.LoginPolicy = "local"
	}
	handler := Handler{
		services:       services,
		auth:           authenticator,
		secret:         secret,
		registry:       registry,
		playgroundAuth: options.PlaygroundAuth,
		loginPolicy:    options.LoginPolicy,
		runtimeInfo:    options.RuntimeInfo,
		settings:       options.Settings,
		themeRegistry:  options.ThemeRegistry,
		metaRegistry:   options.MetaRegistry,
		authn:          services.Authn,
		health:         options.Health,
		diagnostics:    options.Diagnostics,
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
	h.registry.AddRoutes(
		h.coreRoute("POST /go-admin/preferences/{screen}", authz.CapabilityControlPanelAccess, h.screenOptionsSave),
	)
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
	routes = append(routes,
		h.coreRoute("POST "+resource.BasePath+"/quick-edit", authz.CapabilityContentReadPrivate, h.contentQuickEdit(resource.Kind)),
		h.coreRoute("POST "+resource.BasePath+"/bulk", authz.CapabilityContentReadPrivate, h.contentBulk(resource.Kind, string(resource.ID))),
	)
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
	session, ok, err := h.sessionForLogin(r)
	if err != nil || !ok {
		h.audit(r, authz.Principal{ID: strings.TrimSpace(r.PostForm.Get("email"))}, "auth.login", "session", "", domainaudit.StatusFailure, map[string]any{"login_policy": h.loginPolicy})
		if err != nil {
			h.logError(r, "admin.login", err.Error(), "warning", map[string]any{"login_policy": h.loginPolicy})
		}
		message := fixture.Login.ErrorInvalidCredentials
		if errors.Is(err, appauthn.ErrLoginLocked) {
			message = "Too many login attempts. Try again later."
		}
		data := views.LoginPageData{
			Title:         fixture.Login.Title,
			Subtitle:      fixture.Login.Subtitle,
			Error:         message,
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
	if err := h.auth.Issue(w, session); err != nil {
		h.logError(r, "admin.login", err.Error(), "error", nil)
		http.Error(w, "Unable to issue session.", http.StatusInternalServerError)
		return
	}
	h.audit(r, authz.NewPrincipal(session.UserID), "auth.login", "session", session.UserID, domainaudit.StatusSuccess, map[string]any{"provider": session.ProviderID, "must_change_password": session.MustChangePassword})
	http.Redirect(w, r, adminReturnTo(r.PostForm.Get("return_to"), "/go-admin/content-types"), http.StatusSeeOther)
}

func (h Handler) logoutSubmit(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err == nil && h.validToken(r.PostForm.Get("action_token"), "logout") {
		if principal, ok := h.auth.Principal(r); ok {
			h.audit(r, principal, "auth.logout", "session", principal.ID, domainaudit.StatusSuccess, nil)
		}
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

func (h Handler) sessionForLogin(r *http.Request) (rest.SessionData, bool, error) {
	identifier := r.PostForm.Get("email")
	password := r.PostForm.Get("password")
	switch strings.TrimSpace(strings.ToLower(h.loginPolicy)) {
	case "playground":
		session, ok := fixtureSession(identifier, password, "playground", h.playgroundAuth)
		return session, ok, nil
	case "", "fixture":
		session, ok := fixtureSession(identifier, password, "fixture", h.playgroundAuth)
		return session, ok, nil
	case "local":
		if !h.authn.Enabled() {
			return rest.SessionData{}, false, nil
		}
		result, err := h.authn.AuthenticatePassword(r.Context(), appauthn.PasswordLoginInput{
			Identifier: identifier,
			Password:   password,
			RemoteAddr: r.RemoteAddr,
			UserAgent:  r.UserAgent(),
		}, h.sessionPolicy())
		if err != nil {
			return rest.SessionData{}, false, err
		}
		return sessionFromLoginResult(result, h.sessionPolicy()), true, nil
	case "disabled", "external", "oidc":
		return rest.SessionData{}, false, nil
	default:
		return rest.SessionData{}, false, nil
	}
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
		fixture := h.fixture(r)
		screenFixture, _ := fixture.Screen(screen)
		screenTitle := fallbackValue(screenFixture.Title, fallbackValue(resource.Label, titleFor(screen)))
		screenDescription := fallbackValue(screenFixture.Description, fallbackValue(resource.Description, "Create, edit, publish, schedule, trash, and restore content."))
		basePath := fallbackValue(resource.BasePath, "/go-admin/"+screen)
		state := listui.ParseState(
			r,
			screen,
			resource.Table,
			listui.FirstPositive(resource.Table.PerPage, 25),
			h.screenPerPagePreference(r.Context(), screen, listui.FirstPositive(resource.Table.PerPage, 25)),
			h.screenColumnsPreference(r.Context(), screen),
		)
		query := domaincontent.Query{
			Kinds:    []domaincontent.Kind{kind},
			Statuses: contentStatusesFilter(state.Filters["status"]),
			AuthorID: strings.TrimSpace(state.Filters["author"]),
			Search:   state.Search,
			Page:     state.Page,
			PerPage:  state.PerPage,
			SortBy:   contentSortField(state.Sort),
			SortDesc: state.Order == string(panel.SortDesc),
		}
		result, _ := h.services.Content.List(r.Context(), query)
		rows := make([]blocks.ContentRow, 0, len(result.Items))
		for _, entry := range result.Items {
			rows = append(rows, blocks.ContentRow{
				ID:           string(entry.ID),
				Title:        entry.Title.Value("en", "en"),
				Slug:         entry.Slug.Value("en", "en"),
				Status:       string(entry.Status),
				Author:       entry.AuthorID,
				EditURL:      basePath + "/" + string(entry.ID) + "/edit",
				QuickEditURL: listui.BuildEditHref(state, basePath, string(entry.ID)),
			})
		}
		quickEdit := h.contentQuickEditForm(r, principal, fixture, state, kind, basePath)
		data := views.ContentListPageData{
			Layout: h.layout(r, principal, screenTitle, basePath),
			Screen: screen,
			Table: blocks.ContentTableData{
				Title:       screenTitle,
				Description: screenDescription,
				Rows:        rows,
				Actions:     h.panelActions(fixture, principal, resource.Actions),
				Pagination: elements.PaginationData{
					Page: result.Page, TotalPages: result.TotalPages, BaseHref: state.BaseHref(basePath),
					PreviousLabel: fixture.Label("action_previous", "Previous"),
					NextLabel:     fixture.Label("action_next", "Next"),
				},
				Headers:    h.contentTableHeadersFromSchema(fixture, resource.Table),
				EditLabel:  fixture.Label("action_edit", "Edit"),
				QuickLabel: fixture.Label("action_quick_edit", "Quick edit"),
				Toolbar:    listui.BuildToolbarForm(fixture, resource.Table, basePath, state),
				QuickEdit:  quickEdit,
				ScreenOpts: listui.BuildScreenOptionsForm(fixture, screen, resource.Table, state, state.ReturnTo(basePath), h.token("screen-options-write")),
				Bulk:       h.contentBulkForm(fixture, principal, resource.Table, basePath, state.ReturnTo(basePath)),
				Visible:    state.VisibleMap(resource.Table),
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
			FeaturedMediaID: r.PostForm.Get("featured_media_id"), Template: r.PostForm.Get("template"), Metadata: h.formMetadata(r, kind), Terms: formTerms(r),
		})
		if err == nil {
			_, err = h.applyStatus(r, principal, entry.ID)
		}
		if err != nil {
			http.Error(w, err.Error(), statusFromError(err))
			return
		}
		h.audit(r, principal, "admin.content.create", "content", string(entry.ID), domainaudit.StatusSuccess, map[string]any{"kind": entry.Kind, "status": entry.Status})
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
	current, err := h.services.Content.Get(r.Context(), principal, domaincontent.ID(r.PathValue("id")))
	if err != nil {
		http.Error(w, err.Error(), statusFromError(err))
		return
	}
	entry, err := h.services.Content.Update(r.Context(), principal, appcontent.UpdateCommand{
		ID: domaincontent.ID(r.PathValue("id")), Title: localized(r.PostForm.Get("title")), Slug: localized(r.PostForm.Get("slug")),
		Body: localized(r.PostForm.Get("content")), Excerpt: localized(r.PostForm.Get("excerpt")), AuthorID: r.PostForm.Get("author_id"),
		FeaturedMediaID: r.PostForm.Get("featured_media_id"), Template: r.PostForm.Get("template"), Metadata: h.formMetadata(r, current.Kind), Terms: formTerms(r),
	})
	if err == nil {
		entry, err = h.applyStatus(r, principal, entry.ID)
	}
	if err != nil {
		http.Error(w, err.Error(), statusFromError(err))
		return
	}
	h.audit(r, principal, "admin.content.update", "content", string(entry.ID), domainaudit.StatusSuccess, map[string]any{"kind": entry.Kind, "status": entry.Status})
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
	h.audit(r, principal, "admin.content.trash", "content", string(entry.ID), domainaudit.StatusSuccess, map[string]any{"kind": entry.Kind})
	http.Redirect(w, r, contentListPath(entry.Kind), http.StatusSeeOther)
}

func (h Handler) contentQuickEdit(kind domaincontent.Kind) func(http.ResponseWriter, *http.Request, authz.Principal) {
	return func(w http.ResponseWriter, r *http.Request, principal authz.Principal) {
		if err := r.ParseForm(); err != nil || !h.validToken(r.PostForm.Get("action_token"), "content-quick-edit") {
			http.Error(w, "Invalid content submission.", http.StatusBadRequest)
			return
		}
		current, err := h.services.Content.Get(r.Context(), principal, domaincontent.ID(r.PostForm.Get("id")))
		if err != nil {
			http.Error(w, err.Error(), statusFromError(err))
			return
		}
		entry, err := h.services.Content.Update(r.Context(), principal, appcontent.UpdateCommand{
			ID:              current.ID,
			Title:           localized(r.PostForm.Get("title")),
			Slug:            localized(r.PostForm.Get("slug")),
			Body:            current.Body,
			Excerpt:         current.Excerpt,
			AuthorID:        current.AuthorID,
			FeaturedMediaID: current.FeaturedMediaID,
			Template:        current.Template,
			Metadata:        current.Metadata,
			Terms:           current.Terms,
		})
		if err == nil {
			entry, err = h.applyStatus(r, principal, entry.ID)
		}
		if err != nil {
			http.Error(w, err.Error(), statusFromError(err))
			return
		}
		if entry.Kind != kind {
			http.Error(w, "Invalid content kind.", http.StatusBadRequest)
			return
		}
		h.audit(r, principal, "admin.content.quick_edit", "content", string(entry.ID), domainaudit.StatusSuccess, map[string]any{"kind": entry.Kind, "status": entry.Status})
		http.Redirect(w, r, safeReturnTo(r.PostForm.Get("return_to")), http.StatusSeeOther)
	}
}

func (h Handler) contentBulk(kind domaincontent.Kind, screen string) func(http.ResponseWriter, *http.Request, authz.Principal) {
	return func(w http.ResponseWriter, r *http.Request, principal authz.Principal) {
		if err := r.ParseForm(); err != nil || !h.validToken(r.PostForm.Get("action_token"), "content-bulk") {
			http.Error(w, "Invalid bulk action.", http.StatusBadRequest)
			return
		}
		action := strings.TrimSpace(r.PostForm.Get("bulk_action"))
		ids := listui.SelectedIDs(r)
		if action == "" || len(ids) == 0 {
			http.Redirect(w, r, safeReturnTo(r.PostForm.Get("return_to")), http.StatusSeeOther)
			return
		}
		for _, id := range ids {
			entry, err := h.services.Content.Get(r.Context(), principal, domaincontent.ID(id))
			if err != nil {
				http.Error(w, err.Error(), statusFromError(err))
				return
			}
			if entry.Kind != kind {
				continue
			}
			switch action {
			case "publish":
				if _, err := h.services.Content.Publish(r.Context(), principal, entry.ID); err != nil {
					http.Error(w, err.Error(), statusFromError(err))
					return
				}
			case "trash":
				if _, err := h.services.Content.Trash(r.Context(), principal, entry.ID); err != nil {
					http.Error(w, err.Error(), statusFromError(err))
					return
				}
			case "restore":
				if _, err := h.services.Content.Restore(r.Context(), principal, entry.ID); err != nil {
					http.Error(w, err.Error(), statusFromError(err))
					return
				}
			default:
				http.Error(w, "Unsupported bulk action.", http.StatusBadRequest)
				return
			}
		}
		h.audit(r, principal, "admin.content.bulk", "content", screen, domainaudit.StatusSuccess, map[string]any{"action": action, "count": len(ids)})
		http.Redirect(w, r, safeReturnTo(r.PostForm.Get("return_to")), http.StatusSeeOther)
	}
}

func (h Handler) contentTypesPage(w http.ResponseWriter, r *http.Request, principal authz.Principal) {
	fixture := h.fixture(r)
	items, _ := h.services.ContentTypes.List(r.Context())
	rows := make([]blocks.SimpleListRow, 0, len(items))
	for _, item := range items {
		rows = append(rows, blocks.SimpleListRow{ID: string(item.ID), Label: string(item.ID), Description: item.Label, Status: visibleStatus(item.RESTVisible), ActionURL: ""})
	}
	h.renderSimple(w, r, principal, "content-types", fixture, rows, "content-types", h.simpleFields(r, "content-types"), "/go-admin/content-types", nil, nil)
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
	http.Redirect(w, r, adminReturnTo(r.PostForm.Get("return_to"), "/go-admin/taxonomies"), http.StatusSeeOther)
}

func (h Handler) taxonomiesPage(w http.ResponseWriter, r *http.Request, principal authz.Principal) {
	fixture := h.fixture(r)
	items, _ := h.services.Taxonomy.ListDefinitions(r.Context())
	schema := h.pageTableSchema("taxonomies")
	state := listui.ParseState(r, "taxonomies", schema, listui.FirstPositive(schema.PerPage, 25), h.screenPerPagePreference(r.Context(), "taxonomies", listui.FirstPositive(schema.PerPage, 25)), h.screenColumnsPreference(r.Context(), "taxonomies"))
	rows := make([]blocks.SimpleListRow, 0, len(items))
	for _, item := range items {
		rows = append(rows, blocks.SimpleListRow{
			ID:           string(item.Type),
			Label:        string(item.Type),
			Description:  item.Label,
			Status:       string(item.Mode),
			ActionURL:    "/go-admin/taxonomies/" + string(item.Type) + "/terms",
			QuickEditURL: listui.BuildEditHref(state, "/go-admin/taxonomies", string(item.Type)),
		})
	}
	h.renderSimple(w, r, principal, "taxonomies", fixture, rows, "taxonomies", h.simpleFields(r, "taxonomies"), "/go-admin/taxonomies", h.taxonomyQuickEditForm(fixture, state, items, "/go-admin/taxonomies"), nil)
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
	http.Redirect(w, r, adminReturnTo(r.PostForm.Get("return_to"), "/go-admin/taxonomies/"+r.PathValue("type")+"/terms"), http.StatusSeeOther)
}

func (h Handler) termsPage(w http.ResponseWriter, r *http.Request, principal authz.Principal) {
	fixture := h.fixture(r)
	taxonomyType := domaintaxonomy.Type(r.PathValue("type"))
	items, _ := h.services.Taxonomy.ListTerms(r.Context(), taxonomyType)
	basePath := "/go-admin/taxonomies/" + string(taxonomyType) + "/terms"
	schema := h.pageTableSchema("terms")
	state := listui.ParseState(r, "terms", schema, listui.FirstPositive(schema.PerPage, 25), h.screenPerPagePreference(r.Context(), "terms", listui.FirstPositive(schema.PerPage, 25)), h.screenColumnsPreference(r.Context(), "terms"))
	rows := make([]blocks.SimpleListRow, 0, len(items))
	for _, item := range items {
		rows = append(rows, blocks.SimpleListRow{
			ID:           string(item.ID),
			Label:        string(item.ID),
			Description:  item.Name.Value("en", "en"),
			Status:       string(item.Type),
			QuickEditURL: listui.BuildEditHref(state, basePath, string(item.ID)),
		})
	}
	h.renderSimple(w, r, principal, "terms", fixture, rows, "terms", h.simpleFields(r, "terms"), basePath, h.termQuickEditForm(fixture, state, items, basePath), nil)
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
	http.Redirect(w, r, adminReturnTo(r.PostForm.Get("return_to"), "/go-admin/media"), http.StatusSeeOther)
}

func (h Handler) mediaPage(w http.ResponseWriter, r *http.Request, principal authz.Principal) {
	fixture := h.fixture(r)
	items, _ := h.services.Media.List(r.Context())
	schema := h.pageTableSchema("media")
	state := listui.ParseState(r, "media", schema, listui.FirstPositive(schema.PerPage, 25), h.screenPerPagePreference(r.Context(), "media", listui.FirstPositive(schema.PerPage, 25)), h.screenColumnsPreference(r.Context(), "media"))
	rows := make([]blocks.SimpleListRow, 0, len(items))
	for _, item := range items {
		description := item.Filename
		if strings.TrimSpace(item.ProviderRef.Provider) != "" {
			description = item.Filename + " (" + item.ProviderRef.Provider + ")"
		}
		rows = append(rows, blocks.SimpleListRow{
			ID:           string(item.ID),
			Label:        string(item.ID),
			Description:  description,
			Status:       item.MimeType,
			QuickEditURL: listui.BuildEditHref(state, "/go-admin/media", string(item.ID)),
		})
	}
	h.renderSimple(w, r, principal, "media", fixture, rows, "media", h.mediaFields(r), "/go-admin/media", h.mediaQuickEditForm(fixture, state, items, "/go-admin/media"), nil)
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
	asset := domainmedia.Asset{
		ID:         domainmedia.ID(r.PostForm.Get("id")),
		Filename:   r.PostForm.Get("filename"),
		MimeType:   r.PostForm.Get("mime_type"),
		SizeBytes:  parseFormInt64(r, "size_bytes"),
		Width:      parseFormInt(r, "width"),
		Height:     parseFormInt(r, "height"),
		AltText:    r.PostForm.Get("alt_text"),
		Caption:    r.PostForm.Get("caption"),
		PublicURL:  r.PostForm.Get("public_url"),
		PublicMeta: map[string]any{},
		ProviderRef: domainmedia.BlobRef{
			Provider: r.PostForm.Get("provider"),
			Key:      r.PostForm.Get("provider_key"),
			URL:      r.PostForm.Get("provider_url"),
			Checksum: r.PostForm.Get("provider_checksum"),
			ETag:     r.PostForm.Get("provider_etag"),
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := h.services.Media.SaveMetadata(r.Context(), principal, asset); err != nil {
		http.Error(w, err.Error(), statusFromError(err))
		return
	}
	http.Redirect(w, r, adminReturnTo(r.PostForm.Get("return_to"), "/go-admin/menus"), http.StatusSeeOther)
}

func (h Handler) menusPage(w http.ResponseWriter, r *http.Request, principal authz.Principal) {
	fixture := h.fixture(r)
	items, _ := h.services.Menus.List(r.Context())
	schema := h.pageTableSchema("menus")
	state := listui.ParseState(r, "menus", schema, listui.FirstPositive(schema.PerPage, 25), h.screenPerPagePreference(r.Context(), "menus", listui.FirstPositive(schema.PerPage, 25)), h.screenColumnsPreference(r.Context(), "menus"))
	rows := make([]blocks.SimpleListRow, 0, len(items))
	for _, item := range items {
		rows = append(rows, blocks.SimpleListRow{
			ID:           string(item.ID),
			Label:        string(item.ID),
			Description:  item.Name,
			Status:       string(item.Location),
			QuickEditURL: listui.BuildEditHref(state, "/go-admin/menus", string(item.ID)),
		})
	}
	h.renderSimple(w, r, principal, "menus", fixture, rows, "menus", h.simpleFields(r, "menus"), "/go-admin/menus", h.menuQuickEditForm(fixture, state, items, "/go-admin/menus"), nil)
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
	http.Redirect(w, r, safeReturnTo(r.PostForm.Get("return_to")), http.StatusSeeOther)
}

func (h Handler) usersPage(w http.ResponseWriter, r *http.Request, principal authz.Principal) {
	fixture := h.fixture(r)
	items, _ := h.services.Users.List(r.Context())
	schema := h.pageTableSchema("users")
	state := listui.ParseState(r, "users", schema, listui.FirstPositive(schema.PerPage, 25), h.screenPerPagePreference(r.Context(), "users", listui.FirstPositive(schema.PerPage, 25)), h.screenColumnsPreference(r.Context(), "users"))
	rows := make([]blocks.SimpleListRow, 0, len(items))
	for _, item := range items {
		description := item.DisplayName
		if summary := h.userSecuritySummary(r.Context(), item); summary != "" {
			description = strings.TrimSpace(description + " | " + summary)
		}
		rows = append(rows, blocks.SimpleListRow{
			ID:           string(item.ID),
			Label:        string(item.ID),
			Description:  description,
			Status:       string(item.Status),
			QuickEditURL: listui.BuildEditHref(state, "/go-admin/users", string(item.ID)),
		})
	}
	h.renderSimple(w, r, principal, "users", fixture, rows, "users", h.simpleFields(r, "users"), "/go-admin/users", h.userQuickEditForm(fixture, state, items, "/go-admin/users"), h.userBulkForm(fixture, principal, state.ReturnTo("/go-admin/users")))
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
	if bulkAction := strings.TrimSpace(r.PostForm.Get("bulk_action")); bulkAction != "" {
		ids := listui.SelectedIDs(r)
		status := strings.TrimSpace(strings.TrimPrefix(bulkAction, "status:"))
		for _, id := range ids {
			current, ok, err := h.userByID(r.Context(), domainusers.ID(id))
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if !ok {
				continue
			}
			current.Status = domainusers.Status(status)
			if err := h.services.Users.Save(r.Context(), current); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		}
		http.Redirect(w, r, adminReturnTo(r.PostForm.Get("return_to"), "/go-admin/users"), http.StatusSeeOther)
		return
	}
	status := domainusers.Status(defaultValue(r.PostForm.Get("status"), string(domainusers.StatusActive)))
	securityAction := strings.TrimSpace(r.PostForm.Get("security_action"))
	if securityAction != "" && securityAction != "profile" {
		current, ok, err := h.userByID(r.Context(), domainusers.ID(r.PostForm.Get("id")))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if !ok {
			http.Error(w, "User not found.", http.StatusNotFound)
			return
		}
		switch securityAction {
		case "set_password":
			password := strings.TrimSpace(r.PostForm.Get("new_password"))
			if password == "" {
				http.Error(w, "New password is required.", http.StatusBadRequest)
				return
			}
			updated, err := h.authn.SetPassword(r.Context(), current.ID, password, r.PostForm.Get("must_change_password") != "")
			if err != nil {
				h.logError(r, "admin.user.password", err.Error(), "warning", map[string]any{"user_id": current.ID})
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			h.audit(r, principal, "auth.password.set", "user", string(updated.ID), domainaudit.StatusSuccess, nil)
			h.renderSecurityResult(w, r, principal, "Password updated", "The local password was rotated successfully.", []blocks.SimpleListRow{{Label: string(updated.ID), Description: "Password updated for local sign-in.", Status: visibleStatus(updated.MustChangePassword)}})
			return
		case "generate_recovery_codes":
			codes, err := h.authn.CreateRecoveryCodes(r.Context(), current.ID, max(parseFormInt(r, "recovery_code_count"), 1))
			if err != nil {
				h.logError(r, "admin.user.recovery_codes", err.Error(), "warning", map[string]any{"user_id": current.ID})
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			rows := make([]blocks.SimpleListRow, 0, len(codes))
			for i, code := range codes {
				rows = append(rows, blocks.SimpleListRow{ID: strconv.Itoa(i + 1), Label: "Recovery code", Description: code, Status: "copy-now"})
			}
			h.audit(r, principal, "auth.recovery_codes.generate", "user", string(current.ID), domainaudit.StatusSuccess, map[string]any{"count": len(codes)})
			h.renderSecurityResult(w, r, principal, "Recovery codes", "These codes are shown once. Copy them to offline storage now.", rows)
			return
		case "issue_reset_token":
			raw, token, err := h.authn.CreateResetToken(r.Context(), principal.ID, current.ID, r.PostForm.Get("reset_token_label"), time.Duration(max(parseFormInt(r, "reset_token_ttl_minutes"), 1))*time.Minute, false, false)
			if err != nil {
				h.logError(r, "admin.user.reset_token", err.Error(), "warning", map[string]any{"user_id": current.ID})
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			h.audit(r, principal, "auth.reset_token.issue", "user", string(current.ID), domainaudit.StatusSuccess, map[string]any{"expires_at": token.ExpiresAt.UTC().Format(time.RFC3339)})
			h.renderSecurityResult(w, r, principal, "Reset token", "This reset token is shown once. Deliver it over a trusted local/admin channel.", []blocks.SimpleListRow{
				{Label: "Token", Description: raw, Status: "copy-now"},
				{Label: "Expires", Description: token.ExpiresAt.UTC().Format(time.RFC3339), Status: "timed"},
			})
			return
		case "create_app_token":
			raw, token, err := h.authn.CreateAppToken(r.Context(), current.ID, r.PostForm.Get("app_token_name"), parseCapabilities(r.PostForm.Get("app_token_capabilities")), time.Duration(max(parseFormInt(r, "app_token_ttl_hours"), 1))*time.Hour)
			if err != nil {
				h.logError(r, "admin.user.app_token", err.Error(), "warning", map[string]any{"user_id": current.ID})
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			description := "Inherited user capabilities."
			if len(token.Capabilities) > 0 {
				parts := make([]string, 0, len(token.Capabilities))
				for _, capability := range token.Capabilities {
					parts = append(parts, string(capability))
				}
				description = strings.Join(parts, ", ")
			}
			h.audit(r, principal, "auth.app_token.create", "user", string(current.ID), domainaudit.StatusSuccess, map[string]any{"token_name": token.Name})
			h.renderSecurityResult(w, r, principal, "App token", "This app token is shown once. Copy it to the client now.", []blocks.SimpleListRow{
				{Label: "Token", Description: raw, Status: "copy-now"},
				{Label: "Scope", Description: description, Status: "granted"},
			})
			return
		case "revoke_app_token":
			if err := h.authn.RevokeAppToken(r.Context(), strings.TrimSpace(r.PostForm.Get("app_token_id"))); err != nil {
				h.logError(r, "admin.user.app_token_revoke", err.Error(), "warning", map[string]any{"user_id": current.ID})
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			h.audit(r, principal, "auth.app_token.revoke", "user", string(current.ID), domainaudit.StatusSuccess, nil)
			h.renderSecurityResult(w, r, principal, "App token revoked", "The selected app token was revoked.", []blocks.SimpleListRow{{Label: string(current.ID), Description: strings.TrimSpace(r.PostForm.Get("app_token_id")), Status: "revoked"}})
			return
		default:
			http.Error(w, "Unsupported security action.", http.StatusBadRequest)
			return
		}
	}
	user := domainusers.User{
		ID:          domainusers.ID(r.PostForm.Get("id")),
		Login:       r.PostForm.Get("login"),
		DisplayName: r.PostForm.Get("display_name"),
		Email:       r.PostForm.Get("email"),
		Status:      status,
	}
	if current, ok, _ := h.userByID(r.Context(), user.ID); ok {
		user.Roles = current.Roles
		user.Profile = current.Profile
		user.PasswordHash = current.PasswordHash
		user.PasswordUpdatedAt = current.PasswordUpdatedAt
		user.MustChangePassword = current.MustChangePassword
		user.LastLoginAt = current.LastLoginAt
	}
	if err := h.services.Users.Save(r.Context(), user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	h.audit(r, principal, "admin.user.save", "user", string(user.ID), domainaudit.StatusSuccess, map[string]any{"status": user.Status, "roles": user.Roles})
	http.Redirect(w, r, adminReturnTo(r.PostForm.Get("return_to"), "/go-admin/users"), http.StatusSeeOther)
}

func (h Handler) authorsPage(w http.ResponseWriter, r *http.Request, principal authz.Principal) {
	fixture := h.fixture(r)
	items, _ := h.services.Users.List(r.Context())
	rows := make([]blocks.SimpleListRow, 0, len(items))
	for _, item := range items {
		author := item.PublicAuthor()
		rows = append(rows, blocks.SimpleListRow{ID: string(author.ID), Label: string(author.ID), Description: author.DisplayName, Status: author.Slug})
	}
	h.renderSimple(w, r, principal, "authors", fixture, rows, "authors", nil, "", nil, nil)
}

func (h Handler) capabilitiesPage(w http.ResponseWriter, r *http.Request, principal authz.Principal) {
	fixture := h.fixture(r)
	rows := []blocks.SimpleListRow{
		{Label: fixture.Label("capability_content", "Content"), Description: fixture.Label("capability_content_description", "Create, edit, publish, schedule, trash, restore"), Status: "core"},
		{Label: fixture.Label("capability_taxonomies", "Taxonomies"), Description: fixture.Label("capability_taxonomies_description", "Manage and assign terms"), Status: "core"},
		{Label: fixture.Label("capability_settings", "Settings"), Description: fixture.Label("capability_settings_description", "Manage private and public settings"), Status: "restricted"},
		{Label: fixture.Label("capability_users", "Users"), Description: fixture.Label("capability_users_description", "Manage accounts and roles"), Status: "restricted"},
	}
	h.renderSimple(w, r, principal, "capabilities", fixture, rows, "capabilities", nil, "", nil, nil)
}

func (h Handler) settingsPage(w http.ResponseWriter, r *http.Request, principal authz.Principal) {
	fixture := h.fixture(r)
	definitions := h.settingDefinitions("site.title", "public.rendering")
	fields := h.settingFields(r.Context(), definitions)
	updateFieldID(fields, "site.title", "site_title", fixture.Label("field_site_title", "Site title"))
	updateFieldID(fields, "public.rendering", "public_rendering", fixture.Label("field_public_rendering", "Public rendering"))
	form := h.contentEditorFromFixture(fixture, "settings", "/go-admin/settings", h.token("settings-write"), fields)
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
		{Key: "site.title", Value: r.PostForm.Get("site_title")},
		{Key: "public.rendering", Value: r.PostForm.Get("public_rendering")},
	} {
		if err := h.services.Settings.Save(r.Context(), principal, value); err != nil {
			http.Error(w, err.Error(), statusFromError(err))
			return
		}
	}
	h.audit(r, principal, "admin.settings.save", "settings", "site", domainaudit.StatusSuccess, map[string]any{"keys": []string{"site.title", "public.rendering"}})
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
			ID:          string(item.ID),
			Label:       item.Name,
			Description: item.Version + " | contract " + item.Contract + " | roles " + strings.Join(themeRoles(item), ", ") + " | presets " + strings.Join(themePresetIDs(presets), ", ") + " | preview " + previewURL,
			Status:      status,
		})
	}
	themeOptions := h.themeOptionDefinitions(activeID)
	formFields := []blocks.FieldData{
		{ID: "theme_active", Name: "theme_active", Label: fixture.Label("field_theme_active", "Active theme"), Value: activeID, Required: true, Component: "select", Options: h.themeOptions()},
		{ID: "theme_style_preset", Name: "theme_style_preset", Label: fixture.Label("field_theme_style_preset", "Style preset"), Value: activePreset, Required: true, Component: "select", Options: h.themePresetOptions(activeID)},
		{ID: "theme_preview", Name: "theme_preview", Label: fixture.Label("field_theme_preview", "Preview theme"), Value: previewTheme, Component: "select", Options: h.themeOptions()},
		{ID: "theme_preview_preset", Name: "theme_preview_preset", Label: fixture.Label("field_theme_preview_preset", "Preview preset"), Value: previewPreset, Component: "select", Options: h.themePresetOptions(previewTheme)},
	}
	formFields = append(formFields, h.settingFields(r.Context(), themeOptions)...)
	form := h.contentEditorFromFixture(fixture, "themes", "/go-admin/themes", h.token("admin-write"), formFields)
	updateFieldValue(form.Fields, "theme_active", activeID)
	updateFieldValue(form.Fields, "theme_style_preset", activePreset)
	updateFieldValue(form.Fields, "theme_preview", previewTheme)
	updateFieldValue(form.Fields, "theme_preview_preset", previewPreset)
	screen, _ := fixture.Screen("themes")
	title := fallbackValue(screen.Title, "Themes")
	description := fallbackValue(screen.Description, "Inspect installed themes, choose the active theme, and select a style preset.")
	form.Title = title
	form.Description = description
	h.renderSimple(w, r, principal, "themes", fixture, rows, "themes", form.Fields, "/go-admin/themes", nil, nil)
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
		{Key: domainsettings.Key(platformthemes.ActiveThemeKey), Value: string(activeTheme)},
		{Key: domainsettings.Key(platformthemes.StylePresetKey), Value: activePreset},
		{Key: domainsettings.Key(platformthemes.PreviewThemeKey), Value: string(previewTheme)},
		{Key: "theme.preview_preset", Value: previewPreset},
	} {
		if err := h.services.Settings.Save(r.Context(), principal, value); err != nil {
			http.Error(w, err.Error(), statusFromError(err))
			return
		}
	}
	for _, definition := range h.themeOptionDefinitions(string(activeTheme)) {
		if err := h.services.Settings.Save(r.Context(), principal, domainsettings.Value{
			Key:   definition.Key,
			Value: r.PostForm.Get(string(definition.Key)),
		}); err != nil {
			http.Error(w, err.Error(), statusFromError(err))
			return
		}
	}
	http.Redirect(w, r, "/go-admin/themes", http.StatusSeeOther)
}

func (h Handler) permalinksPage(w http.ResponseWriter, r *http.Request, principal authz.Principal) {
	fixture := h.fixture(r)
	definitions := h.settingDefinitions("permalinks.post_pattern", "permalinks.page_pattern")
	fields := h.settingFields(r.Context(), definitions)
	updateFieldID(fields, "permalinks.post_pattern", "post_pattern", fixture.Label("field_post_pattern", "Post permalink pattern"))
	updateFieldID(fields, "permalinks.page_pattern", "page_pattern", fixture.Label("field_page_pattern", "Page permalink pattern"))
	form := h.contentEditorFromFixture(fixture, "permalinks", "/go-admin/permalinks", h.token("permalinks-write"), fields)
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
		{Key: "permalinks.post_pattern", Value: r.PostForm.Get("post_pattern")},
		{Key: "permalinks.page_pattern", Value: r.PostForm.Get("page_pattern")},
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
	h.renderSimple(w, r, principal, "headless", fixture, rows, "headless-settings", nil, "", nil, nil)
}

func (h Handler) runtimePage(w http.ResponseWriter, r *http.Request, principal authz.Principal) {
	fixture := h.fixture(r)
	info := h.runtimeInfo
	switchRule := fallbackValue(info.ProviderSwitchRule, "Durable provider switches require export, migration, restart or redeploy, and import through the JSON/site-package handoff. They are not runtime toggles.")
	rows := []blocks.SimpleListRow{
		{Label: fixture.Label("runtime_preset", "Preset"), Description: valueOrUnset(info.Preset), Status: "resolved"},
		{Label: fixture.Label("runtime_profile", "Runtime profile"), Description: valueOrUnset(info.RuntimeProfile), Status: "resolved"},
		{Label: fixture.Label("runtime_storage", "Storage profile"), Description: valueOrUnset(info.StorageProfile), Status: "resolved"},
		{Label: fixture.Label("runtime_deployment", "Deployment profile"), Description: valueOrUnset(info.DeploymentProfile), Status: "resolved"},
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
	for _, result := range h.health.Results(r.Context()) {
		description := result.Description
		if result.Error != "" {
			description = result.Error
		}
		rows = append(rows, blocks.SimpleListRow{
			ID:          result.ID,
			Label:       "Health: " + result.Label,
			Description: description,
			Status:      result.Status,
		})
	}
	if events, err := h.services.Audit.Recent(r.Context(), 10); err == nil {
		for _, event := range events {
			rows = append(rows, blocks.SimpleListRow{
				ID:          event.ID,
				Label:       "Audit: " + event.Action,
				Description: strings.TrimSpace(strings.Join([]string{event.ActorID, event.Resource, event.ResourceID}, " | ")),
				Status:      string(event.Status),
			})
		}
	}
	if errors, err := h.diagnostics.Recent(r.Context(), 10); err == nil {
		for _, item := range errors {
			rows = append(rows, blocks.SimpleListRow{
				ID:          item.ID,
				Label:       "Error: " + item.Source,
				Description: item.Message,
				Status:      item.Severity,
			})
		}
	}
	h.renderSimple(w, r, principal, "runtime", fixture, rows, "runtime-status", nil, "", nil, nil)
}

func (h Handler) screenOptionsSave(w http.ResponseWriter, r *http.Request, principal authz.Principal) {
	if err := r.ParseForm(); err != nil || !h.validToken(r.PostForm.Get("action_token"), "screen-options-write") {
		http.Error(w, "Invalid screen options submission.", http.StatusBadRequest)
		return
	}
	screen := r.PathValue("screen")
	values := []domainsettings.Value{
		{Key: domainsettings.Key("admin.screen." + screen + ".per_page"), Value: parseFormInt(r, "per_page")},
		{Key: domainsettings.Key("admin.screen." + screen + ".columns"), Value: strings.TrimSpace(r.PostForm.Get("columns"))},
	}
	for _, value := range values {
		if err := h.services.Settings.Save(r.Context(), principal, value); err != nil {
			http.Error(w, err.Error(), statusFromError(err))
			return
		}
	}
	http.Redirect(w, r, safeReturnTo(r.PostForm.Get("return_to")), http.StatusSeeOther)
}

func (h Handler) renderSimple(w http.ResponseWriter, r *http.Request, principal authz.Principal, screen string, bundle adminfixtures.AdminFixture, rows []blocks.SimpleListRow, marker string, fields []blocks.FieldData, formAction string, quickEdit *blocks.InlineFormData, bulk *blocks.BulkActionData) {
	screenData, _ := bundle.Screen(screen)
	title := fallbackValue(screenData.Title, titleFor(screen))
	description := fallbackValue(screenData.Description, "Manage admin content.")
	schema := h.pageTableSchema(screen)
	state := listui.ParseState(
		r,
		screen,
		schema,
		listui.FirstPositive(schema.PerPage, 25),
		h.screenPerPagePreference(r.Context(), screen, listui.FirstPositive(schema.PerPage, 25)),
		h.screenColumnsPreference(r.Context(), screen),
	)
	basePath := r.URL.Path
	pagedRows, _, totalPages, page := listui.ApplySimpleListState(rows, state, schema)
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
			Rows:        pagedRows,
			Actions:     actions,
			FormAction:  formAction, Token: h.token("admin-write"), Fields: fields,
			Headers:    h.simpleListHeaders(bundle),
			OpenLabel:  bundle.Label("action_open", "Open"),
			SaveLabel:  bundle.Label("action_save", "Save"),
			QuickLabel: bundle.Label("action_quick_edit", "Quick edit"),
			Pagination: elements.PaginationData{
				Page:          page,
				TotalPages:    totalPages,
				BaseHref:      state.BaseHref(basePath),
				PreviousLabel: bundle.Label("action_previous", "Previous"),
				NextLabel:     bundle.Label("action_next", "Next"),
			},
			Toolbar:    listui.BuildToolbarForm(bundle, schema, basePath, state),
			QuickEdit:  quickEdit,
			ScreenOpts: listui.BuildScreenOptionsForm(bundle, screen, schema, state, state.ReturnTo(basePath), h.token("screen-options-write")),
			Bulk:       bulk,
			Visible:    state.VisibleMap(schema),
		},
	}
	_ = web.Render(r.Context(), w, views.SimpleListPage(data))
}

func (h Handler) renderSecurityResult(w http.ResponseWriter, r *http.Request, principal authz.Principal, title string, description string, rows []blocks.SimpleListRow) {
	fixture := h.fixture(r)
	_ = web.Render(r.Context(), w, views.SimpleListPage(views.SimpleListPageData{
		Layout: h.layout(r, principal, title, "/go-admin/users"),
		Screen: "users",
		List: blocks.SimpleListData{
			Title:       title,
			Description: description,
			Marker:      "security-result",
			Rows:        rows,
			Headers:     h.simpleListHeaders(fixture),
			OpenLabel:   fixture.Label("action_open", "Open"),
			SaveLabel:   fixture.Label("action_save", "Save"),
			QuickLabel:  fixture.Label("action_quick_edit", "Quick edit"),
		},
	}))
}

func (h Handler) userSecuritySummary(ctx context.Context, user domainusers.User) string {
	parts := []string{}
	if h.authn.Enabled() {
		if recovery, err := h.authn.ListRecoveryCodes(ctx, user.ID); err == nil {
			active := 0
			for _, item := range recovery {
				if item.Available(time.Now().UTC()) {
					active++
				}
			}
			parts = append(parts, fmt.Sprintf("recovery:%d", active))
		}
		if resets, err := h.authn.ListResetTokens(ctx, user.ID); err == nil {
			active := 0
			for _, item := range resets {
				if item.Available(time.Now().UTC()) {
					active++
				}
			}
			parts = append(parts, fmt.Sprintf("reset-tokens:%d", active))
		}
		if tokens, err := h.authn.ListAppTokens(ctx, user.ID); err == nil {
			active := 0
			activeIDs := []string{}
			for _, item := range tokens {
				if item.Active(time.Now().UTC()) {
					active++
					activeIDs = append(activeIDs, item.Prefix)
				}
			}
			parts = append(parts, fmt.Sprintf("tokens:%d", active))
			if len(activeIDs) > 0 {
				parts = append(parts, "token-ids:"+strings.Join(activeIDs, ","))
			}
		}
	}
	if user.LastLoginAt != nil {
		parts = append(parts, "last-login:"+user.LastLoginAt.UTC().Format(time.RFC3339))
	}
	if user.MustChangePassword {
		parts = append(parts, "password-change-required")
	}
	return strings.Join(parts, " | ")
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
	schema := resource.Form
	schema.Fields = append(schema.Fields, cmspanel.MetadataFields(h.metaDefinitions(entry.Kind))...)
	fields := h.formFieldsFromFixture(bundle, "content-editor", contentFormFields(schema, entry, editorProvider))
	fields = fieldsWithoutIDs(fields, "meta_key", "meta_value")

	updateFieldValue(fields, "title", entry.Title.Value("en", "en"))
	updateFieldValue(fields, "slug", entry.Slug.Value("en", "en"))
	updateFieldValue(fields, "content", entry.Body.Value("en", "en"))
	updateFieldValue(fields, "excerpt", entry.Excerpt.Value("en", "en"))
	updateFieldValue(fields, "author_id", defaultValue(entry.AuthorID, "author-1"))
	updateFieldValue(fields, "featured_media_id", entry.FeaturedMediaID)
	updateFieldValue(fields, "template", entry.Template)
	updateFieldValue(fields, "terms", formatTerms(entry.Terms))
	h.updateMetaFieldValues(fields, entry)
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
	case "disabled", "external", "repository":
		return rest.SessionData{}, false
	case "", "fixture":
	default:
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
	return sessionFromCapabilities(principal.ID, principal.Capabilities, "fixture", false, 0)
}

func sessionFromLoginResult(result appauthn.PasswordLoginResult, policy domainauthn.SessionPolicy) rest.SessionData {
	capabilities := map[authz.Capability]struct{}{}
	for capability := range result.Principal.Capabilities {
		capabilities[capability] = struct{}{}
	}
	return sessionFromCapabilities(result.Principal.ID, capabilities, result.ProviderID, result.MustChangePassword, policy.AbsoluteTTL)
}

func sessionFromCapabilities(userID string, capabilitiesMap map[authz.Capability]struct{}, providerID string, mustChange bool, absoluteTTL time.Duration) rest.SessionData {
	capabilities := make([]string, 0, len(capabilitiesMap))
	for capability := range capabilitiesMap {
		capabilities = append(capabilities, string(capability))
	}
	session := rest.SessionData{
		UserID:             userID,
		Capabilities:       capabilities,
		ProviderID:         providerID,
		MustChangePassword: mustChange,
		IssuedAtUnix:       time.Now().UTC().Unix(),
	}
	if absoluteTTL > 0 {
		session.AbsoluteExpiryUnix = time.Now().UTC().Add(absoluteTTL).Unix()
	}
	return session
}

func (h Handler) sessionPolicy() domainauthn.SessionPolicy {
	secureCookies := false
	switch strings.TrimSpace(strings.ToLower(h.runtimeInfo.DeploymentProfile)) {
	case string(runtimeprofile.DeploymentProfileContainer), string(runtimeprofile.DeploymentProfileServerless), string(runtimeprofile.DeploymentProfileSSH):
		secureCookies = true
	}
	policy := domainauthn.DefaultSessionPolicy(secureCookies)
	policy.IdleTTL = time.Duration(h.intSetting(context.Background(), "auth.session.idle_ttl_hours", int(policy.IdleTTL/time.Hour))) * time.Hour
	policy.AbsoluteTTL = time.Duration(h.intSetting(context.Background(), "auth.session.absolute_ttl_hours", int(policy.AbsoluteTTL/time.Hour))) * time.Hour
	policy.MaxAttempts = h.intSetting(context.Background(), "auth.login.max_attempts", policy.MaxAttempts)
	policy.AttemptWindow = time.Duration(h.intSetting(context.Background(), "auth.login.attempt_window_minutes", int(policy.AttemptWindow/time.Minute))) * time.Minute
	policy.LockoutWindow = time.Duration(h.intSetting(context.Background(), "auth.login.lockout_minutes", int(policy.LockoutWindow/time.Minute))) * time.Minute
	return policy.Normalized()
}

func (h Handler) audit(r *http.Request, principal authz.Principal, action string, resource string, resourceID string, status domainaudit.Status, details map[string]any) {
	if !h.services.Audit.Enabled() {
		return
	}
	_ = h.services.Audit.Record(r.Context(), domainaudit.Event{
		ActorID:    principal.ID,
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
		Status:     status,
		RemoteAddr: r.RemoteAddr,
		UserAgent:  r.UserAgent(),
		Details:    details,
	})
}

func (h Handler) intSetting(ctx context.Context, key domainsettings.Key, fallback int) int {
	value, ok, err := h.services.Settings.Get(ctx, key)
	if err != nil || !ok || value.Value == nil {
		return fallback
	}
	switch typed := value.Value.(type) {
	case int:
		if typed > 0 {
			return typed
		}
	case int64:
		if typed > 0 {
			return int(typed)
		}
	case float64:
		if typed > 0 {
			return int(typed)
		}
	case string:
		parsed, err := strconv.Atoi(strings.TrimSpace(typed))
		if err == nil && parsed > 0 {
			return parsed
		}
	}
	return fallback
}

func (h Handler) logError(r *http.Request, source string, message string, severity string, details map[string]any) {
	if !h.diagnostics.Enabled() {
		return
	}
	_ = h.diagnostics.Record(r.Context(), source, message, severity, details)
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

func adminReturnTo(value string, fallback string) string {
	if strings.HasPrefix(strings.TrimSpace(value), "/go-admin") {
		return value
	}
	return fallback
}

func localized(value string) domaincontent.LocalizedText {
	return domaincontent.LocalizedText{"en": strings.TrimSpace(value)}
}

func (h Handler) formMetadata(r *http.Request, kind domaincontent.Kind) domaincontent.Metadata {
	metadata := domaincontent.Metadata{}
	for _, definition := range h.metaDefinitions(kind) {
		fieldName := appmeta.FormFieldName(definition.Key)
		value := r.PostForm.Get(fieldName)
		if definition.Type == domainmeta.ValueTypeBoolean && r.PostForm.Get(fieldName) == "" {
			value = "false"
		}
		metadata[definition.Key] = domaincontent.MetaValue{Value: value, Public: definition.Public}
	}
	customKey := strings.TrimSpace(r.PostForm.Get("custom_meta_key"))
	if customKey != "" {
		publicValue, _ := strconv.ParseBool(defaultValue(r.PostForm.Get("custom_meta_public"), "false"))
		metadata[customKey] = domaincontent.MetaValue{
			Value:  strings.TrimSpace(r.PostForm.Get("custom_meta_value")),
			Public: publicValue,
		}
	}
	if len(metadata) == 0 {
		return nil
	}
	return metadata
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

func (h Handler) metaDefinitions(kind domaincontent.Kind) []domainmeta.Definition {
	if h.metaRegistry == nil {
		return nil
	}
	return h.metaRegistry.Definitions(kind)
}

func (h Handler) updateMetaFieldValues(fields []blocks.FieldData, entry domaincontent.Entry) {
	metadata := entry.Metadata
	for _, definition := range h.metaDefinitions(entry.Kind) {
		value := ""
		if item, ok := metadata[definition.Key]; ok {
			value = formatMetaValue(item.Value)
		} else if definition.Default != nil {
			value = formatMetaValue(definition.Default)
		}
		updateFieldValue(fields, appmeta.FormFieldName(definition.Key), value)
	}
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

func formatMetaValue(value any) string {
	switch typed := value.(type) {
	case nil:
		return ""
	case bool:
		if typed {
			return "true"
		}
		return "false"
	case string:
		return typed
	default:
		return fmt.Sprint(typed)
	}
}

func formatSettingValue(value any) string {
	switch typed := value.(type) {
	case nil:
		return ""
	case bool:
		if typed {
			return "true"
		}
		return "false"
	case string:
		return typed
	default:
		return fmt.Sprint(typed)
	}
}

func parseFormInt(r *http.Request, key string) int {
	value := strings.TrimSpace(r.PostForm.Get(key))
	if value == "" {
		return 0
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0
	}
	return parsed
}

func parseFormInt64(r *http.Request, key string) int64 {
	value := strings.TrimSpace(r.PostForm.Get(key))
	if value == "" {
		return 0
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0
	}
	return parsed
}

func parseCapabilities(raw string) []authz.Capability {
	parts := strings.Split(raw, ",")
	result := []authz.Capability{}
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		result = append(result, authz.Capability(part))
	}
	return result
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

func (h Handler) screenPerPagePreference(ctx context.Context, screen string, fallback int) int {
	value, ok, err := h.services.Settings.Get(ctx, domainsettings.Key("admin.screen."+screen+".per_page"))
	if err != nil || !ok {
		return fallback
	}
	return appsettings.IntValue(value, fallback)
}

func (h Handler) screenColumnsPreference(ctx context.Context, screen string) []string {
	value, ok, err := h.services.Settings.Get(ctx, domainsettings.Key("admin.screen."+screen+".columns"))
	if err != nil || !ok {
		return nil
	}
	return appsettings.VisibleColumns(value)
}

func (h Handler) settingDefinitions(keys ...string) []domainsettings.Definition {
	if h.settings == nil {
		return nil
	}
	result := make([]domainsettings.Definition, 0, len(keys))
	for _, key := range keys {
		definition, ok := h.settings.Definition(domainsettings.Key(key))
		if ok {
			result = append(result, definition)
		}
	}
	return result
}

func (h Handler) themeOptionDefinitions(activeTheme string) []domainsettings.Definition {
	if h.settings == nil {
		return nil
	}
	result := []domainsettings.Definition{}
	owner := "theme:" + activeTheme
	for _, definition := range h.settings.DefinitionsByGroup(domainsettings.GroupTheme) {
		if definition.Owner == owner && !strings.HasPrefix(string(definition.Key), "theme.preview") && string(definition.Key) != platformthemes.ActiveThemeKey && string(definition.Key) != platformthemes.StylePresetKey {
			result = append(result, definition)
		}
	}
	return result
}

func (h Handler) settingFields(ctx context.Context, definitions []domainsettings.Definition) []blocks.FieldData {
	fields := make([]blocks.FieldData, 0, len(definitions))
	for _, definition := range definitions {
		value, ok, err := h.services.Settings.Get(ctx, definition.Key)
		if err != nil || !ok {
			value = domainsettings.Value{Key: definition.Key, Value: definition.Default, Public: definition.Public}
		}
		fields = append(fields, settingField(definition, value))
	}
	return fields
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

func (h Handler) contentBulkForm(bundle adminfixtures.AdminFixture, principal authz.Principal, schema panel.TableSchema[authz.Capability], basePath string, returnTo string) *blocks.BulkActionData {
	options := []ui.FieldOption{}
	for _, action := range schema.BulkActions {
		if action.Capability == "" || principal.Has(action.Capability) {
			options = append(options, ui.FieldOption{Value: action.ID, Label: action.Label})
		}
	}
	return listui.BuildBulkForm(bundle, basePath+"/bulk", h.token("content-bulk"), returnTo, options)
}

func (h Handler) contentQuickEditForm(r *http.Request, principal authz.Principal, bundle adminfixtures.AdminFixture, state listui.State, kind domaincontent.Kind, basePath string) *blocks.InlineFormData {
	if strings.TrimSpace(state.EditID) == "" {
		return nil
	}
	entry, err := h.services.Content.Get(r.Context(), principal, domaincontent.ID(state.EditID))
	if err != nil || entry.Kind != kind {
		return nil
	}
	return listui.BuildContentQuickEditForm(bundle, state, basePath, h.token("content-quick-edit"), entry, h.statusOptions(bundle))
}

func (h Handler) taxonomyQuickEditForm(bundle adminfixtures.AdminFixture, state listui.State, items []domaintaxonomy.Definition, basePath string) *blocks.InlineFormData {
	for _, item := range items {
		if string(item.Type) != state.EditID {
			continue
		}
		return listui.BuildTaxonomyQuickEditForm(bundle, state, basePath, h.token("admin-write"), item)
	}
	return nil
}

func (h Handler) termQuickEditForm(bundle adminfixtures.AdminFixture, state listui.State, items []domaintaxonomy.Term, basePath string) *blocks.InlineFormData {
	for _, item := range items {
		if string(item.ID) != state.EditID {
			continue
		}
		return listui.BuildTermQuickEditForm(bundle, state, basePath, h.token("admin-write"), item)
	}
	return nil
}

func (h Handler) mediaQuickEditForm(bundle adminfixtures.AdminFixture, state listui.State, items []domainmedia.Asset, basePath string) *blocks.InlineFormData {
	for _, item := range items {
		if string(item.ID) != state.EditID {
			continue
		}
		return listui.BuildMediaQuickEditForm(bundle, state, basePath, h.token("admin-write"), item)
	}
	return nil
}

func (h Handler) menuQuickEditForm(bundle adminfixtures.AdminFixture, state listui.State, items []domainmenus.Menu, basePath string) *blocks.InlineFormData {
	for _, item := range items {
		if string(item.ID) != state.EditID {
			continue
		}
		return listui.BuildMenuQuickEditForm(bundle, state, basePath, h.token("admin-write"), item)
	}
	return nil
}

func (h Handler) userQuickEditForm(bundle adminfixtures.AdminFixture, state listui.State, items []domainusers.User, basePath string) *blocks.InlineFormData {
	for _, item := range items {
		if string(item.ID) != state.EditID {
			continue
		}
		return listui.BuildUserQuickEditForm(bundle, state, basePath, h.token("admin-write"), item)
	}
	return nil
}

func (h Handler) userBulkForm(bundle adminfixtures.AdminFixture, principal authz.Principal, returnTo string) *blocks.BulkActionData {
	if !principal.Has(authz.CapabilityUsersManage) {
		return nil
	}
	return listui.BuildBulkForm(bundle, "/go-admin/users", h.token("admin-write"), returnTo, []ui.FieldOption{
		{Value: "status:active", Label: "Set active"},
		{Value: "status:suspended", Label: "Set suspended"},
		{Value: "status:deleted", Label: "Set deleted"},
	})
}

func contentStatusesFilter(value string) []domaincontent.Status {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return []domaincontent.Status{domaincontent.Status(value)}
}

func contentSortField(value string) domaincontent.SortField {
	switch strings.TrimSpace(value) {
	case string(domaincontent.SortTitle):
		return domaincontent.SortTitle
	case string(domaincontent.SortSlug):
		return domaincontent.SortSlug
	case string(domaincontent.SortCreatedAt):
		return domaincontent.SortCreatedAt
	case string(domaincontent.SortPublishedAt):
		return domaincontent.SortPublishedAt
	default:
		return domaincontent.SortUpdatedAt
	}
}

func (h Handler) pageTableSchema(screen string) panel.TableSchema[authz.Capability] {
	for _, page := range cmspanel.AdminPages() {
		if string(page.ID) == screen {
			return page.Table
		}
	}
	return panel.TableSchema[authz.Capability]{}
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

func settingField(definition domainsettings.Definition, value domainsettings.Value) blocks.FieldData {
	field := blocks.FieldData{
		ID:          string(definition.Key),
		Name:        string(definition.Key),
		Label:       definition.Label,
		Value:       formatSettingValue(value.Value),
		Placeholder: definition.Placeholder,
		Hint:        definition.Description,
	}
	switch definition.FieldHint {
	case domainsettings.FieldHintTextarea:
		field.Component = "textarea"
		field.Rows = 3
	case domainsettings.FieldHintCheckbox:
		field.Type = "checkbox"
	case domainsettings.FieldHintNumber:
		field.Type = "number"
	case domainsettings.FieldHintSelect:
		field.Component = "select"
		field.Options = settingOptions(definition.Options)
	}
	return field
}

func settingOptions(options []domainsettings.Option) []ui.FieldOption {
	result := make([]ui.FieldOption, 0, len(options))
	for _, option := range options {
		result = append(result, ui.FieldOption{Value: option.Value, Label: option.Label})
	}
	return result
}

func updateFieldID(fields []blocks.FieldData, oldID string, newID string, label string) {
	for i := range fields {
		if fields[i].ID != oldID {
			continue
		}
		fields[i].ID = newID
		fields[i].Name = newID
		if label != "" {
			fields[i].Label = label
		}
	}
}

func (h Handler) themeOptions() []ui.FieldOption {
	items := h.themeRegistry.List()
	result := make([]ui.FieldOption, 0, len(items))
	for _, item := range items {
		result = append(result, ui.FieldOption{Value: string(item.ID), Label: item.Name})
	}
	return result
}

func (h Handler) themePresetOptions(themeID string) []ui.FieldOption {
	presets := h.themeRegistry.ListPresets(themeID)
	result := make([]ui.FieldOption, 0, len(presets))
	for _, preset := range presets {
		result = append(result, ui.FieldOption{Value: preset.ID, Label: preset.Name})
	}
	return result
}

func (h Handler) userByID(ctx context.Context, id domainusers.ID) (domainusers.User, bool, error) {
	items, err := h.services.Users.List(ctx)
	if err != nil {
		return domainusers.User{}, false, err
	}
	for _, item := range items {
		if item.ID == id {
			return item, true, nil
		}
	}
	return domainusers.User{}, false, nil
}

func fieldsWithoutIDs(fields []blocks.FieldData, ids ...string) []blocks.FieldData {
	excluded := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		excluded[id] = struct{}{}
	}
	result := make([]blocks.FieldData, 0, len(fields))
	for _, field := range fields {
		if _, ok := excluded[field.ID]; ok {
			continue
		}
		result = append(result, field)
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
	used := map[string]struct{}{}
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
			used[override.Name] = struct{}{}
		}
		if override.ID != "" {
			used[override.ID] = struct{}{}
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
	for _, field := range fallback {
		if field.ID != "" {
			if _, ok := used[field.ID]; ok {
				continue
			}
		}
		if field.Name != "" {
			if _, ok := used[field.Name]; ok {
				continue
			}
		}
		result = append(result, field)
	}
	return result
}

func (h Handler) screenActions(screen string, fixture adminfixtures.AdminFixture) []elements.Action {
	return h.registry.ScreenActions(screen, fixture)
}

func (h Handler) mediaFields(r *http.Request) []blocks.FieldData {
	fixture := h.fixture(r)
	return h.formFieldsFromFixture(fixture, "media", []blocks.FieldData{
		{ID: "id", Name: "id", Label: fixture.Label("field_id", "ID"), Required: true},
		{ID: "filename", Name: "filename", Label: fixture.Label("field_filename", "Filename"), Required: true},
		{ID: "mime_type", Name: "mime_type", Label: fixture.Label("field_mime_type", "MIME type"), Required: true, Placeholder: "image/webp"},
		{ID: "size_bytes", Name: "size_bytes", Label: fixture.Label("field_size_bytes", "Size bytes"), Type: "number"},
		{ID: "width", Name: "width", Label: fixture.Label("field_width", "Width"), Type: "number"},
		{ID: "height", Name: "height", Label: fixture.Label("field_height", "Height"), Type: "number"},
		{ID: "public_url", Name: "public_url", Label: fixture.Label("field_public_url", "Public URL"), Required: true, Placeholder: "https://cdn.example.test/image.webp"},
		{ID: "alt_text", Name: "alt_text", Label: fixture.Label("field_alt_text", "Alt text")},
		{ID: "caption", Name: "caption", Label: fixture.Label("field_caption", "Caption"), Component: "textarea", Rows: 3},
		{ID: "provider", Name: "provider", Label: fixture.Label("field_provider", "Provider"), Placeholder: "s3"},
		{ID: "provider_key", Name: "provider_key", Label: fixture.Label("field_provider_key", "Provider key"), Placeholder: "media/originals/cover.webp"},
		{ID: "provider_url", Name: "provider_url", Label: fixture.Label("field_provider_url", "Provider URL"), Placeholder: "https://bucket.example.test/cover.webp"},
		{ID: "provider_checksum", Name: "provider_checksum", Label: fixture.Label("field_provider_checksum", "Checksum")},
		{ID: "provider_etag", Name: "provider_etag", Label: fixture.Label("field_provider_etag", "ETag")},
	})
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
