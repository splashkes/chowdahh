package components

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/splashkes/chowdahh_recipes/examples/tui/api"
	"github.com/splashkes/chowdahh_recipes/examples/tui/ui/style"
)

// CardItem wraps a Card to satisfy the list.Item interface.
type CardItem struct {
	Card api.Card
}

func (c CardItem) Title() string       { return c.Card.Headline }
func (c CardItem) Description() string { return c.Card.Summary }
func (c CardItem) FilterValue() string { return c.Card.Headline + " " + c.Card.Summary }

// CardDelegate renders card items in the list.
type CardDelegate struct{}

func (d CardDelegate) Height() int                             { return 4 }
func (d CardDelegate) Spacing() int                            { return 1 }
func (d CardDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d CardDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	ci, ok := item.(CardItem)
	if !ok {
		return
	}
	card := ci.Card
	width := m.Width() - 4

	selected := index == m.Index()

	// Line 1: headline + time (color-coded)
	headlineStyle := lipgloss.NewStyle().Foreground(style.ColorFg)
	if selected {
		headlineStyle = headlineStyle.Bold(true).Foreground(style.ColorHighlight)
	}
	timeStr := StyledTimeAgo(card.LatestSourceAt)
	headlineMax := width - 20 // reserve space for time
	headline := Truncate(card.Headline, headlineMax)
	line1 := headlineStyle.Render(headline) + "  " + timeStr

	// Line 2: topics + sources
	var chips []string
	for i, t := range card.Topics {
		if i >= 3 {
			chips = append(chips, style.Dim.Render(fmt.Sprintf("+%d", len(card.Topics)-3)))
			break
		}
		chips = append(chips, style.TopicChip.Render(t))
	}
	metaParts := fmt.Sprintf("  %d sources", card.SourceCount)
	if card.HasImage() {
		metaParts += "  📷"
	}
	meta := style.Dim.Render(metaParts)
	line2 := strings.Join(chips, " ") + meta

	// Line 3: share link — always show
	link := card.ShareLink()
	var line3 string
	if link != "" {
		line3 = style.ShareLabel.Render("share: ") + style.ShareURL.Render(link)
	} else {
		line3 = style.Dim.Render(Truncate(card.Summary, width))
	}

	cursor := "  "
	if selected {
		cursor = style.Highlight.Render("> ")
	}

	fmt.Fprintf(w, "%s%s\n%s  %s\n%s  %s\n", cursor, line1, "  ", line2, "  ", line3)
}

// Truncate shortens a string to max runes, adding an ellipsis if needed.
func Truncate(s string, max int) string {
	if max <= 3 {
		return s
	}
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max-1]) + "…"
}

// StyledTimeAgo returns a color-coded time string.
func StyledTimeAgo(dateStr string) string {
	if dateStr == "" {
		return ""
	}
	t, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		t, err = time.Parse(time.RFC3339Nano, dateStr)
		if err != nil {
			return ""
		}
	}
	d := time.Since(t)
	switch {
	case d < 2*time.Minute:
		return style.TimeJustNow.Render("just now")
	case d < time.Hour:
		return style.TimeMinutes.Render(fmt.Sprintf("%dm ago", int(d.Minutes())))
	case d < 6*time.Hour:
		return style.TimeHours.Render(fmt.Sprintf("%dh ago", int(d.Hours())))
	case d < 24*time.Hour:
		return style.TimeDays.Render(fmt.Sprintf("%dh ago", int(d.Hours())))
	case d < 7*24*time.Hour:
		return style.TimeDays.Render(fmt.Sprintf("%dd ago", int(d.Hours()/24)))
	default:
		return style.TimeOld.Render(fmt.Sprintf("%dd ago", int(d.Hours()/24)))
	}
}

// PlainTimeAgo returns an unstyled time string (for use in other contexts).
func PlainTimeAgo(dateStr string) string {
	if dateStr == "" {
		return ""
	}
	t, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		t, err = time.Parse(time.RFC3339Nano, dateStr)
		if err != nil {
			return ""
		}
	}
	d := time.Since(t)
	switch {
	case d < 2*time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	}
}
