package bundle

import (
	"bytes"
	_ "embed"

	"github.com/woliveiras/blind-dev-setup/internal/manifest"
)

//go:embed windows-x64.json
var windowsManifestJSON []byte

func WindowsManifest() (manifest.Manifest, error) {
	return manifest.Parse(bytes.NewReader(windowsManifestJSON))
}

func WindowsManifestJSON() []byte {
	result := make([]byte, len(windowsManifestJSON))
	copy(result, windowsManifestJSON)
	return result
}
