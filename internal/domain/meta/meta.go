package meta

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/fastygo/cms/internal/domain/authz"
	domaincontent "github.com/fastygo/cms/internal/domain/content"
)

type Scope string

const (
	ScopeContent Scope = "content"
)

type ValueType string

const (
	ValueTypeString  ValueType = "string"
	ValueTypeText    ValueType = "text"
	ValueTypeURL     ValueType = "url"
	ValueTypeBoolean ValueType = "boolean"
	ValueTypeInteger ValueType = "integer"
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

type Option struct {
	Value string
	Label string
}

type ValidationRule struct {
	Name string
	Arg  string
}

type Definition struct {
	Key         string
	Label       string
	Owner       string
	Scope       Scope
	Kinds       []domaincontent.Kind
	Type        ValueType
	FieldHint   FieldHint
	Public      bool
	Required    bool
	Default     any
	Capability  authz.Capability
	Rules       []ValidationRule
	Description string
	Placeholder string
	Options     []Option
}

func (d Definition) AppliesTo(kind domaincontent.Kind) bool {
	if len(d.Kinds) == 0 {
		return true
	}
	for _, candidate := range d.Kinds {
		if candidate == kind {
			return true
		}
	}
	return false
}

func (d Definition) Validate() error {
	switch {
	case strings.TrimSpace(d.Key) == "":
		return fmt.Errorf("meta key is required")
	case !validKey(d.Key):
		return fmt.Errorf("meta key %q is invalid", d.Key)
	case strings.TrimSpace(d.Label) == "":
		return fmt.Errorf("meta %q label is required", d.Key)
	case strings.TrimSpace(d.Owner) == "":
		return fmt.Errorf("meta %q owner is required", d.Key)
	case d.Scope == "":
		return fmt.Errorf("meta %q scope is required", d.Key)
	case d.Type == "":
		return fmt.Errorf("meta %q type is required", d.Key)
	}
	if d.FieldHint == "" {
		switch d.Type {
		case ValueTypeBoolean:
			d.FieldHint = FieldHintCheckbox
		case ValueTypeInteger:
			d.FieldHint = FieldHintNumber
		case ValueTypeText, ValueTypeJSON:
			d.FieldHint = FieldHintTextarea
		default:
			d.FieldHint = FieldHintText
		}
	}
	for _, rule := range d.Rules {
		switch strings.TrimSpace(strings.ToLower(rule.Name)) {
		case "", "required", "min_length", "max_length", "one_of":
		default:
			return fmt.Errorf("meta %q validation rule %q is not supported", d.Key, rule.Name)
		}
	}
	return nil
}

func validKey(value string) bool {
	for _, char := range value {
		switch {
		case char >= 'a' && char <= 'z':
		case char >= '0' && char <= '9':
		case char == '-', char == '_', char == '.':
		default:
			return false
		}
	}
	return true
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
