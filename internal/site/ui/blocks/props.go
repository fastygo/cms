package blocks

import (
	"github.com/fastygo/cms/internal/site/ui/elements"
	"github.com/fastygo/ui8kit/ui"
)

type StatCard struct {
	Label string
	Value string
	Href  string
}

type LoginPanelData struct {
	Title       string
	Subtitle    string
	Error       string
	ReturnTo    string
	ActionToken string
}

type ContentRow struct {
	ID      string
	Title   string
	Slug    string
	Status  string
	Author  string
	EditURL string
}

type ContentTableData struct {
	Title       string
	Description string
	Rows        []ContentRow
	Actions     []elements.Action
	Pagination  elements.PaginationData
}

type PanelData struct {
	Title       string
	Description string
	Marker      string
}

type FieldData struct {
	ID          string
	Name        string
	Label       string
	Value       string
	Type        string
	Component   string
	Placeholder string
	Required    bool
	Rows        int
	Options     []ui.FieldOption
}

type ContentEditorData struct {
	Title       string
	Description string
	Action      string
	Token       string
	Fields      []FieldData
	Status      string
	Actions     []elements.Action
	Errors      []elements.FieldError
}

type SimpleListRow struct {
	Label       string
	Description string
	Status      string
	ActionURL   string
}

type SimpleListData struct {
	Title       string
	Description string
	Marker      string
	Rows        []SimpleListRow
	Actions     []elements.Action
	FormAction  string
	Token       string
	Fields      []FieldData
}
