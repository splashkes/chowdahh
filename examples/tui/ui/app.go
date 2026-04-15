package ui

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/splashkes/chowdahh_recipes/examples/tui/api"
	"github.com/splashkes/chowdahh_recipes/examples/tui/audio"
	"github.com/splashkes/chowdahh_recipes/examples/tui/ui/components"
	umsg "github.com/splashkes/chowdahh_recipes/examples/tui/ui/msg"
	"github.com/splashkes/chowdahh_recipes/examples/tui/ui/screens"
)

// App is the root Bubble Tea model.
type App struct {
	client *api.Client
	screen Screen
	prev   Screen

	// Sub-models
	auth    screens.AuthModel
	streams screens.StreamsModel
	cards   screens.CardListModel
	detail  screens.CardDetailModel
	replay  screens.ReplayModel
	prefs   screens.PreferencesModel
	search  screens.SearchModel

	// Radio state — persists across screens
	player       *audio.Player
	radioSession *api.RadioSessionData
	radioTracks  []api.RadioTrack
	radioIdx     int
	nowPlaying   components.NowPlaying

	statusBar components.StatusBar
	showHelp  bool

	width, height int
}

func NewApp(client *api.Client, initial Screen) App {
	player := audio.New()
	return App{
		client:    client,
		screen:    initial,
		auth:      screens.NewAuthModel(),
		streams:   screens.NewStreamsModel(client, 80, 24),
		statusBar: components.NewStatusBar(),
		player:    player,
		nowPlaying: components.NowPlaying{Player: player},
	}
}

func (a App) Init() tea.Cmd {
	switch a.screen {
	case ScreenAuth:
		return a.auth.Init()
	case ScreenStreams:
		return a.streams.Init()
	default:
		return nil
	}
}

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.statusBar.Width = msg.Width
		a.nowPlaying.Width = msg.Width
		return a.propagate(msg)

	case tea.KeyMsg:
		// Help overlay intercepts all keys
		if a.showHelp {
			a.showHelp = false
			return a, nil
		}

		// Global keys (not during auth or text input screens)
		if a.screen != ScreenAuth && a.screen != ScreenSearch {
			switch {
			case key.Matches(msg, DefaultKeys.Quit):
				a.player.Stop()
				return a, tea.Quit
			case key.Matches(msg, DefaultKeys.Help):
				a.showHelp = true
				return a, nil
			case key.Matches(msg, DefaultKeys.Logout):
				a.player.Stop()
				api.ClearToken()
				a.client.Token = ""
				a.screen = ScreenAuth
				a.auth = screens.NewAuthModel()
				a.auth.Update(tea.WindowSizeMsg{Width: a.width, Height: a.height - 1})
				return a, a.auth.Init()
			case key.Matches(msg, DefaultKeys.Search) && a.screen != ScreenCardDetail:
				a.prev = a.screen
				a.screen = ScreenSearch
				a.search = screens.NewSearchModel(a.client, a.width, a.height-1)
				return a, a.search.Init()
			case key.Matches(msg, DefaultKeys.Replay) && (a.screen == ScreenStreams || a.screen == ScreenCardList):
				a.prev = a.screen
				a.screen = ScreenReplay
				a.replay = screens.NewReplayModel(a.client, a.width, a.height-1)
				return a, a.replay.Init()
			case key.Matches(msg, DefaultKeys.Prefs) && (a.screen == ScreenStreams || a.screen == ScreenCardList):
				a.prev = a.screen
				a.screen = ScreenPreferences
				a.prefs = screens.NewPreferencesModel(a.client, a.width, a.height-1)
				return a, a.prefs.Init()
			}

			// Radio controls — global when radio is active
			if a.player.IsActive() {
				switch msg.String() {
				case " ":
					a.player.Pause()
					return a, nil
				case ">":
					return a, a.radioSkip()
				case "x":
					a.radioStop()
					return a, nil
				}
			}

			// Radio toggle
			if key.Matches(msg, DefaultKeys.Radio) && !a.player.IsActive() && !a.nowPlaying.Loading {
				a.nowPlaying.Loading = true
				return a, umsg.StartRadioSession(a.client, "briefing", 10)
			}
		}

		// Quit from auth/search with ctrl+c
		if (a.screen == ScreenAuth || a.screen == ScreenSearch) && msg.String() == "ctrl+c" {
			a.player.Stop()
			return a, tea.Quit
		}

	// Screen transitions
	case AuthSuccessMsg:
		if msg.Token != "" {
			api.SaveToken(msg.Token)
			a.client.Token = msg.Token
		}
		a.screen = ScreenStreams
		a.streams = screens.NewStreamsModel(a.client, a.width, a.height-1)
		return a, a.streams.Init()

	case StreamSelectedMsg:
		a.prev = ScreenStreams
		a.screen = ScreenCardList
		a.cards = screens.NewCardListModel(a.client, msg.Slug, a.width, a.height-1)
		return a, a.cards.Init()

	case CardSelectedMsg:
		a.prev = a.screen
		a.screen = ScreenCardDetail
		a.detail = screens.NewCardDetailModel(a.client, msg.Card, a.width, a.height-1)
		return a, a.detail.Init()

	case BackMsg:
		return a.goBack()

	// Radio messages
	case umsg.RadioStartedMsg:
		a.nowPlaying.Loading = false
		if msg.Err != nil {
			a.statusBar.Flash = "Radio failed: " + msg.Err.Error()
			return a, nil
		}
		return a, a.radioStart(msg.Data)

	case umsg.RadioTrackEndedMsg:
		return a, a.radioAdvance()

	case umsg.RadioTickMsg:
		if a.player.IsActive() {
			a.nowPlaying.Tick++
			return a, umsg.RadioTick()
		}

	case umsg.StartRadioMsg:
		return a, umsg.StartRadioSession(a.client, msg.Mode, 10)

	}

	return a.propagate(msg)
}

