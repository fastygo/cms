package authn

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	domainauthn "github.com/fastygo/cms/internal/domain/authn"
	"github.com/fastygo/cms/internal/domain/authz"
	domainusers "github.com/fastygo/cms/internal/domain/users"
	"github.com/fastygo/cms/internal/platform/loginbrute"
	"github.com/google/uuid"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrLoginLocked        = errors.New("login temporarily locked")
	ErrUserNotActive      = errors.New("user is not active")
	ErrRecoveryFailed     = errors.New("recovery code is invalid or unavailable")
	ErrResetTokenFailed   = errors.New("reset token is invalid or unavailable")
	ErrAppTokenFailed     = errors.New("app token is invalid or unavailable")
)

type UserRepository interface {
	GetUser(context.Context, domainusers.ID) (domainusers.User, bool, error)
	ListUsers(context.Context) ([]domainusers.User, error)
	SaveUser(context.Context, domainusers.User) error
}

type Repository interface {
	SaveRecoveryCode(context.Context, domainauthn.RecoveryCode) error
	ListRecoveryCodes(context.Context, domainusers.ID) ([]domainauthn.RecoveryCode, error)
	SaveResetToken(context.Context, domainauthn.ResetToken) error
	GetResetToken(context.Context, string) (domainauthn.ResetToken, bool, error)
	ListResetTokens(context.Context, domainusers.ID) ([]domainauthn.ResetToken, error)
	SaveAppToken(context.Context, domainauthn.AppToken) error
	GetAppTokenByPrefix(context.Context, string) (domainauthn.AppToken, bool, error)
	ListAppTokens(context.Context, domainusers.ID) ([]domainauthn.AppToken, error)
	SaveLoginAttempt(context.Context, domainauthn.LoginAttempt) error
	ListLoginAttempts(context.Context, string, time.Time) ([]domainauthn.LoginAttempt, error)
}

type PasswordLoginInput struct {
	Identifier string
	Password   string
	RemoteAddr string
	UserAgent  string
}

type ExternalIdentity struct {
	ProviderID string
	Subject    string
	Email      string
	Groups     []string
}

type PasswordLoginResult struct {
	User               domainusers.User
	Principal          authz.Principal
	ProviderID         string
	MustChangePassword bool
}

type IdentityProvider interface {
	ID() string
	AuthenticatePassword(context.Context, PasswordLoginInput, domainauthn.SessionPolicy) (PasswordLoginResult, error)
}

type ExternalIdentityProvider interface {
	ID() string
	ResolveIdentity(context.Context, ExternalIdentity) (PasswordLoginResult, error)
}

type Service struct {
	users  UserRepository
	repo   Repository
	hasher PasswordHasher
	now    func() time.Time
}

type Option func(*Service)

func WithHasher(hasher PasswordHasher) Option {
	return func(s *Service) {
		if hasher != nil {
			s.hasher = hasher
		}
	}
}

func WithNow(now func() time.Time) Option {
	return func(s *Service) {
		if now != nil {
			s.now = now
		}
	}
}

func NewService(users UserRepository, repo Repository, options ...Option) Service {
	service := Service{
		users:  users,
		repo:   repo,
		hasher: DefaultPasswordHasher(),
		now:    time.Now,
	}
	for _, option := range options {
		option(&service)
	}
	return service
}

func (s Service) Enabled() bool {
	return s.users != nil && s.repo != nil && s.hasher != nil && s.now != nil
}

func (s Service) LocalProvider() IdentityProvider {
	return LocalPasswordProvider{service: s}
}

func (s Service) AuthenticatePassword(ctx context.Context, input PasswordLoginInput, policy domainauthn.SessionPolicy) (PasswordLoginResult, error) {
	return s.LocalProvider().AuthenticatePassword(ctx, input, policy)
}

