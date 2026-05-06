package settings

import (
	"context"
	"fmt"

	domainauthz "github.com/fastygo/cms/internal/domain/authz"
	domainsettings "github.com/fastygo/cms/internal/domain/settings"
)

type Repository interface {
	SaveSetting(context.Context, domainsettings.Value) error
	GetSetting(context.Context, domainsettings.Key) (domainsettings.Value, bool, error)
	ListPublicSettings(context.Context) ([]domainsettings.Value, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return Service{repo: repo}
}

func (s Service) Save(ctx context.Context, principal domainauthz.Principal, value domainsettings.Value) error {
	if !principal.Has(domainauthz.CapabilitySettingsManage) {
		return fmt.Errorf("capability %q is required", domainauthz.CapabilitySettingsManage)
	}
	return s.repo.SaveSetting(ctx, value)
}

func (s Service) Get(ctx context.Context, key domainsettings.Key) (domainsettings.Value, bool, error) {
	return s.repo.GetSetting(ctx, key)
}

func (s Service) Public(ctx context.Context) ([]domainsettings.Value, error) {
	return s.repo.ListPublicSettings(ctx)
}
