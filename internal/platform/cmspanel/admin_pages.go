package cmspanel

import (
	"github.com/fastygo/cms/internal/domain/authz"
	"github.com/fastygo/panel"
)

type AdminRouteRole string

const (
	AdminRouteIndex  AdminRouteRole = "index"
	AdminRouteCreate AdminRouteRole = "create"
	AdminRouteUpdate AdminRouteRole = "update"
)

type AdminPageRoute struct {
	Role       AdminRouteRole
	Pattern    string
	Capability authz.Capability
}

type AdminPage struct {
	panel.Page[authz.Capability]
	Routes []AdminPageRoute
}

func AdminPages() []AdminPage {
	return []AdminPage{
		DashboardPage(),
		ContentTypesPage(),
		TaxonomiesPage(),
		TermsPage(),
		MediaPage(),
		MenusPage(),
		UsersPage(),
		AuthorsPage(),
		CapabilitiesPage(),
		SettingsPage(),
		ThemesPage(),
		PermalinksPage(),
		HeadlessPage(),
		RuntimePage(),
	}
}

func DashboardPage() AdminPage {
	return adminPage(adminPageOptions{
		id: "dashboard", kind: panel.PageDashboard, title: "Dashboard", path: "/go-admin", icon: "home", order: 0,
	})
}

func ContentTypesPage() AdminPage {
	return adminPage(adminPageOptions{
		id: "content-types", kind: panel.PageSettings, title: "Content types", path: "/go-admin/content-types", icon: "box", order: 3,
		capability: authz.CapabilitySettingsManage,
		table:      simpleTable("content-types", "ID", "Label", "REST"),
		form:       form("content-type", textField("id", "ID", true), textField("label", "Label", true)),
		routes: []AdminPageRoute{
			{Role: AdminRouteIndex, Pattern: "GET /go-admin/content-types", Capability: authz.CapabilitySettingsManage},
			{Role: AdminRouteCreate, Pattern: "POST /go-admin/content-types", Capability: authz.CapabilitySettingsManage},
		},
	})
}

func TaxonomiesPage() AdminPage {
	return adminPage(adminPageOptions{
		id: "taxonomies", kind: panel.PageSettings, title: "Taxonomies", path: "/go-admin/taxonomies", icon: "boxes", order: 4,
		capability: authz.CapabilityTaxonomiesManage,
		table:      simpleTable("taxonomies", "Type", "Label", "Mode"),
		form: form("taxonomy", textField("type", "Type", true), textField("label", "Label", true), selectField("mode", "Mode", false,
			panel.Option{Value: "flat", Label: "Flat"},
			panel.Option{Value: "hierarchical", Label: "Hierarchical"},
		)),
		routes: []AdminPageRoute{
			{Role: AdminRouteIndex, Pattern: "GET /go-admin/taxonomies", Capability: authz.CapabilityTaxonomiesManage},
			{Role: AdminRouteCreate, Pattern: "POST /go-admin/taxonomies", Capability: authz.CapabilityTaxonomiesManage},
		},
	})
}

func TermsPage() AdminPage {
	return adminPage(adminPageOptions{
		id: "terms", kind: panel.PageSettings, title: "Terms", path: "/go-admin/taxonomies/{type}/terms",
		capability: authz.CapabilityTaxonomiesManage,
		table:      simpleTable("terms", "ID", "Name", "Taxonomy"),
		form:       form("term", textField("id", "ID", true), textField("name", "Name", true), textField("slug", "Slug", true)),
		routes: []AdminPageRoute{
			{Role: AdminRouteIndex, Pattern: "GET /go-admin/taxonomies/{type}/terms", Capability: authz.CapabilityTaxonomiesManage},
			{Role: AdminRouteCreate, Pattern: "POST /go-admin/taxonomies/{type}/terms", Capability: authz.CapabilityTaxonomiesManage},
		},
	})
}

func MediaPage() AdminPage {
	return adminPage(adminPageOptions{
		id: "media", kind: panel.PageCustom, title: "Media", path: "/go-admin/media", icon: "image", order: 5,
		capability: authz.CapabilityMediaUpload,
		table:      simpleTable("media", "ID", "Filename", "MIME type"),
		form:       form("media", textField("id", "ID", true), textField("filename", "Filename", true), textField("mime_type", "MIME type", true), textField("public_url", "Public URL", true)),
		routes: []AdminPageRoute{
			{Role: AdminRouteIndex, Pattern: "GET /go-admin/media", Capability: authz.CapabilityMediaUpload},
			{Role: AdminRouteCreate, Pattern: "POST /go-admin/media", Capability: authz.CapabilityMediaUpload},
		},
	})
}

