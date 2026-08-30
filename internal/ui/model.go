package ui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func renderPane(title, content string, width, height int, isActive bool) string {
	borderColor := inactiveBorderColor
	titleColor := inactiveTitleColor
	if isActive {
		borderColor = activeBorderColor
		titleColor = activeTitleColor
	}

	border := lipgloss.RoundedBorder()
	borderStyle := lipgloss.NewStyle().Foreground(borderColor)

	// 1. Style and render the title
	titleStyle := lipgloss.NewStyle().
		Foreground(titleColor).
		Bold(isActive)
	renderedTitle := titleStyle.Render(" " + title + " ")

	// 2. Construct the custom top border line
	// Format: ╭─ Title ────────────╮
	titleLen := lipgloss.Width(renderedTitle)
	topBarLength := width - titleLen - 2 // -2 for the ╭ and ╮ corners

	// Ensure we don't panic on extremely small terminals
	if topBarLength < 0 {
		topBarLength = 0
	}

	leftDashCount := 1 // We want one dash before the title
	rightDashCount := topBarLength - leftDashCount
	if rightDashCount < 0 {
		rightDashCount = 0
	}

	topLeft := borderStyle.Render(border.TopLeft)
	topRight := borderStyle.Render(border.TopRight)
	horizLineLeft := borderStyle.Render(strings.Repeat(border.Top, leftDashCount))
	horizLineRight := borderStyle.Render(strings.Repeat(border.Top, rightDashCount))

	topBorder := topLeft + horizLineLeft + renderedTitle + horizLineRight + topRight

	// 3. Render the content box
	// We disable the top border because we just built it manually above
	contentStyle := lipgloss.NewStyle().
		Border(border).
		BorderTop(false).
		BorderForeground(borderColor).
		Width(width).
		Height(height - 1) // -1 because the manual top border takes up 1 line

	renderedContent := contentStyle.Render(content)

	// 4. Stack the top border and the content box
	return lipgloss.JoinVertical(lipgloss.Left, topBorder, renderedContent)
}

type App struct {
	// sidebar *Sidebar
	header *Header
	result *ResultView

	width  int
	height int
}

func InitialModel() App {
	header := NewHeader()
	resultView := NewResultView()

	return App{
		header: header,
		result: resultView,
	}
}

func (a App) Init() tea.Cmd {
	return a.header.urlInput.Focus()
}

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return a, tea.Quit
		}
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.header.SetWindowWidth(msg.Width)
	}

	header, cmd := a.header.Update(msg)
	*a.header = header

	cmds = append(cmds, cmd)

	a.result.audits = a.header.GetLighthouseResult()
	a.result.loaded = true

	return a, tea.Batch(cmds...)
}

func (a App) View() tea.View {
	v := tea.NewView(lipgloss.JoinVertical(lipgloss.Center, renderPane("Header", a.header.View().Content, a.width, 2, false), a.result.View().Content))

	v.AltScreen = true

	return v
}

func (a App) displaySize() string {
	return fmt.Sprintf("Width: %d, height: %d", a.width, a.height)
}

func (a App) setLayout() {}
