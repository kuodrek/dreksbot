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
	if track.AudioURL == "" {
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

// TODO: add test for yt-dlp process failing (non-zero exit code)
// TODO: add test for ExtractPlaylist
// TODO: add integration test (//go:build integration) that calls real yt-dlp
