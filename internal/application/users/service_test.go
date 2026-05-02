package users_test

import (
	"context"
	"testing"

	appusers "github.com/fastygo/cms/internal/application/users"
	domainusers "github.com/fastygo/cms/internal/domain/users"
)

func TestPublicAuthorProjectionHidesPrivateUserData(t *testing.T) {
	ctx := context.Background()
	repo := &memoryUserRepo{users: map[domainusers.ID]domainusers.User{
		"user-1": {
			ID:          "user-1",
			Login:       "jane",
			DisplayName: "Jane Doe",
			Email:       "jane@example.test",
			Status:      domainusers.StatusActive,
			Profile: domainusers.AuthorProfile{
				Bio:       "Editor",
				AvatarURL: "/avatars/jane.png",
			},
		},
	}}
	service := appusers.NewService(repo)

	author, ok, err := service.PublicAuthor(ctx, "user-1")
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("expected public author projection")
	}
	if author.ID != "user-1" || author.DisplayName != "Jane Doe" || author.Slug != "jane" {
		t.Fatalf("unexpected author projection: %+v", author)
	}
}

type memoryUserRepo struct {
	users map[domainusers.ID]domainusers.User
}

func (r *memoryUserRepo) GetUser(_ context.Context, id domainusers.ID) (domainusers.User, bool, error) {
	user, ok := r.users[id]
	return user, ok, nil
}

func (r *memoryUserRepo) ListUsers(context.Context) ([]domainusers.User, error) {
	users := make([]domainusers.User, 0, len(r.users))
	for _, user := range r.users {
		users = append(users, user)
	}
	return users, nil
}

func (r *memoryUserRepo) SaveUser(_ context.Context, user domainusers.User) error {
	r.users[user.ID] = user
	return nil
}
