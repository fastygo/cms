package fixtures

import (
	"context"
	"time"

	appauthn "github.com/fastygo/cms/internal/application/authn"
	"github.com/fastygo/cms/internal/domain/authz"
	domaincontent "github.com/fastygo/cms/internal/domain/content"
	domaincontenttype "github.com/fastygo/cms/internal/domain/contenttype"
	domainmedia "github.com/fastygo/cms/internal/domain/media"
	domainmenus "github.com/fastygo/cms/internal/domain/menus"
	domainsettings "github.com/fastygo/cms/internal/domain/settings"
	domaintaxonomy "github.com/fastygo/cms/internal/domain/taxonomy"
	domainusers "github.com/fastygo/cms/internal/domain/users"
)

type Store interface {
	Save(context.Context, domaincontent.Entry) error
	SaveContentType(context.Context, domaincontenttype.Type) error
	SaveDefinition(context.Context, domaintaxonomy.Definition) error
	SaveTerm(context.Context, domaintaxonomy.Term) error
	SaveMedia(context.Context, domainmedia.Asset) error
	SaveUser(context.Context, domainusers.User) error
	SaveSetting(context.Context, domainsettings.Value) error
	SaveMenu(context.Context, domainmenus.Menu) error
}

func Seed(ctx context.Context, store Store) error {
	now := time.Date(2026, 5, 2, 10, 0, 0, 0, time.UTC)
	future := now.Add(24 * time.Hour)
	published := now.Add(-24 * time.Hour)
	hasher := appauthn.DefaultPasswordHasher()
	adminHash, err := hasher.Hash("admin")
	if err != nil {
		return err
	}
	editorHash, err := hasher.Hash("editor")
	if err != nil {
		return err
	}
	viewerHash, err := hasher.Hash("viewer")
	if err != nil {
		return err
	}

	for _, contentType := range []domaincontenttype.Type{
		domaincontenttype.BuiltInPost(),
		domaincontenttype.BuiltInPage(),
		{
			ID:             domaincontent.Kind("product"),
			Label:          "Products",
			Public:         true,
			RESTVisible:    true,
			GraphQLVisible: true,
			Archive:        true,
			Permalink:      "/products/{slug}",
			Supports:       domaincontenttype.Supports{Title: true, Editor: true, Excerpt: true, FeaturedMedia: true, Revisions: true, Taxonomies: true, CustomFields: true},
		},
	} {
		if err := store.SaveContentType(ctx, contentType); err != nil {
			return err
		}
	}

	for _, definition := range []domaintaxonomy.Definition{
		{Type: domaintaxonomy.TypeCategory, Label: "Categories", Mode: domaintaxonomy.ModeHierarchical, AssignedToKinds: []domaincontent.Kind{domaincontent.KindPost}, Public: true, RESTVisible: true, GraphQLVisible: true},
		{Type: domaintaxonomy.TypeTag, Label: "Tags", Mode: domaintaxonomy.ModeFlat, AssignedToKinds: []domaincontent.Kind{domaincontent.KindPost}, Public: true, RESTVisible: true, GraphQLVisible: true},
		{Type: domaintaxonomy.Type("topic"), Label: "Topics", Mode: domaintaxonomy.ModeFlat, AssignedToKinds: []domaincontent.Kind{domaincontent.Kind("product")}, Public: true, RESTVisible: true, GraphQLVisible: true},
	} {
		if err := store.SaveDefinition(ctx, definition); err != nil {
			return err
		}
	}

	for _, term := range []domaintaxonomy.Term{
		{ID: "term-news", Type: domaintaxonomy.TypeCategory, Name: domaincontent.LocalizedText{"en": "News", "ru": "Новости"}, Slug: domaincontent.LocalizedText{"en": "news", "ru": "news"}},
		{ID: "term-featured", Type: domaintaxonomy.TypeTag, Name: domaincontent.LocalizedText{"en": "Featured", "ru": "Избранное"}, Slug: domaincontent.LocalizedText{"en": "featured", "ru": "featured"}},
		{ID: "term-security", Type: domaintaxonomy.Type("topic"), Name: domaincontent.LocalizedText{"en": "Security", "ru": "Безопасность"}, Slug: domaincontent.LocalizedText{"en": "security", "ru": "security"}},
	} {
		if err := store.SaveTerm(ctx, term); err != nil {
			return err
		}
	}

	if err := store.SaveMedia(ctx, domainmedia.Asset{
		ID: "media-cover", Filename: "go-cms-itgarage.webp", MimeType: "image/webp", SizeBytes: 48_000,
		Width: 1200, Height: 800, AltText: "GoCMS IT garage cover",
		PublicURL: "/static/img/go-cms-itgarage.webp", PublicMeta: map[string]any{"source": "fixture"},
		CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		return err
	}
	if err := store.SaveMedia(ctx, domainmedia.Asset{
		ID: "media-avatar", Filename: "gosms-banner.png", MimeType: "image/png", SizeBytes: 120_000,
		Width: 1200, Height: 630, AltText: "Mr Gopher",
		PublicURL: "/static/img/gosms-banner.png", PublicMeta: map[string]any{"source": "fixture"},
		CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		return err
	}

	passwordUpdatedAt := now
	for _, user := range []domainusers.User{
		{
			ID:                "admin",
			Login:             "admin",
			DisplayName:       "Admin",
			Email:             "admin@example.test",
			Status:            domainusers.StatusActive,
			Roles:             []string{authz.RoleAdmin},
			PasswordHash:      adminHash,
			PasswordUpdatedAt: &passwordUpdatedAt,
		},
		{
			ID:          "author-1",
			Login:       "mrgopher",
			DisplayName: "Mr Gopher",
			Email:       "mr.gopher@gocms.example.test",
			Status:      domainusers.StatusActive,
			Roles:       []string{authz.RoleEditor},
			Profile: domainusers.AuthorProfile{
				Slug: "mr-gopher", Bio: "GoCMS mascot and fixture author.", AvatarMediaID: "media-avatar",
			},
			PasswordHash:      editorHash,
			PasswordUpdatedAt: &passwordUpdatedAt,
		},
		{
			ID:                "viewer",
			Login:             "viewer",
			DisplayName:       "Viewer",
			Email:             "viewer@example.test",
			Status:            domainusers.StatusActive,
			Roles:             []string{authz.RoleViewer},
			PasswordHash:      viewerHash,
			PasswordUpdatedAt: &passwordUpdatedAt,
		},
	} {
		if err := store.SaveUser(ctx, user); err != nil {
			return err
		}
	}
	if err := store.SaveUser(ctx, domainusers.User{
		ID:                "legacy-editor",
		Login:             "editor",
		DisplayName:       "Editor",
		Email:             "editor@example.test",
		Status:            domainusers.StatusActive,
		Roles:             []string{authz.RoleEditor},
		PasswordHash:      editorHash,
		PasswordUpdatedAt: &passwordUpdatedAt,
	}); err != nil {
		return err
	}

	publishedAt := published
	scheduledAt := future
	entries := []domaincontent.Entry{
		{
			ID: "content-post-published", Kind: domaincontent.KindPost, Status: domaincontent.StatusPublished, Visibility: domaincontent.VisibilityPublic,
			Title:    domaincontent.LocalizedText{"en": "Published Post", "ru": "Опубликованная запись"},
			Slug:     domaincontent.LocalizedText{"en": "published-post", "ru": "published-post"},
			Body:     domaincontent.LocalizedText{"en": "Public fixture content", "ru": "Тестовое публичное содержимое"},
			Excerpt:  domaincontent.LocalizedText{"en": "Public excerpt", "ru": "Краткое описание"},
			AuthorID: "author-1", FeaturedMediaID: "media-cover", Terms: []domaincontent.TermRef{{Taxonomy: "category", TermID: "term-news"}, {Taxonomy: "tag", TermID: "term-featured"}},
			Metadata:  domaincontent.Metadata{"public_key": {Value: "public", Public: true}, "private_key": {Value: "private", Public: false}},
			CreatedAt: now.Add(-48 * time.Hour), UpdatedAt: now.Add(-24 * time.Hour), PublishedAt: &publishedAt,
		},
		{
			ID: "content-post-draft", Kind: domaincontent.KindPost, Status: domaincontent.StatusDraft, Visibility: domaincontent.VisibilityPublic,
			Title:    domaincontent.LocalizedText{"en": "Draft Post", "ru": "Черновик"},
			Slug:     domaincontent.LocalizedText{"en": "draft-post", "ru": "draft-post"},
			Body:     domaincontent.LocalizedText{"en": "Draft fixture content", "ru": "Черновое содержимое"},
			AuthorID: "author-1", CreatedAt: now, UpdatedAt: now,
		},
		{
			ID: "content-post-scheduled", Kind: domaincontent.KindPost, Status: domaincontent.StatusScheduled, Visibility: domaincontent.VisibilityPublic,
			Title:    domaincontent.LocalizedText{"en": "Scheduled Post", "ru": "Отложенная публикация"},
			Slug:     domaincontent.LocalizedText{"en": "scheduled-post", "ru": "scheduled-post"},
			Body:     domaincontent.LocalizedText{"en": "Scheduled fixture content", "ru": "Отложенное содержимое"},
			AuthorID: "author-1", CreatedAt: now, UpdatedAt: now, PublishedAt: &scheduledAt,
		},
		{
			ID: "content-page-about", Kind: domaincontent.KindPage, Status: domaincontent.StatusPublished, Visibility: domaincontent.VisibilityPublic,
			Title:    domaincontent.LocalizedText{"en": "About", "ru": "О проекте"},
			Slug:     domaincontent.LocalizedText{"en": "about", "ru": "about"},
			Body:     domaincontent.LocalizedText{"en": "About page fixture", "ru": "Тестовая страница «О проекте»"},
			AuthorID: "author-1", CreatedAt: now.Add(-72 * time.Hour), UpdatedAt: now.Add(-24 * time.Hour), PublishedAt: &publishedAt,
		},
	}
	for _, entry := range entries {
		if err := store.Save(ctx, entry); err != nil {
			return err
		}
	}

	for _, value := range []domainsettings.Value{
		{Key: "site.title", Value: "GoCMS", Public: true},
		{Key: "site.private_note", Value: "hidden", Public: false},
	} {
		if err := store.SaveSetting(ctx, value); err != nil {
			return err
		}
	}

	if err := store.SaveMenu(ctx, domainmenus.Menu{
		ID:       "menu-primary",
		Name:     "Primary",
		Location: "primary",
		Items: []domainmenus.Item{
			{ID: "menu-home", Label: "Home", URL: "/"},
			{
				ID:    "menu-blog",
				Label: "Blog",
				URL:   "/blog/",
				Children: []domainmenus.Item{
					{ID: "menu-blog-news", Label: "News", URL: "/category/news/"},
				},
			},
			{ID: "menu-about", Label: "About", URL: "/about/"},
		},
	}); err != nil {
		return err
	}
	return store.SaveMenu(ctx, domainmenus.Menu{
		ID:       "menu-footer",
		Name:     "Footer",
		Location: "footer",
		Items: []domainmenus.Item{
			{ID: "menu-footer-home", Label: "Home", URL: "/"},
			{
				ID:    "menu-footer-resources",
				Label: "Resources",
				URL:   "/blog/",
				Children: []domainmenus.Item{
					{ID: "menu-footer-author", Label: "Mr Gopher", URL: "/author/mr-gopher/"},
				},
			},
		},
	})
}
