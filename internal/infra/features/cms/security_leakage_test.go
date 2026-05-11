package cms

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestCompatibilitySecurityNoDraftsOrPrivateDataLeakAcrossPublicSurfaces(t *testing.T) {
	mux, closeFn := newPublicMux(t, "full", []string{"graphql"})
	defer closeFn()

	checks := []struct {
		name string
		rec  func(*testing.T) string
	}{
		{
			name: "public home",
			rec: func(t *testing.T) string {
				t.Helper()
				rec := requestPublic(mux, http.MethodGet, "/", "", "")
				if err := expectStatus(rec.Code, http.StatusOK, rec.Body.String()); err != nil {
					t.Fatal(err)
				}
				return rec.Body.String()
			},
		},
		{
			name: "REST posts",
			rec: func(t *testing.T) string {
				t.Helper()
				rec := requestPublic(mux, http.MethodGet, "/go-json/go/v2/posts?per_page=20", "", "")
				if err := expectStatus(rec.Code, http.StatusOK, rec.Body.String()); err != nil {
					t.Fatal(err)
				}
				return rec.Body.String()
			},
		},
		{
			name: "REST search",
			rec: func(t *testing.T) string {
				t.Helper()
				rec := requestPublic(mux, http.MethodGet, "/go-json/go/v2/search?q=Post", "", "")
				if err := expectStatus(rec.Code, http.StatusOK, rec.Body.String()); err != nil {
					t.Fatal(err)
				}
				return rec.Body.String()
			},
		},
		{
			name: "GraphQL posts",
			rec: func(t *testing.T) string {
				t.Helper()
				payload, err := json.Marshal(map[string]any{"query": `query { posts { items { id status metadata title } } }`})
				if err != nil {
					t.Fatal(err)
				}
				rec := requestPublic(mux, http.MethodPost, "/go-graphql", string(payload), "")
				if err := expectStatus(rec.Code, http.StatusOK, rec.Body.String()); err != nil {
					t.Fatal(err)
				}
				return rec.Body.String()
			},
		},
	}

	for _, check := range checks {
		t.Run(check.name, func(t *testing.T) {
			body := check.rec(t)
			for _, forbidden := range []string{
				"content-post-draft",
				"Draft Post",
				"content-post-scheduled",
				"Scheduled Post",
				"private_key",
				"site.private_note",
			} {
				if strings.Contains(body, forbidden) {
					t.Fatalf("%s leaked %q: %s", check.name, forbidden, body)
				}
			}
		})
	}

	admin := requestPublic(mux, http.MethodGet, "/go-admin/posts", "", "")
	if admin.Code != http.StatusSeeOther {
		t.Fatalf("anonymous admin list status = %d, want redirect", admin.Code)
	}
}
