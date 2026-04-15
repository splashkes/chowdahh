package components

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/splashkes/chowdahh_recipes/examples/tui/api"
	"github.com/splashkes/chowdahh_recipes/examples/tui/audio"
	"github.com/splashkes/chowdahh_recipes/examples/tui/ui/style"
)

// NowPlaying is a compact, portable radio player widget.
type NowPlaying struct {
	Track     *api.RadioTrack
	Player    *audio.Player
	Loading   bool
	Width     int
	Tick      int // incremented each RadioTick for animations
	TrackIdx  int
	TrackTotal int
}

// Visualizer bar characters, low to high energy.
var vizBars = []string{"▁", "▂", "▃", "▄", "▅", "▆", "▇", "█"}

const playerWidth = 50

func (np NowPlaying) View() string {
	innerW := playerWidth - 4 // border + padding

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(style.ColorBorder).
		Padding(0, 1).
		Width(playerWidth - 2)

	// Active playback — teal border
	if np.Track != nil && np.Player != nil && np.Player.IsActive() {
		return boxStyle.BorderForeground(style.ColorAccent).Render(np.activeView(innerW))
	}

	// Loading
	if np.Loading {
		header := np.headerLine(innerW)
		body := style.Accent.Render(" Starting radio…")
		return boxStyle.Render(header + "\n" + body)
	}

	// Idle — single line
	idle := style.Dim.Render("♫ Radio ") + style.Highlight.Render("[r]") + style.Dim.Render(" start")
	return boxStyle.Render(idle)
}

func (np NowPlaying) headerLine(innerW int) string {
	title := style.Accent.Render("♫ CHOWDAHH RADIO")
	pad := innerW - lipgloss.Width(title)
	if pad < 0 {
		pad = 0
	}
	return title + strings.Repeat(" ", pad)
}

func (np NowPlaying) activeView(innerW int) string {
	var lines []string

	// Line 1: header + track counter
	header := style.Accent.Render("♫ CHOWDAHH RADIO")
	counter := ""
	if np.TrackTotal > 0 {
		counter = style.Dim.Render(fmt.Sprintf("track %d/%d", np.TrackIdx+1, np.TrackTotal))
	}
	headerPad := innerW - lipgloss.Width(header) - lipgloss.Width(counter)
	if headerPad < 1 {
		headerPad = 1
	}
	lines = append(lines, header+strings.Repeat(" ", headerPad)+counter)

	// Line 2: scrolling title
	icon := style.TimeMinutes.Render(" ▶ ")
	if np.Player.IsPaused() {
		icon = style.Highlight.Render(" ⏸ ")
	}
	titleW := innerW - 5
	scrolled := np.scrollTitle(np.Track.Headline, titleW)
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(style.ColorFg)
	lines = append(lines, icon+titleStyle.Render(scrolled))

	// Line 3: progress bar + time
	pos := np.Player.Position()
	dur := np.Player.Duration()
	timeStr := formatDuration(pos)
	if dur > 0 {
		timeStr += " / " + formatDuration(dur)
	}
	timeRendered := style.Dim.Render(timeStr)
	barW := innerW - lipgloss.Width(timeRendered) - 4
	if barW < 10 {
		barW = 10
	}
	bar := np.progressBar(pos, dur, barW)
	lines = append(lines, "  "+bar+"  "+timeRendered)

	// Line 4: visualizer + controls
	vizW := 16
	viz := np.visualizer(vizW)
	controls := style.Dim.Render("  ⏸ space") +
		style.Dim.Render("  ⏭ >") +
		style.Dim.Render("  ⏹ x")
	lines = append(lines, "  "+viz+controls)

	return strings.Join(lines, "\n")
}

// scrollTitle scrolls the headline if it's wider than the available space.
func (np NowPlaying) scrollTitle(title string, maxW int) string {
	runes := []rune(title)
	if len(runes) <= maxW {
		return title
	}
	// Scroll speed: 1 char every 2 ticks
	padded := append(runes, []rune("   ·   ")...)
	padded = append(padded, runes...)
	offset := (np.Tick / 2) % len(padded)
	end := offset + maxW
	if end > len(padded) {
		end = len(padded)
	}
	return string(padded[offset:end])
}

// progressBar renders a filled/unfilled bar.
func (np NowPlaying) progressBar(pos, dur time.Duration, width int) string {
	if dur <= 0 {
		return style.Dim.Render(strings.Repeat("─", width))
	}
	ratio := float64(pos) / float64(dur)
	if ratio > 1 {
		ratio = 1
	}
	filled := int(ratio * float64(width))
	return style.Accent.Render(strings.Repeat("━", filled)) +
		style.Dim.Render(strings.Repeat("─", width-filled))
}

// visualizer renders animated bars that bounce with a pseudo-wave pattern.
func (np NowPlaying) visualizer(width int) string {
	if np.Player.IsPaused() {
		return style.Dim.Render(strings.Repeat("▁", width))
	}
	var b strings.Builder
	for i := 0; i < width; i++ {
		// Create a wave pattern using sin with offset per bar
		phase := float64(np.Tick)*0.4 + float64(i)*0.7
		val := (math.Sin(phase) + 1) / 2 // 0..1
		// Add some noise from a second wave
		val += (math.Sin(phase*2.3+1.7) + 1) / 4
		if val > 1 {
			val = 1
		}
		idx := int(val * float64(len(vizBars)-1))
		if idx >= len(vizBars) {
			idx = len(vizBars) - 1
		}
		b.WriteString(style.Accent.Render(vizBars[idx]))
	}
	return b.String()
}

func formatDuration(d time.Duration) string {
	m := int(d.Minutes())
	s := int(d.Seconds()) % 60
	return fmt.Sprintf("%d:%02d", m, s)
}
