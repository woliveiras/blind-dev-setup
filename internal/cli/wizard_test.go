package cli

import (
	"bytes"
	"context"
	"errors"
	"io"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/woliveiras/blind-dev-setup/internal/target"
)

func wizardCandidate() target.Candidate {
	return target.Candidate{DiskNumber: 2, Model: "Meu USB", DriveLetter: "E", Label: "PROJETOS", FileSystem: "NTFS", SizeBytes: 32 << 30, FreeBytes: 20 << 30}
}

func wizardDependencies(input string) Dependencies {
	return Dependencies{
		Manifest: cliManifest(), Input: strings.NewReader(input),
		ListTargets: func() ([]target.Candidate, error) { return []target.Candidate{wizardCandidate()}, nil },
		Prepare:     func(PrepareRequest) error { return nil },
		Verify:      func(string, io.Writer) error { return nil },
	}
}

func TestWizardPreparesVerifiesAndOpensInOneSession(t *testing.T) {
	var output bytes.Buffer
	var events []string
	deps := wizardDependencies("1\nACEITO\nSIM\n1\n0\n")
	deps.Prepare = func(request PrepareRequest) error {
		if request.Target != "E:\\" {
			t.Fatalf("target = %q", request.Target)
		}
		events = append(events, "prepare")
		return nil
	}
	deps.Verify = func(path string, out io.Writer) error {
		if path != "E:\\" {
			t.Fatalf("verify target = %q", path)
		}
		events = append(events, "verify")
		return nil
	}
	deps.OpenEnvironment = func(path string) error {
		if path != filepath.Join("E:\\", target.DirectoryName) {
			t.Fatalf("open target = %q", path)
		}
		events = append(events, "open")
		return nil
	}
	if code := Run(nil, &output, &output, deps); code != 0 {
		t.Fatalf("code=%d; %s", code, output.String())
	}
	if strings.Join(events, ",") != "prepare,verify,open" {
		t.Fatalf("events = %v", events)
	}
	for _, expected := range []string{"Meu USB", "PROJETOS", "32.0 GiB", "E:\\", "https://example.test/license", "ACEITO", "SIM", "Preparação e verificação concluídas.", "START.cmd"} {
		if !strings.Contains(output.String(), expected) {
			t.Errorf("missing %q in %s", expected, output.String())
		}
	}
}

func TestWizardNeverInfersConsentFromEOFOrRefusal(t *testing.T) {
	for _, input := range []string{"", "0\n", "1\n", "1\nACEITO\n", "1\nACEITO\nSIM", "1\nNAO\n0\n", "1\nACEITO\nV\n0\n", "1\nsim\n0\n"} {
		t.Run(strings.ReplaceAll(input, "\n", "_"), func(t *testing.T) {
			deps := wizardDependencies(input)
			deps.Prepare = func(PrepareRequest) error { t.Fatal("prepare called without full consent"); return nil }
			deps.Verify = func(string, io.Writer) error { t.Fatal("verify unexpectedly called"); return nil }
			var output bytes.Buffer
			if code := Run(nil, &output, &output, deps); code != 0 {
				t.Fatalf("code=%d; %s", code, output.String())
			}
		})
	}
}

func TestWizardInvalidChoicesAndRescanStayInSameSession(t *testing.T) {
	deps := wizardDependencies("R\n99\n1\nerrado\nACEITO\nSIM\n0\n")
	lists, prepared := 0, 0
	deps.ListTargets = func() ([]target.Candidate, error) {
		lists++
		if lists == 1 {
			return nil, nil
		}
		return []target.Candidate{wizardCandidate()}, nil
	}
	deps.Prepare = func(PrepareRequest) error { prepared++; return nil }
	var output bytes.Buffer
	if code := Run(nil, &output, &output, deps); code != 0 || prepared != 1 {
		t.Fatalf("code=%d prepared=%d; %s", code, prepared, output.String())
	}
	if !strings.Contains(output.String(), "Nenhum pendrive") || !strings.Contains(output.String(), "Opção inválida") || !strings.Contains(output.String(), "Resposta inválida") {
		t.Fatalf("missing recovery guidance: %s", output.String())
	}
}

