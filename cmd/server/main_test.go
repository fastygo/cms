package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	platformconfig "github.com/fastygo/cms/internal/platform/config"
	"github.com/fastygo/cms/internal/platform/logging"
)

func TestBuildAppServesHealthAndSystemRoutes(t *testing.T) {
	t.Setenv("APP_DATA_SOURCE", "file:server-test?mode=memory&cache=shared")
	cfg, err := platformconfig.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	app, err := buildApp(cfg, logging.New(io.Discard, "error", "text"))
	if err != nil {
		t.Fatalf("buildApp() error = %v", err)
	}

	tests := []struct {
		name string
		path string
	}{
		{name: "liveness", path: cfg.Framework.HealthLivePath},
		{name: "readiness", path: cfg.Framework.HealthReadyPath},
		{name: "system", path: "/"},
		{name: "admin-css", path: "/static/css/app.css"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			req.Header.Set("User-Agent", "gocms-test")
			rec := httptest.NewRecorder()

			app.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusOK, rec.Body.String())
			}
		})
	}
}
