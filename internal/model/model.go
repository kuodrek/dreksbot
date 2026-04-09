package model

import (
	"sync"
	"time"
)

// Track represents a single audio track from YouTube.
type Track struct {
	Title    string        // Human-readable track title
	URL      string        // Original YouTube URL (e.g. https://youtube.com/watch?v=...)
	Duration time.Duration // Track length

	mu       sync.Mutex // guards audioURL
	audioURL string     // Direct audio stream URL (from yt-dlp, extracted lazily before playback)
}

// GetAudioURL returns the direct audio stream URL (thread-safe).
func (t *Track) GetAudioURL() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.audioURL
}

// SetAudioURL sets the direct audio stream URL (thread-safe).
func (t *Track) SetAudioURL(url string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.audioURL = url
}

// PlayResult is returned by PlayerService.Play to provide both the track
// and optional playlist information when a playlist URL is detected.
type PlayResult struct {
	Track         *Track // The track that is now playing or was enqueued
	PlaylistName  string // Non-empty if a playlist was detected
	PlaylistCount int    // Number of additional tracks queued from playlist
}

// PlaybackState is the state machine for a guild's player.
// Transitions: Idle -> Playing -> Paused -> Playing -> Idle
type PlaybackState int

const (
	StateIdle    PlaybackState = iota // Not connected, nothing playing
	StatePlaying                      // Actively sending audio to Discord
	StatePaused                       // Connected to voice but audio paused
)
