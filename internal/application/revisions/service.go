package revisions

import (
	"context"
	"fmt"
	"time"

	domainauthz "github.com/fastygo/cms/internal/domain/authz"
	domaincontent "github.com/fastygo/cms/internal/domain/content"
	domainrevisions "github.com/fastygo/cms/internal/domain/revisions"
)

type Repository interface {
	NextRevisionID(context.Context) (domainrevisions.ID, error)
	SaveRevision(context.Context, domainrevisions.Revision) error
	GetRevision(context.Context, domainrevisions.ID) (domainrevisions.Revision, error)
}

type EntryRepository interface {
	Get(context.Context, domaincontent.ID) (domaincontent.Entry, error)
	Save(context.Context, domaincontent.Entry) error
}

type Service struct {
	repo    Repository
	entries EntryRepository
	clock   func() time.Time
}

func NewService(repo Repository, entries EntryRepository, clock func() time.Time) Service {
	if clock == nil {
		clock = time.Now
	}
	return Service{repo: repo, entries: entries, clock: clock}
}

func (s Service) Create(ctx context.Context, principal domainauthz.Principal, entryID domaincontent.ID, reason string) (domainrevisions.Revision, error) {
	if !principal.Has(domainauthz.CapabilityContentEdit) && !principal.Has(domainauthz.CapabilityContentEditOwn) {
		return domainrevisions.Revision{}, fmt.Errorf("capability %q is required", domainauthz.CapabilityContentEdit)
	}
	entry, err := s.entries.Get(ctx, entryID)
	if err != nil {
		return domainrevisions.Revision{}, err
	}
	id, err := s.repo.NextRevisionID(ctx)
	if err != nil {
		return domainrevisions.Revision{}, err
	}
	revision := domainrevisions.Revision{
		ID:        id,
		EntryID:   entryID,
		Snapshot:  entry,
		AuthorID:  principal.ID,
		Reason:    reason,
		CreatedAt: s.clock(),
	}
	return revision, s.repo.SaveRevision(ctx, revision)
}

func (s Service) Restore(ctx context.Context, principal domainauthz.Principal, revisionID domainrevisions.ID) (domaincontent.Entry, error) {
	if !principal.Has(domainauthz.CapabilityContentRestore) {
		return domaincontent.Entry{}, fmt.Errorf("capability %q is required", domainauthz.CapabilityContentRestore)
	}
	revision, err := s.repo.GetRevision(ctx, revisionID)
	if err != nil {
		return domaincontent.Entry{}, err
	}
	entry := revision.Snapshot
	entry.UpdatedAt = s.clock()
	return entry, s.entries.Save(ctx, entry)
}
