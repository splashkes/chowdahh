package components

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"

	"github.com/splashkes/chowdahh_recipes/examples/tui/api"
	"github.com/splashkes/chowdahh_recipes/examples/tui/ui/style"
)

// StatusBar renders auth mode, rate limit, breadcrumb, and help hint.
type StatusBar struct {
	AuthMode  string
	RateLimit *api.RateLimit
	Screen    string
	Flash     string
	Width     int
}

func NewStatusBar() StatusBar {
	return StatusBar{AuthMode: "anonymous"}
}

func (s StatusBar) UpdateFromClient(client *api.Client) StatusBar {
	rl, am := client.RateInfo()
	if rl != nil {
		s.RateLimit = rl
	}
	if am != "" {
		s.AuthMode = am
	}
	return s
}

func (s StatusBar) View() string {
	barStyle := style.StatusBar.Width(s.Width)

	auth := style.Dim.Render(fmt.Sprintf("[%s]", s.AuthMode))

	rate := ""
	if s.RateLimit != nil {
		rate = style.Dim.Render(fmt.Sprintf("%d/%d remaining", s.RateLimit.Remaining, s.RateLimit.Limit))
	}

	screen := style.Accent.Render(s.Screen)
	help := style.Dim.Render("? help")

	var flash string
	if s.Flash != "" {
		flash = style.Highlight.Render(s.Flash)
	}

	left := fmt.Sprintf(" %s  %s  %s", auth, rate, screen)
	right := fmt.Sprintf("%s  %s ", flash, help)

	gap := s.Width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 0 {
		gap = 1
	}
	spaces := ""
	for i := 0; i < gap; i++ {
		spaces += " "
	}

	return barStyle.Render(left + spaces + right)
}
