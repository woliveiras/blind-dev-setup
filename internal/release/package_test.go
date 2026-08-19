package release_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestWindowsBuildCreatesCompletePortableRelease(t *testing.T) {
	root := projectRoot(t)
	script := readFile(t, filepath.Join(root, "scripts", "build.ps1"))
	for _, required := range []string{
		"go test ./...",
		"go vet ./...",
		"blind-dev-setup-windows-x64.exe",
		"blind-dev-setup-windows-x64.zip",
		"SHA256SUMS.txt",
		"INICIAR-AQUI.cmd",
		"LEIA-ME.txt",
		"Compress-Archive",
		"Get-FileHash",
	} {
		if !strings.Contains(script, required) {
			t.Errorf("scripts/build.ps1 does not contain %q", required)
		}
	}

	starter := readFile(t, filepath.Join(root, "packaging", "windows", "INICIAR-AQUI.cmd"))
	if !strings.Contains(starter, "blind-dev-setup-windows-x64.exe\" list-targets") {
		t.Error("INICIAR-AQUI.cmd does not start accessible target discovery")
	}
	if !strings.Contains(starter, "pause") {
		t.Error("INICIAR-AQUI.cmd closes before the user can read the result")
	}

	guide := readFile(t, filepath.Join(root, "packaging", "windows", "LEIA-ME.txt"))
	for _, required := range []string{"INICIAR-AQUI.cmd", "list-targets", "nao formata"} {
		if !strings.Contains(guide, required) {
			t.Errorf("packaged guide does not contain %q", required)
		}
	}
}

func projectRoot(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller() failed")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(filename), "..", ".."))
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(contents)
}
