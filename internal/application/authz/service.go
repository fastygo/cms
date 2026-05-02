package authz

import (
	"context"
	"fmt"

	domainauthz "github.com/fastygo/cms/internal/domain/authz"
)

type Service struct{}

func NewService() Service {
	return Service{}
}

func (Service) Require(_ context.Context, principal domainauthz.Principal, capability domainauthz.Capability) error {
	if !principal.Has(capability) {
		return fmt.Errorf("capability %q is required", capability)
	}
	return nil
}
