package cli

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/term"
)

type wizard struct {
	deps           Dependencies
	ctx            context.Context
	input          *wizardInput
	rawInput       io.Reader
	output         io.Writer
	accessible     bool
	explicitTarget string
	cache          string
	status         int
	logPath        string
}

type wizardLine struct {
	value string
	err   error
}

// huh v0.8 accessible forms cannot propagate EOF. The adapter supplies exactly
// one complete line per read and remembers termination; callers must check it
// after each form. No default or truncated token can become consent.
type wizardInput struct {
	ctx      context.Context
	requests chan chan wizardLine
	pending  []byte
	ended    error
}

type oneByteReader struct{ io.Reader }

func (r oneByteReader) Read(p []byte) (int, error) {
	if len(p) > 1 {
		p = p[:1]
	}
	return r.Reader.Read(p)
}

func newWizardInput(ctx context.Context, input io.Reader) *wizardInput {
	r := &wizardInput{ctx: ctx, requests: make(chan chan wizardLine)}
	go func() {
		// On-demand reads are essential: after the initial mode question the
		// visual renderer must have exclusive access to keyboard input.
		reader := bufio.NewReader(oneByteReader{input})
		for {
			select {
			case <-ctx.Done():
				return
			case response := <-r.requests:
				value, err := reader.ReadString('\n')
				response <- wizardLine{value, err}
				if err != nil {
					return
				}
			}
		}
	}()
	return r
}

func (r *wizardInput) next() (string, error) {
	if r.ended != nil {
		return "", r.ended
	}
	response := make(chan wizardLine, 1)
	select {
	case r.requests <- response:
	case <-r.ctx.Done():
		r.ended = r.ctx.Err()
		return "", r.ended
	}
	select {
	case line := <-response:
		if line.err != nil {
			r.ended = line.err
			return "", line.err
		}
		return line.value, nil
	case <-r.ctx.Done():
		r.ended = r.ctx.Err()
		return "", r.ended
	}
}

func (r *wizardInput) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	if len(r.pending) == 0 {
		line, err := r.next()
		if err != nil {
			return 0, err
		}
		r.pending = []byte(line)
	}
	n := copy(p, r.pending)
	r.pending = r.pending[n:]
	return n, nil
}

type wizardOption struct{ value, label string }

func terminalIO(input io.Reader, output io.Writer) bool {
	in, ok := input.(interface{ Fd() uintptr })
	if !ok {
		return false
	}
	out, ok := output.(interface{ Fd() uintptr })
	if !ok {
		return false
	}
	return term.IsTerminal(in.Fd()) && term.IsTerminal(out.Fd())
}

func (w *wizard) chooseMode(accessible, visual, terminal bool) bool {
	if accessible || !terminal {
		w.accessible = true
		if visual && !terminal {
			fmt.Fprintln(w.output, "Sem terminal interativo: usando o modo para leitor de tela.")
		}
		return true
	}
	if visual {
		w.accessible = false
		return true
	}
	fmt.Fprintln(w.output, "Como deseja iniciar?")
	fmt.Fprintln(w.output, "0. Sair")
	fmt.Fprintln(w.output, "1. Com leitor de tela/NVDA (padrão: Enter)")
	fmt.Fprintln(w.output, "2. Sem leitor de tela")
	for {
		line, err := w.input.next()
		if err != nil {
			w.endInput(err)
			return false
		}
		switch strings.TrimSpace(line) {
		case "", "1":
			w.accessible = true
			return true
		case "2":
			w.accessible = false
			return true
		case "0":
			return false
		default:
			fmt.Fprintln(w.output, "Escolha 1, 2 ou 0.")
		}
	}
}

func (w *wizard) step(number int, title string) {
	text := fmt.Sprintf("%d de 5 · %s", number, title)
	if w.accessible {
		fmt.Fprintf(w.output, "\n%s\n\n", text)
		return
	}
	style := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")).MarginTop(1).MarginBottom(1)
	fmt.Fprintln(w.output, style.Render(text))
}

func (w *wizard) form(field huh.Field) bool {
	form := huh.NewForm(huh.NewGroup(field)).
		WithProgramOptions(tea.WithoutSignalHandler(), tea.WithFilter(func(_ tea.Model, message tea.Msg) tea.Msg {
			if size, ok := message.(tea.WindowSizeMsg); ok && size.Width > 76 {
				size.Width = 76
				return size
			}
			return message
		})).
		WithAccessible(w.accessible).WithOutput(w.output).
		WithKeyMap(wizardKeyMap()).WithShowHelp(true).
		WithShowErrors(true)
	if w.accessible {
		form.WithInput(w.input).WithTheme(&huh.Theme{})
	} else {
		theme := huh.ThemeCharm()
		theme.Focused.Base = theme.Focused.Base.Padding(1, 2)
		form.WithInput(w.rawInput).WithTheme(theme)
	}
	err := form.RunWithContext(w.ctx)
	if w.ctx.Err() != nil {
		w.endInput(w.ctx.Err())
		return false
	}
	if w.accessible && w.input.ended != nil {
		w.endInput(w.input.ended)
		return false
	}
	if err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			fmt.Fprintln(w.output, "Assistente cancelado.")
			return false
		}
		w.fail("Não foi possível ler sua escolha", err)
		return false
	}
	if !w.accessible && form.State != huh.StateCompleted {
		fmt.Fprintln(w.output, "Assistente cancelado.")
		return false
	}
	return true
}

