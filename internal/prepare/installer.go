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
		if err := os.MkdirAll(destination, 0o755); err != nil {
			return fmt.Errorf("criar destino: %w", err)
		}
		return copyExclusive(artifactPath, filepath.Join(destination, filepath.Base(artifactPath)))
	default:
		return fmt.Errorf("tipo de instalação não suportado: %s", component.Kind)
	}
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
