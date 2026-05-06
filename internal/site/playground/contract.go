package playground

import (
	"encoding/json"
	"time"
)

type Source struct {
	Kind     string    `json:"kind"`
	BaseURL  string    `json:"base_url"`
	Imported time.Time `json:"imported_at"`
}

type Snapshot struct {
	Version string                     `json:"snapshot_version"`
	Source  Source                     `json:"source"`
	Routes  map[string]json.RawMessage `json:"routes"`
	Local   SnapshotLocal              `json:"local"`
}

type SnapshotLocal struct {
	MediaBlobs    string          `json:"media_blobs"`
	MediaMetadata []MediaMetadata `json:"media_metadata,omitempty"`
}

type MediaMetadata struct {
	ID         string    `json:"id"`
	Filename   string    `json:"filename"`
	MimeType   string    `json:"mime_type"`
	Width      int       `json:"width"`
	Height     int       `json:"height"`
	Size       int64     `json:"size"`
	Alt        string    `json:"alt"`
	Caption    string    `json:"caption"`
	CreatedAt  time.Time `json:"created_at"`
	AttachedTo string    `json:"attached_to"`
	BlobStatus string    `json:"blob_status"`
}

const (
	DefaultSnapshotVersion = "gocms.playground.v1"
	IndexedDBDatabaseName  = "gocms-playground"
	SnapshotStore          = "snapshots"
	MediaMetadataStore     = "media_metadata"
	MediaBlobStore         = "media_blobs"
	SettingsStore          = "settings"
	BlobStatusLocalOnly    = "local-only"
	BlobStatusMissing      = "missing-local-blob"
)
