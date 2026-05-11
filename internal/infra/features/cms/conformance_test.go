package cms

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/fastygo/cms/internal/conformance"
	platformplugins "github.com/fastygo/cms/internal/platform/plugins"
)

func TestGoCMSConformanceBaselineLevelFull(t *testing.T) {
	mux, closeFn := newPublicMux(t, "full", []string{"graphql"})
	defer closeFn()
	filterDescriptor := moduleTestDescriptor{
		manifest: platformplugins.Manifest{
			ID:          "conformance-render-filter",
			Name:        "Conformance Render Filter",
			Version:     "1.0.0",
			Contract:    "0.1",
			Description: "Adds a conformance marker to public content.",
			Hooks: []platformplugins.HookRegistration{
				{HookID: "render.content.filter", HandlerID: "conformance.render", OwnerID: "conformance-render-filter", Category: platformplugins.HookCategoryFilter},
			},
		},
		register: func(_ context.Context, registry *platformplugins.Registry) error {
			hook := platformplugins.HookRegistration{HookID: "render.content.filter", HandlerID: "conformance.render", OwnerID: "conformance-render-filter", Category: platformplugins.HookCategoryFilter}
			registry.AddHooks(hook)
			registry.AddFilterHandlers(platformplugins.FilterHandlerRegistration{
				Hook: hook,
				Handle: func(_ context.Context, _ platformplugins.HookContext, value any) (any, error) {
					return `<span data-conformance-hook="render"></span>` + value.(string), nil
				},
			})
			return nil
		},
	}
	filterModule := newModuleForPublicTestsWithDescriptors(t, "full", []string{"conformance-render-filter"}, []platformplugins.Descriptor{filterDescriptor})
	t.Cleanup(func() {
		_ = filterModule.Close(t.Context())
	})
	filterMux := http.NewServeMux()
	filterModule.Routes(filterMux)

	runner := conformance.NewRunner(conformance.Options{
		ContractVersion: "go-codex.0.1",
		Implementation:  "GoCMS",
		Level:           conformance.LevelFull,
		Profiles:        []string{"public-rendering", "admin", "rest", "graphql", "compiled-plugins", "compiled-themes"},
		Now: func() time.Time {
			return time.Date(2026, 5, 9, 10, 0, 0, 0, time.UTC)
		},
	},
		conformance.Case{ID: "level0.core_public_visibility", Level: conformance.LevelCore, Run: func(context.Context) error {
			home := requestPublic(mux, http.MethodGet, "/", "", "")
			if err := expectStatus(home.Code, http.StatusOK, home.Body.String()); err != nil {
				return err
			}
			for _, leaked := range []string{"Draft Post", "Scheduled Post"} {
				if strings.Contains(home.Body.String(), leaked) {
					return fmt.Errorf("private content leaked on home page: %q", leaked)
				}
			}
			draft := requestPublic(mux, http.MethodGet, "/draft-post/", "", "")
			return expectStatus(draft.Code, http.StatusNotFound, draft.Body.String())
		}},
		conformance.Case{ID: "level1.rest_discovery_envelopes_and_errors", Level: conformance.LevelREST, Run: func(context.Context) error {
			discovery := requestPublic(mux, http.MethodGet, "/go-json", "", "")
			if err := expectStatus(discovery.Code, http.StatusOK, discovery.Body.String()); err != nil {
				return err
			}
			list := requestPublic(mux, http.MethodGet, "/go-json/go/v2/posts?per_page=1", "", "")
			if err := expectStatus(list.Code, http.StatusOK, list.Body.String()); err != nil {
				return err
			}
			var envelope struct {
				Data       []map[string]any `json:"data"`
				Pagination struct {
					Page       int `json:"page"`
					PerPage    int `json:"per_page"`
					Total      int `json:"total"`
					TotalPages int `json:"total_pages"`
				} `json:"pagination"`
			}
			if err := json.Unmarshal(list.Body.Bytes(), &envelope); err != nil {
				return err
			}
			if len(envelope.Data) != 1 || envelope.Pagination.Page != 1 || envelope.Pagination.PerPage != 1 || envelope.Pagination.Total == 0 {
				return fmt.Errorf("unexpected REST list envelope: %+v", envelope)
			}
			invalid := requestPublic(mux, http.MethodGet, "/go-json/go/v2/posts?page=bad", "", "")
			if err := expectStatus(invalid.Code, http.StatusBadRequest, invalid.Body.String()); err != nil {
				return err
			}
			return expectRESTErrorCode(invalid.Body.Bytes(), "validation_error")
		}},
		conformance.Case{ID: "level2.admin_auth_and_capabilities", Level: conformance.LevelAdmin, Run: func(context.Context) error {
			blocked := requestPublic(mux, http.MethodGet, "/go-admin", "", "")
			if err := expectStatus(blocked.Code, http.StatusSeeOther, blocked.Body.String()); err != nil {
				return err
			}
			viewerSettings := requestPublic(mux, http.MethodGet, "/go-admin/settings", "", "Bearer viewer-token")
			if err := expectStatus(viewerSettings.Code, http.StatusForbidden, viewerSettings.Body.String()); err != nil {
				return err
			}
			adminDashboard := requestPublic(mux, http.MethodGet, "/go-admin", "", "Bearer admin-token")
			return expectStatus(adminDashboard.Code, http.StatusOK, adminDashboard.Body.String())
		}},
		conformance.Case{ID: "level3.extension_graphql_plugin_route", Level: conformance.LevelExtension, Profiles: []string{"graphql"}, Run: func(context.Context) error {
			payload, err := json.Marshal(map[string]any{"query": `query { posts { pagination { total } } }`})
			if err != nil {
				return err
			}
			rec := requestPublic(mux, http.MethodPost, "/go-graphql", string(payload), "")
			return expectStatus(rec.Code, http.StatusOK, rec.Body.String())
		}},
		conformance.Case{ID: "level3.extension_render_filter_hook", Level: conformance.LevelExtension, Profiles: []string{"compiled-plugins"}, Run: func(context.Context) error {
			rec := requestPublic(filterMux, http.MethodGet, "/published-post/", "", "")
			if err := expectStatus(rec.Code, http.StatusOK, rec.Body.String()); err != nil {
				return err
			}
			if !strings.Contains(rec.Body.String(), `data-conformance-hook="render"`) {
				return fmt.Errorf("render filter marker missing")
			}
			for _, leaked := range []string{"Draft Post", "Scheduled Post"} {
				if strings.Contains(rec.Body.String(), leaked) {
					return fmt.Errorf("private content leaked through filter: %q", leaked)
				}
			}
			return nil
		}},
		conformance.Case{ID: "level4.full_theme_rendering", Level: conformance.LevelFull, Profiles: []string{"public-rendering"}, Run: func(context.Context) error {
			rec := requestPublic(mux, http.MethodGet, "/?preview_theme=company&preview_preset=company-bold-tech", "", "")
			if err := expectStatus(rec.Code, http.StatusOK, rec.Body.String()); err != nil {
				return err
			}
			for _, expected := range []string{`data-gocms-theme="company"`, `data-gocms-style-preset="company-bold-tech"`, "Published Post"} {
				if !strings.Contains(rec.Body.String(), expected) {
					return fmt.Errorf("theme render missing %q", expected)
				}
			}
			return nil
		}},
	)

	report := runner.Run(context.Background())
	if len(report.Failed) > 0 {
		t.Fatalf("conformance failures: %+v", report.Failed)
	}
	if len(report.Skipped) > 0 {
		t.Fatalf("unexpected skipped conformance cases: %+v", report.Skipped)
	}
	if len(report.Passed) != 6 {
		t.Fatalf("passed = %v, want all critical level cases", report.Passed)
	}
}

func expectStatus(got int, want int, body string) error {
	if got != want {
		return fmt.Errorf("status = %d, want %d: %s", got, want, body)
	}
	return nil
}

func expectRESTErrorCode(payload []byte, want string) error {
	var envelope struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(payload, &envelope); err != nil {
		return err
	}
	if envelope.Error.Code != want {
		return fmt.Errorf("REST error code = %q, want %q", envelope.Error.Code, want)
	}
	return nil
}
