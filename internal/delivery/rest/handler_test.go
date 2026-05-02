package rest_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fastygo/cms/internal/infra/features/cms"
)

func TestRESTDiscoveryPublicReadsAndWrites(t *testing.T) {
	module, err := cms.New("file:rest-test?mode=memory&cache=shared", "test-session-secret-test-session-secret", true)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := module.Close(t.Context()); err != nil {
			t.Fatal(err)
		}
	}()
	mux := http.NewServeMux()
	module.Routes(mux)

	rec := request(mux, http.MethodGet, "/go-json", "", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("discovery status = %d body = %s", rec.Code, rec.Body.String())
	}

	rec = request(mux, http.MethodGet, "/go-json/go/v2/posts?per_page=1", "", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("posts status = %d body = %s", rec.Code, rec.Body.String())
	}
	var list struct {
		Data       []map[string]any `json:"data"`
		Pagination struct {
			Page       int `json:"page"`
			PerPage    int `json:"per_page"`
			Total      int `json:"total"`
			TotalPages int `json:"total_pages"`
		} `json:"pagination"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &list); err != nil {
		t.Fatal(err)
	}
	if len(list.Data) != 1 || list.Pagination.PerPage != 1 || list.Pagination.Total != 1 {
		t.Fatalf("unexpected public list: %+v", list)
	}
	metadata, _ := list.Data[0]["metadata"].(map[string]any)
	if _, ok := metadata["private_key"]; ok {
		t.Fatalf("private metadata leaked: %+v", metadata)
	}

	rec = request(mux, http.MethodGet, "/go-json/go/v2/posts/by-slug/published-post", "", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("slug detail status = %d body = %s", rec.Code, rec.Body.String())
	}

	body := `{"title":{"en":"Created Through REST"},"slug":{"en":"created-through-rest"},"content":{"en":"Body"},"status":"published"}`
	rec = request(mux, http.MethodPost, "/go-json/go/v2/posts", body, "Bearer admin-token")
	if rec.Code != http.StatusCreated {
		t.Fatalf("create status = %d body = %s", rec.Code, rec.Body.String())
	}

	rec = request(mux, http.MethodPost, "/go-json/go/v2/posts", body, "Bearer viewer-token")
	if rec.Code != http.StatusForbidden {
		t.Fatalf("low privilege create status = %d body = %s", rec.Code, rec.Body.String())
	}

	taxonomyBody := `{"type":"genre","label":"Genres","mode":"flat","assigned_to_kinds":["post"],"public":true,"rest_visible":true}`
	rec = request(mux, http.MethodPost, "/go-json/go/v2/taxonomies", taxonomyBody, "Bearer admin-token")
	if rec.Code != http.StatusCreated {
		t.Fatalf("taxonomy create status = %d body = %s", rec.Code, rec.Body.String())
	}
	termBody := `{"id":"term-review","name":{"en":"Review"},"slug":{"en":"review"}}`
	rec = request(mux, http.MethodPost, "/go-json/go/v2/taxonomies/genre/terms", termBody, "Bearer admin-token")
	if rec.Code != http.StatusCreated {
		t.Fatalf("term create status = %d body = %s", rec.Code, rec.Body.String())
	}

	rec = request(mux, http.MethodGet, "/go-json/go/v2/settings", "", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("settings status = %d body = %s", rec.Code, rec.Body.String())
	}
	if bytes.Contains(rec.Body.Bytes(), []byte("site.private_note")) {
		t.Fatalf("private setting leaked: %s", rec.Body.String())
	}
}

func request(handler http.Handler, method string, path string, body string, authorization string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("User-Agent", "gocms-test")
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if authorization != "" {
		req.Header.Set("Authorization", authorization)
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}
