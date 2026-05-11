package authn

import (
	"strings"
	"time"

	"github.com/fastygo/cms/internal/domain/authz"
	domainusers "github.com/fastygo/cms/internal/domain/users"
)

const (
	LocalProviderID    = "local"
	ExternalProviderID = "external"
)

type PasswordCredential struct {
	UserID             domainusers.ID `json:"user_id"`
	Hash               string         `json:"hash"`
	UpdatedAt          time.Time      `json:"updated_at"`
	MustChangePassword bool           `json:"must_change_password"`
	LastLoginAt        *time.Time     `json:"last_login_at,omitempty"`
}

type RecoveryCode struct {
	ID        string         `json:"id"`
	UserID    domainusers.ID `json:"user_id"`
	Label     string         `json:"label"`
	Hash      string         `json:"hash"`
	CreatedAt time.Time      `json:"created_at"`
	ExpiresAt *time.Time     `json:"expires_at,omitempty"`
	UsedAt    *time.Time     `json:"used_at,omitempty"`
}

type ResetToken struct {
	ID          string         `json:"id"`
	UserID      domainusers.ID `json:"user_id"`
	Label       string         `json:"label"`
	Hash        string         `json:"hash"`
	CreatedAt   time.Time      `json:"created_at"`
	ExpiresAt   time.Time      `json:"expires_at"`
	UsedAt      *time.Time     `json:"used_at,omitempty"`
	IssuedBy    string         `json:"issued_by,omitempty"`
	Bootstrap   bool           `json:"bootstrap"`
	Maintenance bool           `json:"maintenance"`
}

type AppToken struct {
	ID           string             `json:"id"`
	Prefix       string             `json:"prefix"`
	UserID       domainusers.ID     `json:"user_id"`
	Name         string             `json:"name"`
	Hash         string             `json:"hash"`
	Capabilities []authz.Capability `json:"capabilities"`
	CreatedAt    time.Time          `json:"created_at"`
	ExpiresAt    *time.Time         `json:"expires_at,omitempty"`
	RevokedAt    *time.Time         `json:"revoked_at,omitempty"`
	LastUsedAt   *time.Time         `json:"last_used_at,omitempty"`
	LastUsedIP   string             `json:"last_used_ip,omitempty"`
}

type LoginAttempt struct {
	ID         string         `json:"id"`
	Key        string         `json:"key"`
	UserID     domainusers.ID `json:"user_id,omitempty"`
	Login      string         `json:"login"`
	RemoteAddr string         `json:"remote_addr,omitempty"`
	UserAgent  string         `json:"user_agent,omitempty"`
	Success    bool           `json:"success"`
	CreatedAt  time.Time      `json:"created_at"`
}

type SessionPolicy struct {
	IdleTTL          time.Duration `json:"idle_ttl"`
	AbsoluteTTL      time.Duration `json:"absolute_ttl"`
	SecureCookies    bool          `json:"secure_cookies"`
	SameSite         string        `json:"same_site"`
	RotateAfterLogin bool          `json:"rotate_after_login"`
	MaxAttempts      int           `json:"max_attempts"`
	AttemptWindow    time.Duration `json:"attempt_window"`
	LockoutWindow    time.Duration `json:"lockout_window"`
}

func DefaultSessionPolicy(secureCookies bool) SessionPolicy {
	return SessionPolicy{
		IdleTTL:          12 * time.Hour,
		AbsoluteTTL:      7 * 24 * time.Hour,
		SecureCookies:    secureCookies,
		SameSite:         "lax",
		RotateAfterLogin: true,
		MaxAttempts:      3,
		AttemptWindow:    24 * time.Hour,
		LockoutWindow:    24 * time.Hour,
	}
}

func (p SessionPolicy) Normalized() SessionPolicy {
	if p.IdleTTL <= 0 {
		p.IdleTTL = 12 * time.Hour
	}
	if p.AbsoluteTTL <= 0 {
		p.AbsoluteTTL = 7 * 24 * time.Hour
	}
	if p.MaxAttempts <= 0 {
		p.MaxAttempts = 3
	}
	if p.AttemptWindow <= 0 {
		p.AttemptWindow = 24 * time.Hour
	}
	if p.LockoutWindow <= 0 {
		p.LockoutWindow = 24 * time.Hour
	}
	switch strings.TrimSpace(strings.ToLower(p.SameSite)) {
	case "strict", "none":
		p.SameSite = strings.ToLower(strings.TrimSpace(p.SameSite))
	default:
		p.SameSite = "lax"
	}
	return p
}

func (c RecoveryCode) Available(now time.Time) bool {
	if c.UsedAt != nil {
		return false
	}
	return c.ExpiresAt == nil || c.ExpiresAt.After(now)
}

func (t ResetToken) Available(now time.Time) bool {
	return t.UsedAt == nil && t.ExpiresAt.After(now)
}

func (t AppToken) Active(now time.Time) bool {
	if t.RevokedAt != nil {
		return false
	}
	return t.ExpiresAt == nil || t.ExpiresAt.After(now)
}
