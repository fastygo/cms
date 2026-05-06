package plugins

import (
	"context"
	"fmt"
	"net/http"
	"slices"
	"sort"
	"strings"

	"github.com/fastygo/cms/internal/domain/authz"
	"github.com/fastygo/cms/internal/site/adminfixtures"
	"github.com/fastygo/cms/internal/site/ui/elements"
	"github.com/fastygo/framework/pkg/app"
)

type State string

const (
	StateInstalled   State = "installed"
	StateActive      State = "active"
	StateInactive    State = "inactive"
	StateFailed      State = "failed"
	StateUninstalled State = "uninstalled"
)

type Surface string

const (
	SurfaceAdmin Surface = "admin"
	SurfaceREST  Surface = "rest"
	SurfacePublic Surface = "public"
)

type Manifest struct {
	ID           string
	Name         string
	Version      string
	Contract     string
	Description  string
	Author       string
	Requires     map[string]string
	Capabilities []CapabilityDefinition
	Settings     []SettingDefinition
	Hooks        []HookRegistration
	Assets       []Asset
}

func (m Manifest) Validate() error {
	if !validPluginID(m.ID) {
		return fmt.Errorf("plugin id %q is invalid", m.ID)
	}
	if strings.TrimSpace(m.Name) == "" {
		return fmt.Errorf("plugin %q name is required", m.ID)
	}
	if strings.TrimSpace(m.Version) == "" {
		return fmt.Errorf("plugin %q version is required", m.ID)
	}
	if strings.TrimSpace(m.Contract) == "" {
		return fmt.Errorf("plugin %q contract is required", m.ID)
	}
	return nil
}

type CapabilityDefinition struct {
	ID          string
	Description string
}

type SettingDefinition struct {
	Key        string
	Type       string
	Default    string
	Public     bool
	Capability authz.Capability
}

type HookRegistration struct {
	HookID   string
	HandlerID string
	OwnerID  string
	Category string
	Priority int
}

type Asset struct {
	ID      string
	Surface Surface
	Path    string
}

type AdminMenuItem struct {
	ID         string
	Label      string
	Path       string
	Icon       string
	Order      int
	Capability authz.Capability
}

type ScreenActionRegistration struct {
	ScreenID string
	Build    func(adminfixtures.AdminFixture) []elements.Action
}

type Route struct {
	Pattern          string
	Surface          Surface
	Capability       authz.Capability
	Protected        bool
	Handler          http.HandlerFunc
	ProtectedHandler func(http.ResponseWriter, *http.Request, authz.Principal)
}

type Registry struct {
	adminMenu      []AdminMenuItem
	screenActions  map[string][]ScreenActionRegistration
	routes         []Route
	assets         []Asset
	capabilities   []CapabilityDefinition
	settings       []SettingDefinition
	hooks          []HookRegistration
}

func NewRegistry() *Registry {
	return &Registry{screenActions: map[string][]ScreenActionRegistration{}}
}

func (r *Registry) AddAdminMenu(item AdminMenuItem) {
	r.adminMenu = append(r.adminMenu, item)
}

func (r *Registry) AddScreenActions(actions ...ScreenActionRegistration) {
	for _, action := range actions {
		r.screenActions[action.ScreenID] = append(r.screenActions[action.ScreenID], action)
	}
}

func (r *Registry) AddRoutes(routes ...Route) {
	r.routes = append(r.routes, routes...)
}

func (r *Registry) AddAssets(assets ...Asset) {
	r.assets = append(r.assets, assets...)
}

func (r *Registry) AddCapabilities(capabilities ...CapabilityDefinition) {
	r.capabilities = append(r.capabilities, capabilities...)
}

func (r *Registry) AddSettings(settings ...SettingDefinition) {
	r.settings = append(r.settings, settings...)
}

func (r *Registry) AddHooks(hooks ...HookRegistration) {
	r.hooks = append(r.hooks, hooks...)
}

func (r *Registry) AdminMenu(principal authz.Principal) []app.NavItem {
	items := make([]app.NavItem, 0, len(r.adminMenu))
	for _, item := range r.adminMenu {
		if item.Capability != "" && !principal.Has(item.Capability) {
			continue
		}
		items = append(items, app.NavItem{Label: item.Label, Path: item.Path, Icon: item.Icon, Order: item.Order})
	}
	sort.SliceStable(items, func(i, j int) bool { return items[i].Order < items[j].Order })
	return items
}

func (r *Registry) ScreenActions(screenID string, fixture adminfixtures.AdminFixture) []elements.Action {
	builders := r.screenActions[screenID]
	actions := make([]elements.Action, 0, len(builders))
	for _, builder := range builders {
		actions = append(actions, builder.Build(fixture)...)
	}
	return actions
}

