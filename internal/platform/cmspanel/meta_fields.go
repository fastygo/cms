package cmspanel

import (
	appmeta "github.com/fastygo/cms/internal/application/meta"
	domainmeta "github.com/fastygo/cms/internal/domain/meta"
	"github.com/fastygo/panel"
)

func MetadataFields(definitions []domainmeta.Definition) []panel.Field {
	fields := make([]panel.Field, 0, len(definitions)+3)
	for _, definition := range definitions {
		field := panel.Field{
			ID:          appmeta.FormFieldName(definition.Key),
			Label:       definition.Label,
			Description: definition.Description,
			Placeholder: definition.Placeholder,
			Required:    definition.Required,
		}
		switch {
		case len(definition.Options) > 0:
			field.Type = panel.FieldSelect
			field.Options = options(definition.Options)
		case definition.FieldHint == domainmeta.FieldHintTextarea:
			field.Type = panel.FieldTextarea
		case definition.FieldHint == domainmeta.FieldHintCheckbox:
			field.Type = panel.FieldBoolean
		case definition.FieldHint == domainmeta.FieldHintNumber:
			field.Type = panel.FieldNumber
		default:
			field.Type = panel.FieldText
		}
		fields = append(fields, field)
	}
	fields = append(fields,
		panel.Field{ID: "custom_meta_key", Label: "Custom metadata key", Type: panel.FieldText, Placeholder: "plugin.custom_key", Description: "Optional fallback key for unregistered metadata; saved as private by default."},
		panel.Field{ID: "custom_meta_value", Label: "Custom metadata value", Type: panel.FieldText},
		panel.Field{ID: "custom_meta_public", Label: "Expose custom metadata publicly", Type: panel.FieldBoolean, Description: "Only applies to unregistered custom metadata keys."},
	)
	return fields
}

func options(values []domainmeta.Option) []panel.Option {
	result := make([]panel.Option, 0, len(values))
	for _, value := range values {
		result = append(result, panel.Option{Value: value.Value, Label: value.Label})
	}
	return result
}
