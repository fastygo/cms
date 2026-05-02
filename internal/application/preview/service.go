package preview

import (
	"context"
	"fmt"
	"time"

	domainauthz "github.com/fastygo/cms/internal/domain/authz"
	domaincontent "github.com/fastygo/cms/internal/domain/content"
	domainpreview "github.com/fastygo/cms/internal/domain/preview"
)

type Repository interface {
	SavePreview(context.Context, domainpreview.Access) error
	GetPreview(context.Context, domainpreview.Token) (domainpreview.Access, bool, error)
}

type Service struct {
	repo  Repository
	clock func() time.Time
}

func NewService(repo Repository, clock func() time.Time) Service {
	if clock == nil {
		clock = time.Now
	}
	return Service{repo: repo, clock: clock}
}

func (s Service) Create(ctx context.Context, principal domainauthz.Principal, entryID domaincontent.ID, ttl time.Duration) (domainpreview.Access, error) {
	if !principal.Has(domainauthz.CapabilityContentReadPrivate) && !principal.Has(domainauthz.CapabilityContentEdit) {
		return domainpreview.Access{}, fmt.Errorf("capability %q is required", domainauthz.CapabilityContentReadPrivate)
	}
	token, err := domainpreview.NewToken()
	if err != nil {
		return domainpreview.Access{}, err
	}
	now := s.clock()
	access := domainpreview.Access{
		Token:       token,
		EntryID:     entryID,
		PrincipalID: principal.ID,
		CreatedAt:   now,
		ExpiresAt:   now.Add(ttl),
	}
	return access, s.repo.SavePreview(ctx, access)
}

func (s Service) Validate(ctx context.Context, token domainpreview.Token) (domainpreview.Access, bool, error) {
	access, ok, err := s.repo.GetPreview(ctx, token)
	if err != nil || !ok {
		return domainpreview.Access{}, ok, err
	}
	return access, access.ValidAt(s.clock()), nil
}
