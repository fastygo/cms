package ui_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestGoCMSAppTemplatesDoNotUseRawTags(t *testing.T) {
	rawTag := regexp.MustCompile(`(^|[^\w])</?[a-z][\w:-]*(\s|>|/)`)
	allowedSuppliers := map[string]struct{}{
		filepath.FromSlash("internal/site/ui/elements/markers.templ"):    {},
		filepath.FromSlash("internal/site/ui/blocks/auth_document.templ"): {},
	}
	forEachFile(t, filepath.FromSlash("../../../internal/site"), func(path string, content string) {
		if !strings.HasSuffix(path, ".templ") {
			return
		}
		normalized := filepath.ToSlash(path)
		for suffix := range allowedSuppliers {
			if strings.HasSuffix(normalized, filepath.ToSlash(suffix)) {
				return
			}
		}
		for index, line := range strings.Split(content, "\n") {
			if rawTag.MatchString(strings.TrimSpace(line)) {
				t.Fatalf("%s:%d uses raw HTML tag; use UI8Kit/local UI suppliers", path, index+1)
			}
		}
	})
}

func TestGoCMSUIImportBoundaries(t *testing.T) {
	forEachFile(t, filepath.FromSlash("../../../internal/site/ui/elements"), func(path string, content string) {
		if strings.Contains(content, "github.com/fastygo/ui8kit/components") {
			t.Fatalf("%s imports UI8Kit composites from elements layer", path)
		}
	})
	forEachFile(t, filepath.FromSlash("../../../internal/site/views"), func(path string, content string) {
		if strings.Contains(content, "github.com/fastygo/ui8kit/components") {
			t.Fatalf("%s imports UI8Kit composites directly from views", path)
		}
	})
}

func TestGoCMSCSSUsesApplyOnlySelectors(t *testing.T) {
	declaration := regexp.MustCompile(`^\s*[a-z-]+\s*:`)
	forEachFile(t, filepath.FromSlash("../../../web/static/css"), func(path string, content string) {
		if !strings.HasSuffix(path, ".css") || strings.HasSuffix(path, "tokens.css") || strings.HasSuffix(path, "input.css") {
			return
		}
		if strings.Contains(filepath.ToSlash(path), "/ui8kit/") || strings.HasSuffix(path, "fonts.css") {
			return
		}
		for index, line := range strings.Split(content, "\n") {
			trimmed := strings.TrimSpace(line)
			if trimmed == "" || strings.HasPrefix(trimmed, "/*") || strings.HasPrefix(trimmed, "*") || strings.HasPrefix(trimmed, "*/") {
				continue
			}
			if declaration.MatchString(line) && !strings.HasPrefix(trimmed, "@apply") {
				t.Fatalf("%s:%d uses raw CSS declaration; app selectors must use @apply only", path, index+1)
			}
		}
	})
}

func TestGoCMSDoesNotUseRawPaletteUtilities(t *testing.T) {
	banned := []string{"bg-blue-", "text-slate-", "border-red-", "from-purple-", "ring-zinc-"}
	forEachFile(t, filepath.FromSlash("../../../internal/site"), func(path string, content string) {
		if strings.HasSuffix(path, "architecture_test.go") {
			return
		}
		for _, token := range banned {
			if strings.Contains(content, token) {
				t.Fatalf("%s uses raw palette utility %q", path, token)
			}
		}
	})
	forEachFile(t, filepath.FromSlash("../../../web/static/css"), func(path string, content string) {
		if strings.HasSuffix(path, "app.css") || isVersionedAppCSS(path) {
			return
		}
		for _, token := range banned {
			if strings.Contains(content, token) {
				t.Fatalf("%s uses raw palette utility %q", path, token)
			}
		}
	})
}

func isVersionedAppCSS(path string) bool {
	name := filepath.Base(path)
	return regexp.MustCompile(`^app\.[a-f0-9]{12}\.css$`).MatchString(name)
}

func forEachFile(t *testing.T, root string, fn func(path string, content string)) {
	t.Helper()
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		fn(path, string(content))
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
