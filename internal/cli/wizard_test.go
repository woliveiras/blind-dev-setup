package cli

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"
	"github.com/woliveiras/blind-dev-setup/internal/progress"
	"github.com/woliveiras/blind-dev-setup/internal/target"
)

func wizardCandidate() target.Candidate {
	return target.Candidate{DiskNumber: 2, Model: "Meu USB", DriveLetter: "E", Label: "PROJETOS", FileSystem: "NTFS", SizeBytes: 32 << 30, FreeBytes: 20 << 30}
}

func wizardDependencies(t *testing.T, input string) Dependencies {
	directory := t.TempDir()
	t.Setenv("TMPDIR", directory)
	t.Setenv("TEMP", directory)
	t.Setenv("TMP", directory)
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
	deps := wizardDependencies(t, "1\nACEITO\nSIM\n1\n0\n")
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
	for _, expected := range []string{"Meu USB", "PROJETOS", "32.0 GiB", "E:\\", "ACEITO", "SIM", "Preparação e verificação concluídas.", "START.cmd"} {
		if !strings.Contains(output.String(), expected) {
			t.Errorf("missing %q in %s", expected, output.String())
		}
	}
}

func TestWizardLaunchesOnlyAfterConfirmationAndFreshVerification(t *testing.T) {
	deps := wizardDependencies(t, "1\nACEITO\nSIM\nI\n0\n")
	var events []string
	installed := false
	deps.ListTargets = func() ([]target.Candidate, error) {
		candidate := wizardCandidate()
		candidate.DestinationExists = installed
		return []target.Candidate{candidate}, nil
	}
	deps.Prepare = func(PrepareRequest) error {
		installed = true
		events = append(events, "prepare")
		return nil
	}
	deps.Verify = func(string, io.Writer) error {
		events = append(events, "verify")
		return nil
	}
	deps.LaunchEnvironment = func(root string, log io.Writer) error {
		if root != filepath.Join("E:\\", target.DirectoryName) {
			t.Fatalf("launch root = %q", root)
		}
		events = append(events, "launch")
		fmt.Fprintln(log, "launcher diagnostic")
		return nil
	}
	var output bytes.Buffer
	if code := Run(nil, &output, &output, deps); code != 0 {
		t.Fatalf("exit=%d output=%s", code, output.String())
	}
	if strings.Join(events, ",") != "prepare,verify,verify,launch" {
		t.Fatalf("operation order = %v", events)
	}
	if !strings.Contains(output.String(), "Inicialização solicitada para NVDA e VS Code") || strings.Contains(output.String(), "launcher diagnostic") {
		t.Fatal(output.String())
	}
	logs, _ := filepath.Glob(filepath.Join(os.TempDir(), "blind-dev-setup-*.log"))
	if len(logs) != 1 {
		t.Fatalf("logs = %v", logs)
	}
	log, err := os.ReadFile(logs[0])
	if err != nil || !bytes.Contains(log, []byte("launcher diagnostic")) {
		t.Fatalf("missing launch diagnostics: %v; %s", err, log)
	}
}

func TestWizardWillNotLaunchChangedOrUnverifiedEnvironment(t *testing.T) {
	for _, scenario := range []string{"changed device", "failed verification", "not requested"} {
		t.Run(scenario, func(t *testing.T) {
			input := "1\nSIM\nI\n0\n"
			if scenario == "not requested" {
				input = "1\nSIM\n0\n"
			}
			deps := wizardDependencies(t, input)
			lists, verifications := 0, 0
			deps.ListTargets = func() ([]target.Candidate, error) {
				lists++
				candidate := wizardCandidate()
				candidate.DestinationExists = true
				if scenario == "changed device" && lists >= 3 {
					candidate.DiskNumber++
				}
				return []target.Candidate{candidate}, nil
			}
			deps.Verify = func(string, io.Writer) error {
				verifications++
				if scenario == "failed verification" && verifications > 1 {
					return errors.New("environment changed after preparation")
				}
				return nil
			}
			deps.LaunchEnvironment = func(string, io.Writer) error {
				t.Fatal("unconfirmed or changed environment was launched")
				return nil
			}
			var output bytes.Buffer
			code := Run(nil, &output, &output, deps)
			want := 1
			if scenario == "not requested" {
				want = 0
			}
			if code != want || strings.Contains(output.String(), "Inicialização solicitada") {
				t.Fatalf("exit=%d output=%s", code, output.String())
			}
		})
	}
}

