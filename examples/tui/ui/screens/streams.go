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

// 8-bit pixel art logo — bowl icon from brand assets + CHOW/DAHH text.
// Colors match the production pixel art bowl: gold steam, dark bowl, orange fill.

func renderLogo(_ int) string {
	// Text is 25 chars wide; bowl is indented to center over it
	// Bowl visual center ~9 chars wide, offset ~8 chars to center over 25-char text
	bowlLines := []struct {
		text  string
		color lipgloss.Color
	}{
		{"            ✦", lipgloss.Color("#e8c547")},
		{"         ╱  ╱  ╱", lipgloss.Color("#f59e0b")},
		{"        ╱  ╱  ╱", lipgloss.Color("#d97706")},
		{"       ▄██████████▄", lipgloss.Color("#1b3a4b")},
		{"      █▓▓▓▓▓▓▓▓▓▓▓▓█", lipgloss.Color("#e8c547")},
		{"       ▀██████████▀", lipgloss.Color("#1b3a4b")},
		{"         ▀▀▀▀▀▀▀▀", lipgloss.Color("#111122")},
	}

	textLines := []struct {
		text  string
		color lipgloss.Color
	}{
		{"  ▄▀▀▀ █  █  ▄▀▀▄  █   █", lipgloss.Color("#22d3ee")},
		{"  █    █▀▀█  █  █  █▄█▄█", lipgloss.Color("#2dd4bf")},
		{"  ▀▀▀▀ ▀  ▀  ▀▀▀    ▀ ▀", lipgloss.Color("#34d399")},
		{"  █▀▀▄  ▄▀▀▄  █  █  █  █", lipgloss.Color("#e8c547")},
		{"  █  █  █▀▀█  █▀▀█  █▀▀█", lipgloss.Color("#f59e0b")},
		{"  ▀▀▀   ▀  ▀  ▀  ▀  ▀  ▀", lipgloss.Color("#d97706")},
	}

	var lines []string
	for _, l := range bowlLines {
		lines = append(lines, lipgloss.NewStyle().Foreground(l.color).Bold(true).Render(l.text))
	}
	lines = append(lines, "")
	for _, l := range textLines {
		lines = append(lines, lipgloss.NewStyle().Foreground(l.color).Bold(true).Render(l.text))
	}
	return strings.Join(lines, "\n")
}

// Fallback if /api/v1/categories is unavailable.
var fallbackCategories = []api.Category{
	{Slug: "top", Label: "Top Stories"},
	{Slug: "latest", Label: "Latest"},
	{Slug: "science", Label: "Science"},
	{Slug: "world", Label: "World"},
	{Slug: "business", Label: "Business"},
	{Slug: "culture", Label: "Culture"},
	{Slug: "tech", Label: "Tech"},
	{Slug: "health", Label: "Health"},
	{Slug: "sports", Label: "Sports"},
	{Slug: "good-news", Label: "Good News"},
	{Slug: "local", Label: "Local"},
}

type StreamItem struct {
	Slug    string
	Label   string
	Count   int
	HasMore bool
	Loaded  bool
}

func (s StreamItem) Title() string       { return s.Label }
func (s StreamItem) Description() string { return "" }
func (s StreamItem) FilterValue() string { return s.Label + " " + s.Slug }

type StreamDelegate struct{}

func (d StreamDelegate) Height() int                             { return 1 }
func (d StreamDelegate) Spacing() int                            { return 0 }
func (d StreamDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d StreamDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	si, ok := item.(StreamItem)
	if !ok {
		return
	}

	selected := index == m.Index()

	cursor := "  "
	nameStyle := lipgloss.NewStyle().Foreground(style.ColorFg)
	if selected {
		cursor = style.Highlight.Render("> ")
		nameStyle = nameStyle.Bold(true).Foreground(style.ColorHighlight)
	}

	name := nameStyle.Render(si.Label)

	var count string
	if si.Loaded {
		if si.Count > 0 && si.HasMore {
			count = style.Accent.Render(fmt.Sprintf("  %d+", si.Count))
		} else if si.Count > 0 {
			count = style.Accent.Render(fmt.Sprintf("  %d", si.Count))
		} else {
			count = style.Dim.Render("  —")
		}
	} else {
		count = style.Dim.Render("  ·")
	}

	fmt.Fprintf(w, "%s%s%s", cursor, name, count)
}

