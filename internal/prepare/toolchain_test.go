package prepare

import (
	"context"
	"io"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/woliveiras/blind-dev-setup/internal/manifest"
)

func TestMiseToolchainInstallsAndProbesPinnedTools(t *testing.T) {
	runner := &recordingRunner{}
	root := t.TempDir()
	current := manifest.Manifest{Toolchain: map[string]string{
		"python": "3.14.7",
		"node":   "24.19.0",
		"uv":     "0.12.5",
		"pnpm":   "11.22.0",
	}}

	if err := (MiseToolchain{Runner: runner}).Install(context.Background(), root, current, io.Discard); err != nil {
		t.Fatalf("Install() error = %v", err)
	}
	if len(runner.commands) != 7 {
		t.Fatalf("command count = %d, commands = %#v", len(runner.commands), runner.commands)
	}
	wantInstall := []string{
		"install", "--yes", "--locked", "python@3.14.7", "node@24.19.0", "uv@0.12.5", "pnpm@11.22.0",
	}
	wantTrust := []string{"trust", "--yes", filepath.Join(root, "config", "mise.toml")}
	if !slices.Equal(runner.commands[0].arguments, wantTrust) {
		t.Fatalf("trust arguments = %#v", runner.commands[0].arguments)
	}
	if !slices.Equal(runner.commands[1].arguments, wantInstall) {
		t.Fatalf("install arguments = %#v", runner.commands[1].arguments)
	}
	wantProbes := [][]string{
		{"exec", "python@3.14.7", "--", "python", "--version"},
		{"exec", "node@24.19.0", "--", "node", "--version"},
		{"exec", "uv@0.12.5", "--", "uv", "--version"},
		{"exec", "pnpm@11.22.0", "--", "pnpm", "--version"},
	}
	for index, want := range wantProbes {
		if !slices.Equal(runner.commands[index+3].arguments, want) {
			t.Errorf("probe %d = %#v, want %#v", index, runner.commands[index+3].arguments, want)
		}
	}
	dataDir := "MISE_DATA_DIR=" + filepath.Join(root, "data", "mise")
	found := false
	for _, entry := range portableEnvironment(root) {
		if strings.EqualFold(entry, dataDir) {
			found = true
		}
	}
	if !found {
		t.Errorf("portable environment does not contain %q", dataDir)
	}
}
