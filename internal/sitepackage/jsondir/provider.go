package jsondir

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/fastygo/cms/internal/application/snapshot"
)

const DefaultFilename = "gocms-site-package.json"

type Provider struct {
	Dir string
}

func (p Provider) Enabled() bool {
	return p.Dir != ""
}

func (p Provider) SnapshotPath() string {
	return filepath.Join(p.Dir, DefaultFilename)
}

func (p Provider) Load() (snapshot.Bundle, error) {
	payload, err := os.ReadFile(p.SnapshotPath())
	if err != nil {
		return snapshot.Bundle{}, err
	}
	var bundle snapshot.Bundle
	if err := json.Unmarshal(payload, &bundle); err != nil {
		return snapshot.Bundle{}, err
	}
	return bundle, nil
}

func (p Provider) Save(bundle snapshot.Bundle) error {
	if err := os.MkdirAll(p.Dir, 0o755); err != nil {
		return err
	}
	payload, err := json.MarshalIndent(bundle, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p.SnapshotPath(), append(payload, '\n'), 0o644)
}
