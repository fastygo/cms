package rest

import (
	"context"
	"net/http"
	"strings"
	"time"

	domainauthz "github.com/fastygo/cms/internal/domain/authz"
	"github.com/fastygo/framework/pkg/auth"
)

type SessionData struct {
	UserID             string   `json:"user_id"`
	Capabilities       []string `json:"capabilities"`
	ProviderID         string   `json:"provider_id,omitempty"`
	IssuedAtUnix       int64    `json:"issued_at_unix,omitempty"`
	AbsoluteExpiryUnix int64    `json:"absolute_expiry_unix,omitempty"`
	MustChangePassword bool     `json:"must_change_password,omitempty"`
}

type Authenticator struct {
	Session        auth.CookieSession[SessionData]
	BearerTokens   map[string]domainauthz.Principal
	BearerResolver BearerPrincipalResolver
	AbsoluteTTL    time.Duration
}

type AuthenticatorOptions struct {
	SessionName        string
	SessionPath        string
	SessionTTL         time.Duration
	SessionSecure      bool
	SessionSameSite    http.SameSite
	AbsoluteSessionTTL time.Duration
	BearerResolver     BearerPrincipalResolver
}

type BearerPrincipalResolver interface {
	AuthenticateAppToken(context.Context, string, string) (domainauthz.Principal, bool, error)
}

func NewAuthenticator(sessionSecret string, bearerTokens map[string]domainauthz.Principal) Authenticator {
	return NewAuthenticatorWithOptions(sessionSecret, bearerTokens, AuthenticatorOptions{})
}

func NewAuthenticatorWithOptions(sessionSecret string, bearerTokens map[string]domainauthz.Principal, options AuthenticatorOptions) Authenticator {
	if options.SessionName == "" {
		options.SessionName = "gocms_session"
	}
	if options.SessionPath == "" {
		options.SessionPath = "/"
	}
	if options.SessionTTL <= 0 {
		options.SessionTTL = 12 * time.Hour
	}
	if options.SessionSameSite == 0 {
		options.SessionSameSite = http.SameSiteLaxMode
	}
	return Authenticator{
		Session: auth.CookieSession[SessionData]{
			Name:     options.SessionName,
			Path:     options.SessionPath,
			Secret:   sessionSecret,
			TTL:      options.SessionTTL,
			HTTPOnly: true,
			Secure:   options.SessionSecure,
			SameSite: options.SessionSameSite,
		},
		BearerTokens:   bearerTokens,
		BearerResolver: options.BearerResolver,
		AbsoluteTTL:    options.AbsoluteSessionTTL,
	}
}

func (a Authenticator) Principal(r *http.Request) (domainauthz.Principal, bool) {
	if principal, ok := a.bearerPrincipal(r); ok {
		return principal, true
	}
	if session, ok := a.Session.Read(r); ok && session.UserID != "" {
		if session.AbsoluteExpiryUnix > 0 && time.Now().UTC().Unix() > session.AbsoluteExpiryUnix {
			return domainauthz.Principal{}, false
		}
		capabilities := make([]domainauthz.Capability, 0, len(session.Capabilities))
		for _, capability := range session.Capabilities {
			capabilities = append(capabilities, domainauthz.Capability(capability))
		}
		return domainauthz.NewPrincipal(session.UserID, capabilities...), true
	}
	return domainauthz.Principal{}, false
}

func (a Authenticator) Issue(w http.ResponseWriter, session SessionData) error {
	now := time.Now().UTC()
	if session.IssuedAtUnix == 0 {
		session.IssuedAtUnix = now.Unix()
	}
	if session.AbsoluteExpiryUnix == 0 && a.AbsoluteTTL > 0 {
		session.AbsoluteExpiryUnix = now.Add(a.AbsoluteTTL).Unix()
	}
	return a.Session.Issue(w, session)
}

func (a Authenticator) bearerPrincipal(r *http.Request) (domainauthz.Principal, bool) {
	const prefix = "bearer "
	header := r.Header.Get("Authorization")
	if !strings.HasPrefix(strings.ToLower(header), prefix) {
		return domainauthz.Principal{}, false
	}
	token := strings.TrimSpace(header[len(prefix):])
	principal, ok := a.BearerTokens[token]
	if ok {
		return principal, true
	}
	if a.BearerResolver == nil {
		return domainauthz.Principal{}, false
	}
	principal, ok, err := a.BearerResolver.AuthenticateAppToken(r.Context(), token, r.RemoteAddr)
	if err != nil {
		return domainauthz.Principal{}, false
	}
	return principal, ok
}

func DevBearerPrincipals() map[string]domainauthz.Principal {
	return map[string]domainauthz.Principal{
		"admin-token": domainauthz.NewPrincipal("admin",
			domainauthz.CapabilityControlPanelAccess,
			domainauthz.CapabilityContentCreate,
			domainauthz.CapabilityContentReadPrivate,
			domainauthz.CapabilityContentEdit,
			domainauthz.CapabilityContentPublish,
			domainauthz.CapabilityContentSchedule,
			domainauthz.CapabilityContentDelete,
			domainauthz.CapabilityContentRestore,
			domainauthz.CapabilityMediaUpload,
			domainauthz.CapabilityMediaEdit,
			domainauthz.CapabilityTaxonomiesManage,
			domainauthz.CapabilityTaxonomiesAssign,
			domainauthz.CapabilityMenusManage,
			domainauthz.CapabilitySettingsManage,
			domainauthz.CapabilityThemesManage,
			domainauthz.CapabilityUsersManage,
			domainauthz.CapabilityRolesManage,
		),
		"editor-token": domainauthz.NewPrincipal("author-1",
			domainauthz.CapabilityControlPanelAccess,
			domainauthz.CapabilityContentCreate,
			domainauthz.CapabilityContentReadPrivate,
			domainauthz.CapabilityContentEditOwn,
			domainauthz.CapabilityContentPublish,
			domainauthz.CapabilityContentSchedule,
			domainauthz.CapabilityContentDelete,
			domainauthz.CapabilityContentRestore,
			domainauthz.CapabilityMediaUpload,
			domainauthz.CapabilityMediaEdit,
			domainauthz.CapabilityTaxonomiesAssign,
		),
		"viewer-token": domainauthz.NewPrincipal("viewer",
			domainauthz.CapabilityControlPanelAccess,
			domainauthz.CapabilityContentReadPrivate,
		),
	}
}
