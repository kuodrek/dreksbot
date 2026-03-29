package model

import "time"

// Track represents a single audio track from YouTube.
type Track struct {
	Title    string        // Human-readable track title
	URL      string        // Original YouTube URL (e.g. https://youtube.com/watch?v=...)
	AudioURL string        // Direct audio stream URL (from yt-dlp, extracted lazily before playback)
	Duration time.Duration // Track length
}

// PlaybackState is the state machine for a guild's player.
// Transitions: Idle -> Playing -> Paused -> Playing -> Idle
type PlaybackState int

const (
	StateIdle    PlaybackState = iota // Not connected, nothing playing
	StatePlaying                      // Actively sending audio to Discord
	StatePaused                       // Connected to voice but audio paused
)
