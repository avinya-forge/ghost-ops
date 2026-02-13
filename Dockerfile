# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /ghost-ops cmd/ghost-ops/main.go

# Final stage
FROM alpine:latest

WORKDIR /root/

COPY --from=builder /ghost-ops .

CMD ["./ghost-ops"]
