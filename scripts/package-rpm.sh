#!/bin/bash
set -e

# RHEL/Fedora RPM package builder for Venture
# Usage: ./scripts/package-rpm.sh [version] [arch]
# Example: ./scripts/package-rpm.sh 1.0.0 amd64

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

# Map Go arch to RPM arch
RPM_ARCH="x86_64"
if [ "$ARCH" = "arm64" ]; then
    RPM_ARCH="aarch64"
fi

PACKAGE_NAME="venture"
RPM_BUILD_DIR="$BUILD_DIR/rpmbuild"

echo_info "Building RPM package for Venture $VERSION ($ARCH)..."

# Check prerequisites
if ! command -v rpmbuild &> /dev/null; then
    echo_warn "rpmbuild not found, creating spec file only"
    RPMBUILD_AVAILABLE=false
else
    RPMBUILD_AVAILABLE=true
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

# Create RPM build directory structure
echo_info "Creating RPM build structure..."
rm -rf "$RPM_BUILD_DIR"
mkdir -p "$RPM_BUILD_DIR"/{BUILD,RPMS,SOURCES,SPECS,SRPMS}
mkdir -p "$RPM_BUILD_DIR/BUILDROOT/${PACKAGE_NAME}-${VERSION}-1.${RPM_ARCH}"

# Create source tarball
SOURCE_DIR="$RPM_BUILD_DIR/SOURCES/${PACKAGE_NAME}-${VERSION}"
mkdir -p "$SOURCE_DIR"
cp "$SERVER_BIN" "$SOURCE_DIR/venture-server"
cp "$CLIENT_BIN" "$SOURCE_DIR/venture-client"

# Create systemd service file in sources
cat > "$SOURCE_DIR/venture-server.service" << EOF
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

NoNewPrivileges=yes
ProtectSystem=strict
ProtectHome=yes
PrivateTmp=yes

[Install]
WantedBy=multi-user.target
EOF

# Create tarball
cd "$RPM_BUILD_DIR/SOURCES"
tar czf "${PACKAGE_NAME}-${VERSION}.tar.gz" "${PACKAGE_NAME}-${VERSION}"
rm -rf "${PACKAGE_NAME}-${VERSION}"

# Create spec file
cat > "$RPM_BUILD_DIR/SPECS/${PACKAGE_NAME}.spec" << EOF
Name:           ${PACKAGE_NAME}
Version:        ${VERSION}
Release:        1%{?dist}
Summary:        Procedural Multiplayer Action-RPG

License:        MIT
URL:            https://github.com/opd-ai/venture
Source0:        %{name}-%{version}.tar.gz

BuildArch:      ${RPM_ARCH}
Requires:       mesa-libGL
Requires:       libXcursor
Requires:       libXi
Requires:       libXinerama
Requires:       libXrandr
Requires:       libXxf86vm
Requires:       alsa-lib

%description
Venture is a fully procedural multiplayer action-RPG built with Go and Ebiten.
Every aspect of the game—graphics, audio, and gameplay content—is generated
at runtime with no external asset files.

Features include:
- 100% procedural content generation
- Real-time multiplayer with high-latency support
- Cross-platform (Linux, macOS, Windows, WebAssembly)
- Player housing and guild systems
- Advanced physics (vehicles, fluids, destruction)

%prep
%setup -q

%build
# Binaries are pre-built

%install
rm -rf %{buildroot}
mkdir -p %{buildroot}%{_bindir}
mkdir -p %{buildroot}%{_unitdir}
mkdir -p %{buildroot}%{_sharedstatedir}/%{name}

install -m 755 venture-server %{buildroot}%{_bindir}/
install -m 755 venture-client %{buildroot}%{_bindir}/
install -m 644 venture-server.service %{buildroot}%{_unitdir}/

%pre
getent group venture >/dev/null || groupadd -r venture
getent passwd venture >/dev/null || \\
    useradd -r -g venture -d %{_sharedstatedir}/%{name} -s /sbin/nologin \\
    -c "Venture Game Server" venture
exit 0

%post
%systemd_post venture-server.service
echo "Venture installed successfully!"
echo "  - Start server: venture-server"
echo "  - Start client: venture-client"
echo "  - Enable service: sudo systemctl enable venture-server"

%preun
%systemd_preun venture-server.service

%postun
%systemd_postun_with_restart venture-server.service

%files
%{_bindir}/venture-server
%{_bindir}/venture-client
%{_unitdir}/venture-server.service
%dir %attr(755, venture, venture) %{_sharedstatedir}/%{name}

%changelog
* $(date "+%a %b %d %Y") OPD AI <dev@opd-ai.example> - ${VERSION}-1
- Initial release
EOF

# Create output directory
mkdir -p "$DIST_DIR/rpm"

# Build the package
if [ "$RPMBUILD_AVAILABLE" = true ]; then
    echo_info "Building .rpm package..."
    rpmbuild --define "_topdir $RPM_BUILD_DIR" \
             --define "_rpmdir $DIST_DIR/rpm" \
             -bb "$RPM_BUILD_DIR/SPECS/${PACKAGE_NAME}.spec"
    
    # Move package to expected location
    find "$DIST_DIR/rpm" -name "*.rpm" -exec mv {} "$DIST_DIR/rpm/${PACKAGE_NAME}-${VERSION}-1.${RPM_ARCH}.rpm" \; 2>/dev/null || true
    echo_info "Package created: $DIST_DIR/rpm/${PACKAGE_NAME}-${VERSION}-1.${RPM_ARCH}.rpm"
else
    echo_warn "rpmbuild not available, spec file created at: $RPM_BUILD_DIR/SPECS/${PACKAGE_NAME}.spec"
    cp "$RPM_BUILD_DIR/SPECS/${PACKAGE_NAME}.spec" "$DIST_DIR/rpm/"
    echo_warn "Run 'rpmbuild -bb $DIST_DIR/rpm/${PACKAGE_NAME}.spec' on an RPM-based system to complete"
fi

echo_info "RPM package build complete!"
