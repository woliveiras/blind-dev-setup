package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/woliveiras/blind-dev-setup/internal/progress"
	"github.com/woliveiras/blind-dev-setup/internal/target"
)

func runWizard(args []string, stdout, stderr io.Writer, deps Dependencies) int {
	flags := flag.NewFlagSet("wizard", flag.ContinueOnError)
	flags.SetOutput(stderr)
	explicit := flags.String("target", "", "destino explícito; ainda exige seleção e confirmação")
	cache := flags.String("cache", "", "diretório opcional para downloads verificados")
	accessible := flags.Bool("accessible", false, "iniciar com leitor de tela/NVDA")
	visual := flags.Bool("visual", false, "iniciar sem leitor de tela")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if flags.NArg() != 0 || (*accessible && *visual) {
		fmt.Fprintln(stderr, "Erro: use --accessible ou --visual, sem argumentos adicionais.")
		return 2
	}
	input := deps.Input
	if input == nil {
		input = strings.NewReader("")
	}
	ctx := deps.Context
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	w := wizard{deps: deps, ctx: ctx, rawInput: input, output: stdout, explicitTarget: strings.TrimSpace(*explicit), cache: *cache}
	w.input = newWizardInput(ctx, input)
	fmt.Fprintf(stdout, "blind-dev-setup %s\n", deps.Manifest.Release)
	if !w.chooseMode(*accessible, *visual, terminalIO(input, stdout)) {
		return w.status
	}
	for {
		if w.ctx.Err() != nil {
			w.endInput(w.ctx.Err())
			return w.status
		}
		w.step(1, "Escolher pendrive")
		choices, err := w.discover()
		if err != nil {
			w.fail("Não foi possível procurar pendrives", err)
		}
		if err == nil && len(choices) == 0 {
			fmt.Fprintln(stdout, "Nenhum pendrive encontrado. Conecte-o e procure novamente.")
		}
		options := make([]wizardOption, 0, len(choices)+3)
		for i, choice := range choices {
			label := choice.label()
			if issues := w.issues(choice); len(issues) > 0 {
				label += " · precisa de atenção"
			} else if choice.exists() {
				label += " · verificar instalação"
			}
			options = append(options, wizardOption{strconv.Itoa(i + 1), label})
		}
		options = append(options, wizardOption{"R", "Procurar novamente"})
		if w.logPath != "" {
			options = append(options, wizardOption{"D", "Ver detalhes da última tentativa"})
		}
		options = append(options, wizardOption{"0", "Sair"})
		answer, ok := w.ask("Qual pendrive deseja usar?", "", options)
		if !ok || answer == "0" {
			return w.status
		}
		if answer == "R" {
			continue
		}
		if answer == "D" {
			w.showDetails()
			continue
		}
		number, _ := strconv.Atoi(answer)
		chosen := choices[number-1]
		if issues := w.issues(chosen); len(issues) > 0 {
			fmt.Fprintf(stdout, "Não é possível usar este destino: %s.\n", strings.Join(issues, "; "))
			fmt.Fprintln(stdout, "Use outro pendrive NTFS ou corrija o problema. Nenhuma alteração foi feita.")
			continue
		}
		if !chosen.exists() {
			w.step(2, "Revisar aplicativos e licenças")
			summary := w.applicationSummary()
			action := w.consent("Aceitar licenças", summary, "ACEITO", w.showLicenses)
			if action == "exit" {
				return w.status
			}
			if action == "back" {
				continue
			}
		}
		w.step(3, "Confirmar destino")
		description := chosen.summary() + "\nOs arquivos atuais serão preservados."
		title := "Confirmar preparação"
		if chosen.exists() {
			title = "Confirmar verificação"
			description = chosen.summary() + "\nA instalação existente será apenas verificada."
		}
		action := w.consent(title, description, "SIM", nil)
		if action == "exit" {
			return w.status
		}
		if action == "back" {
			continue
		}
		if w.ctx.Err() != nil {
			w.endInput(w.ctx.Err())
			return w.status
		}
		if err := w.revalidate(chosen); err != nil {
			w.fail("O destino mudou ou não está mais disponível. Selecione-o novamente", err)
			continue
		}
		if deps.Verify == nil || (!chosen.exists() && deps.Prepare == nil) {
			w.fail("Esta compilação não pode concluir a operação", errors.New("preparação ou verificação indisponível"))
			continue
		}
		w.step(4, "Preparar e verificar")
		if err := w.perform(chosen); err != nil {
			w.fail("Não foi possível concluir a operação", err)
			fmt.Fprintln(stdout, "O ambiente não foi aprovado. Selecione o destino para tentar novamente ou D para detalhes.")
			continue
		}
		w.status = 0
		w.step(5, "Concluir")
		if chosen.exists() {
			fmt.Fprintln(stdout, "Verificação concluída.")
		} else {
			fmt.Fprintln(stdout, "Preparação e verificação concluídas.")
		}
		fmt.Fprintln(stdout, "Escolha Iniciar ambiente para abrir NVDA e VS Code.")
		if w.finished(chosen) {
			return w.status
		}
	}
}

