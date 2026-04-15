package screens

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/splashkes/chowdahh_recipes/examples/tui/api"
	"github.com/splashkes/chowdahh_recipes/examples/tui/ui/components"
	umsg "github.com/splashkes/chowdahh_recipes/examples/tui/ui/msg"
	"github.com/splashkes/chowdahh_recipes/examples/tui/ui/style"
)

type CardListModel struct {
	list    list.Model
	slug    string
	cards   []api.Card
	cursor  string
	hasMore bool
	loading bool
	spinner spinner.Model
	client  *api.Client
	errMsg  string
	width   int
	height  int
}

func NewCardListModel(client *api.Client, slug string, width, height int) CardListModel {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = style.Accent

	l := list.New(nil, components.CardDelegate{}, width, height-4)
	l.Title = slug
	l.SetShowStatusBar(true)
	l.SetShowHelp(false)
	l.Styles.Title = style.Title

	return CardListModel{
		list:    l,
		slug:    slug,
		client:  client,
		spinner: sp,
		loading: true,
		width:   width,
		height:  height,
	}
}

func (m CardListModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		umsg.FetchStream(m.client, m.slug, 20, ""),
	)
}

func (m CardListModel) Update(msg tea.Msg) (CardListModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(msg.Width, msg.Height-4)

	case umsg.StreamLoadedMsg:
		m.loading = false
		if msg.Err != nil {
			m.errMsg = msg.Err.Error()
			return m, nil
		}
		if msg.Data != nil {
			m.cards = append(m.cards, msg.Data.Items...)
			items := make([]list.Item, len(m.cards))
			for i, c := range m.cards {
				items[i] = components.CardItem{Card: c}
			}
			m.list.SetItems(items)
			m.list.Title = msg.Data.Stream
		}
		if msg.Meta != nil {
			m.cursor = msg.Meta.NextCursor
			m.hasMore = msg.Meta.HasMore
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
		case "esc", "h", "left":
			return m, func() tea.Msg { return umsg.BackMsg{} }
		case "enter", "l", "right":
			if item, ok := m.list.SelectedItem().(components.CardItem); ok {
				return m, func() tea.Msg { return umsg.CardSelectedMsg{Card: item.Card} }
			}
		case "n", "ctrl+f":
			if m.hasMore && !m.loading {
				m.loading = true
				return m, tea.Batch(
					m.spinner.Tick,
					umsg.FetchStream(m.client, m.slug, 20, m.cursor),
				)
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m CardListModel) View() string {
	if m.loading && len(m.cards) == 0 {
		return "\n  " + m.spinner.View() + " Loading " + m.slug + "…"
	}
	if m.errMsg != "" {
		return "\n  " + style.Error.Render("Error: "+m.errMsg) + "\n\n  Press esc to go back"
	}
	v := m.list.View()
	if m.loading {
		v += "\n  " + m.spinner.View() + " Loading more…"
	}
	return v
}
