package publicsite

import (
	"net/http"
	"strconv"
	"strings"

	appcontent "github.com/fastygo/cms/internal/application/content"
	appmedia "github.com/fastygo/cms/internal/application/media"
	appmenus "github.com/fastygo/cms/internal/application/menus"
	"github.com/fastygo/cms/internal/application/publicrender"
	appsettings "github.com/fastygo/cms/internal/application/settings"
	apptaxonomy "github.com/fastygo/cms/internal/application/taxonomy"
	appusers "github.com/fastygo/cms/internal/application/users"
	domainsettings "github.com/fastygo/cms/internal/domain/settings"
	"github.com/fastygo/cms/internal/platform/locales"
	"github.com/fastygo/cms/internal/platform/permalinks"
	platformplugins "github.com/fastygo/cms/internal/platform/plugins"
	platformthemes "github.com/fastygo/cms/internal/platform/themes"
	"github.com/fastygo/cms/internal/site/publicfixtures"
	"github.com/fastygo/framework/pkg/web"
)

type Services struct {
	Content  appcontent.Service
	Media    appmedia.Service
	Menus    appmenus.Service
	Settings appsettings.Service
	Taxonomy apptaxonomy.Service
	Users    appusers.Service
}

type Handler struct {
	services  Services
	themes    *platformthemes.Registry
	assembler pageAssembler
	registry  *platformplugins.Registry
}

func New(services Services, themes *platformthemes.Registry) Handler {
	return NewWithRegistry(services, themes, nil)
}

func NewWithRegistry(services Services, themes *platformthemes.Registry, registry *platformplugins.Registry) Handler {
	if themes == nil {
		themes = platformthemes.DefaultRegistry()
	}
	return Handler{
		services:  services,
		themes:    themes,
		assembler: newPageAssembler(services, themes),
		registry:  registry,
	}
}

func (h Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/", h.handle)
}

func (h Handler) handle(w http.ResponseWriter, r *http.Request) {
	config := h.loadConfig(r)
	if strings.EqualFold(config.PublicRendering, "disabled") {
		pub := publicfixtures.MustLoad(locales.FromContext(r.Context()))
		h.renderPage(w, r, h.assembler.notFound(r, config, pub.Routes.PublicRenderingDisabled), http.StatusNotFound)
		return
	}

	candidates := permalinks.Resolve(r.URL.Path, r.URL.Query(), config.Permalinks)
	for _, candidate := range candidates {
		switch candidate.Kind {
		case permalinks.CandidateHome:
			request, err := h.assembler.home(r, config)
			if err != nil {
				http.Error(w, "Unable to load public content.", http.StatusInternalServerError)
				return
			}
			h.renderPage(w, r, request, http.StatusOK)
			return
		case permalinks.CandidateBlog:
			request, err := h.assembler.blog(r, config)
			if err != nil {
				http.Error(w, "Unable to load blog archive.", http.StatusInternalServerError)
				return
			}
			h.renderPage(w, r, request, http.StatusOK)
			return
		case permalinks.CandidateSearch:
			request, err := h.assembler.search(r, config, candidate)
			if err != nil {
				http.Error(w, "Unable to load search results.", http.StatusInternalServerError)
				return
			}
			h.renderPage(w, r, request, http.StatusOK)
			return
		case permalinks.CandidateTaxonomy:
			request, ok, err := h.assembler.taxonomy(r, config, candidate)
			if err != nil {
				http.Error(w, "Unable to load taxonomy archive.", http.StatusInternalServerError)
				return
			}
			if ok {
				h.renderPage(w, r, request, http.StatusOK)
				return
			}
		case permalinks.CandidateAuthor:
			request, ok, err := h.assembler.author(r, config, candidate)
			if err != nil {
				http.Error(w, "Unable to load author archive.", http.StatusInternalServerError)
				return
			}
			if ok {
				h.renderPage(w, r, request, http.StatusOK)
				return
			}
		case permalinks.CandidatePageSlug:
			request, ok, err := h.assembler.page(r, config, candidate.Slug)
			if err != nil {
				http.Error(w, "Unable to load public page.", http.StatusInternalServerError)
				return
			}
			if ok {
				h.renderPage(w, r, request, http.StatusOK)
				return
			}
		case permalinks.CandidatePostSlug, permalinks.CandidatePostID:
			request, ok, err := h.assembler.post(r, config, candidate)
			if err != nil {
				http.Error(w, "Unable to load public post.", http.StatusInternalServerError)
				return
			}
			if ok {
				h.renderPage(w, r, request, http.StatusOK)
				return
			}
		}
	}

	h.renderPage(w, r, h.assembler.notFound(r, config, ""), http.StatusNotFound)
}

