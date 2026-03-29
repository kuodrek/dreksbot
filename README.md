# dreksbot (now in go!)

- [invite bot with this link](https://discord.com/oauth2/authorize?client_id=1374531119357100103&scope=bot&permissions=1126450236246016)

## Features

- `/play` - Play audio from YouTube in a voice channel (search by name or paste a URL)

Other features will be available in future releases.

## Known issues

- Using a [forked version of discordgo](https://github.com/yeongaori/discordgo-fork) because the upstream package does not yet support Discord's DAVE protocol (E2EE for voice channels). The fork includes a fix for the DAVE/E2EE handshake.

## How to run bot

### Locally

- Run with go 1.26.1
- Install dependencies in environment (yt-dlp, ffmpeg)
- Create `.env` with credentials (see `.env.example`)
- Run `bash run.sh`

### Production

Runs on an EC2 instance via Docker. On every push to `main`, GitHub Actions builds a Docker image, pushes it to GHCR, and deploys it automatically. See [docs/deployment.md](docs/deployment.md) for details.
