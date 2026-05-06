package publicsite

import (
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/fastygo/cms/internal/application/publicrender"
	"github.com/fastygo/cms/internal/domain/authz"
	domaincontent "github.com/fastygo/cms/internal/domain/content"
	domainmedia "github.com/fastygo/cms/internal/domain/media"
	domainmenus "github.com/fastygo/cms/internal/domain/menus"
	domainthemes "github.com/fastygo/cms/internal/domain/themes"
	domaintaxonomy "github.com/fastygo/cms/internal/domain/taxonomy"
	domainusers "github.com/fastygo/cms/internal/domain/users"
	"github.com/fastygo/cms/internal/platform/permalinks"
	platformthemes "github.com/fastygo/cms/internal/platform/themes"
	"github.com/fastygo/cms/internal/site/adminfixtures"
	"github.com/fastygo/cms/internal/site/assets"
	"github.com/fastygo/framework/pkg/web/locale"
	"github.com/fastygo/framework/pkg/web/view"
)

type pageAssembler struct {
	services Services
	themes   *platformthemes.Registry
}

func newPageAssembler(services Services, themes *platformthemes.Registry) pageAssembler {
	return pageAssembler{services: services, themes: themes}
}

func (a pageAssembler) home(r *http.Request, config publicrender.SiteConfig) (publicrender.RenderRequest, error) {
	page := parsePage(r, 1)
	result, err := a.services.Content.List(r.Context(), domaincontent.Query{
		Kinds:      []domaincontent.Kind{domaincontent.KindPost},
		PublicOnly: true,
		Page:       page,
		PerPage:    6,
		SortBy:     domaincontent.SortPublishedAt,
		SortDesc:   true,
	})
	if err != nil {
		return publicrender.RenderRequest{}, err
	}
	return a.renderRequest(r, config, publicrender.PublicPage{
		Kind:         publicrender.RenderKindHome,
		Screen:       "home",
		TemplateRole: "front",
		Title:        config.Title,
		Description:  config.HomeIntro,
		Intro:        config.HomeIntro,
		Pagination:   paginationData("/", page, result.TotalPages),
		Items:        a.archiveItems(r, result.Items, config.Permalinks),
		SEO:          seoModel(config.Title, config.HomeIntro, "/", "website", nil, time.Time{}, time.Time{}, false),
	}, "/")
}

func (a pageAssembler) blog(r *http.Request, config publicrender.SiteConfig) (publicrender.RenderRequest, error) {
	page := parsePage(r, 1)
	result, err := a.services.Content.List(r.Context(), domaincontent.Query{
		Kinds:      []domaincontent.Kind{domaincontent.KindPost},
		PublicOnly: true,
		Page:       page,
		PerPage:    10,
		SortBy:     domaincontent.SortPublishedAt,
		SortDesc:   true,
	})
	if err != nil {
		return publicrender.RenderRequest{}, err
	}
	return a.renderRequest(r, config, publicrender.PublicPage{
		Kind:         publicrender.RenderKindBlog,
		Screen:       "blog",
		TemplateRole: "archive",
		Title:        "Blog",
		Description:  "Latest published posts.",
		Pagination:   paginationData("/blog/", page, result.TotalPages),
		Breadcrumbs:  []publicrender.Breadcrumb{{Label: "Home", URL: "/"}, {Label: "Blog", URL: "/blog/"}},
		Items:        a.archiveItems(r, result.Items, config.Permalinks),
		SEO:          seoModel("Blog", "Latest published posts.", "/blog/", "website", nil, time.Time{}, time.Time{}, false),
	}, "/blog/")
}

