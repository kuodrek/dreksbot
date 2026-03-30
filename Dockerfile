# ---- Build stage ----
# Uses gcc and libopus-dev for CGO (required by layeh.com/gopus for Opus encoding)
FROM golang:1.26.1 AS builder

RUN apt-get update && apt-get install -y gcc libopus-dev

WORKDIR /app

# Download dependencies first (separate layer for caching)
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /bot ./cmd/bot

# ---- Runtime stage ----
# Debian slim to match the glibc builder stage
FROM debian:bookworm-slim

# ffmpeg: audio transcoding
# yt-dlp: YouTube audio extraction
# python3: required by yt-dlp
# libopus0: required at runtime by the Go binary (gopus CGO dependency)
RUN apt-get update && apt-get install -y --no-install-recommends \
    ffmpeg python3 libopus0 curl ca-certificates \
 && curl -L https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp -o /usr/local/bin/yt-dlp \
 && chmod a+rx /usr/local/bin/yt-dlp \
 && apt-get clean && rm -rf /var/lib/apt/lists/*

COPY --from=builder /bot /bot

CMD ["/bot"]
