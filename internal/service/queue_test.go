package service_test

import (
	"testing"

	"github.com/drek/dreksbot/internal/model"
	"github.com/drek/dreksbot/internal/service"
)

// This is the example test showing the pattern. You will write more below.
func TestQueueService_AddAndNext_ReturnsFIFO(t *testing.T) {
	q := service.NewQueueService()

	track1 := &model.Track{Title: "Track 1", URL: "https://youtube.com/1"}
	track2 := &model.Track{Title: "Track 2", URL: "https://youtube.com/2"}

	q.Add("guild1", track1)
	q.Add("guild1", track2)

	got := q.Next("guild1")
	if got != track1 {
		t.Errorf("expected track1 first, got %v", got)
	}
	got = q.Next("guild1")
	if got != track2 {
		t.Errorf("expected track2 second, got %v", got)
	}
	got = q.Next("guild1")
	if got != nil {
		t.Errorf("expected nil on empty queue, got %v", got)
	}
}

// TODO: TestQueueService_Guilds_AreIsolated — Add to guild1 should not affect guild2
// TODO: TestQueueService_List_ReturnsCopy — mutating result of List should not change internal queue
// TODO: TestQueueService_Clear_EmptiesQueue
// TODO: TestQueueService_Len_ReflectsQueueSize
// TODO: TestQueueService_AddAll_AppendsInOrder
// TODO: TestQueueService_ConcurrentAccess — run with -race flag to catch data races
