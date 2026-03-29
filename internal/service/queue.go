package service

import (
	"sync"

	"github.com/drek/dreksbot/internal/model"
)

// QueueService manages per-guild FIFO track queues.
// All methods are safe for concurrent use from multiple goroutines.
type QueueService interface {
	// Add appends a single track to the end of the guild's queue.
	Add(guildID string, track *model.Track)

	// AddAll appends multiple tracks in order to the guild's queue.
	AddAll(guildID string, tracks []*model.Track)

	// Next pops and returns the next track from the front of the queue.
	// Returns nil if the queue is empty.
	Next(guildID string) *model.Track

	// List returns a copy of all queued tracks (does not remove them).
	List(guildID string) []*model.Track

	// Clear removes all tracks from the guild's queue.
	Clear(guildID string)

	// Len returns the number of tracks waiting in the queue.
	Len(guildID string) int
}

// queueService is the in-memory implementation of QueueService.
type queueService struct {
	mu     sync.RWMutex
	queues map[string][]*model.Track // guildID -> ordered slice of tracks
}

// NewQueueService returns a new in-memory QueueService.
func NewQueueService() QueueService {
	return &queueService{
		queues: make(map[string][]*model.Track),
	}
}

func (q *queueService) Add(guildID string, track *model.Track) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.queues[guildID] = append(q.queues[guildID], track)
}

func (q *queueService) AddAll(guildID string, tracks []*model.Track) {
	// TODO: implement
	// Hint: similar to Add but appends a slice. Use append(q.queues[guildID], tracks...).
}

func (q *queueService) Next(guildID string) *model.Track {
	q.mu.Lock()
	defer q.mu.Unlock()
	queue := q.queues[guildID]
	if len(queue) == 0 {
		return nil
	}
	// Pop from the front (FIFO)
	track := queue[0]
	q.queues[guildID] = queue[1:]
	return track
}

func (q *queueService) List(guildID string) []*model.Track {
	// TODO: implement
	// Important: return a COPY of the slice, not the internal slice itself.
	// If you return the internal slice, callers could mutate it and corrupt the queue.
	return nil
}

func (q *queueService) Clear(guildID string) {
	// TODO: implement
	// Delete the guild's entry from the map (or set it to an empty slice).
}

func (q *queueService) Len(guildID string) int {
	// TODO: implement
	return 0
}
