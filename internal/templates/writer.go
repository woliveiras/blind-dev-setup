package templates

import (
	"crypto/sha256"
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
	managed, err := ManagedFiles(w.Manifest)
	if err != nil {
		return err
	}
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
	for relative, content := range managed {
		if err := os.WriteFile(filepath.Join(root, filepath.FromSlash(relative)), content, 0o644); err != nil {
			return err
		}
	}
	return nil
}

// ManagedFiles returns configuration controlled by this generator. User settings
// and project files are deliberately excluded so verify permits customization.
func ManagedFiles(current manifest.Manifest) (map[string][]byte, error) {
	miseConfig := fmt.Sprintf(
		"[tools]\npython = %q\nnode = %q\nuv = %q\npnpm = %q\n",
		current.Toolchain["python"],
		current.Toolchain["node"],
		current.Toolchain["uv"],
		current.Toolchain["pnpm"],
	)
	managed := map[string][]byte{"config/mise.toml": []byte(miseConfig)}
	for _, relative := range []string{
		"config/mise.lock",
		"config/environment.cmd",
		"config/initialize-toolchain.cmd",
		"config/initialize-toolchain.ps1",
	} {
		content, err := files.ReadFile("files/" + relative)
		if err != nil {
			return nil, err
		}
		managed[relative] = content
	}
	initializer := string(managed["config/initialize-toolchain.ps1"])
	initializer = strings.ReplaceAll(initializer, "__MISE_CONFIG_SHA256__", fmt.Sprintf("%x", sha256.Sum256([]byte(miseConfig))))
	initializer = strings.ReplaceAll(initializer, "__MISE_LOCK_SHA256__", fmt.Sprintf("%x", sha256.Sum256(managed["config/mise.lock"])))
	managed["config/initialize-toolchain.ps1"] = []byte(initializer)
	return managed, nil
}
