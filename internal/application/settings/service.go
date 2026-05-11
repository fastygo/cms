package settings

import (
	"context"
	"fmt"

	domainauthz "github.com/fastygo/cms/internal/domain/authz"
	domainsettings "github.com/fastygo/cms/internal/domain/settings"
	platformplugins "github.com/fastygo/cms/internal/platform/plugins"
)

type Repository interface {
	SaveSetting(context.Context, domainsettings.Value) error
	GetSetting(context.Context, domainsettings.Key) (domainsettings.Value, bool, error)
	ListPublicSettings(context.Context) ([]domainsettings.Value, error)
	ListSettings(context.Context) ([]domainsettings.Value, error)
}

type Service struct {
	repo       Repository
	registry   *Registry
	hookGetter func() *platformplugins.Registry
}

type Option func(*Service)

func WithRegistry(registry *Registry) Option {
	return func(service *Service) {
		service.registry = registry
	}
}

func WithHookRegistry(getter func() *platformplugins.Registry) Option {
	return func(service *Service) {
		service.hookGetter = getter
	}
}

type HookPayload struct {
	Operation string
	Key       domainsettings.Key
	Value     domainsettings.Value
}

func NewService(repo Repository, options ...Option) Service {
	service := Service{repo: repo}
	for _, option := range options {
		if option != nil {
			option(&service)
		}
	}
	return service
}

func (s Service) Save(ctx context.Context, principal domainauthz.Principal, value domainsettings.Value) error {
	operation := "save"
	if err := s.dispatchAction(ctx, "settings.validate.before", operation, value); err != nil {
		return err
	}
	definition, ok := s.definition(value.Key)
	if ok {
		if definition.Capability != "" && !principal.Has(definition.Capability) {
			return fmt.Errorf("capability %q is required", definition.Capability)
		}
		normalized, err := NormalizeValue(definition, value.Value)
		if err != nil {
			return fmt.Errorf("setting %q %w", value.Key, err)
		}
		value.Value = normalized
		value.Public = definition.Public
	} else {
		if !principal.Has(domainauthz.CapabilitySettingsManage) {
			return fmt.Errorf("capability %q is required", domainauthz.CapabilitySettingsManage)
		}
	}
	if err := s.dispatchAction(ctx, "settings.validate.after", operation, value); err != nil {
		return err
	}
	if err := s.dispatchAction(ctx, "settings.update.before", operation, value); err != nil {
		return err
	}
	if err := s.repo.SaveSetting(ctx, value); err != nil {
		return err
	}
	return s.dispatchAction(ctx, "settings.update.after", operation, value)
}

func (s Service) Get(ctx context.Context, key domainsettings.Key) (domainsettings.Value, bool, error) {
	value, ok, err := s.repo.GetSetting(ctx, key)
	if err != nil {
		return domainsettings.Value{}, false, err
	}
	definition, defined := s.definition(key)
	if !ok {
		if !defined || definition.Default == nil {
			return domainsettings.Value{}, false, nil
		}
		return domainsettings.Value{Key: key, Value: definition.Default, Public: definition.Public}, true, nil
	}
	if defined {
		value.Public = definition.Public
	}
	return value, true, nil
}

func (s Service) Public(ctx context.Context) ([]domainsettings.Value, error) {
	if s.registry == nil {
		return s.repo.ListPublicSettings(ctx)
	}
	values, err := s.repo.ListSettings(ctx)
	if err != nil {
		return nil, err
	}
	return MergeStoredWithDefinitions(s.registry, values, true), nil
}

func (s Service) All(ctx context.Context) ([]domainsettings.Value, error) {
	if s.registry == nil {
		return s.repo.ListSettings(ctx)
	}
	values, err := s.repo.ListSettings(ctx)
	if err != nil {
		return nil, err
	}
	return MergeStoredWithDefinitions(s.registry, values, false), nil
}

func (s Service) definition(key domainsettings.Key) (domainsettings.Definition, bool) {
	if s.registry == nil {
		return domainsettings.Definition{}, false
	}
	return s.registry.Definition(key)
}

func (s Service) dispatchAction(ctx context.Context, hookID string, operation string, value domainsettings.Value) error {
	if s.hookGetter == nil {
		return nil
	}
	registry := s.hookGetter()
	if registry == nil {
		return nil
	}
	return registry.DispatchAction(ctx, hookID, platformplugins.HookContext{
		Metadata: map[string]any{
			"operation": operation,
			"key":       string(value.Key),
		},
	}, HookPayload{
		Operation: operation,
		Key:       value.Key,
		Value:     value,
	})
}
