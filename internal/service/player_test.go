package service_test

import (
	"context"
	"testing"

	"github.com/drek/dreksbot/internal/infra"
	"github.com/drek/dreksbot/internal/model"
	"github.com/drek/dreksbot/internal/service"
)

// --- Mocks ---
// Hand-written mocks are idiomatic Go. No framework needed.

type mockExtractor struct {
	extractTrackFn func(ctx context.Context, query string) (*model.Track, error)
}

func (m *mockExtractor) ExtractTrack(ctx context.Context, query string) (*model.Track, error) {
	return m.extractTrackFn(ctx, query)
}
func (m *mockExtractor) ExtractPlaylist(ctx context.Context, url string) ([]*model.Track, error) {
	return nil, nil
}

type mockEncoder struct{}

func (m *mockEncoder) NewStream(ctx context.Context, url string) (infra.AudioStream, error) {
	return &mockStream{}, nil
}

type mockStream struct{}

func (m *mockStream) Read(p []byte) (int, error) { return 0, nil }
func (m *mockStream) Stop() error                { return nil }

type mockVoiceConn struct{}

func (m *mockVoiceConn) SendAudio(ctx context.Context, s infra.AudioStream) error { return nil }
func (m *mockVoiceConn) Pause()                                                   {}
func (m *mockVoiceConn) Resume()                                                  {}
func (m *mockVoiceConn) Disconnect() error                                        { return nil }
func (m *mockVoiceConn) IsConnected() bool                                        { return true }

type mockVoiceFactory struct {
	joinFn func(guildID, channelID string) (infra.VoiceConnection, error)
}

func (m *mockVoiceFactory) Join(guildID, channelID string) (infra.VoiceConnection, error) {
	return m.joinFn(guildID, channelID)
}

// --- Tests ---

func TestPlayerService_Play_ReturnsExtractedTrack(t *testing.T) {
	expected := &model.Track{
		Title:    "Never Gonna Give You Up",
		URL:      "https://youtube.com/watch?v=dQw4w9WgXcQ",
		AudioURL: "https://audio.example.com/stream.webm",
	}

	player := service.NewPlayerService(
		&mockExtractor{extractTrackFn: func(_ context.Context, _ string) (*model.Track, error) {
			return expected, nil
		}},
		&mockEncoder{},
		&mockVoiceFactory{joinFn: func(_, _ string) (infra.VoiceConnection, error) {
			return &mockVoiceConn{}, nil
		}},
		service.NewQueueService(),
	)

	track, err := player.Play(context.Background(), "guild1", "channel1", "never gonna give you up")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if track.Title != expected.Title {
		t.Errorf("expected %q, got %q", expected.Title, track.Title)
	}
}

// TODO: TestPlayerService_Play_AlreadyPlaying_AddsToQueue
// TODO: TestPlayerService_Play_ExtractorError_ReturnsError
// TODO: TestPlayerService_Play_VoiceJoinError_ReturnsError
// TODO: TestPlayerService_Skip_AdvancesToNextTrack (implement after you write Skip)
// TODO: TestPlayerService_Stop_ClearsEverything (implement after you write Stop)
