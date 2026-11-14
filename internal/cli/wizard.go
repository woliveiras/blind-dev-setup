package cli

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/woliveiras/blind-dev-setup/internal/target"
)

type wizard struct {
	deps           Dependencies
	lines          chan wizardLine
	ctx            context.Context
	output         io.Writer
	explicitTarget string
	cache          string
	status         int
}

type wizardLine struct {
	value string
	err   error
}
type wizardTarget struct {
	candidate target.Candidate
	details   *target.Details
}

func runWizard(args []string, stdout, stderr io.Writer, deps Dependencies) int {
	flags := flag.NewFlagSet("wizard", flag.ContinueOnError)
	flags.SetOutput(stderr)
	explicit := flags.String("target", "", "destino explícito; ainda exige seleção e confirmação")
	cache := flags.String("cache", "", "diretório opcional para downloads verificados")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if flags.NArg() != 0 {
		fmt.Fprintln(stderr, "Erro: argumento inesperado no assistente.")
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
	w := wizard{deps: deps, lines: make(chan wizardLine), ctx: ctx, output: stdout, explicitTarget: strings.TrimSpace(*explicit), cache: *cache}
	// One reader goroutine lets Ctrl+C interrupt a blocked terminal read. Closing
	// the caller-owned input is avoided; at most this one read can remain blocked
	// until the process exits or the caller provides input.
	go func() {
		reader := bufio.NewReader(input)
		for {
			line, err := reader.ReadString('\n')
			select {
			case w.lines <- wizardLine{line, err}:
			case <-ctx.Done():
				return
			}
			if err != nil {
				return
			}
		}
	}()
	fmt.Fprintf(stdout, "blind-dev-setup %s — assistente de preparação\n", deps.Manifest.Release)
	fmt.Fprintln(stdout, "Todas as etapas acontecem nesta janela. Digite a opção e pressione Enter.")
	fmt.Fprintln(stdout, "O programa não formata unidades, não escolhe por você e não sobrescreve instalações.")
	for {
		if w.ctx.Err() != nil {
			fmt.Fprintln(stdout, "Assistente interrompido.")
			return 130
		}
		choices, err := w.discover()
		if err != nil {
			w.fail("Não foi possível consultar os destinos", err)
		}
		if err == nil && len(choices) == 0 {
			fmt.Fprintln(stdout, "Nenhum pendrive foi encontrado. Conecte-o e escolha R para procurar novamente.")
		}
		for i, choice := range choices {
			fmt.Fprintf(stdout, "\nOpção %d\n", i+1)
			w.describe(choice)
			if issues := w.issues(choice); len(issues) != 0 {
				fmt.Fprintf(stdout, "Precisa de atenção: %s.\n", strings.Join(issues, "; "))
			} else if choice.exists() {
				fmt.Fprintln(stdout, "Já existe uma instalação. Esta opção permite verificá-la, sem sobrescrever.")
			}
		}
		fmt.Fprintln(stdout, "\nDigite o número do destino, R para procurar novamente ou 0 para sair.")
		answer, ok := w.read()
		if !ok || answer == "0" {
			return w.status
		}
		if answer == "R" {
			continue
		}
		number, err := strconv.Atoi(answer)
		if err != nil || number < 1 || number > len(choices) {
			fmt.Fprintln(stdout, "Opção inválida. Escolha um número mostrado, R ou 0.")
			continue
		}
		chosen := choices[number-1]
		if issues := w.issues(chosen); len(issues) != 0 {
			fmt.Fprintf(stdout, "Não é possível usar este destino: %s.\n", strings.Join(issues, "; "))
			fmt.Fprintln(stdout, "Use outro pendrive NTFS ou corrija o problema e procure novamente. Nenhuma alteração foi feita.")
			continue
		}
		if !chosen.exists() {
			runPlan([]string{"--target", chosen.path()}, stdout, stderr, deps.Manifest)
			fmt.Fprintln(stdout, "Licenças dos aplicativos de terceiros:")
			for _, component := range deps.Manifest.Components {
				fmt.Fprintf(stdout, "%s: %s\n%s\n", component.Name, component.License, component.LicenseURL)
			}
			action := w.confirm("Leia as licenças. Digite ACEITO para aceitá-las e continuar, V para voltar ou 0 para sair.", "ACEITO")
			if action == "exit" {
				return w.status
			}
			if action == "back" {
				continue
			}
		}
		fmt.Fprintln(stdout, "Confirme o destino escolhido:")
		w.describe(chosen)
		message := "Digite SIM para preparar este destino, V para voltar ou 0 para sair."
		if chosen.exists() {
			message = "Digite SIM para verificar esta instalação, V para voltar ou 0 para sair."
		}
		action := w.confirm(message, "SIM")
		if action == "exit" {
			return w.status
		}
		if action == "back" {
			continue
		}
		if w.ctx.Err() != nil {
			fmt.Fprintln(stdout, "Assistente interrompido.")
			return 130
		}
		if err := w.revalidate(chosen); err != nil {
			w.fail("O destino mudou ou não está mais disponível. Selecione-o novamente", err)
			continue
		}
		if !chosen.exists() {
			if deps.Prepare == nil {
				w.fail("Preparação indisponível", errors.New("esta compilação não oferece preparação"))
				continue
			}
			fmt.Fprintln(stdout, "Preparando o ambiente. Aguarde as mensagens de cada componente.")
			if err := deps.Prepare(PrepareRequest{Target: chosen.path(), Cache: w.cache, Output: stdout}); err != nil {
				w.fail("Não foi possível concluir a preparação", err)
				fmt.Fprintln(stdout, "Confira o erro acima. Você pode selecionar o destino novamente para tentar outra vez nesta janela.")
				continue
			}
		}
		if deps.Verify == nil {
			w.fail("Verificação indisponível", errors.New("esta compilação não oferece verificação"))
			continue
		}
		fmt.Fprintln(stdout, "Verificando a instalação...")
		if err := deps.Verify(chosen.path(), stdout); err != nil {
			w.fail("A instalação não passou na verificação", err)
			fmt.Fprintln(stdout, "O ambiente não foi aprovado. Selecione o destino novamente para verificar; nenhuma instalação será sobrescrita.")
			continue
		}
		w.status = 0
		if chosen.exists() {
			fmt.Fprintln(stdout, "Verificação concluída.")
		} else {
			fmt.Fprintln(stdout, "Preparação e verificação concluídas.")
		}
		fmt.Fprintln(stdout, "Para testar o ambiente com NVDA, siga docs/VALIDACAO.md.")
		fmt.Fprintln(stdout, "As extensões Python são opcionais; a instalação guiada está em docs/PYTHON-COM-UV.md.")
		if w.finished(chosen.path()) {
			return w.status
		}
	}
}

// EOF, including an unterminated token, cancels instead of converting truncated
// redirected input into consent.
func (w *wizard) read() (string, bool) {
	select {
	case <-w.ctx.Done():
		w.status = 130
		fmt.Fprintln(w.output, "Assistente interrompido.")
		return "", false
	case line := <-w.lines:
		if line.err != nil {
			if line.err != io.EOF {
				w.fail("Falha ao ler a resposta", line.err)
			}
			fmt.Fprintln(w.output, "Entrada encerrada. Assistente encerrado sem confirmar novas ações.")
			return "", false
		}
		return strings.ToUpper(strings.TrimSpace(line.value)), true
	}
}

func (w *wizard) confirm(message, consent string) string {
	for {
		fmt.Fprintln(w.output, message)
		answer, ok := w.read()
		if !ok || answer == "0" {
			return "exit"
		}
		if answer == consent {
			return "yes"
		}
		if answer == "V" || answer == "R" || answer == "NÃO" || answer == "NAO" {
			fmt.Fprintln(w.output, "Operação cancelada. Voltando à seleção.")
			return "back"
		}
		fmt.Fprintln(w.output, "Resposta inválida. Nenhuma ação foi confirmada.")
	}
}

func (w *wizard) fail(message string, err error) {
	w.status = 1
	fmt.Fprintf(w.output, "%s: %v\n", message, err)
}

func (w *wizard) discover() ([]wizardTarget, error) {
	if w.explicitTarget != "" {
		if w.deps.Inspect == nil {
			return nil, errors.New("inspeção de destino indisponível")
		}
		details, err := w.deps.Inspect(w.explicitTarget)
		if err != nil {
			return nil, err
		}
		return []wizardTarget{{details: &details}}, nil
	}
	if w.deps.ListTargets == nil {
		return nil, errors.New("procura de pendrives indisponível")
	}
	candidates, err := w.deps.ListTargets()
	if err != nil {
		return nil, err
	}
	choices := make([]wizardTarget, len(candidates))
	for i, candidate := range candidates {
		choices[i].candidate = candidate
	}
	return choices, nil
}

func (choice wizardTarget) path() string {
	if choice.details != nil {
		return choice.details.Path
	}
	return choice.candidate.RootPath()
}

func (choice wizardTarget) exists() bool {
	if choice.details != nil {
		return choice.details.DestinationExists
	}
	return choice.candidate.DestinationExists
}

func (w *wizard) describe(choice wizardTarget) {
	if choice.details != nil {
		fmt.Fprintf(w.output, "Destino informado explicitamente: %s\nVolume: %s\nSistema de arquivos: %s\nEspaço livre: %.1f GiB\n", choice.path(), choice.details.Volume, choice.details.FileSystem, bytesInGiB(choice.details.FreeBytes))
		fmt.Fprintln(w.output, "Nome do dispositivo e tamanho total não disponíveis nesta inspeção.")
		return
	}
	c := choice.candidate
	fmt.Fprintf(w.output, "Nome: %s\nLetra: %s\nNome do volume: %s\nTamanho total: %.1f GiB\nSistema de arquivos: %s\nEspaço livre: %.1f GiB\n", valueOr(c.Model, "não informado"), valueOr(c.RootPath(), "não atribuída"), valueOr(c.Label, "não informado"), bytesInGiB(c.SizeBytes), valueOr(c.FileSystem, "não informado"), bytesInGiB(c.FreeBytes))
}

func (w *wizard) issues(choice wizardTarget) []string {
	if choice.details != nil {
		details := *choice.details
		if choice.exists() {
			details.DestinationExists = false
			details.FreeBytes = uint64(w.deps.Manifest.MinimumFreeBytes)
		}
		if err := target.Validate(details, w.deps.Manifest.MinimumFreeBytes); err != nil {
			return []string{err.Error()}
		}
		return nil
	}
	c := choice.candidate
	minimum := w.deps.Manifest.MinimumFreeBytes
	if c.DestinationExists {
		minimum = 0
	}
	return c.PreparationIssues(minimum)
}

func (w *wizard) revalidate(chosen wizardTarget) error {
	fresh, err := w.discover()
	if err != nil {
		return err
	}
	for _, candidate := range fresh {
		if !strings.EqualFold(candidate.path(), chosen.path()) {
			continue
		}
		if chosen.details != nil {
			if candidate.details == nil || !strings.EqualFold(candidate.details.Volume, chosen.details.Volume) || !strings.EqualFold(candidate.details.SystemVolume, chosen.details.SystemVolume) || !strings.EqualFold(candidate.details.FileSystem, chosen.details.FileSystem) {
				return errors.New("a identificação do volume mudou")
			}
		} else {
			a, b := chosen.candidate, candidate.candidate
			if a.DiskNumber != b.DiskNumber || a.Model != b.Model || a.SizeBytes != b.SizeBytes || a.Label != b.Label || !strings.EqualFold(a.FileSystem, b.FileSystem) || a.IsSystem != b.IsSystem {
				return errors.New("o dispositivo ou volume associado à letra mudou")
			}
		}
		if candidate.exists() != chosen.exists() {
			return errors.New("a presença de uma instalação mudou")
		}
		if issues := w.issues(candidate); len(issues) > 0 {
			return errors.New(strings.Join(issues, "; "))
		}
		return nil
	}
	return errors.New("o destino selecionado não foi encontrado")
}

func (w *wizard) finished(path string) bool {
	for {
		fmt.Fprintln(w.output, "\n1: abrir pasta do ambiente.\n2: escolher outro destino.\n0: sair.")
		answer, ok := w.read()
		if !ok || answer == "0" {
			return true
		}
		if answer == "2" || answer == "R" {
			return false
		}
		if answer != "1" {
			fmt.Fprintln(w.output, "Opção inválida. Escolha 1, 2 ou 0.")
			continue
		}
		if w.deps.OpenEnvironment == nil {
			w.fail("Não foi possível abrir a pasta", errors.New("abertura indisponível nesta compilação"))
			continue
		}
		if err := w.deps.OpenEnvironment(filepath.Join(path, target.DirectoryName)); err != nil {
			w.fail("Não foi possível abrir a pasta", err)
			continue
		}
		w.status = 0
		fmt.Fprintln(w.output, "Pasta do ambiente aberta. Abra START.cmd para iniciar NVDA e VS Code.")
	}
}
