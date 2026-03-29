package handler

import "github.com/bwmarrin/discordgo"

func (h *Handler) handleStop(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// TODO: implement
	// 1. Call h.player.Stop(i.GuildID).
	// 2. If error, respond with error.
	// 3. Respond "Stopped and disconnected."
	respond(s, i, "stop: not implemented yet")
}
