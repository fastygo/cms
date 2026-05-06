package playground

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
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

func containsBinaryPayload(value string) bool {
	return strings.Contains(value, "data:image/") || strings.Contains(value, "base64,") || strings.Contains(value, `"blob":`)
}
