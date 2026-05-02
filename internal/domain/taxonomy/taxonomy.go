package taxonomy

import (
	"fmt"
	"strings"

	"github.com/fastygo/cms/internal/domain/content"
)

type Type string
type Mode string
type TermID string

const (
	TypeCategory Type = "category"
	TypeTag      Type = "tag"
)

const (
	ModeHierarchical Mode = "hierarchical"
	ModeFlat         Mode = "flat"
)

type Definition struct {
	Type            Type
	Label           string
	Mode            Mode
	AssignedToKinds []content.Kind
	Public          bool
	RESTVisible     bool
	GraphQLVisible  bool
}

type Term struct {
	ID          TermID
	Type        Type
	Name        content.LocalizedText
	Slug        content.LocalizedText
	Description content.LocalizedText
	ParentID    TermID
}

type Assignment struct {
	ContentID content.ID
	TermID    TermID
	Type      Type
}

func ValidateType(value Type) error {
	if strings.TrimSpace(string(value)) == "" {
		return fmt.Errorf("taxonomy type is required")
	}
	return nil
}
