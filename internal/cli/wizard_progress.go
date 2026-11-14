package cli

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/woliveiras/blind-dev-setup/internal/progress"
)

type wizardActivity string
type wizardProgressDone struct{}
type wizardProgressModel struct {
	activity string
	width    int
	done     bool
}

func (m wizardProgressModel) Init() tea.Cmd { return nil }
func (m wizardProgressModel) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch value := message.(type) {
	case wizardActivity:
		m.activity = string(value)
	case tea.WindowSizeMsg:
		m.width = value.Width
	case wizardProgressDone:
		m.done = true
		return m, tea.Quit
	}
	return m, nil
}
func (m wizardProgressModel) View() string {
	if m.done {
		return ""
	}
	width := m.width
	if width <= 0 {
		width = 76
	}
	if width > 76 {
		width = 76
	}
	if width < 24 {
		return ansi.Truncate("Preparando...", width, "") + "\n"
	}
	content := "Preparando seu ambiente\n\n" + m.activity + "\n\nCtrl+C para cancelar"
	return lipgloss.NewStyle().Width(width-6).Padding(1, 2).Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("39")).Render(content) + "\n"
}

func eventDescription(event progress.Event) string {
	var action string
	switch event.Stage {
	case progress.Download:
		action = "Baixando"
	case progress.Extract:
		action = "Extraindo"
	case progress.Configure:
		return "Criando configurações e atalhos..."
	case progress.Toolchain:
		return "Instalando Python, Node.js, uv e pnpm..."
	case progress.Publish:
		return "Finalizando o ambiente no pendrive..."
	default:
		return ""
	}
	component := strings.TrimSpace(cleanWizardText(event.Component))
	if event.Index > 0 && event.Total >= event.Index {
		return fmt.Sprintf("%s %s (%d de %d)...", action, component, event.Index, event.Total)
	}
	return fmt.Sprintf("%s %s...", action, component)
}

func (w *wizard) withProgress(operation func(func(progress.Event), func(string)) error) error {
	if w.accessible {
		activity := func(text string) { fmt.Fprintln(w.output, text) }
		return operation(func(event progress.Event) {
			if text := eventDescription(event); text != "" {
				activity(text)
			}
		}, activity)
	}
	program := tea.NewProgram(wizardProgressModel{activity: "Iniciando...", width: 76},
		tea.WithInput(nil), tea.WithOutput(w.output), tea.WithContext(w.ctx), tea.WithoutSignalHandler())
	done := make(chan error, 1)
	go func() { _, err := program.Run(); done <- err }()
	activity := func(text string) { program.Send(wizardActivity(text)) }
	err := operation(func(event progress.Event) {
		if text := eventDescription(event); text != "" {
			activity(text)
		}
	}, activity)
	program.Send(wizardProgressDone{})
	renderErr := <-done
	if err != nil {
		return err
	}
	return renderErr
}
