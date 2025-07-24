package statusbar

import (
	"github.com/brunoluiz/xpdig/internal/bubbles/component/navigator"
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	m.statusbar.FourthColumn = ""
	m.statusbar.FourthColumnColors = m.neutralColor

	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		cmd = m.onResize(msg)
	case navigator.EventItemCopied:
		m.statusbar.FourthColumn = "copied"
		m.statusbar.FourthColumnColors = m.secondaryColor
	}

	var statusbarCmd tea.Cmd
	m.statusbar, statusbarCmd = m.statusbar.Update(msg)

	return m, tea.Batch(cmd, statusbarCmd)
}

func (m *Model) onResize(msg tea.WindowSizeMsg) tea.Cmd {
	m.statusbar.Width = msg.Width
	return nil
}