func (w *wizard) ask(title, description string, options []wizardOption) (string, bool) {
	value := ""
	validate := func(input string) error {
		input = strings.ToUpper(strings.TrimSpace(input))
		for _, option := range options {
			if input == option.value {
				return nil
			}
		}
		return errors.New("Opção inválida. Digite uma das opções mostradas.")
	}
	if w.accessible {
		if description != "" {
			fmt.Fprintln(w.output, description)
		}
		for _, option := range options {
			fmt.Fprintf(w.output, "%s: %s\n", option.value, option.label)
		}
		field := huh.NewInput().Title(title).Value(&value).Validate(validate)
		if !w.form(field) {
			return "", false
		}
	} else {
		items := make([]huh.Option[string], len(options))
		for i, option := range options {
			items[i] = huh.NewOption(option.label, option.value)
		}
		field := huh.NewSelect[string]().Title(title).Description(description).Options(items...).Value(&value)
		if len(items) > 8 {
			field.Height(10)
		}
		if !w.form(field) {
			return "", false
		}
	}
	// Accessible huh can return the last invalid value on EOF. This validation
	// is intentionally repeated after the termination check.
	if validate(value) != nil {
		w.fail("Escolha não confirmada", errors.New("resposta inválida"))
		return "", false
	}
	return strings.ToUpper(strings.TrimSpace(value)), true
}

func wizardKeyMap() *huh.KeyMap {
	keys := huh.NewDefaultKeyMap()
	keys.Quit.SetHelp("ctrl+c", "cancelar")
	keys.Input.Submit.SetHelp("enter", "confirmar")
	keys.Input.Next.SetHelp("enter", "continuar")
	keys.Input.Prev.SetHelp("shift+tab", "voltar")
	keys.Input.AcceptSuggestion.Unbind()
	keys.Select.Up.SetHelp("↑", "acima")
	keys.Select.Down.SetHelp("↓", "abaixo")
	keys.Select.Next.SetHelp("enter", "selecionar")
	keys.Select.Submit.SetHelp("enter", "selecionar")
	keys.Select.Prev.SetHelp("shift+tab", "voltar")
	keys.Select.HalfPageUp.SetHelp("ctrl+u", "subir")
	keys.Select.HalfPageDown.SetHelp("ctrl+d", "descer")
	keys.Select.GotoTop.SetHelp("home", "primeiro")
	keys.Select.GotoBottom.SetHelp("end", "último")
	keys.Select.Filter.Unbind()
	keys.Select.SetFilter.Unbind()
	keys.Select.ClearFilter.Unbind()
	return keys
}

func (w *wizard) consent(title, description, token string, details func()) string {
	for {
		value := ""
		validate := func(input string) error {
			switch strings.ToUpper(strings.TrimSpace(input)) {
			case token, "V", "0", "NAO", "NÃO":
				return nil
			case "D":
				if details != nil {
					return nil
				}
			}
			return fmt.Errorf("Resposta inválida. Digite %s, V ou 0.", token)
		}
		help := fmt.Sprintf("Digite %s para continuar; V para voltar; 0 para sair.", token)
		if details != nil {
			help += " D: ver aplicativos e licenças."
		}
		if w.accessible {
			fmt.Fprintln(w.output, description)
			fmt.Fprintln(w.output, help)
		}
		field := huh.NewInput().Title(title).Description(description + "\n\n" + help).Value(&value).Validate(validate)
		if !w.form(field) {
			return "exit"
		}
		if validate(value) != nil {
			return "exit"
		}
		switch strings.ToUpper(strings.TrimSpace(value)) {
		case token:
			return "yes"
		case "D":
			details()
		case "0":
			return "exit"
		default:
			fmt.Fprintln(w.output, "Operação cancelada. Voltando à seleção.")
			return "back"
		}
	}
}

func (w *wizard) endInput(err error) {
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		w.status = 130
		fmt.Fprintln(w.output, "Assistente interrompido.")
		return
	}
	if err != io.EOF {
		w.status = 1
	}
	fmt.Fprintln(w.output, "Entrada encerrada. Nenhuma nova ação foi confirmada.")
}

func cleanWizardText(text string) string {
	return strings.Map(func(r rune) rune {
		if r == '\n' || r == '\t' {
			return r
		}
		if unicode.IsControl(r) {
			return -1
		}
		return r
	}, ansi.Strip(text))
}

func (w *wizard) fail(message string, err error) {
	w.status = 1
	if w.logPath != "" {
		if file, openErr := os.OpenFile(w.logPath, os.O_APPEND|os.O_WRONLY, 0); openErr == nil {
			fmt.Fprintf(file, "\n%s: %v\n", message, err)
			_ = file.Close()
		}
		fmt.Fprintf(w.output, "%s. Escolha D para ver detalhes.\n", message)
		return
	}
	cause := []rune(cleanWizardText(err.Error()))
	if len(cause) > 180 {
		cause = append(cause[:180], []rune("…")...)
	}
	fmt.Fprintf(w.output, "%s: %s\n", message, string(cause))
}
