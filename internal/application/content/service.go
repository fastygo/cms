package content

import (
	"context"
	"fmt"
	"time"

	appmeta "github.com/fastygo/cms/internal/application/meta"
	domainauthz "github.com/fastygo/cms/internal/domain/authz"
	domaincontent "github.com/fastygo/cms/internal/domain/content"
	domaincontenttype "github.com/fastygo/cms/internal/domain/contenttype"
	platformplugins "github.com/fastygo/cms/internal/platform/plugins"
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
	repo       Repository
	types      TypeRegistry
	clock      Clock
	meta       *appmeta.Registry
	hookGetter func() *platformplugins.Registry
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

type Option func(*Service)

func WithMetadataRegistry(registry *appmeta.Registry) Option {
	return func(service *Service) {
		service.meta = registry
	}
}

func WithHookRegistry(getter func() *platformplugins.Registry) Option {
	return func(service *Service) {
		service.hookGetter = getter
	}
}

type MetadataHookPayload struct {
	Operation string
	ContentID domaincontent.ID
	Kind      domaincontent.Kind
	Metadata  domaincontent.Metadata
}

func NewService(repo Repository, types TypeRegistry, clock Clock, options ...Option) Service {
	if clock == nil {
		clock = time.Now
	}
	service := Service{repo: repo, types: types, clock: clock}
	for _, option := range options {
		if option != nil {
			option(&service)
		}
	}
	return service
}

func (s Service) CreateDraft(ctx context.Context, principal domainauthz.Principal, cmd CreateDraftCommand) (domaincontent.Entry, error) {
	if !principal.Has(domainauthz.CapabilityContentCreate) {
		return domaincontent.Entry{}, fmt.Errorf("capability %q is required", domainauthz.CapabilityContentCreate)
	}
	if err := s.ensureKind(ctx, cmd.Kind); err != nil {
		return domaincontent.Entry{}, err
	}
	metadata, err := s.normalizeMetadata(ctx, principal, "create", "", cmd.Kind, cmd.Metadata)
	if err != nil {
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
		Metadata:        metadata,
		Terms:           cmd.Terms,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if err := s.dispatchMetadataAction(ctx, "content.metadata.persist.before", "create", entry.ID, entry.Kind, entry.Metadata); err != nil {
		return domaincontent.Entry{}, err
	}
	if err := s.repo.Save(ctx, entry); err != nil {
		return domaincontent.Entry{}, err
	}
	if err := s.dispatchMetadataAction(ctx, "content.metadata.persist.after", "create", entry.ID, entry.Kind, entry.Metadata); err != nil {
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
	metadata, err := s.normalizeMetadata(ctx, principal, "update", entry.ID, entry.Kind, cmd.Metadata)
	if err != nil {
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
	entry.Metadata = metadata
	entry.Terms = cmd.Terms
	entry.UpdatedAt = s.clock()
	if err := s.dispatchMetadataAction(ctx, "content.metadata.persist.before", "update", entry.ID, entry.Kind, entry.Metadata); err != nil {
		return domaincontent.Entry{}, err
	}
	if err := s.repo.Save(ctx, entry); err != nil {
		return domaincontent.Entry{}, err
	}
	if err := s.dispatchMetadataAction(ctx, "content.metadata.persist.after", "update", entry.ID, entry.Kind, entry.Metadata); err != nil {
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

func (s Service) normalizeMetadata(ctx context.Context, principal domainauthz.Principal, operation string, contentID domaincontent.ID, kind domaincontent.Kind, metadata domaincontent.Metadata) (domaincontent.Metadata, error) {
	if err := s.dispatchMetadataAction(ctx, "content.metadata.validate.before", operation, contentID, kind, metadata); err != nil {
		return nil, err
	}
	if s.meta == nil {
		if err := s.dispatchMetadataAction(ctx, "content.metadata.validate.after", operation, contentID, kind, metadata); err != nil {
			return nil, err
		}
		return metadata, nil
	}
	normalized, err := s.meta.Normalize(principal, kind, metadata)
	if err != nil {
		return nil, err
	}
	if err := s.dispatchMetadataAction(ctx, "content.metadata.validate.after", operation, contentID, kind, normalized); err != nil {
		return nil, err
	}
	return normalized, nil
}

func (s Service) dispatchMetadataAction(ctx context.Context, hookID string, operation string, contentID domaincontent.ID, kind domaincontent.Kind, metadata domaincontent.Metadata) error {
	if s.hookGetter == nil {
		return nil
	}
	registry := s.hookGetter()
	if registry == nil {
		return nil
	}
	return registry.DispatchAction(ctx, hookID, platformplugins.HookContext{
		Metadata: map[string]any{
			"operation":  operation,
			"content_id": string(contentID),
			"kind":       string(kind),
		},
	}, MetadataHookPayload{
		Operation: operation,
		ContentID: contentID,
		Kind:      kind,
		Metadata:  metadata,
	})
}
