package authz

import "testing"

func TestRootIncludesDeclaredCoreCapabilities(t *testing.T) {
	root := Root()
	required := []Capability{
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
	}
	for _, capability := range required {
		if !root.Has(capability) {
			t.Fatalf("root principal does not include %q", capability)
		}
	}
}
