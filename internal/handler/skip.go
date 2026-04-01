package handler

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

func (h *Handler) handleSkip(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// TODO: implement
	// 1. Call h.player.Skip(i.GuildID).
	// 2. If error (e.g. "not implemented" or "nothing playing"), respond with error.
	// 3. If next track is nil (queue empty), respond "Queue is now empty."
	// 4. Otherwise respond "Skipped! Now playing: **<track.Title>**"
	// Note: Skip is fast (no yt-dlp call), so use respond() not deferResponse().
	guildId := i.GuildID
	nextTrack, err := h.player.Skip(guildId)
	if err != nil {
		respond(s, i, "Failed skip operation. Call dreks if hes available")
		return
	}
	respond(s, i, fmt.Sprintf("Track skipped. Now playing: **%s**", nextTrack))
}
