package snapshot

import (
	"context"
	"time"

	domaincontent "github.com/fastygo/cms/internal/domain/content"
	domaincontenttype "github.com/fastygo/cms/internal/domain/contenttype"
	domainmedia "github.com/fastygo/cms/internal/domain/media"
	domainmenus "github.com/fastygo/cms/internal/domain/menus"
	domainsettings "github.com/fastygo/cms/internal/domain/settings"
	domaintaxonomy "github.com/fastygo/cms/internal/domain/taxonomy"
	domainusers "github.com/fastygo/cms/internal/domain/users"
)

const SnapshotVersion = "gocms.snapshot.v1"

type Repository interface {
	List(context.Context, domaincontent.Query) (domaincontent.ListResult, error)
	Save(context.Context, domaincontent.Entry) error
	ListContentTypes(context.Context) ([]domaincontenttype.Type, error)
	SaveContentType(context.Context, domaincontenttype.Type) error
	ListDefinitions(context.Context) ([]domaintaxonomy.Definition, error)
	SaveDefinition(context.Context, domaintaxonomy.Definition) error
	ListTerms(context.Context, domaintaxonomy.Type) ([]domaintaxonomy.Term, error)
	SaveTerm(context.Context, domaintaxonomy.Term) error
	ListMedia(context.Context) ([]domainmedia.Asset, error)
	SaveMedia(context.Context, domainmedia.Asset) error
	ListUsers(context.Context) ([]domainusers.User, error)
	SaveUser(context.Context, domainusers.User) error
	ListSettings(context.Context) ([]domainsettings.Value, error)
	SaveSetting(context.Context, domainsettings.Value) error
	ListMenus(context.Context) ([]domainmenus.Menu, error)
	SaveMenu(context.Context, domainmenus.Menu) error
}

type Service struct {
	repo Repository
	now  func() time.Time
}

type Bundle struct {
	Version             string                      `json:"version"`
	ExportedAt          time.Time                   `json:"exported_at"`
	Content             []domaincontent.Entry       `json:"content"`
	ContentTypes        []domaincontenttype.Type    `json:"content_types"`
	TaxonomyDefinitions []domaintaxonomy.Definition `json:"taxonomy_definitions"`
	TaxonomyTerms       []domaintaxonomy.Term       `json:"taxonomy_terms"`
	Media               []domainmedia.Asset         `json:"media"`
	Users               []domainusers.User          `json:"users"`
	Settings            []domainsettings.Value      `json:"settings"`
	Menus               []domainmenus.Menu          `json:"menus"`
}

func NewService(repo Repository, now func() time.Time) Service {
	if now == nil {
		now = time.Now
	}
	return Service{repo: repo, now: now}
}

func (s Service) Export(ctx context.Context) (Bundle, error) {
	content, err := listAllContent(ctx, s.repo)
	if err != nil {
		return Bundle{}, err
	}
	types, err := s.repo.ListContentTypes(ctx)
	if err != nil {
		return Bundle{}, err
	}
	definitions, err := s.repo.ListDefinitions(ctx)
	if err != nil {
		return Bundle{}, err
	}
	terms := make([]domaintaxonomy.Term, 0)
	for _, definition := range definitions {
		rows, err := s.repo.ListTerms(ctx, definition.Type)
		if err != nil {
			return Bundle{}, err
		}
		terms = append(terms, rows...)
	}
	media, err := s.repo.ListMedia(ctx)
	if err != nil {
		return Bundle{}, err
	}
	users, err := s.repo.ListUsers(ctx)
	if err != nil {
		return Bundle{}, err
	}
	settings, err := s.repo.ListSettings(ctx)
	if err != nil {
		return Bundle{}, err
	}
	menus, err := s.repo.ListMenus(ctx)
	if err != nil {
		return Bundle{}, err
	}
	return Bundle{
		Version:             SnapshotVersion,
		ExportedAt:          s.now().UTC(),
		Content:             content,
		ContentTypes:        types,
		TaxonomyDefinitions: definitions,
		TaxonomyTerms:       terms,
		Media:               media,
		Users:               users,
		Settings:            settings,
		Menus:               menus,
	}, nil
}

func (s Service) Import(ctx context.Context, bundle Bundle) error {
	for _, contentType := range bundle.ContentTypes {
		if err := s.repo.SaveContentType(ctx, contentType); err != nil {
			return err
		}
	}
	for _, definition := range bundle.TaxonomyDefinitions {
		if err := s.repo.SaveDefinition(ctx, definition); err != nil {
			return err
		}
	}
	for _, term := range bundle.TaxonomyTerms {
		if err := s.repo.SaveTerm(ctx, term); err != nil {
			return err
		}
	}
	for _, asset := range bundle.Media {
		if err := s.repo.SaveMedia(ctx, asset); err != nil {
			return err
		}
	}
	for _, user := range bundle.Users {
		if err := s.repo.SaveUser(ctx, user); err != nil {
			return err
		}
	}
	for _, value := range bundle.Settings {
		if err := s.repo.SaveSetting(ctx, value); err != nil {
			return err
		}
	}
	for _, menu := range bundle.Menus {
		if err := s.repo.SaveMenu(ctx, menu); err != nil {
			return err
		}
	}
	for _, entry := range bundle.Content {
		if err := s.repo.Save(ctx, entry); err != nil {
			return err
		}
	}
	return nil
}

func listAllContent(ctx context.Context, repo Repository) ([]domaincontent.Entry, error) {
	page := 1
	perPage := 200
	result := []domaincontent.Entry{}
	for {
		list, err := repo.List(ctx, domaincontent.Query{Page: page, PerPage: perPage})
		if err != nil {
			return nil, err
		}
		result = append(result, list.Items...)
		if list.TotalPages == 0 || page >= list.TotalPages {
			return result, nil
		}
		page++
	}
}
