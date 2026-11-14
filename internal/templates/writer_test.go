package templates

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/woliveiras/blind-dev-setup/internal/bundle"
)

func TestWriterCreatesPortableLaunchersAndPinnedToolchain(t *testing.T) {
	root := t.TempDir()
	current, err := bundle.WindowsManifest()
	if err != nil {
		t.Fatal(err)
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
		"config/initialize-toolchain.cmd",
		"config/initialize-toolchain.ps1",
		"config/manifest.json",
		"config/mise.toml",
		"config/mise.lock",
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
	for tool, version := range current.Toolchain {
		want := tool + " = \"" + version + "\""
		if !strings.Contains(string(miseConfig), want) {
			t.Errorf("mise.toml does not contain %q:\n%s", want, miseConfig)
		}
	}
	miseLock, err := os.ReadFile(filepath.Join(root, "config", "mise.lock"))
	if err != nil {
		t.Fatal(err)
	}
	for _, version := range current.Toolchain {
		want := "version = \"" + version + "\""
		if !strings.Contains(string(miseLock), want) {
			t.Errorf("mise.lock does not contain %q", want)
		}
	}
	for _, want := range []string{"platforms.windows-x64", "checksum = \"sha256:"} {
		if !strings.Contains(string(miseLock), want) {
			t.Errorf("mise.lock does not contain %q", want)
		}
	}
}

func TestDevelopmentLaunchersShareToolchainInitialization(t *testing.T) {
	for _, launcher := range []string{"START.cmd", "START-VSCODE.cmd", "DEV-SHELL.cmd"} {
		content, err := files.ReadFile("files/" + launcher)
		if err != nil {
			t.Fatal(err)
		}
		text := strings.ReplaceAll(string(content), "\r\n", "\n")
		if !strings.Contains(text, "call \"%~dp0config\\initialize-toolchain.cmd\"\nif errorlevel 1 goto error") {
			t.Errorf("%s must initialize the toolchain and stop on failure", launcher)
		}
		if strings.Contains(string(content), "trust --yes") {
			t.Errorf("%s must not trust configuration outside the validated initializer", launcher)
		}
	}
}
