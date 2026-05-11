package cms

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/fastygo/cms/internal/platform/cmspanel"
)

type publicFrontendValidationFixture struct {
	contracts []cmspanel.PublicProjectionContract
}

func newPublicFrontendValidationFixture() publicFrontendValidationFixture {
	return publicFrontendValidationFixture{contracts: cmspanel.PublicProjectionContracts()}
}

func TestCompanyThemePackageRendersAsCompiledPublicTheme(t *testing.T) {
	mux, closeFn := newPublicMux(t, "full", nil)
	defer closeFn()

	rec := requestPublic(mux, http.MethodGet, "/?preview_theme=company&preview_preset=company-bold-tech", "", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("company theme preview status = %d body = %s", rec.Code, rec.Body.String())
	}
	for _, expected := range []string{
		`data-gocms-theme="company"`,
		`data-gocms-theme-package="company"`,
		`data-gocms-public-header="company"`,
		`data-gocms-style-preset="company-bold-tech"`,
		`/static/themes/company/theme.css`,
		"Theme: company | Preset: company-bold-tech",
		"Published Post",
	} {
		if !strings.Contains(rec.Body.String(), expected) {
			t.Fatalf("expected company theme render to contain %q", expected)
		}
	}
}

func TestPublicFrontendRESTProjectionContracts(t *testing.T) {
	mux, closeFn := newPublicMux(t, "full", []string{"graphql"})
	defer closeFn()
	fixture := newPublicFrontendValidationFixture()

	for _, contract := range fixture.contracts {
		t.Run(contract.ID, func(t *testing.T) {
			rec := requestPublic(mux, http.MethodGet, contract.REST.CollectionPath, "", "")
			if rec.Code != http.StatusOK {
				t.Fatalf("%s status = %d body = %s", contract.REST.CollectionPath, rec.Code, rec.Body.String())
			}
			item := firstRESTDataItem(t, rec.Body.Bytes())
			assertFields(t, item, contract.REST.RequiredFields)
			if stringValue(item["status"]) == "draft" || stringValue(item["status"]) == "scheduled" {
				t.Fatalf("private content leaked through %s: %+v", contract.REST.CollectionPath, item)
			}

			if contract.REST.BySlugPath != "" {
				single := requestPublic(mux, http.MethodGet, contract.REST.BySlugPath, "", "")
				if single.Code != http.StatusOK {
					t.Fatalf("%s status = %d body = %s", contract.REST.BySlugPath, single.Code, single.Body.String())
				}
				assertFields(t, mustMapValue(t, restResourceData(t, single.Body.Bytes())), contract.REST.RequiredFields)
			}
		})
	}
}

func TestPublicFrontendGraphQLProjectionContracts(t *testing.T) {
	mux, closeFn := newPublicMux(t, "full", []string{"graphql"})
	defer closeFn()
	fixture := newPublicFrontendValidationFixture()

	for _, contract := range fixture.contracts {
		t.Run(contract.ID, func(t *testing.T) {
			payload, err := json.Marshal(map[string]any{"query": publicFrontendGraphQLQuery(contract)})
			if err != nil {
				t.Fatal(err)
			}
			rec := requestPublic(mux, http.MethodPost, "/go-graphql", string(payload), "")
			if rec.Code != http.StatusOK {
				t.Fatalf("graphql status = %d body = %s", rec.Code, rec.Body.String())
			}

			response := decodePublicGraphQLResponse(t, rec.Body.Bytes())
			if len(response.Errors) > 0 {
				t.Fatalf("unexpected graphql errors: %+v", response.Errors)
			}
			if response.Data == nil {
				t.Fatalf("graphql response has no data: %s", rec.Body.String())
			}
			var item map[string]any
			switch contract.ID {
			case "posts":
				item = firstGraphQLListItem(t, response.Data["posts"])
				assertFields(t, item, contract.GraphQL.RequiredFields)
				assertFields(t, mustMapValue(t, response.Data["post"]), contract.GraphQL.RequiredFields)
			case "pages":
				item = firstGraphQLListItem(t, response.Data["pages"])
				assertFields(t, item, contract.GraphQL.RequiredFields)
				assertFields(t, mustMapValue(t, response.Data["page"]), contract.GraphQL.RequiredFields)
			default:
				item = firstMapItem(t, response.Data[contract.GraphQL.CollectionField])
				assertFields(t, item, contract.GraphQL.RequiredFields)
			}
		})
	}
}

