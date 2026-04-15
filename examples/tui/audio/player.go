package audio

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/mp3"
	"github.com/gopxl/beep/v2/speaker"
)

// Player manages MP3 audio playback for Chowdahh Radio.
type Player struct {
	mu          sync.Mutex
	streamer    beep.StreamSeekCloser
	ctrl        *beep.Ctrl
	format      beep.Format
	playing     bool
	paused      bool
	done        chan struct{} // closed when track finishes or is stopped
	tmpFile     string
	speakerOn   bool
	speakerRate beep.SampleRate
}

// New creates a new Player.
func New() *Player {
	return &Player{}
}

// PlayURL downloads an MP3 from the given URL and starts playback.
// Returns immediately; playback runs in the background.
// The returned channel is closed when the track finishes or is stopped.
func (p *Player) PlayURL(url string) (chan struct{}, error) {
	p.Stop() // stop any existing playback

	p.mu.Lock()
	defer p.mu.Unlock()

	// Download to temp file (beep needs a seekable reader)
	tmpFile, err := downloadToTemp(url)
	if err != nil {
		return nil, fmt.Errorf("download: %w", err)
	}
	p.tmpFile = tmpFile

	f, err := os.Open(tmpFile)
	if err != nil {
		os.Remove(tmpFile)
		return nil, fmt.Errorf("open: %w", err)
	}

	streamer, format, err := mp3.Decode(f)
	if err != nil {
		f.Close()
		os.Remove(tmpFile)
		return nil, fmt.Errorf("decode: %w", err)
	}

	// Always play at the file's native sample rate — no resampling
	if !p.speakerOn {
		err = speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
		if err != nil {
			streamer.Close()
			os.Remove(tmpFile)
			return nil, fmt.Errorf("speaker init: %w", err)
		}
		p.speakerOn = true
		p.speakerRate = format.SampleRate
	}

	// Resample to match the speaker's rate if this track differs
	var src beep.Streamer = streamer
	if format.SampleRate != p.speakerRate {
		src = beep.Resample(4, format.SampleRate, p.speakerRate, streamer)
	}

	p.streamer = streamer
	p.format = format
	p.ctrl = &beep.Ctrl{Streamer: src, Paused: false}
	p.playing = true
	p.paused = false
	p.done = make(chan struct{})

	done := p.done
	speaker.Play(beep.Seq(p.ctrl, beep.Callback(func() {
		p.mu.Lock()
		p.playing = false
		p.mu.Unlock()
		select {
		case <-done:
		default:
			close(done)
		}
	})))

	return done, nil
}

// Pause toggles pause state.
func (p *Player) Pause() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.ctrl != nil {
		speaker.Lock()
		p.ctrl.Paused = !p.ctrl.Paused
		p.paused = p.ctrl.Paused
		speaker.Unlock()
	}
}

// Stop stops playback and cleans up.
func (p *Player) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.speakerOn {
		speaker.Clear()
	}

	if p.streamer != nil {
		p.streamer.Close()
		p.streamer = nil
	}
	if p.tmpFile != "" {
		os.Remove(p.tmpFile)
		p.tmpFile = ""
	}
	if p.done != nil {
		select {
		case <-p.done:
		default:
			close(p.done)
		}
		p.done = nil
	}
	p.playing = false
	p.paused = false
	p.ctrl = nil
}

// IsPlaying returns true if audio is currently playing (not paused).
func (p *Player) IsPlaying() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.playing && !p.paused
}

// IsPaused returns true if audio is paused.
func (p *Player) IsPaused() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.paused
}

// IsActive returns true if player has a track loaded (playing or paused).
func (p *Player) IsActive() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.playing
}

// Position returns the current playback position.
func (p *Player) Position() time.Duration {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.streamer == nil {
		return 0
	}
	speaker.Lock()
	pos := p.format.SampleRate.D(p.streamer.Position())
	speaker.Unlock()
	return pos
}

// Duration returns the total track duration.
func (p *Player) Duration() time.Duration {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.streamer == nil {
		return 0
	}
	return p.format.SampleRate.D(p.streamer.Len())
}

func downloadToTemp(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	f, err := os.CreateTemp("", "chowdahh-*.mp3")
	if err != nil {
		return "", err
	}
	defer f.Close()

	if _, err := io.Copy(f, resp.Body); err != nil {
		os.Remove(f.Name())
		return "", err
	}

	return f.Name(), nil
}
