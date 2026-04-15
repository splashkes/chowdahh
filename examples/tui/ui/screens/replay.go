package screens

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/splashkes/chowdahh_recipes/examples/tui/api"
	umsg "github.com/splashkes/chowdahh_recipes/examples/tui/ui/msg"
	"github.com/splashkes/chowdahh_recipes/examples/tui/ui/style"
)

var periods = []string{"today", "last_7_days", "this_month"}

// Signal type icons
var signalIcons = map[string]string{
	"seen":        "👁",
	"open":        "📖",
	"save":        "⭐",
	"share":       "🔗",
	"dismiss":     "✕",
	"source_open": "🌐",
}

type ReplayItem struct {
	Event api.ReplayEvent
}

func (r ReplayItem) Title() string       { return r.Event.Headline }
func (r ReplayItem) Description() string { return r.Event.SignalType + " · " + r.Event.OccurredAt }
func (r ReplayItem) FilterValue() string { return r.Event.Headline }

type ReplayDelegate struct{}

func (d ReplayDelegate) Height() int                             { return 2 }
func (d ReplayDelegate) Spacing() int                            { return 0 }
func (d ReplayDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d ReplayDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	ri, ok := item.(ReplayItem)
	if !ok {
		return
	}

	selected := index == m.Index()
	icon := signalIcons[ri.Event.SignalType]
	if icon == "" {
		icon = "·"
	}

	cursor := "  "
	nameStyle := lipgloss.NewStyle().Foreground(style.ColorFg)
	if selected {
		cursor = style.Highlight.Render("> ")
		nameStyle = nameStyle.Bold(true).Foreground(style.ColorHighlight)
	}

	headline := ri.Event.Headline
	if headline == "" {
		headline = ri.Event.CardID
	}

	fmt.Fprintf(w, "%s%s %s\n%s%s",
		cursor, icon, nameStyle.Render(truncate(headline, m.Width()-8)),
		"    ", style.Dim.Render(ri.Event.SignalType+" · "+ri.Event.OccurredAt),
	)
}

type ReplayModel struct {
	list       list.Model
	loading    bool
	spinner    spinner.Model
	client     *api.Client
	errMsg     string
	periodIdx  int
	needsAuth  bool
	width      int
	height     int
}

func NewReplayModel(client *api.Client, width, height int) ReplayModel {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = style.Accent

	l := list.New(nil, ReplayDelegate{}, width, height-6)
	l.Title = "Replay — today"
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.Styles.Title = style.Title

	needsAuth := client.Token == ""

	return ReplayModel{
		list:      l,
		client:    client,
		spinner:   sp,
		loading:   !needsAuth,
		needsAuth: needsAuth,
		width:     width,
		height:    height,
	}
}

func (m ReplayModel) Init() tea.Cmd {
	if m.needsAuth {
		return nil
	}
	return tea.Batch(m.spinner.Tick, umsg.FetchReplay(m.client, periods[m.periodIdx], 50))
}

func (m ReplayModel) Update(msg tea.Msg) (ReplayModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(msg.Width, msg.Height-6)

	case umsg.ReplayLoadedMsg:
		m.loading = false
		if msg.Err != nil {
			m.errMsg = msg.Err.Error()
			return m, nil
		}
		if msg.Data != nil {
			items := make([]list.Item, len(msg.Data.Events))
			for i, e := range msg.Data.Events {
				items[i] = ReplayItem{Event: e}
			}
			m.list.SetItems(items)
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
		case "tab":
			m.periodIdx = (m.periodIdx + 1) % len(periods)
			period := periods[m.periodIdx]
			m.list.Title = "Replay — " + strings.ReplaceAll(period, "_", " ")
			m.loading = true
			return m, tea.Batch(m.spinner.Tick, umsg.FetchReplay(m.client, period, 50))
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m ReplayModel) View() string {
	if m.needsAuth {
		return "\n  " + style.Error.Render("Replay requires authentication.") +
			"\n\n  " + style.Dim.Render("Log in with a ch_person_* token to view history.") +
			"\n\n  " + style.Dim.Render("Press esc to go back")
	}
	if m.loading {
		return "\n  " + m.spinner.View() + " Loading replay…"
	}
	if m.errMsg != "" {
		return "\n  " + style.Error.Render("Error: "+m.errMsg) + "\n\n  Press esc to go back"
	}
	header := style.Dim.Render("  tab to switch period")
	return header + "\n" + m.list.View()
}

func truncate(s string, max int) string {
	if max <= 3 {
		return s
	}
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max-1]) + "…"
}
