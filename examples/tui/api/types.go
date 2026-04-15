package api

import "fmt"

// Envelope wraps every successful API response.
type Envelope[T any] struct {
	Data     T         `json:"data"`
	Guidance *Guidance `json:"guidance,omitempty"`
	Meta     *Meta     `json:"meta,omitempty"`
}

type Guidance struct {
	StatusExplanation string        `json:"status_explanation"`
	NextBestActions   []Action      `json:"next_best_actions,omitempty"`
	AccountState      *AccountState `json:"account_state,omitempty"`
	CapabilityHints   []string      `json:"capability_hints,omitempty"`
}

type AccountState struct {
	AuthMode  string     `json:"auth_mode"`
	RateLimit *RateLimit `json:"rate_limit,omitempty"`
}

type RateLimit struct {
	Limit     int    `json:"limit"`
	Remaining int    `json:"remaining"`
	ResetAt   string `json:"reset_at"`
}

type Action struct {
	ActionID         string   `json:"action_id"`
	Title            string   `json:"title"`
	Why              string   `json:"why,omitempty"`
	Kind             string   `json:"kind,omitempty"`
	Priority         string   `json:"priority,omitempty"`
	Available        bool     `json:"available,omitempty"`
	UserFacingPrompt string   `json:"user_facing_prompt,omitempty"`
	APIHint          *APIHint `json:"api_hint,omitempty"`
}

type APIHint struct {
	Method string `json:"method"`
	Path   string `json:"path"`
}

type Meta struct {
	RequestID  string `json:"request_id"`
	NextCursor string `json:"next_cursor,omitempty"`
	HasMore    bool   `json:"has_more,omitempty"`
}

// SourceRef is a single source article backing a card.
type SourceRef struct {
	Title       string `json:"title"`
	SourceURL   string `json:"source_url"`
	Domain      string `json:"domain,omitempty"`
	CreatorName string `json:"creator_name,omitempty"`
	PublishedAt string `json:"published_at,omitempty"`
}

// Card is the atomic content unit.
type Card struct {
	ID                string      `json:"id"`
	Headline          string      `json:"headline"`
	Summary           string      `json:"summary,omitempty"`
	LeadText          string      `json:"lead_text,omitempty"`
	ContentType       string      `json:"content_type,omitempty"`
	Topics            []string    `json:"topics,omitempty"`
	Sources           []SourceRef `json:"sources,omitempty"`
	SourceCount       int         `json:"source_count"`
	DomainCount       int         `json:"domain_count,omitempty"`
	SignificanceScore float64     `json:"significance_score,omitempty"`
	Velocity          float64     `json:"velocity,omitempty"`
	ImageURL          string      `json:"image_url,omitempty"`
	ShortURL          string      `json:"short_url,omitempty"`
	CanonicalURL      string      `json:"canonical_url,omitempty"`
	ShareURL          string      `json:"share_url,omitempty"`
	LatestSourceAt    string      `json:"latest_source_at,omitempty"`
	CreatedAt         string      `json:"created_at,omitempty"`
	UpdatedAt         string      `json:"updated_at,omitempty"`
}

// ShareLink returns the short URL if available, empty string otherwise.
func (c Card) ShareLink() string {
	if c.ShortURL != "" {
		return c.ShortURL
	}
	return ""
}

// HasImage returns true if the card has an associated image.
func (c Card) HasImage() bool {
	return c.ImageURL != ""
}

// StreamData is the response for GET /api/v1/streams/{slug}.
type StreamData struct {
	Items  []Card `json:"items"`
	Count  int    `json:"count"`
	Stream string `json:"stream"`
}

// Category represents an active content category.
type Category struct {
	Slug  string `json:"slug"`
	Label string `json:"label"`
	Count int    `json:"count,omitempty"`
}

// CategoriesData is the response for GET /api/v1/categories.
type CategoriesData struct {
	Categories []Category `json:"categories"`
	Count      int        `json:"count"`
}