func (a pageAssembler) search(r *http.Request, config publicrender.SiteConfig, candidate permalinks.Candidate) (publicrender.RenderRequest, error) {
	page := parsePage(r, 1)
	result, err := a.services.Content.List(r.Context(), domaincontent.Query{
		Kinds:      []domaincontent.Kind{domaincontent.KindPost, domaincontent.KindPage},
		PublicOnly: true,
		Search:     candidate.Query,
		Page:       page,
		PerPage:    12,
		SortBy:     domaincontent.SortPublishedAt,
		SortDesc:   true,
	})
	if err != nil {
		return publicrender.RenderRequest{}, err
	}
	description := "Search results"
	if strings.TrimSpace(candidate.Query) != "" {
		description = `Results for "` + candidate.Query + `"`
	}
	canonical := "/search/"
	if strings.TrimSpace(candidate.Query) != "" {
		canonical += "?q=" + url.QueryEscape(candidate.Query)
	}
	return a.renderRequest(r, config, publicrender.PublicPage{
		Kind:         publicrender.RenderKindSearch,
		Screen:       "search",
		TemplateRole: "search",
		Title:        "Search",
		Description:  description,
		Query:        candidate.Query,
		Pagination:   paginationData(searchBase(candidate.Query), page, result.TotalPages),
		Breadcrumbs:  []publicrender.Breadcrumb{{Label: "Home", URL: "/"}, {Label: "Search", URL: canonical}},
		Items:        a.archiveItems(r, result.Items, config.Permalinks),
		SEO:          seoModel("Search", description, canonical, "website", nil, time.Time{}, time.Time{}, true),
	}, "/search/")
}

func (a pageAssembler) taxonomy(r *http.Request, config publicrender.SiteConfig, candidate permalinks.Candidate) (publicrender.RenderRequest, bool, error) {
	term, ok := a.findTerm(r, domaintaxonomy.Type(candidate.Taxonomy), candidate.Slug)
	if !ok {
		return publicrender.RenderRequest{}, false, nil
	}
	page := parsePage(r, 1)
	result, err := a.services.Content.List(r.Context(), domaincontent.Query{
		Kinds:      []domaincontent.Kind{domaincontent.KindPost},
		PublicOnly: true,
		Taxonomy:   candidate.Taxonomy,
		TermID:     string(term.ID),
		Page:       page,
		PerPage:    12,
		SortBy:     domaincontent.SortPublishedAt,
		SortDesc:   true,
	})
	if err != nil {
		return publicrender.RenderRequest{}, true, err
	}
	title := term.Name.Value(locale.From(r.Context()), "en")
	description := term.Description.Value(locale.From(r.Context()), "en")
	if description == "" {
		description = "Archive for " + candidate.Taxonomy + ": " + title
	}
	path := candidate.Path
	request, err := a.renderRequest(r, config, publicrender.PublicPage{
		Kind:         publicrender.RenderKindTaxonomy,
		Screen:       "taxonomy",
		TemplateRole: "taxonomy",
		Title:        title,
		Description:  description,
		CurrentTerm:  a.termView(term),
		Pagination:   paginationData(path, page, result.TotalPages),
		Breadcrumbs:  []publicrender.Breadcrumb{{Label: "Home", URL: "/"}, {Label: title, URL: path}},
		Items:        a.archiveItems(r, result.Items, config.Permalinks),
		SEO:          seoModel(title, description, path, "website", nil, time.Time{}, time.Time{}, false),
	}, path)
	return request, true, err
}

func (a pageAssembler) author(r *http.Request, config publicrender.SiteConfig, candidate permalinks.Candidate) (publicrender.RenderRequest, bool, error) {
	author, ok := a.findAuthorBySlug(r, candidate.Slug)
	if !ok {
		return publicrender.RenderRequest{}, false, nil
	}
	page := parsePage(r, 1)
	result, err := a.services.Content.List(r.Context(), domaincontent.Query{
		Kinds:      []domaincontent.Kind{domaincontent.KindPost},
		PublicOnly: true,
		AuthorID:   author.ID,
		Page:       page,
		PerPage:    10,
		SortBy:     domaincontent.SortPublishedAt,
		SortDesc:   true,
	})
	if err != nil {
		return publicrender.RenderRequest{}, true, err
	}
	description := author.Bio
	if strings.TrimSpace(description) == "" {
		description = "Posts by " + author.DisplayName
	}
	path := candidate.Path
	request, err := a.renderRequest(r, config, publicrender.PublicPage{
		Kind:         publicrender.RenderKindAuthor,
		Screen:       "author",
		TemplateRole: "author",
		Title:        author.DisplayName,
		Description:  description,
		Author:       author,
		Pagination:   paginationData(path, page, result.TotalPages),
		Breadcrumbs:  []publicrender.Breadcrumb{{Label: "Home", URL: "/"}, {Label: "Authors", URL: "/author/"}, {Label: author.DisplayName, URL: path}},
		Items:        a.archiveItems(r, result.Items, config.Permalinks),
		SEO:          seoModel(author.DisplayName, description, path, "profile", nil, time.Time{}, time.Time{}, false),
	}, path)
	return request, true, err
}

