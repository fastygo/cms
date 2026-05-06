package blocks

import "github.com/fastygo/ui8kit/ui"

func toFieldProps(field FieldData) ui.FieldProps {
	component := field.Component
	if component == "" && field.Rows > 0 {
		component = "textarea"
	}
	return ui.FieldProps{
		ID:          field.ID,
		Name:        field.Name,
		Label:       field.Label,
		Value:       field.Value,
		Type:        field.Type,
		Component:   component,
		Placeholder: field.Placeholder,
		Required:    field.Required,
		Rows:        field.Rows,
		Options:     field.Options,
		Hint:        field.Hint,
	}
}

func statusOptions(options []ui.FieldOption) []ui.FieldOption {
	if len(options) > 0 {
		return options
	}
	return []ui.FieldOption{
		{Value: "draft", Label: "Draft"},
		{Value: "published", Label: "Published"},
		{Value: "scheduled", Label: "Scheduled"},
		{Value: "trashed", Label: "Trashed"},
	}
}

func fallbackLabel(value string, fallback string) string {
	if value != "" {
		return value
	}
	return fallback
}