func MenusPage() AdminPage {
	return adminPage(adminPageOptions{
		id: "menus", kind: panel.PageCustom, title: "Menus", path: "/go-admin/menus", icon: "menu", order: 6,
		capability: authz.CapabilityMenusManage,
		table:      simpleTable("menus", "ID", "Name", "Location"),
		form:       form("menu", textField("id", "ID", true), textField("name", "Name", true), textField("location", "Location", true)),
		routes: []AdminPageRoute{
			{Role: AdminRouteIndex, Pattern: "GET /go-admin/menus", Capability: authz.CapabilityMenusManage},
			{Role: AdminRouteCreate, Pattern: "POST /go-admin/menus", Capability: authz.CapabilityMenusManage},
		},
	})
}

func UsersPage() AdminPage {
	return adminPage(adminPageOptions{
		id: "users", kind: panel.PageCustom, title: "Users", path: "/go-admin/users", icon: "users", order: 7,
		capability: authz.CapabilityUsersManage,
		table:      simpleTable("users", "ID", "Display name", "Status"),
		form:       form("user", textField("id", "ID", true), textField("login", "Login", true), textField("display_name", "Display name", true), textField("email", "Email", true)),
		routes: []AdminPageRoute{
			{Role: AdminRouteIndex, Pattern: "GET /go-admin/users", Capability: authz.CapabilityUsersManage},
			{Role: AdminRouteCreate, Pattern: "POST /go-admin/users", Capability: authz.CapabilityUsersManage},
		},
	})
}

func AuthorsPage() AdminPage {
	return adminPage(adminPageOptions{
		id: "authors", kind: panel.PageReport, title: "Authors", path: "/go-admin/authors", icon: "users", order: 8,
		capability: authz.CapabilityContentReadPrivate,
		table:      simpleTable("authors", "ID", "Display name", "Slug"),
		routes:     []AdminPageRoute{{Role: AdminRouteIndex, Pattern: "GET /go-admin/authors", Capability: authz.CapabilityContentReadPrivate}},
	})
}

func CapabilitiesPage() AdminPage {
	return adminPage(adminPageOptions{
		id: "capabilities", kind: panel.PageReport, title: "Roles and capabilities", path: "/go-admin/capabilities", icon: "shield", order: 9,
		capability: authz.CapabilityRolesManage,
		table:      simpleTable("capabilities", "Capability", "Description", "Scope"),
		routes:     []AdminPageRoute{{Role: AdminRouteIndex, Pattern: "GET /go-admin/capabilities", Capability: authz.CapabilityRolesManage}},
	})
}

func SettingsPage() AdminPage {
	return adminPage(adminPageOptions{
		id: "settings", kind: panel.PageSettings, title: "Settings", path: "/go-admin/settings", icon: "sliders", order: 10,
		capability: authz.CapabilitySettingsManage,
		form:       form("settings", textField("site_title", "Site title", true), textField("public_rendering", "Public rendering", false)),
		routes: []AdminPageRoute{
			{Role: AdminRouteIndex, Pattern: "GET /go-admin/settings", Capability: authz.CapabilitySettingsManage},
			{Role: AdminRouteUpdate, Pattern: "POST /go-admin/settings", Capability: authz.CapabilitySettingsManage},
		},
	})
}

func ThemesPage() AdminPage {
	return adminPage(adminPageOptions{
		id: "themes", kind: panel.PageSettings, title: "Themes", path: "/go-admin/themes", icon: "palette", order: 11,
		capability: authz.CapabilityThemesManage,
		table:      simpleTable("themes", "Theme", "Contract", "Status"),
		form:       form("themes", textField("theme_active", "Active theme", true), textField("theme_style_preset", "Style preset", true), textField("theme_preview", "Preview theme", false), textField("theme_preview_preset", "Preview preset", false)),
		routes: []AdminPageRoute{
			{Role: AdminRouteIndex, Pattern: "GET /go-admin/themes", Capability: authz.CapabilityThemesManage},
			{Role: AdminRouteUpdate, Pattern: "POST /go-admin/themes", Capability: authz.CapabilityThemesManage},
		},
	})
}

