//go:build windows

package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"golang.org/x/sys/windows"
)

// launchEnvironment runs the bundle's established initializer. The selected
// path is a working directory, never shell command text. A successful return
// means START.cmd completed, not that the graphical apps were inspected.
func launchEnvironment(ctx context.Context, root string, output io.Writer) error {
	absolute, err := validateLaunchRoot(root)
	if err != nil {
		return err
	}
	systemDirectory, err := windows.GetSystemDirectory()
	if err != nil {
		return fmt.Errorf("localizar o interpretador de comandos do Windows: %w", err)
	}
	command := exec.CommandContext(ctx, filepath.Join(systemDirectory, "cmd.exe"), "/d", "/c", ".\\START.cmd")
	command.Dir = absolute
	command.Env = append(os.Environ(), "BLIND_DEV_NO_PAUSE=1")
	command.Stdout = output
	command.Stderr = output
	// Nil stdin is NUL, so older launchers containing pause cannot block.
	command.Stdin = nil
	// Inherit the wizard's console. Do not request a new or hidden window:
	// the applications launched by START.cmd should open normally.
	// The wizard supplies a log file directly (no pipe copying). Bound waits
	// nevertheless protect other callers from inherited handles in descendants.
	command.WaitDelay = 2 * time.Second
	if err := command.Run(); err != nil {
		if ctx.Err() != nil {
			return fmt.Errorf("inicialização interrompida: %w", ctx.Err())
		}
		return fmt.Errorf("executar START.cmd: %w", err)
	}
	return nil
}
