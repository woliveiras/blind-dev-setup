package prepare

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/woliveiras/blind-dev-setup/internal/manifest"
	"github.com/woliveiras/blind-dev-setup/internal/target"
)

type fakeFetcher struct {
	path string
	err  error
}

func (f fakeFetcher) Fetch(context.Context, manifest.Component, string) (string, error) {
	return f.path, f.err
}

type fakeInstaller struct {
	err error
}

func (f fakeInstaller) Install(_ context.Context, component manifest.Component, _, root string, _ io.Writer) error {
	if f.err != nil {
		return f.err
	}
	for _, expected := range component.ExpectedFiles {
		path := filepath.Join(root, component.Destination, expected)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(path, []byte("installed"), 0o755); err != nil {
			return err
		}
	}
	return nil
}

type fakeMaterializer struct{}

func (fakeMaterializer) Write(root string, manifestJSON []byte) error {
	if err := os.MkdirAll(filepath.Join(root, "config"), 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(root, "config", "manifest.json"), manifestJSON, 0o644)
}

type fakeToolchain struct {
	err error
}

func (f fakeToolchain) Install(context.Context, string, manifest.Manifest, io.Writer) error {
	return f.err
}

func TestCreatorPublishesOnlyCompletePreparation(t *testing.T) {
	targetRoot := t.TempDir()
	current := preparationManifest()
	var output strings.Builder
	creator := Creator{
		Manifest:     current,
		ManifestJSON: []byte("{\"release\":\"0.1.0\"}"),
		Fetch:        fakeFetcher{path: "artifact.zip"},
		Install:      fakeInstaller{},
		Materialize:  fakeMaterializer{},
		Toolchain:    fakeToolchain{},
		Inspect: func(string) (target.Details, error) {
			return target.Details{
				Path:         targetRoot,
				Volume:       "E:",
				SystemVolume: "C:",
				FileSystem:   "NTFS",
				FreeBytes:    20 << 30,
			}, nil
		},
	}

	if err := creator.Run(context.Background(), targetRoot, t.TempDir(), &output); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	final := filepath.Join(targetRoot, target.DirectoryName)
	if _, err := os.Stat(filepath.Join(final, "tools", "editor", "editor.exe")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(final, "READY.txt")); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "Preparação concluída") {
		t.Fatalf("output = %q", output.String())
	}
}

func TestCreatorCleansStagingAfterFailure(t *testing.T) {
	targetRoot := t.TempDir()
	creator := Creator{
		Manifest:     preparationManifest(),
		ManifestJSON: []byte("{}"),
		Fetch:        fakeFetcher{err: errors.New("network unavailable")},
		Install:      fakeInstaller{},
		Materialize:  fakeMaterializer{},
		Toolchain:    fakeToolchain{},
		Inspect: func(string) (target.Details, error) {
			return target.Details{
				Path:         targetRoot,
				Volume:       "E:",
				SystemVolume: "C:",
				FileSystem:   "NTFS",
				FreeBytes:    20 << 30,
			}, nil
		},
	}

	err := creator.Run(context.Background(), targetRoot, t.TempDir(), io.Discard)
	if err == nil || !strings.Contains(err.Error(), "network unavailable") {
		t.Fatalf("Run() error = %v", err)
	}
	entries, readErr := os.ReadDir(targetRoot)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if len(entries) != 0 {
		t.Fatalf("target contains incomplete preparation: %v", entries)
	}
}

func preparationManifest() manifest.Manifest {
	return manifest.Manifest{
		SchemaVersion:    1,
		Release:          "0.1.0",
		Platform:         "windows-x64",
		MinimumFreeBytes: 12 << 30,
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
			Destination:   "tools/editor",
			ExpectedFiles: []string{"editor.exe"},
		}},
	}
}
