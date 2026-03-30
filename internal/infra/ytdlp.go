package infra

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/drek/dreksbot/internal/model"
)

// execCommand is a variable so tests can override it to avoid real subprocess calls.
// In production this is exec.CommandContext — the standard library function.
var execCommand = exec.CommandContext

// AudioExtractor fetches metadata and a streamable audio URL from YouTube.
type AudioExtractor interface {
	// ExtractTrack takes a YouTube URL or search keywords and returns track metadata
	// including a direct audio stream URL. If query is not a URL, it is treated as
	// a YouTube search (uses yt-dlp's ytsearch1: prefix).
	ExtractTrack(ctx context.Context, query string) (*model.Track, error)

	// ExtractPlaylist fetches all tracks from a YouTube playlist URL.
	// NOTE: only title and original URL are populated — AudioURL is left empty
	// and will be lazily extracted before playback to prevent URL expiration.
	ExtractPlaylist(ctx context.Context, playlistURL string) ([]*model.Track, error)
}

// ytdlpExtractor implements AudioExtractor using the yt-dlp CLI.
type ytdlpExtractor struct{}

// NewYTDLPExtractor returns an AudioExtractor backed by the yt-dlp CLI.
// yt-dlp must be installed and available on PATH.
func NewYTDLPExtractor() AudioExtractor {
	return &ytdlpExtractor{}
}

// ytdlpOutput is the subset of yt-dlp's JSON output we care about.
type ytdlpOutput struct {
	Title      string  `json:"title"`
	WebpageURL string  `json:"webpage_url"`
	URL        string  `json:"url"`      // Direct audio stream URL
	Duration   float64 `json:"duration"` // Seconds
}

func (e *ytdlpExtractor) ExtractTrack(ctx context.Context, query string) (*model.Track, error) {
	// If query is not a URL, prefix with ytsearch1: to search YouTube.
	if !strings.HasPrefix(query, "http://") && !strings.HasPrefix(query, "https://") {
		query = "ytsearch1:" + query
	}

	args := []string{
		"-v",                    // Verbose output to stderr for debugging
		"--no-download",         // Don't actually download the file
		"--print-json",          // Output metadata as JSON to stdout
		"-f", "bestaudio*",      // Best audio; * allows fallback to audio from combined formats
		"--no-playlist",         // Treat playlist URLs as single videos
		"--js-runtimes", "node", // Use Node.js to solve YouTube JS challenges
	}

	const cookiesPath = "/cookies/cookies.txt"
	if _, err := os.Stat(cookiesPath); err == nil {
		args = append(args, "--cookies", cookiesPath)
	}

	args = append(args, query)
	cmd := execCommand(ctx, "yt-dlp", args...)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	out, err := cmd.Output()
	if err != nil {
		log.Printf("[yt-dlp] failed for query %q: %v\nstderr:\n%s", query, err, stderr.String())
		return nil, fmt.Errorf("yt-dlp failed for query %q: %w", query, err)
	}

	log.Printf("[yt-dlp] success for query %q", query)

	var result ytdlpOutput
	if err := json.Unmarshal(out, &result); err != nil {
		return nil, fmt.Errorf("parsing yt-dlp JSON: %w", err)
	}

	return &model.Track{
		Title:    result.Title,
		URL:      result.WebpageURL,
		AudioURL: result.URL,
		Duration: time.Duration(result.Duration) * time.Second,
	}, nil
}

func (e *ytdlpExtractor) ExtractPlaylist(ctx context.Context, playlistURL string) ([]*model.Track, error) {
	// TODO: implement
	// Use yt-dlp with --flat-playlist --print-json to get all entries.
	// For each entry, create a Track with only Title and URL populated.
	// Leave AudioURL empty — the playback goroutine will call ExtractTrack
	// lazily before playing each track to avoid URL expiration.
	//
	// Hint: yt-dlp --flat-playlist outputs one JSON object per line.
	// Use a bufio.Scanner to read line-by-line and unmarshal each one.
	return nil, fmt.Errorf("not implemented")
}
