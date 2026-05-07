package admin

import (
	"testing"

	"github.com/fastygo/cms/internal/domain/authz"
	domaincontent "github.com/fastygo/cms/internal/domain/content"
	"github.com/fastygo/cms/internal/platform/cmspanel"
	"github.com/fastygo/cms/internal/platform/plugins"
	"github.com/fastygo/cms/internal/site/adminfixtures"
	"github.com/fastygo/cms/internal/site/ui/blocks"
)

func TestContentFormFieldsPreserveRichEditorMetadata(t *testing.T) {
	resource := cmspanel.PostsResource()
	fields := contentFormFields(resource.Form, domaincontent.Entry{
		Kind:     domaincontent.KindPost,
		Title:    domaincontent.LocalizedText{"en": "Hello"},
		Slug:     domaincontent.LocalizedText{"en": "hello"},
		Body:     domaincontent.LocalizedText{"en": "<p>Hello</p>"},
		Excerpt:  domaincontent.LocalizedText{"en": "Short"},
		AuthorID: "author-7",
		Terms:    []domaincontent.TermRef{{Taxonomy: "category", TermID: "news"}},
	}, plugins.EditorProviderRegistration{ID: "tiptap-basic"})

	content := blockFieldByID(fields, "content")
	if content.Component != "richtext" {
		t.Fatalf("content component = %q, want richtext", content.Component)
	}
	if content.Editor == nil || content.Editor.ProviderID != "tiptap-basic" {
		t.Fatalf("content editor = %+v, want tiptap-basic provider", content.Editor)
	}
	if content.Value != "<p>Hello</p>" {
		t.Fatalf("content value = %q, want editor HTML", content.Value)
	}
	if terms := blockFieldByID(fields, "terms"); terms.Value != "category:news" || terms.Placeholder == "" {
		t.Fatalf("terms field = %+v, want formatted terms with placeholder", terms)
	}
}

func TestPanelTableSchemaAdaptsToContentHeadersAndActions(t *testing.T) {
	handler := Handler{}
	resource := cmspanel.PagesResource()
	bundle := adminfixtures.MustLoad("en")
	headers := handler.contentTableHeadersFromSchema(bundle, resource.Table)
	if headers.Title == "" || headers.Slug == "" || headers.Status == "" || headers.Author == "" {
		t.Fatalf("headers = %+v, want content table headers", headers)
	}

	actions := handler.panelActions(bundle, authz.NewPrincipal("editor", authz.CapabilityContentCreate), resource.Actions)
	if len(actions) != 1 {
		t.Fatalf("actions = %v, want create action", actions)
	}
	if !actions[0].Enabled || actions[0].Href != "/go-admin/pages/new" {
		t.Fatalf("action = %+v, want enabled create link", actions[0])
	}
}

func blockFieldByID(fields []blocks.FieldData, id string) blocks.FieldData {
	for _, field := range fields {
		if field.ID == id {
			return field
		}
	}
	return blocks.FieldData{}
}