func TestWizardCanRetryFailedLaunchWithoutRepreparing(t *testing.T) {
	deps := wizardDependencies(t, "1\nSIM\nI\nI\n0\n")
	deps.ListTargets = func() ([]target.Candidate, error) {
		candidate := wizardCandidate()
		candidate.DestinationExists = true
		return []target.Candidate{candidate}, nil
	}
	deps.Prepare = func(PrepareRequest) error { t.Fatal("must not prepare again"); return nil }
	launches := 0
	deps.LaunchEnvironment = func(string, io.Writer) error {
		launches++
		if launches == 1 {
			return errors.New("launcher failed")
		}
		return nil
	}
	var output bytes.Buffer
	if code := Run(nil, &output, &output, deps); code != 0 || launches != 2 {
		t.Fatalf("exit=%d launches=%d output=%s", code, launches, output.String())
	}
	if strings.Count(output.String(), "Inicialização solicitada") != 1 || !strings.Contains(output.String(), "Não foi possível iniciar o ambiente") {
		t.Fatal(output.String())
	}
}

func TestWizardNeverInfersConsentFromEOFOrRefusal(t *testing.T) {
	for _, input := range []string{"", "0\n", "1\n", "1\nACEITO\n", "1\nACEITO\nSIM", "1\nNAO\n0\n", "1\nACEITO\nV\n0\n", "1\nsim\n0\n", "1\nACEITO\nerrado\n", "1\nerrado\n", "1\nACEITO\n\n"} {
		t.Run(strings.ReplaceAll(input, "\n", "_"), func(t *testing.T) {
			deps := wizardDependencies(t, input)
			deps.Prepare = func(PrepareRequest) error { t.Fatal("prepare called without full consent"); return nil }
			deps.Verify = func(string, io.Writer) error { t.Fatal("verify unexpectedly called"); return nil }
			var output bytes.Buffer
			if code := Run(nil, &output, &output, deps); code != 0 {
				t.Fatalf("code=%d; %s", code, output.String())
			}
			logs, _ := filepath.Glob(filepath.Join(os.TempDir(), "blind-dev-setup-*.log"))
			if len(logs) != 0 {
				t.Fatalf("cancellation created logs: %v", logs)
			}
		})
	}
}

func TestWizardShowsLicensesOnlyOnRequest(t *testing.T) {
	deps := wizardDependencies(t, "1\nD\nACEITO\nSIM\n0\n")
	var output bytes.Buffer
	if code := Run([]string{"wizard", "--accessible"}, &output, &output, deps); code != 0 {
		t.Fatalf("code=%d; %s", code, output.String())
	}
	if !strings.Contains(output.String(), "https://example.test/license") {
		t.Fatal("D did not show license source")
	}
}

func TestWizardCapturesTechnicalLogsAndShowsRealProgress(t *testing.T) {
	deps := wizardDependencies(t, "1\nACEITO\nSIM\n0\n")
	deps.Prepare = func(request PrepareRequest) error {
		fmt.Fprintln(request.Output, "\x1b[31mprivate technical install output\x1b[0m")
		request.Report(progress.Event{Stage: progress.Download, Component: "Editor", Index: 1, Total: 2})
		request.Report(progress.Event{Stage: progress.Extract, Component: "Editor", Index: 1, Total: 2})
		return nil
	}
	deps.Verify = func(_ string, output io.Writer) error { fmt.Fprintln(output, "technical verification"); return nil }
	var output bytes.Buffer
	if code := Run([]string{"wizard", "--accessible"}, &output, &output, deps); code != 0 {
		t.Fatalf("code=%d; %s", code, output.String())
	}
	text := output.String()
	if strings.Contains(text, "private technical") || strings.Contains(text, "technical verification") || strings.Contains(text, "\x1b") {
		t.Fatalf("technical/control output leaked: %q", text)
	}
	for _, expected := range []string{"Baixando Editor (1 de 2)", "Extraindo Editor (1 de 2)", "Conferindo arquivos"} {
		if !strings.Contains(text, expected) {
			t.Errorf("missing %q", expected)
		}
	}
	logs, _ := filepath.Glob(filepath.Join(os.TempDir(), "blind-dev-setup-*.log"))
	if len(logs) != 1 {
		t.Fatalf("logs=%v", logs)
	}
	content, err := os.ReadFile(logs[0])
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(content, []byte("private technical")) || !bytes.Contains(content, []byte("technical verification")) {
		t.Fatalf("log=%s", content)
	}
}

