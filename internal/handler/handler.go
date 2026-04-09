package handler

import (
	"github.com/bwmarrin/discordgo"
	"github.com/drek/dreksbot/internal/service"
)

// Handler routes Discord slash command interactions to the correct handler function.
// It sits at the top of the layer stack: Discord → Handler → Service → Infra.
type Handler struct {
	player service.PlayerService
	queue  service.QueueService
}

// New creates a Handler wired to the given services.
func New(player service.PlayerService, queue service.QueueService) *Handler {
	return &Handler{player: player, queue: queue}
}

// commandDefinitions returns the slash commands to register with Discord.
// Add a new entry here when you implement a new command.
func commandDefinitions() []*discordgo.ApplicationCommand {
	return []*discordgo.ApplicationCommand{
		{
			Name:        "play",
			Description: "Play a YouTube URL or search by keywords",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "query",
					Description: "YouTube URL or search keywords",
					Required:    true,
				},
			},
		},
		{
			Name:        "skip",
			Description: "Skip the current track",
		},
		{
			Name:        "queue",
			Description: "Show the current queue",
		},
		{
			Name:        "pause",
			Description: "Pause playback",
		},
		{
			Name:        "resume",
			Description: "Resume playback",
		},
		{
			Name:        "stop",
			Description: "Stop playback, clear queue, and disconnect",
		},
	}
}

// RegisterCommands registers all slash commands with Discord.
// Pass guildID to register for a specific server (instant, good for development).
// Pass "" for guildID to register globally (takes up to 1 hour to propagate).
func (h *Handler) RegisterCommands(s *discordgo.Session, appID, guildID string) error {
	for _, cmd := range commandDefinitions() {
		if _, err := s.ApplicationCommandCreate(appID, guildID, cmd); err != nil {
			return err
		}
	}
	return nil
}

// OnInteraction is the discordgo event handler for all slash command interactions.
// Register it with: session.AddHandler(h.OnInteraction)
func (h *Handler) OnInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Only handle slash commands (not buttons, selects, etc.)
	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}

	switch i.ApplicationCommandData().Name {
	case "play":
		h.handlePlay(s, i)
	case "skip":
		h.handleSkip(s, i)
	case "queue":
		h.handleQueue(s, i)
	case "pause":
		h.handlePause(s, i)
	case "resume":
		h.handleResume(s, i)
	case "stop":
		h.handleStop(s, i)
	}
}

// respond sends a simple text response to an interaction.
// Use this for commands that don't need to defer (fast operations).
func respond(s *discordgo.Session, i *discordgo.InteractionCreate, content string) {
	// best-effort: once we're in the handler, there's nothing useful to do if Discord rejects the response
	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Content: content},
	})
}

// deferResponse sends an immediate "thinking..." acknowledgement.
// Call this at the start of any handler that calls yt-dlp (slow operations).
// Discord requires a response within 3 seconds or the interaction times out.
func deferResponse(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// best-effort: Discord requires a response within 3 seconds; if this fails the interaction is already lost
	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
}

// editResponse updates a previously deferred response with the actual content.
func editResponse(s *discordgo.Session, i *discordgo.InteractionCreate, content string) {
	// best-effort: if editing the deferred response fails the user sees "thinking..." but there's no recovery path
	_, _ = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &content,
	})
}
