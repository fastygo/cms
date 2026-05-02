package media

import "time"

type ID string

type Variant struct {
	Name   string
	URL    string
	Width  int
	Height int
}

type Asset struct {
	ID          ID
	Filename    string
	MimeType    string
	SizeBytes   int64
	Width       int
	Height      int
	AltText     string
	Caption     string
	PublicURL   string
	PrivateMeta map[string]any
	PublicMeta  map[string]any
	Variants    []Variant
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type AttachmentRef struct {
	AssetID ID
	Role    string
}
