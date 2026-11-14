package verify

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/woliveiras/blind-dev-setup/internal/manifest"
	"github.com/woliveiras/blind-dev-setup/internal/target"
	"github.com/woliveiras/blind-dev-setup/internal/templates"
)

func TestRunChecksManifestComponentsLaunchersAndToolchain(t *testing.T) {
	targetRoot, root, current := verificationFixture(t)
	// User preferences are intentionally editable, unlike the managed bootstrap.
	settings := filepath.Join(root, "tools", "vscode", "data", "user-data", "User", "settings.json")
	customSettings := []byte(`{"editor.fontSize": 22}`)
	if err := os.WriteFile(settings, customSettings, 0o644); err != nil {
		t.Fatal(err)
	}
	var output strings.Builder
	if err := Run(targetRoot, current, &output); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	for _, message := range []string{"Verificação concluída", "programas não foram executados", "não tiveram seus hashes recalculados"} {
		if !strings.Contains(output.String(), message) {
			t.Errorf("output does not explain %q: %s", message, output.String())
		}
	}
	content, err := os.ReadFile(settings)
	if err != nil || string(content) != string(customSettings) {
		t.Fatalf("verify modified user settings: %q, %v", content, err)
	}
}

func TestRunReportsMissingExpectedFile(t *testing.T) {
	targetRoot, root, current := verificationFixture(t)
	if err := os.Remove(filepath.Join(root, "START.cmd")); err != nil {
		t.Fatal(err)
	}
	err := Run(targetRoot, current, &strings.Builder{})
	if err == nil || !strings.Contains(err.Error(), "START.cmd") {
		t.Fatalf("Run() error = %v", err)
	}
}

func TestRunRejectsInstalledManifestThatDiffersFromBuilder(t *testing.T) {
	targetRoot, root, expected := verificationFixture(t)
	installed := verificationManifest()
	installed.Release = "0.1.0-tampered"
	raw, err := json.Marshal(installed)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "config", "manifest.json"), raw, 0o644); err != nil {
		t.Fatal(err)
	}
	err = Run(targetRoot, expected, &strings.Builder{})
	if err == nil || !strings.Contains(err.Error(), "difere") {
		t.Fatalf("Run() error = %v", err)
	}
}

func TestRunRejectsMissingEmptyOrNonFileToolchainExecutable(t *testing.T) {
	for _, tool := range []string{"python", "node", "uv", "pnpm"} {
		for _, damage := range []string{"missing", "empty", "directory"} {
			t.Run(tool+"/"+damage, func(t *testing.T) {
				targetRoot, root, current := verificationFixture(t)
				path := filepath.Join(root, "data", "mise", "installs", tool, current.Toolchain[tool], tool+".exe")
				if err := os.Remove(path); err != nil {
					t.Fatal(err)
				}
				switch damage {
				case "empty":
					if err := os.WriteFile(path, nil, 0o644); err != nil {
						t.Fatal(err)
					}
				case "directory":
					if err := os.Mkdir(path, 0o755); err != nil {
						t.Fatal(err)
					}
				}
				var output strings.Builder
				err := Run(targetRoot, current, &output)
				if err == nil || !strings.Contains(err.Error(), tool+".exe") {
					t.Fatalf("Run() error = %v; expected broken executable path", err)
				}
				if strings.Contains(output.String(), "Verificação concluída") {
					t.Fatal("broken toolchain reported as verified")
				}
			})
		}
	}
}

func TestRunRejectsEmptyComponentExecutable(t *testing.T) {
	targetRoot, root, current := verificationFixture(t)
	if err := os.WriteFile(filepath.Join(root, "tools", "editor", "editor.exe"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	err := Run(targetRoot, current, &strings.Builder{})
	if err == nil || !strings.Contains(err.Error(), "vazio") || !strings.Contains(err.Error(), "editor.exe") {
		t.Fatalf("Run() error = %v", err)
	}
}

func TestRunRejectsModifiedManagedConfigurationWithoutRepairingIt(t *testing.T) {
	for _, relative := range []string{"config/mise.toml", "config/mise.lock", "config/environment.cmd"} {
		t.Run(relative, func(t *testing.T) {
			targetRoot, root, current := verificationFixture(t)
			path := filepath.Join(root, filepath.FromSlash(relative))
			modified := []byte("unexpected configuration\n")
			if err := os.WriteFile(path, modified, 0o644); err != nil {
				t.Fatal(err)
			}
			err := Run(targetRoot, current, &strings.Builder{})
			if err == nil || !strings.Contains(err.Error(), relative) || !strings.Contains(err.Error(), "difere") {
				t.Fatalf("Run() error = %v", err)
			}
			content, err := os.ReadFile(path)
			if err != nil || string(content) != string(modified) {
				t.Fatalf("verify modified the managed configuration: %q, %v", content, err)
			}
		})
	}
}

func verificationFixture(t *testing.T) (string, string, manifest.Manifest) {
	t.Helper()
	targetRoot := t.TempDir()
	root := filepath.Join(targetRoot, target.DirectoryName)
	current := verificationManifest()
	raw, err := json.Marshal(current)
	if err != nil {
		t.Fatal(err)
	}
	if err := (templates.Writer{Manifest: current}).Write(root, raw); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(root, "READY.txt"))
	writeFile(t, filepath.Join(root, "tools", "editor", "editor.exe"))
	for tool, version := range current.Toolchain {
		writeFile(t, filepath.Join(root, "data", "mise", "installs", tool, version, tool+".exe"))
	}
	return targetRoot, root, current
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
