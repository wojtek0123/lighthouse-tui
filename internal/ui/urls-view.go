package ui

import (
	"fmt"
	"io"
	"strings"

	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

const listHeight = 14

type styles struct {
	title        lipgloss.Style
	item         lipgloss.Style
	selectedItem lipgloss.Style
	pagination   lipgloss.Style
	help         lipgloss.Style
	quitText     lipgloss.Style
}

func newStyles(darkBG bool) styles {
	var s styles
	s.title = lipgloss.NewStyle().MarginLeft(2)
	s.item = lipgloss.NewStyle().PaddingLeft(4)
	s.selectedItem = lipgloss.NewStyle().PaddingLeft(2).Background(lipgloss.Color("247")).Width(58)
	s.pagination = list.DefaultStyles(darkBG).PaginationStyle.PaddingLeft(4)
	s.help = list.DefaultStyles(darkBG).HelpStyle.PaddingLeft(4).PaddingBottom(1)
	s.quitText = lipgloss.NewStyle().Margin(1, 0, 2, 4)
	return s
}

type RestartMsg struct {
	Url   string
	Index int
}

func restart(url string, index int) tea.Cmd {
	return func() tea.Msg {
		return RestartMsg{Url: url, Index: index}
	}
}

type Item struct {
	url     string
	loading bool
}

func (i Item) FilterValue() string { return "" }

type itemDelegate struct {
	spinner *spinner.Model
	styles  styles
}

func (d itemDelegate) Height() int                                  { return 1 }
func (d itemDelegate) Spacing() int                                 { return 0 }
func (d itemDelegate) Update(msg tea.Msg, list *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(Item)

	if !ok {
		return
	}

	spinner := ""

	if i.loading == true {
		spinner = d.spinner.View()
	}

	str := fmt.Sprintf("%s %s", i.url, spinner)

	fn := d.styles.item.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return d.styles.selectedItem.Render(strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}

type UrlsView struct {
	list    list.Model
	spinner *spinner.Model
}

func NewUrlsView() *UrlsView {
	items := []list.Item{}

	const defaultWidth = 20

	s := spinner.New()
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	spinner := &s

	list := list.New(items, itemDelegate{}, defaultWidth, listHeight)
	list.SetFilteringEnabled(false)
	list.SetShowStatusBar(false)
	list.SetShowHelp(false)
	list.SetDelegate(itemDelegate{styles: newStyles(false), spinner: spinner})

	return &UrlsView{list: list,
		spinner: spinner,
	}
}

func (v UrlsView) Init() tea.Cmd {
	return nil
}

func (v UrlsView) Update(msg tea.Msg) (UrlsView, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case spinner.TickMsg:

		var cmd tea.Cmd
		var newSpinner spinner.Model
		newSpinner, cmd = v.spinner.Update(msg)

		*v.spinner = newSpinner

		cmds = append(cmds, cmd)
	case tea.KeyPressMsg:
		switch msg.String() {
		case "r":
			item, ok := v.list.SelectedItem().(Item)

			if !ok {
				return v, nil
			}

			return v, restart(item.url, v.list.Index())
		case "d":
			i := v.list.Index()
			v.list.RemoveItem(i)
		}
	}

	var listCmd tea.Cmd
	v.list, listCmd = v.list.Update(msg)
	cmds = append(cmds, listCmd)

	return v, tea.Batch(cmds...)
}

func (v UrlsView) View() string {
	return "\n" + v.list.View()
}

func (v UrlsView) SetLoading(index int, loading bool) {
	item, ok := v.list.Items()[index].(Item)

	if !ok {
		return
	}

	item.loading = loading

	v.list.SetItem(index, item)
}

func (v UrlsView) IsLoading() bool {
	items := v.list.Items()

	for _, item := range items {
		i := item.(Item)

		if i.loading == true {
			return true
		}
	}

	return false

}

func (v UrlsView) StopLoading(index int) {
	items := v.list.Items()

	i, ok := items[index].(Item)

	if !ok {
		return
	}

	i.loading = !i.loading
	v.list.SetItem(index, i)
}