func (s Service) SetPassword(ctx context.Context, userID domainusers.ID, password string, mustChange bool) (domainusers.User, error) {
	user, ok, err := s.users.GetUser(ctx, userID)
	if err != nil {
		return domainusers.User{}, err
	}
	if !ok {
		return domainusers.User{}, fmt.Errorf("user %q not found", userID)
	}
	hash, err := s.hasher.Hash(password)
	if err != nil {
		return domainusers.User{}, err
	}
	now := s.now().UTC()
	user.PasswordHash = hash
	user.MustChangePassword = mustChange
	user.PasswordUpdatedAt = &now
	if err := s.users.SaveUser(ctx, user); err != nil {
		return domainusers.User{}, err
	}
	return user, nil
}

func (s Service) CreateRecoveryCodes(ctx context.Context, userID domainusers.ID, count int) ([]string, error) {
	if count <= 0 {
		count = 8
	}
	now := s.now().UTC()
	result := make([]string, 0, count)
	for i := 0; i < count; i++ {
		raw, err := randomNumericCode(10)
		if err != nil {
			return nil, err
		}
		hash, err := s.hasher.Hash(raw)
		if err != nil {
			return nil, err
		}
		if err := s.repo.SaveRecoveryCode(ctx, domainauthn.RecoveryCode{
			ID:        uuid.NewString(),
			UserID:    userID,
			Label:     fmt.Sprintf("recovery-%02d", i+1),
			Hash:      hash,
			CreatedAt: now,
		}); err != nil {
			return nil, err
		}
		result = append(result, raw)
	}
	return result, nil
}

func (s Service) ListRecoveryCodes(ctx context.Context, userID domainusers.ID) ([]domainauthn.RecoveryCode, error) {
	return s.repo.ListRecoveryCodes(ctx, userID)
}

func (s Service) UseRecoveryCode(ctx context.Context, identifier string, code string, newPassword string, policy domainauthn.SessionPolicy) (PasswordLoginResult, error) {
	user, ok, err := s.lookupUser(ctx, identifier)
	if err != nil {
		return PasswordLoginResult{}, err
	}
	if !ok || user.Status != domainusers.StatusActive {
		return PasswordLoginResult{}, ErrRecoveryFailed
	}
	items, err := s.repo.ListRecoveryCodes(ctx, user.ID)
	if err != nil {
		return PasswordLoginResult{}, err
	}
	now := s.now().UTC()
	for _, item := range items {
		if !item.Available(now) || !s.hasher.Verify(code, item.Hash) {
			continue
		}
		item.UsedAt = &now
		if err := s.repo.SaveRecoveryCode(ctx, item); err != nil {
			return PasswordLoginResult{}, err
		}
		if _, err := s.SetPassword(ctx, user.ID, newPassword, false); err != nil {
			return PasswordLoginResult{}, err
		}
		return s.authenticateLocal(ctx, PasswordLoginInput{Identifier: identifier, Password: newPassword}, policy)
	}
	return PasswordLoginResult{}, ErrRecoveryFailed
}

func (s Service) CreateResetToken(ctx context.Context, issuedBy string, userID domainusers.ID, label string, ttl time.Duration, bootstrap bool, maintenance bool) (string, domainauthn.ResetToken, error) {
	if ttl <= 0 {
		ttl = 30 * time.Minute
	}
	id := uuid.NewString()
	raw, err := prefixedToken(id, 24)
	if err != nil {
		return "", domainauthn.ResetToken{}, err
	}
	hash, err := s.hasher.Hash(raw)
	if err != nil {
		return "", domainauthn.ResetToken{}, err
	}
	token := domainauthn.ResetToken{
		ID:          id,
		UserID:      userID,
		Label:       strings.TrimSpace(label),
		Hash:        hash,
		CreatedAt:   s.now().UTC(),
		ExpiresAt:   s.now().UTC().Add(ttl),
		IssuedBy:    strings.TrimSpace(issuedBy),
		Bootstrap:   bootstrap,
		Maintenance: maintenance,
	}
	if err := s.repo.SaveResetToken(ctx, token); err != nil {
		return "", domainauthn.ResetToken{}, err
	}
	return raw, token, nil
}

