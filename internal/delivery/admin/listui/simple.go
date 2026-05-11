package listui

import (
	"cmp"
	"slices"
	"strings"

	"github.com/fastygo/cms/internal/domain/authz"
	"github.com/fastygo/cms/internal/site/ui/blocks"
	"github.com/fastygo/panel"
)

func ApplySimpleListState(rows []blocks.SimpleListRow, state State, schema panel.TableSchema[authz.Capability]) ([]blocks.SimpleListRow, int, int, int) {
	filtered := make([]blocks.SimpleListRow, 0, len(rows))
	for _, row := range rows {
		if !matchesSimpleList(row, state) {
			continue
		}
		filtered = append(filtered, row)
	}
	sortSimpleRows(filtered, state.Sort, state.Order)
	total := len(filtered)
	page := state.Page
	if page <= 0 {
		page = 1
	}
	perPage := state.PerPage
	if perPage <= 0 {
		perPage = FirstPositive(schema.PerPage, 25)
	}
	totalPages := 1
	if total > 0 {
		totalPages = (total + perPage - 1) / perPage
	}
	if page > totalPages {
		page = totalPages
	}
	start := (page - 1) * perPage
	if start < 0 || start >= total {
		if total == 0 {
			return nil, total, totalPages, page
		}
		start = 0
	}
	end := min(start+perPage, total)
	return filtered[start:end], total, totalPages, page
}

func matchesSimpleList(row blocks.SimpleListRow, state State) bool {
	search := strings.ToLower(strings.TrimSpace(state.Search))
	if search != "" {
		blob := strings.ToLower(strings.Join([]string{row.ID, row.Label, row.Description, row.Status}, " "))
		if !strings.Contains(blob, search) {
			return false
		}
	}
	for key, value := range state.Filters {
		value = strings.ToLower(strings.TrimSpace(value))
		if value == "" {
			continue
		}
		switch key {
		case "status":
			if !strings.EqualFold(strings.TrimSpace(row.Status), value) {
				return false
			}
		}
	}
	return true
}

func sortSimpleRows(rows []blocks.SimpleListRow, sortID string, order string) {
	sortID = strings.TrimSpace(sortID)
	if sortID == "" {
		sortID = "name"
	}
	desc := strings.EqualFold(order, "desc")
	slices.SortFunc(rows, func(left blocks.SimpleListRow, right blocks.SimpleListRow) int {
		var result int
		switch sortID {
		case "status":
			result = cmp.Compare(strings.ToLower(left.Status), strings.ToLower(right.Status))
		case "description":
			result = cmp.Compare(strings.ToLower(left.Description), strings.ToLower(right.Description))
		default:
			result = cmp.Compare(strings.ToLower(left.Label), strings.ToLower(right.Label))
		}
		if desc {
			return -result
		}
		return result
	})
}
