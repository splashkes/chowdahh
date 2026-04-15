package screens

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/splashkes/chowdahh_recipes/examples/tui/api"
	"github.com/splashkes/chowdahh_recipes/examples/tui/ui/components"
	umsg "github.com/splashkes/chowdahh_recipes/examples/tui/ui/msg"
	"github.com/splashkes/chowdahh_recipes/examples/tui/ui/style"
)

type SearchModel struct {
	input    textinput.Model
	results  list.Model
	loading  bool
	spinner  spinner.Model
	client   *api.Client
	searched bool
	errMsg   string
	focused  int // 0=input, 1=results
	width    int
	height   int
}

func NewSearchModel(client *api.Client, width, height int) SearchModel {
	ti := textinput.New()
	ti.Placeholder = "Search stories…"
	ti.Focus()
	ti.Width = 40

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = style.Accent

	l := list.New(nil, components.CardDelegate{}, width, height-8)
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.SetShowTitle(false)

	return SearchModel{
		input:   ti,
		results: l,
		spinner: sp,
		client:  client,
		width:   width,
		height:  height,
	}
}

func (m SearchModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m SearchModel) Update(msg tea.Msg) (SearchModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.results.SetSize(msg.Width, msg.Height-8)

	case umsg.SearchResultsMsg:
		m.loading = false
		if msg.Err != nil {
			m.errMsg = msg.Err.Error()
			return m, nil
		}
		m.searched = true
		if msg.Data != nil {
			items := make([]list.Item, len(msg.Data.Cards))
			for i, c := range msg.Data.Cards {
				items[i] = components.CardItem{Card: c}
			}
			m.results.SetItems(items)
		}
		return m, nil

	case spinner.TickMsg:
		if m.loading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if m.focused == 1 {
				m.focused = 0
				m.input.Focus()
				return m, textinput.Blink
			}
			return m, func() tea.Msg { return umsg.BackMsg{} }
		case "enter":
			if m.focused == 0 {
				q := m.input.Value()
				if q == "" {
					return m, nil
				}
				m.loading = true
				m.focused = 1
				m.input.Blur()
				return m, tea.Batch(m.spinner.Tick, umsg.DoSearch(m.client, q, 20))
			}
			// In results — open card
			if item, ok := m.results.SelectedItem().(components.CardItem); ok {
				return m, func() tea.Msg { return umsg.CardSelectedMsg{Card: item.Card} }
			}
		case "/":
			if m.focused == 1 {
				m.focused = 0
				m.input.Focus()
				return m, textinput.Blink
			}
		}

		// Route to active component
		if m.focused == 0 {
			var cmd tea.Cmd
			m.input, cmd = m.input.Update(msg)
			return m, cmd
		}
	}

	if m.focused == 1 {
		var cmd tea.Cmd
		m.results, cmd = m.results.Update(msg)
		return m, cmd
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m SearchModel) View() string {
	header := "  " + style.Title.Render("Search") + "\n\n"
	header += "  " + m.input.View() + "\n\n"

	if m.loading {
		return header + "  " + m.spinner.View() + " Searching…"
	}
	if m.errMsg != "" {
		return header + "  " + style.Error.Render("Error: "+m.errMsg)
	}
	if !m.searched {
		return header + "  " + style.Dim.Render("Type a query and press Enter")
	}
	return header + m.results.View()
}
