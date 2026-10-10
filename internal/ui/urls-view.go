package ui

import (
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"

	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/wojtek0123/lighthouse-tui/internal/lighthouse"
)

type UrlSelectedMsg struct {
	hostname string
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

type UrlItem struct {
	hostname string
	loading  bool

	cmd *runningCmd
}

func (i UrlItem) FilterValue() string { return "" }

type urlItemDelegate struct {
	spinner *spinner.Model
	styles  styles
	focus   bool
}

func (d urlItemDelegate) Height() int                                  { return 1 }
func (d urlItemDelegate) Spacing() int                                 { return 0 }
func (d urlItemDelegate) Update(msg tea.Msg, list *list.Model) tea.Cmd { return nil }
func (d urlItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(UrlItem)

	if !ok {
		return
	}

	spinner := ""

	if i.loading == true {
		spinner = d.spinner.View()
	}

	str := fmt.Sprintf("%s %s", i.hostname, spinner)

	fn := d.styles.item.Render
	if index == m.Index() && d.focus {
		fn = func(s ...string) string {
			return d.styles.selectedItem.Render(strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}

type UrlsView struct {
	list     list.Model
	spinner  *spinner.Model
	delegate *urlItemDelegate

	active bool
}

func NewUrlsView(active bool) *UrlsView {
	urls, err := lighthouse.GetHostnames()

	if err != nil {
	}
	items := []list.Item{}

	for _, url := range urls {
		items = append(items, UrlItem{
			hostname: url,
			loading:  false,
		})
	}

	const defaultWidth = 20

	s := spinner.New()
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	spinner := &s
	delegate := &urlItemDelegate{styles: newStyles(false), focus: active, spinner: spinner}

	list := list.New(items, urlItemDelegate{}, defaultWidth, listHeight)
	list.SetFilteringEnabled(false)
	list.SetShowStatusBar(false)
	list.SetShowHelp(false)
	list.SetShowTitle(false)
	list.SetShowPagination(false)
	list.SetDelegate(delegate)

	return &UrlsView{list: list,
		spinner:  spinner,
		active:   active,
		delegate: delegate,
	}
}

func (v UrlsView) Update(msg tea.Msg) (UrlsView, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case ActivePaneMsg:
		if msg.pane != UrlsPane {
			return v, nil
		}

		if selectedItem, ok := v.list.SelectedItem().(UrlItem); ok {
			return v, selectUrl(selectedItem.hostname)
		}
	case spinner.TickMsg:

		var cmd tea.Cmd
		var newSpinner spinner.Model
		newSpinner, cmd = v.spinner.Update(msg)

		*v.spinner = newSpinner

		cmds = append(cmds, cmd)
	case tea.KeyPressMsg:
		switch msg.String() {
		case "r":
			item, ok := v.list.SelectedItem().(UrlItem)

			if !ok {
				return v, nil
			}

			return v, restart(item.hostname, v.list.Index())
		case "d":
			// TODO:
			// [ ] delete confirmation
			if selectedItem, ok := v.list.SelectedItem().(UrlItem); ok {

				err := lighthouse.RemoveHostname(selectedItem.hostname)

				if err != nil {
					return v, nil
				}

				i := v.list.Index()
				v.list.RemoveItem(i)

				return v, nil
			}
		case "c", "esc":
			item, ok := v.list.SelectedItem().(UrlItem)

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
		}
	}

	previousIndex := v.list.Index()

	if v.active {
		v.list, cmd = v.list.Update(msg)
		cmds = append(cmds, cmd)
	}

	if v.list.Index() != previousIndex {
		if selectedItem, ok := v.list.SelectedItem().(UrlItem); ok {
			cmds = append(cmds, selectUrl(selectedItem.hostname))
		}
	}

	return v, tea.Batch(cmds...)
}

func (v UrlsView) View(width, height int) string {
	v.list.SetHeight(height)
	v.list.SetWidth(width)

	return v.list.View()
}

func (v UrlsView) SetLoading(cmd *runningCmd, index int, loading bool) {
	item, ok := v.list.Items()[index].(UrlItem)

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
		i := item.(UrlItem)

		if i.loading == true {
			return true
		}
	}

	return false

}

func cancel(index int) tea.Cmd {
	return func() tea.Msg {
		return CancelMsg{Index: index}
	}
}

func selectUrl(hostname string) tea.Cmd {
	return func() tea.Msg {
		return UrlSelectedMsg{hostname: hostname}
	}
}

func (v UrlsView) GetSelectedItem() (UrlItem, error) {
	i, ok := v.list.SelectedItem().(UrlItem)

	if !ok {
		return UrlItem{}, errors.New("Error parsing item")
	}

	return i, nil
}

func AddUniqueUrl(l *list.Model, newItem UrlItem) tea.Cmd {
	for i, item := range l.Items() {
		if existingItem, ok := item.(UrlItem); ok {
			if existingItem.hostname == newItem.hostname {
				existingItem.cmd = newItem.cmd
				existingItem.loading = true

				return l.SetItem(i, existingItem)
			}
		}
	}

	return l.InsertItem(0, newItem)
}

func StopLoading(l *list.Model, text string) tea.Cmd {
	u, err := url.Parse(text)

	if err != nil {
		return nil
	}

	for index, item := range l.Items() {
		if i, ok := item.(UrlItem); ok {
			if u.Hostname() == i.hostname {

				if !ok {
					return nil
				}

				i.loading = !i.loading
				i.cmd = nil
				return l.SetItem(index, i)
			}
		}
	}

	return nil
}

func (v *UrlsView) SetActive(active bool) {
	v.active = active
	v.delegate.focus = active
}
