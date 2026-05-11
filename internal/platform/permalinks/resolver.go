package permalinks

import (
	"fmt"
	"net/url"
	"path"
	"slices"
	"strconv"
	"strings"
	"time"

	domaincontent "github.com/fastygo/cms/internal/domain/content"
	"github.com/fastygo/cms/internal/platform/locales"
)

const (
	DefaultPostPattern = "/%postname%/"
	DefaultPagePattern = "/{slug}/"
	DefaultBlogPath    = "/blog/"
)

type Settings struct {
	PostPattern string
	PagePattern string
}

type CandidateKind string

const (
	CandidateHome     CandidateKind = "home"
	CandidateBlog     CandidateKind = "blog"
	CandidateSearch   CandidateKind = "search"
	CandidateTaxonomy CandidateKind = "taxonomy"
	CandidateAuthor   CandidateKind = "author"
	CandidatePageSlug CandidateKind = "page"
	CandidatePostSlug CandidateKind = "post_slug"
	CandidatePostID   CandidateKind = "post_id"
)

type Candidate struct {
	Kind     CandidateKind
	Path     string
	Slug     string
	ID       string
	Taxonomy string
	Query    string
	Year     int
	Month    int
	Day      int
}

func NormalizeSettings(settings Settings) Settings {
	return Settings{
		PostPattern: normalizePattern(settings.PostPattern, DefaultPostPattern, true),
		PagePattern: normalizePattern(settings.PagePattern, DefaultPagePattern, true),
	}
}

func Resolve(requestPath string, query url.Values, settings Settings) []Candidate {
	settings = NormalizeSettings(settings)
	cleanPath := normalizePath(requestPath)
	trimmed := strings.Trim(cleanPath, "/")
	switch {
	case cleanPath == "/":
		return []Candidate{{Kind: CandidateHome, Path: cleanPath}}
	case cleanPath == DefaultBlogPath || cleanPath == strings.TrimRight(DefaultBlogPath, "/"):
		return []Candidate{{Kind: CandidateBlog, Path: ensureTrailingSlash(cleanPath)}}
	case isAuthorPath(cleanPath):
		return []Candidate{{Kind: CandidateAuthor, Path: cleanPath, Slug: authorSlug(cleanPath)}}
	case cleanPath == "/search" || cleanPath == "/search/":
		return []Candidate{{Kind: CandidateSearch, Path: cleanPath, Query: strings.TrimSpace(query.Get("q"))}}
	case isTaxonomyPath(cleanPath, "category"):
		return []Candidate{{Kind: CandidateTaxonomy, Path: cleanPath, Taxonomy: "category", Slug: taxonomySlug(cleanPath)}}
	case isTaxonomyPath(cleanPath, "tag"):
		return []Candidate{{Kind: CandidateTaxonomy, Path: cleanPath, Taxonomy: "tag", Slug: taxonomySlug(cleanPath)}}
	}

	var candidates []Candidate
	if slug, ok := matchPagePattern(settings.PagePattern, trimmed); ok {
		candidates = append(candidates, Candidate{Kind: CandidatePageSlug, Path: cleanPath, Slug: slug})
	}
	if post, ok := matchPostPattern(settings.PostPattern, trimmed); ok {
		candidates = append(candidates, post.withPath(cleanPath))
	}
	return candidates
}

func EntryPath(entry domaincontent.Entry, settings Settings) string {
	settings = NormalizeSettings(settings)
	switch entry.Kind {
	case domaincontent.KindPage:
		return buildPagePath(entry, settings.PagePattern)
	default:
		return buildPostPath(entry, settings.PostPattern)
	}
}

func normalizePattern(value string, fallback string, trailingSlash bool) string {
	value = strings.TrimSpace(value)
	if value == "" {
		value = fallback
	}
	if !strings.HasPrefix(value, "/") {
		value = "/" + value
	}
	if trailingSlash && value != "/" && !strings.HasSuffix(value, "/") {
		value += "/"
	}
	return value
}

func normalizePath(value string) string {
	clean := path.Clean("/" + strings.TrimSpace(value))
	if clean == "." {
		return "/"
	}
	return clean
}

func isTaxonomyPath(cleanPath string, taxonomy string) bool {
	prefix := "/" + taxonomy + "/"
	if !strings.HasPrefix(cleanPath, prefix) {
		return false
	}
	slug := strings.Trim(strings.TrimPrefix(cleanPath, prefix), "/")
	return slug != ""
}

