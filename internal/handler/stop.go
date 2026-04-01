package handler

import (
	"github.com/bwmarrin/discordgo"
)

func (h *Handler) handleStop(s *discordgo.Session, i *discordgo.InteractionCreate) {
	guildId := i.GuildID
	err := h.player.Stop(guildId)
	if err != nil {
		respond(s, i, "Failed stop operation. Call dreks if hes available")
		return
	}
	respond(s, i, "Bot stopped playing")
}
