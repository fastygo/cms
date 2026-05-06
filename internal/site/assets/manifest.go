package assets

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const manifestRelativePath = "web/static/asset-manifest.json"

type Paths struct {
	CSS          string
	ThemeJS      string
	AppJS        string
	PlaygroundJS string
}

func Resolve() Paths {
	return Paths{
		CSS:          ResolvePath("/static/css/app.css"),
		ThemeJS:      ResolvePath("/static/js/theme.js"),
		AppJS:        ResolvePath("/static/js/ui8kit.js"),
		PlaygroundJS: ResolvePath("/static/js/playground.js"),
	}
}

func ResolvePath(publicPath string) string {
	manifest := loadManifest()
	if resolved := manifest[publicPath]; resolved != "" {
		return resolved
	}
	return publicPath
}

func loadManifest() map[string]string {
	path := findManifest()
	if path == "" {
		return nil
	}
	payload, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var manifest map[string]string
	if err := json.Unmarshal(payload, &manifest); err != nil {
		return nil
	}
	return manifest
}

func findManifest() string {
	workingDir, err := os.Getwd()
	if err != nil {
		return ""
	}
	for dir := workingDir; ; dir = filepath.Dir(dir) {
		candidate := filepath.Join(dir, manifestRelativePath)
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return candidate
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
	}
}
