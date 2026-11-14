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
	starter = strings.ReplaceAll(starter, "\r\n", "\n")
	if !strings.Contains(starter, "\"%~dp0blind-dev-setup-windows-x64.exe\"\n") {
		t.Error("INICIAR-AQUI.cmd must start the wizard without command arguments")
	}
	if strings.Contains(starter, "list-targets") {
		t.Error("INICIAR-AQUI.cmd must not stop after target discovery")
	}
	normalExit := strings.Index(starter, "if \"%exitCode%\"==\"0\" exit /b 0")
	errorPause := strings.Index(starter, "pause")
	if normalExit < 0 || errorPause < 0 || normalExit >= errorPause {
		t.Error("starter must pause on startup failure but exit normally without a second pause")
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
