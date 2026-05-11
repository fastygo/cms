package cmspanel

import (
	"github.com/fastygo/cms/internal/domain/authz"
	domaincontent "github.com/fastygo/cms/internal/domain/content"
	"github.com/fastygo/panel"
)

type ContentResource struct {
	Kind domaincontent.Kind
	panel.Resource[authz.Capability]
}

func PostsResource() ContentResource {
	return contentResource(contentResourceOptions{
		id:       "posts",
		kind:     domaincontent.KindPost,
		singular: "Post",
		plural:   "Posts",
		prefix:   "/go-admin/posts",
		icon:     "file",
		order:    1,
	})
}

func PagesResource() ContentResource {
	return contentResource(contentResourceOptions{
		id:       "pages",
		kind:     domaincontent.KindPage,
		singular: "Page",
		plural:   "Pages",
		prefix:   "/go-admin/pages",
		icon:     "book",
		order:    2,
	})
}

func ContentResources() []ContentResource {
	return []ContentResource{PostsResource(), PagesResource()}
}

func ResourceByID(id string) (ContentResource, bool) {
	for _, resource := range ContentResources() {
		if string(resource.ID) == id {
			return resource, true
		}
	}
	return ContentResource{}, false
}

func ResourceByKind(kind domaincontent.Kind) (ContentResource, bool) {
	for _, resource := range ContentResources() {
		if resource.Kind == kind {
			return resource, true
		}
	}
	return ContentResource{}, false
}

type contentResourceOptions struct {
	id       string
	kind     domaincontent.Kind
	singular string
	plural   string
	prefix   string
	icon     string
	order    int
}

func contentResource(options contentResourceOptions) ContentResource {
	return ContentResource{
		Kind: options.kind,
		Resource: panel.Resource[authz.Capability]{
			ID:       panel.ResourceID(options.id),
			Label:    options.plural,
			Singular: options.singular,
			Plural:   options.plural,
			BasePath: options.prefix,
			Icon:     options.icon,
			Navigation: panel.MenuItem[authz.Capability]{
				ID:         options.id,
				Label:      options.plural,
				Path:       options.prefix,
				Icon:       options.icon,
				Order:      options.order,
				Capability: authz.CapabilityContentReadPrivate,
			},
			Capabilities: []panel.ResourceCapability[authz.Capability]{
				{Operation: panel.OperationList, Capability: authz.CapabilityContentReadPrivate},
				{Operation: panel.OperationCreate, Capability: authz.CapabilityContentCreate},
				{Operation: panel.OperationEdit, Capability: authz.CapabilityContentEdit},
				{Operation: panel.OperationDelete, Capability: authz.CapabilityContentDelete},
			},
			Routes: []panel.ResourceRoute[authz.Capability]{
				{Role: panel.RouteIndex, Pattern: "GET " + options.prefix, Capability: authz.CapabilityContentReadPrivate},
				{Role: panel.RouteNew, Pattern: "GET " + options.prefix + "/new", Capability: authz.CapabilityContentCreate},
				{Role: panel.RouteCreate, Pattern: "POST " + options.prefix, Capability: authz.CapabilityContentCreate},
				{Role: panel.RouteEdit, Pattern: "GET " + options.prefix + "/{id}/edit", Capability: authz.CapabilityContentEdit},
				{Role: panel.RouteUpdate, Pattern: "POST " + options.prefix + "/{id}", Capability: authz.CapabilityContentEdit},
				{Role: panel.RouteDelete, Pattern: "POST " + options.prefix + "/{id}/trash", Capability: authz.CapabilityContentDelete},
			},
			Table: panel.TableSchema[authz.Capability]{
				ID: "content-list",
				Columns: []panel.Column{
					{ID: "title", Label: "Title", Type: panel.ColumnText, Searchable: true, Sortable: true, Toggleable: true},
					{ID: "slug", Label: "Slug", Type: panel.ColumnText, Searchable: true, Sortable: true, Toggleable: true},
					{ID: "status", Label: "Status", Type: panel.ColumnBadge, Toggleable: true},
					{ID: "author", Label: "Author", Type: panel.ColumnText, Searchable: true, Toggleable: true},
				},
				Filters: []panel.Filter{
					{
						ID:      "status",
						Label:   "Status",
						Type:    panel.FilterSelect,
						Options: []panel.Option{{Value: "draft", Label: "Draft"}, {Value: "published", Label: "Published"}, {Value: "scheduled", Label: "Scheduled"}, {Value: "trashed", Label: "Trashed"}},
					},
					{ID: "author", Label: "Author", Type: panel.FilterText},
				},
				DefaultSort: panel.Sort{ColumnID: "updated_at", Direction: panel.SortDesc},
				Searchable:  true,
				PerPage:     []int{25, 50, 100},
				RowActions: []panel.Action[authz.Capability]{
					{ID: "edit", Label: "Edit", Placement: panel.ActionRow, Style: panel.ActionLink, Capability: authz.CapabilityContentEdit},
				},
				BulkActions: []panel.Action[authz.Capability]{
					{ID: "publish", Label: "Publish", Placement: panel.ActionBulk, Style: panel.ActionButton, Capability: authz.CapabilityContentPublish},
					{ID: "trash", Label: "Trash", Placement: panel.ActionBulk, Style: panel.ActionButton, Capability: authz.CapabilityContentDelete},
					{ID: "restore", Label: "Restore", Placement: panel.ActionBulk, Style: panel.ActionButton, Capability: authz.CapabilityContentRestore},
				},
			},
			Form: panel.FormSchema{
				ID:        "content-editor",
				Operation: "write",
				Fields: []panel.Field{
					{ID: "title", Label: "Title", Type: panel.FieldText, Required: true},
					{ID: "slug", Label: "Slug", Type: panel.FieldText, Required: true},
					{ID: "content", Label: "Content", Type: panel.FieldRichText, Description: "HTML is stored in the content field; the editor provider can be swapped later."},
					{ID: "excerpt", Label: "Excerpt", Type: panel.FieldTextarea},
					{ID: "author_id", Label: "Author ID", Type: panel.FieldText},
					{ID: "featured_media_id", Label: "Featured media ID", Type: panel.FieldText},
					{ID: "template", Label: "Template", Type: panel.FieldText},
					{ID: "terms", Label: "Taxonomy terms", Type: panel.FieldText, Placeholder: "category:news,tag:go"},
				},
			},
			Actions: []panel.Action[authz.Capability]{
				{ID: "create", Label: "Create", Placement: panel.ActionHeader, Style: panel.ActionButton, Capability: authz.CapabilityContentCreate, URL: options.prefix + "/new"},
			},
		},
	}
}