func (a pageAssembler) page(r *http.Request, config publicrender.SiteConfig, slug string) (publicrender.RenderRequest, bool, error) {
	entry, err := a.services.Content.GetBySlug(r.Context(), authz.Principal{}, domaincontent.KindPage, slug, locale.From(r.Context()))
	if err != nil {
		if isResolvableMiss(err) {
			return publicrender.RenderRequest{}, false, nil
		}
		return publicrender.RenderRequest{}, true, err
	}
	return a.content(r, config, entry, publicrender.RenderKindPage, "page", "page"), true, nil
}

func (a pageAssembler) post(r *http.Request, config publicrender.SiteConfig, candidate permalinks.Candidate) (publicrender.RenderRequest, bool, error) {
	var (
		entry domaincontent.Entry
		err   error
	)
	switch candidate.Kind {
	case permalinks.CandidatePostID:
		entry, err = a.services.Content.Get(r.Context(), authz.Principal{}, domaincontent.ID(candidate.ID))
	default:
		entry, err = a.services.Content.GetBySlug(r.Context(), authz.Principal{}, domaincontent.KindPost, candidate.Slug, locale.From(r.Context()))
	}
	if err != nil {
		if isResolvableMiss(err) {
			return publicrender.RenderRequest{}, false, nil
		}
		return publicrender.RenderRequest{}, true, err
	}
	if entry.Kind != domaincontent.KindPost || !permalinks.MatchesEntry(candidate, entry) {
		return publicrender.RenderRequest{}, false, nil
	}
	return a.content(r, config, entry, publicrender.RenderKindPost, "post", "post"), true, nil
}

func (a pageAssembler) content(r *http.Request, config publicrender.SiteConfig, entry domaincontent.Entry, kind publicrender.RenderKind, screen string, role string) publicrender.RenderRequest {
	path := permalinks.EntryPath(entry, config.Permalinks)
	title := entry.Title.Value(locale.From(r.Context()), "en")
	description := entry.Excerpt.Value(locale.From(r.Context()), "en")
	if strings.TrimSpace(description) == "" {
		description = title
	}
	publishedAt := permalinks.PublishTime(entry)
	featured := a.mediaView(r, entry.FeaturedMediaID)
	author := a.authorView(r, entry.AuthorID)
	terms := a.entryTerms(r, entry)
	related := a.relatedPosts(r, entry, config.Permalinks)
	breadcrumbs := []publicrender.Breadcrumb{{Label: "Home", URL: "/"}}
	if kind == publicrender.RenderKindPost {
		breadcrumbs = append(breadcrumbs, publicrender.Breadcrumb{Label: "Blog", URL: "/blog/"})
	}
	breadcrumbs = append(breadcrumbs, publicrender.Breadcrumb{Label: title, URL: path})
	request, _ := a.renderRequest(r, config, publicrender.PublicPage{
		Kind:         kind,
		Screen:       screen,
		TemplateRole: domaintaxonomyRole(role),
		Title:        title,
		Description:  description,
		Content:      entry.Body.Value(locale.From(r.Context()), "en"),
		CanonicalURL: path,
		Published:    "Published: " + publishedAt.UTC().Format(time.RFC3339),
		Breadcrumbs:  breadcrumbs,
		Author:       author,
		Featured:     featured,
		Terms:        terms,
		RelatedItems: related,
		SEO: seoModel(
			title,
			description,
			path,
			contentType(kind),
			featured,
			publishedAt,
			entry.UpdatedAt,
			false,
		),
	}, path)
	return request
}

