package publicfixtures

import (
	"embed"
	"fmt"
	"strings"

	"github.com/fastygo/cms/internal/platform/locales"
	"github.com/fastygo/framework/pkg/web/i18n"
)

//go:embed */public.json
var fixtureFS embed.FS

// MetaFixture holds document-level public metadata.
type MetaFixture struct {
	Lang string `json:"lang"`
}

// SiteDefaultsFixture is used when settings do not override site strings.
type SiteDefaultsFixture struct {
	Title         string `json:"title"`
	BrandName     string `json:"brand_name"`
	HomeHeroTitle string `json:"home_hero_title"`
	HomeIntro     string `json:"home_intro"`
}

// RoutesFixture holds fixed route / screen copy.
type RoutesFixture struct {
	BlogTitle                string `json:"blog_title"`
	BlogDescription          string `json:"blog_description"`
	SearchTitle              string `json:"search_title"`
	SearchResultsBase        string `json:"search_results_base"`
	SearchResultsQuery       string `json:"search_results_query"`
	NotFoundTitle            string `json:"not_found_title"`
	NotFoundDescription      string `json:"not_found_description"`
	NotFoundAction           string `json:"not_found_action"`
	PublicRenderingDisabled  string `json:"public_rendering_disabled"`
	PublishedPrefix          string `json:"published_prefix"`
}

// BreadcrumbsFixture holds breadcrumb segment labels.
type BreadcrumbsFixture struct {
	Home   string `json:"home"`
	Blog   string `json:"blog"`
	Search string `json:"search"`
	Authors string `json:"authors"`
}

// AuthorFixture holds author block chrome strings.
type AuthorFixture struct {
	SectionTitle  string `json:"section_title"`
	ExpertTitle   string `json:"expert_title"`
	LinePrefix    string `json:"line_prefix"`
	PostsFallback string `json:"posts_fallback"`
}

// LabelsFixture holds printf-style templates.
type LabelsFixture struct {
	TaxonomyArchiveTemplate string `json:"taxonomy_archive_template"`
}

// ChromeFixture holds theme-visible chrome strings.
type ChromeFixture struct {
	EmptyArchive       string `json:"empty_archive"`
	OpenAction         string `json:"open_action"`
	ReadAction         string `json:"read_action"`
	RelatedTitle             string `json:"related_title"`
	RelatedInsightsTitle     string `json:"related_insights_title"`
	NoMenuItems              string `json:"no_menu_items"`
	QueryBadgeTemplate string `json:"query_badge_template"`
	ByAuthorPrefix     string `json:"by_author_prefix"`
	ThemePresetBadge   string `json:"theme_preset_badge"`
	FooterThanks       string `json:"footer_thanks"`
}

// PaginationFixture holds pagination control labels.
type PaginationFixture struct {
	Previous string `json:"previous"`
	Next     string `json:"next"`
}

// PublicSite is the root object under the "public" JSON key.
type PublicSite struct {
	Meta         MetaFixture         `json:"meta"`
	SiteDefaults SiteDefaultsFixture `json:"site_defaults"`
	Routes       RoutesFixture       `json:"routes"`
	Breadcrumbs  BreadcrumbsFixture  `json:"breadcrumbs"`
	Author       AuthorFixture       `json:"author"`
	Labels       LabelsFixture       `json:"labels"`
	Chrome       ChromeFixture       `json:"chrome"`
	Pagination   PaginationFixture   `json:"pagination"`
	Menus        map[string]string   `json:"menus"`
}

type bundle struct {
	Public PublicSite `json:"public"`
}

var store = i18n.New[bundle](fixtureFS, locales.Supported(), locales.DefaultForI18n(), func(reader i18n.Reader, loc string) (bundle, error) {
	b, err := i18n.DecodeSection[bundle](reader, loc, "public")
	if err != nil {
		return bundle{}, err
	}
	return b, nil
})

// Load returns the public-site copy bundle for a locale.
func Load(locale string) (PublicSite, error) {
	b, err := store.Load(locale)
	if err != nil {
		return PublicSite{}, err
	}
	return b.Public, nil
}

// MustLoad returns localized public copy, falling back through Supported().
func MustLoad(locale string) PublicSite {
	loc := locales.NormalizeOrDefault(locale)
	if pub, err := Load(loc); err == nil {
		return pub
	}
	if loc != locales.Default {
		if pub, err := Load(locales.Default); err == nil {
			return pub
		}
	}
	for _, fb := range locales.Supported() {
		if fb == loc {
			continue
		}
		if pub, err := Load(fb); err == nil {
			return pub
		}
	}
	return PublicSite{}
}

// MenuLabel returns a localized menu item label when menus[id] is set.
func (p PublicSite) MenuLabel(id string, stored string) string {
	if p.Menus != nil {
		if v := strings.TrimSpace(p.Menus[id]); v != "" {
			return v
		}
	}
	return stored
}

// ThemePresetBadge formats the theme + preset badge for the default theme header.
func (p PublicSite) ThemePresetBadge(themeID string, presetID string) string {
	tpl := strings.TrimSpace(p.Chrome.ThemePresetBadge)
	if tpl == "" {
		return fmt.Sprintf("Theme: %s | Preset: %s", themeID, presetID)
	}
	return fmt.Sprintf(tpl, themeID, presetID)
}
