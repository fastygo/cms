package graphqlplugin

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	graphql "github.com/graph-gophers/graphql-go"

	"github.com/fastygo/cms/internal/delivery/rest"
	"github.com/fastygo/cms/internal/domain/authz"
	"github.com/fastygo/framework/pkg/web"
)

type Handler struct {
	schema   *graphql.Schema
	auth     rest.Authenticator
	settings Settings
}

type requestPayload struct {
	Query         string         `json:"query"`
	OperationName string         `json:"operationName"`
	Variables     map[string]any `json:"variables"`
}

type requestState struct {
	principal     authz.Principal
	authenticated bool
}

type requestStateKey struct{}

func NewHandler(root *rootResolver, authenticator rest.Authenticator, settings Settings) (Handler, error) {
	handler := Handler{auth: authenticator, settings: settings}
	schema, err := graphql.ParseSchema(schema, root,
		graphql.MaxDepth(settings.MaxDepth),
		graphql.MaxParallelism(settings.MaxParallelism),
		graphql.MaxQueryLength(settings.MaxQueryLength),
		graphql.RestrictIntrospection(func(ctx context.Context) bool {
			state := stateFromContext(ctx)
			return settings.PublicIntrospection ||
				(state.authenticated && (state.principal.Has(authz.CapabilitySettingsManage) || state.principal.Has(authz.CapabilityPrivateAPIRead)))
		}),
	)
	if err != nil {
		return Handler{}, err
	}
	handler.schema = schema
	return handler, nil
}

func (h Handler) Get(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodAllowed(w, http.MethodGet)
		return
	}
	payload := requestPayload{
		Query:         r.URL.Query().Get("query"),
		OperationName: r.URL.Query().Get("operationName"),
	}
	if payload.Query == "" {
		h.writeBadRequest(w, "GraphQL query is required for GET requests.")
		return
	}
	if raw := strings.TrimSpace(r.URL.Query().Get("variables")); raw != "" {
		if err := json.Unmarshal([]byte(raw), &payload.Variables); err != nil {
			h.writeBadRequest(w, "Invalid GraphQL variables payload.")
			return
		}
	}
	h.execute(w, r, payload)
}

func (h Handler) Post(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodAllowed(w, http.MethodPost)
		return
	}
	var payload requestPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		h.writeBadRequest(w, "Invalid GraphQL request body.")
		return
	}
	if strings.TrimSpace(payload.Query) == "" {
		h.writeBadRequest(w, "GraphQL query is required.")
		return
	}
	h.execute(w, r, payload)
}

func (h Handler) Options(w http.ResponseWriter, r *http.Request) {
	h.applyHeaders(w, r)
	w.Header().Set("Allow", strings.Join([]string{http.MethodGet, http.MethodPost, http.MethodOptions}, ", "))
	w.WriteHeader(http.StatusNoContent)
}

func (h Handler) Status(w http.ResponseWriter, r *http.Request, _ authz.Principal) {
	h.applyHeaders(w, r)
	_ = web.WriteJSON(w, http.StatusOK, pWrap(h))
}

func (h Handler) execute(w http.ResponseWriter, r *http.Request, payload requestPayload) {
	h.applyHeaders(w, r)
	state := requestState{}
	if principal, ok := h.auth.Principal(r); ok {
		state.principal = principal
		state.authenticated = true
	}
	ctx := context.WithValue(r.Context(), requestStateKey{}, state)
	response := h.schema.Exec(ctx, payload.Query, payload.OperationName, payload.Variables)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

func (h Handler) applyHeaders(w http.ResponseWriter, r *http.Request) {
	if policy := strings.TrimSpace(h.settings.CachePolicy); policy != "" {
		w.Header().Set("Cache-Control", policy)
	}
	if origin := strings.TrimSpace(h.settings.CORSAllowOrigin); origin != "" {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", strings.Join([]string{http.MethodGet, http.MethodPost, http.MethodOptions}, ", "))
	}
}

func (h Handler) writeBadRequest(w http.ResponseWriter, message string) {
	h.applyHeaders(w, &http.Request{})
	http.Error(w, message, http.StatusBadRequest)
}

func stateFromContext(ctx context.Context) requestState {
	state, _ := ctx.Value(requestStateKey{}).(requestState)
	return state
}

func pWrap(h Handler) statusResponse {
	plugin := Plugin{settings: h.settings}
	return plugin.statusPayload()
}
