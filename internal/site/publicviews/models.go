package publicviews

import (
	"github.com/fastygo/cms/internal/site/assets"
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
	Theme    view.ThemeToggleData
	Language view.LanguageToggleData
	Assets   assets.Paths
}

type ArchiveItemData struct {
	Title       string
	Summary     string
	Href        string
	Meta        string
	ActionLabel string
}

type HomePageData struct {
	Layout      LayoutData
	Title       string
	Description string
	Items       []ArchiveItemData
	Pagination  elements.PaginationData
}

type ArchivePageData struct {
	Layout      LayoutData
	Screen      string
	Title       string
	Description string
	Items       []ArchiveItemData
	Pagination  elements.PaginationData
}

type ContentPageData struct {
	Layout       LayoutData
	Screen       string
	Title        string
	Description  string
	Content      string
	Published    string
	CanonicalURL string
}

type NotFoundPageData struct {
	Layout      LayoutData
	Title       string
	Description string
	ActionLabel string
}