func (a pageAssembler) notFound(r *http.Request, config publicrender.SiteConfig, description string) publicrender.RenderRequest {
	request, _ := a.renderRequest(r, config, publicrender.PublicPage{
		Kind:         publicrender.RenderKindNotFound,
		Screen:       "not-found",
		TemplateRole: "not_found",
		Title:        "Not Found",
		Description:  description,
		ActionLabel:  "Go home",
		Breadcrumbs:  []publicrender.Breadcrumb{{Label: "Home", URL: "/"}},
		SEO:          seoModel("Not Found", description, "", "website", nil, time.Time{}, time.Time{}, true),
	}, r.URL.Path)
	return request
}

func (a pageAssembler) renderRequest(r *http.Request, config publicrender.SiteConfig, page publicrender.PublicPage, activePath string) (publicrender.RenderRequest, error) {
	theme := a.themes.ResolveActive(config.ActiveTheme)
	manifest := theme.Manifest()
	preset := a.themes.ResolvePreset(string(manifest.ID), config.StylePreset)
	path := activePath
	if strings.TrimSpace(path) == "" {
		path = r.URL.Path
	}
	page.Layout = a.layout(r, config, page.Title, path, manifest.ID, preset)
	if page.CanonicalURL == "" {
		page.CanonicalURL = path
	}
	if page.SEO.CanonicalURL == "" {
		page.SEO.CanonicalURL = page.CanonicalURL
	}
	return publicrender.RenderRequest{
		Context: publicrender.RenderContext{
			Path:        r.URL.Path,
			Query:       r.URL.RawQuery,
			Locale:      locale.From(r.Context()),
			ThemeID:     manifest.ID,
			StylePreset: preset.ID,
		},
		Page: page,
	}, nil
}

func (a pageAssembler) layout(r *http.Request, config publicrender.SiteConfig, title string, activePath string, themeID domainthemes.ThemeID, preset domainthemes.StylePreset) publicrender.Layout {
	fixture := adminfixtures.MustLoad(locale.From(r.Context()))
	stylesheets, scripts := a.themes.ThemeAssets(string(themeID))
	stylesheets = append(stylesheets, preset.Stylesheets...)
	scripts = append(scripts, preset.Scripts...)
	return publicrender.Layout{
		Title:         title,
		Lang:          fixture.Meta.Lang,
		Brand:         firstNonEmpty(config.BrandName, config.Title, platformthemes.DefaultThemeLabel),
		Tagline:       config.HomeIntro,
		ActivePath:    activePath,
		ThemeID:       themeID,
		StylePresetID: firstNonEmpty(preset.ID, "default"),
		HeaderMenu:    a.menuItems(r, domainmenus.Location("primary"), activePath),
		FooterMenu:    a.menuItems(r, domainmenus.Location("footer"), activePath),
		Assets: publicrender.AssetBundle{
			Base:        assets.Resolve(),
			Stylesheets: stylesheets,
			Scripts:     scripts,
		},
		ThemeToggle: view.ThemeToggleData{
			Label:              fixture.Theme.Label,
			SwitchToDarkLabel:  fixture.Theme.SwitchToDarkLabel,
			SwitchToLightLabel: fixture.Theme.SwitchToLightLabel,
		},
		Language: view.BuildLanguageToggleFromContext(r.Context(),
			view.WithLabel(fixture.Language.Label),
			view.WithCurrentLabel(fixture.Language.CurrentLabel),
			view.WithNextLocale(fixture.Language.NextLocale),
			view.WithNextLabel(fixture.Language.NextLabel),
			view.WithLocaleLabels(fixture.Language.LocaleLabels),
		),
	}
}

func (a pageAssembler) menuItems(r *http.Request, location domainmenus.Location, activePath string) []publicrender.MenuItem {
	menu, ok, err := a.services.Menus.ByLocation(r.Context(), location)
	if err != nil || !ok {
		return nil
	}
	return mapMenuItems(menu.Items, activePath)
}

