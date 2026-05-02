package contenttype

import (
	"context"

	domaincontent "github.com/fastygo/cms/internal/domain/content"
	domaincontenttype "github.com/fastygo/cms/internal/domain/contenttype"
)

type Repository interface {
	SaveContentType(context.Context, domaincontenttype.Type) error
	GetContentType(context.Context, domaincontent.Kind) (domaincontenttype.Type, bool, error)
	ListContentTypes(context.Context) ([]domaincontenttype.Type, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return Service{repo: repo}
}

func (s Service) InstallBuiltIns(ctx context.Context) error {
	if err := s.Register(ctx, domaincontenttype.BuiltInPost()); err != nil {
		return err
	}
	return s.Register(ctx, domaincontenttype.BuiltInPage())
}

func (s Service) Register(ctx context.Context, contentType domaincontenttype.Type) error {
	if err := domaincontenttype.Validate(contentType); err != nil {
		return err
	}
	return s.repo.SaveContentType(ctx, contentType)
}

func (s Service) Get(ctx context.Context, kind domaincontent.Kind) (domaincontenttype.Type, bool, error) {
	return s.repo.GetContentType(ctx, kind)
}

func (s Service) List(ctx context.Context) ([]domaincontenttype.Type, error) {
	return s.repo.ListContentTypes(ctx)
}
