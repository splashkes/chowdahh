package style

import "github.com/charmbracelet/lipgloss"

// Colors derived from the Chowdahh web app CSS.
var (
	ColorBg        = lipgloss.Color("#07080c")
	ColorFg        = lipgloss.Color("#e0e0e0")
	ColorDim       = lipgloss.Color("#555566")
	ColorAccent    = lipgloss.Color("#2c5364")
	ColorHighlight = lipgloss.Color("#e8c547")
	ColorError     = lipgloss.Color("#e63946")
	ColorSuccess   = lipgloss.Color("#2d6a4f")
	ColorBorder    = lipgloss.Color("#1b3a4b")
	ColorChip      = lipgloss.Color("#3a1255")
	ColorBarBg     = lipgloss.Color("#111122")

	Title = lipgloss.NewStyle().Bold(true).Foreground(ColorFg)
	Dim   = lipgloss.NewStyle().Foreground(ColorDim)
	Error = lipgloss.NewStyle().Foreground(ColorError)

	Accent    = lipgloss.NewStyle().Foreground(ColorAccent)
	Highlight = lipgloss.NewStyle().Foreground(ColorHighlight).Bold(true)
	Success   = lipgloss.NewStyle().Foreground(ColorSuccess)

	ShareURL = lipgloss.NewStyle().
			Foreground(ColorHighlight).
			Bold(true).
			Underline(true)

	TopicChip = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#c0c0d0")).
			Background(ColorChip).
			Padding(0, 1)

	StatusBar = lipgloss.NewStyle().
			Background(ColorBarBg).
			Foreground(ColorDim).
			Padding(0, 1)

	Card = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorBorder).
		Padding(1, 2)

	// Time colors — brighter = more recent
	TimeJustNow = lipgloss.NewStyle().Foreground(lipgloss.Color("#4ade80")).Bold(true) // bright green
	TimeMinutes = lipgloss.NewStyle().Foreground(lipgloss.Color("#22d3ee"))            // cyan
	TimeHours   = lipgloss.NewStyle().Foreground(lipgloss.Color("#e8c547"))            // gold
	TimeDays    = lipgloss.NewStyle().Foreground(lipgloss.Color("#8888aa"))             // muted
	TimeOld     = lipgloss.NewStyle().Foreground(ColorDim)                             // dim

	// Share link label
	ShareLabel = lipgloss.NewStyle().Foreground(lipgloss.Color("#888899"))
)
