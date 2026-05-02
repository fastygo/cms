package views

import (
	"github.com/a-h/templ"
	"github.com/fastygo/cms/internal/site/ui/blocks"
	"github.com/fastygo/cms/internal/site/ui/elements"
	"github.com/fastygo/framework/pkg/app"
)

type LayoutData struct {
	Title    string
	Lang     string
	Brand    string
	Active   string
	NavItems []app.NavItem
	Account  elements.AccountActionsData
	Assets   AssetPaths
}

type AssetPaths struct {
	CSS     string
	ThemeJS string
	AppJS   string
}

type LoginPageData struct {
	Title       string
	Subtitle    string
	Error       string
	ReturnTo    string
	ActionToken string
}

type ScreenData struct {
	Layout      LayoutData
	Screen      string
	Title       string
	Description string
	Actions     []elements.Action
	Body        templ.Component
}

type DashboardData struct {
	Layout LayoutData
	Cards  []blocks.StatCard
}

type ContentListPageData struct {
	Layout LayoutData
	Screen string
	Table  blocks.ContentTableData
}

type ContentEditPageData struct {
	Layout LayoutData
	Screen string
	Editor blocks.ContentEditorData
}

type SimpleListPageData struct {
	Layout LayoutData
	Screen string
	List   blocks.SimpleListData
}

type SettingsPageData struct {
	Layout LayoutData
	Screen string
	Form   blocks.ContentEditorData
}
