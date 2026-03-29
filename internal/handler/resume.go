package handler

import "github.com/bwmarrin/discordgo"

func (h *Handler) handleResume(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// TODO: implement
	// 1. Call h.player.Resume(i.GuildID).
	// 2. If error, respond with error.
	// 3. Respond "Resumed."
	respond(s, i, "resume: not implemented yet")
}