// --- Radio management ---

func (a *App) radioStart(data *api.RadioSessionData) tea.Cmd {
	a.radioSession = data
	a.radioTracks = data.Tracks
	a.radioIdx = 0

	// If no tracks with audio_url, build from queue IDs
	if len(a.radioTracks) == 0 && len(data.Queue) > 0 {
		for _, id := range data.Queue {
			a.radioTracks = append(a.radioTracks, api.RadioTrack{
				ID:       id,
				AudioURL: "/audio/" + id,
			})
		}
	}

	if len(a.radioTracks) == 0 {
		a.statusBar.Flash = "Radio: no tracks available"
		return nil
	}

	return a.radioPlayCurrent()
}

func (a *App) radioPlayCurrent() tea.Cmd {
	if a.radioIdx >= len(a.radioTracks) {
		a.radioStop()
		a.statusBar.Flash = "Radio: session complete"
		return nil
	}

	track := a.radioTracks[a.radioIdx]
	a.nowPlaying.Track = &track
	a.nowPlaying.TrackIdx = a.radioIdx
	a.nowPlaying.TrackTotal = len(a.radioTracks)
	a.nowPlaying.Tick = 0

	// Build full audio URL
	audioURL := track.AudioURL
	if len(audioURL) > 0 && audioURL[0] == '/' {
		audioURL = a.client.BaseURL + audioURL
	}

	done, err := a.player.PlayURL(audioURL)
	if err != nil {
		a.statusBar.Flash = "Audio error: " + err.Error()
		// Skip to next
		a.radioIdx++
		return a.radioPlayCurrent()
	}

	return tea.Batch(
		umsg.WaitForTrackEnd(done),
		umsg.RadioTick(),
	)
}

func (a App) radioSkip() tea.Cmd {
	a.player.Stop()
	a.radioIdx++
	return a.radioPlayCurrent()
}

func (a *App) radioAdvance() tea.Cmd {
	a.radioIdx++
	return a.radioPlayCurrent()
}

func (a *App) radioStop() {
	a.player.Stop()
	a.nowPlaying.Track = nil
	a.radioSession = nil
	a.radioTracks = nil
}

// --- Screen routing ---

func (a App) propagate(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch a.screen {
	case ScreenAuth:
		a.auth, cmd = a.auth.Update(msg)
	case ScreenStreams:
		a.streams, cmd = a.streams.Update(msg)
	case ScreenCardList:
		a.cards, cmd = a.cards.Update(msg)
	case ScreenCardDetail:
		a.detail, cmd = a.detail.Update(msg)
	case ScreenReplay:
		a.replay, cmd = a.replay.Update(msg)
	case ScreenPreferences:
		a.prefs, cmd = a.prefs.Update(msg)
	case ScreenSearch:
		a.search, cmd = a.search.Update(msg)
	}
	return a, cmd
}

func (a App) goBack() (tea.Model, tea.Cmd) {
	switch a.screen {
	case ScreenCardDetail:
		if a.prev == ScreenSearch {
			a.screen = ScreenSearch
		} else {
			a.screen = ScreenCardList
		}
	case ScreenCardList, ScreenReplay, ScreenPreferences, ScreenSearch:
		a.screen = ScreenStreams
	default:
		a.screen = ScreenStreams
	}
	return a, nil
}

func (a App) View() string {
	// Always sync auth/rate-limit from client
	a.statusBar = a.statusBar.UpdateFromClient(a.client)

	var content string
	switch a.screen {
	case ScreenAuth:
		content = a.auth.View()
		if a.showHelp {
			return components.HelpOverlay(a.width, a.height)
		}
		return content + "\n" + a.nowPlaying.View()
	case ScreenStreams:
		a.statusBar.Screen = "streams"
		content = a.streams.View()
	case ScreenCardList:
		a.statusBar.Screen = "cards"
		content = a.cards.View()
	case ScreenCardDetail:
		a.statusBar.Screen = "card detail"
		content = a.detail.View()
	case ScreenReplay:
		a.statusBar.Screen = "replay"
		content = a.replay.View()
	case ScreenPreferences:
		a.statusBar.Screen = "preferences"
		content = a.prefs.View()
	case ScreenSearch:
		a.statusBar.Screen = "search"
		content = a.search.View()
	}

	if a.showHelp {
		return components.HelpOverlay(a.width, a.height)
	}

	// Build bottom bars: now-playing (if active) + status bar
	bars := a.statusBar.View()
	np := a.nowPlaying.View()
	if np != "" {
		bars = np + "\n" + bars
	}

	return content + "\n" + bars
}
