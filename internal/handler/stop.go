package handler

import (
	"github.com/bwmarrin/discordgo"
)

func (h *Handler) handleStop(s *discordgo.Session, i *discordgo.InteractionCreate) {
	guildId := i.GuildID
	_ = h.player.Stop(guildId)
	respond(s, i, "Bot stopped playing")
}