func publicFrontendGraphQLQuery(contract cmspanel.PublicProjectionContract) string {
	switch contract.ID {
	case "posts":
		return `
		query PublicFrontendPostsProjectionValidation {
			posts {
				items {
					id
					kind
					status
					slug
					title
					content
					excerpt
					authorID
					taxonomies {
						taxonomy
						termID
					}
					links
				}
			}
			post(slug: "published-post") {
				id
				kind
				status
				slug
				title
				content
				excerpt
				authorID
				taxonomies {
					taxonomy
					termID
				}
				links
			}
		}
	`
	case "pages":
		return `
		query PublicFrontendPagesProjectionValidation {
			pages {
				items {
					id
					kind
					status
					slug
					title
				}
			}
			page(slug: "about") {
				id
				kind
				status
				slug
				title
			}
		}
	`
	case "media":
		return `
		query PublicFrontendMediaProjectionValidation {
			media {
				id
				filename
				publicURL
			}
		}
	`
	case "taxonomies":
		return `
		query PublicFrontendTaxonomyProjectionValidation {
			taxonomies {
				type
				label
				public
				restVisible
				graphqlVisible
			}
		}
	`
	case "menus":
		return `
		query PublicFrontendMenuProjectionValidation {
			menus {
				id
				name
				location
				items {
					id
					label
					url
				}
			}
		}
	`
	default:
		return `query PublicFrontendUnknownProjectionValidation { __typename }`
	}
}

type publicGraphQLResponse struct {
	Data   map[string]any   `json:"data"`
	Errors []map[string]any `json:"errors"`
}

func decodePublicGraphQLResponse(t *testing.T, payload []byte) publicGraphQLResponse {
	t.Helper()
	var response publicGraphQLResponse
	if err := json.Unmarshal(payload, &response); err != nil {
		t.Fatal(err)
	}
	return response
}

func firstRESTDataItem(t *testing.T, payload []byte) map[string]any {
	t.Helper()
	data := restResourceData(t, payload)
	return firstMapItem(t, data)
}

func restResourceData(t *testing.T, payload []byte) any {
	t.Helper()
	var envelope map[string]any
	if err := json.Unmarshal(payload, &envelope); err != nil {
		t.Fatal(err)
	}
	data, ok := envelope["data"]
	if !ok {
		t.Fatalf("REST envelope does not contain data: %+v", envelope)
	}
	return data
}

func firstGraphQLListItem(t *testing.T, value any) map[string]any {
	t.Helper()
	list := mustMapValue(t, value)
	return firstMapItem(t, list["items"])
}

func firstMapItem(t *testing.T, value any) map[string]any {
	t.Helper()
	items, ok := value.([]any)
	if !ok {
		t.Fatalf("expected list, got %T: %+v", value, value)
	}
	if len(items) == 0 {
		t.Fatal("expected at least one public projection item")
	}
	return mustMapValue(t, items[0])
}

func mustMapValue(t *testing.T, value any) map[string]any {
	t.Helper()
	mapped, ok := value.(map[string]any)
	if !ok {
		t.Fatalf("expected object, got %T: %+v", value, value)
	}
	return mapped
}

func assertFields(t *testing.T, item map[string]any, fields []string) {
	t.Helper()
	for _, field := range fields {
		if _, ok := item[field]; !ok {
			t.Fatalf("expected projection field %q in %+v", field, item)
		}
	}
}

func stringValue(value any) string {
	if text, ok := value.(string); ok {
		return text
	}
	return ""
}
