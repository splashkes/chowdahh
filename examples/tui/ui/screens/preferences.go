package screens

import (
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/splashkes/chowdahh_recipes/examples/tui/api"
	umsg "github.com/splashkes/chowdahh_recipes/examples/tui/ui/msg"
	"github.com/splashkes/chowdahh_recipes/examples/tui/ui/style"
)

type PreferencesModel struct {
	followed  []string
	avoided   []string
	focus     int // 0=followed, 1=avoided
	cursor    int
	editing   bool
	input     textinput.Model
	client    *api.Client
	loading   bool
	spinner   spinner.Model
	flash     string
	errMsg    string
	needsAuth bool
	width     int
	height    int
}

func NewPreferencesModel(client *api.Client, width, height int) PreferencesModel {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = style.Accent

	ti := textinput.New()
	ti.Placeholder = "topic name"
	ti.Width = 30

	needsAuth := client.Token == ""

	return PreferencesModel{
		client:    client,
		spinner:   sp,
		input:     ti,
		loading:   !needsAuth,
		needsAuth: needsAuth,
		width:     width,
		height:    height,
	}
}

func (m PreferencesModel) Init() tea.Cmd {
	if m.needsAuth {
		return nil
	}
	// Use "me" as person ID — the API resolves it from the token
	return tea.Batch(m.spinner.Tick, umsg.FetchPreferences(m.client, "me"))
}

func (m PreferencesModel) Update(msg tea.Msg) (PreferencesModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case umsg.PreferencesLoadedMsg:
		m.loading = false
		if msg.Err != nil {
			m.errMsg = msg.Err.Error()
			return m, nil
		}
		if msg.Data != nil && msg.Data.SavedPreferences != nil {
			m.followed = msg.Data.SavedPreferences.TopicsFollowed
			m.avoided = msg.Data.SavedPreferences.TopicsAvoided
		}
		return m, nil

	case umsg.PreferencesSavedMsg:
		if msg.Err != nil {
			m.flash = "Save failed: " + msg.Err.Error()
		} else {
			m.flash = "Preferences saved!"
		}
		return m, nil

	case spinner.TickMsg:
		if m.loading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

	case tea.KeyMsg:
		if m.editing {
			switch msg.String() {
			case "enter":
				val := strings.TrimSpace(m.input.Value())
				if val != "" {
					if m.focus == 0 {
						m.followed = append(m.followed, val)
					} else {
						m.avoided = append(m.avoided, val)
					}
				}
				m.editing = false
				m.input.SetValue("")
				m.input.Blur()
				return m, nil
			case "esc":
				m.editing = false
				m.input.SetValue("")
				m.input.Blur()
				return m, nil
			}
			var cmd tea.Cmd
			m.input, cmd = m.input.Update(msg)
			return m, cmd
		}

		m.flash = ""
		switch msg.String() {
		case "esc", "h", "left":
			return m, func() tea.Msg { return umsg.BackMsg{} }
		case "tab":
			m.focus = (m.focus + 1) % 2
			m.cursor = 0
		case "j", "down":
			list := m.currentList()
			if m.cursor < len(list)-1 {
				m.cursor++
			}
		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
			}
		case "a":
			m.editing = true
			m.input.Focus()
			return m, textinput.Blink
		case "x", "backspace":
			if m.focus == 0 && m.cursor < len(m.followed) {
				m.followed = append(m.followed[:m.cursor], m.followed[m.cursor+1:]...)
				if m.cursor > 0 {
					m.cursor--
				}
			} else if m.focus == 1 && m.cursor < len(m.avoided) {
				m.avoided = append(m.avoided[:m.cursor], m.avoided[m.cursor+1:]...)
				if m.cursor > 0 {
					m.cursor--
				}
			}
		case "ctrl+s":
			return m, func() tea.Msg {
				_, err := m.client.SetPreferences("me", api.Preferences{
					TopicsFollowed: m.followed,
					TopicsAvoided:  m.avoided,
				})
				return umsg.PreferencesSavedMsg{Err: err}
			}
		}
	}
	return m, nil
}

func (m PreferencesModel) currentList() []string {
	if m.focus == 0 {
		return m.followed
	}
	return m.avoided
}

func (m PreferencesModel) View() string {
	if m.needsAuth {
		return "\n  " + style.Error.Render("Preferences require authentication.") +
			"\n\n  " + style.Dim.Render("Log in with a ch_person_* token.") +
			"\n\n  " + style.Dim.Render("Press esc to go back")
	}
	if m.loading {
		return "\n  " + m.spinner.View() + " Loading preferences…"
	}
	if m.errMsg != "" {
		return "\n  " + style.Error.Render("Error: "+m.errMsg) + "\n\n  Press esc to go back"
	}

	var b strings.Builder

	header := style.Dim.Render("  tab switch section  a add  x remove  ctrl+s save")
	if m.flash != "" {
		header += "  " + style.Highlight.Render(m.flash)
	}
	b.WriteString(header + "\n\n")

	// Followed
	followedTitle := "  Followed Topics"
	if m.focus == 0 {
		followedTitle = style.Highlight.Render(followedTitle)
	} else {
		followedTitle = style.Title.Render(followedTitle)
	}
	b.WriteString(followedTitle + "\n")
	b.WriteString(m.renderTopicList(m.followed, m.focus == 0))
	b.WriteString("\n")

	// Avoided
	avoidedTitle := "  Avoided Topics"
	if m.focus == 1 {
		avoidedTitle = style.Highlight.Render(avoidedTitle)
	} else {
		avoidedTitle = style.Title.Render(avoidedTitle)
	}
	b.WriteString(avoidedTitle + "\n")
	b.WriteString(m.renderTopicList(m.avoided, m.focus == 1))

	if m.editing {
		b.WriteString("\n  Add topic: " + m.input.View())
	}

	return b.String()
}

func (m PreferencesModel) renderTopicList(topics []string, active bool) string {
	if len(topics) == 0 {
		return "  " + style.Dim.Render("(none)") + "\n"
	}
	var b strings.Builder
	for i, t := range topics {
		cursor := "  "
		itemStyle := lipgloss.NewStyle().Foreground(style.ColorFg)
		if active && i == m.cursor {
			cursor = style.Highlight.Render("> ")
			itemStyle = itemStyle.Bold(true).Foreground(style.ColorHighlight)
		}
		b.WriteString("  " + cursor + itemStyle.Render(t) + "\n")
	}
	return b.String()
}
