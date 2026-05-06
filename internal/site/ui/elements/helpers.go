package elements

import (
	"fmt"
	"strings"
)

func statusLabel(status string) string {
	if strings.TrimSpace(status) == "" {
		return "unknown"
	}
	return status
}

func statusClass(status string) string {
	switch strings.TrimSpace(status) {
	case "published":
		return "gocms-status-badge--published"
	case "scheduled":
		return "gocms-status-badge--scheduled"
	case "trashed":
		return "gocms-status-badge--trashed"
	default:
		return "gocms-status-badge--draft"
	}
}

func buttonVariant(style string) string {
	switch strings.TrimSpace(style) {
	case "danger":
		return "destructive"
	case "secondary":
		return "outline"
	case "link":
		return "link"
	default:
		return "default"
	}
}

func pageHref(base string, page int) string {
	if page <= 0 {
		page = 1
	}
	sep := "?"
	if strings.Contains(base, "?") {
		sep = "&"
	}
	return fmt.Sprintf("%s%spage=%d", base, sep, page)
}

func pageLabel(page int, total int) string {
	return fmt.Sprintf("Page %d of %d", page, total)
}

func paginationPreviousLabel(value string) string {
	if strings.TrimSpace(value) != "" {
		return value
	}
	return "Previous"
}

func paginationNextLabel(value string) string {
	if strings.TrimSpace(value) != "" {
		return value
	}
	return "Next"
}
