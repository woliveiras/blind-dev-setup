package verify

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"sort"

	"github.com/woliveiras/blind-dev-setup/internal/manifest"
	"github.com/woliveiras/blind-dev-setup/internal/target"
	"github.com/woliveiras/blind-dev-setup/internal/templates"
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

func Run(requestedTarget string, expected manifest.Manifest, output io.Writer) error {
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
	if !reflect.DeepEqual(current, expected) {
		return fmt.Errorf("o manifesto instalado difere do manifesto incorporado neste gerador")
	}
	fmt.Fprintf(output, "Manifesto: versão %s; plataforma %s.\n", current.Release, current.Platform)
	managed, err := templates.ManagedFiles(current)
	if err != nil {
		return fmt.Errorf("obter configurações gerenciadas: %w", err)
	}
	paths := make([]string, 0, len(managed))
	for relative := range managed {
		paths = append(paths, relative)
	}
	sort.Strings(paths)
	for _, relative := range paths {
		if err := requireRegularFile(root, relative); err != nil {
			return err
		}
		actual, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(relative)))
		if err != nil {
			return fmt.Errorf("ler configuração gerenciada %s: %w", relative, err)
		}
		if !bytes.Equal(actual, managed[relative]) {
			return fmt.Errorf("configuração gerenciada difere do gerador: %s; restaure o arquivo original desta versão", relative)
		}
	}
	fmt.Fprintln(output, "Configurações gerenciadas: conteúdo confere com este gerador.")

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
		// The pinned Windows core and aqua distributions expose their executable
		// at the install root. A version directory alone is not an installation.
		executable := filepath.Join(relative, tool+".exe")
		if err := requireRegularFile(root, executable); err != nil {
			return fmt.Errorf("toolchain %s %s: %w", tool, current.Toolchain[tool], err)
		}
		fmt.Fprintf(output, "Executável presente e não vazio: %s %s.\n", tool, current.Toolchain[tool])
	}

	fmt.Fprintln(output, "Verificação estrutural: os programas não foram executados; os binários instalados não tiveram seus hashes recalculados.")
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
	if info.Size() == 0 {
		return fmt.Errorf("arquivo obrigatório vazio: %s", relative)
	}
	return nil
}
