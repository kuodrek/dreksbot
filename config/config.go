package config

import (
	"fmt"
	"os"
)

// Config holds all configuration loaded from environment variables.
type Config struct {
	DiscordToken string // Required. Bot token from Discord Developer Portal.
	AppID        string // Required. Application ID from Discord Developer Portal.
	DevGuildID   string // Optional. Register commands to this guild during development (instant).
	              //         Leave empty to register globally (takes up to 1 hour).
}

// Load reads configuration from environment variables.
// Returns an error if any required variable is missing.
func Load() (*Config, error) {
	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("DISCORD_TOKEN environment variable is required")
	}

	appID := os.Getenv("DISCORD_APP_ID")
	if appID == "" {
		return nil, fmt.Errorf("DISCORD_APP_ID environment variable is required")
	}

	return &Config{
		DiscordToken: token,
		AppID:        appID,
		DevGuildID:   os.Getenv("DEV_GUILD_ID"),
	}, nil
}
