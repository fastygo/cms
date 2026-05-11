package publicrender

import (
	"github.com/fastygo/cms/internal/domain/themes"
	"github.com/fastygo/cms/internal/platform/permalinks"
	"github.com/fastygo/cms/internal/site/assets"
	"github.com/fastygo/framework/pkg/web/view"
)

type RenderKind string

const (
	RenderKindHome     RenderKind = "home"
	RenderKindBlog     RenderKind = "blog"
	RenderKindArchive  RenderKind = "archive"
	RenderKindTaxonomy RenderKind = "taxonomy"
	RenderKindAuthor   RenderKind = "author"
	RenderKindSearch   RenderKind = "search"
	RenderKindPost     RenderKind = "post"
	RenderKindPage     RenderKind = "page"
	RenderKindNotFound RenderKind = "not_found"
)

type SiteConfig struct {
	Title           string
	BrandName       string
	HomeHeroTitle   string
	HomeIntro       string
	ActiveTheme     string
	StylePreset     string
	PublicRendering string
	Permalinks      permalinks.Settings
}

type RenderContext struct {
	Path        string
	Query       string
	Locale      string
	ThemeID     themes.ThemeID
	StylePreset string
}

type RenderRequest struct {
	Context RenderContext
	Page    PublicPage
}

type PublicPage struct {
	Kind                  RenderKind
	Screen                string
	TemplateRole          themes.TemplateRole
	Layout                Layout
	DocumentTitle         string
	Title                 string
	Description           string
	Intro                 string
	Query                 string
	QueryBadge            string
	ActionLabel           string
	Content               string
	CanonicalURL          string
	Published             string
	AuthorAttributionLine string
	Breadcrumbs           []Breadcrumb
	Pagination            Pagination
	SEO                   SEOModel
	Items                 []ArchiveItem
	RelatedItems          []ArchiveItem
	Featured              *MediaView
	Author                *AuthorView
	CurrentTerm           *TermView
	Terms                 []TermView
}

type Layout struct {
	Title         string
	Lang          string
	Brand         string
	Tagline       string
	ActivePath    string
	ThemeID       themes.ThemeID
	StylePresetID string
	HeaderMenu    []MenuItem
	FooterMenu    []MenuItem
	Assets        AssetBundle
	ThemeToggle   view.ThemeToggleData
	Language      view.LanguageToggleData
	Chrome        PublicChrome
}

// PublicChrome carries localized strings for the default public theme shell.
type PublicChrome struct {
	EmptyArchive         string
	OpenAction           string
	ReadAction           string
	RelatedTitle         string
	RelatedInsightsTitle string
	AuthorSectionTitle   string
	AuthorExpertTitle    string
	NoMenuItems          string
	ThemePresetBadge     string
	ByAuthorPrefix       string
	AuthorLinePrefix     string
	FooterThanks         string
}

type AssetBundle struct {
	Base        assets.Paths
	Stylesheets []string
	Scripts     []string
}

type MenuItem struct {
	Label    string
	URL      string
	Active   bool
	Children []MenuItem
}

type Breadcrumb struct {
	Label string
	URL   string
}

type Pagination struct {
	Page          int
	TotalPages    int
	BaseHref      string
	PreviousLabel string
	NextLabel     string
}

type SEOModel struct {
	Title        string
	Description  string
	CanonicalURL string
	Type         string
	ImageURL     string
	PublishedAt  string
	ModifiedAt   string
	NoIndex      bool
}

type MediaView struct {
	ID      string
	URL     string
	AltText string
	Caption string
	Width   int
	Height  int
}

type AuthorView struct {
	ID          string
	Slug        string
	DisplayName string
	Bio         string
	AvatarURL   string
	WebsiteURL  string
}

type TermView struct {
	Taxonomy string
	ID       string
	Slug     string
	Label    string
}

type ArchiveItem struct {
	ID       string
	Kind     string
	Title    string
	Summary  string
	Href     string
	Meta     string
	Author   *AuthorView
	Featured *MediaView
	Terms    []TermView
}