func TestWizardRefusesChangedOrUnavailableTargetsBeforeWriting(t *testing.T) {
	for _, scenario := range []string{"disappeared", "disk", "model", "size", "label", "filesystem", "existing", "space", "discovery error"} {
		t.Run(scenario, func(t *testing.T) {
			deps := wizardDependencies("1\nACEITO\nSIM\n0\n")
			calls := 0
			deps.ListTargets = func() ([]target.Candidate, error) {
				calls++
				c := wizardCandidate()
				if calls == 1 {
					return []target.Candidate{c}, nil
				}
				switch scenario {
				case "disappeared":
					return nil, nil
				case "disk":
					c.DiskNumber++
				case "model":
					c.Model = "Other USB"
				case "size":
					c.SizeBytes++
				case "label":
					c.Label = "OTHER"
				case "filesystem":
					c.FileSystem = "exFAT"
				case "existing":
					c.DestinationExists = true
				case "space":
					c.FreeBytes = 0
				case "discovery error":
					return nil, errors.New("USB lookup failed")
				}
				return []target.Candidate{c}, nil
			}
			deps.Prepare = func(PrepareRequest) error { t.Fatal("wrote after target changed"); return nil }
			var output bytes.Buffer
			if code := Run(nil, &output, &output, deps); code != 1 || !strings.Contains(output.String(), "O destino mudou") {
				t.Fatalf("code=%d; %s", code, output.String())
			}
		})
	}
}

func TestWizardIneligibleTargetsNeverPrepare(t *testing.T) {
	for _, scenario := range []string{"system", "exfat", "space", "no letter"} {
		t.Run(scenario, func(t *testing.T) {
			deps := wizardDependencies("1\n0\n")
			deps.ListTargets = func() ([]target.Candidate, error) {
				c := wizardCandidate()
				switch scenario {
				case "system":
					c.IsSystem = true
				case "exfat":
					c.FileSystem = "exFAT"
				case "space":
					c.FreeBytes = 0
				case "no letter":
					c.DriveLetter = ""
				}
				return []target.Candidate{c}, nil
			}
			deps.Prepare = func(PrepareRequest) error { t.Fatal("prepared ineligible target"); return nil }
			var output bytes.Buffer
			Run(nil, &output, &output, deps)
			if !strings.Contains(output.String(), "Não é possível usar este destino") {
				t.Fatal(output.String())
			}
		})
	}
}

func TestWizardRetriesFailedPreparationWithoutRestarting(t *testing.T) {
	deps := wizardDependencies("1\nACEITO\nSIM\n1\nACEITO\nSIM\n0\n")
	prepares, verifies := 0, 0
	deps.Prepare = func(PrepareRequest) error {
		prepares++
		if prepares == 1 {
			return errors.New("temporary network error")
		}
		return nil
	}
	deps.Verify = func(string, io.Writer) error { verifies++; return nil }
	var output bytes.Buffer
	if code := Run(nil, &output, &output, deps); code != 0 || prepares != 2 || verifies != 1 {
		t.Fatalf("code=%d prepares=%d verifies=%d; %s", code, prepares, verifies, output.String())
	}
}

func TestWizardExistingInstallationCanVerifyWithoutFreeSpace(t *testing.T) {
	deps := wizardDependencies("1\nSIM\n0\n")
	deps.ListTargets = func() ([]target.Candidate, error) {
		c := wizardCandidate()
		c.DestinationExists = true
		c.FreeBytes = 0
		return []target.Candidate{c}, nil
	}
	deps.Prepare = func(PrepareRequest) error { t.Fatal("existing installation must not prepare"); return nil }
	calls := 0
	deps.Verify = func(string, io.Writer) error { calls++; return nil }
	var output bytes.Buffer
	if code := Run(nil, &output, &output, deps); code != 0 || calls != 1 {
		t.Fatalf("code=%d verifies=%d; %s", code, calls, output.String())
	}
	if strings.Contains(output.String(), "Digite ACEITO") || !strings.Contains(output.String(), "Verificação concluída.") {
		t.Fatal(output.String())
	}
}

