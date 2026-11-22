package components

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

const (
	arrowRight = "→" // or "\u2192"
	arrowDown  = "↓" // or "\u2193"
)

type TableChildRow []string

type TableRow struct {
	expanded bool
	title    string
	chips    []string
	children []TableChildRow
}

type TableHeader struct {
	title string
	width int
}

type Table struct {
	headers []TableHeader
	rows    []TableRow

	cursor        int
	width, height int
}

func NewTable() Table {
	return Table{
		headers: []TableHeader{},
		rows:    []TableRow{},
		cursor:  0,
		width:   20,
		height:  20,
	}
}

func (t Table) Update(msg tea.Msg) (Table, tea.Cmd) {
	return t, nil
}

func (t Table) View() string {
	return ""
}

func (t Table) renderHeaders() string {
	headers := []string{}

	for _, h := range t.headers {
		headers = append(headers, lipgloss.NewStyle().Width(h.width).Height(1).Padding(0, 1).Render(h.title))
	}

	return lipgloss.JoinHorizontal(lipgloss.Left, headers...)
}

func (t Table) renderRow(i int) string {
	row := t.rows[i]

	parentRow := lipgloss.JoinHorizontal(lipgloss.Left,
		lipgloss.NewStyle().Width(1).Padding(0, 1).Render(rowIcon(row.expanded)),
		lipgloss.NewStyle().Height(1).Padding(0, 1).Render(row.title),
		lipgloss.NewStyle().Height(1).Render(renderChips(row.chips)...),
	)

	if row.expanded {
		children := t.renderChildren(row.children)

		return lipgloss.JoinVertical(lipgloss.Top, parentRow, children)
	}

	return parentRow
}

func rowIcon(expanded bool) string {
	if expanded {
		return arrowDown
	}

	return arrowRight
}

func (t Table) renderChildren(children []TableChildRow) string {
	widths := t.headerWidths()
	rows := []string{}

	for _, child := range children {
		cells := []string{""}

		for i, cell := range child {
			cells = append(cells, lipgloss.NewStyle().Width(widths[i]).Height(1).Padding(0, 1).Render(cell))
		}

		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Left, cells...))
	}

	return lipgloss.JoinVertical(lipgloss.Top, rows...)
}

func renderChips(chips []string) []string {
	view := make([]string, 0, len(chips))

	for _, chip := range chips {
		view = append(view, lipgloss.NewStyle().Height(1).Foreground(lipgloss.BrightBlack).Render(lipgloss.JoinHorizontal(lipgloss.Left, "∙ ", chip)))
	}

	return view
}

func (t Table) headerWidths() []int {
	widths := make([]int, 0, len(t.headers))

	for _, h := range t.headers {
		widths = append(widths, h.width)
	}

	return widths
}

func (t *Table) SetHeaders(h []TableHeader) {
	t.headers = h
}

func (t *Table) SetRows(r []TableRow) {
	t.rows = r

}

func (t *Table) SetRowChildren(i int, children []TableChildRow) {
	t.rows[i].children = children
}

// var idPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,127}$`)
//
// func validateID(id string) error {
// 	if !idPattern.MatchString(id) {
// 		return fmt.Errorf("%w: %q", ErrInvalidID, id)
// 	}
// 	return nil
// }
