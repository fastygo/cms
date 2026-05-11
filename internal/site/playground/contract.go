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
	Version  string                     `json:"snapshot_version"`
	Source   Source                     `json:"source"`
	Routes   map[string]json.RawMessage `json:"routes"`
	Settings []SnapshotSetting          `json:"settings,omitempty"`
	Local    SnapshotLocal              `json:"local"`
}

type SnapshotLocal struct {
	MediaBlobs    string          `json:"media_blobs"`
	MediaMetadata []MediaMetadata `json:"media_metadata,omitempty"`
}

type SnapshotSetting struct {
	Key    string          `json:"key"`
	Value  json.RawMessage `json:"value"`
	Public bool            `json:"public"`
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

type Blueprint struct {
	Version string        `json:"blueprint_version"`
	Name    string        `json:"name,omitempty"`
	Launch  LaunchOptions `json:"launch"`
}

type LaunchOptions struct {
	Version     string `json:"launch_version"`
	SourceURL   string `json:"source_url,omitempty"`
	SnapshotURL string `json:"snapshot_url,omitempty"`
	InitialPath string `json:"initial_path,omitempty"`
	Theme       string `json:"theme,omitempty"`
	Preset      string `json:"preset,omitempty"`
	DemoMode    bool   `json:"demo_mode,omitempty"`
	Embedded    bool   `json:"embedded,omitempty"`
}

const (
	DefaultSnapshotVersion  = "gocms.playground.v1"
	DefaultBlueprintVersion = "gocms.playground.blueprint.v1"
	DefaultLaunchVersion    = "gocms.playground.launch.v1"
	IndexedDBDatabaseName   = "gocms-playground"
	SnapshotStore           = "snapshots"
	MediaMetadataStore      = "media_metadata"
	MediaBlobStore          = "media_blobs"
	SettingsStore           = "settings"
	BlobStatusExcluded      = "excluded"
	BlobStatusLocalOnly     = "local-only"
	BlobStatusMissing       = "missing-local-blob"
	RoutePosts              = "/wp-json/wp/v2/posts"
	RoutePages              = "/wp-json/wp/v2/pages"
	RouteCategories         = "/wp-json/wp/v2/categories"
	RouteTags               = "/wp-json/wp/v2/tags"
	RouteMedia              = "/wp-json/wp/v2/media"
)
