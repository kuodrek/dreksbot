#!/bin/bash
# Local development runner.
# Prerequisites:
#   - Go 1.22+
#   - ffmpeg installed (apt install ffmpeg / brew install ffmpeg)
#   - yt-dlp installed (pip install yt-dlp or apt install yt-dlp)
#   - libopus-dev installed (apt install libopus-dev / brew install opus)
#   - .env file copied from .env.example with real values

set -e

if [ ! -f .env ]; then
  echo "Error: .env file not found. Copy .env.example to .env and fill in your values."
  exit 1
fi

source .env

if [ -z "$DISCORD_TOKEN" ]; then
  echo "Error: DISCORD_TOKEN is not set in .env"
  exit 1
fi

echo "Starting dreksbot in development mode..."
if [ -n "$DEV_GUILD_ID" ]; then
  echo "Commands will register to guild $DEV_GUILD_ID (instant)"
else
  echo "Commands will register globally (up to 1 hour to propagate)"
fi

go run ./cmd/bot
