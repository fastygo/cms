package plugins

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/fastygo/cms/internal/domain/authz"
	"github.com/fastygo/cms/internal/site/adminfixtures"
	"github.com/fastygo/cms/internal/site/ui/elements"
	"github.com/fastygo/framework/pkg/app"
	"github.com/fastygo/panel"
)

type State string

const (
	StateInstalled   State = "installed"
	StateActive      State = "active"
	StateInactive    State = "inactive"
	StateFailed      State = "failed"
	StateUninstalled State = "uninstalled"
)

const (
	SurfaceAdmin  = panel.SurfaceAdmin
	SurfaceREST   = panel.SurfaceREST
	SurfacePublic = panel.SurfacePublic
)

type Surface = panel.Surface

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

type HookCategory string

const (
	HookCategoryAction HookCategory = "action"
	HookCategoryFilter HookCategory = "filter"
)

type HookErrorPolicy string

const (
	HookErrorPolicyFail    HookErrorPolicy = "fail"
	HookErrorPolicyCollect HookErrorPolicy = "collect"
)

type HookRegistration struct {
	HookID      string
	HandlerID   string
	OwnerID     string
	Category    HookCategory
	Priority    int
	ErrorPolicy HookErrorPolicy
}

type Asset = panel.Asset
type AdminMenuItem = panel.MenuItem[authz.Capability]

type ScreenActionRegistration struct {
	ScreenID string
	Build    func(adminfixtures.AdminFixture) []elements.Action
}

type EditorProviderRegistration = panel.EditorProviderRegistration
type Route = panel.Route[authz.Principal, authz.Capability]

type HookContext struct {
	HookID        string
	Surface       Surface
	Path          string
	Locale        string
	Principal     authz.Principal
	Authenticated bool
	Metadata      map[string]any
}

type ContentProjection struct {
	Slug     map[string]string
	Title    map[string]string
	Content  map[string]string
	Excerpt  map[string]string
	Metadata map[string]any
}

type ActionHandler func(context.Context, HookContext, any) error
type FilterHandler func(context.Context, HookContext, any) (any, error)

type ActionHandlerRegistration struct {
	Hook   HookRegistration
	Handle ActionHandler
}

type FilterHandlerRegistration struct {
	Hook   HookRegistration
	Handle FilterHandler
}

type registeredActionHandler struct {
	ActionHandlerRegistration
	order int
}

type registeredFilterHandler struct {
	FilterHandlerRegistration
	order int
}

type Registry struct {
	panel         *panel.Registry[authz.Principal, authz.Capability]
	screenActions map[string][]ScreenActionRegistration
	capabilities  []CapabilityDefinition
	settings      []SettingDefinition
	hooks         []HookRegistration
	actionHooks   map[string][]registeredActionHandler
	filterHooks   map[string][]registeredFilterHandler
	hookOrder     int
}

func NewRegistry() *Registry {
	return &Registry{
		panel:         panel.NewRegistry[authz.Principal, authz.Capability](),
		screenActions: map[string][]ScreenActionRegistration{},
		actionHooks:   map[string][]registeredActionHandler{},
		filterHooks:   map[string][]registeredFilterHandler{},
	}
}

func (r *Registry) AddAdminMenu(item AdminMenuItem) {
	r.panel.AddMenuItems(item)
}

func (r *Registry) AddScreenActions(actions ...ScreenActionRegistration) {
	for _, action := range actions {
		r.screenActions[action.ScreenID] = append(r.screenActions[action.ScreenID], action)
	}
}

func (r *Registry) AddEditorProviders(providers ...EditorProviderRegistration) {
	r.panel.AddEditorProviders(providers...)
}

func (r *Registry) AddRoutes(routes ...Route) {
	r.panel.AddRoutes(routes...)
}

func (r *Registry) AddAssets(assets ...Asset) {
	r.panel.AddAssets(assets...)
}

func (r *Registry) AddCapabilities(capabilities ...CapabilityDefinition) {
	r.capabilities = append(r.capabilities, capabilities...)
}

func (r *Registry) AddSettings(settings ...SettingDefinition) {
	r.settings = append(r.settings, settings...)
}

func (r *Registry) AddHooks(hooks ...HookRegistration) {
	for _, hook := range hooks {
		hook.Category = normalizeHookCategory(hook.Category, HookCategoryAction)
		hook.Priority = normalizeHookPriority(hook.Priority)
		hook.ErrorPolicy = normalizeHookErrorPolicy(hook.ErrorPolicy)
		r.registerHook(hook)
	}
}