func (w *wizard) applicationSummary() string {
	names := make([]string, 0, len(w.deps.Manifest.Components))
	var bytes int64
	for _, component := range w.deps.Manifest.Components {
		names = append(names, component.Name)
		bytes += component.DownloadBytes
	}
	return fmt.Sprintf("%s.\nPython, Node.js, uv e pnpm também serão instalados.\nDownload dos aplicativos: %.0f MiB, mais as ferramentas de desenvolvimento.\nCada aplicativo mantém sua licença.", strings.Join(names, ", "), float64(bytes)/(1<<20))
}

func (w *wizard) showLicenses() {
	fmt.Fprintln(w.output, "\nAplicativos e licenças")
	for _, component := range w.deps.Manifest.Components {
		fmt.Fprintf(w.output, "%s %s — %s\n%s\n", component.Name, component.Version, component.License, component.LicenseURL)
	}
	fmt.Fprintln(w.output, "As fontes e licenças das ferramentas de desenvolvimento estão em docs/FONTES-E-LICENCAS.md.")
}

func (w *wizard) perform(chosen wizardTarget) error {
	// Logs are created only after both explicit consents and target revalidation.
	file, err := os.CreateTemp("", "blind-dev-setup-*.log")
	if err != nil {
		return fmt.Errorf("criar registro da operação: %w", err)
	}
	w.logPath = file.Name()
	defer file.Close()
	fmt.Fprintf(file, "blind-dev-setup %s\nDestino: %s\n", w.deps.Manifest.Release, chosen.path())
	return w.withProgress(func(report func(progress.Event), activity func(string)) error {
		if !chosen.exists() {
			if err := w.deps.Prepare(PrepareRequest{Target: chosen.path(), Cache: w.cache, Output: file, Report: report}); err != nil {
				return err
			}
		}
		activity("Conferindo arquivos e configurações...")
		if err := w.deps.Verify(chosen.path(), file); err != nil {
			return err
		}
		return file.Sync()
	})
}

func (w *wizard) showDetails() {
	if w.logPath == "" {
		fmt.Fprintln(w.output, "Nenhuma operação foi executada.")
		return
	}
	fmt.Fprintf(w.output, "\nRegistro completo: %s\n", w.logPath)
	fmt.Fprintln(w.output, "Para testar com NVDA, siga docs/VALIDACAO.md. Extensões Python: docs/PYTHON-COM-UV.md.")
	file, err := os.Open(w.logPath)
	if err != nil {
		fmt.Fprintln(w.output, "Não foi possível ler o registro.")
		return
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		fmt.Fprintln(w.output, "Não foi possível ler o registro.")
		return
	}
	const limit = 24 * 1024
	if info.Size() > limit {
		_, _ = file.Seek(-limit, io.SeekEnd)
		fmt.Fprintln(w.output, "Exibindo apenas o final do registro.")
	}
	content, err := io.ReadAll(io.LimitReader(file, limit))
	if err != nil {
		fmt.Fprintln(w.output, "Não foi possível ler o registro.")
		return
	}
	fmt.Fprintln(w.output, cleanWizardText(string(content)))
}

func (w *wizard) finished(chosen wizardTarget) bool {
	for {
		answer, ok := w.ask("O que deseja fazer agora?", "", []wizardOption{
			{"I", "Iniciar ambiente"},
			{"1", "Abrir pasta do ambiente"}, {"2", "Escolher outro destino"}, {"D", "Ver detalhes da operação"}, {"0", "Sair"},
		})
		if !ok || answer == "0" {
			return true
		}
		if answer == "2" {
			return false
		}
		if answer == "D" {
			w.showDetails()
			continue
		}
		if answer == "I" {
			fmt.Fprintln(w.output, "Iniciando ambiente...")
			if err := w.launch(chosen); err != nil {
				w.fail("Não foi possível iniciar o ambiente", err)
				continue
			}
			w.status = 0
			fmt.Fprintln(w.output, "Inicialização solicitada para NVDA e VS Code. Você pode sair do assistente.")
			continue
		}
		if w.deps.OpenEnvironment == nil {
			w.fail("Não foi possível abrir a pasta", errors.New("abertura indisponível"))
			continue
		}
		if err := w.deps.OpenEnvironment(filepath.Join(chosen.path(), target.DirectoryName)); err != nil {
			w.fail("Não foi possível abrir a pasta", err)
			continue
		}
		w.status = 0
		fmt.Fprintln(w.output, "Pasta do ambiente aberta. Abra START.cmd para iniciar NVDA e VS Code.")
	}
}

func (w *wizard) launch(chosen wizardTarget) error {
	if w.deps.LaunchEnvironment == nil || w.deps.Verify == nil {
		return errors.New("inicialização indisponível nesta compilação")
	}
	// Preparation has now published the destination. Recheck its device identity
	// and contents before execution, including after a long stay in the menu.
	if chosen.details != nil {
		details := *chosen.details
		details.DestinationExists = true
		chosen.details = &details
	} else {
		chosen.candidate.DestinationExists = true
	}
	if err := w.revalidate(chosen); err != nil {
		return fmt.Errorf("conferir destino antes de iniciar: %w", err)
	}
	log, err := os.OpenFile(w.logPath, os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		return fmt.Errorf("abrir registro da inicialização: %w", err)
	}
	defer log.Close()
	if err := w.deps.Verify(chosen.path(), log); err != nil {
		return fmt.Errorf("verificar ambiente antes de iniciar: %w", err)
	}
	return w.deps.LaunchEnvironment(filepath.Join(chosen.path(), target.DirectoryName), log)
}
