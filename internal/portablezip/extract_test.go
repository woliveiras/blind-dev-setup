package portablezip

import (
	"archive/zip"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExtractStripsRootDirectory(t *testing.T) {
	archive := makeZip(t, map[string]string{
		"dbeaver/dbeaver.exe":        "executable",
		"dbeaver/plugins/readme.txt": "plugin",
	})
	destination := filepath.Join(t.TempDir(), "dbeaver")

	if err := Extract(archive, destination, 1); err != nil {
		t.Fatalf("Extract() error = %v", err)
	}
	got, err := os.ReadFile(filepath.Join(destination, "dbeaver.exe"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "executable" {
		t.Fatalf("content = %q", got)
	}
}

func TestExtractRejectsPathTraversal(t *testing.T) {
	archive := makeZip(t, map[string]string{"../outside.txt": "unsafe"})
	destination := filepath.Join(t.TempDir(), "tool")

	err := Extract(archive, destination, 0)
	if err == nil || !strings.Contains(err.Error(), "unsafe ZIP path") {
		t.Fatalf("Extract() error = %v", err)
	}
}

func makeZip(t *testing.T, files map[string]string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "archive.zip")
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	writer := zip.NewWriter(file)
	for name, content := range files {
		entry, createErr := writer.Create(name)
		if createErr != nil {
			t.Fatal(createErr)
		}
		if _, writeErr := entry.Write([]byte(content)); writeErr != nil {
			t.Fatal(writeErr)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	return path
}