func TestWizardDoesNotAnnounceSuccessAfterVerificationFailure(t *testing.T) {
	deps := wizardDependencies("1\nACEITO\nSIM\n0\n")
	deps.Verify = func(string, io.Writer) error { return errors.New("python.exe missing") }
	var output bytes.Buffer
	if code := Run(nil, &output, &output, deps); code != 1 {
		t.Fatalf("code=%d; %s", code, output.String())
	}
	if strings.Contains(output.String(), "Preparação e verificação concluídas.") || !strings.Contains(output.String(), "O ambiente não foi aprovado") {
		t.Fatal(output.String())
	}
}

func TestWizardOpenFailureHasRecoveryMenu(t *testing.T) {
	deps := wizardDependencies("1\nACEITO\nSIM\n1\n0\n")
	deps.OpenEnvironment = func(string) error { return errors.New("Explorer unavailable") }
	var output bytes.Buffer
	if code := Run(nil, &output, &output, deps); code != 1 {
		t.Fatalf("code=%d; %s", code, output.String())
	}
	if !strings.Contains(output.String(), "Não foi possível abrir a pasta") || strings.Contains(output.String(), "Pasta do ambiente aberta.") {
		t.Fatal(output.String())
	}
}

func TestWizardExplicitTargetUsesInspectionAndStillRequiresConsent(t *testing.T) {
	deps := wizardDependencies("1\nACEITO\nSIM\n0\n")
	deps.ListTargets = func() ([]target.Candidate, error) { t.Fatal("explicit target must use inspection"); return nil, nil }
	inspected := 0
	deps.Inspect = func(path string) (target.Details, error) {
		inspected++
		if path != "E:\\" {
			t.Fatalf("inspect path=%q", path)
		}
		return target.Details{Path: path, Volume: "E:", SystemVolume: "C:", FileSystem: "NTFS", FreeBytes: 20 << 30}, nil
	}
	prepared := false
	deps.Prepare = func(r PrepareRequest) error {
		prepared = true
		if r.Target != "E:\\" || r.Cache != "cache" {
			t.Fatalf("request=%+v", r)
		}
		return nil
	}
	var output bytes.Buffer
	if code := Run([]string{"wizard", "--target", "E:\\", "--cache", "cache"}, &output, &output, deps); code != 0 || !prepared || inspected != 2 {
		t.Fatalf("code=%d prepared=%v inspected=%d; %s", code, prepared, inspected, output.String())
	}
}

func TestWizardExplicitTargetRevalidatesBeforePreparation(t *testing.T) {
	deps := wizardDependencies("1\nACEITO\nSIM\n0\n")
	count := 0
	deps.Inspect = func(path string) (target.Details, error) {
		count++
		details := target.Details{Path: path, Volume: "E:", SystemVolume: "C:", FileSystem: "NTFS", FreeBytes: 20 << 30}
		if count > 1 {
			details.SystemVolume = "E:"
		}
		return details, nil
	}
	deps.Prepare = func(PrepareRequest) error { t.Fatal("must not prepare changed volume"); return nil }
	var output bytes.Buffer
	if code := Run([]string{"wizard", "--target", "E:\\"}, &output, &output, deps); code != 1 {
		t.Fatalf("code=%d; %s", code, output.String())
	}
}

func TestWizardCancellationInterruptsBlockedInput(t *testing.T) {
	reader, writer := io.Pipe()
	defer reader.Close()
	defer writer.Close()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	deps := wizardDependencies("")
	deps.Context = ctx
	deps.Input = reader
	listed := make(chan struct{})
	deps.ListTargets = func() ([]target.Candidate, error) { close(listed); return nil, nil }
	done := make(chan int, 1)
	go func() { done <- Run(nil, io.Discard, io.Discard, deps) }()
	<-listed
	cancel()
	select {
	case code := <-done:
		if code != 130 {
			t.Fatalf("code=%d", code)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("wizard ignored cancellation")
	}
}
