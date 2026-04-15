package infra

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/drek/dreksbot/internal/model"
)

// execCommand is a variable so tests can override it to avoid real subprocess calls.
// In production this is exec.CommandContext — the standard library function.
var execCommand = exec.CommandContext

// ParseYouTubeURL extracts the video ID and playlist ID from a YouTube URL.
// Returns empty strings for fields not present. Non-YouTube URLs return ("", "").
func ParseYouTubeURL(rawURL string) (videoID, listID string) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", ""
	}
	if !strings.Contains(u.Host, "youtube.com") && !strings.Contains(u.Host, "youtu.be") {
		return "", ""
	}
	q := u.Query()
	listID = q.Get("list")
	if u.Host == "youtu.be" {
		videoID = strings.TrimPrefix(u.Path, "/")
	} else {
		videoID = q.Get("v")
	}
	return videoID, listID
}

// PlaylistResult holds the tracks and title extracted from a YouTube playlist.
type PlaylistResult struct {
	PlaylistTitle string
	Tracks        []*model.Track
}

// AudioExtractor fetches metadata and a streamable audio URL from YouTube.
type AudioExtractor interface {
	// ExtractTrack takes a YouTube URL or search keywords and returns track metadata
	// including a direct audio stream URL. If query is not a URL, it is treated as
	// a YouTube search (uses yt-dlp's ytsearch1: prefix).
	ExtractTrack(ctx context.Context, query string) (*model.Track, error)

	// ExtractPlaylist fetches all tracks from a YouTube playlist URL.
	// NOTE: only title and original URL are populated — AudioURL is left empty
	// and will be lazily extracted before playback to prevent URL expiration.
	ExtractPlaylist(ctx context.Context, playlistURL string) (*PlaylistResult, error)
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
		log.Printf("[yt-dlp] using cookies from %s", cookiesPath)
	} else {
		log.Printf("[yt-dlp] no cookies file found at %s, running without authentication", cookiesPath)
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

	track := &model.Track{
		Title:    result.Title,
		URL:      result.WebpageURL,
		Duration: time.Duration(result.Duration) * time.Second,
	}
	track.SetAudioURL(result.URL)
	return track, nil
}

// flatEntry is the subset of yt-dlp's flat-playlist JSON we care about.
type flatEntry struct {
	ID            string  `json:"id"`
	Title         string  `json:"title"`
	Duration      float64 `json:"duration"`
	PlaylistTitle string  `json:"playlist_title"`
}

func (e *ytdlpExtractor) ExtractPlaylist(ctx context.Context, playlistURL string) (*PlaylistResult, error) {
	args := []string{
		"--flat-playlist",
		"--dump-json",
		"--no-warnings",
	}

	const cookiesPath = "/cookies/cookies.txt"
	if _, err := os.Stat(cookiesPath); err == nil {
		args = append(args, "--cookies", cookiesPath)
	}

	args = append(args, playlistURL)
	cmd := execCommand(ctx, "yt-dlp", args...)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	out, err := cmd.Output()
	if err != nil {
		log.Printf("[yt-dlp] ExtractPlaylist failed for %q: %v\nstderr:\n%s", playlistURL, err, stderr.String())
		return nil, fmt.Errorf("yt-dlp ExtractPlaylist failed for %q: %w", playlistURL, err)
	}

	var result PlaylistResult
	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var entry flatEntry
		if err := json.Unmarshal(line, &entry); err != nil {
			log.Printf("[yt-dlp] skipping unparseable playlist entry: %v", err)
			continue
		}
		if result.PlaylistTitle == "" && entry.PlaylistTitle != "" {
			result.PlaylistTitle = entry.PlaylistTitle
		}
		track := &model.Track{
			Title:    entry.Title,
			URL:      "https://www.youtube.com/watch?v=" + entry.ID,
			Duration: time.Duration(entry.Duration) * time.Second,
		}
		result.Tracks = append(result.Tracks, track)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading yt-dlp playlist output: %w", err)
	}

	log.Printf("[yt-dlp] ExtractPlaylist: found %d tracks for %q", len(result.Tracks), playlistURL)
	return &result, nil
}
