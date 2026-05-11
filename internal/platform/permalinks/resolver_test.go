package permalinks

import (
	"net/url"
	"testing"
	"time"

	domaincontent "github.com/fastygo/cms/internal/domain/content"
)

func TestResolveReturnsDeterministicCandidatesForPrettySlug(t *testing.T) {
	candidates := Resolve("/published-post/", url.Values{}, Settings{})
	if len(candidates) != 2 {
		t.Fatalf("expected 2 candidates, got %d", len(candidates))
	}
	if got := candidates[0].Kind; got != CandidatePageSlug {
		t.Fatalf("first candidate = %q, want %q", got, CandidatePageSlug)
	}
	if got := candidates[1].Kind; got != CandidatePostSlug {
		t.Fatalf("second candidate = %q, want %q", got, CandidatePostSlug)
	}
	if got := candidates[1].Slug; got != "published-post" {
		t.Fatalf("post slug = %q", got)
	}
}

func TestResolveSupportsSearchAndTaxonomyArchives(t *testing.T) {
	blog := Resolve("/blog/", url.Values{}, Settings{})
	if len(blog) != 1 || blog[0].Kind != CandidateBlog {
		t.Fatalf("unexpected blog candidate: %+v", blog)
	}

	search := Resolve("/search", url.Values{"q": {"go cms"}}, Settings{})
	if len(search) != 1 || search[0].Kind != CandidateSearch || search[0].Query != "go cms" {
		t.Fatalf("unexpected search candidate: %+v", search)
	}

	taxonomy := Resolve("/category/news/", url.Values{}, Settings{})
	if len(taxonomy) != 1 || taxonomy[0].Kind != CandidateTaxonomy {
		t.Fatalf("unexpected taxonomy candidate: %+v", taxonomy)
	}
	if got := taxonomy[0].Taxonomy; got != "category" {
		t.Fatalf("taxonomy = %q", got)
	}
	if got := taxonomy[0].Slug; got != "news" {
		t.Fatalf("slug = %q", got)
	}

	author := Resolve("/author/mr-gopher/", url.Values{}, Settings{})
	if len(author) != 1 || author[0].Kind != CandidateAuthor || author[0].Slug != "mr-gopher" {
		t.Fatalf("unexpected author candidate: %+v", author)
	}
}

func TestEntryPathBuildsPrettyLinks(t *testing.T) {
	publishedAt := time.Date(2026, 5, 2, 10, 0, 0, 0, time.UTC)
	post := domaincontent.Entry{
		ID:          "post-1",
		Kind:        domaincontent.KindPost,
		Slug:        domaincontent.LocalizedText{"en": "hello-go"},
		CreatedAt:   publishedAt,
		PublishedAt: &publishedAt,
	}
	page := domaincontent.Entry{
		ID:        "page-1",
		Kind:      domaincontent.KindPage,
		Slug:      domaincontent.LocalizedText{"en": "about"},
		CreatedAt: publishedAt,
	}

	if got := EntryPath(post, Settings{}); got != "/hello-go/" {
		t.Fatalf("default post path = %q", got)
	}
	if got := EntryPath(page, Settings{}); got != "/about/" {
		t.Fatalf("default page path = %q", got)
	}
	if got := EntryPath(post, Settings{PostPattern: "/%year%/%monthnum%/%day%/%postname%/"}); got != "/2026/05/02/hello-go/" {
		t.Fatalf("dated post path = %q", got)
	}
	if got := EntryPath(post, Settings{PostPattern: "/archives/%id%/"}); got != "/archives/post-1/" {
		t.Fatalf("id post path = %q", got)
	}
}

func TestEntryPathUsesRussianSlugWhenEnglishMissing(t *testing.T) {
	publishedAt := time.Date(2026, 5, 2, 10, 0, 0, 0, time.UTC)
	post := domaincontent.Entry{
		ID:          "post-ru",
		Kind:        domaincontent.KindPost,
		Slug:        domaincontent.LocalizedText{"ru": "welcome-to-garage"},
		CreatedAt:   publishedAt,
		PublishedAt: &publishedAt,
	}
	if got := EntryPath(post, Settings{}); got != "/welcome-to-garage/" {
		t.Fatalf("post path = %q, want /welcome-to-garage/", got)
	}
	if !MatchesEntry(Candidate{Kind: CandidatePostSlug, Slug: "welcome-to-garage"}, post) {
		t.Fatal("expected URL slug to match ru-only LocalizedText slug")
	}
}

func TestMatchesEntryHonoursDateComponents(t *testing.T) {
	publishedAt := time.Date(2026, 5, 2, 10, 0, 0, 0, time.UTC)
	entry := domaincontent.Entry{
		ID:          "post-1",
		Kind:        domaincontent.KindPost,
		Slug:        domaincontent.LocalizedText{"en": "hello-go"},
		CreatedAt:   publishedAt,
		PublishedAt: &publishedAt,
	}

	if !MatchesEntry(Candidate{Kind: CandidatePostSlug, Slug: "hello-go", Year: 2026, Month: 5, Day: 2}, entry) {
		t.Fatal("expected dated candidate to match")
	}
	if MatchesEntry(Candidate{Kind: CandidatePostSlug, Slug: "hello-go", Year: 2025}, entry) {
		t.Fatal("expected wrong year to fail")
	}
	if !MatchesEntry(Candidate{Kind: CandidatePostID, ID: "post-1"}, entry) {
		t.Fatal("expected id candidate to match")
	}
}
