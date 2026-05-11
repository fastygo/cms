package media

import "time"

type ID string

type BlobRef struct {
	Provider string `json:"provider"`
	Key      string `json:"key"`
	URL      string `json:"url"`
	Checksum string `json:"checksum"`
	ETag     string `json:"etag"`
}

type Variant struct {
	Name   string `json:"name"`
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

type Asset struct {
	ID          ID             `json:"id"`
	Filename    string         `json:"filename"`
	MimeType    string         `json:"mime_type"`
	SizeBytes   int64          `json:"size_bytes"`
	Width       int            `json:"width"`
	Height      int            `json:"height"`
	AltText     string         `json:"alt_text"`
	Caption     string         `json:"caption"`
	PublicURL   string         `json:"public_url"`
	ProviderRef BlobRef        `json:"provider_ref"`
	PrivateMeta map[string]any `json:"private_metadata,omitempty"`
	PublicMeta  map[string]any `json:"metadata,omitempty"`
	Variants    []Variant      `json:"variants,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

type AttachmentRef struct {
	AssetID ID
	Role    string
}