func (h Handler) renderPage(w http.ResponseWriter, r *http.Request, request publicrender.RenderRequest, status int) {
	if request.Page.Content != "" {
		filtered, err := platformplugins.FilterValue(r.Context(), h.registry, "render.content.filter", platformplugins.HookContext{
			Surface: pluginsSurfacePublic(),
			Path:    r.URL.Path,
			Locale:  locales.FromContext(r.Context()),
			Metadata: map[string]any{
				"screen": string(request.Page.Screen),
				"kind":   string(request.Page.Kind),
			},
		}, request.Page.Content)
		if err != nil {
			http.Error(w, "Unable to render public page.", http.StatusInternalServerError)
			return
		}
		request.Page.Content = filtered
	}
	component, err := h.themes.ResolveActive(string(request.Context.ThemeID)).Render(r.Context(), request)
	if err != nil {
		http.Error(w, "Unable to render public page.", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(status)
	_ = web.Render(r.Context(), w, component)
}

func (h Handler) loadConfig(r *http.Request) publicrender.SiteConfig {
	pub := publicfixtures.MustLoad(locales.FromContext(r.Context()))
	config := publicrender.SiteConfig{
		Title:           pub.SiteDefaults.Title,
		BrandName:       pub.SiteDefaults.BrandName,
		HomeHeroTitle:   pub.SiteDefaults.HomeHeroTitle,
		HomeIntro:       pub.SiteDefaults.HomeIntro,
		ActiveTheme:     string(platformthemes.DefaultThemeID),
		StylePreset:     "default",
		PublicRendering: "enabled",
		Permalinks:      permalinks.NormalizeSettings(permalinks.Settings{}),
	}
	if value, ok, err := h.services.Settings.Get(r.Context(), domainsettings.Key("site.title")); err == nil && ok {
		config.Title = settingString(value, config.Title)
	}
	if value, ok, err := h.services.Settings.Get(r.Context(), domainsettings.Key("theme.brand_name")); err == nil && ok {
		config.BrandName = settingString(value, config.BrandName)
	}
	if value, ok, err := h.services.Settings.Get(r.Context(), domainsettings.Key("theme.home_intro")); err == nil && ok {
		config.HomeIntro = settingString(value, config.HomeIntro)
	}
	if value, ok, err := h.services.Settings.Get(r.Context(), domainsettings.Key("theme.home_hero_title")); err == nil && ok {
		config.HomeHeroTitle = settingString(value, config.HomeHeroTitle)
	}
	if value, ok, err := h.services.Settings.Get(r.Context(), domainsettings.Key(platformthemes.ActiveThemeKey)); err == nil && ok {
		config.ActiveTheme = settingString(value, config.ActiveTheme)
	}
	if value, ok, err := h.services.Settings.Get(r.Context(), domainsettings.Key(platformthemes.StylePresetKey)); err == nil && ok {
		config.StylePreset = settingString(value, config.StylePreset)
	}
	if value, ok, err := h.services.Settings.Get(r.Context(), domainsettings.Key("public.rendering")); err == nil && ok {
		config.PublicRendering = firstNonEmpty(settingString(value, config.PublicRendering), config.PublicRendering)
	}
	if value, ok, err := h.services.Settings.Get(r.Context(), domainsettings.Key("permalinks.post_pattern")); err == nil && ok {
		config.Permalinks.PostPattern = settingString(value, config.Permalinks.PostPattern)
	}
	if value, ok, err := h.services.Settings.Get(r.Context(), domainsettings.Key("permalinks.page_pattern")); err == nil && ok {
		config.Permalinks.PagePattern = settingString(value, config.Permalinks.PagePattern)
	}
	if previewTheme := strings.TrimSpace(r.URL.Query().Get("preview_theme")); previewTheme != "" {
		config.ActiveTheme = previewTheme
	}
	if previewPreset := strings.TrimSpace(r.URL.Query().Get("preview_preset")); previewPreset != "" {
		config.StylePreset = previewPreset
	}
	config.Permalinks = permalinks.NormalizeSettings(config.Permalinks)
	config.ActiveTheme = string(h.themes.ResolveActive(config.ActiveTheme).Manifest().ID)
	config.StylePreset = h.themes.ResolvePreset(config.ActiveTheme, config.StylePreset).ID
	return config
}

func parsePage(r *http.Request, fallback int) int {
	parsed, err := strconv.Atoi(strings.TrimSpace(r.URL.Query().Get("page")))
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}

func settingString(value domainsettings.Value, fallback string) string {
	if raw, ok := value.Value.(string); ok && strings.TrimSpace(raw) != "" {
		return raw
	}
	return fallback
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func isResolvableMiss(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "not found") || strings.Contains(message, "not public")
}

func pluginsSurfacePublic() platformplugins.Surface {
	return platformplugins.SurfacePublic
}