func (r *Registry) AddActionHandlers(registrations ...ActionHandlerRegistration) {
	for _, registration := range registrations {
		registration.Hook.Category = normalizeHookCategory(registration.Hook.Category, HookCategoryAction)
		registration.Hook.Priority = normalizeHookPriority(registration.Hook.Priority)
		registration.Hook.ErrorPolicy = normalizeHookErrorPolicy(registration.Hook.ErrorPolicy)
		r.registerHook(registration.Hook)
		r.actionHooks[registration.Hook.HookID] = append(r.actionHooks[registration.Hook.HookID], registeredActionHandler{
			ActionHandlerRegistration: registration,
			order:                     r.hookOrder,
		})
		r.hookOrder++
	}
}

func (r *Registry) AddFilterHandlers(registrations ...FilterHandlerRegistration) {
	for _, registration := range registrations {
		registration.Hook.Category = normalizeHookCategory(registration.Hook.Category, HookCategoryFilter)
		registration.Hook.Priority = normalizeHookPriority(registration.Hook.Priority)
		registration.Hook.ErrorPolicy = normalizeHookErrorPolicy(registration.Hook.ErrorPolicy)
		r.registerHook(registration.Hook)
		r.filterHooks[registration.Hook.HookID] = append(r.filterHooks[registration.Hook.HookID], registeredFilterHandler{
			FilterHandlerRegistration: registration,
			order:                     r.hookOrder,
		})
		r.hookOrder++
	}
}

func (r *Registry) AdminMenu(principal authz.Principal) []app.NavItem {
	menu := r.panel.MenuItems(principal)
	items := make([]app.NavItem, 0, len(menu))
	for _, item := range menu {
		items = append(items, app.NavItem{Label: item.Label, Path: item.Path, Icon: item.Icon, Order: item.Order})
	}
	return items
}

func (r *Registry) AdminMenuItems(principal authz.Principal) []AdminMenuItem {
	return r.panel.MenuItems(principal)
}

func (r *Registry) ScreenActions(screenID string, fixture adminfixtures.AdminFixture) []elements.Action {
	builders := r.screenActions[screenID]
	actions := make([]elements.Action, 0, len(builders))
	for _, builder := range builders {
		actions = append(actions, builder.Build(fixture)...)
	}
	return actions
}

func (r *Registry) EditorProviders() []EditorProviderRegistration {
	return r.panel.EditorProviders()
}

func (r *Registry) ResolveEditorProvider(id string) (EditorProviderRegistration, bool) {
	return r.panel.ResolveEditorProvider(id)
}

func (r *Registry) AssetsForSurface(surface Surface) []Asset {
	return r.panel.AssetsForSurface(surface)
}

