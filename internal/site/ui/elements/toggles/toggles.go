package toggles

import (
	"strings"

	"github.com/a-h/templ"
	"github.com/fastygo/framework/pkg/web/view"
	"github.com/fastygo/ui8kit/ui"
)

type HeaderLanguageToggleProps struct {
	Toggle view.LanguageToggleData
}

func HeaderLanguageToggle(props HeaderLanguageToggleProps) templ.Component {
	return LanguageToggle(props.Toggle)
}

func LanguageToggle(data view.LanguageToggleData) templ.Component {
	if strings.TrimSpace(data.CurrentLabel) == "" && strings.TrimSpace(data.CurrentLocale) == "" {
		return templ.NopComponent
	}
	return ui.Button(ui.ButtonProps{
		ID:        languageToggleID(""),
		Href:      data.NextHref,
		Variant:   "unstyled",
		Class:     "ui-header-action-btn",
		AriaLabel: languageToggleAriaLabel(data),
		Attrs:     languageToggleAttrs(data),
	}, languageToggleLabel(data.CurrentLabel, data.CurrentLocale))
}

func languageToggleID(value string) string {
	if strings.TrimSpace(value) != "" {
		return value
	}
	return "language-toggle"
}

func languageToggleLabel(label, locale string) string {
	if strings.TrimSpace(label) != "" {
		return label
	}
	return strings.ToUpper(strings.TrimSpace(locale))
}

func languageToggleAriaLabel(data view.LanguageToggleData) string {
	if strings.TrimSpace(data.Label) != "" {
		return data.Label
	}
	return "Switch language"
}

func languageToggleAttrs(data view.LanguageToggleData) templ.Attributes {
	attrs := templ.Attributes{
		"data-default-locale":    data.DefaultLocale,
		"data-current-locale":    data.CurrentLocale,
		"data-next-locale":       data.NextLocale,
		"data-next-label":        data.NextLabel,
		"data-available-locales": strings.Join(data.AvailableLocales, ","),
	}
	if data.EnhanceWithJS {
		attrs["data-ui8kit-spa-lang"] = "1"
		attrs["data-spa-target"] = data.SPATarget
	}
	return attrs
}
