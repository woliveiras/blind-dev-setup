package release_test

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestReadmeStartsUsersWithPrebuiltDownloads(t *testing.T) {
	readme := readFile(t, filepath.Join(projectRoot(t), "README.md"))
	downloadHeading := strings.Index(readme, "## Baixar e usar")
	buildHeading := strings.Index(readme, "## Compilar o projeto")
	if downloadHeading < 0 {
		t.Fatal("README does not contain the prebuilt download section")
	}
	if buildHeading < 0 || downloadHeading > buildHeading {
		t.Error("README presents source builds before prebuilt downloads")
	}

	for _, required := range []string{
		"https://github.com/woliveiras/blind-dev-setup/releases/latest",
		"https://github.com/woliveiras/blind-dev-setup/releases/latest/download/blind-dev-setup-windows-x64.exe",
		"https://github.com/woliveiras/blind-dev-setup/releases/latest/download/blind-dev-setup-windows-x64.zip",
		"curl.exe -fL",
		"wget --https-only",
		"INICIAR-AQUI.cmd",
		"não precisa instalar Go",
		"não possui assinatura digital",
	} {
		if !strings.Contains(readme, required) {
			t.Errorf("README does not contain %q", required)
		}
	}
}
