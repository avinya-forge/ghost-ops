# syntax=docker/dockerfile:1.7

# ---------- Build stage ----------
FROM golang:1.24-alpine AS builder

WORKDIR /src

# Cache module downloads.
COPY go.mod go.sum ./
RUN go mod download

# Build static binary with version baked in from /VERSION.
COPY . .
RUN CGO_ENABLED=0 GOOS=linux \
    go build \
      -ldflags "-s -w -X main.Version=$(cat VERSION)" \
      -trimpath \
      -o /out/ghost-ops \
      ./cmd/ghost-ops

# ---------- Runtime stage ----------
FROM alpine:3.20

# ca-certificates for HTTPS to LLM providers; wget for the HEALTHCHECK.
# Then create a non-root user with no shell.
RUN apk add --no-cache ca-certificates wget && \
    addgroup -S ghost && adduser -S -G ghost -h /home/ghost ghost

WORKDIR /home/ghost
COPY --from=builder /out/ghost-ops /usr/local/bin/ghost-ops

# Default config + state directories, owned by the non-root user.
RUN mkdir -p /home/ghost/blueprints /home/ghost/state && \
    chown -R ghost:ghost /home/ghost

USER ghost
EXPOSE 8080

# Healthy when the version subcommand exits 0 (cheap readiness probe).
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD ghost-ops version >/dev/null 2>&1 || exit 1

ENTRYPOINT ["ghost-ops"]
CMD ["-engine", "mock", "-llm", "mock"]
