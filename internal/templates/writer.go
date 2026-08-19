package templates

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/woliveiras/blind-dev-setup/internal/manifest"
)

//go:embed files
var files embed.FS

type Writer struct {
	Manifest manifest.Manifest
}

func (w Writer) Write(root string, manifestJSON []byte) error {
	if err := fs.WalkDir(files, "files", func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relative := strings.TrimPrefix(path, "files/")
		if relative == "files" || relative == "" {
			return nil
		}
		destination := filepath.Join(root, filepath.FromSlash(relative))
		if entry.IsDir() {
			return os.MkdirAll(destination, 0o755)
		}
		content, err := files.ReadFile(path)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(destination, content, 0o644); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}

	for _, relative := range []string{
		"cache/mise",
		"cache/uv",
		"data/dbeaver/workspace",
		"data/mise",
		"data/mise-state",
		"data/pnpm/home",
		"data/pnpm/store",
		"tools/vscode/data/extensions",
		"tools/vscode/data/tmp",
		"workspace/projects",
	} {
		if err := os.MkdirAll(filepath.Join(root, filepath.FromSlash(relative)), 0o755); err != nil {
			return err
		}
	}

	if err := os.WriteFile(filepath.Join(root, "config", "manifest.json"), manifestJSON, 0o644); err != nil {
		return err
	}
	miseConfig := fmt.Sprintf(
		"[tools]\npython = %q\nnode = %q\nuv = %q\npnpm = %q\n",
		w.Manifest.Toolchain["python"],
		w.Manifest.Toolchain["node"],
		w.Manifest.Toolchain["uv"],
		w.Manifest.Toolchain["pnpm"],
	)
	return os.WriteFile(filepath.Join(root, "config", "mise.toml"), []byte(miseConfig), 0o644)
}
