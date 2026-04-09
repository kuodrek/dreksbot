package infra

import (
	"context"
	"os/exec"
	"testing"
)

// testYTDLPJSON mimics the JSON output of: yt-dlp --print-json --no-download -f bestaudio <url>
var testYTDLPJSON = `{
	"title": "Never Gonna Give You Up",
	"webpage_url": "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
	"url": "https://audio.example.com/stream.webm",
	"duration": 213
}`

func TestYTDLPExtractor_ExtractTrack_ParsesJSON(t *testing.T) {
	// Override the package-level execCommand to return fake JSON
	// instead of actually calling yt-dlp.
	orig := execCommand
	defer func() { execCommand = orig }()

	execCommand = func(ctx context.Context, name string, arg ...string) *exec.Cmd {
		// echo the fake JSON to stdout, simulating yt-dlp output
		return exec.Command("echo", testYTDLPJSON)
	}

	e := NewYTDLPExtractor()
	track, err := e.ExtractTrack(context.Background(), "https://www.youtube.com/watch?v=dQw4w9WgXcQ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if track.Title != "Never Gonna Give You Up" {
		t.Errorf("expected 'Never Gonna Give You Up', got %q", track.Title)
	}
	if track.URL != "https://www.youtube.com/watch?v=dQw4w9WgXcQ" {
		t.Errorf("unexpected URL: %s", track.URL)
	}
	if track.GetAudioURL() == "" {
		t.Error("expected AudioURL to be set from yt-dlp output")
	}
}

func TestYTDLPExtractor_ExtractTrack_SearchQuery_PrefixesYTSearch(t *testing.T) {
	var capturedArgs []string
	orig := execCommand
	defer func() { execCommand = orig }()

	execCommand = func(ctx context.Context, name string, arg ...string) *exec.Cmd {
		capturedArgs = arg
		return exec.Command("echo", testYTDLPJSON)
	}

	e := NewYTDLPExtractor()
	_, _ = e.ExtractTrack(context.Background(), "never gonna give you up")

	found := false
	for _, a := range capturedArgs {
		if a == "ytsearch1:never gonna give you up" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected ytsearch1: prefix in args, got: %v", capturedArgs)
	}
}

func TestParseYouTubeURL(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantVideo string
		wantList  string
	}{
		{
			name:      "video only",
			input:     "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
			wantVideo: "dQw4w9WgXcQ",
			wantList:  "",
		},
		{
			name:      "video with playlist",
			input:     "https://www.youtube.com/watch?v=dQw4w9WgXcQ&list=PLxxx",
			wantVideo: "dQw4w9WgXcQ",
			wantList:  "PLxxx",
		},
		{
			name:      "playlist only",
			input:     "https://www.youtube.com/playlist?list=PLxxx",
			wantVideo: "",
			wantList:  "PLxxx",
		},
		{
			name:      "search query (not a URL)",
			input:     "never gonna give you up",
			wantVideo: "",
			wantList:  "",
		},
		{
			name:      "non-YouTube URL",
			input:     "https://soundcloud.com/foo/bar",
			wantVideo: "",
			wantList:  "",
		},
		{
			name:      "youtu.be short URL",
			input:     "https://youtu.be/dQw4w9WgXcQ",
			wantVideo: "dQw4w9WgXcQ",
			wantList:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotVideo, gotList := ParseYouTubeURL(tt.input)
			if gotVideo != tt.wantVideo {
				t.Errorf("videoID: got %q, want %q", gotVideo, tt.wantVideo)
			}
			if gotList != tt.wantList {
				t.Errorf("listID: got %q, want %q", gotList, tt.wantList)
			}
		})
	}
}

// testPlaylistJSON mimics two lines of yt-dlp --flat-playlist --dump-json output.
var testPlaylistJSON = `{"id":"aaa111","title":"Song One","duration":180,"playlist_title":"My Playlist"}
{"id":"bbb222","title":"Song Two","duration":240,"playlist_title":"My Playlist"}
`

func TestYTDLPExtractor_ExtractPlaylist_ParsesMultilineJSON(t *testing.T) {
	orig := execCommand
	defer func() { execCommand = orig }()

	execCommand = func(ctx context.Context, name string, arg ...string) *exec.Cmd {
		return exec.Command("echo", "-n", testPlaylistJSON)
	}

	e := NewYTDLPExtractor()
	result, err := e.ExtractPlaylist(context.Background(), "https://www.youtube.com/playlist?list=PLxxx")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.PlaylistTitle != "My Playlist" {
		t.Errorf("expected playlist title %q, got %q", "My Playlist", result.PlaylistTitle)
	}
	if len(result.Tracks) != 2 {
		t.Fatalf("expected 2 tracks, got %d", len(result.Tracks))
	}
	if result.Tracks[0].Title != "Song One" {
		t.Errorf("expected %q, got %q", "Song One", result.Tracks[0].Title)
	}
	if result.Tracks[1].URL != "https://www.youtube.com/watch?v=bbb222" {
		t.Errorf("unexpected URL: %s", result.Tracks[1].URL)
	}
	if result.Tracks[0].GetAudioURL() != "" {
		t.Error("AudioURL should be empty for lazy extraction")
	}
}

// TODO: add test for yt-dlp process failing (non-zero exit code)
// TODO: add integration test (//go:build integration) that calls real yt-dlp
