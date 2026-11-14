package prepare

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/woliveiras/blind-dev-setup/internal/manifest"
	"github.com/woliveiras/blind-dev-setup/internal/portablezip"
)

type CommandRunner interface {
	Run(context.Context, string, []string, string, []string, io.Writer) error
}

type ExecRunner struct{}

func (ExecRunner) Run(
	ctx context.Context,
	executable string,
	arguments []string,
	directory string,
	environment []string,
	output io.Writer,
) error {
	command := exec.CommandContext(ctx, executable, arguments...)
	command.Dir = directory
	if environment == nil {
		command.Env = os.Environ()
	} else {
		command.Env = environment
	}
	command.Stdout = output
	command.Stderr = output
	if err := command.Run(); err != nil {
		return fmt.Errorf("executar %s: %w", filepath.Base(executable), err)
	}
	return nil
}

type Installer struct {
	Runner CommandRunner
}

func (i Installer) Install(
	ctx context.Context,
	component manifest.Component,
	artifactPath string,
	root string,
	output io.Writer,
) error {
	destination := filepath.Join(root, filepath.FromSlash(component.Destination))
	switch component.Kind {
	case "zip":
		return portablezip.Extract(artifactPath, destination, component.StripComponents)
	case "git-portable-7zip":
		return i.installPortableGit(ctx, component, artifactPath, root, destination, output)
	case "self-extracting-7zip":
		if i.Runner == nil {
			return errors.New("executor de comandos não configurado")
		}
		if err := os.MkdirAll(destination, 0o755); err != nil {
			return fmt.Errorf("criar destino: %w", err)
		}
		arguments := replaceDestination(component.Arguments, destination)
		return i.Runner.Run(ctx, artifactPath, arguments, root, nil, output)
	case "nvda-portable":
		if i.Runner == nil {
			return errors.New("executor de comandos não configurado")
		}
		if err := os.MkdirAll(destination, 0o755); err != nil {
			return fmt.Errorf("criar destino: %w", err)
		}
		arguments := []string{"--create-portable-silent", "--portable-path=" + destination}
		return i.Runner.Run(ctx, artifactPath, arguments, root, nil, output)
	case "executable":
		if len(component.ExpectedFiles) != 1 || component.ExpectedFiles[0] == "" || component.ExpectedFiles[0] == "." || component.ExpectedFiles[0] == ".." || strings.ContainsAny(component.ExpectedFiles[0], "/\\:") {
			return errors.New("componente executável exige exatamente um nome de arquivo de destino")
		}
		if err := os.MkdirAll(destination, 0o755); err != nil {
			return fmt.Errorf("criar destino: %w", err)
		}
		return copyExclusive(artifactPath, filepath.Join(destination, component.ExpectedFiles[0]))
	default:
		return fmt.Errorf("tipo de instalação não suportado: %s", component.Kind)
	}
}

func (i Installer) installPortableGit(ctx context.Context, component manifest.Component, artifactPath, root, destination string, output io.Writer) error {
	if component.ID != "git" {
		return errors.New("extração git-portable-7zip é exclusiva do componente Git verificado")
	}
	if i.Runner == nil {
		return errors.New("executor de comandos não configurado")
	}
	archivePath, err := filepath.Abs(artifactPath)
	if err != nil {
		return fmt.Errorf("resolver arquivo do Git Portable: %w", err)
	}
	if err := os.MkdirAll(destination, 0o755); err != nil {
		return fmt.Errorf("criar destino: %w", err)
	}
	// The current Git SFX does not implement the old -gm2 switch. Its LZMA lc=8
	// also exceeds what common libarchive/liblzma builds accept. Use the official
	// standalone console extractor installed from our pinned, verified manifest:
	// https://www.7-zip.org/download.html
	// Extract the SHA-256-verified archive without invoking the GUI, then run its
	// exact upstream post-install command (portable/release.sh in build-extra):
	// https://github.com/git-for-windows/build-extra/blob/main/portable/release.sh
	// Do not resolve an extractor from the host PATH. -bsp0 disables its progress
	// ticker; normal diagnostics and stderr remain available in the build log.
	extractor := filepath.Join(root, "tools", "7zip", "7zr.exe")
	if err := i.Runner.Run(ctx, extractor, []string{"x", "-t7z", "-y", "-bb0", "-bsp0", "-o" + destination, archivePath}, destination, nil, output); err != nil {
		return fmt.Errorf("extrair Git Portable com 7-Zip verificado: %w", err)
	}
	arguments := []string{"--needs-console", "--hide", "--no-cd", "--command=post-install.bat"}
	if err := i.Runner.Run(ctx, filepath.Join(destination, "git-bash.exe"), arguments, destination, nil, output); err != nil {
		return fmt.Errorf("configurar Git Portable após extração: %w", err)
	}
	return nil
}

func replaceDestination(arguments []string, destination string) []string {
	result := make([]string, len(arguments))
	for index, argument := range arguments {
		result[index] = strings.ReplaceAll(argument, "{destination}", destination)
	}
	return result
}

func copyExclusive(sourcePath, destinationPath string) error {
	source, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer source.Close()
	destination, err := os.OpenFile(destinationPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o755)
	if err != nil {
		return err
	}
	keep := false
	defer func() {
		_ = destination.Close()
		if !keep {
			_ = os.Remove(destinationPath)
		}
	}()
	if _, err := io.Copy(destination, source); err != nil {
		return err
	}
	if err := destination.Close(); err != nil {
		return err
	}
	keep = true
	return nil
}
