package menus

import (
	"context"
	"fmt"

	domainauthz "github.com/fastygo/cms/internal/domain/authz"
	domainmenus "github.com/fastygo/cms/internal/domain/menus"
)

type Repository interface {
	SaveMenu(context.Context, domainmenus.Menu) error
	ListMenus(context.Context) ([]domainmenus.Menu, error)
	GetMenuByLocation(context.Context, domainmenus.Location) (domainmenus.Menu, bool, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return Service{repo: repo}
}

func (s Service) Save(ctx context.Context, principal domainauthz.Principal, menu domainmenus.Menu) error {
	if !principal.Has(domainauthz.CapabilityMenusManage) {
		return fmt.Errorf("capability %q is required", domainauthz.CapabilityMenusManage)
	}
	return s.repo.SaveMenu(ctx, menu)
}

func (s Service) ByLocation(ctx context.Context, location domainmenus.Location) (domainmenus.Menu, bool, error) {
	return s.repo.GetMenuByLocation(ctx, location)
}

func (s Service) List(ctx context.Context) ([]domainmenus.Menu, error) {
	return s.repo.ListMenus(ctx)
}