func taxonomySlug(cleanPath string) string {
	parts := strings.Split(strings.Trim(cleanPath, "/"), "/")
	if len(parts) < 2 {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

func isAuthorPath(cleanPath string) bool {
	prefix := "/author/"
	if !strings.HasPrefix(cleanPath, prefix) {
		return false
	}
	slug := strings.Trim(strings.TrimPrefix(cleanPath, prefix), "/")
	return slug != ""
}

func authorSlug(cleanPath string) string {
	parts := strings.Split(strings.Trim(cleanPath, "/"), "/")
	if len(parts) < 2 {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

func matchPagePattern(pattern string, trimmedPath string) (string, bool) {
	pattern = strings.Trim(pattern, "/")
	if pattern == "{slug}" && trimmedPath != "" && !strings.Contains(trimmedPath, "/") {
		return strings.TrimSpace(trimmedPath), true
	}
	return "", false
}

func matchPostPattern(pattern string, trimmedPath string) (Candidate, bool) {
	patternSegments := patternSegments(pattern)
	pathSegments := splitPath(trimmedPath)
	if len(patternSegments) == 0 || len(patternSegments) != len(pathSegments) {
		return Candidate{}, false
	}
	candidate := Candidate{Kind: CandidatePostSlug}
	for i, segment := range patternSegments {
		value := pathSegments[i]
		switch segment {
		case "%postname%":
			candidate.Slug = value
		case "%id%":
			candidate.Kind = CandidatePostID
			candidate.ID = value
		case "%year%":
			year, err := strconv.Atoi(value)
			if err != nil || year <= 0 {
				return Candidate{}, false
			}
			candidate.Year = year
		case "%monthnum%":
			month, err := strconv.Atoi(value)
			if err != nil || month < 1 || month > 12 {
				return Candidate{}, false
			}
			candidate.Month = month
		case "%day%":
			day, err := strconv.Atoi(value)
			if err != nil || day < 1 || day > 31 {
				return Candidate{}, false
			}
			candidate.Day = day
		default:
			if segment != value {
				return Candidate{}, false
			}
		}
	}
	if candidate.Kind == CandidatePostSlug && candidate.Slug == "" {
		return Candidate{}, false
	}
	if candidate.Kind == CandidatePostID && candidate.ID == "" {
		return Candidate{}, false
	}
	return candidate, true
}

// slugForPermalink returns a non-empty slug for URL building, independent of which
// LocalizedText key the editor stored (admin uses locales.Default, often "ru").
func slugForPermalink(entry domaincontent.Entry) string {
	active := locales.Default
	fb := locales.ContentFallback(active)
	if s := strings.TrimSpace(entry.Slug.Value(active, fb)); s != "" {
		return s
	}
	if s := strings.TrimSpace(entry.Slug.Value(fb, active)); s != "" {
		return s
	}
	keys := make([]string, 0, len(entry.Slug))
	for k := range entry.Slug {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	for _, k := range keys {
		if s := strings.TrimSpace(entry.Slug[k]); s != "" {
			return s
		}
	}
	return ""
}

// entrySlugMatchesURL reports whether the path slug from the request matches this entry
// (any localized slug value).
func entrySlugMatchesURL(requestSlug string, entry domaincontent.Entry) bool {
	requestSlug = strings.TrimSpace(requestSlug)
	if requestSlug == "" {
		return false
	}
	for _, v := range entry.Slug {
		if strings.TrimSpace(v) == requestSlug {
			return true
		}
	}
	return false
}

func buildPagePath(entry domaincontent.Entry, pattern string) string {
	slug := slugForPermalink(entry)
	if slug == "" {
		return "/"
	}
	return ensureTrailingSlash(strings.ReplaceAll(pattern, "{slug}", slug))
}

func buildPostPath(entry domaincontent.Entry, pattern string) string {
	slug := slugForPermalink(entry)
	if slug == "" {
		return "/"
	}
	point := entry.CreatedAt
	if entry.PublishedAt != nil {
		point = *entry.PublishedAt
	}
	replacer := strings.NewReplacer(
		"%postname%", slug,
		"%id%", string(entry.ID),
		"%year%", fmt.Sprintf("%04d", point.UTC().Year()),
		"%monthnum%", fmt.Sprintf("%02d", int(point.UTC().Month())),
		"%day%", fmt.Sprintf("%02d", point.UTC().Day()),
	)
	return ensureTrailingSlash(replacer.Replace(pattern))
}

func MatchesEntry(candidate Candidate, entry domaincontent.Entry) bool {
	point := entry.CreatedAt
	if entry.PublishedAt != nil {
		point = *entry.PublishedAt
	}
	switch candidate.Kind {
	case CandidatePostID:
		return string(entry.ID) == candidate.ID
	case CandidatePostSlug:
		if !entrySlugMatchesURL(candidate.Slug, entry) {
			return false
		}
		if candidate.Year != 0 && point.UTC().Year() != candidate.Year {
			return false
		}
		if candidate.Month != 0 && int(point.UTC().Month()) != candidate.Month {
			return false
		}
		if candidate.Day != 0 && point.UTC().Day() != candidate.Day {
			return false
		}
		return true
	default:
		return false
	}
}

func ensureTrailingSlash(value string) string {
	value = normalizePath(value)
	if value != "/" && !strings.HasSuffix(value, "/") {
		value += "/"
	}
	return value
}

func patternSegments(pattern string) []string {
	return splitPath(strings.Trim(pattern, "/"))
}

func splitPath(value string) []string {
	value = strings.TrimSpace(strings.Trim(value, "/"))
	if value == "" {
		return nil
	}
	parts := strings.Split(value, "/")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func (c Candidate) withPath(path string) Candidate {
	c.Path = path
	return c
}

func PublishTime(entry domaincontent.Entry) time.Time {
	if entry.PublishedAt != nil {
		return *entry.PublishedAt
	}
	return entry.CreatedAt
}
