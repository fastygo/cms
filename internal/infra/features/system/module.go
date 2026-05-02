package system

import (
	"context"
	"net/http"

	"github.com/fastygo/framework/pkg/app"
	"github.com/fastygo/framework/pkg/web"
)

// Module is the pass-0 feature proving host wiring without CMS business logic.
type Module struct{}

// New creates the system feature.
func New() Module {
	return Module{}
}

// ID returns the stable feature identifier.
func (Module) ID() string {
	return "system"
}

// Routes registers pass-0 system routes.
func (Module) Routes(mux *http.ServeMux) {
	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		_ = web.WriteJSON(w, http.StatusOK, map[string]string{
			"name":   "GoCMS",
			"status": "pass0",
		})
	})
}

// NavItems returns no navigation because the admin UI starts in a later pass.
func (Module) NavItems() []app.NavItem {
	return nil
}

// HealthCheck reports that the pass-0 feature is ready.
func (Module) HealthCheck(context.Context) error {
	return nil
}
