package cms

import (
	"testing"

	"github.com/fastygo/cms/internal/platform/runtimeprofile"
)

func TestNewWithOptionsUsesBrowserLocalProfileWithoutDurableDataSource(t *testing.T) {
	module, err := NewWithOptions(Options{
		DataSource:     "file:/path/that/should/not/be/created/gocms.db",
		SessionKey:     "test-session-key",
		SeedFixtures:   true,
		RuntimeProfile: string(runtimeprofile.RuntimeProfilePlayground),
		StorageProfile: string(runtimeprofile.StorageProfileBrowserIndexedDB),
	})
	if err != nil {
		t.Fatalf("NewWithOptions() error = %v", err)
	}
	t.Cleanup(func() {
		_ = module.Close(t.Context())
	})
}
