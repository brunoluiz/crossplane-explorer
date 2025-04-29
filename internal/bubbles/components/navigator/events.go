package navigator

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"

	tea "github.com/charmbracelet/bubbletea"
)

type EventItemCopied struct {
	ID   string
	Data any
}

type EventItemSelected struct {
	ID   string
	Data any
}

type EventItemFocused struct {
	ID   string
	Data any
}

type EventQuitted struct{}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		cmd = m.onResize(msg)
	case tea.KeyMsg:
		cmd = m.onKey(msg)
	}

	switch m.searchMode {
	case searchModeInput:
		var searchCmd tea.Cmd
		m.searchInput, searchCmd = m.searchInput.Update(msg)
		return m, searchCmd
	case searchModeInit:
		m.searchMode = searchModeInput
	}

	var tableCmd tea.Cmd
	m.table, tableCmd = m.table.Update(msg)

	return m, tea.Batch(cmd, tableCmd)
}

func (m *Model) onResize(msg tea.WindowSizeMsg) tea.Cmd {
	m.setSize(msg.Width, msg.Height)
	m.table.SetWidth(msg.Width)
	m.table.SetHeight(msg.Height)
	m.SetColumns(m.table.Columns())
	return nil
}

func (m *Model) onNavUp() tea.Cmd {
	m.cursor--
	if m.cursor < 0 {
		m.cursor = 0
	}
	m.doLoadTable()

	return func() tea.Msg {
		return EventItemFocused{ID: m.Current().ID, Data: m.Current().Data}
	}
}

func (m *Model) onNavDown() tea.Cmd {
	m.cursor++
	if m.cursor >= len(m.data) {
		m.cursor = len(m.data) - 1
	}
	m.doLoadTable()

	return func() tea.Msg {
		return EventItemFocused{ID: m.Current().ID, Data: m.Current().Data}
	}
}

func (m *Model) onSearch(msg tea.KeyMsg) tea.Cmd {
	switch {
	case key.Matches(msg, m.KeyMap.SearchConfirm):
		m.searchResult = m.searchInput.Value()
		m.searchInput.Blur()
		m.searchMode = searchModeFilter
		m.doSearch()
		m.doLoadTable()
	case key.Matches(msg, m.KeyMap.SearchQuit):
		m.searchInput.Blur()
		m.searchMode = searchModeOff
		m.searchResult = ""
		m.searchInput.Reset()
	}
	return nil
}

func (m *Model) doSearch() {
	searchTerm := strings.ToLower(m.searchInput.Value())
	m.searchResultPos = []int{}
	for i, v := range m.data {
		if strings.Contains(strings.ToLower(v.ID), searchTerm) {
			m.searchResultPos = append(m.searchResultPos, i)
		}
	}
	if len(m.searchResultPos) > 0 {
		m.searchCursor = 0
		m.cursor = m.searchResultPos[0]
		m.table.SetCursor(m.cursor)
	}
}

func (m *Model) onSearchInit() {
	m.searchMode = searchModeInit
	m.searchInput.Reset()
	m.searchInput.Focus()
	m.searchResult = ""
}

func (m *Model) onSearchQuit() {
	if m.searchMode == searchModeOff {
		return
	}
	m.searchInput.Blur()
	m.searchInput.Reset()
	m.searchMode = searchModeOff
	m.searchResult = ""
	m.searchCursor = 0
	m.searchResultPos = []int{}
	m.doLoadTable()
}

func (m *Model) onSearchNext() tea.Cmd {
	if len(m.searchResultPos) == 0 {
		return nil
	}

	if m.searchCursor <= m.cursor {
		m.searchCursor = m.cursor
	}

	m.searchCursor++
	if m.searchCursor >= len(m.searchResultPos) {
		m.searchCursor = 0 // Wrap around to the first result
	}

	m.cursor = m.searchResultPos[m.searchCursor]
	m.table.SetCursor(m.cursor)
	return func() tea.Msg {
		return EventItemFocused{ID: m.Current().ID, Data: m.Current().Data}
	}
}

func (m *Model) onSearchPrev() tea.Cmd {
	if len(m.searchResultPos) == 0 {
		return nil
	}

	if m.cursor <= m.searchCursor {
		m.searchCursor = m.cursor
	}

	m.searchCursor--
	if m.searchCursor < 0 {
		m.searchCursor = len(m.searchResultPos) - 1 // Wrap around to the last result
	}

	m.cursor = m.searchResultPos[m.searchCursor]
	m.table.SetCursor(m.cursor)
	return func() tea.Msg {
		return EventItemFocused{ID: m.Current().ID, Data: m.Current().Data}
	}
}

func (m *Model) onKey(msg tea.KeyMsg) tea.Cmd {
	if m.searchMode == searchModeInput {
		return m.onSearch(msg)
	}

	switch {
	case key.Matches(msg, m.KeyMap.Search):
		m.onSearchInit()
	case key.Matches(msg, m.KeyMap.Up):
		return m.onNavUp()
	case key.Matches(msg, m.KeyMap.Down):
		return m.onNavDown()
	case key.Matches(msg, m.KeyMap.Help):
		m.showHelp = !m.showHelp
		// m.Help.ShowAll = !m.Help.ShowAll
	case key.Matches(msg, m.KeyMap.Copy):
		return func() tea.Msg {
			return EventItemCopied{ID: m.Current().ID, Data: m.Current().Data}
		}
	case key.Matches(msg, m.KeyMap.SearchQuit):
		m.onSearchQuit()
	case key.Matches(msg, m.KeyMap.Show):
		return func() tea.Msg {
			return EventItemSelected{ID: m.Current().ID, Data: m.Current().Data}
		}
	case key.Matches(msg, m.KeyMap.Quit):
		return func() tea.Msg {
			return EventQuitted{}
		}
	case key.Matches(msg, m.KeyMap.SearchNext):
		return m.onSearchNext()
	case key.Matches(msg, m.KeyMap.SearchPrevious):
		return m.onSearchPrev()
	}
	return nil
}
