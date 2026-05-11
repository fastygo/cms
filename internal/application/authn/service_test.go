package authn

import (
	"context"
	"errors"
	"testing"
	"time"

	domainauthn "github.com/fastygo/cms/internal/domain/authn"
	"github.com/fastygo/cms/internal/domain/authz"
	domainusers "github.com/fastygo/cms/internal/domain/users"
)

func TestLocalPasswordRecoveryTokensAndLockout(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 5, 9, 12, 0, 0, 0, time.UTC)
	users := &memoryUserRepo{items: map[domainusers.ID]domainusers.User{
		"admin": {ID: "admin", Login: "admin", Email: "admin@example.test", Status: domainusers.StatusActive, Roles: []string{authz.RoleAdmin}},
	}}
	repo := &memoryRepo{}
	service := NewService(users, repo, WithNow(func() time.Time { return now }))

	if _, err := service.SetPassword(ctx, "admin", "admin", false); err != nil {
		t.Fatalf("SetPassword() error = %v", err)
	}

	result, err := service.AuthenticatePassword(ctx, PasswordLoginInput{Identifier: "admin@example.test", Password: "admin", RemoteAddr: "127.0.0.1"}, domainauthn.DefaultSessionPolicy(false))
	if err != nil {
		t.Fatalf("AuthenticatePassword() error = %v", err)
	}
	if !result.Principal.Has(authz.CapabilitySettingsManage) {
		t.Fatalf("principal = %+v, want admin capabilities", result.Principal)
	}

	recoveryCodes, err := service.CreateRecoveryCodes(ctx, "admin", 2)
	if err != nil {
		t.Fatalf("CreateRecoveryCodes() error = %v", err)
	}
	if len(recoveryCodes) != 2 {
		t.Fatalf("recovery code count = %d", len(recoveryCodes))
	}
	recovered, err := service.UseRecoveryCode(ctx, "admin@example.test", recoveryCodes[0], "admin-rotated", domainauthn.DefaultSessionPolicy(false))
	if err != nil {
		t.Fatalf("UseRecoveryCode() error = %v", err)
	}
	if recovered.User.PasswordHash == "" {
		t.Fatalf("expected recovered user password hash to be set")
	}
	if _, err := service.UseRecoveryCode(ctx, "admin@example.test", recoveryCodes[0], "admin-second", domainauthn.DefaultSessionPolicy(false)); !errors.Is(err, ErrRecoveryFailed) {
		t.Fatalf("second recovery use error = %v, want ErrRecoveryFailed", err)
	}

	resetRaw, _, err := service.CreateResetToken(ctx, "root", "admin", "Admin reset", 30*time.Minute, false, false)
	if err != nil {
		t.Fatalf("CreateResetToken() error = %v", err)
	}
	if _, err := service.ResetPasswordWithToken(ctx, resetRaw, "admin-reset"); err != nil {
		t.Fatalf("ResetPasswordWithToken() error = %v", err)
	}
	if _, err := service.ResetPasswordWithToken(ctx, resetRaw, "admin-reset-again"); !errors.Is(err, ErrResetTokenFailed) {
		t.Fatalf("second reset token use error = %v, want ErrResetTokenFailed", err)
	}

	appTokenRaw, appToken, err := service.CreateAppToken(ctx, "admin", "CLI", []authz.Capability{authz.CapabilityContentReadPrivate}, 24*time.Hour)
	if err != nil {
		t.Fatalf("CreateAppToken() error = %v", err)
	}
	principal, ok, err := service.AuthenticateAppToken(ctx, appTokenRaw, "127.0.0.1")
	if err != nil || !ok {
		t.Fatalf("AuthenticateAppToken() ok=%v err=%v", ok, err)
	}
	if !principal.Has(authz.CapabilityContentReadPrivate) {
		t.Fatalf("app token principal = %+v", principal)
	}
	if err := service.RevokeAppToken(ctx, appToken.Prefix); err != nil {
		t.Fatalf("RevokeAppToken() error = %v", err)
	}
	if _, ok, err := service.AuthenticateAppToken(ctx, appTokenRaw, "127.0.0.1"); err != nil || ok {
		t.Fatalf("revoked app token ok=%v err=%v", ok, err)
	}

	policy := domainauthn.DefaultSessionPolicy(false)
	policy.MaxAttempts = 2
	policy.AttemptWindow = time.Hour
	policy.LockoutWindow = time.Hour
	for i := 0; i < 2; i++ {
		if _, err := service.AuthenticatePassword(ctx, PasswordLoginInput{Identifier: "admin@example.test", Password: "wrong", RemoteAddr: "127.0.0.1"}, policy); !errors.Is(err, ErrInvalidCredentials) {
			t.Fatalf("wrong password #%d error = %v", i+1, err)
		}
	}
	if _, err := service.AuthenticatePassword(ctx, PasswordLoginInput{Identifier: "admin@example.test", Password: "wrong", RemoteAddr: "127.0.0.1"}, policy); !errors.Is(err, ErrLoginLocked) {
		t.Fatalf("expected lockout error, got %v", err)
	}
}

