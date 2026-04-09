package service

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/drek/dreksbot/internal/infra"
	"github.com/drek/dreksbot/internal/model"
)

// PlayerService orchestrates playback for all guilds.
// It manages voice connections, the playback goroutine lifecycle,
// and delegates queue management to QueueService.
type PlayerService interface {
	// Play resolves query (URL or keywords) via yt-dlp, joins the user's voice
	// channel if not already connected, and starts playing or enqueues the track.
	// If query is a playlist URL, all tracks are queued and PlayResult includes
	// PlaylistName and PlaylistCount.
	Play(ctx context.Context, guildID, channelID, query string) (*model.PlayResult, error)

	// Skip cancels the current track. The playback goroutine automatically advances
	// to the next track in the queue. Returns the newly playing track (or nil).
	Skip(guildID string) (*model.Track, error)

	// Pause pauses audio without disconnecting from voice.
	Pause(guildID string) error

	// Resume resumes paused playback.
	Resume(guildID string) error

	// Stop ends playback, clears the queue, and disconnects from voice.
	Stop(guildID string) error

	// NowPlaying returns the track currently playing, or nil if idle.
	NowPlaying(guildID string) *model.Track

	// Queue returns a snapshot of tracks waiting to play (not including current).
	Queue(guildID string) []*model.Track

	IsGuildActive(guildID string) bool
}

// guildPlayer holds all per-guild playback state.
// Each guild has its own mutex so operations on guild A never block guild B.
type guildPlayer struct {
	mu           sync.Mutex
	ctx          context.Context    // Guild-level context. Cancelled by Stop.
	cancelGuild  context.CancelFunc // Cancels ctx — ends the playback goroutine.
	cancelTrack  context.CancelFunc // Cancels current track only — used by Skip.
	state        model.PlaybackState
	currentTrack *model.Track
	voice        infra.VoiceConnection
}

// playerServiceImpl is the concrete PlayerService.
type playerServiceImpl struct {
	mu        sync.RWMutex
	guilds    map[string]*guildPlayer // guildID -> per-guild state
	extractor infra.AudioExtractor
	encoder   infra.AudioEncoder
	factory   infra.VoiceFactory
	queue     QueueService
}

// NewPlayerService returns a new PlayerService.
func NewPlayerService(
	extractor infra.AudioExtractor,
	encoder infra.AudioEncoder,
	factory infra.VoiceFactory,
	queue QueueService,
) PlayerService {
	return &playerServiceImpl{
		guilds:    make(map[string]*guildPlayer),
		extractor: extractor,
		encoder:   encoder,
		factory:   factory,
		queue:     queue,
	}
}

// Play resolves a track (or playlist) and either starts playback or enqueues it.
func (p *playerServiceImpl) Play(ctx context.Context, guildID, channelID, query string) (*model.PlayResult, error) {
	videoID, listID := infra.ParseYouTubeURL(query)

	var firstTrack *model.Track
	result := &model.PlayResult{}

	if listID == "" || videoID != "" {
		// Single video (or video+playlist): extract the specific video first.
		track, err := p.extractor.ExtractTrack(ctx, query)
		if err != nil {
			return nil, fmt.Errorf("extracting track: %w", err)
		}
		firstTrack = track
		result.Track = track
	}

	if listID != "" {
		// Fetch all playlist tracks (AudioURL empty — lazy extracted before playback).
		pl, err := p.extractor.ExtractPlaylist(ctx, query)
		if err != nil {
			return nil, fmt.Errorf("extracting playlist: %w", err)
		}
		result.PlaylistName = pl.PlaylistTitle

		if firstTrack == nil && len(pl.Tracks) > 0 {
			// playlist?list=Y — use first playlist track as the playing track.
			// Extract its AudioURL now so playback can start immediately.
			resolved, err := p.extractor.ExtractTrack(ctx, pl.Tracks[0].URL)
			if err != nil {
				return nil, fmt.Errorf("extracting first playlist track: %w", err)
			}
			firstTrack = resolved
			result.Track = resolved
			pl.Tracks = pl.Tracks[1:]
		} else if firstTrack != nil {
			// watch?v=X&list=Y — deduplicate: remove the video already playing.
			filtered := pl.Tracks[:0]
			for _, t := range pl.Tracks {
				tid, _ := infra.ParseYouTubeURL(t.URL)
				if tid != videoID {
					filtered = append(filtered, t)
				}
			}
			pl.Tracks = filtered
		}

		result.PlaylistCount = len(pl.Tracks)

		p.mu.Lock()
		gp, exists := p.guilds[guildID]
		if !exists {
			voice, err := p.factory.Join(ctx, guildID, channelID)
			if err != nil {
				p.mu.Unlock()
				return nil, fmt.Errorf("joining voice channel: %w", err)
			}
			gctx, cancel := context.WithCancel(context.Background())
			gp = &guildPlayer{
				ctx:         gctx,
				cancelGuild: cancel,
				cancelTrack: func() {},
				state:       model.StateIdle,
				voice:       voice,
			}
			p.guilds[guildID] = gp
			p.queue.Add(guildID, firstTrack)
			p.queue.AddAll(guildID, pl.Tracks)
			p.mu.Unlock()
			p.startPlayback(guildID, gp)
		} else {
			p.queue.Add(guildID, firstTrack)
			p.queue.AddAll(guildID, pl.Tracks)
			p.mu.Unlock()
		}
		return result, nil
	}

	// No playlist — single track path.
	p.mu.Lock()
	gp, exists := p.guilds[guildID]
	if !exists {
		voice, err := p.factory.Join(ctx, guildID, channelID)
		if err != nil {
			p.mu.Unlock()
			return nil, fmt.Errorf("joining voice channel: %w", err)
		}
		gctx, cancel := context.WithCancel(context.Background())
		gp = &guildPlayer{
			ctx:         gctx,
			cancelGuild: cancel,
			cancelTrack: func() {},
			state:       model.StateIdle,
			voice:       voice,
		}
		p.guilds[guildID] = gp
		p.queue.Add(guildID, firstTrack)
		p.mu.Unlock()
		p.startPlayback(guildID, gp)
	} else {
		p.queue.Add(guildID, firstTrack)
		p.mu.Unlock()
	}
	return result, nil
}

