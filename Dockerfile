# ---- Build stage ----
# Uses gcc and libopus-dev for CGO (required by layeh.com/gopus for Opus encoding)
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache gcc musl-dev libopus-dev

WORKDIR /app

# Download dependencies first (separate layer for caching)
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /bot ./cmd/bot

# ---- Runtime stage ----
# Minimal Alpine image with only what's needed to run the bot
FROM alpine:3.19

# ffmpeg: audio transcoding
# yt-dlp: YouTube audio extraction
# python3: required by yt-dlp
# libopus: required at runtime by the Go binary (gopus CGO dependency)
RUN apk add --no-cache ffmpeg python3 libopus curl \
 && curl -L https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp -o /usr/local/bin/yt-dlp \
 && chmod a+rx /usr/local/bin/yt-dlp

COPY --from=builder /bot /bot

CMD ["/bot"]
