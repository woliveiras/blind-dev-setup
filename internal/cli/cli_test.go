package cli

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/woliveiras/blind-dev-setup/internal/manifest"
	"github.com/woliveiras/blind-dev-setup/internal/target"
)

func TestListTargetsExplainsEachPendriveAndItsNextStep(t *testing.T) {
	var stdout, stderr bytes.Buffer
	exitCode := Run(
		[]string{"list-targets"},
		&stdout,
		&stderr,
		Dependencies{
			Manifest: cliManifest(),
			ListTargets: func() ([]target.Candidate, error) {
				return []target.Candidate{
					{
						DiskNumber:  2,
						Model:       "Kingston DataTraveler",
						DriveLetter: "E",
						Label:       "MEU PENDRIVE",
						FileSystem:  "NTFS",
						SizeBytes:   32 << 30,
						FreeBytes:   20 << 30,
					},
					{
						DiskNumber: 3,
						Model:      "USB Flash Disk",
						SizeBytes:  16 << 30,
					},
				}, nil
			},
		},
	)

	if exitCode != 0 {
		t.Fatalf("Run() exit = %d, stderr = %q", exitCode, stderr.String())
	}
	for _, want := range []string{
		"2 pendrives encontrados",
		"Pendrive 1",
		"Nome: Kingston DataTraveler",
		"Letra: E:\\",
		"Nome do volume: MEU PENDRIVE",
		"Sistema de arquivos: NTFS",
		"Espaço livre: 20.0 GiB",
		"Situação: pronto para preparar",
		`.\blind-dev-setup-windows-x64.exe plan --target E:\`,
		"Pendrive 2",
		"Letra: não atribuída pelo Windows",
		"Situação: precisa de atenção",
		"Motivo: o Windows não atribuiu uma letra a este pendrive",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Errorf("stdout does not contain %q:\n%s", want, stdout.String())
		}
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q", stderr.String())
	}
}

func TestListTargetsExplainsWhenNoPendriveWasFound(t *testing.T) {
	var stdout, stderr bytes.Buffer
	exitCode := Run(
		[]string{"list-targets"},
		&stdout,
		&stderr,
		Dependencies{
			Manifest: cliManifest(),
			ListTargets: func() ([]target.Candidate, error) {
				return nil, nil
			},
		},
	)

	if exitCode != 0 {
		t.Fatalf("Run() exit = %d, stderr = %q", exitCode, stderr.String())
	}
	for _, want := range []string{
		"Nenhum pendrive foi encontrado.",
		"Conecte o pendrive, aguarde alguns segundos",
		`.\blind-dev-setup-windows-x64.exe list-targets`,
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Errorf("stdout does not contain %q:\n%s", want, stdout.String())
		}
	}
}

func TestListTargetsTurnsDiscoveryFailureIntoActionableMessage(t *testing.T) {
	var stdout, stderr bytes.Buffer
	exitCode := Run(
		[]string{"list-targets"},
		&stdout,
		&stderr,
		Dependencies{
			Manifest: cliManifest(),
			ListTargets: func() ([]target.Candidate, error) {
				return nil, errors.New("PowerShell indisponível")
			},
		},
	)

	if exitCode != 1 {
		t.Fatalf("Run() exit = %d, stderr = %q", exitCode, stderr.String())
	}
	for _, want := range []string{
		"Erro: não foi possível procurar pendrives.",
		"Detalhes: PowerShell indisponível",
		"Feche e abra o programa novamente",
	} {
		if !strings.Contains(stderr.String(), want) {
			t.Errorf("stderr does not contain %q:\n%s", want, stderr.String())
		}
	}
}

func TestListTargetsDirectsPreparedPendriveToVerification(t *testing.T) {
	var stdout, stderr bytes.Buffer
	exitCode := Run(
		[]string{"list-targets"},
		&stdout,
		&stderr,
		Dependencies{
			Manifest: cliManifest(),
			ListTargets: func() ([]target.Candidate, error) {
				return []target.Candidate{{
					Model:             "Kingston",
					DriveLetter:       "E",
					FileSystem:        "NTFS",
					FreeBytes:         20 << 30,
					DestinationExists: true,
				}}, nil
			},
		},
	)

	if exitCode != 0 {
		t.Fatalf("Run() exit = %d, stderr = %q", exitCode, stderr.String())
	}
	for _, want := range []string{
		"Situação: o ambiente já está preparado",
		`.\blind-dev-setup-windows-x64.exe verify --target E:\`,
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Errorf("stdout does not contain %q:\n%s", want, stdout.String())
		}
	}
}

func TestListTargetsExplainsThatItDoesNotFormatAnIncompatiblePendrive(t *testing.T) {
	var stdout, stderr bytes.Buffer
	exitCode := Run(
		[]string{"list-targets"},
		&stdout,
		&stderr,
		Dependencies{
			Manifest: cliManifest(),
			ListTargets: func() ([]target.Candidate, error) {
				return []target.Candidate{{
					Model:       "Kingston",
					DriveLetter: "E",
					FileSystem:  "exFAT",
					FreeBytes:   20 << 30,
				}}, nil
			},
		},
	)

	if exitCode != 0 {
		t.Fatalf("Run() exit = %d, stderr = %q", exitCode, stderr.String())
	}
	for _, want := range []string{
		"Motivo: o sistema de arquivos é exFAT; esta versão precisa de NTFS.",
		"Este programa não formata pendrives.",
		"Formatar apaga os arquivos",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Errorf("stdout does not contain %q:\n%s", want, stdout.String())
		}
	}
}

func TestPlanPrintsAllActionsWithoutWriting(t *testing.T) {
	var stdout, stderr bytes.Buffer
	exitCode := Run(
		[]string{"plan", "--target", "E:\\"},
		&stdout,
		&stderr,
		Dependencies{Manifest: cliManifest()},
	)

	if exitCode != 0 {
		t.Fatalf("Run() exit = %d, stderr = %q", exitCode, stderr.String())
	}
	for _, want := range []string{
		"Destino: E:\\",
		"Editor 1.2.3",
		"Python 3.14.7",
		"nenhum arquivo será modificado",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Errorf("stdout does not contain %q:\n%s", want, stdout.String())
		}
	}
}

func TestPrepareRequiresLicenseAcceptance(t *testing.T) {
	var stdout, stderr bytes.Buffer
	exitCode := Run(
		[]string{"prepare", "--target", "E:\\"},
		&stdout,
		&stderr,
		Dependencies{Manifest: cliManifest()},
	)

	if exitCode != 2 || !strings.Contains(stderr.String(), "--accept-licenses") {
		t.Fatalf("Run() exit = %d, stderr = %q", exitCode, stderr.String())
	}
}

func TestUnknownCommandReturnsUsageError(t *testing.T) {
	var stdout, stderr bytes.Buffer
	exitCode := Run([]string{"destroy"}, &stdout, &stderr, Dependencies{Manifest: cliManifest()})
	if exitCode != 2 || !strings.Contains(stderr.String(), "Comando desconhecido") {
		t.Fatalf("Run() exit = %d, stderr = %q", exitCode, stderr.String())
	}
}

func cliManifest() manifest.Manifest {
	return manifest.Manifest{
		SchemaVersion:    1,
		Release:          "0.1.0",
		Platform:         "windows-x64",
		MinimumFreeBytes: 1024,
		Toolchain: map[string]string{
			"python": "3.14.7",
			"node":   "24.19.0",
			"uv":     "0.12.5",
			"pnpm":   "11.22.0",
		},
		Components: []manifest.Component{{
			ID:            "editor",
			Name:          "Editor",
			Version:       "1.2.3",
			Required:      true,
			URL:           "https://example.test/editor.zip",
			SHA256:        "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			License:       "Example",
			LicenseURL:    "https://example.test/license",
			Kind:          "zip",
			Destination:   "tools/editor",
			DownloadBytes: 100,
			ExpectedFiles: []string{"editor.exe"},
		}},
	}
}
