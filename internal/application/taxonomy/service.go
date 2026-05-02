package taxonomy

import (
	"context"
	"fmt"

	domainauthz "github.com/fastygo/cms/internal/domain/authz"
	domaincontent "github.com/fastygo/cms/internal/domain/content"
	domaintaxonomy "github.com/fastygo/cms/internal/domain/taxonomy"
)

type Repository interface {
	SaveDefinition(context.Context, domaintaxonomy.Definition) error
	GetDefinition(context.Context, domaintaxonomy.Type) (domaintaxonomy.Definition, bool, error)
	ListDefinitions(context.Context) ([]domaintaxonomy.Definition, error)
	SaveTerm(context.Context, domaintaxonomy.Term) error
	GetTerm(context.Context, domaintaxonomy.TermID) (domaintaxonomy.Term, bool, error)
	ListTerms(context.Context, domaintaxonomy.Type) ([]domaintaxonomy.Term, error)
}

type EntryRepository interface {
	Get(context.Context, domaincontent.ID) (domaincontent.Entry, error)
	Save(context.Context, domaincontent.Entry) error
}

type Service struct {
	repo    Repository
	entries EntryRepository
}

func NewService(repo Repository, entries EntryRepository) Service {
	return Service{repo: repo, entries: entries}
}

func (s Service) Register(ctx context.Context, principal domainauthz.Principal, definition domaintaxonomy.Definition) error {
	if !principal.Has(domainauthz.CapabilityTaxonomiesManage) {
		return fmt.Errorf("capability %q is required", domainauthz.CapabilityTaxonomiesManage)
	}
	if err := domaintaxonomy.ValidateType(definition.Type); err != nil {
		return err
	}
	return s.repo.SaveDefinition(ctx, definition)
}

func (s Service) CreateTerm(ctx context.Context, principal domainauthz.Principal, term domaintaxonomy.Term) error {
	if !principal.Has(domainauthz.CapabilityTaxonomiesManage) {
		return fmt.Errorf("capability %q is required", domainauthz.CapabilityTaxonomiesManage)
	}
	if _, ok, err := s.repo.GetDefinition(ctx, term.Type); err != nil {
		return err
	} else if !ok {
		return fmt.Errorf("taxonomy %q is not registered", term.Type)
	}
	return s.repo.SaveTerm(ctx, term)
}

func (s Service) ListDefinitions(ctx context.Context) ([]domaintaxonomy.Definition, error) {
	return s.repo.ListDefinitions(ctx)
}

func (s Service) GetDefinition(ctx context.Context, taxonomyType domaintaxonomy.Type) (domaintaxonomy.Definition, bool, error) {
	return s.repo.GetDefinition(ctx, taxonomyType)
}

func (s Service) ListTerms(ctx context.Context, taxonomyType domaintaxonomy.Type) ([]domaintaxonomy.Term, error) {
	return s.repo.ListTerms(ctx, taxonomyType)
}

func (s Service) GetTerm(ctx context.Context, id domaintaxonomy.TermID) (domaintaxonomy.Term, bool, error) {
	return s.repo.GetTerm(ctx, id)
}

func (s Service) AssignTerms(ctx context.Context, principal domainauthz.Principal, contentID domaincontent.ID, refs []domaincontent.TermRef) (domaincontent.Entry, error) {
	if !principal.Has(domainauthz.CapabilityTaxonomiesAssign) {
		return domaincontent.Entry{}, fmt.Errorf("capability %q is required", domainauthz.CapabilityTaxonomiesAssign)
	}
	entry, err := s.entries.Get(ctx, contentID)
	if err != nil {
		return domaincontent.Entry{}, err
	}
	for _, ref := range refs {
		term, ok, err := s.repo.GetTerm(ctx, domaintaxonomy.TermID(ref.TermID))
		if err != nil {
			return domaincontent.Entry{}, err
		}
		if !ok || string(term.Type) != ref.Taxonomy {
			return domaincontent.Entry{}, fmt.Errorf("term %q is not registered for taxonomy %q", ref.TermID, ref.Taxonomy)
		}
	}
	entry.Terms = refs
	return entry, s.entries.Save(ctx, entry)
}
