package prepare

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/woliveiras/blind-dev-setup/internal/manifest"
)

type MiseToolchain struct {
	Runner CommandRunner
}

func (m MiseToolchain) Install(
	ctx context.Context,
	root string,
	current manifest.Manifest,
	output io.Writer,
) error {
	if m.Runner == nil {
		return errors.New("executor do mise não configurado")
	}
	miseExecutable := filepath.Join(root, "tools", "mise", "mise.exe")
	environment := portableEnvironment(root)
	configFile := filepath.Join(root, "config", "mise.toml")
	if err := m.Runner.Run(
		ctx,
		miseExecutable,
		[]string{"trust", "--yes", configFile},
		root,
		environment,
		output,
	); err != nil {
		return fmt.Errorf("confiar na configuração incorporada: %w", err)
	}
	arguments := []string{
		"install",
		"--yes",
		"--locked",
		"python@" + current.Toolchain["python"],
		"node@" + current.Toolchain["node"],
		"uv@" + current.Toolchain["uv"],
		"pnpm@" + current.Toolchain["pnpm"],
	}
	if err := m.Runner.Run(ctx, miseExecutable, arguments, root, environment, output); err != nil {
		return err
	}
	if err := m.Runner.Run(ctx, miseExecutable, []string{"reshim"}, root, environment, output); err != nil {
		return err
	}
	probes := []struct {
		tool    string
		command string
	}{
		{tool: "python", command: "python"},
		{tool: "node", command: "node"},
		{tool: "uv", command: "uv"},
		{tool: "pnpm", command: "pnpm"},
	}
	for _, probe := range probes {
		arguments := []string{
			"exec",
			probe.tool + "@" + current.Toolchain[probe.tool],
			"--",
			probe.command,
			"--version",
		}
		if err := m.Runner.Run(ctx, miseExecutable, arguments, root, environment, output); err != nil {
			return fmt.Errorf("validar %s: %w", probe.tool, err)
		}
	}
	return nil
}

func portableEnvironment(root string) []string {
	overrides := map[string]string{
		"MISE_CONFIG_FILE": filepath.Join(root, "config", "mise.toml"),
		"MISE_DATA_DIR":    filepath.Join(root, "data", "mise"),
		"MISE_CACHE_DIR":   filepath.Join(root, "cache", "mise"),
		"MISE_STATE_DIR":   filepath.Join(root, "data", "mise-state"),
		"MISE_CONFIG_DIR":  filepath.Join(root, "config", "mise"),
		"UV_CACHE_DIR":     filepath.Join(root, "cache", "uv"),
		"PNPM_HOME":        filepath.Join(root, "data", "pnpm", "home"),
		"PNPM_STORE_DIR":   filepath.Join(root, "data", "pnpm", "store"),
	}
	result := make([]string, 0, len(os.Environ())+len(overrides))
	for _, entry := range os.Environ() {
		key, _, found := strings.Cut(entry, "=")
		if !found {
			continue
		}
		overridden := false
		for override := range overrides {
			if strings.EqualFold(key, override) {
				overridden = true
				break
			}
		}
		if !overridden {
			result = append(result, entry)
		}
	}
	for key, value := range overrides {
		result = append(result, fmt.Sprintf("%s=%s", key, value))
	}
	return result
}
