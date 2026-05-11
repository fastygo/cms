package playground

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	appsnapshot "github.com/fastygo/cms/internal/application/snapshot"
	domaincontent "github.com/fastygo/cms/internal/domain/content"
	domainmedia "github.com/fastygo/cms/internal/domain/media"
	domainsettings "github.com/fastygo/cms/internal/domain/settings"
)

func TestSnapshotExportsMediaMetadataWithoutBinaryPayload(t *testing.T) {
	snapshot := Snapshot{
		Version: DefaultSnapshotVersion,
		Source:  Source{Kind: "wp-json", BaseURL: "https://example.test/wp-json", Imported: time.Unix(1, 0).UTC()},
		Routes: map[string]json.RawMessage{
			"/wp-json/wp/v2/posts": json.RawMessage(`[]`),
		},
		Local: SnapshotLocal{
			MediaBlobs: BlobStatusLocalOnly,
			MediaMetadata: []MediaMetadata{{
				ID: "media-1", Filename: "hero.jpg", MimeType: "image/jpeg",
				Width: 1200, Height: 800, Size: 42, BlobStatus: BlobStatusLocalOnly,
			}},
		},
	}

	payload, err := json.Marshal(snapshot)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	body := string(payload)
	if containsBinaryPayload(body) {
		t.Fatalf("snapshot should not include binary payload data: %s", body)
	}
	if !json.Valid(payload) {
		t.Fatalf("snapshot JSON is invalid")
	}
	if !strings.Contains(body, `"media_blobs":"local-only"`) || !strings.Contains(body, `"filename":"hero.jpg"`) {
		t.Fatalf("snapshot should include media metadata without blobs: %s", body)
	}
}

func TestBlueprintAndLaunchContractsMarshalStableFields(t *testing.T) {
	blueprint := Blueprint{
		Version: DefaultBlueprintVersion,
		Name:    "demo",
		Launch: LaunchOptions{
			Version:     DefaultLaunchVersion,
			SourceURL:   "https://example.test/wp-json",
			SnapshotURL: "https://example.test/demo.json",
			InitialPath: "/published-post/",
			Theme:       "blank",
			Preset:      "minimal",
			DemoMode:    true,
			Embedded:    true,
		},
	}
	payload, err := json.Marshal(blueprint)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	body := string(payload)
	for _, expected := range []string{
		`"blueprint_version":"gocms.playground.blueprint.v1"`,
		`"launch_version":"gocms.playground.launch.v1"`,
		`"snapshot_url":"https://example.test/demo.json"`,
		`"initial_path":"/published-post/"`,
		`"preset":"minimal"`,
		`"embedded":true`,
	} {
		if !strings.Contains(body, expected) {
			t.Fatalf("blueprint payload missing %q: %s", expected, body)
		}
	}
}

func TestSnapshotBridgeAlignsWithCompatibilitySnapshots(t *testing.T) {
	now := time.Date(2026, 5, 9, 12, 0, 0, 0, time.UTC)
	publishedAt := now.Add(-2 * time.Hour)
	bundle := appsnapshot.Bundle{
		Version:    appsnapshot.SnapshotVersion,
		ExportedAt: now,
		Content: []domaincontent.Entry{{
			ID:          "post-1",
			Kind:        domaincontent.KindPost,
			Status:      domaincontent.StatusPublished,
			Visibility:  domaincontent.VisibilityPublic,
			Title:       domaincontent.LocalizedText{"en": "Hello"},
			Slug:        domaincontent.LocalizedText{"en": "hello"},
			Body:        domaincontent.LocalizedText{"en": "<p>World</p>"},
			Excerpt:     domaincontent.LocalizedText{"en": "Summary"},
			AuthorID:    "author-1",
			CreatedAt:   now.Add(-4 * time.Hour),
			UpdatedAt:   now.Add(-3 * time.Hour),
			PublishedAt: &publishedAt,
		}},
		Media: []domainmedia.Asset{{
			ID:        "media-1",
			Filename:  "hero.jpg",
			MimeType:  "image/jpeg",
			SizeBytes: 42,
			Width:     1200,
			Height:    800,
			AltText:   "Hero",
			Caption:   "Welcome",
			PublicURL: "/media/hero.jpg",
			CreatedAt: now.Add(-5 * time.Hour),
			UpdatedAt: now.Add(-5 * time.Hour),
		}},
		Settings: []domainsettings.Value{{
			Key:    "site.title",
			Value:  "GoCMS Playground",
			Public: true,
		}},
	}

	snapshot, err := SnapshotFromBundle(bundle, Source{Kind: "gocms-snapshot", BaseURL: "https://example.test", Imported: now})
	if err != nil {
		t.Fatalf("SnapshotFromBundle() error = %v", err)
	}
	if snapshot.Version != DefaultSnapshotVersion {
		t.Fatalf("snapshot version = %q", snapshot.Version)
	}
	if snapshot.Local.MediaBlobs != BlobStatusExcluded {
		t.Fatalf("media blob policy = %q", snapshot.Local.MediaBlobs)
	}
	if len(snapshot.Settings) != 1 || snapshot.Settings[0].Key != "site.title" {
		t.Fatalf("snapshot settings = %+v", snapshot.Settings)
	}
	var posts []map[string]any
	if err := json.Unmarshal(snapshot.Routes[RoutePosts], &posts); err != nil {
		t.Fatalf("unmarshal post route: %v", err)
	}
	if len(posts) != 1 || posts[0]["slug"] != "hello" {
		t.Fatalf("post route payload = %+v", posts)
	}
	content, _ := posts[0]["content"].(map[string]any)
	if content["rendered"] != "<p>World</p>" {
		t.Fatalf("post content payload = %+v", posts)
	}

	roundTrip, err := BundleFromSnapshot(snapshot, func() time.Time { return now })
	if err != nil {
		t.Fatalf("BundleFromSnapshot() error = %v", err)
	}
	if roundTrip.Version != appsnapshot.SnapshotVersion {
		t.Fatalf("bundle version = %q", roundTrip.Version)
	}
	if len(roundTrip.Content) != 1 || roundTrip.Content[0].Slug.Value("en", "en") != "hello" {
		t.Fatalf("round-trip content = %+v", roundTrip.Content)
	}
	if len(roundTrip.Media) != 1 || roundTrip.Media[0].Filename != "hero.jpg" {
		t.Fatalf("round-trip media = %+v", roundTrip.Media)
	}
	if len(roundTrip.Settings) != 1 || roundTrip.Settings[0].Value != "GoCMS Playground" {
		t.Fatalf("round-trip settings = %+v", roundTrip.Settings)
	}
}

func containsBinaryPayload(value string) bool {
	return strings.Contains(value, "data:image/") || strings.Contains(value, "base64,") || strings.Contains(value, `"blob":`)
}
