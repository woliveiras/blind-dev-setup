package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func validateLaunchRoot(root string) (string, error) {
	if strings.TrimSpace(root) == "" {
		return "", fmt.Errorf("a pasta do ambiente é obrigatória")
	}
	absolute, err := filepath.Abs(root)
	if err != nil {
		return "", fmt.Errorf("resolver pasta do ambiente: %w", err)
	}
	info, err := os.Stat(absolute)
	if err != nil {
		return "", fmt.Errorf("abrir pasta do ambiente: %w", err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("o ambiente não é uma pasta: %s", absolute)
	}
	// Validate the established launch contract without requiring the exact
	// current template or newly introduced files in already prepared bundles.
	for _, relative := range []string{
		"START.cmd", "READY.txt", "config/environment.cmd",
		"tools/nvda/nvda.exe", "tools/vscode/Code.exe",
	} {
		info, err := os.Stat(filepath.Join(absolute, filepath.FromSlash(relative)))
		if err != nil {
			return "", fmt.Errorf("arquivo necessário para iniciar ausente: %s: %w", relative, err)
		}
		if !info.Mode().IsRegular() || info.Size() == 0 {
			return "", fmt.Errorf("arquivo necessário para iniciar inválido: %s", relative)
		}
	}
	return absolute, nil
}