func mapMenuItems(items []domainmenus.Item, activePath string) []publicrender.MenuItem {
	out := make([]publicrender.MenuItem, 0, len(items))
	for _, item := range items {
		childActive := strings.TrimRight(item.URL, "/") == strings.TrimRight(activePath, "/")
		out = append(out, publicrender.MenuItem{
			Label:    item.Label,
			URL:      defaultPath(item.URL),
			Active:   childActive,
			Children: mapMenuItems(item.Children, activePath),
		})
	}
	return out
}

func (a pageAssembler) archiveItems(r *http.Request, entries []domaincontent.Entry, settings permalinks.Settings) []publicrender.ArchiveItem {
	items := make([]publicrender.ArchiveItem, 0, len(entries))
	for _, entry := range entries {
		items = append(items, publicrender.ArchiveItem{
			ID:       string(entry.ID),
			Kind:     string(entry.Kind),
			Title:    entry.Title.Value(locale.From(r.Context()), "en"),
			Summary:  entry.Excerpt.Value(locale.From(r.Context()), "en"),
			Href:     permalinks.EntryPath(entry, settings),
			Meta:     permalinks.PublishTime(entry).UTC().Format("2006-01-02"),
			Author:   a.authorView(r, entry.AuthorID),
			Featured: a.mediaView(r, entry.FeaturedMediaID),
			Terms:    a.entryTerms(r, entry),
		})
	}
	return items
}

func (a pageAssembler) findTerm(r *http.Request, taxonomy domaintaxonomy.Type, value string) (domaintaxonomy.Term, bool) {
	terms, err := a.services.Taxonomy.ListTerms(r.Context(), taxonomy)
	if err != nil {
		return domaintaxonomy.Term{}, false
	}
	for _, item := range terms {
		if strings.EqualFold(item.Slug.Value(locale.From(r.Context()), "en"), value) ||
			strings.EqualFold(item.Slug.Value("en", "en"), value) ||
			string(item.ID) == value {
			return item, true
		}
	}
	return domaintaxonomy.Term{}, false
}

func (a pageAssembler) termView(item domaintaxonomy.Term) *publicrender.TermView {
	return &publicrender.TermView{
		Taxonomy: string(item.Type),
		ID:       string(item.ID),
		Slug:     item.Slug.Value("en", "en"),
		Label:    item.Name.Value("en", "en"),
	}
}

func (a pageAssembler) entryTerms(r *http.Request, entry domaincontent.Entry) []publicrender.TermView {
	if len(entry.Terms) == 0 {
		return nil
	}
	views := make([]publicrender.TermView, 0, len(entry.Terms))
	seen := map[string]struct{}{}
	for _, ref := range entry.Terms {
		key := ref.Taxonomy + ":" + ref.TermID
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		term, ok := a.findTerm(r, domaintaxonomy.Type(ref.Taxonomy), ref.TermID)
		if !ok {
			continue
		}
		if view := a.termView(term); view != nil {
			views = append(views, *view)
		}
	}
	slices.SortFunc(views, func(left publicrender.TermView, right publicrender.TermView) int {
		return strings.Compare(left.Label, right.Label)
	})
	return views
}

func (a pageAssembler) findAuthorBySlug(r *http.Request, slug string) (*publicrender.AuthorView, bool) {
	users, err := a.services.Users.List(r.Context())
	if err != nil {
		return nil, false
	}
	for _, user := range users {
		if user.Status != domainusers.StatusActive {
			continue
		}
		profile := user.PublicAuthor()
		if strings.EqualFold(profile.Slug, slug) {
			return &publicrender.AuthorView{
				ID:          string(profile.ID),
				Slug:        profile.Slug,
				DisplayName: profile.DisplayName,
				Bio:         profile.Bio,
				AvatarURL:   profile.AvatarURL,
				WebsiteURL:  profile.WebsiteURL,
			}, true
		}
	}
	return nil, false
}

