package templates

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/woliveiras/blind-dev-setup/internal/bundle"
)

// The subprocess acts as mise without downloading or executing third-party code.
// This exercises the generated PowerShell control flow and native exit handling.
func init() {
	logPath := os.Getenv("BLIND_DEV_TEST_MISE_LOG")
	if logPath == "" || filepath.Base(os.Args[0]) != "mise.exe" {
		return
	}
	cwd, _ := os.Getwd()
	line, _ := json.Marshal(struct {
		Args []string `json:"args"`
		Cwd  string   `json:"cwd"`
	}{os.Args[1:], cwd})
	file, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		os.Exit(99)
	}
	_, _ = file.Write(append(line, '\n'))
	_ = file.Close()
	if len(os.Args) > 1 && os.Args[1] == os.Getenv("BLIND_DEV_TEST_MISE_FAIL") {
		os.Exit(17)
	}
	os.Exit(0)
}

func TestInitializerValidatesBeforeTrustAndWorksAfterRelocation(t *testing.T) {
	shell := "pwsh"
	if runtime.GOOS == "windows" {
		shell = "powershell.exe"
	}
	powershell, err := exec.LookPath(shell)
	if err != nil {
		t.Skip("PowerShell runtime unavailable; Windows CI runs this integration test")
	}
	current, err := bundle.WindowsManifest()
	if err != nil {
		t.Fatal(err)
	}
	testBinary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	binary, err := os.ReadFile(testBinary)
	if err != nil {
		t.Fatal(err)
	}
	for _, scenario := range []struct {
		name            string
		alter           string
		missing         bool
		fail            string
		isolatedModules bool
		wantCalls       int
		wantError       bool
	}{
		{name: "relocated", wantCalls: 2},
		{name: "isolated module path", isolatedModules: true, wantCalls: 2},
		{name: "changed config", alter: "mise.toml", wantError: true},
		{name: "changed lock", alter: "mise.lock", wantError: true},
		{name: "missing config", alter: "mise.toml", missing: true, wantError: true},
		{name: "trust failure", fail: "trust", wantCalls: 1, wantError: true},
		{name: "reshim failure", fail: "reshim", wantCalls: 2, wantError: true},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			parent := t.TempDir()
			staging := filepath.Join(parent, "staging")
			if err := (Writer{Manifest: current}).Write(staging, bundle.WindowsManifestJSON()); err != nil {
				t.Fatal(err)
			}
			root := filepath.Join(parent, "published path & spaces")
			if err := os.Rename(staging, root); err != nil {
				t.Fatal(err)
			}
			mise := filepath.Join(root, "tools", "mise", "mise.exe")
			if err := os.MkdirAll(filepath.Dir(mise), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(mise, binary, 0o755); err != nil {
				t.Fatal(err)
			}
			if scenario.alter != "" {
				path := filepath.Join(root, "config", scenario.alter)
				if scenario.missing {
					err = os.Remove(path)
				} else {
					err = os.WriteFile(path, []byte("changed"), 0o644)
				}
				if err != nil {
					t.Fatal(err)
				}
			}
			logPath := filepath.Join(parent, "mise-calls.jsonl")
			command := exec.Command(powershell, "-NoLogo", "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-File", filepath.Join(root, "config", "initialize-toolchain.ps1"))
			command.Env = append(os.Environ(), "BLIND_DEV_TEST_MISE_LOG="+logPath, "BLIND_DEV_TEST_MISE_FAIL="+scenario.fail)
			if scenario.isolatedModules {
				modulePath := filepath.Join(parent, "empty-modules")
				if err := os.Mkdir(modulePath, 0o755); err != nil {
					t.Fatal(err)
				}
				filtered := command.Env[:0]
				for _, entry := range command.Env {
					key, _, _ := strings.Cut(entry, "=")
					if !strings.EqualFold(key, "PSModulePath") {
						filtered = append(filtered, entry)
					}
				}
				command.Env = append(filtered, "PSModulePath="+modulePath)
			}
			output, runErr := command.CombinedOutput()
			if (runErr != nil) != scenario.wantError {
				t.Fatalf("error = %v, want error %v; output: %s", runErr, scenario.wantError, output)
			}
			log, _ := os.ReadFile(logPath)
			lines := strings.Split(strings.TrimSpace(string(log)), "\n")
			if len(log) == 0 {
				lines = nil
			}
			if len(lines) != scenario.wantCalls {
				t.Fatalf("mise calls = %d, want %d; log: %s; output: %s", len(lines), scenario.wantCalls, log, output)
			}
			for index, line := range lines {
				var call struct {
					Args []string `json:"args"`
					Cwd  string   `json:"cwd"`
				}
				if err := json.Unmarshal([]byte(line), &call); err != nil {
					t.Fatal(err)
				}
				if index == 0 && (len(call.Args) != 3 || call.Args[0] != "trust" || call.Args[1] != "--yes" || call.Args[2] != filepath.Join(root, "config", "mise.toml")) {
					t.Errorf("trust arguments = %v", call.Args)
				}
				if index == 1 && strings.Join(call.Args, " ") != "reshim --force" {
					t.Errorf("reshim arguments = %v", call.Args)
				}
				actualDir, actualErr := os.Stat(call.Cwd)
				wantedDir, wantedErr := os.Stat(root)
				if actualErr != nil || wantedErr != nil || !os.SameFile(actualDir, wantedDir) {
					t.Errorf("working directory = %q, want %q", call.Cwd, root)
				}
			}
		})
	}
}
