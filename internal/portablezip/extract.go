package portablezip

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const maxExtractedBytes uint64 = 8 << 30

func Extract(archivePath, destination string, stripComponents int) error {
	if stripComponents < 0 {
		return errors.New("stripComponents cannot be negative")
	}
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return fmt.Errorf("abrir ZIP: %w", err)
	}
	defer reader.Close()

	var total uint64
	for _, entry := range reader.File {
		total += entry.UncompressedSize64
		if total > maxExtractedBytes {
			return errors.New("ZIP excede o limite de 8 GiB extraídos")
		}
	}
	if err := os.MkdirAll(destination, 0o755); err != nil {
		return fmt.Errorf("criar destino do ZIP: %w", err)
	}

	for _, entry := range reader.File {
		relative, include, err := strippedPath(entry.Name, stripComponents)
		if err != nil {
			return err
		}
		if !include {
			continue
		}
		target, err := safeDestination(destination, relative)
		if err != nil {
			return err
		}
		if entry.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0o755); err != nil {
				return fmt.Errorf("criar diretório %s: %w", relative, err)
			}
			continue
		}
		if entry.Mode()&os.ModeSymlink != 0 || !entry.Mode().IsRegular() {
			return fmt.Errorf("tipo de arquivo ZIP não suportado: %s", entry.Name)
		}
		if err := extractFile(entry, target, relative); err != nil {
			return err
		}
	}
	return nil
}

func strippedPath(name string, stripComponents int) (string, bool, error) {
	if strings.Contains(name, "\\") {
		return "", false, fmt.Errorf("unsafe ZIP path: %s", name)
	}
	clean := strings.TrimSuffix(name, "/")
	parts := strings.Split(clean, "/")
	if len(parts) <= stripComponents {
		return "", false, nil
	}
	parts = parts[stripComponents:]
	for _, part := range parts {
		if part == "" || part == "." || part == ".." {
			return "", false, fmt.Errorf("unsafe ZIP path: %s", name)
		}
	}
	return filepath.Join(parts...), true, nil
}

func safeDestination(root, relative string) (string, error) {
	target := filepath.Join(root, relative)
	rel, err := filepath.Rel(root, target)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("unsafe ZIP path: %s", relative)
	}
	return target, nil
}

func extractFile(entry *zip.File, target, relative string) error {
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return fmt.Errorf("criar diretório para %s: %w", relative, err)
	}
	source, err := entry.Open()
	if err != nil {
		return fmt.Errorf("abrir %s no ZIP: %w", relative, err)
	}
	defer source.Close()
	destination, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o755)
	if err != nil {
		return fmt.Errorf("criar %s: %w", relative, err)
	}
	keep := false
	defer func() {
		_ = destination.Close()
		if !keep {
			_ = os.Remove(target)
		}
	}()
	written, err := io.Copy(destination, io.LimitReader(source, int64(entry.UncompressedSize64)+1))
	if err != nil {
		return fmt.Errorf("extrair %s: %w", relative, err)
	}
	if written != int64(entry.UncompressedSize64) {
		return fmt.Errorf("extrair %s: tamanho inesperado", relative)
	}
	if err := destination.Close(); err != nil {
		return fmt.Errorf("fechar %s: %w", relative, err)
	}
	keep = true
	return nil
}