type memoryUserRepo struct {
	items map[domainusers.ID]domainusers.User
}

func (r *memoryUserRepo) GetUser(_ context.Context, id domainusers.ID) (domainusers.User, bool, error) {
	user, ok := r.items[id]
	return user, ok, nil
}

func (r *memoryUserRepo) ListUsers(context.Context) ([]domainusers.User, error) {
	result := make([]domainusers.User, 0, len(r.items))
	for _, user := range r.items {
		result = append(result, user)
	}
	return result, nil
}

func (r *memoryUserRepo) SaveUser(_ context.Context, user domainusers.User) error {
	r.items[user.ID] = user
	return nil
}

type memoryRepo struct {
	recovery []domainauthn.RecoveryCode
	resets   map[string]domainauthn.ResetToken
	tokens   map[string]domainauthn.AppToken
	attempts []domainauthn.LoginAttempt
}

func (r *memoryRepo) SaveRecoveryCode(_ context.Context, code domainauthn.RecoveryCode) error {
	for i, item := range r.recovery {
		if item.ID == code.ID {
			r.recovery[i] = code
			return nil
		}
	}
	r.recovery = append(r.recovery, code)
	return nil
}

func (r *memoryRepo) ListRecoveryCodes(_ context.Context, userID domainusers.ID) ([]domainauthn.RecoveryCode, error) {
	result := []domainauthn.RecoveryCode{}
	for _, item := range r.recovery {
		if item.UserID == userID {
			result = append(result, item)
		}
	}
	return result, nil
}

func (r *memoryRepo) SaveResetToken(_ context.Context, token domainauthn.ResetToken) error {
	if r.resets == nil {
		r.resets = map[string]domainauthn.ResetToken{}
	}
	r.resets[token.ID] = token
	return nil
}

func (r *memoryRepo) GetResetToken(_ context.Context, id string) (domainauthn.ResetToken, bool, error) {
	item, ok := r.resets[id]
	return item, ok, nil
}

func (r *memoryRepo) ListResetTokens(_ context.Context, userID domainusers.ID) ([]domainauthn.ResetToken, error) {
	result := []domainauthn.ResetToken{}
	for _, item := range r.resets {
		if item.UserID == userID {
			result = append(result, item)
		}
	}
	return result, nil
}

func (r *memoryRepo) SaveAppToken(_ context.Context, token domainauthn.AppToken) error {
	if r.tokens == nil {
		r.tokens = map[string]domainauthn.AppToken{}
	}
	r.tokens[token.Prefix] = token
	return nil
}

func (r *memoryRepo) GetAppTokenByPrefix(_ context.Context, prefix string) (domainauthn.AppToken, bool, error) {
	item, ok := r.tokens[prefix]
	return item, ok, nil
}

func (r *memoryRepo) ListAppTokens(_ context.Context, userID domainusers.ID) ([]domainauthn.AppToken, error) {
	result := []domainauthn.AppToken{}
	for _, item := range r.tokens {
		if item.UserID == userID {
			result = append(result, item)
		}
	}
	return result, nil
}

func (r *memoryRepo) SaveLoginAttempt(_ context.Context, attempt domainauthn.LoginAttempt) error {
	r.attempts = append(r.attempts, attempt)
	return nil
}

func (r *memoryRepo) ListLoginAttempts(_ context.Context, key string, since time.Time) ([]domainauthn.LoginAttempt, error) {
	result := []domainauthn.LoginAttempt{}
	for _, item := range r.attempts {
		if item.Key == key && (item.CreatedAt.After(since) || item.CreatedAt.Equal(since)) {
			result = append(result, item)
		}
	}
	return result, nil
}