func TestWizardDetailsSanitizeLogsAfterFailure(t *testing.T) {
	deps := wizardDependencies(t, "1\nACEITO\nSIM\nD\n0\n")
	deps.Prepare = func(request PrepareRequest) error {
		fmt.Fprintln(request.Output, "\x1b[2Jdiagnostic detail\x00")
		return errors.New("network interrupted")
	}
	var output bytes.Buffer
	if code := Run([]string{"wizard", "--accessible"}, &output, &output, deps); code != 1 {
		t.Fatalf("code=%d; %s", code, output.String())
	}
	text := output.String()
	for _, expected := range []string{"Registro completo:", "diagnostic detail", "network interrupted"} {
		if !strings.Contains(text, expected) {
			t.Errorf("missing %q; %s", expected, text)
		}
	}
	if strings.ContainsAny(text, "\x1b\x00") {
		t.Fatalf("unsanitized control codes: %q", text)
	}
}

func TestWizardModeFlagsFallbackAndConflict(t *testing.T) {
	for _, flag := range []string{"--accessible", "--visual"} {
		deps := wizardDependencies(t, "0\n")
		var output bytes.Buffer
		if code := Run([]string{"wizard", flag}, &output, &output, deps); code != 0 {
			t.Fatalf("code=%d", code)
		}
		if strings.Contains(output.String(), "\x1b") {
			t.Fatalf("non-TTY received escape sequences: %q", output.String())
		}
		if flag == "--visual" && !strings.Contains(output.String(), "modo para leitor de tela") {
			t.Fatal("missing visual fallback explanation")
		}
	}
	deps := wizardDependencies(t, "")
	if code := Run([]string{"wizard", "--accessible", "--visual"}, io.Discard, io.Discard, deps); code != 2 {
		t.Fatalf("conflict code=%d", code)
	}
}

func TestWizardModeBootstrapDoesNotConsumeVisualKeys(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	raw := strings.NewReader("2\nj\r")
	w := wizard{ctx: ctx, input: newWizardInput(ctx, raw), rawInput: raw, output: io.Discard}
	if !w.chooseMode(false, false, true) || w.accessible {
		t.Fatal("visual mode not selected")
	}
	remaining, _ := io.ReadAll(raw)
	if string(remaining) != "j\r" {
		t.Fatalf("bootstrap consumed keys: %q", remaining)
	}
}

func TestWizardVisualHuhFormActuallyRendersAndAcceptsSelection(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var output bytes.Buffer
	reader, writer := io.Pipe()
	defer reader.Close()
	defer writer.Close()
	go func() {
		select {
		case <-ctx.Done():
			return
		case <-time.After(100 * time.Millisecond):
		}
		_, _ = io.WriteString(writer, "\x1b[B\r")
	}()
	w := wizard{ctx: ctx, rawInput: reader, output: &output, accessible: false}
	value, ok := w.ask("Escolha visual", "Teste do formulário", []wizardOption{{"1", "Primeira opção"}, {"2", "Segunda opção"}})
	if !ok || value != "2" {
		t.Fatalf("ok=%v value=%q output=%q", ok, value, output.String())
	}
	if !strings.Contains(ansi.Strip(output.String()), "Escolha visual") {
		t.Fatalf("visual form did not render: %q", output.String())
	}
}

func TestWizardProgressPanelFitsNarrowTerminal(t *testing.T) {
	for _, width := range []int{12, 24, 40, 76, 120} {
		view := wizardProgressModel{activity: "Baixando um componente com nome longo (1 de 6)...", width: width}.View()
		for _, line := range strings.Split(view, "\n") {
			if ansi.StringWidth(line) > width {
				t.Fatalf("width=%d, line width=%d: %q", width, ansi.StringWidth(line), line)
			}
		}
	}
}

func TestWizardInvalidChoicesAndRescanStayInSameSession(t *testing.T) {
	deps := wizardDependencies(t, "R\n99\n1\nerrado\nACEITO\nSIM\n0\n")
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
			deps := wizardDependencies(t, "1\nACEITO\nSIM\n0\n")
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
			deps := wizardDependencies(t, "1\n0\n")
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
	deps := wizardDependencies(t, "1\nACEITO\nSIM\n1\nACEITO\nSIM\n0\n")
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
	deps := wizardDependencies(t, "1\nSIM\n0\n")
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
	deps := wizardDependencies(t, "1\nACEITO\nSIM\n0\n")
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
	deps := wizardDependencies(t, "1\nACEITO\nSIM\n1\n0\n")
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
	deps := wizardDependencies(t, "1\nACEITO\nSIM\n0\n")
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
	deps := wizardDependencies(t, "1\nACEITO\nSIM\n0\n")
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
	deps := wizardDependencies(t, "")
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
