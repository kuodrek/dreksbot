# Deployment

## Overview

On every push to `main`, GitHub Actions builds a Docker image, pushes it to GitHub Container Registry (GHCR), and deploys it to an EC2 instance via SSH.

```
push to main → GitHub Actions → build image → push to GHCR → SSH into EC2 → pull & restart
```

Each deploy is tagged with a version like `v2026.03.29-a1b2c3d` (date + commit SHA).

## How it works

1. Code is pushed/merged to `main`
2. GitHub Actions (`.github/workflows/deploy.yml`) runs:
   - Builds the Docker image using the multi-stage `Dockerfile`
   - Pushes it to `ghcr.io/kuodrek/dreksbot:latest`
   - Creates and pushes a git tag
   - SSHs into EC2 and pulls the new image
   - Restarts the container with `docker compose -f docker-compose.prod.yml up -d --force-recreate`

## GitHub Secrets required

| Secret | Description |
|--------|-------------|
| `EC2_HOST` | EC2 instance public IP |
| `EC2_SSH_KEY` | Contents of the `.pem` key file |
| `DISCORD_TOKEN` | Discord bot token |
| `DISCORD_APP_ID` | Discord application ID |

Configure at: repo Settings → Secrets and variables → Actions → Repository secrets.

## EC2 one-time setup

```bash
# Install Docker
sudo dnf install -y docker
sudo systemctl enable --now docker
sudo usermod -aG docker ec2-user
# Log out and back in for group change to take effect

# Install Docker Compose plugin
sudo mkdir -p /usr/local/lib/docker/cli-plugins
sudo curl -SL https://github.com/docker/compose/releases/latest/download/docker-compose-linux-x86_64 \
  -o /usr/local/lib/docker/cli-plugins/docker-compose
sudo chmod +x /usr/local/lib/docker/cli-plugins/docker-compose

# Create app directory
mkdir -p ~/dreksbot

# Log in to GHCR (needs a classic GitHub PAT with read:packages scope)
echo YOUR_PAT | docker login ghcr.io -u kuodrek --password-stdin
```

## Docker image

The `Dockerfile` uses a multi-stage build:

- **Builder stage** (`golang:1.26.1`): compiles the Go binary with CGO (needed for opus encoding)
- **Runtime stage** (`alpine:3.19`): minimal image with ffmpeg, yt-dlp, python3, and opus

## Files

| File | Purpose |
|------|---------|
| `.github/workflows/deploy.yml` | CI/CD pipeline |
| `docker-compose.prod.yml` | Production compose (pulls from GHCR) |
| `docker-compose.yml` | Local development (builds from Dockerfile) |
| `Dockerfile` | Multi-stage Docker build |

## Useful commands (on EC2)

```bash
# Check running containers
docker ps

# View bot logs
docker compose -f ~/dreksbot/docker-compose.prod.yml logs -f

# Restart manually
cd ~/dreksbot && docker compose -f docker-compose.prod.yml up -d --force-recreate

# Stop the bot
cd ~/dreksbot && docker compose -f docker-compose.prod.yml down
```
