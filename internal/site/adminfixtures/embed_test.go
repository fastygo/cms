package adminfixtures

import (
	"strings"
	"testing"

	"github.com/fastygo/cms/internal/platform/locales"
)

func TestLoadAdminFixtures(t *testing.T) {
	for _, loc := range Locales {
		t.Run(loc, func(t *testing.T) {
			bundle, err := Load(loc)
			if err != nil {
				t.Fatalf("Load(%q): %v", loc, err)
			}
			if bundle.Admin.Meta.Title == "" {
				t.Fatalf("admin.meta.title is empty for locale %q", loc)
			}
			if len(bundle.Admin.Navigation) == 0 {
				t.Fatalf("admin.navigation is empty for locale %q", loc)
			}
			if len(bundle.Admin.Dashboard.Cards) == 0 {
				t.Fatalf("admin.dashboard.cards is empty for locale %q", loc)
			}
			if _, ok := bundle.Admin.Screen("dashboard"); !ok {
				t.Fatalf("admin.screens.dashboard missing for locale %q", loc)
			}
		})
	}
}

func TestMustLoadFallsBackToDefaultLocale(t *testing.T) {
	fixture := MustLoad("does-not-exist")
	if fixture.Meta.Title == "" {
		t.Fatalf("fallback fixture meta title is empty")
	}
	if fixture.Meta.Lang != locales.Default {
		t.Fatalf("fallback fixture lang = %q, want %q", fixture.Meta.Lang, locales.Default)
	}
}

func TestAdminLabelKeysMatchAcrossLocales(t *testing.T) {
	en, err := Load("en")
	if err != nil {
		t.Fatalf("Load(en): %v", err)
	}
	ru, err := Load("ru")
	if err != nil {
		t.Fatalf("Load(ru): %v", err)
	}
	enLabels := en.Admin.Labels
	ruLabels := ru.Admin.Labels
	if len(enLabels) != len(ruLabels) {
		t.Fatalf("labels count mismatch en=%d ru=%d", len(enLabels), len(ruLabels))
	}
	for k := range enLabels {
		if _, ok := ruLabels[k]; !ok {
			t.Fatalf("label key %q present in en but missing in ru", k)
		}
		if strings.TrimSpace(ruLabels[k]) == "" {
			t.Fatalf("label key %q is empty in ru", k)
		}
	}
	for k := range ruLabels {
		if _, ok := enLabels[k]; !ok {
			t.Fatalf("label key %q present in ru but missing in en", k)
		}
		if strings.TrimSpace(enLabels[k]) == "" {
			t.Fatalf("label key %q is empty in en", k)
		}
	}
}

func TestAdminLabelKeysCoverKnownCodeKeys(t *testing.T) {
	// Static inventory of keys referenced from admin delivery + listui (keep in sync when adding UI).
	required := []string{
		"action_apply", "action_back", "action_cancel", "action_content_create", "action_content_update",
		"action_edit", "action_new", "action_next", "action_open", "action_previous", "action_quick_edit",
		"action_save", "action_sign_in", "action_sign_out",
		"default_user_app_token_name", "default_user_reset_token_label",
		"editor_provider_tiptap_basic_description", "editor_provider_tiptap_basic_label",
		"error_app_token_inherited_capabilities", "error_login_locked", "error_new_password_required",
		"error_unsupported_security_action", "error_user_not_found",
		"field_app_token_capabilities", "field_app_token_id", "field_app_token_name", "field_app_token_ttl_hours",
		"field_avatar_media_id", "field_avatar_media_id_placeholder", "field_bulk_action", "field_columns",
		"field_columns_hint", "field_display_name", "field_email", "field_login", "field_mode", "field_must_change_password",
		"field_new_password", "field_password", "field_per_page", "field_public_rendering", "field_recovery_code_count",
		"field_reset_token_label", "field_reset_token_ttl_minutes", "field_search", "field_security_action",
		"field_site_title", "field_slug", "field_sort", "field_status", "field_title", "field_order",
		"field_theme_active", "field_theme_preview", "field_theme_preview_preset", "field_theme_style_preset",
		"field_id", "field_filename", "field_mime_type", "field_size_bytes", "field_width", "field_height",
		"field_public_url", "field_alt_text", "field_caption", "field_provider", "field_provider_key",
		"field_provider_url", "field_provider_checksum", "field_provider_etag",
		"headless_graphql", "headless_graphql_description", "headless_graphql_enabled_description",
		"headless_rendering", "headless_rendering_description", "headless_rest", "headless_rest_description",
		"language_switch_aria_label", "mode_flat", "mode_hierarchical", "option_all", "panel_list_controls",
		"panel_list_controls_description", "panel_publish", "panel_quick_edit", "panel_quick_edit_description",
		"panel_quick_edit_users_description", "panel_screen_options", "panel_screen_options_description",
		"placeholder_search", "runtime_active_plugins", "runtime_admin_policy", "runtime_browser_stateless",
		"runtime_content_provider", "runtime_deployment", "runtime_dev_bearer", "runtime_login_policy",
		"runtime_playground_auth", "runtime_preset", "runtime_profile", "runtime_provider_switch_description",
		"runtime_provider_switch_rule", "runtime_row_audit", "runtime_row_error", "runtime_row_health",
		"runtime_site_package", "runtime_storage",
		"security_action_create_app_token", "security_action_generate_recovery_codes", "security_action_issue_reset_token",
		"security_action_profile", "security_action_revoke_app_token", "security_action_set_password",
		"security_app_token_description", "security_app_token_revoked_description", "security_app_token_revoked_title",
		"security_app_token_scope_label", "security_app_token_title", "security_expires_label",
		"security_password_updated_description", "security_password_updated_row_description", "security_password_updated_title",
		"security_recovery_code_row_label", "security_recovery_codes_description", "security_recovery_codes_title",
		"security_reset_token_description", "security_reset_token_title", "security_token_label",
		"sort_ascending", "sort_descending", "status_draft", "status_published", "status_scheduled", "status_trashed",
		"status_user_active", "status_user_deleted", "status_user_suspended",
		"table_actions", "table_author", "table_description", "table_name", "table_slug", "table_status", "table_title",
	}
	en := MustLoad("en")
	for _, k := range required {
		if v := en.Label(k, ""); v == "" {
			t.Fatalf("required label key %q missing or empty in en bundle", k)
		}
	}
}
