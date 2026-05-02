package users

import (
	"context"

	domainusers "github.com/fastygo/cms/internal/domain/users"
)

type Repository interface {
	GetUser(context.Context, domainusers.ID) (domainusers.User, bool, error)
	ListUsers(context.Context) ([]domainusers.User, error)
	SaveUser(context.Context, domainusers.User) error
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return Service{repo: repo}
}

func (s Service) Save(ctx context.Context, user domainusers.User) error {
	return s.repo.SaveUser(ctx, user)
}

func (s Service) PublicAuthor(ctx context.Context, id domainusers.ID) (domainusers.AuthorProfile, bool, error) {
	user, ok, err := s.repo.GetUser(ctx, id)
	if err != nil || !ok {
		return domainusers.AuthorProfile{}, ok, err
	}
	if user.Status != domainusers.StatusActive {
		return domainusers.AuthorProfile{}, false, nil
	}
	return user.PublicAuthor(), true, nil
}

func (s Service) List(ctx context.Context) ([]domainusers.User, error) {
	return s.repo.ListUsers(ctx)
}
