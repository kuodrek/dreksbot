package handler

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func (h *Handler) handleQueue(s *discordgo.Session, i *discordgo.InteractionCreate) {
	current := h.player.NowPlaying(i.GuildID)
	queue := h.player.Queue(i.GuildID)

	if current == nil && len(queue) == 0 {
		respond(s, i, "The queue is empty.")
		return
	}

	var b strings.Builder
	if current != nil {
		fmt.Fprintf(&b, "Now playing: **%s**\n", current.Title)
	}
	for i, track := range queue {
		fmt.Fprintf(&b, "%d. %s\n", i+1, track.Title)
	}
	respond(s, i, b.String())
}
