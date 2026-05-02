package taxonomy_test

import (
	"context"
	"fmt"
	"testing"

	apptaxonomy "github.com/fastygo/cms/internal/application/taxonomy"
	"github.com/fastygo/cms/internal/domain/authz"
	domaincontent "github.com/fastygo/cms/internal/domain/content"
	domaintaxonomy "github.com/fastygo/cms/internal/domain/taxonomy"
)

func TestTaxonomyRegisterCreateAndAssignTerms(t *testing.T) {
	ctx := context.Background()
	taxRepo := newMemoryTaxonomyRepo()
	entryRepo := &memoryEntryRepo{entries: map[domaincontent.ID]domaincontent.Entry{
		"content-1": {ID: "content-1", Kind: domaincontent.KindPost},
	}}
	service := apptaxonomy.NewService(taxRepo, entryRepo)
	manager := authz.NewPrincipal("editor-1", authz.CapabilityTaxonomiesManage, authz.CapabilityTaxonomiesAssign)

	if err := service.Register(ctx, manager, domaintaxonomy.Definition{
		Type:            domaintaxonomy.TypeCategory,
		Label:           "Categories",
		Mode:            domaintaxonomy.ModeHierarchical,
		AssignedToKinds: []domaincontent.Kind{domaincontent.KindPost},
		Public:          true,
		RESTVisible:     true,
	}); err != nil {
		t.Fatal(err)
	}
	if err := service.CreateTerm(ctx, manager, domaintaxonomy.Term{
		ID:   "term-1",
		Type: domaintaxonomy.TypeCategory,
		Name: domaincontent.LocalizedText{"en": "News"},
		Slug: domaincontent.LocalizedText{"en": "news"},
	}); err != nil {
		t.Fatal(err)
	}
	entry, err := service.AssignTerms(ctx, manager, "content-1", []domaincontent.TermRef{{Taxonomy: "category", TermID: "term-1"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(entry.Terms) != 1 || entry.Terms[0].TermID != "term-1" {
		t.Fatalf("expected assigned term, got %+v", entry.Terms)
	}
}

type memoryTaxonomyRepo struct {
	definitions map[domaintaxonomy.Type]domaintaxonomy.Definition
	terms       map[domaintaxonomy.TermID]domaintaxonomy.Term
}

func newMemoryTaxonomyRepo() *memoryTaxonomyRepo {
	return &memoryTaxonomyRepo{
		definitions: make(map[domaintaxonomy.Type]domaintaxonomy.Definition),
		terms:       make(map[domaintaxonomy.TermID]domaintaxonomy.Term),
	}
}

func (r *memoryTaxonomyRepo) SaveDefinition(_ context.Context, definition domaintaxonomy.Definition) error {
	r.definitions[definition.Type] = definition
	return nil
}

func (r *memoryTaxonomyRepo) GetDefinition(_ context.Context, value domaintaxonomy.Type) (domaintaxonomy.Definition, bool, error) {
	definition, ok := r.definitions[value]
	return definition, ok, nil
}

func (r *memoryTaxonomyRepo) ListDefinitions(context.Context) ([]domaintaxonomy.Definition, error) {
	definitions := make([]domaintaxonomy.Definition, 0, len(r.definitions))
	for _, definition := range r.definitions {
		definitions = append(definitions, definition)
	}
	return definitions, nil
}

func (r *memoryTaxonomyRepo) SaveTerm(_ context.Context, term domaintaxonomy.Term) error {
	r.terms[term.ID] = term
	return nil
}

func (r *memoryTaxonomyRepo) GetTerm(_ context.Context, id domaintaxonomy.TermID) (domaintaxonomy.Term, bool, error) {
	term, ok := r.terms[id]
	return term, ok, nil
}

func (r *memoryTaxonomyRepo) ListTerms(_ context.Context, taxonomyType domaintaxonomy.Type) ([]domaintaxonomy.Term, error) {
	terms := make([]domaintaxonomy.Term, 0, len(r.terms))
	for _, term := range r.terms {
		if term.Type == taxonomyType {
			terms = append(terms, term)
		}
	}
	return terms, nil
}

type memoryEntryRepo struct {
	entries map[domaincontent.ID]domaincontent.Entry
}

func (r *memoryEntryRepo) Get(_ context.Context, id domaincontent.ID) (domaincontent.Entry, error) {
	entry, ok := r.entries[id]
	if !ok {
		return domaincontent.Entry{}, fmt.Errorf("content %q not found", id)
	}
	return entry, nil
}

func (r *memoryEntryRepo) Save(_ context.Context, entry domaincontent.Entry) error {
	r.entries[entry.ID] = entry
	return nil
}
