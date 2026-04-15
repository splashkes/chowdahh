package screens

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/splashkes/chowdahh_recipes/examples/tui/api"
	"github.com/splashkes/chowdahh_recipes/examples/tui/ui/components"
	umsg "github.com/splashkes/chowdahh_recipes/examples/tui/ui/msg"
	"github.com/splashkes/chowdahh_recipes/examples/tui/ui/style"
)

type CardDetailModel struct {
	card       api.Card
	viewport   viewport.Model
	client     *api.Client
	flash      string
	showPanel  bool
	panelIdx   int
	panelTotal int // sources count + actions count
	width      int
	height     int
	ready      bool
}

var actionItems = []struct {
	Key   string
	Label string
}{
	{"c", "Copy share URL"},
	{"o", "Open in browser"},
	{"s", "Save card"},
	{"d", "Dismiss card"},
}

func NewCardDetailModel(client *api.Client, card api.Card, width, height int) CardDetailModel {
	vp := viewport.New(width, height-4)
	vp.SetContent(renderCardContent(card, width-4))

	m := CardDetailModel{
		card:     card,
		viewport: vp,
		client:   client,
		width:    width,
		height:   height,
		ready:    true,
	}
	return m
}

func (m CardDetailModel) Init() tea.Cmd {
	// Record "open" signal
	return umsg.RecordSignal(m.client, "open", m.card.ID)
}

func (m CardDetailModel) Update(msg tea.Msg) (CardDetailModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - 4
		m.viewport.SetContent(renderCardContent(m.card, msg.Width-4))

	case umsg.ClipboardMsg:
		if msg.Err != nil {
			m.flash = "Copy failed"
		} else {
			m.flash = "Copied!"
		}
		return m, nil

	case umsg.SignalSentMsg:
		return m, nil

	case tea.KeyMsg:
		m.flash = ""

		// Side panel is open (sources + actions)
		if m.showPanel {
			switch msg.String() {
			case "esc", "left", "h":
				m.showPanel = false
				return m, nil
			case "j", "down":
				if m.panelIdx < m.panelTotal-1 {
					m.panelIdx++
				}
				return m, nil
			case "k", "up":
				if m.panelIdx > 0 {
					m.panelIdx--
				}
				return m, nil
			case "enter", "o":
				return m.executePanelItem()
			}
			return m, nil
		}

		switch msg.String() {
		case "esc", "h", "left":
			return m, func() tea.Msg { return umsg.BackMsg{} }
		case "right":
			m.showPanel = true
			m.panelIdx = 0
			m.panelTotal = len(m.card.Sources) + len(actionItems)
			return m, nil
		case "s":
			m.flash = "Saved!"
			return m, umsg.RecordSignal(m.client, "save", m.card.ID)
		case "d":
			return m, tea.Batch(
				umsg.RecordSignal(m.client, "dismiss", m.card.ID),
				func() tea.Msg { return umsg.BackMsg{} },
			)
		case "o":
			link := m.card.ShareLink()
			if link != "" {
				return m, umsg.OpenURL(link)
			}
		case "c":
			link := m.card.ShareLink()
			if link != "" {
				return m, umsg.CopyToClipboard(link)
			}
		}
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m CardDetailModel) executePanelItem() (CardDetailModel, tea.Cmd) {
	srcCount := len(m.card.Sources)

	// Source item — open its URL
	if m.panelIdx < srcCount {
		src := m.card.Sources[m.panelIdx]
		if src.SourceURL != "" {
			return m, umsg.OpenURL(src.SourceURL)
		}
		return m, nil
	}

	// Action item
	actionIdx := m.panelIdx - srcCount
	if actionIdx >= 0 && actionIdx < len(actionItems) {
		m.showPanel = false
		key := actionItems[actionIdx].Key
		link := m.card.ShareLink()
		switch key {
		case "c":
			if link != "" {
				return m, umsg.CopyToClipboard(link)
			}
		case "o":
			if link != "" {
				return m, umsg.OpenURL(link)
			}
		case "s":
			m.flash = "Saved!"
			return m, umsg.RecordSignal(m.client, "save", m.card.ID)
		case "d":
			return m, tea.Batch(
				umsg.RecordSignal(m.client, "dismiss", m.card.ID),
				func() tea.Msg { return umsg.BackMsg{} },
			)
		}
	}
	return m, nil
}

func (m CardDetailModel) View() string {
	header := style.Dim.Render("  ← back  → sources & actions")
	if m.flash != "" {
		header += "  " + style.Highlight.Render(m.flash)
	}

	if m.showPanel {
		return header + "\n" + m.renderPanel()
	}

	return header + "\n" + m.viewport.View()
}

