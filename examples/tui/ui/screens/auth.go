package screens

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/splashkes/chowdahh_recipes/examples/tui/api"
	umsg "github.com/splashkes/chowdahh_recipes/examples/tui/ui/msg"
	"github.com/splashkes/chowdahh_recipes/examples/tui/ui/style"
)

type AuthModel struct {
	input  textinput.Model
	err    string
	width  int
	height int
}

func NewAuthModel() AuthModel {
	ti := textinput.New()
	ti.Placeholder = "ch_person_xxxxxxxxxxxx"
	ti.Focus()
	ti.EchoMode = textinput.EchoPassword
	ti.Width = 40
	return AuthModel{input: ti}
}

func (m AuthModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m AuthModel) Update(msg tea.Msg) (AuthModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			token := m.input.Value()
			if token == "" {
				m.err = ""
				return m, func() tea.Msg { return umsg.AuthSuccessMsg{Token: ""} }
			}
			if !api.ValidateTokenFormat(token) {
				m.err = "Token must start with ch_person_ or ch_cur_"
				return m, nil
			}
			m.err = ""
			return m, func() tea.Msg { return umsg.AuthSuccessMsg{Token: token} }
		case "tab":
			// Skip auth — continue anonymous
			return m, func() tea.Msg { return umsg.AuthSuccessMsg{Token: ""} }
		}
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m AuthModel) View() string {
	logo := lipgloss.NewStyle().
		Bold(true).
		Foreground(style.ColorFg).
		MarginBottom(2).
		Render("C H O W D A H H")

	prompt := style.Dim.Render("Paste your person token to log in")
	input := m.input.View()

	skip := style.Dim.Render("Press Tab to skip (anonymous)")

	var errLine string
	if m.err != "" {
		errLine = "\n" + style.Error.Render(fmt.Sprintf("  %s", m.err))
	}

	block := lipgloss.JoinVertical(lipgloss.Center,
		logo,
		prompt,
		"",
		input,
		errLine,
		"",
		skip,
	)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, block)
}
