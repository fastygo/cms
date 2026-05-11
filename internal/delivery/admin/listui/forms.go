package listui

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/fastygo/cms/internal/domain/authz"
	domaincontent "github.com/fastygo/cms/internal/domain/content"
	domainmedia "github.com/fastygo/cms/internal/domain/media"
	domainmenus "github.com/fastygo/cms/internal/domain/menus"
	domaintaxonomy "github.com/fastygo/cms/internal/domain/taxonomy"
	domainusers "github.com/fastygo/cms/internal/domain/users"
	"github.com/fastygo/cms/internal/site/adminfixtures"
	"github.com/fastygo/cms/internal/site/ui/blocks"
	"github.com/fastygo/panel"
	"github.com/fastygo/ui8kit/ui"
)

func BuildToolbarForm(bundle adminfixtures.AdminFixture, schema panel.TableSchema[authz.Capability], action string, state State) *blocks.InlineFormData {
	fields := []blocks.FieldData{
		{ID: "search", Name: "search", Label: bundle.Label("field_search", "Search"), Value: state.Search, Placeholder: bundle.Label("placeholder_search", "Search current records")},
	}
	for _, filter := range schema.Filters {
		id := "filter_" + filter.ID
		field := blocks.FieldData{ID: id, Name: id, Label: filter.Label, Value: state.Filters[filter.ID]}
		if filter.Type == panel.FilterSelect {
			field.Component = "select"
			field.Options = append([]ui.FieldOption{{Value: "", Label: bundle.Label("option_all", "All")}}, panelOptions(filter.Options)...)
		}
		fields = append(fields, field)
	}
	if options := SortOptions(schema); len(options) > 0 {
		fields = append(fields,
			blocks.FieldData{ID: "sort", Name: "sort", Label: bundle.Label("field_sort", "Sort by"), Value: state.Sort, Component: "select", Options: options},
			blocks.FieldData{ID: "order", Name: "order", Label: bundle.Label("field_order", "Order"), Value: state.Order, Component: "select", Options: []ui.FieldOption{{Value: "asc", Label: bundle.Label("sort_ascending", "Ascending")}, {Value: "desc", Label: bundle.Label("sort_descending", "Descending")}}},
		)
	}
	return &blocks.InlineFormData{
		Title:       bundle.Label("panel_list_controls", "List controls"),
		Description: bundle.Label("panel_list_controls_description", "Search, filter, and sort the current screen."),
		Action:      action,
		Method:      http.MethodGet,
		Fields:      fields,
		SubmitLabel: bundle.Label("action_apply", "Apply"),
	}
}

func BuildScreenOptionsForm(bundle adminfixtures.AdminFixture, screen string, schema panel.TableSchema[authz.Capability], state State, returnTo string, token string) *blocks.InlineFormData {
	if len(schema.PerPage) == 0 && len(schema.Columns) == 0 {
		return nil
	}
	fields := []blocks.FieldData{
		{ID: "return_to", Name: "return_to", Type: "hidden", Value: returnTo},
	}
	if len(schema.PerPage) > 0 {
		fields = append(fields, blocks.FieldData{
			ID:        "per_page",
			Name:      "per_page",
			Label:     bundle.Label("field_per_page", "Rows per page"),
			Value:     strconv.Itoa(state.PerPage),
			Component: "select",
			Options:   intOptions(schema.PerPage),
		})
	}
	toggleable := []string{}
	for _, column := range schema.Columns {
		if column.Toggleable {
			toggleable = append(toggleable, column.ID)
		}
	}
	if len(toggleable) > 0 {
		fields = append(fields, blocks.FieldData{
			ID:          "columns",
			Name:        "columns",
			Label:       bundle.Label("field_columns", "Visible columns"),
			Value:       strings.Join(state.VisibleColumns, ","),
			Placeholder: strings.Join(toggleable, ","),
			Hint:        bundle.Label("field_columns_hint", "Comma-separated column IDs from the current descriptor."),
		})
	}
	return &blocks.InlineFormData{
		Title:       bundle.Label("panel_screen_options", "Screen options"),
		Description: bundle.Label("panel_screen_options_description", "Persist pagination and column preferences for this screen."),
		Action:      "/go-admin/preferences/" + screen,
		Method:      http.MethodPost,
		Token:       token,
		Fields:      fields,
		SubmitLabel: bundle.Label("action_save", "Save"),
	}
}