func (a pageAssembler) authorView(r *http.Request, authorID string) *publicrender.AuthorView {
	if strings.TrimSpace(authorID) == "" {
		return nil
	}
	profile, ok, err := a.services.Users.PublicAuthor(r.Context(), domainusers.ID(authorID))
	if err != nil || !ok {
		return nil
	}
	return &publicrender.AuthorView{
		ID:          string(profile.ID),
		Slug:        profile.Slug,
		DisplayName: profile.DisplayName,
		Bio:         profile.Bio,
		AvatarURL:   profile.AvatarURL,
		WebsiteURL:  profile.WebsiteURL,
	}
}

func (a pageAssembler) mediaView(r *http.Request, mediaID string) *publicrender.MediaView {
	if strings.TrimSpace(mediaID) == "" {
		return nil
	}
	asset, ok, err := a.services.Media.Get(r.Context(), domainmedia.ID(mediaID))
	if err != nil || !ok {
		return nil
	}
	return &publicrender.MediaView{
		ID:      string(asset.ID),
		URL:     asset.PublicURL,
		AltText: asset.AltText,
		Caption: asset.Caption,
		Width:   asset.Width,
		Height:  asset.Height,
	}
}

func (a pageAssembler) relatedPosts(r *http.Request, entry domaincontent.Entry, settings permalinks.Settings) []publicrender.ArchiveItem {
	var items []domaincontent.Entry
	for _, ref := range entry.Terms {
		result, err := a.services.Content.List(r.Context(), domaincontent.Query{
			Kinds:      []domaincontent.Kind{domaincontent.KindPost},
			PublicOnly: true,
			Taxonomy:   ref.Taxonomy,
			TermID:     ref.TermID,
			Page:       1,
			PerPage:    4,
			SortBy:     domaincontent.SortPublishedAt,
			SortDesc:   true,
		})
		if err == nil && len(result.Items) > 0 {
			items = result.Items
			break
		}
	}
	if len(items) == 0 {
		result, err := a.services.Content.List(r.Context(), domaincontent.Query{
			Kinds:      []domaincontent.Kind{domaincontent.KindPost},
			PublicOnly: true,
			Page:       1,
			PerPage:    5,
			SortBy:     domaincontent.SortPublishedAt,
			SortDesc:   true,
		})
		if err == nil {
			items = result.Items
		}
	}
	filtered := make([]domaincontent.Entry, 0, len(items))
	for _, item := range items {
		if item.ID != entry.ID {
			filtered = append(filtered, item)
		}
		if len(filtered) == 3 {
			break
		}
	}
	return a.archiveItems(r, filtered, settings)
}

func seoModel(title string, description string, canonical string, typ string, featured *publicrender.MediaView, publishedAt time.Time, modifiedAt time.Time, noindex bool) publicrender.SEOModel {
	model := publicrender.SEOModel{
		Title:        title,
		Description:  description,
		CanonicalURL: canonical,
		Type:         typ,
		NoIndex:      noindex,
	}
	if featured != nil {
		model.ImageURL = featured.URL
	}
	if !publishedAt.IsZero() {
		model.PublishedAt = publishedAt.UTC().Format(time.RFC3339)
	}
	if !modifiedAt.IsZero() {
		model.ModifiedAt = modifiedAt.UTC().Format(time.RFC3339)
	}
	return model
}

func paginationData(base string, page int, totalPages int) publicrender.Pagination {
	return publicrender.Pagination{
		Page:          page,
		TotalPages:    totalPages,
		BaseHref:      base,
		PreviousLabel: "Previous",
		NextLabel:     "Next",
	}
}

func searchBase(query string) string {
	if strings.TrimSpace(query) == "" {
		return "/search/"
	}
	return "/search/?q=" + url.QueryEscape(query)
}

func defaultPath(path string) string {
	if strings.TrimSpace(path) == "" {
		return "/"
	}
	return path
}

func contentType(kind publicrender.RenderKind) string {
	switch kind {
	case publicrender.RenderKindPost:
		return "article"
	default:
		return "website"
	}
}

func domaintaxonomyRole(value string) domainthemes.TemplateRole {
	return domainthemes.TemplateRole(value)
}
