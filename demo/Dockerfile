# syntax=docker/dockerfile:1

# build stage - compile snitch
FROM golang:1.25-trixie AS builder
WORKDIR /src
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -o snitch .

# runtime stage - official vhs image has ffmpeg, chromium, ttyd pre-installed
FROM ghcr.io/charmbracelet/vhs

# install only lightweight tools for fake services
RUN --mount=type=cache,target=/var/cache/apt,sharing=locked \
    --mount=type=cache,target=/var/lib/apt,sharing=locked \
    apt-get update --allow-releaseinfo-change && apt-get install -y --no-install-recommends \
    netcat-openbsd \
    procps \
    socat \
    nginx-light

WORKDIR /app

# copy built binary from builder
COPY --from=builder /src/snitch /app/snitch

# copy demo files
COPY demo/demo.tape /app/demo.tape
COPY demo/entrypoint.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh

ENV TERM=xterm-256color
ENV COLORTERM=truecolor

ENTRYPOINT ["/app/entrypoint.sh"]
