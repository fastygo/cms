package media

import (
	"context"
	"fmt"
	"net/url"
	"strings"

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
	if err := validateAsset(asset); err != nil {
		return err
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

func validateAsset(asset domainmedia.Asset) error {
	switch {
	case strings.TrimSpace(string(asset.ID)) == "":
		return fmt.Errorf("media id is required")
	case strings.TrimSpace(asset.Filename) == "":
		return fmt.Errorf("media filename is required")
	case !allowedMIME(asset.MimeType):
		return fmt.Errorf("invalid mime type %q", asset.MimeType)
	case strings.TrimSpace(asset.PublicURL) == "":
		return fmt.Errorf("public_url is required")
	case !validAssetURL(asset.PublicURL):
		return fmt.Errorf("invalid public_url")
	case asset.SizeBytes < 0:
		return fmt.Errorf("invalid size_bytes")
	case asset.Width < 0 || asset.Height < 0:
		return fmt.Errorf("invalid media dimensions")
	case (asset.Width == 0) != (asset.Height == 0):
		return fmt.Errorf("invalid media dimensions")
	}
	if err := validateProviderRef(asset.ProviderRef); err != nil {
		return err
	}
	return nil
}

func validateProviderRef(ref domainmedia.BlobRef) error {
	if strings.TrimSpace(ref.Provider) == "" {
		if strings.TrimSpace(ref.Key) != "" || strings.TrimSpace(ref.URL) != "" || strings.TrimSpace(ref.Checksum) != "" || strings.TrimSpace(ref.ETag) != "" {
			return fmt.Errorf("invalid provider_ref: provider is required when key, url, checksum, or etag is set")
		}
		return nil
	}
	if strings.TrimSpace(ref.Key) == "" && strings.TrimSpace(ref.URL) == "" {
		return fmt.Errorf("invalid provider_ref: key or url is required")
	}
	if strings.TrimSpace(ref.URL) != "" && !validAssetURL(ref.URL) {
		return fmt.Errorf("invalid provider_ref url")
	}
	return nil
}

func allowedMIME(value string) bool {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case "image/jpeg", "image/png", "image/webp", "image/avif", "image/gif":
		return true
	default:
		return false
	}
}

func validAssetURL(value string) bool {
	if strings.HasPrefix(value, "/") {
		return true
	}
	parsed, err := url.Parse(strings.TrimSpace(value))
	return err == nil && (parsed.Scheme == "http" || parsed.Scheme == "https") && parsed.Host != ""
}
