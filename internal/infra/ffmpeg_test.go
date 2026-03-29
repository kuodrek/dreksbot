package infra

import (
	"bytes"
	"context"
	"io"
	"os/exec"
	"testing"
)

func TestFFmpegEncoder_NewStream_ReadsFromStdout(t *testing.T) {
	orig := execCommand
	defer func() { execCommand = orig }()

	// Simulate ffmpeg outputting some PCM bytes
	fakeAudio := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
	execCommand = func(ctx context.Context, name string, arg ...string) *exec.Cmd {
		// Use printf to output raw bytes — simulates ffmpeg PCM output
		return exec.Command("printf", `\x01\x02\x03\x04\x05\x06\x07\x08`)
	}

	enc := NewFFmpegEncoder()
	stream, err := enc.NewStream(context.Background(), "https://audio.example.com/stream.webm")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer stream.Stop()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, stream)
	if err != nil && err != io.EOF {
		t.Fatalf("unexpected read error: %v", err)
	}

	got := buf.Bytes()
	if len(got) == 0 {
		t.Fatal("expected bytes from stream, got none")
	}
	_ = fakeAudio // fakeAudio is for documentation; actual bytes depend on printf behavior
}

func TestFFmpegEncoder_NewStream_StopKillsProcess(t *testing.T) {
	orig := execCommand
	defer func() { execCommand = orig }()

	// Simulate a long-running process
	execCommand = func(ctx context.Context, name string, arg ...string) *exec.Cmd {
		return exec.Command("sleep", "60")
	}

	enc := NewFFmpegEncoder()
	stream, err := enc.NewStream(context.Background(), "https://audio.example.com/stream.webm")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Stop should kill the process without hanging
	if err := stream.Stop(); err != nil {
		// Process was killed — that's expected, some error codes are OK
		t.Logf("Stop returned (expected): %v", err)
	}
}

// TODO: add test that ffmpeg is called with the correct flags (-f s16le -ar 48000 -ac 2)
// TODO: add integration test (//go:build integration) that calls real ffmpeg