func (s Service) ResetPasswordWithToken(ctx context.Context, raw string, newPassword string) (domainusers.User, error) {
	id, _, ok := normalizePlaintextToken(raw)
	if !ok {
		return domainusers.User{}, ErrResetTokenFailed
	}
	token, found, err := s.repo.GetResetToken(ctx, id)
	if err != nil {
		return domainusers.User{}, err
	}
	now := s.now().UTC()
	if !found || !token.Available(now) || !s.hasher.Verify(raw, token.Hash) {
		return domainusers.User{}, ErrResetTokenFailed
	}
	token.UsedAt = &now
	if err := s.repo.SaveResetToken(ctx, token); err != nil {
		return domainusers.User{}, err
	}
	return s.SetPassword(ctx, token.UserID, newPassword, false)
}

func (s Service) ListResetTokens(ctx context.Context, userID domainusers.ID) ([]domainauthn.ResetToken, error) {
	return s.repo.ListResetTokens(ctx, userID)
}

func (s Service) CreateAppToken(ctx context.Context, userID domainusers.ID, name string, capabilities []authz.Capability, ttl time.Duration) (string, domainauthn.AppToken, error) {
	id := uuid.NewString()
	raw, err := prefixedToken(id, 24)
	if err != nil {
		return "", domainauthn.AppToken{}, err
	}
	hash, err := s.hasher.Hash(raw)
	if err != nil {
		return "", domainauthn.AppToken{}, err
	}
	token := domainauthn.AppToken{
		ID:           id,
		Prefix:       id,
		UserID:       userID,
		Name:         strings.TrimSpace(name),
		Capabilities: append([]authz.Capability(nil), capabilities...),
		Hash:         hash,
		CreatedAt:    s.now().UTC(),
	}
	if ttl > 0 {
		expiresAt := s.now().UTC().Add(ttl)
		token.ExpiresAt = &expiresAt
	}
	if err := s.repo.SaveAppToken(ctx, token); err != nil {
		return "", domainauthn.AppToken{}, err
	}
	return raw, token, nil
}

func (s Service) RevokeAppToken(ctx context.Context, tokenID string) error {
	token, found, err := s.repo.GetAppTokenByPrefix(ctx, tokenID)
	if err != nil {
		return err
	}
	if !found {
		return nil
	}
	now := s.now().UTC()
	token.RevokedAt = &now
	return s.repo.SaveAppToken(ctx, token)
}

func (s Service) AuthenticateAppToken(ctx context.Context, raw string, remoteAddr string) (authz.Principal, bool, error) {
	prefix, _, ok := normalizePlaintextToken(raw)
	if !ok {
		return authz.Principal{}, false, nil
	}
	token, found, err := s.repo.GetAppTokenByPrefix(ctx, prefix)
	if err != nil {
		return authz.Principal{}, false, err
	}
	now := s.now().UTC()
	if !found || !token.Active(now) || !s.hasher.Verify(raw, token.Hash) {
		return authz.Principal{}, false, nil
	}
	user, ok, err := s.users.GetUser(ctx, token.UserID)
	if err != nil {
		return authz.Principal{}, false, err
	}
	if !ok || user.Status != domainusers.StatusActive {
		return authz.Principal{}, false, nil
	}
	token.LastUsedAt = &now
	token.LastUsedIP = strings.TrimSpace(remoteAddr)
	if err := s.repo.SaveAppToken(ctx, token); err != nil {
		return authz.Principal{}, false, err
	}
	principal := s.principalForUser(user)
	if len(token.Capabilities) > 0 {
		principal = authz.NewPrincipal(string(user.ID), token.Capabilities...)
	}
	return principal, true, nil
}

func (s Service) ListAppTokens(ctx context.Context, userID domainusers.ID) ([]domainauthn.AppToken, error) {
	return s.repo.ListAppTokens(ctx, userID)
}

