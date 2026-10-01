package ui

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/wojtek0123/lighthouse-tui/internal/lighthouse"
)

type SelectedUrlMsg struct {
	url string
}

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
	s.item = lipgloss.NewStyle().PaddingLeft(1)
	s.selectedItem = lipgloss.NewStyle().PaddingLeft(1).Background(lipgloss.Color("60")).Width(58)
	s.pagination = list.DefaultStyles(darkBG).PaginationStyle.PaddingLeft(4)
	s.help = list.DefaultStyles(darkBG).HelpStyle.PaddingLeft(4).PaddingBottom(1)
	s.quitText = lipgloss.NewStyle().Margin(1, 0, 2, 4)
	return s
}

type RestartMsg struct {
	Url   string
	Index int
}

type CancelMsg struct {
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

	cmd *runningCmd
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
	urls, err := lighthouse.GetUrls()

	if err != nil {
	}
	items := []list.Item{}

	for _, url := range urls {
		items = append(items, Item{
			url:     url,
			loading: false,
		})
	}

	const defaultWidth = 20

	s := spinner.New()
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	spinner := &s

	list := list.New(items, itemDelegate{}, defaultWidth, listHeight)
	list.SetFilteringEnabled(false)
	list.SetShowStatusBar(false)
	list.SetShowHelp(false)
	list.SetShowTitle(false)
	list.SetShowPagination(false)
	list.SetDelegate(itemDelegate{styles: newStyles(false), spinner: spinner})

	return &UrlsView{list: list,
		spinner: spinner,
	}
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
		case "c", "esc":
			item, ok := v.list.SelectedItem().(Item)

			if !ok {
				return v, nil
			}

			if !item.loading || item.cmd == nil {
				return v, nil
			}

			index := v.list.Index()

			item.cmd.cancel()
			item.loading = false
			item.cmd = nil
			v.list.SetItem(index, item)

			return v, nil
		case "j":
			i, ok := v.list.SelectedItem().(Item)

			if !ok {
				return v, nil
			}

			if v.list.Index() < len(v.list.Items())-1 {
				v.list.Select(v.list.Index() + 1)
			}

			return v, selectUrl(i.url)
		case "k":
			i, ok := v.list.SelectedItem().(Item)

			if !ok {
				return v, nil
			}

			if v.list.Index() != 0 {
				v.list.Select(v.list.Index() - 1)
			}

			return v, selectUrl(i.url)
		}
	}

	var listCmd tea.Cmd
	v.list, listCmd = v.list.Update(msg)
	cmds = append(cmds, listCmd)

	i, ok := v.list.SelectedItem().(Item)

	if ok {
		cmds = append(cmds, selectUrl(i.url))
	}

	return v, tea.Batch(cmds...)
}

func (v UrlsView) View() string {
	return "\n" + v.list.View()
}

func (v UrlsView) SetLoading(cmd *runningCmd, index int, loading bool) {
	item, ok := v.list.Items()[index].(Item)

	if !ok {
		return
	}

	item.loading = loading
	item.cmd = cmd

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
	i.cmd = nil
	v.list.SetItem(index, i)
}

func cancel(index int) tea.Cmd {
	return func() tea.Msg {
		return CancelMsg{Index: index}
	}
}

func selectUrl(url string) tea.Cmd {
	return func() tea.Msg {
		return SelectedUrlMsg{url: url}
	}
}

func (v UrlsView) GetSelectedItem() (Item, error) {
	i, ok := v.list.SelectedItem().(Item)

	if !ok {
		return Item{}, errors.New("Error parsing item")
	}

	return i, nil
}
