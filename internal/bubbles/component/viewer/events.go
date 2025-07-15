package viewer

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/samber/lo"
)

type EventQuit struct{}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		cmds = append(cmds, m.onKey(msg))
	case tea.WindowSizeMsg:
		cmds = append(cmds, m.onResize(msg))
	}

	switch m.searchMode {
	case searchModeInput:
		var searchCmd tea.Cmd
		m.searchInput, searchCmd = m.searchInput.Update(msg)
		cmds = append(cmds, searchCmd)
	case searchModeInit:
		m.searchMode = searchModeInput
	}

	// Handle keyboard and mouse events in the viewport
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *Model) onKey(msg tea.KeyMsg) tea.Cmd {
	//nolint
	switch {
	case key.Matches(msg, m.KeyMap.Quit):
		defer m.onSearchQuit()
		if key.Matches(msg, m.KeyMap.SearchQuit) && m.searchMode != searchModeOff {
			return nil
		}

		return m.cmdQuit
	case key.Matches(msg, m.KeyMap.Search):
		m.onSearchInit()
	case key.Matches(msg, m.KeyMap.SearchConfirm):
		m.onSearchConfirm()
	case key.Matches(msg, m.KeyMap.SearchNext):
		m.onSearchNext()
	case key.Matches(msg, m.KeyMap.SearchPrevious):
		m.onSearchPrevious()
	case key.Matches(msg, m.KeyMap.SearchQuit):
		m.onSearchQuit()
	}
	return nil
}

func (m *Model) setViewportHeight(h int) {
	headerHeight := lipgloss.Height(m.headerView())
	footerHeight := lipgloss.Height(m.footerView())
	if m.searchMode != searchModeOff {
		footerHeight += lipgloss.Height(m.searchInput.View())
	}
	verticalMarginHeight := headerHeight + footerHeight
	m.viewport.Height = h - verticalMarginHeight
}

func (m *Model) onResize(msg tea.WindowSizeMsg) tea.Cmd {
	m.height = msg.Height

	if !m.ready {
		// Since this program is using the full size of the viewport we
		// need to wait until we've received the window dimensions before
		// we can initialize the viewport. The initial dimensions come in
		// quickly, though asynchronously, which is why we wait for them
		// here.
		headerHeight := lipgloss.Height(m.headerView())
		m.viewport = viewport.New(msg.Width, 0)
		m.viewport.Style = m.Styles.Viewport
		m.viewport.YPosition = headerHeight
		m.viewport.KeyMap.PageUp.SetKeys(
			lo.Flatten([][]string{m.KeyMap.PageUp.Keys(), m.viewport.KeyMap.PageUp.Keys()})...)
		m.viewport.KeyMap.PageDown.SetKeys(
			lo.Flatten([][]string{m.KeyMap.PageDown.Keys(), m.viewport.KeyMap.PageDown.Keys()})...)
		m.viewport.SetContent(m.content)
		m.ready = true

		// This is only necessary for high performance rendering, which in
		// most cases you won't need.
		//
		// Render the viewport one line below the header.
		m.viewport.YPosition = headerHeight + 1
		m.setViewportHeight(msg.Height)
	} else {
		m.viewport.Width = msg.Width
		m.setViewportHeight(msg.Height)
		m.setContent(m.content)
	}

	return nil
}

func (m *Model) onSearchInit() {
	if m.searchMode != searchModeOff {
		m.onSearchQuit()
	}
	m.searchMode = searchModeInit
	m.searchInput.Focus()
	m.setViewportHeight(m.height)
}

func (m *Model) onSearchConfirm() {
	m.searchResult = m.searchInput.Value()
	m.searchInput.Blur()
	m.searchMode = searchModeFilter
	m.doSearch()
}

func (m *Model) onSearchQuit() {
	m.searchInput.Blur()
	m.searchMode = searchModeOff
	m.searchResult = ""
	m.searchInput.Reset()
	m.searchResultPos = []int{}
	m.viewport.SetContent(m.content)
	m.setViewportHeight(m.height)
}

func (m *Model) doSearch() {
	m.searchResultPos = []int{}
	if m.searchResult == "" {
		return
	}

	contentLower := strings.ToLower(m.content)
	searchLower := strings.ToLower(m.searchResult)
	offset := 0

	for {
		index := strings.Index(contentLower[offset:], searchLower)
		if index == -1 {
			break
		}
		m.searchResultPos = append(m.searchResultPos, offset+index)
		offset += index + len(searchLower)
	}

	m.updateViewportContent()
}

func (m *Model) updateViewportContent() {
	if m.searchResult == "" {
		m.viewport.SetContent(m.content)
		return
	}

	var highlightedContent strings.Builder
	lastIndex := 0
	for i, matchIndex := range m.searchResultPos {
		highlightedContent.WriteString(m.content[lastIndex:matchIndex])
		style := m.Styles.SearchItem

		// Highlight the current search result
		if i == m.searchCursor {
			style = m.Styles.SearchCurrentItem
		}
		highlightedContent.WriteString(style.Render(m.content[matchIndex : matchIndex+len(m.searchResult)]))
		lastIndex = matchIndex + len(m.searchResult)
	}
	highlightedContent.WriteString(m.content[lastIndex:])

	m.viewport.SetContent(highlightedContent.String())
}

func (m *Model) onSearchNext() {
	if len(m.searchResultPos) == 0 {
		return
	}

	m.searchCursor++
	if m.searchCursor >= len(m.searchResultPos) {
		m.searchCursor = 0 // Wrap around to the first result
	}

	// Adjust viewport to show the next result
	m.adjustViewportPosition()
	m.updateViewportContent()
}

func (m *Model) onSearchPrevious() {
	if len(m.searchResultPos) == 0 {
		return
	}

	// // If no selection exists, start at the end
	// if m.searchCursor == 0 {
	// 	m.searchCursor = len(m.searchResultPos) - 1
	// 	m.updateViewportContent()
	// 	return
	// }

	m.searchCursor--
	if m.searchCursor < 0 {
		m.searchCursor = len(m.searchResultPos) - 1 // Wrap around to the last result
	}

	// Adjust viewport to show the previous result
	m.adjustViewportPosition()
	m.updateViewportContent()
}

// adjustViewportPosition adjust the viewport position to show the selected search result
func (m *Model) adjustViewportPosition() {
	if len(m.searchResultPos) == 0 {
		return
	}

	selectedPosition := m.searchResultPos[m.searchCursor]

	// Calculate the line number of the selected position
	lineNumber := strings.Count(m.content[:selectedPosition], "\n")

	// Adjust the viewport to show the selected line
	m.viewport.SetYOffset(lineNumber)
}
