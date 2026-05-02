package admin

import (
	"context"
	"fmt"
	"net/http"
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
	domainusers "github.com/fastygo/cms/internal/domain/users"
	"github.com/fastygo/cms/internal/site/assets"
	"github.com/fastygo/cms/internal/site/ui/blocks"
	"github.com/fastygo/cms/internal/site/ui/elements"
	"github.com/fastygo/cms/internal/site/views"
	"github.com/fastygo/framework/pkg/app"
	frameworkauth "github.com/fastygo/framework/pkg/auth"
	"github.com/fastygo/framework/pkg/web"
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
	services Services
	auth     rest.Authenticator
	secret   string
}

type actionToken struct {
	Action string `json:"action"`
	Exp    int64  `json:"exp"`
}

func NewHandler(services Services, authenticator rest.Authenticator, secret string) Handler {
	return Handler{services: services, auth: authenticator, secret: secret}
}

func (h Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /go-login", h.loginPage)
	mux.HandleFunc("POST /go-login", h.loginSubmit)
	mux.HandleFunc("POST /go-logout", h.logoutSubmit)
	mux.HandleFunc("GET /go-admin", h.protect(h.dashboard))
	mux.HandleFunc("GET /go-admin/posts", h.protect(h.contentList(domaincontent.KindPost, "posts")))
	mux.HandleFunc("GET /go-admin/posts/new", h.protect(h.contentNew(domaincontent.KindPost, "posts")))
	mux.HandleFunc("POST /go-admin/posts", h.protect(h.contentCreate(domaincontent.KindPost)))
	mux.HandleFunc("GET /go-admin/posts/{id}/edit", h.protect(h.contentEdit("posts")))
	mux.HandleFunc("POST /go-admin/posts/{id}", h.protect(h.contentUpdate))
	mux.HandleFunc("POST /go-admin/posts/{id}/trash", h.protect(h.contentTrash))
	mux.HandleFunc("GET /go-admin/pages", h.protect(h.contentList(domaincontent.KindPage, "pages")))
	mux.HandleFunc("GET /go-admin/pages/new", h.protect(h.contentNew(domaincontent.KindPage, "pages")))
	mux.HandleFunc("POST /go-admin/pages", h.protect(h.contentCreate(domaincontent.KindPage)))
	mux.HandleFunc("GET /go-admin/pages/{id}/edit", h.protect(h.contentEdit("pages")))
	mux.HandleFunc("POST /go-admin/pages/{id}", h.protect(h.contentUpdate))
	mux.HandleFunc("POST /go-admin/pages/{id}/trash", h.protect(h.contentTrash))
	mux.HandleFunc("GET /go-admin/content-types", h.protect(h.contentTypesPage))
	mux.HandleFunc("POST /go-admin/content-types", h.protect(h.contentTypeCreate))
	mux.HandleFunc("GET /go-admin/taxonomies", h.protect(h.taxonomiesPage))
	mux.HandleFunc("POST /go-admin/taxonomies", h.protect(h.taxonomyCreate))
	mux.HandleFunc("GET /go-admin/taxonomies/{type}/terms", h.protect(h.termsPage))
	mux.HandleFunc("POST /go-admin/taxonomies/{type}/terms", h.protect(h.termCreate))
	mux.HandleFunc("GET /go-admin/media", h.protect(h.mediaPage))
	mux.HandleFunc("POST /go-admin/media", h.protect(h.mediaSave))
	mux.HandleFunc("GET /go-admin/menus", h.protect(h.menusPage))
	mux.HandleFunc("POST /go-admin/menus", h.protect(h.menuSave))
	mux.HandleFunc("GET /go-admin/users", h.protect(h.usersPage))
	mux.HandleFunc("POST /go-admin/users", h.protect(h.userSave))
	mux.HandleFunc("GET /go-admin/authors", h.protect(h.authorsPage))
	mux.HandleFunc("GET /go-admin/capabilities", h.protect(h.capabilitiesPage))
	mux.HandleFunc("GET /go-admin/settings", h.protect(h.settingsPage))
	mux.HandleFunc("POST /go-admin/settings", h.protect(h.settingsSave))
	mux.HandleFunc("GET /go-admin/headless", h.protect(h.headlessPage))
}

