package verify

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/woliveiras/blind-dev-setup/internal/manifest"
	"github.com/woliveiras/blind-dev-setup/internal/target"
)

var requiredPortableFiles = []string{
	"READY.txt",
	"START.cmd",
	"DEV-SHELL.cmd",
	"START-NVDA.cmd",
	"START-VSCODE.cmd",
	"START-NOTEPADPP.cmd",
	"START-DBEAVER.cmd",
	"config/environment.cmd",
	"config/mise.lock",
	"config/mise.toml",
	"docs/LEIA-ME.txt",
	"docs/NVDA.md",
	"docs/VALIDACAO.md",
	"tools/vscode/data/user-data/User/settings.json",
}

func Run(requestedTarget string, output io.Writer) error {
	root, err := filepath.Abs(filepath.Join(requestedTarget, target.DirectoryName))
	if err != nil {
		return fmt.Errorf("resolver destino: %w", err)
	}
	fmt.Fprintf(output, "Verificando %s\n", root)
	for _, relative := range requiredPortableFiles {
		if err := requireRegularFile(root, relative); err != nil {
			return err
		}
	}
	fmt.Fprintln(output, "Launchers, configurações e documentação: presentes.")

	manifestPath := filepath.Join(root, "config", "manifest.json")
	file, err := os.Open(manifestPath)
	if err != nil {
		return fmt.Errorf("abrir manifesto instalado: %w", err)
	}
	current, parseErr := manifest.Parse(file)
	closeErr := file.Close()
	if parseErr != nil {
		return fmt.Errorf("validar manifesto instalado: %w", parseErr)
	}
	if closeErr != nil {
		return fmt.Errorf("fechar manifesto instalado: %w", closeErr)
	}
	fmt.Fprintf(output, "Manifesto: versão %s; plataforma %s.\n", current.Release, current.Platform)

	for _, component := range current.Components {
		for _, expected := range component.ExpectedFiles {
			relative := filepath.ToSlash(filepath.Join(component.Destination, filepath.FromSlash(expected)))
			if err := requireRegularFile(root, relative); err != nil {
				return fmt.Errorf("%s: %w", component.ID, err)
			}
		}
		fmt.Fprintf(output, "Componente conferido: %s %s.\n", component.Name, component.Version)
	}

	for _, tool := range []string{"python", "node", "uv", "pnpm"} {
		relative := filepath.Join("data", "mise", "installs", tool, current.Toolchain[tool])
		info, err := os.Stat(filepath.Join(root, relative))
		if err != nil {
			return fmt.Errorf("toolchain ausente: %s %s", tool, current.Toolchain[tool])
		}
		if !info.IsDir() {
			return fmt.Errorf("toolchain inválida: %s não é diretório", relative)
		}
		fmt.Fprintf(output, "Toolchain conferida: %s %s.\n", tool, current.Toolchain[tool])
	}

	fmt.Fprintln(output, "Verificação concluída sem modificar o pendrive.")
	return nil
}

func requireRegularFile(root, relative string) error {
	path := filepath.Join(root, filepath.FromSlash(relative))
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("arquivo obrigatório ausente: %s", relative)
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("caminho obrigatório não é arquivo: %s", relative)
	}
	return nil
}
