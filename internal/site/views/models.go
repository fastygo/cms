package views

import (
	"github.com/a-h/templ"
	"github.com/fastygo/cms/internal/site/ui/blocks"
	"github.com/fastygo/cms/internal/site/ui/elements"
	"github.com/fastygo/framework/pkg/app"
	"github.com/fastygo/framework/pkg/web/view"
)

type LayoutData struct {
	Title    string
	Lang     string
	Brand    string
	Active   string
	NavItems []app.NavItem
	Account  elements.AccountActionsData
	Theme    view.ThemeToggleData
	Language view.LanguageToggleData
	Assets   AssetPaths
}

type AssetPaths struct {
	CSS          string
	ThemeJS      string
	AppJS        string
	PlaygroundJS string
}

type LoginPageData struct {
	Title         string
	Subtitle      string
	Lang          string
	Error         string
	ReturnTo      string
	ActionToken   string
	EmailLabel    string
	PasswordLabel string
	SubmitLabel   string
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
	Layout      LayoutData
	Title       string
	Cards       []blocks.StatCard
	Description string
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
	Layout      LayoutData
	Screen      string
	Form        blocks.ContentEditorData
	Title       string
	Description string
}
