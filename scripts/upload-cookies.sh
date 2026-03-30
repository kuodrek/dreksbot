#!/usr/bin/env bash
set -euo pipefail

# Upload a cookies.txt file to the EC2 instance running dreksbot.
# Usage: ./scripts/upload-cookies.sh <cookies.txt> <ec2-host>
#
# The EC2 host can also be set via the EC2_HOST environment variable.

COOKIES_FILE="${1:?Usage: $0 <cookies.txt> [ec2-host]}"
EC2_HOST="${2:-${EC2_HOST:?EC2 host not provided. Pass as second arg or set EC2_HOST env var.}}"

if [ ! -f "$COOKIES_FILE" ]; then
    echo "Error: $COOKIES_FILE not found"
    exit 1
fi

scp -i $EC2PEM "$COOKIES_FILE" "ec2-user@${EC2_HOST}:~/dreksbot/cookies.txt"
echo "Cookies uploaded. Restart the container to pick them up:"
echo "  ssh ec2-user@${EC2_HOST} 'cd ~/dreksbot && docker compose -f docker-compose.prod.yml restart'"
