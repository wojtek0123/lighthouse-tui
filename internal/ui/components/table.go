package components

import (
	"math"
	"strconv"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

var (
	highlightRowStyle = lipgloss.NewStyle().Background(lipgloss.Black)
	cellStyle         = lipgloss.NewStyle().Padding(0, 1)
)

const (
	arrowRight = "→"
	arrowDown  = "↓"
)

type TableChildRow []string

type TableRow struct {
	expanded bool
	Values   []string
	Children []TableChildRow
}

type TableHeader struct {
	Title string
	Width int
}

type flatRow struct {
	parentIdx int
	isChild   bool
	parent    *TableRow
	child     TableChildRow
}

type Table struct {
	headers []TableHeader

	rows    []*TableRow
	flatten []flatRow
	visible []*flatRow

	cursor        int
	width, height int

	start, end, mid int
}

func NewTable() *Table {
	return &Table{
		headers: []TableHeader{},
		rows:    []*TableRow{},
		cursor:  0,
		width:   20,
		height:  20,
		start:   0,
		end:     20,
		mid:     10,
	}
}

func (t *Table) Update(msg tea.Msg) (*Table, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "enter", "space":
			if t.visible[t.cursor].isChild {
				t.selectChild(t.cursor)
			} else {
				t.Toggle(t.cursor)
			}
		case "j", "down":
			t.MoveUp()
		case "k", "up":
			t.MoveDown()
		case "G":
			t.MoveBottom()
		case "g":
			t.MoveTop()
		}
	}

	var cmd tea.Cmd

	return t, cmd
}

func (t Table) View() string {
	rows := make([]string, 0, min(t.height, len(t.visible)))

	for i := t.start; i < t.end; i++ {
		rows = append(rows, t.renderRowAt(i))
	}

	body := lipgloss.JoinVertical(lipgloss.Top, rows...)

	return lipgloss.JoinVertical(lipgloss.Top, strconv.Itoa(t.end), t.renderHeaders(), body)
}

func (t Table) renderHeaders() string {
	headers := make([]string, 1, len(t.headers))

	headers[0] = cellStyle.Render(" ")

	for _, h := range t.headers {
		headers = append(headers, cellStyle.Width(h.Width).Height(1).Padding(0, 1).Render(h.Title))
	}

	return lipgloss.JoinHorizontal(lipgloss.Left, headers...)
}

func (t Table) renderRowAt(i int) string {
	r := *t.visible[i]

	var row string
	if r.isChild {
		row = t.renderChildRow(r)
	} else {
		row = t.renderRow(r)
	}

	if i == t.cursor {
		row = highlightRowStyle.Render(row)
	}
	return row
}

func (t Table) renderRow(row flatRow) string {
	widths := t.headerWidths()
	cells := make([]string, 1, len(t.headers)+1)

	icon := arrowRight

	if row.parent.expanded {
		icon = arrowDown
	}

	cells[0] = cellStyle.Render(icon)

	values := padSlice(row.parent.Values, len(t.headers))

	for i, cell := range values {
		cells = append(cells, cellStyle.Width(widths[i]).Render(cell))
	}

	return lipgloss.JoinHorizontal(lipgloss.Left, cells...)
}

func (t Table) renderChildRow(row flatRow) string {
	widths := t.headerWidths()
	cells := make([]string, 1, len(t.headers)+1)

	cells[0] = lipgloss.NewStyle().Padding(0, 1).Render(" ")

	for i, cell := range row.child {
		cells = append(cells, cellStyle.Width(widths[i]).Render(cell))
	}

	return lipgloss.JoinHorizontal(lipgloss.Left, cells...)
}

func (t Table) headerWidths() []int {
	widths := make([]int, 0, len(t.headers))

	for _, h := range t.headers {
		widths = append(widths, h.Width)
	}

	return widths
}

func (t *Table) SetHeaders(h []TableHeader) {
	t.headers = h
}

func (t *Table) SetRows(r []TableRow) {
	t.rows = nil

	for i := 0; i < len(r); i++ {
		r[i].expanded = false
	}

	for _, row := range r {
		t.rows = append(t.rows, &row)
	}

	t.flattenRows()
}

func (t *Table) Toggle(i int) {
	t.rows[t.visible[i].parentIdx].expanded = !t.rows[t.visible[i].parentIdx].expanded
	t.UpdateVisible()
}

func (t *Table) SetWidth(w int) {
	t.width = w
}

func (t *Table) SetHeight(h int) {
	header := 1

	t.height = h - header
	t.mid = int(math.Round(float64(h-header) / 2))
	t.UpdateVisible()
}

func (t *Table) MoveUp() {
	if t.cursor == len(t.visible)-1 {
		return
	}

	t.cursor = t.cursor + 1
	t.UpdateViewport()
}

func (t *Table) MoveDown() {
	if t.cursor == 0 {
		return
	}

	t.cursor = t.cursor - 1
	t.UpdateVisible()
}

func (t *Table) flattenRows() {
	n := len(t.rows)
	for _, r := range t.rows {
		n += len(r.Children)
	}

	flatten := make([]flatRow, 0, n)
	for i, r := range t.rows {
		flatten = append(flatten, flatRow{parentIdx: i, parent: r, isChild: false})
		for _, c := range r.Children {
			flatten = append(flatten, flatRow{parentIdx: i, parent: r, child: c, isChild: true})
		}
	}

	t.flatten = flatten
	t.UpdateVisible()
}

func (t *Table) UpdateVisible() {
	t.visible = make([]*flatRow, 0, len(t.flatten))
	for _, f := range t.flatten {
		if !f.isChild || f.parent.expanded {
			t.visible = append(t.visible, &f)
		}
	}

	t.UpdateViewport()
}

func (t *Table) UpdateViewport() {
	count := len(t.visible)
	t.cursor = clamp(t.cursor, 0, count)

	if t.cursor >= t.mid && t.cursor <= count-t.mid {
		diff := t.cursor - t.mid

		t.start = clamp(diff, 0, t.cursor)
		t.end = min(count, t.height+diff)
	} else if t.cursor > count-t.mid && t.height < count {
		t.start = count - t.height
		t.end = count
	} else {
		t.start = clamp(t.cursor-t.mid, 0, t.cursor)
		t.end = clamp(t.cursor+t.mid, min(t.height, count), count)
	}

	if t.end > count {
		t.end = count
	}
}

func (t *Table) selectChild(i int) {
	// return	t.visible[i].child
}

func (t *Table) MoveTop() {
	t.cursor = 0
	t.UpdateVisible()
}

func (t *Table) MoveBottom() {
	t.cursor = len(t.visible) - 1
	t.UpdateVisible()
}

func clamp(v, low, high int) int {
	return min(max(v, low), high)
}

func padSlice(s []string, targetLen int) []string {
	for len(s) < targetLen {
		s = append(s, "")
	}
	return s
}

// var idPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,127}$`)
//
// func validateID(id string) error {
// 	if !idPattern.MatchString(id) {
// 		return fmt.Errorf("%w: %q", ErrInvalidID, id)
// 	}
// 	return nil
// }