// --- Messages ---

type CategoriesLoadedMsg struct {
	Categories []api.Category
	Err        error
}

type CategoryCountMsg struct {
	Slug    string
	Count   int
	HasMore bool
	Err     error
}

// --- Model ---

type StreamsModel struct {
	list    list.Model
	slugs   []string
	loading bool
	spinner spinner.Model
	client  *api.Client
	width   int
	height  int
}

const logoHeight = 16 // 7 bowl + 1 blank + 6 text + 2 padding
const bottomBarsHeight = 8 // now-playing widget (~5) + status bar (1) + padding

func NewStreamsModel(client *api.Client, width, height int) StreamsModel {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = style.Accent

	listHeight := height - logoHeight - bottomBarsHeight
	l := list.New(nil, StreamDelegate{}, width, listHeight)
	l.Title = ""
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.SetShowTitle(false)
	l.SetFilteringEnabled(true)

	return StreamsModel{
		list:    l,
		loading: true,
		spinner: sp,
		client:  client,
		width:   width,
		height:  height,
	}
}

func (m StreamsModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, fetchCategories(m.client))
}

func fetchCategories(client *api.Client) tea.Cmd {
	return func() tea.Msg {
		env, err := client.GetCategories()
		if err != nil {
			return CategoriesLoadedMsg{Categories: fallbackCategories}
		}
		cats := env.Data.Categories
		if len(cats) == 0 {
			return CategoriesLoadedMsg{Categories: fallbackCategories}
		}
		return CategoriesLoadedMsg{Categories: cats}
	}
}

func fetchCategoryCount(client *api.Client, slug string) tea.Cmd {
	return func() tea.Msg {
		env, err := client.GetStream(slug, 50, "")
		if err != nil {
			return CategoryCountMsg{Slug: slug, Err: err}
		}
		hasMore := env.Meta != nil && env.Meta.HasMore
		return CategoryCountMsg{Slug: slug, Count: env.Data.Count, HasMore: hasMore}
	}
}

func (m StreamsModel) Update(msg tea.Msg) (StreamsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(msg.Width, msg.Height-logoHeight-2)

	case CategoriesLoadedMsg:
		m.loading = false
		items := make([]list.Item, len(msg.Categories))
		m.slugs = make([]string, len(msg.Categories))
		for i, c := range msg.Categories {
			items[i] = StreamItem{Slug: c.Slug, Label: c.Label, Count: c.Count, Loaded: c.Count > 0}
			m.slugs[i] = c.Slug
		}
		m.list.SetItems(items)

		var cmds []tea.Cmd
		for _, c := range msg.Categories {
			if c.Count == 0 {
				cmds = append(cmds, fetchCategoryCount(m.client, c.Slug))
			}
		}
		return m, tea.Batch(cmds...)

	case CategoryCountMsg:
		if msg.Err != nil {
			m.updateItemCount(msg.Slug, 0, false)
		} else {
			m.updateItemCount(msg.Slug, msg.Count, msg.HasMore)
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
		case "enter", "l", "right":
			if item, ok := m.list.SelectedItem().(StreamItem); ok {
				return m, func() tea.Msg { return umsg.StreamSelectedMsg{Slug: item.Slug} }
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m *StreamsModel) updateItemCount(slug string, count int, hasMore bool) {
	items := m.list.Items()
	for i, item := range items {
		if si, ok := item.(StreamItem); ok && si.Slug == slug {
			si.Count = count
			si.HasMore = hasMore
			si.Loaded = true
			items[i] = si
			break
		}
	}
	m.list.SetItems(items)
}

func (m StreamsModel) View() string {
	logo := renderLogo(m.width)

	if m.loading {
		return logo + "\n\n  " + m.spinner.View() + " Loading categories…"
	}
	return logo + "\n\n" + m.list.View()
}
