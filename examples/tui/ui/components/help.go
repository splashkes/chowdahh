package components

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/splashkes/chowdahh_recipes/examples/tui/ui/style"
)

// HelpOverlay renders a centered help panel.
func HelpOverlay(width, height int) string {
	content := `  Navigation
  j/k, ↑/↓    move up / down
  enter, l     open selection
  esc, h       go back
  n            next page
  p            prev page

  Actions
  s            save card
  d            dismiss card
  o            open in browser
  c            copy share URL

  Screens
  /            search
  ctrl+r       replay history
  P            preferences

  Radio
  r            start radio
  space        pause / resume
  >            skip track
  x            stop radio

  System
  ctrl+l       logout
  q            quit
  ?            close this help`

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(style.ColorBorder).
		Padding(1, 3).
		Background(lipgloss.Color("#0d0d18")).
		Foreground(style.ColorFg).
		Width(44)

	title := style.Title.Render("  CHOWDAHH TUI")
	rendered := box.Render(title + "\n\n" + content)

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, rendered)
}
