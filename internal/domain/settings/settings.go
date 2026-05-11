package settings

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/fastygo/cms/internal/domain/authz"
)

type Key string

type Group string

const (
	GroupCore        Group = "core"
	GroupTheme       Group = "theme"
	GroupPlugin      Group = "plugin"
	GroupPublic      Group = "public"
	GroupHeadless    Group = "headless"
	GroupOperational Group = "operational"
	GroupAdmin       Group = "admin"
)

type ValueType string

const (
	ValueTypeString  ValueType = "string"
	ValueTypeText    ValueType = "text"
	ValueTypeBoolean ValueType = "boolean"
	ValueTypeInteger ValueType = "integer"
	ValueTypeSelect  ValueType = "select"
	ValueTypeJSON    ValueType = "json"
)

type FieldHint string

const (
	FieldHintText     FieldHint = "text"
	FieldHintTextarea FieldHint = "textarea"
	FieldHintCheckbox FieldHint = "checkbox"
	FieldHintNumber   FieldHint = "number"
	FieldHintSelect   FieldHint = "select"
)

type AutoloadPolicy string

const (
	AutoloadLazy   AutoloadPolicy = "lazy"
	AutoloadEager  AutoloadPolicy = "eager"
	AutoloadManual AutoloadPolicy = "manual"
)

type Option struct {
	Value string
	Label string
}

type ValidationRule struct {
	Name string
	Arg  string
}

type Definition struct {
	Key         Key
	Label       string
	Owner       string
	Group       Group
	Type        ValueType
	Public      bool
	Default     any
	Description string
	Capability  authz.Capability
	Autoload    AutoloadPolicy
	Rules       []ValidationRule
	FieldHint   FieldHint
	Placeholder string
	Options     []Option
}

type Value struct {
	Key    Key
	Value  any
	Public bool
}

func (d Definition) Validate() error {
	switch {
	case strings.TrimSpace(string(d.Key)) == "":
		return fmt.Errorf("setting key is required")
	case strings.TrimSpace(d.Label) == "":
		return fmt.Errorf("setting %q label is required", d.Key)
	case strings.TrimSpace(d.Owner) == "":
		return fmt.Errorf("setting %q owner is required", d.Key)
	case d.Group == "":
		return fmt.Errorf("setting %q group is required", d.Key)
	case d.Type == "":
		return fmt.Errorf("setting %q type is required", d.Key)
	case d.Autoload == "":
		return fmt.Errorf("setting %q autoload policy is required", d.Key)
	}
	return nil
}

func ParseBoolean(value string) (bool, error) {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case "", "0", "false", "off", "no":
		return false, nil
	case "1", "true", "on", "yes":
		return true, nil
	default:
		return false, fmt.Errorf("must be a boolean")
	}
}

func ParseInteger(value string) (int, error) {
	parsed, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		return 0, fmt.Errorf("must be an integer")
	}
	return parsed, nil
}
