package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/drek/dreksbot/config"
	"github.com/drek/dreksbot/internal/handler"
	"github.com/drek/dreksbot/internal/infra"
	"github.com/drek/dreksbot/internal/service"
)

func main() {
	// --- Load configuration ---
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	// --- Create Discord session ---
	session, err := discordgo.New("Bot " + cfg.DiscordToken)
	if err != nil {
		log.Fatalf("creating Discord session: %v", err)
	}

	// Request the voice state intent so we can see which voice channel users are in.
	session.Identify.Intents = discordgo.IntentsGuilds | discordgo.IntentsGuildVoiceStates

	// --- Wire infrastructure layer ---
	extractor := infra.NewYTDLPExtractor()
	encoder := infra.NewFFmpegEncoder()
	voiceFactory := infra.NewDiscordVoiceFactory(session)

	// --- Wire service layer ---
	queueSvc := service.NewQueueService()
	playerSvc := service.NewPlayerService(extractor, encoder, voiceFactory, queueSvc)

	// --- Wire handler layer ---
	h := handler.New(playerSvc, queueSvc)

	// Register the interaction handler before opening the session
	session.AddHandler(h.OnInteraction)

	// --- Connect to Discord ---
	if err := session.Open(); err != nil {
		log.Fatalf("opening Discord session: %v", err)
	}
	defer session.Close()

	// --- Register slash commands ---
	// DEV_GUILD_ID registers commands to a specific server instantly (great for testing).
	// Empty string registers globally (takes up to 1 hour).
	if err := h.RegisterCommands(session, cfg.AppID, cfg.DevGuildID); err != nil {
		log.Fatalf("registering commands: %v", err)
	}

	log.Printf("Bot is running. Press Ctrl+C to stop.")
	if cfg.DevGuildID != "" {
		log.Printf("Commands registered to guild %s (dev mode)", cfg.DevGuildID)
	} else {
		log.Printf("Commands registered globally (may take up to 1 hour to appear)")
	}

	// Block until Ctrl+C or SIGTERM
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("Shutting down...")
}
