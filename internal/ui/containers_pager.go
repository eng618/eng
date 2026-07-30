package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/eng618/eng/internal/ui/theme"
)

type pagerModel struct {
	viewport viewport.Model
	content  string
	ready    bool
}

func newPagerModel(content string) pagerModel {
	return pagerModel{content: content}
}

func (m pagerModel) Init() tea.Cmd {
	return nil
}

func (m pagerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		headerHeight := 2
		footerHeight := 2
		verticalMarginHeight := headerHeight + footerHeight

		if !m.ready {
			m.viewport = viewport.New(msg.Width, msg.Height-verticalMarginHeight)
			m.viewport.SetContent(m.content)
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - verticalMarginHeight
		}
	}

	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m pagerModel) View() string {
	if !m.ready {
		return "\n  Initializing viewport..."
	}

	header := theme.PrimaryText.Bold(true).Render(" Compose Status Inspection Viewport")
	footer := theme.MutedText.Render(
		fmt.Sprintf(
			" %3.f%%  •  Use j/k or arrow keys to scroll  •  Press 'q' or 'esc' to exit ",
			m.viewport.ScrollPercent()*100,
		),
	)

	return fmt.Sprintf("%s\n%s\n%s", header, m.viewport.View(), footer)
}

// RunContainersPager launches an interactive scrollable viewport displaying rendered container status content.
func RunContainersPager(content string) error {
	p := tea.NewProgram(
		newPagerModel(content),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("failed to launch containers viewport pager: %w", err)
	}

	return nil
}
