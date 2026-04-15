package ui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/splashkes/chowdahh_recipes/examples/tui/api"
	"github.com/splashkes/chowdahh_recipes/examples/tui/ui/msg"
)

// Screen identifies the active screen.
type Screen = msg.Screen

const (
	ScreenAuth        = msg.ScreenAuth
	ScreenStreams      = msg.ScreenStreams
	ScreenCardList     = msg.ScreenCardList
	ScreenCardDetail   = msg.ScreenCardDetail
	ScreenReplay       = msg.ScreenReplay
	ScreenPreferences  = msg.ScreenPreferences
	ScreenSearch       = msg.ScreenSearch
)

// --- Navigation messages (re-exported) ---

type BackMsg = msg.BackMsg
type AuthSuccessMsg = msg.AuthSuccessMsg
type LogoutMsg = msg.LogoutMsg
type StreamSelectedMsg = msg.StreamSelectedMsg
type CardSelectedMsg = msg.CardSelectedMsg
type SearchSubmitMsg = msg.SearchSubmitMsg

// --- API result messages (re-exported) ---

type StreamLoadedMsg = msg.StreamLoadedMsg
type SearchResultsMsg = msg.SearchResultsMsg
type ReplayLoadedMsg = msg.ReplayLoadedMsg
type PreferencesLoadedMsg = msg.PreferencesLoadedMsg
type PreferencesSavedMsg = msg.PreferencesSavedMsg
type SignalSentMsg = msg.SignalSentMsg
type FlashMsg = msg.FlashMsg
type ClipboardMsg = msg.ClipboardMsg
type URLOpenedMsg = msg.URLOpenedMsg

// --- Commands (delegated) ---

func FetchStream(client *api.Client, slug string, limit int, cursor string) tea.Cmd {
	return msg.FetchStream(client, slug, limit, cursor)
}

func DoSearch(client *api.Client, query string, limit int) tea.Cmd {
	return msg.DoSearch(client, query, limit)
}

func FetchReplay(client *api.Client, period string, limit int) tea.Cmd {
	return msg.FetchReplay(client, period, limit)
}

func FetchPreferences(client *api.Client, personID string) tea.Cmd {
	return msg.FetchPreferences(client, personID)
}

func RecordSignal(client *api.Client, signalType, cardID string) tea.Cmd {
	return msg.RecordSignal(client, signalType, cardID)
}

func OpenURL(rawURL string) tea.Cmd {
	return msg.OpenURL(rawURL)
}

func CopyToClipboard(text string) tea.Cmd {
	return msg.CopyToClipboard(text)
}