// SearchResult is the response for GET /api/v1/search.
type SearchResult struct {
	Query string `json:"query"`
	Cards []Card `json:"cards,omitempty"`
	Count int    `json:"count"`
}

// Signal records a user interaction.
type Signal struct {
	SignalType   string `json:"signal_type"`
	CardID       string `json:"card_id"`
	TopicID      string `json:"topic_id,omitempty"`
	SubmissionID string `json:"submission_id,omitempty"`
	SourceURL    string `json:"source_url,omitempty"`
	SharedTo     string `json:"shared_to,omitempty"`
}

type SignalResult struct {
	Recorded int `json:"recorded"`
}

// ReplayEvent is a historical signal event.
type ReplayEvent struct {
	EventID    string `json:"event_id"`
	SignalType string `json:"signal_type"`
	CardID     string `json:"card_id"`
	TopicID    string `json:"topic_id,omitempty"`
	Headline   string `json:"headline,omitempty"`
	SessionID  string `json:"session_id,omitempty"`
	OccurredAt string `json:"occurred_at"`
}

type ReplayData struct {
	PersonID string        `json:"person_id,omitempty"`
	Window   string        `json:"window,omitempty"`
	Events   []ReplayEvent `json:"events"`
	Count    int           `json:"count,omitempty"`
}

// Preferences holds person preference state.
type Preferences struct {
	TopicsFollowed  []string `json:"topics_followed,omitempty"`
	TopicsAvoided   []string `json:"topics_avoided,omitempty"`
	TonePreferences []string `json:"tone_preferences,omitempty"`
}

type PreferencesData struct {
	PersonID         string       `json:"person_id"`
	Status           string       `json:"status,omitempty"`
	SavedPreferences *Preferences `json:"saved_preferences,omitempty"`
}

// FeedbackPayload for POST /api/v1/feedback.
type FeedbackPayload struct {
	FeedbackType string `json:"feedback_type"`
	Title        string `json:"title"`
	Detail       string `json:"detail,omitempty"`
	TopicID      string `json:"topic_id,omitempty"`
	CardID       string `json:"card_id,omitempty"`
}

type FeedbackResult struct {
	Status string `json:"status"`
}

// RadioTrack is a single track in a radio session queue.
type RadioTrack struct {
	ID          string   `json:"id"`
	Headline    string   `json:"headline"`
	AudioURL    string   `json:"audio_url"`
	ImageURL    string   `json:"image_url,omitempty"`
	Topics      []string `json:"topics,omitempty"`
	SourceCount int      `json:"source_count,omitempty"`
}

// RadioSessionData is the response for radio session endpoints.
type RadioSessionData struct {
	RadioSessionID string       `json:"radio_session_id"`
	State          string       `json:"state"` // ready, playing, paused, ended
	Mode           string       `json:"mode,omitempty"`
	QueueLength    int          `json:"queue_length"`
	Position       int          `json:"position,omitempty"`
	Tracks         []RadioTrack `json:"tracks,omitempty"`
	Queue          []string     `json:"queue,omitempty"` // fallback: just IDs
}

// RadioStartPayload for POST /api/v1/radio-sessions.
type RadioStartPayload struct {
	Mode            string   `json:"mode"`
	DurationMinutes int      `json:"duration_minutes"`
	TopicLenses     []string `json:"topic_lenses,omitempty"`
}

// RadioControlPayload for PATCH /api/v1/radio-sessions/{id}.
type RadioControlPayload struct {
	Action string `json:"action"` // pause, resume, skip, stop
}

// APIError represents a non-2xx API response.
type APIError struct {
	StatusCode int
	Code       string `json:"code"`
	Message    string `json:"message"`
	Guidance   *Guidance
	Meta       *Meta
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API %d (%s): %s", e.StatusCode, e.Code, e.Message)
}
