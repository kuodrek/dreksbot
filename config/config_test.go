package config_test

import (
	"testing"

	"github.com/drek/dreksbot/config"
)

func TestLoad_MissingToken_ReturnsError(t *testing.T) {
	t.Setenv("DISCORD_TOKEN", "")
	t.Setenv("DISCORD_APP_ID", "123456")
	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error when DISCORD_TOKEN is missing, got nil")
	}
}

func TestLoad_MissingAppID_ReturnsError(t *testing.T) {
	t.Setenv("DISCORD_TOKEN", "token123")
	t.Setenv("DISCORD_APP_ID", "")
	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error when DISCORD_APP_ID is missing, got nil")
	}
}

func TestLoad_AllRequired_Succeeds(t *testing.T) {
	t.Setenv("DISCORD_TOKEN", "token123")
	t.Setenv("DISCORD_APP_ID", "appid123")
	t.Setenv("DEV_GUILD_ID", "")
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.DiscordToken != "token123" {
		t.Errorf("expected token123, got %s", cfg.DiscordToken)
	}
	if cfg.AppID != "appid123" {
		t.Errorf("expected appid123, got %s", cfg.AppID)
	}
}

func TestLoad_DevGuildID_IsOptional(t *testing.T) {
	t.Setenv("DISCORD_TOKEN", "tok")
	t.Setenv("DISCORD_APP_ID", "app")
	t.Setenv("DEV_GUILD_ID", "guild123")
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.DevGuildID != "guild123" {
		t.Errorf("expected guild123, got %s", cfg.DevGuildID)
	}
}
