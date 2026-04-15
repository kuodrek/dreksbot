package handler

import "github.com/bwmarrin/discordgo"

func (h *Handler) handlePlaylist(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// TODO: implement
	// Pattern is identical to handlePlay — defer first, then do work.
	// 1. deferResponse(s, i) — ExtractPlaylist is slow (calls yt-dlp).
	// 2. Get URL from options: i.ApplicationCommandData().Options[0].StringValue()
	// 3. Get user's voice channel (same s.State.VoiceState pattern as handlePlay).
	// 4. Call h.player.PlayPlaylist(context.Background(), i.GuildID, vs.ChannelID, url).
	// 5. editResponse with fmt.Sprintf("Added %d tracks from playlist.", len(tracks))
	// You'll need to add: import ("context" "fmt")
	deferResponse(s, i)
	editResponse(s, i, "playlist: not implemented yet")
}
