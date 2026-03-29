package handler

import (
	"github.com/bwmarrin/discordgo"
)

func (h *Handler) handleQueue(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// TODO: implement
	// 1. Get h.player.NowPlaying(i.GuildID) for the current track.
	// 2. Get h.queue.List(i.GuildID) for queued tracks.
	// 3. Build a formatted string using fmt.Sprintf and strings.Builder:
	//    "Now playing: **<title>**\nQueue:\n1. <title>\n2. <title>..."
	// 4. If nothing playing and queue empty, respond "The queue is empty."
	// Note: Queue is fast, use respond() not deferResponse().
	// You'll need to add: import ("fmt" "strings")
	respond(s, i, "queue: not implemented yet")
}
