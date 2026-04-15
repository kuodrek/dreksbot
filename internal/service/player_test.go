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
	extractTrackFn    func(ctx context.Context, query string) (*model.Track, error)
	extractPlaylistFn func(ctx context.Context, url string) (*infra.PlaylistResult, error)
}

func (m *mockExtractor) ExtractTrack(ctx context.Context, query string) (*model.Track, error) {
	return m.extractTrackFn(ctx, query)
}
func (m *mockExtractor) ExtractPlaylist(ctx context.Context, url string) (*infra.PlaylistResult, error) {
	if m.extractPlaylistFn != nil {
		return m.extractPlaylistFn(ctx, url)
	}
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
func (m *mockVoiceConn) Disconnect(ctx context.Context) error                     { return nil }
func (m *mockVoiceConn) IsConnected() bool                                        { return true }

type mockVoiceFactory struct {
	joinFn func(ctx context.Context, guildID, channelID string) (infra.VoiceConnection, error)
}

func (m *mockVoiceFactory) Join(ctx context.Context, guildID, channelID string) (infra.VoiceConnection, error) {
	return m.joinFn(ctx, guildID, channelID)
}

// --- Tests ---

func TestPlayerService_Play_ReturnsExtractedTrack(t *testing.T) {
	expected := &model.Track{
		Title: "Never Gonna Give You Up",
		URL:   "https://youtube.com/watch?v=dQw4w9WgXcQ",
	}
	expected.SetAudioURL("https://audio.example.com/stream.webm")

	player := service.NewPlayerService(
		&mockExtractor{extractTrackFn: func(_ context.Context, _ string) (*model.Track, error) {
			return expected, nil
		}},
		&mockEncoder{},
		&mockVoiceFactory{joinFn: func(_ context.Context, _, _ string) (infra.VoiceConnection, error) {
			return &mockVoiceConn{}, nil
		}},
		service.NewQueueService(),
	)

	result, err := player.Play(context.Background(), "guild1", "channel1", "never gonna give you up")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Track.Title != expected.Title {
		t.Errorf("expected %q, got %q", expected.Title, result.Track.Title)
	}
}

func newTestPlayer(extractor infra.AudioExtractor) service.PlayerService {
	return service.NewPlayerService(
		extractor,
		&mockEncoder{},
		&mockVoiceFactory{joinFn: func(_ context.Context, _, _ string) (infra.VoiceConnection, error) {
			return &mockVoiceConn{}, nil
		}},
		service.NewQueueService(),
	)
}

func TestPlayerService_Play_PlaylistOnly_ReturnsPlaylistMetadata(t *testing.T) {
	tracks := []*model.Track{
		{Title: "Track 1", URL: "https://www.youtube.com/watch?v=aaa111"},
		{Title: "Track 2", URL: "https://www.youtube.com/watch?v=bbb222"},
		{Title: "Track 3", URL: "https://www.youtube.com/watch?v=ccc333"},
	}

	firstResolved := &model.Track{Title: "Track 1", URL: "https://www.youtube.com/watch?v=aaa111"}
	firstResolved.SetAudioURL("https://audio.example.com/1.webm")

	extractor := &mockExtractor{
		extractTrackFn: func(_ context.Context, _ string) (*model.Track, error) {
			return firstResolved, nil
		},
		extractPlaylistFn: func(_ context.Context, _ string) (*infra.PlaylistResult, error) {
			return &infra.PlaylistResult{PlaylistTitle: "My Playlist", Tracks: tracks}, nil
		},
	}

	player := newTestPlayer(extractor)
	result, err := player.Play(context.Background(), "guild1", "ch1", "https://www.youtube.com/playlist?list=PLxxx")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.PlaylistName != "My Playlist" {
		t.Errorf("expected playlist name %q, got %q", "My Playlist", result.PlaylistName)
	}
	// First track is used for playback; remaining 2 are reported as queued.
	if result.PlaylistCount != 2 {
		t.Errorf("expected PlaylistCount 2, got %d", result.PlaylistCount)
	}
	if result.Track.Title != "Track 1" {
		t.Errorf("expected playing track %q, got %q", "Track 1", result.Track.Title)
	}
}

func TestPlayerService_Play_VideoWithPlaylist_DeduplicatesVideo(t *testing.T) {
	videoTrack := &model.Track{Title: "The Video", URL: "https://www.youtube.com/watch?v=dQw4w9WgXcQ"}
	videoTrack.SetAudioURL("https://audio.example.com/video.webm")

	playlistTracks := []*model.Track{
		{Title: "The Video", URL: "https://www.youtube.com/watch?v=dQw4w9WgXcQ"}, // duplicate
		{Title: "Other Track", URL: "https://www.youtube.com/watch?v=zzz999"},
	}

	extractor := &mockExtractor{
		extractTrackFn: func(_ context.Context, _ string) (*model.Track, error) {
			return videoTrack, nil
		},
		extractPlaylistFn: func(_ context.Context, _ string) (*infra.PlaylistResult, error) {
			return &infra.PlaylistResult{PlaylistTitle: "Mix", Tracks: playlistTracks}, nil
		},
	}

	player := newTestPlayer(extractor)
	result, err := player.Play(context.Background(), "guild1", "ch1",
		"https://www.youtube.com/watch?v=dQw4w9WgXcQ&list=PLxxx")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Duplicate removed: only 1 track left from playlist portion.
	if result.PlaylistCount != 1 {
		t.Errorf("expected PlaylistCount 1 after dedup, got %d", result.PlaylistCount)
	}
	if result.Track.Title != "The Video" {
		t.Errorf("expected playing track %q, got %q", "The Video", result.Track.Title)
	}
}

// TODO: TestPlayerService_Play_AlreadyPlaying_AddsToQueue
// TODO: TestPlayerService_Play_ExtractorError_ReturnsError
// TODO: TestPlayerService_Play_VoiceJoinError_ReturnsError
// TODO: TestPlayerService_Skip_AdvancesToNextTrack (implement after you write Skip)
// TODO: TestPlayerService_Stop_ClearsEverything (implement after you write Stop)