func (r *Registry) AssetsForSurface(surface Surface) []Asset {
	result := []Asset{}
	for _, asset := range r.assets {
		if asset.Surface == surface {
			result = append(result, asset)
		}
	}
	return result
}

func (r *Registry) RoutesForSurface(surface Surface) []Route {
	result := []Route{}
	for _, route := range r.routes {
		if route.Surface == surface {
			result = append(result, route)
		}
	}
	return result
}

func (r *Registry) Capabilities() []CapabilityDefinition {
	return append([]CapabilityDefinition(nil), r.capabilities...)
}

func (r *Registry) Settings() []SettingDefinition {
	return append([]SettingDefinition(nil), r.settings...)
}

func (r *Registry) Hooks() []HookRegistration {
	return append([]HookRegistration(nil), r.hooks...)
}

type StateRecord struct {
	ID        string
	Version   string
	State     State
	LastError string
}

type StateRepository interface {
	Load(context.Context) (map[string]StateRecord, error)
	Save(context.Context, StateRecord) error
}

type InMemoryStateRepository struct {
	states map[string]StateRecord
}

func NewInMemoryStateRepository() *InMemoryStateRepository {
	return &InMemoryStateRepository{states: map[string]StateRecord{}}
}

func (r *InMemoryStateRepository) Load(context.Context) (map[string]StateRecord, error) {
	result := make(map[string]StateRecord, len(r.states))
	for id, record := range r.states {
		result[id] = record
	}
	return result, nil
}

func (r *InMemoryStateRepository) Save(_ context.Context, record StateRecord) error {
	r.states[record.ID] = record
	return nil
}

type Descriptor interface {
	Manifest() Manifest
	Register(context.Context, *Registry) error
}

type Runtime struct {
	descriptors map[string]Descriptor
	state       StateRepository
}

func NewRuntime(state StateRepository, descriptors ...Descriptor) (*Runtime, error) {
	if state == nil {
		state = NewInMemoryStateRepository()
	}
	registry := make(map[string]Descriptor, len(descriptors))
	for _, descriptor := range descriptors {
		manifest := descriptor.Manifest()
		if err := manifest.Validate(); err != nil {
			return nil, err
		}
		if _, exists := registry[manifest.ID]; exists {
			return nil, fmt.Errorf("duplicate plugin id %q", manifest.ID)
		}
		registry[manifest.ID] = descriptor
	}
	return &Runtime{descriptors: registry, state: state}, nil
}

func (r *Runtime) Activate(ctx context.Context, active []string) (*Registry, error) {
	records, err := r.state.Load(ctx)
	if err != nil {
		return nil, err
	}
	registry := NewRegistry()
	for _, id := range active {
		descriptor, ok := r.descriptors[id]
		if !ok {
			return nil, fmt.Errorf("plugin %q is not compiled into this binary", id)
		}
		manifest := descriptor.Manifest()
		record := StateRecord{ID: manifest.ID, Version: manifest.Version, State: StateInstalled}
		if existing, ok := records[id]; ok {
			record = existing
		}
		if err := descriptor.Register(ctx, registry); err != nil {
			record.State = StateFailed
			record.LastError = err.Error()
			_ = r.state.Save(ctx, record)
			return nil, err
		}
		record.State = StateActive
		record.LastError = ""
		if err := r.state.Save(ctx, record); err != nil {
			return nil, err
		}
	}
	for id, descriptor := range r.descriptors {
		if slices.Contains(active, id) {
			continue
		}
		record := StateRecord{ID: id, Version: descriptor.Manifest().Version, State: StateInactive}
		if err := r.state.Save(ctx, record); err != nil {
			return nil, err
		}
	}
	return registry, nil
}

func (r *Runtime) Deactivate(ctx context.Context, ids []string) error {
	for _, id := range ids {
		descriptor, ok := r.descriptors[id]
		if !ok {
			return fmt.Errorf("plugin %q is not compiled into this binary", id)
		}
		if err := r.state.Save(ctx, StateRecord{
			ID:      id,
			Version: descriptor.Manifest().Version,
			State:   StateInactive,
		}); err != nil {
			return err
		}
	}
	return nil
}

func validPluginID(value string) bool {
	if strings.TrimSpace(value) == "" {
		return false
	}
	for _, char := range value {
		switch {
		case char >= 'a' && char <= 'z':
		case char >= '0' && char <= '9':
		case char == '-', char == '.':
		default:
			return false
		}
	}
	return true
}