func (m CardDetailModel) renderPanel() string {
	var b strings.Builder
	srcCount := len(m.card.Sources)
	idx := 0

	// Sources section
	if srcCount > 0 {
		b.WriteString(style.Title.Render("  Sources") + "\n")
		b.WriteString(style.Dim.Render(fmt.Sprintf("  %d sources, %d domains", m.card.SourceCount, m.card.DomainCount)) + "\n\n")

		for i, src := range m.card.Sources {
			cursor := "  "
			titleStyle := lipgloss.NewStyle().Foreground(style.ColorFg)
			if idx == m.panelIdx {
				cursor = style.Highlight.Render("> ")
				titleStyle = titleStyle.Bold(true).Foreground(style.ColorHighlight)
			}

			title := src.Title
			if title == "" {
				title = src.SourceURL
			}
			_ = i

			line := cursor + titleStyle.Render(components.Truncate(title, m.width-6))
			b.WriteString(line + "\n")

			// Domain + time below
			meta := "    "
			if src.Domain != "" {
				meta += style.Accent.Render(src.Domain)
			}
			if src.PublishedAt != "" {
				meta += "  " + components.StyledTimeAgo(src.PublishedAt)
			}
			if src.SourceURL != "" && idx == m.panelIdx {
				meta += "  " + style.Dim.Render(components.Truncate(src.SourceURL, m.width-30))
			}
			b.WriteString(meta + "\n")

			idx++
		}
		b.WriteString("\n")
	} else {
		b.WriteString(style.Dim.Render(fmt.Sprintf("  %d sources (details not yet available)", m.card.SourceCount)) + "\n\n")
	}

	// Actions section
	b.WriteString(style.Title.Render("  Actions") + "\n")
	for _, item := range actionItems {
		cursor := "  "
		label := style.Dim.Render(fmt.Sprintf("[%s] %s", item.Key, item.Label))
		if idx == m.panelIdx {
			cursor = style.Highlight.Render("> ")
			label = style.Highlight.Render(fmt.Sprintf("[%s] %s", item.Key, item.Label))
		}
		b.WriteString(cursor + label + "\n")
		idx++
	}

	return b.String()
}

func renderCardContent(card api.Card, width int) string {
	var b strings.Builder
	w := min(width, 80)

	// Headline
	headline := style.Title.Bold(true).Width(w).Render(card.Headline)
	b.WriteString(headline)
	b.WriteString("\n\n")

	// Time — prominent, color-coded
	timeStr := components.StyledTimeAgo(card.LatestSourceAt)
	if timeStr != "" {
		b.WriteString("  " + timeStr)
		if card.CreatedAt != "" {
			created := components.PlainTimeAgo(card.CreatedAt)
			if created != "" {
				b.WriteString("  " + style.Dim.Render("(first seen "+created+")"))
			}
		}
		b.WriteString("\n\n")
	}

	// Share URL — prominent, boxed
	link := card.ShareLink()
	if link != "" {
		shareBox := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(style.ColorHighlight).
			Padding(0, 2)
		shareContent := style.ShareLabel.Render("SHARE  ") + style.ShareURL.Render(link)
		b.WriteString(shareBox.Render(shareContent))
		b.WriteString("  " + style.Dim.Render("[c] copy  [o] open"))
		b.WriteString("\n\n")
	}

	// Topics
	if len(card.Topics) > 0 {
		var chips []string
		for _, t := range card.Topics {
			chips = append(chips, style.TopicChip.Render(t))
		}
		b.WriteString(strings.Join(chips, " "))
		b.WriteString("\n\n")
	}

	// Meta line
	meta := fmt.Sprintf("%d sources", card.SourceCount)
	if card.DomainCount > 0 {
		meta += fmt.Sprintf("  %d domains", card.DomainCount)
	}
	if card.HasImage() {
		meta += "  " + style.Accent.Render("📷 photo")
	}
	b.WriteString(style.Dim.Render(meta))
	b.WriteString("\n\n")

	// Summary
	if card.Summary != "" {
		sumStyle := lipgloss.NewStyle().
			Foreground(style.ColorFg).
			Width(w)
		b.WriteString(sumStyle.Render(card.Summary))
		b.WriteString("\n\n")
	}

	// Lead text (full story)
	if card.LeadText != "" {
		b.WriteString(style.Dim.Render("─── Full Story ───"))
		b.WriteString("\n\n")
		ltStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#b0b0c0")).
			Width(w)
		b.WriteString(ltStyle.Render(card.LeadText))
		b.WriteString("\n")
	}

	return b.String()
}