func BuildBulkForm(bundle adminfixtures.AdminFixture, action string, token string, returnTo string, options []ui.FieldOption) *blocks.BulkActionData {
	if len(options) == 0 {
		return nil
	}
	return &blocks.BulkActionData{
		Action: action,
		Token:  token,
		Fields: []blocks.FieldData{
			{ID: "return_to", Name: "return_to", Type: "hidden", Value: returnTo},
			{ID: "bulk_action", Name: "bulk_action", Label: bundle.Label("field_bulk_action", "Bulk action"), Component: "select", Options: options},
		},
		SubmitLabel: bundle.Label("action_apply", "Apply"),
	}
}

func BuildContentQuickEditForm(bundle adminfixtures.AdminFixture, state State, basePath string, token string, entry domaincontent.Entry, statusOptions []ui.FieldOption) *blocks.InlineFormData {
	return &blocks.InlineFormData{
		Title:       bundle.Label("panel_quick_edit", "Quick edit"),
		Description: bundle.Label("panel_quick_edit_description", "Update the most common row fields without leaving the list."),
		Action:      basePath + "/quick-edit",
		Method:      http.MethodPost,
		Token:       token,
		Fields: []blocks.FieldData{
			{ID: "id", Name: "id", Type: "hidden", Value: string(entry.ID)},
			{ID: "return_to", Name: "return_to", Type: "hidden", Value: state.ReturnTo(basePath)},
			{ID: "title", Name: "title", Label: bundle.Label("field_title", "Title"), Value: entry.Title.Value("en", "en"), Required: true},
			{ID: "slug", Name: "slug", Label: bundle.Label("field_slug", "Slug"), Value: entry.Slug.Value("en", "en"), Required: true},
			{ID: "status", Name: "status", Label: bundle.Label("field_status", "Status"), Value: string(defaultStatus(entry.Status)), Component: "select", Options: statusOptions},
		},
		SubmitLabel: bundle.Label("action_save", "Save"),
		CancelLabel: bundle.Label("action_cancel", "Cancel"),
		CancelHref:  state.ReturnTo(basePath),
	}
}

func BuildTaxonomyQuickEditForm(bundle adminfixtures.AdminFixture, state State, basePath string, token string, item domaintaxonomy.Definition) *blocks.InlineFormData {
	return &blocks.InlineFormData{
		Title:       bundle.Label("panel_quick_edit", "Quick edit"),
		Description: bundle.Label("panel_quick_edit_description", "Update the most common row fields without leaving the list."),
		Action:      basePath,
		Method:      http.MethodPost,
		Token:       token,
		Fields: []blocks.FieldData{
			{ID: "type", Name: "type", Type: "hidden", Value: string(item.Type)},
			{ID: "return_to", Name: "return_to", Type: "hidden", Value: state.ReturnTo(basePath)},
			{ID: "label", Name: "label", Label: bundle.Label("field_label", "Label"), Value: item.Label, Required: true},
			{ID: "mode", Name: "mode", Label: bundle.Label("field_mode", "Mode"), Value: string(item.Mode), Component: "select", Options: []ui.FieldOption{{Value: "flat", Label: "Flat"}, {Value: "hierarchical", Label: "Hierarchical"}}},
		},
		SubmitLabel: bundle.Label("action_save", "Save"),
		CancelLabel: bundle.Label("action_cancel", "Cancel"),
		CancelHref:  state.ReturnTo(basePath),
	}
}

func BuildTermQuickEditForm(bundle adminfixtures.AdminFixture, state State, basePath string, token string, item domaintaxonomy.Term) *blocks.InlineFormData {
	return &blocks.InlineFormData{
		Title:       bundle.Label("panel_quick_edit", "Quick edit"),
		Description: bundle.Label("panel_quick_edit_description", "Update the most common row fields without leaving the list."),
		Action:      basePath,
		Method:      http.MethodPost,
		Token:       token,
		Fields: []blocks.FieldData{
			{ID: "id", Name: "id", Type: "hidden", Value: string(item.ID)},
			{ID: "return_to", Name: "return_to", Type: "hidden", Value: state.ReturnTo(basePath)},
			{ID: "name", Name: "name", Label: bundle.Label("field_name", "Name"), Value: item.Name.Value("en", "en"), Required: true},
			{ID: "slug", Name: "slug", Label: bundle.Label("field_slug", "Slug"), Value: item.Slug.Value("en", "en"), Required: true},
		},
		SubmitLabel: bundle.Label("action_save", "Save"),
		CancelLabel: bundle.Label("action_cancel", "Cancel"),
		CancelHref:  state.ReturnTo(basePath),
	}
}

