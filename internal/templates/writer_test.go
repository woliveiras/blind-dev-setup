package templates

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/woliveiras/blind-dev-setup/internal/manifest"
)

func TestWriterCreatesPortableLaunchersAndPinnedToolchain(t *testing.T) {
	root := t.TempDir()
	current := manifest.Manifest{
		Toolchain: map[string]string{
			"python": "3.14.7",
			"node":   "24.19.0",
			"uv":     "0.12.5",
			"pnpm":   "11.22.0",
		},
	}
	writer := Writer{Manifest: current}

	if err := writer.Write(root, []byte("{\"release\":\"0.1.0\"}\n")); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	for _, relative := range []string{
		"START.cmd",
		"DEV-SHELL.cmd",
		"START-NVDA.cmd",
		"START-VSCODE.cmd",
		"START-NOTEPADPP.cmd",
		"START-DBEAVER.cmd",
		"config/environment.cmd",
		"config/manifest.json",
		"config/mise.toml",
		"tools/vscode/data/user-data/User/settings.json",
		"workspace/python-example/pyproject.toml",
		"workspace/node-workspace/pnpm-workspace.yaml",
		"docs/VALIDACAO.md",
	} {
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(relative))); err != nil {
			t.Errorf("%s: %v", relative, err)
		}
	}
	miseConfig, err := os.ReadFile(filepath.Join(root, "config", "mise.toml"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"python = \"3.14.7\"",
		"node = \"24.19.0\"",
		"uv = \"0.12.5\"",
		"pnpm = \"11.22.0\"",
	} {
		if !strings.Contains(string(miseConfig), want) {
			t.Errorf("mise.toml does not contain %q:\n%s", want, miseConfig)
		}
	}
}
