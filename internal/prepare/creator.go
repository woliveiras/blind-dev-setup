package prepare

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/woliveiras/blind-dev-setup/internal/manifest"
	"github.com/woliveiras/blind-dev-setup/internal/target"
)

type Fetcher interface {
	Fetch(context.Context, manifest.Component, string) (string, error)
}

type ComponentInstaller interface {
	Install(context.Context, manifest.Component, string, string, io.Writer) error
}

type Materializer interface {
	Write(root string, manifestJSON []byte) error
}

type ToolchainInstaller interface {
	Install(context.Context, string, manifest.Manifest, io.Writer) error
}

type Creator struct {
	Manifest     manifest.Manifest
	ManifestJSON []byte
	Fetch        Fetcher
	Install      ComponentInstaller
	Materialize  Materializer
	Toolchain    ToolchainInstaller
	Inspect      func(string) (target.Details, error)
}

func (c Creator) Run(ctx context.Context, requestedTarget, cache string, output io.Writer) error {
	if err := c.validateDependencies(); err != nil {
		return err
	}
	details, err := c.Inspect(requestedTarget)
	if err != nil {
		return err
	}
	if err := target.Validate(details, c.Manifest.MinimumFreeBytes); err != nil {
		return err
	}
	if cache == "" {
		userCache, err := os.UserCacheDir()
		if err != nil {
			return fmt.Errorf("identificar diretório de cache: %w", err)
		}
		cache = filepath.Join(userCache, "blind-dev-setup", "downloads")
	}

	staging, err := os.MkdirTemp(details.Path, ".blind-dev-setup-staging-")
	if err != nil {
		return fmt.Errorf("criar área temporária no destino: %w", err)
	}
	defer os.RemoveAll(staging)

	fmt.Fprintf(output, "Destino validado: %s; NTFS; %.1f GiB livres.\n", details.Path, float64(details.FreeBytes)/(1<<30))
	for index, component := range c.Manifest.Components {
		fmt.Fprintf(output, "Componente %d de %d: %s %s.\n", index+1, len(c.Manifest.Components), component.Name, component.Version)
		artifactPath, err := c.Fetch.Fetch(ctx, component, cache)
		if err != nil {
			return fmt.Errorf("%s: %w", component.ID, err)
		}
		fmt.Fprintln(output, "Download conferido por tamanho e SHA-256.")
		if err := c.Install.Install(ctx, component, artifactPath, staging, output); err != nil {
			return fmt.Errorf("instalar %s: %w", component.ID, err)
		}
		if err := verifyComponent(staging, component); err != nil {
			return err
		}
		fmt.Fprintln(output, "Componente instalado e estrutura conferida.")
	}

	fmt.Fprintln(output, "Criando configurações, launchers e documentação no pendrive.")
	if err := c.Materialize.Write(staging, c.ManifestJSON); err != nil {
		return fmt.Errorf("criar arquivos do ambiente: %w", err)
	}
	fmt.Fprintln(output, "Instalando Python, Node.js, uv e pnpm pelo mise.")
	if err := c.Toolchain.Install(ctx, staging, c.Manifest, output); err != nil {
		return fmt.Errorf("instalar toolchain: %w", err)
	}

	ready := fmt.Sprintf(
		"blind-dev-setup %s\nStatus: pronto\nPlataforma: %s\nUse START.cmd ou DEV-SHELL.cmd.\n",
		c.Manifest.Release,
		c.Manifest.Platform,
	)
	if err := os.WriteFile(filepath.Join(staging, "READY.txt"), []byte(ready), 0o644); err != nil {
		return fmt.Errorf("criar marcador de conclusão: %w", err)
	}

	finalPath := filepath.Join(details.Path, target.DirectoryName)
	if err := os.Rename(staging, finalPath); err != nil {
		return fmt.Errorf("publicar ambiente concluído: %w", err)
	}
	fmt.Fprintf(output, "Preparação concluída: %s\n", finalPath)
	return nil
}

func (c Creator) validateDependencies() error {
	switch {
	case c.Fetch == nil:
		return errors.New("dependência de download não configurada")
	case c.Install == nil:
		return errors.New("dependência de instalação não configurada")
	case c.Materialize == nil:
		return errors.New("dependência de materialização não configurada")
	case c.Toolchain == nil:
		return errors.New("dependência de toolchain não configurada")
	case c.Inspect == nil:
		return errors.New("dependência de inspeção do destino não configurada")
	}
	return nil
}

func verifyComponent(root string, component manifest.Component) error {
	for _, relative := range component.ExpectedFiles {
		path := filepath.Join(root, component.Destination, filepath.FromSlash(relative))
		info, err := os.Stat(path)
		if err != nil {
			return fmt.Errorf("instalar %s: arquivo esperado ausente: %s", component.ID, relative)
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("instalar %s: caminho esperado não é arquivo: %s", component.ID, relative)
		}
	}
	return nil
}
