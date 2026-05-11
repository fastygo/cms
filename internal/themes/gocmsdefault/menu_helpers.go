package gocmsdefault

import (
	"github.com/fastygo/cms/internal/application/publicrender"
	ui8layout "github.com/fastygo/ui8kit/layout"
)

// flattenPublicMenu returns primary menu items in depth-first order for flat nav UIs.
func flattenPublicMenu(items []publicrender.MenuItem) []publicrender.MenuItem {
	out := make([]publicrender.MenuItem, 0, 16)
	var walk func([]publicrender.MenuItem)
	walk = func(nodes []publicrender.MenuItem) {
		for _, n := range nodes {
			c := n
			c.Children = nil
			out = append(out, c)
			if len(n.Children) > 0 {
				walk(n.Children)
			}
		}
	}
	walk(items)
	return out
}

// publicMenuNavItems maps the primary menu into ui8kit shell sidebar / mobile sheet links.
func publicMenuNavItems(items []publicrender.MenuItem) []ui8layout.NavItem {
	flat := flattenPublicMenu(items)
	out := make([]ui8layout.NavItem, len(flat))
	for i, it := range flat {
		out[i] = ui8layout.NavItem{Path: it.URL, Label: it.Label}
	}
	return out
}

func publicHeaderNavLinkClass(active bool) string {
	if active {
		return "ui-header-nav-link ui-header-nav-link-active"
	}
	return "ui-header-nav-link"
}

func footerNavLinkClass(active bool) string {
	if active {
		return "text-sm font-medium text-foreground underline underline-offset-4 hover:underline"
	}
	return "text-sm text-muted-foreground no-underline underline-offset-4 hover:text-foreground hover:underline"
}