func BuildMediaQuickEditForm(bundle adminfixtures.AdminFixture, state State, basePath string, token string, item domainmedia.Asset) *blocks.InlineFormData {
	return &blocks.InlineFormData{
		Title:       bundle.Label("panel_quick_edit", "Quick edit"),
		Description: bundle.Label("panel_quick_edit_description", "Update the most common row fields without leaving the list."),
		Action:      basePath,
		Method:      http.MethodPost,
		Token:       token,
		Fields: []blocks.FieldData{
			{ID: "id", Name: "id", Type: "hidden", Value: string(item.ID)},
			{ID: "return_to", Name: "return_to", Type: "hidden", Value: state.ReturnTo(basePath)},
			{ID: "filename", Name: "filename", Label: bundle.Label("field_filename", "Filename"), Value: item.Filename, Required: true},
			{ID: "mime_type", Name: "mime_type", Label: bundle.Label("field_mime_type", "MIME type"), Value: item.MimeType, Required: true},
			{ID: "public_url", Name: "public_url", Label: bundle.Label("field_public_url", "Public URL"), Value: item.PublicURL, Required: true},
			{ID: "alt_text", Name: "alt_text", Label: bundle.Label("field_alt_text", "Alt text"), Value: item.AltText},
			{ID: "caption", Name: "caption", Label: bundle.Label("field_caption", "Caption"), Value: item.Caption, Component: "textarea", Rows: 3},
			{ID: "provider", Name: "provider", Label: bundle.Label("field_provider", "Provider"), Value: item.ProviderRef.Provider},
			{ID: "provider_key", Name: "provider_key", Label: bundle.Label("field_provider_key", "Provider key"), Value: item.ProviderRef.Key},
			{ID: "provider_url", Name: "provider_url", Label: bundle.Label("field_provider_url", "Provider URL"), Value: item.ProviderRef.URL},
			{ID: "provider_checksum", Name: "provider_checksum", Label: bundle.Label("field_provider_checksum", "Checksum"), Value: item.ProviderRef.Checksum},
			{ID: "provider_etag", Name: "provider_etag", Label: bundle.Label("field_provider_etag", "ETag"), Value: item.ProviderRef.ETag},
		},
		SubmitLabel: bundle.Label("action_save", "Save"),
		CancelLabel: bundle.Label("action_cancel", "Cancel"),
		CancelHref:  state.ReturnTo(basePath),
	}
}

func BuildMenuQuickEditForm(bundle adminfixtures.AdminFixture, state State, basePath string, token string, item domainmenus.Menu) *blocks.InlineFormData {
	return &blocks.InlineFormData{
		Title:       bundle.Label("panel_quick_edit", "Quick edit"),
		Description: bundle.Label("panel_quick_edit_description", "Update the most common row fields without leaving the list."),
		Action:      basePath,
		Method:      http.MethodPost,
		Token:       token,
		Fields: []blocks.FieldData{
			{ID: "id", Name: "id", Type: "hidden", Value: string(item.ID)},
			{ID: "return_to", Name: "return_to", Type: "hidden", Value: state.ReturnTo(basePath)},
			{ID: "name", Name: "name", Label: bundle.Label("field_name", "Name"), Value: item.Name, Required: true},
			{ID: "location", Name: "location", Label: bundle.Label("field_location", "Location"), Value: string(item.Location), Required: true},
		},
		SubmitLabel: bundle.Label("action_save", "Save"),
		CancelLabel: bundle.Label("action_cancel", "Cancel"),
		CancelHref:  state.ReturnTo(basePath),
	}
}

