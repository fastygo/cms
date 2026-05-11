package authz

import "slices"

// Capability identifies one granular permission.
type Capability string

const (
	CapabilityControlPanelAccess Capability = "control_panel.access"
	CapabilityContentCreate      Capability = "content.create"
	CapabilityContentReadPrivate Capability = "content.read_private"
	CapabilityContentEdit        Capability = "content.edit"
	CapabilityContentEditOwn     Capability = "content.edit_own"
	CapabilityContentEditOthers  Capability = "content.edit_others"
	CapabilityContentPublish     Capability = "content.publish"
	CapabilityContentSchedule    Capability = "content.schedule"
	CapabilityContentDelete      Capability = "content.delete"
	CapabilityContentRestore     Capability = "content.restore"
	CapabilityMediaUpload        Capability = "media.upload"
	CapabilityMediaEdit          Capability = "media.edit"
	CapabilityTaxonomiesManage   Capability = "taxonomies.manage"
	CapabilityTaxonomiesAssign   Capability = "taxonomies.assign"
	CapabilityMenusManage        Capability = "menus.manage"
	CapabilitySettingsManage     Capability = "settings.manage"
	CapabilityUsersManage        Capability = "users.manage"
	CapabilityRolesManage        Capability = "roles.manage"
	CapabilityPluginsManage      Capability = "plugins.manage"
	CapabilityThemesManage       Capability = "themes.manage"
	CapabilityPrivateAPIRead     Capability = "api.read_private"
)

// Principal is the actor asking to perform an operation.
type Principal struct {
	ID           string
	Capabilities map[Capability]struct{}
}

// NewPrincipal creates a principal with the supplied capabilities.
func NewPrincipal(id string, capabilities ...Capability) Principal {
	p := Principal{ID: id, Capabilities: make(map[Capability]struct{}, len(capabilities))}
	for _, capability := range capabilities {
		p.Capabilities[capability] = struct{}{}
	}
	return p
}

// Has reports whether the principal has a capability.
func (p Principal) Has(capability Capability) bool {
	_, ok := p.Capabilities[capability]
	return ok
}

// Root returns a synthetic principal with all currently declared core capabilities.
func Root() Principal {
	return NewPrincipal("root",
		CapabilityControlPanelAccess,
		CapabilityContentCreate,
		CapabilityContentReadPrivate,
		CapabilityContentEdit,
		CapabilityContentEditOwn,
		CapabilityContentEditOthers,
		CapabilityContentPublish,
		CapabilityContentSchedule,
		CapabilityContentDelete,
		CapabilityContentRestore,
		CapabilityMediaUpload,
		CapabilityMediaEdit,
		CapabilityTaxonomiesManage,
		CapabilityTaxonomiesAssign,
		CapabilityMenusManage,
		CapabilitySettingsManage,
		CapabilityUsersManage,
		CapabilityRolesManage,
		CapabilityPluginsManage,
		CapabilityThemesManage,
		CapabilityPrivateAPIRead,
	)
}

// Role is a named capability set.
type Role struct {
	ID           string
	Label        string
	Capabilities []Capability
}

const (
	RoleAdmin  = "admin"
	RoleEditor = "editor"
	RoleViewer = "viewer"
)

func BuiltInRoles() []Role {
	return []Role{
		{
			ID:    RoleAdmin,
			Label: "Administrator",
			Capabilities: []Capability{
				CapabilityControlPanelAccess,
				CapabilityContentCreate,
				CapabilityContentReadPrivate,
				CapabilityContentEdit,
				CapabilityContentEditOwn,
				CapabilityContentEditOthers,
				CapabilityContentPublish,
				CapabilityContentSchedule,
				CapabilityContentDelete,
				CapabilityContentRestore,
				CapabilityMediaUpload,
				CapabilityMediaEdit,
				CapabilityTaxonomiesManage,
				CapabilityTaxonomiesAssign,
				CapabilityMenusManage,
				CapabilitySettingsManage,
				CapabilityUsersManage,
				CapabilityRolesManage,
				CapabilityPluginsManage,
				CapabilityThemesManage,
				CapabilityPrivateAPIRead,
			},
		},
		{
			ID:    RoleEditor,
			Label: "Editor",
			Capabilities: []Capability{
				CapabilityControlPanelAccess,
				CapabilityContentCreate,
				CapabilityContentReadPrivate,
				CapabilityContentEditOwn,
				CapabilityContentPublish,
				CapabilityContentSchedule,
				CapabilityContentDelete,
				CapabilityContentRestore,
				CapabilityMediaUpload,
				CapabilityMediaEdit,
				CapabilityTaxonomiesAssign,
				CapabilityPrivateAPIRead,
			},
		},
		{
			ID:    RoleViewer,
			Label: "Viewer",
			Capabilities: []Capability{
				CapabilityControlPanelAccess,
				CapabilityContentReadPrivate,
			},
		},
	}
}

func ResolveRoleCapabilities(roleIDs []string) []Capability {
	resolved := []Capability{}
	known := map[string][]Capability{}
	for _, role := range BuiltInRoles() {
		known[role.ID] = role.Capabilities
	}
	for _, roleID := range roleIDs {
		for _, capability := range known[roleID] {
			if !slices.Contains(resolved, capability) {
				resolved = append(resolved, capability)
			}
		}
	}
	return resolved
}
