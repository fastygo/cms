package media

import (
	"context"
	"fmt"

	domainauthz "github.com/fastygo/cms/internal/domain/authz"
	domaincontent "github.com/fastygo/cms/internal/domain/content"
	domainmedia "github.com/fastygo/cms/internal/domain/media"
)

type Repository interface {
	GetMedia(context.Context, domainmedia.ID) (domainmedia.Asset, bool, error)
	ListMedia(context.Context) ([]domainmedia.Asset, error)
	SaveMedia(context.Context, domainmedia.Asset) error
}

type EntryRepository interface {
	Get(context.Context, domaincontent.ID) (domaincontent.Entry, error)
	Save(context.Context, domaincontent.Entry) error
}

type Service struct {
	repo    Repository
	entries EntryRepository
}

func NewService(repo Repository, entries EntryRepository) Service {
	return Service{repo: repo, entries: entries}
}

func (s Service) SaveMetadata(ctx context.Context, principal domainauthz.Principal, asset domainmedia.Asset) error {
	if !principal.Has(domainauthz.CapabilityMediaEdit) && !principal.Has(domainauthz.CapabilityMediaUpload) {
		return fmt.Errorf("capability %q is required", domainauthz.CapabilityMediaEdit)
	}
	return s.repo.SaveMedia(ctx, asset)
}

func (s Service) List(ctx context.Context) ([]domainmedia.Asset, error) {
	return s.repo.ListMedia(ctx)
}

func (s Service) Get(ctx context.Context, id domainmedia.ID) (domainmedia.Asset, bool, error) {
	return s.repo.GetMedia(ctx, id)
}

func (s Service) AttachFeatured(ctx context.Context, principal domainauthz.Principal, contentID domaincontent.ID, assetID domainmedia.ID) (domaincontent.Entry, error) {
	if !principal.Has(domainauthz.CapabilityMediaEdit) {
		return domaincontent.Entry{}, fmt.Errorf("capability %q is required", domainauthz.CapabilityMediaEdit)
	}
	if _, ok, err := s.repo.GetMedia(ctx, assetID); err != nil {
		return domaincontent.Entry{}, err
	} else if !ok {
		return domaincontent.Entry{}, fmt.Errorf("media asset %q is not registered", assetID)
	}
	entry, err := s.entries.Get(ctx, contentID)
	if err != nil {
		return domaincontent.Entry{}, err
	}
	entry.FeaturedMediaID = string(assetID)
	return entry, s.entries.Save(ctx, entry)
}
