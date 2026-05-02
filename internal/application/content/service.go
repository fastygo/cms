package content

import (
	"context"
	"fmt"
	"time"

	domainauthz "github.com/fastygo/cms/internal/domain/authz"
	domaincontent "github.com/fastygo/cms/internal/domain/content"
	domaincontenttype "github.com/fastygo/cms/internal/domain/contenttype"
)

type Repository interface {
	NextID(context.Context) (domaincontent.ID, error)
	Save(context.Context, domaincontent.Entry) error
	Get(context.Context, domaincontent.ID) (domaincontent.Entry, error)
	List(context.Context, domaincontent.Query) (domaincontent.ListResult, error)
}

type TypeRegistry interface {
	GetContentType(context.Context, domaincontent.Kind) (domaincontenttype.Type, bool, error)
}

type Clock func() time.Time

type Service struct {
	repo  Repository
	types TypeRegistry
	clock Clock
}

type CreateDraftCommand struct {
	Kind            domaincontent.Kind
	Title           domaincontent.LocalizedText
	Slug            domaincontent.LocalizedText
	Body            domaincontent.LocalizedText
	Excerpt         domaincontent.LocalizedText
	AuthorID        string
	FeaturedMediaID string
	Template        string
	Metadata        domaincontent.Metadata
	Terms           []domaincontent.TermRef
}

type UpdateCommand struct {
	ID              domaincontent.ID
	Title           domaincontent.LocalizedText
	Slug            domaincontent.LocalizedText
	Body            domaincontent.LocalizedText
	Excerpt         domaincontent.LocalizedText
	AuthorID        string
	FeaturedMediaID string
	Template        string
	Metadata        domaincontent.Metadata
	Terms           []domaincontent.TermRef
}

func NewService(repo Repository, types TypeRegistry, clock Clock) Service {
	if clock == nil {
		clock = time.Now
	}
	return Service{repo: repo, types: types, clock: clock}
}

