//go:build windows

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// Open the prepared folder through Explorer's argument API. No command shell
// interprets the selected drive or directory, and no elevation is requested.
func openEnvironment(root string) error {
	absolute, err := filepath.Abs(root)
	if err != nil {
		return err
	}
	info, err := os.Stat(absolute)
	if err != nil {
		return fmt.Errorf("abrir pasta do ambiente: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("o ambiente não é uma pasta: %s", absolute)
	}
	command := exec.Command("explorer.exe", absolute)
	if err := command.Start(); err != nil {
		return fmt.Errorf("abrir Explorador de Arquivos: %w", err)
	}
	// Explorer may hand off to an existing desktop process. Reap the launcher
	// without blocking the wizard or claiming that an application was tested.
	go func() { _ = command.Wait() }()
	return nil
}
