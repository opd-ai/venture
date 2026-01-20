#!/bin/bash
set -e

# Debian/Ubuntu package builder for Venture
# Usage: ./scripts/package-deb.sh [version] [arch]
# Example: ./scripts/package-deb.sh 1.0.0 amd64

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
BUILD_DIR="$PROJECT_ROOT/build"
DIST_DIR="$PROJECT_ROOT/dist"

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
ARCH="${2:-amd64}"

# Validate architecture
if [[ ! "$ARCH" =~ ^(amd64|arm64)$ ]]; then
    echo_error "Invalid architecture: $ARCH. Must be amd64 or arm64"
    exit 1
fi

# Map Go arch to Debian arch
DEB_ARCH="$ARCH"
if [ "$ARCH" = "arm64" ]; then
    DEB_ARCH="arm64"
fi

PACKAGE_NAME="venture"
PACKAGE_DIR="$BUILD_DIR/deb/${PACKAGE_NAME}_${VERSION}_${DEB_ARCH}"

echo_info "Building Debian package for Venture $VERSION ($ARCH)..."

# Check prerequisites
if ! command -v dpkg-deb &> /dev/null; then
    echo_warn "dpkg-deb not found, attempting to create package structure only"
    DPKG_AVAILABLE=false
else
    DPKG_AVAILABLE=true
fi

# Build binaries if they don't exist
SERVER_BIN="$BUILD_DIR/venture-server-linux-$ARCH"
CLIENT_BIN="$BUILD_DIR/venture-client-linux-$ARCH"

if [ ! -f "$SERVER_BIN" ] || [ ! -f "$CLIENT_BIN" ]; then
    echo_info "Building binaries..."
    cd "$PROJECT_ROOT"
    GOARCH="$ARCH" go build -ldflags="-s -w" -o "$SERVER_BIN" ./cmd/server
    GOARCH="$ARCH" go build -ldflags="-s -w" -o "$CLIENT_BIN" ./cmd/client
fi

# Create package directory structure
echo_info "Creating package structure..."
rm -rf "$PACKAGE_DIR"
mkdir -p "$PACKAGE_DIR/DEBIAN"
mkdir -p "$PACKAGE_DIR/usr/bin"
mkdir -p "$PACKAGE_DIR/usr/share/doc/$PACKAGE_NAME"
mkdir -p "$PACKAGE_DIR/usr/share/man/man1"
mkdir -p "$PACKAGE_DIR/etc/systemd/system"

# Copy binaries
cp "$SERVER_BIN" "$PACKAGE_DIR/usr/bin/venture-server"
cp "$CLIENT_BIN" "$PACKAGE_DIR/usr/bin/venture-client"
chmod 755 "$PACKAGE_DIR/usr/bin/venture-server"
chmod 755 "$PACKAGE_DIR/usr/bin/venture-client"

# Create control file
cat > "$PACKAGE_DIR/DEBIAN/control" << EOF
Package: $PACKAGE_NAME
Version: $VERSION
Section: games
Priority: optional
Architecture: $DEB_ARCH
Maintainer: OPD AI <dev@opd-ai.example>
Description: Venture - Procedural Multiplayer Action-RPG
 A fully procedural multiplayer action-RPG built with Go and Ebiten.
 Every aspect of the game—graphics, audio, and gameplay content—is
 generated at runtime with no external asset files.
 .
 Features include:
  - 100% procedural content generation
  - Real-time multiplayer with high-latency support
  - Cross-platform (Linux, macOS, Windows, WebAssembly)
  - Player housing and guild systems
  - Advanced physics (vehicles, fluids, destruction)
Depends: libc6, libgl1-mesa-glx | libgl1, libxcursor1, libxi6, libxinerama1, libxrandr2, libxxf86vm1, libasound2
Homepage: https://github.com/opd-ai/venture
EOF

# Create conffiles (empty, no config files yet)
touch "$PACKAGE_DIR/DEBIAN/conffiles"

# Create postinst script
cat > "$PACKAGE_DIR/DEBIAN/postinst" << 'EOF'
#!/bin/bash
set -e

# Reload systemd daemon to pick up the new service file
if [ -d /run/systemd/system ]; then
    systemctl daemon-reload || true
fi

echo "Venture installed successfully!"
echo "  - Start server: venture-server"
echo "  - Start client: venture-client"
echo "  - Enable service: sudo systemctl enable venture-server"
EOF
chmod 755 "$PACKAGE_DIR/DEBIAN/postinst"

# Create prerm script
cat > "$PACKAGE_DIR/DEBIAN/prerm" << 'EOF'
#!/bin/bash
set -e

# Stop the service if running
if [ -d /run/systemd/system ]; then
    systemctl stop venture-server || true
    systemctl disable venture-server || true
fi
EOF
chmod 755 "$PACKAGE_DIR/DEBIAN/prerm"

# Create systemd service file
cat > "$PACKAGE_DIR/etc/systemd/system/venture-server.service" << EOF
[Unit]
Description=Venture Game Server
Documentation=https://github.com/opd-ai/venture
After=network.target

[Service]
Type=simple
User=venture
Group=venture
ExecStart=/usr/bin/venture-server --port 7777 --max-players 10
Restart=on-failure
RestartSec=5
StandardOutput=journal
StandardError=journal
SyslogIdentifier=venture-server

# Security hardening
NoNewPrivileges=yes
ProtectSystem=strict
ProtectHome=yes
PrivateTmp=yes
ProtectKernelTunables=yes
ProtectKernelModules=yes
ProtectControlGroups=yes
ReadWritePaths=/var/lib/venture

[Install]
WantedBy=multi-user.target
EOF

# Create copyright file
cat > "$PACKAGE_DIR/usr/share/doc/$PACKAGE_NAME/copyright" << EOF
Format: https://www.debian.org/doc/packaging-manuals/copyright-format/1.0/
Upstream-Name: venture
Upstream-Contact: OPD AI <dev@opd-ai.example>
Source: https://github.com/opd-ai/venture

Files: *
Copyright: 2024-2026 OPD AI
License: MIT
EOF

# Copy changelog (if exists)
if [ -f "$PROJECT_ROOT/CHANGELOG.md" ]; then
    cp "$PROJECT_ROOT/CHANGELOG.md" "$PACKAGE_DIR/usr/share/doc/$PACKAGE_NAME/changelog"
    gzip -9 "$PACKAGE_DIR/usr/share/doc/$PACKAGE_NAME/changelog" 2>/dev/null || true
fi

# Create output directory
mkdir -p "$DIST_DIR/deb"

# Build the package
if [ "$DPKG_AVAILABLE" = true ]; then
    echo_info "Building .deb package..."
    dpkg-deb --build --root-owner-group "$PACKAGE_DIR" "$DIST_DIR/deb/${PACKAGE_NAME}_${VERSION}_${DEB_ARCH}.deb"
    echo_info "Package created: $DIST_DIR/deb/${PACKAGE_NAME}_${VERSION}_${DEB_ARCH}.deb"
else
    echo_warn "dpkg-deb not available, package structure created at: $PACKAGE_DIR"
    echo_warn "Run 'dpkg-deb --build $PACKAGE_DIR' on a Debian-based system to complete"
fi

echo_info "Debian package build complete!"
