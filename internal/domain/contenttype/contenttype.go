package contenttype

import (
	"fmt"
	"strings"

	"github.com/fastygo/cms/internal/domain/authz"
	"github.com/fastygo/cms/internal/domain/content"
)

type Supports struct {
	Title         bool
	Editor        bool
	Excerpt       bool
	FeaturedMedia bool
	Revisions     bool
	Taxonomies    bool
	CustomFields  bool
	Comments      bool
}

type Type struct {
	ID             content.Kind
	Label          string
	Public         bool
	RESTVisible    bool
	GraphQLVisible bool
	Supports       Supports
	Archive        bool
	Permalink      string
	Capabilities   map[string]authz.Capability
}

func BuiltInPost() Type {
	return Type{
		ID:             content.KindPost,
		Label:          "Posts",
		Public:         true,
		RESTVisible:    true,
		GraphQLVisible: true,
		Archive:        true,
		Permalink:      "/posts/{slug}",
		Supports: Supports{
			Title: true, Editor: true, Excerpt: true, FeaturedMedia: true,
			Revisions: true, Taxonomies: true, CustomFields: true, Comments: true,
		},
	}
}

func BuiltInPage() Type {
	return Type{
		ID:             content.KindPage,
		Label:          "Pages",
		Public:         true,
		RESTVisible:    true,
		GraphQLVisible: true,
		Archive:        false,
		Permalink:      "/{slug}",
		Supports: Supports{
			Title: true, Editor: true, Excerpt: true, FeaturedMedia: true,
			Revisions: true, Taxonomies: true, CustomFields: true,
		},
	}
}

func Validate(t Type) error {
	if err := content.ValidateKind(t.ID); err != nil {
		return err
	}
	if strings.TrimSpace(t.Label) == "" {
		return fmt.Errorf("content type label is required")
	}
	return nil
}
