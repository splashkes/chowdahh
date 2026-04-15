package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"
)

// Client is the Chowdahh API client.
type Client struct {
	BaseURL    string
	Token      string
	HTTPClient *http.Client

	// Cached from latest response guidance.
	LastRateLimit *RateLimit
	LastAuthMode  string
	mu            sync.RWMutex
}

// NewClient creates a new API client.
func NewClient(baseURL, token string) *Client {
	return &Client{
		BaseURL: baseURL,
		Token:   token,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// RateInfo returns the last-seen rate limit info (thread-safe).
func (c *Client) RateInfo() (*RateLimit, string) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.LastRateLimit, c.LastAuthMode
}

// --- Discovery ---

func (c *Client) GetStream(slug string, limit int, cursor string) (*Envelope[StreamData], error) {
	q := url.Values{"limit": {strconv.Itoa(limit)}}
	if cursor != "" {
		q.Set("cursor", cursor)
	}
	var env Envelope[StreamData]
	if err := c.get(fmt.Sprintf("/api/v1/streams/%s?%s", url.PathEscape(slug), q.Encode()), &env); err != nil {
		return nil, err
	}
	return &env, nil
}

func (c *Client) GetCategories() (*Envelope[CategoriesData], error) {
	var env Envelope[CategoriesData]
	if err := c.get("/api/v1/categories", &env); err != nil {
		return nil, err
	}
	return &env, nil
}

func (c *Client) Search(query string, limit int) (*Envelope[SearchResult], error) {
	q := url.Values{"q": {query}, "limit": {strconv.Itoa(limit)}}
	var env Envelope[SearchResult]
	if err := c.get(fmt.Sprintf("/api/v1/search?%s", q.Encode()), &env); err != nil {
		return nil, err
	}
	return &env, nil
}

// --- Signals ---

func (c *Client) RecordSignals(signals []Signal) error {
	var env Envelope[SignalResult]
	return c.post("/api/v1/signals", signals, &env)
}

// --- Replay ---

func (c *Client) GetReplay(period, signalType string, limit int, cursor string) (*Envelope[ReplayData], error) {
	q := url.Values{"limit": {strconv.Itoa(limit)}}
	if period != "" {
		q.Set("period", period)
	}
	if signalType != "" {
		q.Set("signal_type", signalType)
	}
	if cursor != "" {
		q.Set("cursor", cursor)
	}
	var env Envelope[ReplayData]
	if err := c.get(fmt.Sprintf("/api/v1/replay?%s", q.Encode()), &env); err != nil {
		return nil, err
	}
	return &env, nil
}

// --- Radio ---

func (c *Client) StartRadioSession(payload RadioStartPayload) (*Envelope[RadioSessionData], error) {
	var env Envelope[RadioSessionData]
	if err := c.post("/api/v1/radio-sessions", payload, &env); err != nil {
		return nil, err
	}
	return &env, nil
}

func (c *Client) GetRadioSession(sessionID string) (*Envelope[RadioSessionData], error) {
	var env Envelope[RadioSessionData]
	if err := c.get(fmt.Sprintf("/api/v1/radio-sessions/%s", url.PathEscape(sessionID)), &env); err != nil {
		return nil, err
	}
	return &env, nil
}

func (c *Client) UpdateRadioSession(sessionID string, payload RadioControlPayload) (*Envelope[RadioSessionData], error) {
	var env Envelope[RadioSessionData]
	if err := c.patch(fmt.Sprintf("/api/v1/radio-sessions/%s", url.PathEscape(sessionID)), payload, &env); err != nil {
		return nil, err
	}
	return &env, nil
}

// AudioURL returns the full URL for streaming a track's audio.
func (c *Client) AudioURL(trackID string) string {
	return c.BaseURL + "/audio/" + trackID
}

// --- Preferences ---

func (c *Client) GetPreferences(personID string) (*Envelope[PreferencesData], error) {
	var env Envelope[PreferencesData]
	if err := c.get(fmt.Sprintf("/api/v1/preferences/%s", url.PathEscape(personID)), &env); err != nil {
		return nil, err
	}
	return &env, nil
}

func (c *Client) SetPreferences(personID string, prefs Preferences) (*Envelope[PreferencesData], error) {
	var env Envelope[PreferencesData]
	if err := c.put(fmt.Sprintf("/api/v1/preferences/%s", url.PathEscape(personID)), prefs, &env); err != nil {
		return nil, err
	}
	return &env, nil
}

// --- Feedback ---

func (c *Client) SubmitFeedback(payload FeedbackPayload) (*Envelope[FeedbackResult], error) {
	var env Envelope[FeedbackResult]
	if err := c.post("/api/v1/feedback", payload, &env); err != nil {
		return nil, err
	}
	return &env, nil
}

// --- Internal HTTP ---

func (c *Client) get(path string, out interface{}) error {
	return c.do("GET", path, nil, out)
}

func (c *Client) post(path string, body interface{}, out interface{}) error {
	return c.do("POST", path, body, out)
}

func (c *Client) put(path string, body interface{}, out interface{}) error {
	return c.do("PUT", path, body, out)
}

func (c *Client) patch(path string, body interface{}, out interface{}) error {
	return c.do("PATCH", path, body, out)
}

func (c *Client) do(method, path string, body interface{}, out interface{}) error {
	u := c.BaseURL + path

	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, u, bodyReader)
	if err != nil {
		return fmt.Errorf("request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("fetch: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return c.parseError(resp.StatusCode, respBody)
	}

	if err := json.Unmarshal(respBody, out); err != nil {
		return fmt.Errorf("decode: %w", err)
	}

	// Cache guidance from response
	c.cacheGuidance(respBody)

	return nil
}

func (c *Client) parseError(statusCode int, body []byte) *APIError {
	var errEnv struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
		Guidance *Guidance `json:"guidance,omitempty"`
		Meta     *Meta     `json:"meta,omitempty"`
	}
	json.Unmarshal(body, &errEnv)

	return &APIError{
		StatusCode: statusCode,
		Code:       errEnv.Error.Code,
		Message:    errEnv.Error.Message,
		Guidance:   errEnv.Guidance,
		Meta:       errEnv.Meta,
	}
}

func (c *Client) cacheGuidance(body []byte) {
	var partial struct {
		Guidance *Guidance `json:"guidance"`
	}
	if json.Unmarshal(body, &partial) == nil && partial.Guidance != nil {
		c.mu.Lock()
		defer c.mu.Unlock()
		if partial.Guidance.AccountState != nil {
			c.LastAuthMode = partial.Guidance.AccountState.AuthMode
			c.LastRateLimit = partial.Guidance.AccountState.RateLimit
		}
	}
}
