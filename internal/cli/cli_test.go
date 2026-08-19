package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/woliveiras/blind-dev-setup/internal/manifest"
)

func TestPlanPrintsAllActionsWithoutWriting(t *testing.T) {
	var stdout, stderr bytes.Buffer
	exitCode := Run(
		[]string{"plan", "--target", "E:\\"},
		&stdout,
		&stderr,
		Dependencies{Manifest: cliManifest()},
	)

	if exitCode != 0 {
		t.Fatalf("Run() exit = %d, stderr = %q", exitCode, stderr.String())
	}
	for _, want := range []string{
		"Destino: E:\\",
		"Editor 1.2.3",
		"Python 3.14.7",
		"nenhum arquivo será modificado",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Errorf("stdout does not contain %q:\n%s", want, stdout.String())
		}
	}
}

func TestPrepareRequiresLicenseAcceptance(t *testing.T) {
	var stdout, stderr bytes.Buffer
	exitCode := Run(
		[]string{"prepare", "--target", "E:\\"},
		&stdout,
		&stderr,
		Dependencies{Manifest: cliManifest()},
	)

	if exitCode != 2 || !strings.Contains(stderr.String(), "--accept-licenses") {
		t.Fatalf("Run() exit = %d, stderr = %q", exitCode, stderr.String())
	}
}

func TestUnknownCommandReturnsUsageError(t *testing.T) {
	var stdout, stderr bytes.Buffer
	exitCode := Run([]string{"destroy"}, &stdout, &stderr, Dependencies{Manifest: cliManifest()})
	if exitCode != 2 || !strings.Contains(stderr.String(), "Comando desconhecido") {
		t.Fatalf("Run() exit = %d, stderr = %q", exitCode, stderr.String())
	}
}

func cliManifest() manifest.Manifest {
	return manifest.Manifest{
		SchemaVersion:    1,
		Release:          "0.1.0",
		Platform:         "windows-x64",
		MinimumFreeBytes: 1024,
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
			SHA256:        "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			License:       "Example",
			LicenseURL:    "https://example.test/license",
			Kind:          "zip",
			Destination:   "tools/editor",
			DownloadBytes: 100,
			ExpectedFiles: []string{"editor.exe"},
		}},
	}
}
