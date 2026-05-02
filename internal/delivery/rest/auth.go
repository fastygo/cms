package rest

import (
	"net/http"
	"strings"
	"time"

	domainauthz "github.com/fastygo/cms/internal/domain/authz"
	"github.com/fastygo/framework/pkg/auth"
)

type SessionData struct {
	UserID       string   `json:"user_id"`
	Capabilities []string `json:"capabilities"`
}

type Authenticator struct {
	Session      auth.CookieSession[SessionData]
	BearerTokens map[string]domainauthz.Principal
}

func NewAuthenticator(sessionSecret string, bearerTokens map[string]domainauthz.Principal) Authenticator {
	return Authenticator{
		Session: auth.CookieSession[SessionData]{
			Name:     "gocms_session",
			Path:     "/",
			Secret:   sessionSecret,
			TTL:      12 * time.Hour,
			HTTPOnly: true,
			SameSite: http.SameSiteLaxMode,
		},
		BearerTokens: bearerTokens,
	}
}

func (a Authenticator) Principal(r *http.Request) (domainauthz.Principal, bool) {
	if principal, ok := a.bearerPrincipal(r.Header.Get("Authorization")); ok {
		return principal, true
	}
	if session, ok := a.Session.Read(r); ok && session.UserID != "" {
		capabilities := make([]domainauthz.Capability, 0, len(session.Capabilities))
		for _, capability := range session.Capabilities {
			capabilities = append(capabilities, domainauthz.Capability(capability))
		}
		return domainauthz.NewPrincipal(session.UserID, capabilities...), true
	}
	return domainauthz.Principal{}, false
}

func (a Authenticator) bearerPrincipal(header string) (domainauthz.Principal, bool) {
	const prefix = "bearer "
	if !strings.HasPrefix(strings.ToLower(header), prefix) {
		return domainauthz.Principal{}, false
	}
	token := strings.TrimSpace(header[len(prefix):])
	principal, ok := a.BearerTokens[token]
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
