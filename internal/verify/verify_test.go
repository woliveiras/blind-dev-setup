package verify

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/woliveiras/blind-dev-setup/internal/manifest"
	"github.com/woliveiras/blind-dev-setup/internal/target"
)

func TestRunChecksManifestComponentsLaunchersAndToolchain(t *testing.T) {
	targetRoot := t.TempDir()
	root := filepath.Join(targetRoot, target.DirectoryName)
	current := verificationManifest()
	raw, err := json.Marshal(current)
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(root, "READY.txt"))
	if err := os.MkdirAll(filepath.Join(root, "config"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "config", "manifest.json"), raw, 0o644); err != nil {
		t.Fatal(err)
	}
	for _, required := range requiredPortableFiles {
		writeFile(t, filepath.Join(root, filepath.FromSlash(required)))
	}
	writeFile(t, filepath.Join(root, "tools", "editor", "editor.exe"))
	for tool, version := range current.Toolchain {
		if err := os.MkdirAll(filepath.Join(root, "data", "mise", "installs", tool, version), 0o755); err != nil {
			t.Fatal(err)
		}
	}

	var output strings.Builder
	if err := Run(targetRoot, current, &output); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if !strings.Contains(output.String(), "Verificação concluída") {
		t.Fatalf("output = %q", output.String())
	}
}

func TestRunReportsMissingExpectedFile(t *testing.T) {
	targetRoot := t.TempDir()
	root := filepath.Join(targetRoot, target.DirectoryName)
	current := verificationManifest()
	raw, err := json.Marshal(current)
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(root, "READY.txt"))
	if err := os.MkdirAll(filepath.Join(root, "config"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "config", "manifest.json"), raw, 0o644); err != nil {
		t.Fatal(err)
	}

	err = Run(targetRoot, current, &strings.Builder{})
	if err == nil || !strings.Contains(err.Error(), "START.cmd") {
		t.Fatalf("Run() error = %v", err)
	}
}

func TestRunRejectsInstalledManifestThatDiffersFromBuilder(t *testing.T) {
	targetRoot := t.TempDir()
	root := filepath.Join(targetRoot, target.DirectoryName)
	expected := verificationManifest()
	installed := verificationManifest()
	installed.Release = "0.1.0-tampered"
	raw, err := json.Marshal(installed)
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(root, "READY.txt"))
	if err := os.MkdirAll(filepath.Join(root, "config"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "config", "manifest.json"), raw, 0o644); err != nil {
		t.Fatal(err)
	}
	for _, required := range requiredPortableFiles {
		writeFile(t, filepath.Join(root, filepath.FromSlash(required)))
	}

	err = Run(targetRoot, expected, &strings.Builder{})
	if err == nil || !strings.Contains(err.Error(), "difere") {
		t.Fatalf("Run() error = %v", err)
	}
}

func verificationManifest() manifest.Manifest {
	return manifest.Manifest{
		SchemaVersion:    1,
		Release:          "0.1.0",
		Platform:         "windows-x64",
		MinimumFreeBytes: 1,
		Toolchain: map[string]string{
			"python": "3.14.7",
			"node":   "24.19.0",
			"uv":     "0.12.5",
			"pnpm":   "11.22.0",
		},
		Components: []manifest.Component{{
			ID:            "editor",
			Name:          "Editor",
			Version:       "1.2.3",
			Required:      true,
			URL:           "https://example.test/editor.zip",
			SHA256:        strings.Repeat("a", 64),
			License:       "Example",
			LicenseURL:    "https://example.test/license",
			Kind:          "zip",
			Destination:   "tools/editor",
			DownloadBytes: 100,
			ExpectedFiles: []string{"editor.exe"},
		}},
	}
}

func writeFile(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("present"), 0o644); err != nil {
		t.Fatal(err)
	}
}