func (s Service) authenticateLocal(ctx context.Context, input PasswordLoginInput, policy domainauthn.SessionPolicy) (PasswordLoginResult, error) {
	policy = policy.Normalized()
	key := attemptKey(input.Identifier, input.RemoteAddr)
	locked, err := s.lockedOut(ctx, key, policy)
	if err != nil {
		return PasswordLoginResult{}, err
	}
	if locked {
		_ = s.recordAttempt(ctx, key, domainusers.ID(""), input, false)
		return PasswordLoginResult{}, ErrLoginLocked
	}
	user, ok, err := s.lookupUser(ctx, input.Identifier)
	if err != nil {
		return PasswordLoginResult{}, err
	}
	if !ok || user.Status != domainusers.StatusActive || strings.TrimSpace(user.PasswordHash) == "" || !s.hasher.Verify(input.Password, user.PasswordHash) {
		_ = s.recordAttempt(ctx, key, domainusers.ID(""), input, false)
		return PasswordLoginResult{}, ErrInvalidCredentials
	}
	now := s.now().UTC()
	user.LastLoginAt = &now
	if err := s.users.SaveUser(ctx, user); err != nil {
		return PasswordLoginResult{}, err
	}
	if err := s.recordAttempt(ctx, key, user.ID, input, true); err != nil {
		return PasswordLoginResult{}, err
	}
	return PasswordLoginResult{
		User:               user,
		Principal:          s.principalForUser(user),
		ProviderID:         domainauthn.LocalProviderID,
		MustChangePassword: user.MustChangePassword,
	}, nil
}

func (s Service) lockedOut(ctx context.Context, key string, policy domainauthn.SessionPolicy) (bool, error) {
	if strings.TrimSpace(key) == "" {
		return false, nil
	}
	now := s.now().UTC()
	windowStart := now.Add(-policy.AttemptWindow)
	items, err := s.repo.ListLoginAttempts(ctx, key, windowStart)
	if err != nil {
		return false, err
	}
	attempts := make([]loginbrute.Attempt, 0, len(items))
	for _, item := range items {
		attempts = append(attempts, loginbrute.Attempt{Success: item.Success, CreatedAt: item.CreatedAt})
	}
	lb := loginbrute.Policy{
		MaxAttempts:   policy.MaxAttempts,
		AttemptWindow: policy.AttemptWindow,
		LockoutWindow: policy.LockoutWindow,
	}
	return loginbrute.LockedOut(now, lb, windowStart, attempts), nil
}

func (s Service) recordAttempt(ctx context.Context, key string, userID domainusers.ID, input PasswordLoginInput, success bool) error {
	return s.repo.SaveLoginAttempt(ctx, domainauthn.LoginAttempt{
		ID:         uuid.NewString(),
		Key:        key,
		UserID:     userID,
		Login:      strings.TrimSpace(strings.ToLower(input.Identifier)),
		RemoteAddr: strings.TrimSpace(input.RemoteAddr),
		UserAgent:  strings.TrimSpace(input.UserAgent),
		Success:    success,
		CreatedAt:  s.now().UTC(),
	})
}

func (s Service) lookupUser(ctx context.Context, identifier string) (domainusers.User, bool, error) {
	normalized := strings.TrimSpace(strings.ToLower(identifier))
	if normalized == "" {
		return domainusers.User{}, false, nil
	}
	users, err := s.users.ListUsers(ctx)
	if err != nil {
		return domainusers.User{}, false, err
	}
	for _, user := range users {
		if strings.EqualFold(strings.TrimSpace(user.Login), normalized) || strings.EqualFold(strings.TrimSpace(user.Email), normalized) {
			return user, true, nil
		}
	}
	return domainusers.User{}, false, nil
}

func (s Service) principalForUser(user domainusers.User) authz.Principal {
	capabilities := authz.ResolveRoleCapabilities(user.Roles)
	if len(capabilities) == 0 && slices.Contains(user.Roles, authz.RoleAdmin) {
		capabilities = append(capabilities, authz.ResolveRoleCapabilities([]string{authz.RoleAdmin})...)
	}
	return authz.NewPrincipal(string(user.ID), capabilities...)
}

func attemptKey(identifier string, remoteAddr string) string {
	return strings.TrimSpace(strings.ToLower(identifier)) + "|" + strings.TrimSpace(strings.ToLower(remoteAddr))
}

type LocalPasswordProvider struct {
	service Service
}

func (p LocalPasswordProvider) ID() string {
	return domainauthn.LocalProviderID
}

func (p LocalPasswordProvider) AuthenticatePassword(ctx context.Context, input PasswordLoginInput, policy domainauthn.SessionPolicy) (PasswordLoginResult, error) {
	return p.service.authenticateLocal(ctx, input, policy)
}
