# Build stage
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache gcc musl-dev sqlite-dev

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Copy source code (needed for go mod tidy)
COPY . .

# Tidy up dependencies and download
RUN go mod tidy
RUN go mod download

# Build binaries
RUN CGO_ENABLED=1 go build -ldflags="-s -w" -o bin/curltree-server ./cmd/server
RUN CGO_ENABLED=1 go build -ldflags="-s -w" -o bin/curltree-tui ./cmd/tui

# Production stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache ca-certificates sqlite openssh-keygen

WORKDIR /app

# Copy binaries from build stage
COPY --from=builder /app/bin/curltree-server ./
COPY --from=builder /app/bin/curltree-tui ./

# Copy configuration example
COPY --from=builder /app/config.example.json ./config.json

# Create directories
RUN mkdir -p .ssh data

# Generate SSH host key
RUN ssh-keygen -t rsa -b 4096 -f .ssh/curltree_host_key -N "" -C "curltree-host-key"

# Set environment variables
ENV DB_PATH=/app/data/curltree.db
ENV HOST_KEY_PATH=/app/.ssh/curltree_host_key
ENV SERVER_HOST=0.0.0.0
ENV SSH_HOST=0.0.0.0

# Expose ports
EXPOSE 8080 23234

# Create startup script
RUN echo '#!/bin/sh' > start.sh && \
    echo './curltree-server &' >> start.sh && \
    echo './curltree-tui &' >> start.sh && \
    echo 'wait' >> start.sh && \
    chmod +x start.sh

CMD ["./start.sh"]