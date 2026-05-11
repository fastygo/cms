package listui

import (
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"

	"github.com/fastygo/cms/internal/domain/authz"
	"github.com/fastygo/panel"
)

type State struct {
	Screen         string
	Page           int
	PerPage        int
	Search         string
	Sort           string
	Order          string
	Filters        map[string]string
	EditID         string
	VisibleColumns []string
}

func ParseState(r *http.Request, screen string, schema panel.TableSchema[authz.Capability], defaultPerPage int, preferredPerPage int, preferredColumns []string) State {
	if defaultPerPage <= 0 {
		defaultPerPage = FirstPositive(schema.PerPage, 25)
	}
	perPage := defaultPerPage
	if preferredPerPage > 0 {
		perPage = preferredPerPage
	}
	perPage = normalizePerPage(r.URL.Query().Get("per_page"), schema.PerPage, perPage)
	sortID := normalizeSort(r.URL.Query().Get("sort"), schema)
	if sortID == "" {
		sortID = strings.TrimSpace(schema.DefaultSort.ColumnID)
	}
	order := normalizeOrder(r.URL.Query().Get("order"), schema.DefaultSort.Direction)
	state := State{
		Screen:         screen,
		Page:           positiveInt(r.URL.Query().Get("page"), 1),
		PerPage:        perPage,
		Search:         strings.TrimSpace(firstNonEmpty(r.URL.Query().Get("search"), r.URL.Query().Get("q"))),
		Sort:           sortID,
		Order:          order,
		Filters:        map[string]string{},
		EditID:         strings.TrimSpace(r.URL.Query().Get("edit")),
		VisibleColumns: normalizeVisibleColumns(schema, firstNonEmpty(r.URL.Query().Get("columns"), strings.Join(preferredColumns, ","))),
	}
	for _, filter := range schema.Filters {
		key := "filter_" + filter.ID
		value := strings.TrimSpace(r.URL.Query().Get(key))
		if value == "" {
			value = strings.TrimSpace(filter.Default)
		}
		if value == "" {
			continue
		}
		if filter.Type == panel.FilterSelect && !filterOptionAllowed(filter, value) {
			continue
		}
		state.Filters[filter.ID] = value
	}
	return state
}

func (s State) Values() url.Values {
	values := url.Values{}
	if s.Search != "" {
		values.Set("search", s.Search)
	}
	if s.PerPage > 0 {
		values.Set("per_page", strconv.Itoa(s.PerPage))
	}
	if s.Sort != "" {
		values.Set("sort", s.Sort)
	}
	if s.Order != "" {
		values.Set("order", s.Order)
	}
	if len(s.VisibleColumns) > 0 {
		values.Set("columns", strings.Join(s.VisibleColumns, ","))
	}
	for key, value := range s.Filters {
		if strings.TrimSpace(value) != "" {
			values.Set("filter_"+key, value)
		}
	}
	return values
}

func (s State) BaseHref(basePath string) string {
	values := s.Values()
	values.Del("page")
	encoded := values.Encode()
	if encoded == "" {
		return basePath
	}
	return basePath + "?" + encoded
}

func (s State) ReturnTo(basePath string) string {
	values := s.Values()
	values.Set("page", strconv.Itoa(s.Page))
	encoded := values.Encode()
	if encoded == "" {
		return basePath
	}
	return basePath + "?" + encoded
}

func (s State) VisibleMap(schema panel.TableSchema[authz.Capability]) map[string]bool {
	if len(s.VisibleColumns) == 0 {
		return defaultVisibleColumns(schema)
	}
	visible := map[string]bool{}
	for _, column := range s.VisibleColumns {
		visible[column] = true
	}
	return visible
}

func BuildEditHref(state State, basePath string, id string) string {
	values := state.Values()
	values.Set("edit", id)
	encoded := values.Encode()
	if encoded == "" {
		return basePath + "?edit=" + id
	}
	return basePath + "?" + encoded
}

func SelectedIDs(r *http.Request) []string {
	result := []string{}
	for _, id := range r.PostForm["selected_id"] {
		id = strings.TrimSpace(id)
		if id != "" && !slices.Contains(result, id) {
			result = append(result, id)
		}
	}
	return result
}

func FirstPositive(values []int, fallback int) int {
	if len(values) > 0 && values[0] > 0 {
		return values[0]
	}
	return fallback
}

func positiveInt(value string, fallback int) int {
	parsed, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}

func normalizePerPage(value string, allowed []int, fallback int) int {
	requested := positiveInt(value, fallback)
	if len(allowed) == 0 {
		return requested
	}
	for _, item := range allowed {
		if item == requested {
			return requested
		}
	}
	return fallback
}

func normalizeSort(value string, schema panel.TableSchema[authz.Capability]) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if value == schema.DefaultSort.ColumnID {
		return value
	}
	for _, column := range schema.Columns {
		if column.ID == value && column.Sortable {
			return value
		}
	}
	return ""
}

func normalizeOrder(value string, fallback panel.SortDirection) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case string(panel.SortAsc):
		return string(panel.SortAsc)
	case string(panel.SortDesc):
		return string(panel.SortDesc)
	}
	if fallback == "" {
		return string(panel.SortDesc)
	}
	return string(fallback)
}

func normalizeVisibleColumns(schema panel.TableSchema[authz.Capability], raw string) []string {
	allowed := map[string]bool{}
	defaults := []string{}
	for _, column := range schema.Columns {
		if column.ID == "" {
			continue
		}
		allowed[column.ID] = true
		if !column.Toggleable {
			defaults = append(defaults, column.ID)
		}
	}
	result := append([]string(nil), defaults...)
	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		if part == "" || !allowed[part] || slices.Contains(result, part) {
			continue
		}
		result = append(result, part)
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

func defaultVisibleColumns(schema panel.TableSchema[authz.Capability]) map[string]bool {
	visible := map[string]bool{}
	for _, column := range schema.Columns {
		visible[column.ID] = true
	}
	return visible
}

func filterOptionAllowed(filter panel.Filter, value string) bool {
	if len(filter.Options) == 0 {
		return true
	}
	for _, option := range filter.Options {
		if option.Value == value {
			return true
		}
	}
	return false
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
