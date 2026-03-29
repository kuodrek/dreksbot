.PHONY: build test test-integration lint docker run

# Build the bot binary locally
build:
	go build -o bot ./cmd/bot

# Run all unit tests with race detector
test:
	go test ./... -race -count=1

# Run integration tests (requires yt-dlp and ffmpeg installed locally)
test-integration:
	go test ./... -race -tags=integration -count=1

# Run linter (install: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
lint:
	golangci-lint run

# Build and run with Docker Compose
docker:
	docker-compose up --build

# Run locally (requires .env file with DISCORD_TOKEN, DISCORD_APP_ID, DEV_GUILD_ID)
run:
	./run.sh
