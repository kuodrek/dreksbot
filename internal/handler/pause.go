package handler

import "github.com/bwmarrin/discordgo"

func (h *Handler) handlePause(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// TODO: implement
	// 1. Call h.player.Pause(i.GuildID).
	// 2. If error, respond with error.
	// 3. Respond "Paused."
	respond(s, i, "pause: not implemented yet")
}
