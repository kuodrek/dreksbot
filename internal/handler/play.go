package handler

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
)

// handlePlay implements /play <query>.
//
// Flow:
//  1. Immediately defer the response (Discord has a 3-second timeout).
//  2. Check the user is in a voice channel.
//  3. Call PlayerService.Play — this calls yt-dlp (slow) and starts audio.
//  4. Edit the deferred response with the result.
//
// This is the reference implementation showing the full handler pattern.
// Study this when implementing the other handlers.
func (h *Handler) handlePlay(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Step 1: Acknowledge immediately — "Bot is thinking..."
	// This buys us time to call yt-dlp, which can take 1-3 seconds.
	deferResponse(s, i)

	// Step 2: Get the query from the slash command options
	query := i.ApplicationCommandData().Options[0].StringValue()

	// Step 3: Find which voice channel the user is currently in
	// s.State is a cache of guild state maintained by discordgo.
	vs, err := s.State.VoiceState(i.GuildID, i.Member.User.ID)
	if err != nil || vs == nil {
		editResponse(s, i, "You need to be in a voice channel to use /play.")
		return
	}

	// Step 4: Tell the PlayerService to play
	// Play() calls yt-dlp, joins voice if needed, and starts the playback goroutine.
	track, err := h.player.Play(context.Background(), i.GuildID, vs.ChannelID, query)
	if err != nil {
		editResponse(s, i, fmt.Sprintf("Error: %v", err))
		return
	}

	// Step 5: Edit the deferred response with the result
	editResponse(s, i, fmt.Sprintf("Now playing: **%s**", track.Title))
}
