package themes

import (
	"fmt"
	"strings"
)

type ThemeID string
type TemplateRole string
type Slot string
type AssetType string
type AssetLocation string

const (
	AssetTypeCSS   AssetType = "css"
	AssetTypeJS    AssetType = "js"
	AssetTypeFont  AssetType = "font"
	AssetTypeImage AssetType = "image"
)

const (
	AssetLocationHead    AssetLocation = "head"
	AssetLocationBodyEnd AssetLocation = "body_end"
)

type Asset struct {
	ID           string
	Type         AssetType
	Path         string
	Dependencies []string
	Load         AssetLocation
	Integrity    string
}

type SettingDefinition struct {
	Key         string
	Label       string
	Type        string
	Default     any
	Public      bool
	Validation  string
	Description string
}

type StylePreset struct {
	ID           string
	Name         string
	Description  string
	Stylesheets  []string
	Scripts      []string
	TokenJSON    string
	PreviewClass string
}

type Manifest struct {
	ID          ThemeID
	Name        string
	Version     string
	Contract    string
	Description string
	Author      string
	Templates   map[TemplateRole]string
	Assets      map[string]Asset
	Slots       []Slot
	Settings    []SettingDefinition
}

type Activation struct {
	ActiveID        ThemeID
	PreviewID       ThemeID
	ActivePresetID  string
	PreviewPresetID string
}

func ValidateManifest(manifest Manifest) error {
	switch {
	case strings.TrimSpace(string(manifest.ID)) == "":
		return fmt.Errorf("theme id is required")
	case strings.TrimSpace(manifest.Name) == "":
		return fmt.Errorf("theme name is required")
	case strings.TrimSpace(manifest.Version) == "":
		return fmt.Errorf("theme version is required")
	case strings.TrimSpace(manifest.Contract) == "":
		return fmt.Errorf("theme contract is required")
	}
	if normalized := strings.ToLower(strings.TrimSpace(string(manifest.ID))); normalized != string(manifest.ID) {
		return fmt.Errorf("theme id must be lowercase and stable")
	}
	if strings.ContainsAny(string(manifest.ID), " \t\r\n") {
		return fmt.Errorf("theme id must not contain spaces")
	}
	return nil
}
