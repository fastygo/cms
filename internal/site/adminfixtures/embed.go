package adminfixtures

import (
	"embed"

	"github.com/fastygo/framework/pkg/web/i18n"
)

//go:embed en/*.json ru/*.json
var fixtureFS embed.FS

type ThemeFixture struct {
	Label              string `json:"label"`
	SwitchToDarkLabel  string `json:"switch_to_dark_label"`
	SwitchToLightLabel string `json:"switch_to_light_label"`
}

type LanguageFixture struct {
	Label        string            `json:"label"`
	CurrentLabel string            `json:"current_label"`
	NextLabel    string            `json:"next_label"`
	NextLocale   string            `json:"next_locale"`
	Available    []string          `json:"available"`
	LocaleLabels map[string]string `json:"locale_labels"`
}

type MetaFixture struct {
	Lang      string `json:"lang"`
	Title     string `json:"title"`
	BrandName string `json:"brand_name"`
}

type LoginFixture struct {
	Title                   string `json:"title"`
	Subtitle                string `json:"subtitle"`
	ErrorInvalidCredentials string `json:"error_invalid_credentials"`
}

type NavItemFixture struct {
	Label string `json:"label"`
	Path  string `json:"path"`
	Icon  string `json:"icon"`
	Order int    `json:"order"`
}

type DashboardCardFixture struct {
	Label string `json:"label"`
}

type DashboardFixture struct {
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Cards       []DashboardCardFixture `json:"cards"`
}

type ScreenFixture struct {
	Title           string `json:"title"`
	Description     string `json:"description"`
	Singular        string `json:"singular"`
	FormDescription string `json:"form_description"`
}

type FieldFixture struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Label       string   `json:"label"`
	Value       string   `json:"value"`
	Type        string   `json:"type"`
	Component   string   `json:"component"`
	Placeholder string   `json:"placeholder"`
	Required    bool     `json:"required"`
	Rows        int      `json:"rows"`
	Options     []string `json:"options"`
}

type FormFixture struct {
	Fields []FieldFixture `json:"fields"`
}

type PlaygroundFixture struct {
	ImportLabel       string `json:"import_label"`
	ExportLabel       string `json:"export_label"`
	ResetLabel        string `json:"reset_label"`
	ImportFromSource  string `json:"import_from_source"`
	RefreshFromSource string `json:"refresh_from_source"`
	ImportLimitHint   string `json:"import_limit_hint"`
	MissingBlobLabel  string `json:"missing_blob_label"`
}

type AdminFixture struct {
	Meta       MetaFixture              `json:"meta"`
	Theme      ThemeFixture             `json:"theme"`
	Language   LanguageFixture          `json:"language"`
	Login      LoginFixture             `json:"login"`
	Navigation []NavItemFixture         `json:"navigation"`
	Dashboard  DashboardFixture         `json:"dashboard"`
	Screens    map[string]ScreenFixture `json:"screens"`
	Forms      map[string]FormFixture   `json:"forms"`
	Labels     map[string]string        `json:"labels"`
	Playground PlaygroundFixture        `json:"playground"`
}

type Bundle struct {
	Admin AdminFixture `json:"admin"`
}

var Locales = []string{"en", "ru"}

var store = i18n.New[Bundle](fixtureFS, Locales, "en", func(reader i18n.Reader, loc string) (Bundle, error) {
	bundle, err := i18n.DecodeSection[Bundle](reader, loc, "admin")
	if err != nil {
		return Bundle{}, err
	}
	return bundle, nil
})

func Load(locale string) (Bundle, error) {
	return store.Load(locale)
}

func MustLoad(locale string) AdminFixture {
	admin, err := Load(locale)
	if err == nil {
		return admin.Admin
	}
	admin, err = Load("en")
	if err == nil {
		return admin.Admin
	}
	return AdminFixture{}
}

func (f AdminFixture) Label(key string, fallback string) string {
	if value, ok := f.Labels[key]; ok && value != "" {
		return value
	}
	return fallback
}

func (f AdminFixture) Screen(id string) (ScreenFixture, bool) {
	if f.Screens == nil {
		return ScreenFixture{}, false
	}
	screen, ok := f.Screens[id]
	return screen, ok
}

func (f AdminFixture) Form(id string) (FormFixture, bool) {
	if f.Forms == nil {
		return FormFixture{}, false
	}
	form, ok := f.Forms[id]
	return form, ok
}