// startPlayback launches the per-guild playback goroutine.
// This goroutine loops through the queue until empty, then disconnects.
// It is started once per guild (on the first /play) and self-terminates when idle.
func (p *playerServiceImpl) startPlayback(guildID string, gp *guildPlayer) {
	go func() {
		// When the goroutine exits, clean up guild state.
		defer func() {
			gp.mu.Lock()
			gp.state = model.StateIdle
			gp.currentTrack = nil
			gp.mu.Unlock()

			_ = gp.voice.Disconnect(gp.ctx)

			p.mu.Lock()
			delete(p.guilds, guildID)
			p.mu.Unlock()
		}()

		for {
			// Check if the guild context was cancelled (Stop was called)
			select {
			case <-gp.ctx.Done():
				return
			default:
			}

			// Get next track from queue
			track := p.queue.Next(guildID)
			if track == nil {
				return // Queue empty — go idle
			}

			// Lazy-extract audio URL if not already done (e.g. playlist entries)
			if track.GetAudioURL() == "" {
				resolved, err := p.extractor.ExtractTrack(gp.ctx, track.URL)
				if err != nil {
					continue // Skip broken track, try next
				}
				track.SetAudioURL(resolved.GetAudioURL())
			}

			gp.mu.Lock()
			gp.currentTrack = track
			gp.state = model.StatePlaying
			gp.mu.Unlock()

			// Each track gets its own context so Skip can cancel just this track
			// without stopping the entire guild.
			trackCtx, cancelTrack := context.WithCancel(gp.ctx)
			gp.mu.Lock()
			gp.cancelTrack = cancelTrack
			gp.mu.Unlock()

			// Launch ffmpeg and stream audio
			stream, err := p.encoder.NewStream(trackCtx, track.GetAudioURL())
			if err != nil {
				cancelTrack()
				continue
			}

			// Pre-cache the next track's AudioURL while this one plays.
			p.precacheNext(gp, guildID)

			// SendAudio blocks until the track ends or trackCtx is cancelled.
			_ = gp.voice.SendAudio(trackCtx, stream)
			_ = stream.Stop()
			cancelTrack()

			// Loop back to get the next track.
		}
	}()
}

func (p *playerServiceImpl) Skip(guildID string) (*model.Track, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	gp, exists := p.guilds[guildID]

	if !exists {
		return nil, fmt.Errorf("nothing is playing")
	}

	gp.mu.Lock()
	gp.cancelTrack()
	gp.mu.Unlock()

	queue := p.queue.List(guildID)
	if len(queue) == 0 {
		return nil, nil
	}
	return queue[0], nil
}

func (p *playerServiceImpl) Pause(guildID string) error {
	// TODO: implement
	// 1. Get guildPlayer, check state is StatePlaying.
	// 2. Call gp.voice.Pause() and set gp.state = model.StatePaused.
	return fmt.Errorf("not implemented")
}

func (p *playerServiceImpl) Resume(guildID string) error {
	// TODO: implement
	// 1. Get guildPlayer, check state is StatePaused.
	// 2. Call gp.voice.Resume() and set gp.state = model.StatePlaying.
	return fmt.Errorf("not implemented")
}

func (p *playerServiceImpl) Stop(guildID string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.queue.Clear(guildID)
	gp, exists := p.guilds[guildID]
	if exists {
		gp.cancelGuild()
		delete(p.guilds, guildID)
	}
	return nil
}

func (p *playerServiceImpl) NowPlaying(guildID string) *model.Track {
	p.mu.RLock()
	gp, exists := p.guilds[guildID]
	p.mu.RUnlock()
	if !exists {
		return nil
	}

	gp.mu.Lock()
	defer gp.mu.Unlock()
	return gp.currentTrack
}

func (p *playerServiceImpl) Queue(guildID string) []*model.Track {
	return p.queue.List(guildID)
}

// precacheNext resolves the AudioURL for the next queued track in the background
// so it is ready before the current track ends, reducing the gap between songs.
func (p *playerServiceImpl) precacheNext(gp *guildPlayer, guildID string) {
	tracks := p.queue.List(guildID)
	if len(tracks) == 0 {
		return
	}
	next := tracks[0]
	if next.GetAudioURL() != "" {
		return // already cached
	}
	go func() {
		resolved, err := p.extractor.ExtractTrack(gp.ctx, next.URL)
		if err != nil {
			log.Printf("[precache] failed for %q: %v", next.URL, err)
			return
		}
		next.SetAudioURL(resolved.GetAudioURL())
		log.Printf("[precache] resolved AudioURL for %q", next.Title)
	}()
}

func (p *playerServiceImpl) IsGuildActive(guildID string) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	_, exists := p.guilds[guildID]
	return exists
}
