package statusbar

import tea "github.com/charmbracelet/bubbletea"

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	m.statusbar.FourthColumn = ""
	m.statusbar.FourthColumnColors = m.neutralColor

	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		cmd = m.onResize(msg)
	case tea.KeyMsg:
		cmd = m.onKey(msg)
	}

	var statusbarCmd tea.Cmd
	m.statusbar, statusbarCmd = m.statusbar.Update(msg)

	return m, tea.Batch(cmd, statusbarCmd)
}

func (m *Model) onResize(msg tea.WindowSizeMsg) tea.Cmd {
	m.statusbar.Width = msg.Width
	return nil
}

func (m *Model) onKey(msg tea.KeyMsg) tea.Cmd {
	//nolint // allow usage of switch
	switch msg.String() {
	case "y":
		m.statusbar.FourthColumn = "yanked"
		m.statusbar.FourthColumnColors = m.secondaryColor
	}
	return nil
}
