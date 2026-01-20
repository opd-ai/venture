#!/bin/bash
set -e

# Windows NSIS installer builder for Venture
# Usage: ./scripts/package-windows.sh [version]
# Example: ./scripts/package-windows.sh 1.0.0
#
# Prerequisites: NSIS (makensis) or Wine with NSIS
# On Ubuntu: apt install nsis

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
ARCH="amd64"

PACKAGE_NAME="venture"
NSIS_DIR="$BUILD_DIR/nsis"

echo_info "Building Windows installer for Venture $VERSION..."

# Check prerequisites
if ! command -v makensis &> /dev/null; then
    echo_warn "makensis not found, creating NSIS script only"
    NSIS_AVAILABLE=false
else
    NSIS_AVAILABLE=true
fi

# Build binaries if they don't exist
SERVER_BIN="$BUILD_DIR/venture-server-windows-amd64.exe"
CLIENT_BIN="$BUILD_DIR/venture-client-windows-amd64.exe"

if [ ! -f "$SERVER_BIN" ] || [ ! -f "$CLIENT_BIN" ]; then
    echo_info "Building Windows binaries..."
    cd "$PROJECT_ROOT"
    GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o "$SERVER_BIN" ./cmd/server
    GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o "$CLIENT_BIN" ./cmd/client
fi

# Create NSIS directory
echo_info "Creating NSIS installer files..."
rm -rf "$NSIS_DIR"
mkdir -p "$NSIS_DIR"

# Copy binaries
cp "$SERVER_BIN" "$NSIS_DIR/venture-server.exe"
cp "$CLIENT_BIN" "$NSIS_DIR/venture-client.exe"

# Create NSIS script
cat > "$NSIS_DIR/venture.nsi" << 'NSIS_EOF'
!include "MUI2.nsh"

; Installer attributes
Name "Venture"
OutFile "venture-${VERSION}-setup.exe"
InstallDir "$PROGRAMFILES64\Venture"
InstallDirRegKey HKLM "Software\Venture" "InstallDir"
RequestExecutionLevel admin

; Metadata
!define PRODUCT_NAME "Venture"
!define PRODUCT_VERSION "${VERSION}"
!define PRODUCT_PUBLISHER "OPD AI"
!define PRODUCT_WEB_SITE "https://github.com/opd-ai/venture"

; MUI Settings
!define MUI_ABORTWARNING
!define MUI_ICON "${NSISDIR}\Contrib\Graphics\Icons\modern-install.ico"
!define MUI_UNICON "${NSISDIR}\Contrib\Graphics\Icons\modern-uninstall.ico"

; Installer pages
!insertmacro MUI_PAGE_WELCOME
!insertmacro MUI_PAGE_LICENSE "LICENSE.txt"
!insertmacro MUI_PAGE_DIRECTORY
!insertmacro MUI_PAGE_INSTFILES
!insertmacro MUI_PAGE_FINISH

; Uninstaller pages
!insertmacro MUI_UNPAGE_CONFIRM
!insertmacro MUI_UNPAGE_INSTFILES

; Language
!insertmacro MUI_LANGUAGE "English"

Section "Venture" SecMain
    SetOutPath "$INSTDIR"
    
    ; Install files
    File "venture-server.exe"
    File "venture-client.exe"
    
    ; Create Start Menu shortcuts
    CreateDirectory "$SMPROGRAMS\Venture"
    CreateShortCut "$SMPROGRAMS\Venture\Venture Client.lnk" "$INSTDIR\venture-client.exe"
    CreateShortCut "$SMPROGRAMS\Venture\Venture Server.lnk" "$INSTDIR\venture-server.exe"
    CreateShortCut "$SMPROGRAMS\Venture\Uninstall.lnk" "$INSTDIR\uninstall.exe"
    
    ; Create Desktop shortcut (optional)
    CreateShortCut "$DESKTOP\Venture.lnk" "$INSTDIR\venture-client.exe"
    
    ; Write registry keys
    WriteRegStr HKLM "Software\Venture" "InstallDir" "$INSTDIR"
    WriteRegStr HKLM "Software\Venture" "Version" "${PRODUCT_VERSION}"
    
    ; Add to PATH (optional)
    EnVar::AddValue "PATH" "$INSTDIR"
    
    ; Create uninstaller
    WriteUninstaller "$INSTDIR\uninstall.exe"
    
    ; Add uninstall information to Add/Remove Programs
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\Venture" \
                     "DisplayName" "Venture"
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\Venture" \
                     "UninstallString" "$\"$INSTDIR\uninstall.exe$\""
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\Venture" \
                     "InstallLocation" "$INSTDIR"
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\Venture" \
                     "Publisher" "${PRODUCT_PUBLISHER}"
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\Venture" \
                     "DisplayVersion" "${PRODUCT_VERSION}"
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\Venture" \
                     "URLInfoAbout" "${PRODUCT_WEB_SITE}"
    WriteRegDWORD HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\Venture" \
                      "NoModify" 1
    WriteRegDWORD HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\Venture" \
                      "NoRepair" 1
SectionEnd

Section "Uninstall"
    ; Remove files
    Delete "$INSTDIR\venture-server.exe"
    Delete "$INSTDIR\venture-client.exe"
    Delete "$INSTDIR\uninstall.exe"
    
    ; Remove shortcuts
    Delete "$SMPROGRAMS\Venture\Venture Client.lnk"
    Delete "$SMPROGRAMS\Venture\Venture Server.lnk"
    Delete "$SMPROGRAMS\Venture\Uninstall.lnk"
    RMDir "$SMPROGRAMS\Venture"
    Delete "$DESKTOP\Venture.lnk"
    
    ; Remove from PATH
    EnVar::DeleteValue "PATH" "$INSTDIR"
    
    ; Remove registry keys
    DeleteRegKey HKLM "Software\Venture"
    DeleteRegKey HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\Venture"
    
    ; Remove install directory
    RMDir "$INSTDIR"
SectionEnd
NSIS_EOF

# Replace version placeholder in NSIS script
sed -i "s/\${VERSION}/$VERSION/g" "$NSIS_DIR/venture.nsi"

# Create license file for installer
if [ -f "$PROJECT_ROOT/LICENSE" ]; then
    cp "$PROJECT_ROOT/LICENSE" "$NSIS_DIR/LICENSE.txt"
else
    cat > "$NSIS_DIR/LICENSE.txt" << EOF
MIT License

Copyright (c) 2024-2026 OPD AI

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
EOF
fi

# Create output directory
mkdir -p "$DIST_DIR/windows"

# Build the installer
if [ "$NSIS_AVAILABLE" = true ]; then
    echo_info "Building Windows installer..."
    cd "$NSIS_DIR"
    makensis venture.nsi
    mv "venture-${VERSION}-setup.exe" "$DIST_DIR/windows/"
    echo_info "Installer created: $DIST_DIR/windows/venture-${VERSION}-setup.exe"
else
    echo_warn "makensis not available, NSIS script created at: $NSIS_DIR/venture.nsi"
    cp "$NSIS_DIR/venture.nsi" "$DIST_DIR/windows/"
    cp "$NSIS_DIR/venture-server.exe" "$DIST_DIR/windows/"
    cp "$NSIS_DIR/venture-client.exe" "$DIST_DIR/windows/"
    echo_warn "Install NSIS and run 'makensis $NSIS_DIR/venture.nsi' to build installer"
fi

echo_info "Windows package build complete!"
