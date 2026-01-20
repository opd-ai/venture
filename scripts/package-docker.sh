#!/bin/bash
set -e

# Docker image builder for Venture
# Usage: ./scripts/package-docker.sh [version] [registry]
# Example: ./scripts/package-docker.sh 1.0.0 ghcr.io/opd-ai

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
echo_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
echo_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# Parse arguments
VERSION="${1:-1.0.0}"
REGISTRY="${2:-ghcr.io/opd-ai}"

IMAGE_NAME="venture-server"
FULL_IMAGE="$REGISTRY/$IMAGE_NAME"

echo_info "Building Docker image: $FULL_IMAGE:$VERSION"

# Check prerequisites
if ! command -v docker &> /dev/null; then
    echo_error "Docker is not installed"
    exit 1
fi

cd "$PROJECT_ROOT"

# Create Dockerfile if it doesn't exist
if [ ! -f "Dockerfile" ]; then
    echo_info "Creating Dockerfile..."
    cat > Dockerfile << 'DOCKERFILE_EOF'
# Multi-stage build for Venture server
FROM golang:1.24-bookworm AS builder

# Install build dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    gcc \
    libc6-dev \
    libgl1-mesa-dev \
    libxcursor-dev \
    libxi-dev \
    libxinerama-dev \
    libxrandr-dev \
    libxxf86vm-dev \
    libasound2-dev \
    pkg-config \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /venture-server ./cmd/server

# Runtime image
FROM debian:bookworm-slim

# Install runtime dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Create non-root user
RUN groupadd -r venture && useradd -r -g venture venture
RUN mkdir -p /var/lib/venture && chown venture:venture /var/lib/venture

WORKDIR /app
COPY --from=builder /venture-server .

USER venture

EXPOSE 7777/tcp
EXPOSE 7777/udp

VOLUME ["/var/lib/venture"]

ENTRYPOINT ["./venture-server"]
CMD ["--port", "7777", "--max-players", "10"]
DOCKERFILE_EOF
fi

# Build the image
echo_info "Building Docker image..."
docker build \
    --build-arg VERSION="$VERSION" \
    -t "$FULL_IMAGE:$VERSION" \
    -t "$FULL_IMAGE:latest" \
    .

echo_info "Docker image built successfully!"
echo_info "  $FULL_IMAGE:$VERSION"
echo_info "  $FULL_IMAGE:latest"

# Test the image
echo_info "Testing Docker image..."
docker run --rm "$FULL_IMAGE:$VERSION" --version

echo ""
echo_info "To push to registry:"
echo "  docker push $FULL_IMAGE:$VERSION"
echo "  docker push $FULL_IMAGE:latest"
echo ""
echo_info "To run locally:"
echo "  docker run -p 7777:7777 $FULL_IMAGE:$VERSION"
