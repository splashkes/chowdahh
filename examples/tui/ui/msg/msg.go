package msg

import (
	"os/exec"
	"runtime"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/splashkes/chowdahh_recipes/examples/tui/api"
)

// Screen identifies the active screen.
type Screen int

const (
	ScreenAuth Screen = iota
	ScreenStreams
	ScreenCardList
	ScreenCardDetail
	ScreenReplay
	ScreenPreferences
	ScreenSearch
)

// --- Navigation messages ---

type BackMsg struct{}
type AuthSuccessMsg struct{ Token string }
type LogoutMsg struct{}
type StreamSelectedMsg struct{ Slug string }
type CardSelectedMsg struct{ Card api.Card }
type SearchSubmitMsg struct{ Query string }

// --- API result messages ---

type StreamLoadedMsg struct {
	Data *api.StreamData
	Meta *api.Meta
	Err  error
}

type SearchResultsMsg struct {
	Data *api.SearchResult
	Err  error
}

type ReplayLoadedMsg struct {
	Data *api.ReplayData
	Err  error
}

type PreferencesLoadedMsg struct {
	Data *api.PreferencesData
	Err  error
}

type PreferencesSavedMsg struct {
	Err error
}

type SignalSentMsg struct{}

type FlashMsg struct{ Text string }
type ClipboardMsg struct {
	Text string
	Err  error
}
type URLOpenedMsg struct{}

// --- Radio messages ---

type RadioStartedMsg struct {
	Data *api.RadioSessionData
	Err  error
}

type RadioTrackEndedMsg struct{}

type RadioTickMsg struct{} // periodic UI refresh for progress

type StartRadioMsg struct {
	Mode string
}

// --- Commands ---

// FetchStream loads cards from a stream.
func FetchStream(client *api.Client, slug string, limit int, cursor string) tea.Cmd {
	return func() tea.Msg {
		env, err := client.GetStream(slug, limit, cursor)
		if err != nil {
			return StreamLoadedMsg{Err: err}
		}
		return StreamLoadedMsg{Data: &env.Data, Meta: env.Meta}
	}
}

// DoSearch runs a search query.
func DoSearch(client *api.Client, query string, limit int) tea.Cmd {
	return func() tea.Msg {
		env, err := client.Search(query, limit)
		if err != nil {
			return SearchResultsMsg{Err: err}
		}
		return SearchResultsMsg{Data: &env.Data}
	}
}

// FetchReplay loads replay events.
func FetchReplay(client *api.Client, period string, limit int) tea.Cmd {
	return func() tea.Msg {
		env, err := client.GetReplay(period, "", limit, "")
		if err != nil {
			return ReplayLoadedMsg{Err: err}
		}
		return ReplayLoadedMsg{Data: &env.Data}
	}
}

// FetchPreferences loads the user's preferences.
func FetchPreferences(client *api.Client, personID string) tea.Cmd {
	return func() tea.Msg {
		env, err := client.GetPreferences(personID)
		if err != nil {
			return PreferencesLoadedMsg{Err: err}
		}
		return PreferencesLoadedMsg{Data: &env.Data}
	}
}

// RecordSignal sends a signal (fire-and-forget).
func RecordSignal(client *api.Client, signalType, cardID string) tea.Cmd {
	return func() tea.Msg {
		client.RecordSignals([]api.Signal{{
			SignalType: signalType,
			CardID:    cardID,
		}})
		return SignalSentMsg{}
	}
}

// StartRadioSession starts a radio session and returns tracks.
func StartRadioSession(client *api.Client, mode string, duration int) tea.Cmd {
	return func() tea.Msg {
		env, err := client.StartRadioSession(api.RadioStartPayload{
			Mode:            mode,
			DurationMinutes: duration,
		})
		if err != nil {
			return RadioStartedMsg{Err: err}
		}
		return RadioStartedMsg{Data: &env.Data}
	}
}

// WaitForTrackEnd blocks until the channel closes, then returns RadioTrackEndedMsg.
func WaitForTrackEnd(done chan struct{}) tea.Cmd {
	return func() tea.Msg {
		<-done
		return RadioTrackEndedMsg{}
	}
}

// RadioTick returns a tick command for UI refresh during playback.
func RadioTick() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(time.Time) tea.Msg {
		return RadioTickMsg{}
	})
}

// OpenURL opens a URL in the default browser.
func OpenURL(rawURL string) tea.Cmd {
	return func() tea.Msg {
		var cmd *exec.Cmd
		switch runtime.GOOS {
		case "darwin":
			cmd = exec.Command("open", rawURL)
		default:
			cmd = exec.Command("xdg-open", rawURL)
		}
		cmd.Run()
		return URLOpenedMsg{}
	}
}

// CopyToClipboard copies text to the system clipboard.
func CopyToClipboard(text string) tea.Cmd {
	return func() tea.Msg {
		var cmd *exec.Cmd
		switch runtime.GOOS {
		case "darwin":
			cmd = exec.Command("pbcopy")
		default:
			cmd = exec.Command("xclip", "-selection", "clipboard")
		}
		cmd.Stdin = strings.NewReader(text)
		if err := cmd.Run(); err != nil {
			return ClipboardMsg{Err: err}
		}
		return ClipboardMsg{Text: text}
	}
}
