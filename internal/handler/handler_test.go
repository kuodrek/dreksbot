package handler_test

import (
	"context"
	"testing"

	"github.com/drek/dreksbot/internal/model"
	"github.com/drek/dreksbot/internal/service"
)

// --- Service mocks ---

type mockPlayerService struct {
	playFn func(ctx context.Context, guildID, channelID, query string) (*model.PlayResult, error)
}

func (m *mockPlayerService) Play(ctx context.Context, guildID, channelID, query string) (*model.PlayResult, error) {
	return m.playFn(ctx, guildID, channelID, query)
}
func (m *mockPlayerService) Skip(guildID string) (*model.Track, error) { return nil, nil }
func (m *mockPlayerService) Pause(guildID string) error                { return nil }
func (m *mockPlayerService) Resume(guildID string) error               { return nil }
func (m *mockPlayerService) Stop(guildID string) error                 { return nil }
func (m *mockPlayerService) NowPlaying(guildID string) *model.Track    { return nil }
func (m *mockPlayerService) Queue(guildID string) []*model.Track       { return nil }
func (m *mockPlayerService) IsGuildActive(guildID string) bool         { return false }

var _ service.PlayerService = &mockPlayerService{}

// --- Tests ---

// Note: Testing Discord handlers directly is tricky because discordgo.Session
// and discordgo.InteractionCreate are complex structs tied to the Discord API.
//
// The recommended approach for handler tests:
//  1. Extract pure business logic into testable helper functions.
//  2. Test the service layer (player_test.go) for the real logic.
//  3. Use integration/smoke testing for the full Discord interaction flow.
//
// The test below shows the mock pattern. Write real tests after you implement handlers.

func TestHandler_New_Compiles(t *testing.T) {
	// Sanity check: the handler package compiles with our mock.
	// Replace this with real tests as you implement each command.
	_ = &mockPlayerService{}
}

// TODO: TestHandleSkip_WhenNotPlaying_RespondsWithError
// TODO: TestHandleSkip_WhenPlaying_CallsSkipAndResponds
// TODO: TestHandleQueue_Empty_RespondsWithEmptyMessage
// TODO: TestHandleQueue_WithTracks_ListsTracks
// TODO: TestHandleStop_CallsStop
// TODO: TestHandlePause_WhenPlaying_CallsPause
// TODO: TestHandleResume_WhenPaused_CallsResume
