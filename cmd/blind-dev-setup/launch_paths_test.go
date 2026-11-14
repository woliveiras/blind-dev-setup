package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func launchFixture(t *testing.T, body string) string {
	t.Helper()
	root := filepath.Join(t.TempDir(), "bundle spaces & (literal) %UNCHANGED%")
	for _, relative := range []string{"START.cmd", "READY.txt", "config/environment.cmd", "tools/nvda/nvda.exe", "tools/vscode/Code.exe"} {
		path := filepath.Join(root, filepath.FromSlash(relative))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		content := []byte("fixture")
		if relative == "START.cmd" {
			content = []byte(strings.ReplaceAll(body, "\n", "\r\n"))
		}
		if err := os.WriteFile(path, content, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func TestValidateLaunchRootAcceptsExistingLaunchContract(t *testing.T) {
	root := launchFixture(t, "@echo off\nexit /b 0\n")
	actual, err := validateLaunchRoot(root)
	if err != nil || actual != root {
		t.Fatalf("root=%q error=%v", actual, err)
	}
	// Older prepared environments need not have a new bootstrap helper.
	if _, err := os.Stat(filepath.Join(root, "config", "initialize-toolchain.cmd")); !os.IsNotExist(err) {
		t.Fatal("fixture unexpectedly has new helper")
	}
}

func TestValidateLaunchRootRejectsIncompleteEnvironment(t *testing.T) {
	for _, relative := range []string{"START.cmd", "READY.txt", "config/environment.cmd", "tools/nvda/nvda.exe", "tools/vscode/Code.exe"} {
		t.Run(relative, func(t *testing.T) {
			for _, state := range []string{"missing", "empty", "directory"} {
				t.Run(state, func(t *testing.T) {
					root := launchFixture(t, "@echo off\nexit /b 0\n")
					path := filepath.Join(root, filepath.FromSlash(relative))
					if err := os.Remove(path); err != nil {
						t.Fatal(err)
					}
					switch state {
					case "empty":
						if err := os.WriteFile(path, nil, 0o644); err != nil {
							t.Fatal(err)
						}
					case "directory":
						if err := os.Mkdir(path, 0o755); err != nil {
							t.Fatal(err)
						}
					}
					if _, err := validateLaunchRoot(root); err == nil || !strings.Contains(err.Error(), relative) {
						t.Fatalf("error=%v", err)
					}
				})
			}
		})
	}
}

func TestValidateLaunchRootRejectsInvalidRoot(t *testing.T) {
	root := launchFixture(t, "@echo off\nexit /b 0\n")
	for _, path := range []string{"", " ", filepath.Join(root, "missing"), filepath.Join(root, "READY.txt")} {
		if _, err := validateLaunchRoot(path); err == nil {
			t.Fatalf("accepted root %q", path)
		}
	}
}