func (h Handler) NavItems() []app.NavItem {
	return []app.NavItem{
		{Label: "Dashboard", Path: "/go-admin", Icon: "home", Order: 0},
		{Label: "Posts", Path: "/go-admin/posts", Icon: "file", Order: 1},
		{Label: "Pages", Path: "/go-admin/pages", Icon: "book", Order: 2},
		{Label: "Content types", Path: "/go-admin/content-types", Icon: "box", Order: 3},
		{Label: "Taxonomies", Path: "/go-admin/taxonomies", Icon: "boxes", Order: 4},
		{Label: "Media", Path: "/go-admin/media", Icon: "image", Order: 5},
		{Label: "Menus", Path: "/go-admin/menus", Icon: "menu", Order: 6},
		{Label: "Users", Path: "/go-admin/users", Icon: "users", Order: 7},
		{Label: "Authors", Path: "/go-admin/authors", Icon: "users", Order: 8},
		{Label: "Capabilities", Path: "/go-admin/capabilities", Icon: "shield", Order: 9},
		{Label: "Settings", Path: "/go-admin/settings", Icon: "sliders", Order: 10},
		{Label: "Headless", Path: "/go-admin/headless", Icon: "server", Order: 11},
	}
}

func (h Handler) loginPage(w http.ResponseWriter, r *http.Request) {
	data := views.LoginPageData{
		Title:       "GoCMS Login",
		Subtitle:    "Use fixture credentials: admin@example.test / admin.",
		ReturnTo:    returnTo(r),
		ActionToken: h.token("login"),
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
	session, ok := fixtureSession(r.PostForm.Get("email"), r.PostForm.Get("password"))
	if !ok {
		data := views.LoginPageData{Title: "GoCMS Login", Subtitle: "Use fixture credentials: admin@example.test / admin.", Error: "Invalid credentials.", ReturnTo: returnTo(r), ActionToken: h.token("login")}
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
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := h.auth.Principal(r)
		if !ok || !principal.Has(authz.CapabilityControlPanelAccess) {
			http.Redirect(w, r, "/go-login?return_to="+r.URL.Path, http.StatusSeeOther)
			return
		}
		next(w, r, principal)
	}
}

func (h Handler) dashboard(w http.ResponseWriter, r *http.Request, principal authz.Principal) {
	postCount := h.contentCount(r.Context(), domaincontent.KindPost)
	pageCount := h.contentCount(r.Context(), domaincontent.KindPage)
	taxonomies, _ := h.services.Taxonomy.ListDefinitions(r.Context())
	media, _ := h.services.Media.List(r.Context())
	data := views.DashboardData{
		Layout: h.layout(r, principal, "Dashboard", "/go-admin"),
		Cards: []blocks.StatCard{
			{Label: "Posts", Value: strconv.Itoa(postCount), Href: "/go-admin/posts"},
			{Label: "Pages", Value: strconv.Itoa(pageCount), Href: "/go-admin/pages"},
			{Label: "Taxonomies", Value: strconv.Itoa(len(taxonomies)), Href: "/go-admin/taxonomies"},
			{Label: "Media", Value: strconv.Itoa(len(media)), Href: "/go-admin/media"},
		},
	}
	_ = web.Render(r.Context(), w, views.DashboardPage(data))
}

func (h Handler) contentList(kind domaincontent.Kind, screen string) func(http.ResponseWriter, *http.Request, authz.Principal) {
	return func(w http.ResponseWriter, r *http.Request, principal authz.Principal) {
		result, _ := h.services.Content.List(r.Context(), domaincontent.Query{Kinds: []domaincontent.Kind{kind}, Page: 1, PerPage: 50, SortBy: domaincontent.SortUpdatedAt, SortDesc: true})
		rows := make([]blocks.ContentRow, 0, len(result.Items))
		for _, entry := range result.Items {
			rows = append(rows, blocks.ContentRow{
				ID:      string(entry.ID),
				Title:   entry.Title.Value("en", "en"),
				Slug:    entry.Slug.Value("en", "en"),
				Status:  string(entry.Status),
				Author:  entry.AuthorID,
				EditURL: "/go-admin/" + screen + "/" + string(entry.ID) + "/edit",
			})
		}
		data := views.ContentListPageData{
			Layout: h.layout(r, principal, titleFor(screen), "/go-admin/"+screen),
			Screen: screen,
			Table: blocks.ContentTableData{
				Title:       titleFor(screen),
				Description: "Create, edit, publish, schedule, trash, and restore content.",
				Rows:        rows,
				Actions:     []elements.Action{{Label: "Create", Href: "/go-admin/" + screen + "/new", Enabled: principal.Has(authz.CapabilityContentCreate)}},
				Pagination:  elements.PaginationData{Page: 1, TotalPages: 1, BaseHref: "/go-admin/" + screen},
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
		data := views.ContentEditPageData{
			Layout: h.layout(r, principal, "New "+singular(screen), "/go-admin/"+screen),
			Screen: screen + "-edit",
			Editor: h.contentEditor("New "+singular(screen), "Create a draft and choose publish state.", "/go-admin/"+screen, h.token("content-write"), domaincontent.Entry{Kind: kind, Status: domaincontent.StatusDraft}),
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
		data := views.ContentEditPageData{
			Layout: h.layout(r, principal, "Edit "+entry.Title.Value("en", "en"), "/go-admin/"+screen),
			Screen: screen + "-edit",
			Editor: h.contentEditor("Edit "+entry.Title.Value("en", "en"), "Update content fields and publication state.", "/go-admin/"+screen+"/"+string(entry.ID), h.token("content-write"), entry),
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
	items, _ := h.services.ContentTypes.List(r.Context())
	rows := make([]blocks.SimpleListRow, 0, len(items))
	for _, item := range items {
		rows = append(rows, blocks.SimpleListRow{Label: string(item.ID), Description: item.Label, Status: visibleStatus(item.RESTVisible), ActionURL: ""})
	}
	h.renderSimple(w, r, principal, "content-types", "Content types", "Manage built-in and custom content types.", rows, "content-types", simpleFields("content-types"), "/go-admin/content-types")
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
	items, _ := h.services.Taxonomy.ListDefinitions(r.Context())
	rows := make([]blocks.SimpleListRow, 0, len(items))
	for _, item := range items {
		rows = append(rows, blocks.SimpleListRow{Label: string(item.Type), Description: item.Label, Status: string(item.Mode), ActionURL: "/go-admin/taxonomies/" + string(item.Type) + "/terms"})
	}
	h.renderSimple(w, r, principal, "taxonomies", "Taxonomies", "Manage taxonomy definitions and terms.", rows, "taxonomies", simpleFields("taxonomies"), "/go-admin/taxonomies")
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
	taxonomyType := domaintaxonomy.Type(r.PathValue("type"))
	items, _ := h.services.Taxonomy.ListTerms(r.Context(), taxonomyType)
	rows := make([]blocks.SimpleListRow, 0, len(items))
	for _, item := range items {
		rows = append(rows, blocks.SimpleListRow{Label: string(item.ID), Description: item.Name.Value("en", "en"), Status: string(item.Type)})
	}
	h.renderSimple(w, r, principal, "terms", "Terms", "Manage taxonomy terms.", rows, "terms", simpleFields("terms"), "/go-admin/taxonomies/"+string(taxonomyType)+"/terms")
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
	items, _ := h.services.Media.List(r.Context())
	rows := make([]blocks.SimpleListRow, 0, len(items))
	for _, item := range items {
		rows = append(rows, blocks.SimpleListRow{Label: string(item.ID), Description: item.Filename, Status: item.MimeType})
	}
	h.renderSimple(w, r, principal, "media", "Media", "Manage media metadata and featured media references.", rows, "media", simpleFields("media"), "/go-admin/media")
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
	items, _ := h.services.Menus.List(r.Context())
	rows := make([]blocks.SimpleListRow, 0, len(items))
	for _, item := range items {
		rows = append(rows, blocks.SimpleListRow{Label: string(item.ID), Description: item.Name, Status: string(item.Location)})
	}
	h.renderSimple(w, r, principal, "menus", "Menus", "Manage navigation menus.", rows, "menus", simpleFields("menus"), "/go-admin/menus")
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
	items, _ := h.services.Users.List(r.Context())
	rows := make([]blocks.SimpleListRow, 0, len(items))
	for _, item := range items {
		rows = append(rows, blocks.SimpleListRow{Label: string(item.ID), Description: item.DisplayName, Status: string(item.Status)})
	}
	h.renderSimple(w, r, principal, "users", "Users", "Manage users and account state.", rows, "users", simpleFields("users"), "/go-admin/users")
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
	items, _ := h.services.Users.List(r.Context())
	rows := make([]blocks.SimpleListRow, 0, len(items))
	for _, item := range items {
		author := item.PublicAuthor()
		rows = append(rows, blocks.SimpleListRow{Label: string(author.ID), Description: author.DisplayName, Status: author.Slug})
	}
	h.renderSimple(w, r, principal, "authors", "Authors", "Review public author projections.", rows, "authors", nil, "")
}

func (h Handler) capabilitiesPage(w http.ResponseWriter, r *http.Request, principal authz.Principal) {
	rows := []blocks.SimpleListRow{
		{Label: "Content", Description: "Create, edit, publish, schedule, trash, restore", Status: "core"},
		{Label: "Taxonomies", Description: "Manage and assign terms", Status: "core"},
		{Label: "Settings", Description: "Manage private and public settings", Status: "restricted"},
		{Label: "Users", Description: "Manage accounts and roles", Status: "restricted"},
	}
	h.renderSimple(w, r, principal, "capabilities", "Roles and capabilities", "Review capability groups enforced server-side.", rows, "capabilities", nil, "")
}

func (h Handler) settingsPage(w http.ResponseWriter, r *http.Request, principal authz.Principal) {
	form := blocks.ContentEditorData{
		Title: "Settings", Description: "Configure public site settings.", Action: "/go-admin/settings", Token: h.token("settings-write"),
		Fields: []blocks.FieldData{{ID: "site_title", Name: "site_title", Label: "Site title", Value: "GoCMS", Required: true}, {ID: "public_rendering", Name: "public_rendering", Label: "Public rendering", Value: "disabled"}},
	}
	_ = web.Render(r.Context(), w, views.SettingsPage(views.SettingsPageData{Layout: h.layout(r, principal, "Settings", "/go-admin/settings"), Screen: "settings", Form: form}))
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

func (h Handler) headlessPage(w http.ResponseWriter, r *http.Request, principal authz.Principal) {
	rows := []blocks.SimpleListRow{
		{Label: "REST", Description: "/go-json/go/v2/ is enabled", Status: "enabled"},
		{Label: "GraphQL", Description: "Planned as Pass 4 plugin", Status: "planned"},
		{Label: "Public rendering", Description: "Can remain disabled for headless mode", Status: "disabled"},
	}
	h.renderSimple(w, r, principal, "headless", "API and headless settings", "Inspect API delivery mode and upcoming plugin state.", rows, "headless-settings", nil, "")
}

func (h Handler) renderSimple(w http.ResponseWriter, r *http.Request, principal authz.Principal, screen string, title string, description string, rows []blocks.SimpleListRow, marker string, fields []blocks.FieldData, formAction string) {
	data := views.SimpleListPageData{
		Layout: h.layout(r, principal, title, "/go-admin/"+screen),
		Screen: screen,
		List: blocks.SimpleListData{
			Title: title, Description: description, Marker: marker, Rows: rows,
			FormAction: formAction, Token: h.token("admin-write"), Fields: fields,
		},
	}
	_ = web.Render(r.Context(), w, views.SimpleListPage(data))
}

func (h Handler) layout(r *http.Request, principal authz.Principal, title string, active string) views.LayoutData {
	resolvedAssets := assets.Resolve()
	return views.LayoutData{
		Title: title, Lang: "en", Brand: "GoCMS", Active: active, NavItems: h.NavItems(),
		Account: elements.AccountActionsData{Email: principal.ID, Token: h.token("logout")},
		Assets: views.AssetPaths{
			CSS:     resolvedAssets.CSS,
			ThemeJS: resolvedAssets.ThemeJS,
			AppJS:   resolvedAssets.AppJS,
		},
	}
}

func (h Handler) contentEditor(title string, description string, action string, token string, entry domaincontent.Entry) blocks.ContentEditorData {
	return blocks.ContentEditorData{
		Title: title, Description: description, Action: action, Token: token, Status: string(defaultStatus(entry.Status)),
		Fields: []blocks.FieldData{
			{ID: "title", Name: "title", Label: "Title", Value: entry.Title.Value("en", "en"), Required: true},
			{ID: "slug", Name: "slug", Label: "Slug", Value: entry.Slug.Value("en", "en"), Required: true},
			{ID: "content", Name: "content", Label: "Content", Value: entry.Body.Value("en", "en"), Component: "textarea", Rows: 8},
			{ID: "excerpt", Name: "excerpt", Label: "Excerpt", Value: entry.Excerpt.Value("en", "en"), Component: "textarea", Rows: 3},
			{ID: "author_id", Name: "author_id", Label: "Author ID", Value: defaultValue(entry.AuthorID, "author-1")},
			{ID: "featured_media_id", Name: "featured_media_id", Label: "Featured media ID", Value: entry.FeaturedMediaID},
			{ID: "template", Name: "template", Label: "Template", Value: entry.Template},
			{ID: "terms", Name: "terms", Label: "Taxonomy terms", Value: formatTerms(entry.Terms), Placeholder: "category:news,tag:go"},
			{ID: "meta_key", Name: "meta_key", Label: "Metadata key", Placeholder: "seo_title"},
			{ID: "meta_value", Name: "meta_value", Label: "Metadata value"},
		},
		Actions: []elements.Action{{Label: "Back", Href: contentListPath(entry.Kind), Enabled: true}},
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

func fixtureSession(email string, password string) (rest.SessionData, bool) {
	principals := rest.DevBearerPrincipals()
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

func simpleFields(screen string) []blocks.FieldData {
	switch screen {
	case "content-types":
		return []blocks.FieldData{
			{ID: "id", Name: "id", Label: "Content type ID", Required: true, Placeholder: "product"},
			{ID: "label", Name: "label", Label: "Label", Required: true, Placeholder: "Products"},
		}
	case "taxonomies":
		return []blocks.FieldData{
			{ID: "type", Name: "type", Label: "Taxonomy type", Required: true, Placeholder: "brand"},
			{ID: "label", Name: "label", Label: "Label", Required: true, Placeholder: "Brands"},
			{ID: "mode", Name: "mode", Label: "Mode", Value: "flat"},
		}
	case "terms":
		return []blocks.FieldData{
			{ID: "id", Name: "id", Label: "Term ID", Required: true, Placeholder: "term-1"},
			{ID: "name", Name: "name", Label: "Name", Required: true},
			{ID: "slug", Name: "slug", Label: "Slug", Required: true},
		}
	case "media":
		return []blocks.FieldData{
			{ID: "id", Name: "id", Label: "Media ID", Required: true, Placeholder: "media-1"},
			{ID: "filename", Name: "filename", Label: "Filename", Required: true},
			{ID: "mime_type", Name: "mime_type", Label: "MIME type", Required: true, Value: "image/jpeg"},
			{ID: "public_url", Name: "public_url", Label: "Public URL", Required: true},
		}
	case "menus":
		return []blocks.FieldData{
			{ID: "id", Name: "id", Label: "Menu ID", Required: true, Placeholder: "primary"},
			{ID: "name", Name: "name", Label: "Name", Required: true},
			{ID: "location", Name: "location", Label: "Location", Required: true, Value: "primary"},
		}
	case "users":
		return []blocks.FieldData{
			{ID: "id", Name: "id", Label: "User ID", Required: true},
			{ID: "login", Name: "login", Label: "Login", Required: true},
			{ID: "display_name", Name: "display_name", Label: "Display name", Required: true},
			{ID: "email", Name: "email", Label: "Email", Type: "email", Required: true},
		}
	default:
		return nil
	}
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