func (s Service) CreateDraft(ctx context.Context, principal domainauthz.Principal, cmd CreateDraftCommand) (domaincontent.Entry, error) {
	if !principal.Has(domainauthz.CapabilityContentCreate) {
		return domaincontent.Entry{}, fmt.Errorf("capability %q is required", domainauthz.CapabilityContentCreate)
	}
	if err := s.ensureKind(ctx, cmd.Kind); err != nil {
		return domaincontent.Entry{}, err
	}
	id, err := s.repo.NextID(ctx)
	if err != nil {
		return domaincontent.Entry{}, err
	}
	now := s.clock()
	entry := domaincontent.Entry{
		ID:              id,
		Kind:            cmd.Kind,
		Status:          domaincontent.StatusDraft,
		Visibility:      domaincontent.VisibilityPublic,
		Title:           cmd.Title,
		Slug:            cmd.Slug,
		Body:            cmd.Body,
		Excerpt:         cmd.Excerpt,
		AuthorID:        firstNonEmpty(cmd.AuthorID, principal.ID),
		FeaturedMediaID: cmd.FeaturedMediaID,
		Template:        cmd.Template,
		Metadata:        cmd.Metadata,
		Terms:           cmd.Terms,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if err := s.repo.Save(ctx, entry); err != nil {
		return domaincontent.Entry{}, err
	}
	return entry, nil
}

func (s Service) Update(ctx context.Context, principal domainauthz.Principal, cmd UpdateCommand) (domaincontent.Entry, error) {
	entry, err := s.repo.Get(ctx, cmd.ID)
	if err != nil {
		return domaincontent.Entry{}, err
	}
	if err := requireEdit(principal, entry); err != nil {
		return domaincontent.Entry{}, err
	}
	entry.Title = cmd.Title
	entry.Slug = cmd.Slug
	entry.Body = cmd.Body
	entry.Excerpt = cmd.Excerpt
	if cmd.AuthorID != "" {
		entry.AuthorID = cmd.AuthorID
	}
	entry.FeaturedMediaID = cmd.FeaturedMediaID
	entry.Template = cmd.Template
	entry.Metadata = cmd.Metadata
	entry.Terms = cmd.Terms
	entry.UpdatedAt = s.clock()
	if err := s.repo.Save(ctx, entry); err != nil {
		return domaincontent.Entry{}, err
	}
	return entry, nil
}

func (s Service) Publish(ctx context.Context, principal domainauthz.Principal, id domaincontent.ID) (domaincontent.Entry, error) {
	if !principal.Has(domainauthz.CapabilityContentPublish) {
		return domaincontent.Entry{}, fmt.Errorf("capability %q is required", domainauthz.CapabilityContentPublish)
	}
	entry, err := s.repo.Get(ctx, id)
	if err != nil {
		return domaincontent.Entry{}, err
	}
	now := s.clock()
	entry.Status = domaincontent.StatusPublished
	entry.PublishedAt = &now
	entry.DeletedAt = nil
	entry.UpdatedAt = now
	return entry, s.repo.Save(ctx, entry)
}

func (s Service) Schedule(ctx context.Context, principal domainauthz.Principal, id domaincontent.ID, publishAt time.Time) (domaincontent.Entry, error) {
	if !principal.Has(domainauthz.CapabilityContentSchedule) {
		return domaincontent.Entry{}, fmt.Errorf("capability %q is required", domainauthz.CapabilityContentSchedule)
	}
	entry, err := s.repo.Get(ctx, id)
	if err != nil {
		return domaincontent.Entry{}, err
	}
	now := s.clock()
	entry.Status = domaincontent.StatusScheduled
	entry.PublishedAt = &publishAt
	entry.UpdatedAt = now
	return entry, s.repo.Save(ctx, entry)
}

func (s Service) Trash(ctx context.Context, principal domainauthz.Principal, id domaincontent.ID) (domaincontent.Entry, error) {
	if !principal.Has(domainauthz.CapabilityContentDelete) {
		return domaincontent.Entry{}, fmt.Errorf("capability %q is required", domainauthz.CapabilityContentDelete)
	}
	entry, err := s.repo.Get(ctx, id)
	if err != nil {
		return domaincontent.Entry{}, err
	}
	now := s.clock()
	entry.Status = domaincontent.StatusTrashed
	entry.DeletedAt = &now
	entry.UpdatedAt = now
	return entry, s.repo.Save(ctx, entry)
}

func (s Service) Restore(ctx context.Context, principal domainauthz.Principal, id domaincontent.ID) (domaincontent.Entry, error) {
	if !principal.Has(domainauthz.CapabilityContentRestore) {
		return domaincontent.Entry{}, fmt.Errorf("capability %q is required", domainauthz.CapabilityContentRestore)
	}
	entry, err := s.repo.Get(ctx, id)
	if err != nil {
		return domaincontent.Entry{}, err
	}
	entry.Status = domaincontent.StatusDraft
	entry.DeletedAt = nil
	entry.UpdatedAt = s.clock()
	return entry, s.repo.Save(ctx, entry)
}

func (s Service) List(ctx context.Context, query domaincontent.Query) (domaincontent.ListResult, error) {
	if query.PublicOnly && query.PublishedAt.IsZero() {
		query.PublishedAt = s.clock()
	}
	return s.repo.List(ctx, query)
}

func (s Service) GetBySlug(ctx context.Context, principal domainauthz.Principal, kind domaincontent.Kind, slug string, locale string) (domaincontent.Entry, error) {
	result, err := s.List(ctx, domaincontent.Query{
		Kinds:      []domaincontent.Kind{kind},
		Slug:       slug,
		Locale:     locale,
		PublicOnly: !principal.Has(domainauthz.CapabilityContentReadPrivate),
		Page:       1,
		PerPage:    1,
	})
	if err != nil {
		return domaincontent.Entry{}, err
	}
	if len(result.Items) == 0 {
		return domaincontent.Entry{}, fmt.Errorf("content with slug %q is not found", slug)
	}
	return result.Items[0], nil
}

func (s Service) Get(ctx context.Context, principal domainauthz.Principal, id domaincontent.ID) (domaincontent.Entry, error) {
	entry, err := s.repo.Get(ctx, id)
	if err != nil {
		return domaincontent.Entry{}, err
	}
	if entry.IsPublicAt(s.clock()) || principal.Has(domainauthz.CapabilityContentReadPrivate) {
		return entry, nil
	}
	return domaincontent.Entry{}, fmt.Errorf("content %q is not public", id)
}

func (s Service) ensureKind(ctx context.Context, kind domaincontent.Kind) error {
	if err := domaincontent.ValidateKind(kind); err != nil {
		return err
	}
	if s.types == nil {
		return nil
	}
	_, ok, err := s.types.GetContentType(ctx, kind)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("content type %q is not registered", kind)
	}
	return nil
}

func requireEdit(principal domainauthz.Principal, entry domaincontent.Entry) error {
	if principal.Has(domainauthz.CapabilityContentEdit) {
		return nil
	}
	if principal.ID == entry.AuthorID && principal.Has(domainauthz.CapabilityContentEditOwn) {
		return nil
	}
	if principal.ID != entry.AuthorID && principal.Has(domainauthz.CapabilityContentEditOthers) {
		return nil
	}
	return fmt.Errorf("capability %q is required", domainauthz.CapabilityContentEdit)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
