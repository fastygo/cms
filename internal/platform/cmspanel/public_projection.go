package cmspanel

import (
	domaincontent "github.com/fastygo/cms/internal/domain/content"
)

type PublicProjectionContract struct {
	ID             string
	ResourceID     string
	Kind           domaincontent.Kind
	Label          string
	REST           RESTProjectionContract
	GraphQL        GraphQLProjectionContract
	Implementation string
}

type RESTProjectionContract struct {
	CollectionPath string
	BySlugPath     string
	RequiredFields []string
}

type GraphQLProjectionContract struct {
	CollectionField string
	SingleField     string
	SingleSlug      string
	RequiredFields  []string
}

func PublicProjectionContracts() []PublicProjectionContract {
	return []PublicProjectionContract{
		contentProjectionContract(PostsResource(), "published-post", "/go-json/go/v2/posts", "/go-json/go/v2/posts/by-slug/published-post", "posts", "post"),
		contentProjectionContract(PagesResource(), "about", "/go-json/go/v2/pages", "/go-json/go/v2/pages/by-slug/about", "pages", "page"),
		{
			ID:         "media",
			ResourceID: "media",
			Label:      "Media",
			REST: RESTProjectionContract{
				CollectionPath: "/go-json/go/v2/media",
				RequiredFields: []string{
					"id",
					"filename",
					"mime_type",
					"public_url",
				},
			},
			GraphQL: GraphQLProjectionContract{
				CollectionField: "media",
				RequiredFields: []string{
					"id",
					"filename",
					"publicURL",
				},
			},
			Implementation: "Public media projection exists in REST and GraphQL and is aligned with the cmspanel MediaPage admin descriptor.",
		},
		{
			ID:         "taxonomies",
			ResourceID: "taxonomies",
			Label:      "Taxonomies",
			REST: RESTProjectionContract{
				CollectionPath: "/go-json/go/v2/taxonomies",
				RequiredFields: []string{
					"type",
					"label",
					"public",
					"rest_visible",
					"graphql_visible",
				},
			},
			GraphQL: GraphQLProjectionContract{
				CollectionField: "taxonomies",
				RequiredFields: []string{
					"type",
					"label",
					"public",
					"restVisible",
					"graphqlVisible",
				},
			},
			Implementation: "Public taxonomy projection exists in REST and GraphQL and is aligned with the cmspanel TaxonomiesPage and TermsPage admin descriptors.",
		},
		{
			ID:         "menus",
			ResourceID: "menus",
			Label:      "Menus",
			REST: RESTProjectionContract{
				CollectionPath: "/go-json/go/v2/menus",
				RequiredFields: []string{
					"id",
					"name",
					"location",
					"items",
				},
			},
			GraphQL: GraphQLProjectionContract{
				CollectionField: "menus",
				RequiredFields: []string{
					"id",
					"name",
					"location",
					"items",
				},
			},
			Implementation: "Public menu projection exists in REST and GraphQL and is aligned with the cmspanel MenusPage admin descriptor.",
		},
	}
}

func contentProjectionContract(resource ContentResource, singleSlug string, restCollectionPath string, restBySlugPath string, graphCollectionField string, graphSingleField string) PublicProjectionContract {
	graphFields := []string{
		"id",
		"kind",
		"status",
		"slug",
		"title",
	}
	if resource.Kind == domaincontent.KindPost {
		graphFields = append(graphFields, "content", "excerpt", "authorID", "taxonomies", "links")
	}
	implementation := "Public content projection is backed by cmspanel PostsResource/PagesResource metadata and REST/GraphQL delivery contracts."
	if resource.Kind == domaincontent.KindPage {
		implementation = "Public page identity projection is validated through GraphQL; richer page body, excerpt, and links fields are recorded as a GraphQL projection gap."
	}
	return PublicProjectionContract{
		ID:         string(resource.ID),
		ResourceID: string(resource.ID),
		Kind:       resource.Kind,
		Label:      resource.Label,
		REST: RESTProjectionContract{
			CollectionPath: restCollectionPath,
			BySlugPath:     restBySlugPath,
			RequiredFields: []string{
				"id",
				"kind",
				"status",
				"slug",
				"title",
				"content",
				"excerpt",
				"author_id",
				"taxonomy_ids",
				"links",
			},
		},
		GraphQL: GraphQLProjectionContract{
			CollectionField: graphCollectionField,
			SingleField:     graphSingleField,
			SingleSlug:      singleSlug,
			RequiredFields:  graphFields,
		},
		Implementation: implementation,
	}
}
