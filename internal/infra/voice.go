package infra

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"sync/atomic"
	"time"

	"github.com/bwmarrin/discordgo"
	"layeh.com/gopus"
)

// VoiceConnection represents an active Discord voice channel connection for one guild.
type VoiceConnection interface {
	// SendAudio reads PCM from stream, encodes to Opus, and sends to Discord.
	// Blocks until the stream ends or ctx is cancelled (e.g. by Skip/Stop).
	SendAudio(ctx context.Context, stream AudioStream) error

	// Pause stops sending audio without disconnecting. ffmpeg's stdout buffer naturally
	// fills up, causing ffmpeg to block on write — no process restart needed on resume.
	Pause()

	// Resume restarts audio sending after a Pause.
	Resume()

	// Disconnect leaves the voice channel and cleans up.
	Disconnect(ctx context.Context) error

	// IsConnected reports whether this connection is active.
	IsConnected() bool
}

// VoiceFactory creates VoiceConnection instances by joining a voice channel.
type VoiceFactory interface {
	// Join connects to the given voice channel and returns the connection.
	// guildID and channelID are Discord snowflake IDs.
	Join(ctx context.Context, guildID, channelID string) (VoiceConnection, error)
}

// discordVoiceFactory implements VoiceFactory using discordgo.
type discordVoiceFactory struct {
	session *discordgo.Session
}

// NewDiscordVoiceFactory returns a VoiceFactory backed by a discordgo session.
func NewDiscordVoiceFactory(s *discordgo.Session) VoiceFactory {
	return &discordVoiceFactory{session: s}
}

func (f *discordVoiceFactory) Join(ctx context.Context, guildID, channelID string) (VoiceConnection, error) {
	// mute=false, deaf=true: the bot doesn't need to hear users
	vc, err := f.session.ChannelVoiceJoin(ctx, guildID, channelID, false, true)
	if err != nil {
		return nil, fmt.Errorf("joining voice channel %s in guild %s: %w", channelID, guildID, err)
	}
	return &discordVoiceConn{vc: vc}, nil
}

// discordVoiceConn implements VoiceConnection.
type discordVoiceConn struct {
	vc     *discordgo.VoiceConnection
	paused atomic.Bool
}

const (
	sampleRate   = 48000
	channels     = 2
	frameSize    = 960                      // 20ms at 48kHz
	pcmBytes     = frameSize * channels * 2 // int16 = 2 bytes, stereo
	maxOpusBytes = 1000
)

// SendAudio reads PCM from stream, encodes each 20ms frame to Opus, and sends
// to Discord via vc.OpusSend. Blocks until stream is exhausted or ctx cancelled.
func (c *discordVoiceConn) SendAudio(ctx context.Context, stream AudioStream) error {
	encoder, err := gopus.NewEncoder(sampleRate, channels, gopus.Audio)
	if err != nil {
		return fmt.Errorf("creating Opus encoder: %w", err)
	}

	_ = c.vc.Speaking(true)
	time.Sleep(2 * time.Second)
	defer c.vc.Speaking(false)

	pcmBuf := make([]byte, pcmBytes)
	pcm := make([]int16, frameSize*channels)

	for {
		// Check for cancellation (Skip or Stop)
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// If paused, sleep briefly. ffmpeg's stdout buffer naturally fills up,
		// causing ffmpeg to block on write — no process restart needed on resume.
		if c.paused.Load() {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		// Read exactly one 20ms PCM frame from ffmpeg
		_, err := io.ReadFull(stream, pcmBuf)

		if err == io.EOF || err == io.ErrUnexpectedEOF {
			return nil // Track ended naturally
		}
		if err != nil {
			return fmt.Errorf("reading PCM from ffmpeg: %w", err)
		}

		// Convert []byte (s16le) to []int16
		for i := range pcm {
			pcm[i] = int16(binary.LittleEndian.Uint16(pcmBuf[i*2:]))
		}

		// Encode to Opus
		opusFrame, err := encoder.Encode(pcm, frameSize, maxOpusBytes)

		if err != nil {
			return fmt.Errorf("encoding Opus frame: %w", err)
		}

		// Send to Discord (non-blocking with ctx check)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case c.vc.OpusSend <- opusFrame:
		}
	}
}

func (c *discordVoiceConn) Pause() {
	c.paused.Store(true)
}

func (c *discordVoiceConn) Resume() {
	c.paused.Store(false)
}

func (c *discordVoiceConn) Disconnect(ctx context.Context) error {
	return c.vc.Disconnect(ctx)
}

func (c *discordVoiceConn) IsConnected() bool {
	return c.vc.Status == discordgo.VoiceConnectionStatusReady
}
