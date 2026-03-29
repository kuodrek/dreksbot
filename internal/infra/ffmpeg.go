package infra

import (
	"context"
	"fmt"
	"io"
	"os/exec"
)

// AudioStream represents a running ffmpeg process producing raw PCM audio.
// It implements io.Reader — callers read PCM bytes (s16le, 48kHz, stereo).
type AudioStream interface {
	io.Reader
	// Stop kills the ffmpeg process. Safe to call multiple times.
	Stop() error
}

// AudioEncoder launches ffmpeg subprocesses to transcode audio URLs into PCM.
type AudioEncoder interface {
	// NewStream starts an ffmpeg process reading from audioURL and writing
	// raw PCM to stdout (s16le, 48kHz, stereo). The process is tied to ctx:
	// cancelling ctx kills ffmpeg (used by Skip and Stop commands).
	NewStream(ctx context.Context, audioURL string) (AudioStream, error)
}

// ffmpegEncoder implements AudioEncoder.
type ffmpegEncoder struct{}

// NewFFmpegEncoder returns an AudioEncoder backed by the ffmpeg CLI.
// ffmpeg must be installed and available on PATH.
func NewFFmpegEncoder() AudioEncoder {
	return &ffmpegEncoder{}
}

func (e *ffmpegEncoder) NewStream(ctx context.Context, audioURL string) (AudioStream, error) {
	cmd := execCommand(ctx, "ffmpeg",
		// Reconnect flags: crucial for YouTube stream URLs which can time out
		"-reconnect", "1",
		"-reconnect_streamed", "1",
		"-reconnect_delay_max", "5",
		"-i", audioURL,
		"-f", "s16le", // Raw PCM, 16-bit signed little-endian
		"-ar", "48000", // 48kHz sample rate (Discord requirement)
		"-ac", "2", // Stereo (2 channels)
		"pipe:1", // Output to stdout
	)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("creating ffmpeg stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("starting ffmpeg: %w", err)
	}

	return &ffmpegStream{cmd: cmd, reader: stdout}, nil
}

// ffmpegStream wraps a running ffmpeg process and its stdout.
type ffmpegStream struct {
	cmd    *exec.Cmd
	reader io.ReadCloser
}

// Read implements io.Reader — reads raw PCM bytes from ffmpeg stdout.
// Returns io.EOF when the track ends.
func (s *ffmpegStream) Read(p []byte) (int, error) {
	return s.reader.Read(p)
}

// Stop kills the ffmpeg process and waits for it to exit.
func (s *ffmpegStream) Stop() error {
	if s.cmd.Process != nil {
		_ = s.cmd.Process.Kill()
	}
	return s.cmd.Wait()
}