func (r *Registry) RoutesForSurface(surface Surface) []Route {
	return r.panel.RoutesForSurface(surface)
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

func (r *Registry) DispatchAction(ctx context.Context, hookID string, hookContext HookContext, payload any) error {
	if r == nil {
		return nil
	}
	hookContext.HookID = hookID
	handlers := append([]registeredActionHandler(nil), r.actionHooks[hookID]...)
	slices.SortFunc(handlers, compareActionHandlers)
	failures := make([]string, 0)
	for _, handler := range handlers {
		if handler.Handle == nil {
			continue
		}
		if err := handler.Handle(ctx, hookContext, payload); err != nil {
			switch handler.Hook.ErrorPolicy {
			case HookErrorPolicyCollect:
				failures = append(failures, fmt.Sprintf("%s: %v", handler.Hook.HandlerID, err))
			default:
				return fmt.Errorf("action hook %q handler %q failed: %w", hookID, handler.Hook.HandlerID, err)
			}
		}
	}
	if len(failures) > 0 {
		return fmt.Errorf("action hook %q failed: %s", hookID, strings.Join(failures, "; "))
	}
	return nil
}

func (r *Registry) ApplyFilter(ctx context.Context, hookID string, hookContext HookContext, value any) (any, error) {
	if r == nil {
		return value, nil
	}
	hookContext.HookID = hookID
	current := value
	handlers := append([]registeredFilterHandler(nil), r.filterHooks[hookID]...)
	slices.SortFunc(handlers, compareFilterHandlers)
	failures := make([]string, 0)
	for _, handler := range handlers {
		if handler.Handle == nil {
			continue
		}
		next, err := handler.Handle(ctx, hookContext, current)
		if err != nil {
			switch handler.Hook.ErrorPolicy {
			case HookErrorPolicyCollect:
				failures = append(failures, fmt.Sprintf("%s: %v", handler.Hook.HandlerID, err))
				continue
			default:
				return current, fmt.Errorf("filter hook %q handler %q failed: %w", hookID, handler.Hook.HandlerID, err)
			}
		}
		current = next
	}
	if len(failures) > 0 {
		return current, fmt.Errorf("filter hook %q failed: %s", hookID, strings.Join(failures, "; "))
	}
	return current, nil
}

func FilterValue[T any](ctx context.Context, registry *Registry, hookID string, hookContext HookContext, value T) (T, error) {
	var zero T
	if registry == nil {
		return value, nil
	}
	raw, err := registry.ApplyFilter(ctx, hookID, hookContext, value)
	if err != nil {
		return zero, err
	}
	if raw == nil {
		return zero, fmt.Errorf("filter hook %q returned <nil>", hookID)
	}
	typed, ok := raw.(T)
	if !ok {
		return zero, fmt.Errorf("filter hook %q returned %T", hookID, raw)
	}
	return typed, nil
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
		if err := registry.DispatchAction(ctx, "plugin.activate.before", HookContext{
			Metadata: map[string]any{
				"plugin_id":      manifest.ID,
				"plugin_version": manifest.Version,
			},
		}, map[string]any{
			"plugin_id":      manifest.ID,
			"plugin_version": manifest.Version,
		}); err != nil {
			record.State = StateFailed
			record.LastError = err.Error()
			_ = r.state.Save(ctx, record)
			return nil, err
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
		if err := registry.DispatchAction(ctx, "plugin.activate.after", HookContext{
			Metadata: map[string]any{
				"plugin_id":      manifest.ID,
				"plugin_version": manifest.Version,
			},
		}, map[string]any{
			"plugin_id":      manifest.ID,
			"plugin_version": manifest.Version,
		}); err != nil {
			record.State = StateFailed
			record.LastError = err.Error()
			_ = r.state.Save(ctx, record)
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
	records, err := r.state.Load(ctx)
	if err != nil {
		return err
	}
	registry := NewRegistry()
	for _, record := range records {
		if record.State != StateActive {
			continue
		}
		descriptor, ok := r.descriptors[record.ID]
		if !ok {
			continue
		}
		if err := descriptor.Register(ctx, registry); err != nil {
			return err
		}
	}
	for _, id := range ids {
		descriptor, ok := r.descriptors[id]
		if !ok {
			return fmt.Errorf("plugin %q is not compiled into this binary", id)
		}
		manifest := descriptor.Manifest()
		payload := map[string]any{
			"plugin_id":      manifest.ID,
			"plugin_version": manifest.Version,
		}
		if err := registry.DispatchAction(ctx, "plugin.deactivate.before", HookContext{Metadata: payload}, payload); err != nil {
			return err
		}
		if err := r.state.Save(ctx, StateRecord{
			ID:      id,
			Version: manifest.Version,
			State:   StateInactive,
		}); err != nil {
			return err
		}
		if err := registry.DispatchAction(ctx, "plugin.deactivate.after", HookContext{Metadata: payload}, payload); err != nil {
			return err
		}
	}
	return nil
}

func compareActionHandlers(left registeredActionHandler, right registeredActionHandler) int {
	if left.Hook.Priority != right.Hook.Priority {
		return left.Hook.Priority - right.Hook.Priority
	}
	return left.order - right.order
}

func compareFilterHandlers(left registeredFilterHandler, right registeredFilterHandler) int {
	if left.Hook.Priority != right.Hook.Priority {
		return left.Hook.Priority - right.Hook.Priority
	}
	return left.order - right.order
}

func normalizeHookCategory(value HookCategory, fallback HookCategory) HookCategory {
	if strings.TrimSpace(string(value)) == "" {
		return fallback
	}
	return value
}

func normalizeHookPriority(value int) int {
	if value == 0 {
		return 100
	}
	return value
}

func normalizeHookErrorPolicy(value HookErrorPolicy) HookErrorPolicy {
	switch value {
	case HookErrorPolicyCollect:
		return value
	default:
		return HookErrorPolicyFail
	}
}

func (r *Registry) registerHook(hook HookRegistration) {
	for _, existing := range r.hooks {
		if existing.HookID == hook.HookID && existing.HandlerID == hook.HandlerID && existing.OwnerID == hook.OwnerID {
			return
		}
	}
	r.hooks = append(r.hooks, hook)
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
