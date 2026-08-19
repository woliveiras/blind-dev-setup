package artifact

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/woliveiras/blind-dev-setup/internal/manifest"
)

type Fetcher struct {
	Client *http.Client
}

func (f Fetcher) Fetch(ctx context.Context, component manifest.Component, cache string) (string, error) {
	if err := os.MkdirAll(cache, 0o755); err != nil {
		return "", fmt.Errorf("criar cache: %w", err)
	}
	finalPath := filepath.Join(cache, component.SHA256+".artifact")
	if valid, err := verifyFile(finalPath, component); err != nil {
		return "", err
	} else if valid {
		return finalPath, nil
	}
	if info, err := os.Lstat(finalPath); err == nil {
		if !info.Mode().IsRegular() {
			return "", fmt.Errorf("cache de %s não é um arquivo regular", component.ID)
		}
		if err := os.Remove(finalPath); err != nil {
			return "", fmt.Errorf("remover cache inválido de %s: %w", component.ID, err)
		}
	} else if !os.IsNotExist(err) {
		return "", fmt.Errorf("inspecionar cache de %s: %w", component.ID, err)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, component.URL, nil)
	if err != nil {
		return "", fmt.Errorf("criar requisição para %s: %w", component.ID, err)
	}
	client := f.Client
	if client == nil {
		client = http.DefaultClient
	}
	response, err := client.Do(request)
	if err != nil {
		return "", fmt.Errorf("baixar %s: %w", component.ID, err)
	}
	defer response.Body.Close()
	if response.Request.URL.Scheme != "https" {
		return "", fmt.Errorf("baixar %s: redirecionamento para protocolo não seguro", component.ID)
	}
	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("baixar %s: servidor respondeu HTTP %d", component.ID, response.StatusCode)
	}
	if response.ContentLength >= 0 && response.ContentLength != component.DownloadBytes {
		return "", fmt.Errorf(
			"baixar %s: tamanho inesperado: recebido %d, esperado %d",
			component.ID,
			response.ContentLength,
			component.DownloadBytes,
		)
	}

	temporary, err := os.CreateTemp(cache, ".partial-*")
	if err != nil {
		return "", fmt.Errorf("criar arquivo temporário para %s: %w", component.ID, err)
	}
	temporaryPath := temporary.Name()
	keep := false
	defer func() {
		_ = temporary.Close()
		if !keep {
			_ = os.Remove(temporaryPath)
		}
	}()

	hasher := sha256.New()
	limited := io.LimitReader(response.Body, component.DownloadBytes+1)
	written, copyErr := io.Copy(io.MultiWriter(temporary, hasher), limited)
	if copyErr != nil {
		return "", fmt.Errorf("salvar %s: %w", component.ID, copyErr)
	}
	if written != component.DownloadBytes {
		return "", fmt.Errorf(
			"baixar %s: tamanho inesperado: recebido %d, esperado %d",
			component.ID,
			written,
			component.DownloadBytes,
		)
	}
	actualSHA := fmt.Sprintf("%x", hasher.Sum(nil))
	if actualSHA != component.SHA256 {
		return "", fmt.Errorf("baixar %s: SHA-256 não confere", component.ID)
	}
	if err := temporary.Sync(); err != nil {
		return "", fmt.Errorf("sincronizar %s: %w", component.ID, err)
	}
	if err := temporary.Close(); err != nil {
		return "", fmt.Errorf("fechar %s: %w", component.ID, err)
	}
	if err := os.Rename(temporaryPath, finalPath); err != nil {
		if valid, verifyErr := verifyFile(finalPath, component); verifyErr == nil && valid {
			return finalPath, nil
		}
		return "", fmt.Errorf("publicar %s no cache: %w", component.ID, err)
	}
	keep = true
	return finalPath, nil
}

func verifyFile(path string, component manifest.Component) (bool, error) {
	file, err := os.Open(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("abrir cache de %s: %w", component.ID, err)
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return false, fmt.Errorf("ler cache de %s: %w", component.ID, err)
	}
	if !info.Mode().IsRegular() || info.Size() != component.DownloadBytes {
		return false, nil
	}
	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return false, fmt.Errorf("verificar cache de %s: %w", component.ID, err)
	}
	return fmt.Sprintf("%x", hasher.Sum(nil)) == component.SHA256, nil
}