func PermalinksPage() AdminPage {
	return adminPage(adminPageOptions{
		id: "permalinks", kind: panel.PageSettings, title: "Permalinks", path: "/go-admin/permalinks", icon: "link", order: 12,
		capability: authz.CapabilitySettingsManage,
		form:       form("permalinks", textField("post_pattern", "Post permalink pattern", true), textField("page_pattern", "Page permalink pattern", true)),
		routes: []AdminPageRoute{
			{Role: AdminRouteIndex, Pattern: "GET /go-admin/permalinks", Capability: authz.CapabilitySettingsManage},
			{Role: AdminRouteUpdate, Pattern: "POST /go-admin/permalinks", Capability: authz.CapabilitySettingsManage},
		},
	})
}

func HeadlessPage() AdminPage {
	return adminPage(adminPageOptions{
		id: "headless", kind: panel.PageReport, title: "API and headless settings", path: "/go-admin/headless", icon: "server", order: 13,
		capability: authz.CapabilitySettingsManage,
		table:      simpleTable("headless", "Surface", "Description", "Status"),
		routes:     []AdminPageRoute{{Role: AdminRouteIndex, Pattern: "GET /go-admin/headless", Capability: authz.CapabilitySettingsManage}},
	})
}

func RuntimePage() AdminPage {
	return adminPage(adminPageOptions{
		id: "runtime", kind: panel.PageRuntime, title: "Runtime status", path: "/go-admin/runtime", icon: "server", order: 14,
		capability: authz.CapabilitySettingsManage,
		table:      simpleTable("runtime", "Item", "Value", "Status"),
		routes:     []AdminPageRoute{{Role: AdminRouteIndex, Pattern: "GET /go-admin/runtime", Capability: authz.CapabilitySettingsManage}},
	})
}

type adminPageOptions struct {
	id         string
	kind       panel.PageKind
	title      string
	path       string
	icon       string
	order      int
	capability authz.Capability
	table      panel.TableSchema[authz.Capability]
	form       panel.FormSchema
	routes     []AdminPageRoute
}

func adminPage(options adminPageOptions) AdminPage {
	if len(options.routes) == 0 {
		options.routes = []AdminPageRoute{{Role: AdminRouteIndex, Pattern: "GET " + options.path, Capability: options.capability}}
	}
	navigation := panel.MenuItem[authz.Capability]{}
	if options.icon != "" {
		navigation = panel.MenuItem[authz.Capability]{
			ID:         options.id,
			Label:      options.title,
			Path:       options.path,
			Icon:       options.icon,
			Order:      options.order,
			Capability: options.capability,
		}
	}
	return AdminPage{
		Page: panel.Page[authz.Capability]{
			ID:         panel.PageID(options.id),
			Kind:       options.kind,
			Title:      options.title,
			Path:       options.path,
			Icon:       options.icon,
			Capability: options.capability,
			Navigation: navigation,
			Table:      options.table,
			Form:       options.form,
		},
		Routes: options.routes,
	}
}

func simpleTable(id string, nameLabel string, descriptionLabel string, statusLabel string) panel.TableSchema[authz.Capability] {
	return panel.TableSchema[authz.Capability]{
		ID: id + "-list",
		Columns: []panel.Column{
			{ID: "name", Label: nameLabel, Type: panel.ColumnText, Searchable: true, Sortable: true},
			{ID: "description", Label: descriptionLabel, Type: panel.ColumnText, Searchable: true},
			{ID: "status", Label: statusLabel, Type: panel.ColumnBadge, Sortable: true},
		},
		Searchable: true,
	}
}

func form(id string, fields ...panel.Field) panel.FormSchema {
	return panel.FormSchema{ID: id + "-form", Operation: "write", Fields: fields}
}

func textField(id string, label string, required bool) panel.Field {
	return panel.Field{ID: id, Label: label, Type: panel.FieldText, Required: required}
}

func selectField(id string, label string, required bool, options ...panel.Option) panel.Field {
	return panel.Field{ID: id, Label: label, Type: panel.FieldSelect, Required: required, Options: options}
}