func BuildUserQuickEditForm(bundle adminfixtures.AdminFixture, state State, basePath string, token string, item domainusers.User) *blocks.InlineFormData {
	return &blocks.InlineFormData{
		Title:       bundle.Label("panel_quick_edit", "Quick edit"),
		Description: bundle.Label("panel_quick_edit_description", "Update profile fields, rotate passwords, issue recovery material, or create app tokens."),
		Action:      basePath,
		Method:      http.MethodPost,
		Token:       token,
		Fields: []blocks.FieldData{
			{ID: "id", Name: "id", Type: "hidden", Value: string(item.ID)},
			{ID: "return_to", Name: "return_to", Type: "hidden", Value: state.ReturnTo(basePath)},
			{ID: "security_action", Name: "security_action", Label: bundle.Label("field_security_action", "Security action"), Value: "profile", Component: "select", Options: []ui.FieldOption{
				{Value: "profile", Label: "Update profile"},
				{Value: "set_password", Label: "Set local password"},
				{Value: "generate_recovery_codes", Label: "Generate recovery codes"},
				{Value: "issue_reset_token", Label: "Issue reset token"},
				{Value: "create_app_token", Label: "Create app token"},
				{Value: "revoke_app_token", Label: "Revoke app token"},
			}},
			{ID: "login", Name: "login", Label: bundle.Label("field_login", "Login"), Value: item.Login, Required: true},
			{ID: "display_name", Name: "display_name", Label: bundle.Label("field_display_name", "Display name"), Value: item.DisplayName, Required: true},
			{ID: "email", Name: "email", Label: bundle.Label("field_email", "Email"), Value: item.Email, Required: true},
			{ID: "status", Name: "status", Label: bundle.Label("field_status", "Status"), Value: string(item.Status), Component: "select", Options: []ui.FieldOption{{Value: "active", Label: "Active"}, {Value: "suspended", Label: "Suspended"}, {Value: "deleted", Label: "Deleted"}}},
			{ID: "new_password", Name: "new_password", Label: bundle.Label("field_new_password", "New password"), Type: "password"},
			{ID: "must_change_password", Name: "must_change_password", Label: bundle.Label("field_must_change_password", "Require password change on next sign-in"), Type: "checkbox", Value: "1"},
			{ID: "recovery_code_count", Name: "recovery_code_count", Label: bundle.Label("field_recovery_code_count", "Recovery codes"), Value: "8"},
			{ID: "reset_token_label", Name: "reset_token_label", Label: bundle.Label("field_reset_token_label", "Reset token label"), Value: "Admin reset"},
			{ID: "reset_token_ttl_minutes", Name: "reset_token_ttl_minutes", Label: bundle.Label("field_reset_token_ttl_minutes", "Reset token TTL (minutes)"), Value: "30"},
			{ID: "app_token_name", Name: "app_token_name", Label: bundle.Label("field_app_token_name", "App token name"), Value: "API client"},
			{ID: "app_token_ttl_hours", Name: "app_token_ttl_hours", Label: bundle.Label("field_app_token_ttl_hours", "App token TTL (hours)"), Value: "720"},
			{ID: "app_token_capabilities", Name: "app_token_capabilities", Label: bundle.Label("field_app_token_capabilities", "App token capabilities"), Placeholder: "content.read_private,media.edit"},
			{ID: "app_token_id", Name: "app_token_id", Label: bundle.Label("field_app_token_id", "App token ID/prefix to revoke")},
		},
		SubmitLabel: bundle.Label("action_save", "Save"),
		CancelLabel: bundle.Label("action_cancel", "Cancel"),
		CancelHref:  state.ReturnTo(basePath),
	}
}

func SortOptions(schema panel.TableSchema[authz.Capability]) []ui.FieldOption {
	options := []ui.FieldOption{}
	if strings.TrimSpace(schema.DefaultSort.ColumnID) != "" {
		options = append(options, ui.FieldOption{
			Value: schema.DefaultSort.ColumnID,
			Label: panelColumnLabel(schema, schema.DefaultSort.ColumnID, strings.Title(strings.ReplaceAll(schema.DefaultSort.ColumnID, "_", " "))),
		})
	}
	for _, column := range schema.Columns {
		if !column.Sortable {
			continue
		}
		if slicesContainOption(options, column.ID) {
			continue
		}
		options = append(options, ui.FieldOption{Value: column.ID, Label: fallbackValue(column.Label, column.ID)})
	}
	return options
}

func intOptions(values []int) []ui.FieldOption {
	result := make([]ui.FieldOption, 0, len(values))
	for _, value := range values {
		if value <= 0 {
			continue
		}
		result = append(result, ui.FieldOption{Value: strconv.Itoa(value), Label: strconv.Itoa(value)})
	}
	return result
}

func panelOptions(options []panel.Option) []ui.FieldOption {
	result := make([]ui.FieldOption, 0, len(options))
	for _, option := range options {
		result = append(result, ui.FieldOption{Value: option.Value, Label: option.Label})
	}
	return result
}

func panelColumnLabel(schema panel.TableSchema[authz.Capability], id string, fallback string) string {
	for _, column := range schema.Columns {
		if column.ID == id && strings.TrimSpace(column.Label) != "" {
			return column.Label
		}
	}
	return fallback
}

func fallbackValue(value string, fallback string) string {
	if strings.TrimSpace(value) != "" {
		return value
	}
	return fallback
}

func slicesContainOption(options []ui.FieldOption, value string) bool {
	for _, option := range options {
		if option.Value == value {
			return true
		}
	}
	return false
}

func defaultStatus(status domaincontent.Status) domaincontent.Status {
	if status != "" {
		return status
	}
	return domaincontent.StatusDraft
}
