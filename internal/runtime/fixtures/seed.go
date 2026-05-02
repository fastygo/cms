package fixtures

import (
	"context"
	"time"

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
		{ID: "term-news", Type: domaintaxonomy.TypeCategory, Name: domaincontent.LocalizedText{"en": "News"}, Slug: domaincontent.LocalizedText{"en": "news"}},
		{ID: "term-featured", Type: domaintaxonomy.TypeTag, Name: domaincontent.LocalizedText{"en": "Featured"}, Slug: domaincontent.LocalizedText{"en": "featured"}},
		{ID: "term-security", Type: domaintaxonomy.Type("topic"), Name: domaincontent.LocalizedText{"en": "Security"}, Slug: domaincontent.LocalizedText{"en": "security"}},
	} {
		if err := store.SaveTerm(ctx, term); err != nil {
			return err
		}
	}

	if err := store.SaveUser(ctx, domainusers.User{
		ID:          "author-1",
		Login:       "jane",
		DisplayName: "Jane Editor",
		Email:       "jane@example.test",
		Status:      domainusers.StatusActive,
		Profile:     domainusers.AuthorProfile{Slug: "jane", Bio: "Fixture editor", AvatarURL: "/media/avatar-jane.png"},
	}); err != nil {
		return err
	}

	if err := store.SaveMedia(ctx, domainmedia.Asset{ID: "media-cover", Filename: "cover.jpg", MimeType: "image/jpeg", SizeBytes: 1024, Width: 1200, Height: 800, AltText: "Cover image", PublicURL: "/media/cover.jpg", PublicMeta: map[string]any{"source": "fixture"}, CreatedAt: now, UpdatedAt: now}); err != nil {
		return err
	}

	publishedAt := published
	scheduledAt := future
	entries := []domaincontent.Entry{
		{
			ID: "content-post-published", Kind: domaincontent.KindPost, Status: domaincontent.StatusPublished, Visibility: domaincontent.VisibilityPublic,
			Title: domaincontent.LocalizedText{"en": "Published Post"}, Slug: domaincontent.LocalizedText{"en": "published-post"}, Body: domaincontent.LocalizedText{"en": "Public fixture content"}, Excerpt: domaincontent.LocalizedText{"en": "Public excerpt"},
			AuthorID: "author-1", FeaturedMediaID: "media-cover", Terms: []domaincontent.TermRef{{Taxonomy: "category", TermID: "term-news"}, {Taxonomy: "tag", TermID: "term-featured"}},
			Metadata:  domaincontent.Metadata{"public_key": {Value: "public", Public: true}, "private_key": {Value: "private", Public: false}},
			CreatedAt: now.Add(-48 * time.Hour), UpdatedAt: now.Add(-24 * time.Hour), PublishedAt: &publishedAt,
		},
		{
			ID: "content-post-draft", Kind: domaincontent.KindPost, Status: domaincontent.StatusDraft, Visibility: domaincontent.VisibilityPublic,
			Title: domaincontent.LocalizedText{"en": "Draft Post"}, Slug: domaincontent.LocalizedText{"en": "draft-post"}, Body: domaincontent.LocalizedText{"en": "Draft fixture content"},
			AuthorID: "author-1", CreatedAt: now, UpdatedAt: now,
		},
		{
			ID: "content-post-scheduled", Kind: domaincontent.KindPost, Status: domaincontent.StatusScheduled, Visibility: domaincontent.VisibilityPublic,
			Title: domaincontent.LocalizedText{"en": "Scheduled Post"}, Slug: domaincontent.LocalizedText{"en": "scheduled-post"}, Body: domaincontent.LocalizedText{"en": "Scheduled fixture content"},
			AuthorID: "author-1", CreatedAt: now, UpdatedAt: now, PublishedAt: &scheduledAt,
		},
		{
			ID: "content-page-about", Kind: domaincontent.KindPage, Status: domaincontent.StatusPublished, Visibility: domaincontent.VisibilityPublic,
			Title: domaincontent.LocalizedText{"en": "About"}, Slug: domaincontent.LocalizedText{"en": "about"}, Body: domaincontent.LocalizedText{"en": "About page fixture"},
			AuthorID: "author-1", CreatedAt: now.Add(-72 * time.Hour), UpdatedAt: now.Add(-24 * time.Hour), PublishedAt: &publishedAt,
		},
	}
	for _, entry := range entries {
		if err := store.Save(ctx, entry); err != nil {
			return err
		}
	}

	for _, value := range []domainsettings.Value{
		{Key: "site.title", Value: "GoCMS Fixture", Public: true},
		{Key: "site.private_note", Value: "hidden", Public: false},
	} {
		if err := store.SaveSetting(ctx, value); err != nil {
			return err
		}
	}

	return store.SaveMenu(ctx, domainmenus.Menu{
		ID:       "menu-primary",
		Name:     "Primary",
		Location: "primary",
		Items: []domainmenus.Item{
			{ID: "menu-home", Label: "Home", URL: "/"},
			{ID: "menu-blog", Label: "Blog", URL: "/blog"},
		},
	})
}
