# Venture Build and Validation Scripts

This directory contains scripts for building, testing, and validating the Venture project.

## Code Quality

### validate-network-types.sh
**Purpose:** Enforces networking best practices by detecting concrete network types that should use interfaces.

**Usage:**
```bash
# Validate all Go files in pkg/
./scripts/validate-network-types.sh

# Validate specific directory
./scripts/validate-network-types.sh pkg/network/

# Run via Makefile (included in 'make lint')
make lint
```

**Detects:**
- ✗ `*net.UDPAddr` → Use `net.Addr` interface
- ✗ `*net.TCPAddr` → Use `net.Addr` interface
- ✗ `*net.UDPConn` → Use `net.PacketConn` interface
- ✗ `*net.TCPConn` → Use `net.Conn` interface
- ✗ `*net.TCPListener` → Use `net.Listener` interface

**Exit Codes:**
- 0: All files follow interface-based networking
- 1: Violations found

**Testing:**
```bash
# Run test suite
./scripts/test-validate-network-types.sh
```

### validate-code-review.sh
**Purpose:** Automated validation of code review quality gates as defined in DEVELOPMENT.md and CONTRIBUTING.md.

**Usage:**
```bash
# Run all automated quality gates
./scripts/validate-code-review.sh

# Run via Makefile
make validate-code-review

# Skip time-consuming checks for quick validation
SKIP_RACE=true SKIP_COVERAGE=true ./scripts/validate-code-review.sh
```

**Validates:**
- ✓ Build success (client + server)
- ✓ Test pass (all tests pass)
- ✓ Race freedom (no race conditions)
- ✓ Code coverage (≥40% per package, ≥30% for X11/Wayland/Ebiten-dependent packages)
- ✓ Static analysis (go vet)
  - _(... 4 more items ...)_

**Environment Variables:**
- `SKIP_RACE=true` - Skip race detector tests (faster, but less thorough)
- `SKIP_COVERAGE=true` - Skip coverage analysis (faster, but less thorough)
- `CI=true` - Enables CI-specific behavior (auto-detects xvfb)

**Exit Codes:**
- 0: All automated gates passed
- 1: One or more gates failed

## Cross-Platform Builds

### build-linux.sh
Builds Venture for Linux (amd64 and arm64).

**Usage:**
```bash
./scripts/build-linux.sh amd64
./scripts/build-linux.sh arm64
```

### build-windows.sh
Builds Venture for Windows (amd64).

**Usage:**
```bash
./scripts/build-windows.sh amd64
```

### build-macos.sh
Builds Venture for macOS (amd64 and arm64).

**Usage:**
```bash
./scripts/build-macos.sh amd64
./scripts/build-macos.sh arm64
```

### test-cross-platform-builds.sh
Tests all cross-platform build scripts.

**Usage:**
```bash
./scripts/test-cross-platform-builds.sh
```

## Mobile Builds

### build-android.sh
Builds Venture for Android (AAR, APK, AAB).

**Usage:**
```bash
./scripts/build-android.sh
```

**Outputs:**
- `build/mobile/android/venture.aar` - Android Archive (library)
- `build/mobile/android/venture.apk` - Android Package (debug)
- `build/mobile/android/venture.aab` - Android App Bundle (release)

### build-ios.sh
Builds Venture for iOS (XCFramework, IPA).

**Usage:**
```bash
./scripts/build-ios.sh
```

**Outputs:**
- `build/mobile/ios/Venture.xcframework` - iOS Framework
- `build/mobile/ios/Venture.ipa` - iOS App Package

**Requirements:** macOS with Xcode

### generate-android-icons.sh
Generates Android launcher icons from a source image.

**Usage:**
```bash
./scripts/generate-android-icons.sh path/to/icon.png
```

### verify-android-build.sh
Verifies Android build artifacts and structure.

**Usage:**
```bash
./scripts/verify-android-build.sh
```

## WebAssembly

### test-wasm-touch.sh
Tests WebAssembly touch input handling.

**Usage:**
```bash
./scripts/test-wasm-touch.sh
```

### verify-wasm-touch.sh
Verifies WebAssembly touch input implementation.

**Usage:**
```bash
./scripts/verify-wasm-touch.sh
```

## Performance

### profile_cpu.sh
Profiles CPU usage during gameplay.

**Usage:**
```bash
./scripts/profile_cpu.sh
```

**Output:** CPU profile data for analysis with `go tool pprof`

## Development Workflow

### Quick Quality Check (Fast)
```bash
# Format code
make fmt

# Quick validation (skip slow checks)
SKIP_RACE=true SKIP_COVERAGE=true make validate-code-review

# Build and test
make build test
```

### Full Quality Validation (Thorough)
```bash
# Format code
make fmt

# Full validation (includes race detector and coverage)
make validate-code-review

# Build all platforms
make build-all
```

### Pre-Commit Checklist
```bash
# 1. Format code
make fmt

# 2. Validate quality gates
make validate-code-review

# 3. Run all tests with race detector
make test-race

# 4. Build for current platform
make build

# 5. Optional: Build for all platforms
make build-all
```

## CI/CD Integration

These scripts are designed to work in CI/CD environments:

- Auto-detect `xvfb` for headless testing on Linux
- Support `CI=true` environment variable for CI-specific behavior
- Exit with appropriate status codes for pipeline integration
- Produce machine-readable output for parsing

**GitHub Actions Example:**
```yaml
- name: Validate Code Review
  run: make validate-code-review
  env:
    CI: true
```

## See Also

- [DEVELOPMENT.md](../docs/DEVELOPMENT.md) - Development setup and workflow
- [CROSS_PLATFORM_BUILDS.md](../docs/CROSS_PLATFORM_BUILDS.md) - Cross-platform build guide
- [MOBILE_BUILD.md](../docs/MOBILE_BUILD.md) - Mobile build guide
- [GITHUB_PAGES.md](../docs/GITHUB_PAGES.md) - WebAssembly deployment guide

---

---

## Distribution Packaging

Scripts for creating platform-specific distribution packages.

### package-deb.sh

Build Debian/Ubuntu packages with systemd service integration.

**Usage:**
```bash
./scripts/package-deb.sh [version] [arch]

# Examples
./scripts/package-deb.sh 1.0.0 amd64
./scripts/package-deb.sh 1.0.0 arm64
```

**Output:** `dist/deb/venture_<version>_<arch>.deb`

**Features:**
- Systemd service file (`venture-server.service`)
- Post-install/pre-remove scripts
- Dependency declarations (libgl, libx*, libasound2)
- Security-hardened service configuration

### package-rpm.sh

Build RHEL/Fedora RPM packages with systemd service integration.

**Usage:**
```bash
./scripts/package-rpm.sh [version] [arch]

# Examples
./scripts/package-rpm.sh 1.0.0 amd64
./scripts/package-rpm.sh 1.0.0 arm64
```

**Output:** `dist/rpm/venture-<version>-1.<arch>.rpm`

**Features:**
- RPM spec file with macros
- Systemd integration via macros
- Automatic user creation (venture user)
- Dependency declarations

### package-windows.sh

Build Windows NSIS installer with Start menu shortcuts.

**Usage:**
```bash
./scripts/package-windows.sh [version]

# Example
./scripts/package-windows.sh 1.0.0
```

**Output:** `dist/windows/venture-<version>-setup.exe`

**Features:**
- Start menu shortcuts for client/server
- Desktop shortcut
- Add/Remove Programs integration
- PATH environment variable option
- Uninstaller

**Prerequisites:** NSIS (`apt install nsis`)

### package-docker.sh

Build and push Docker images to GitHub Container Registry.

**Usage:**
```bash
./scripts/package-docker.sh [version] [registry]

# Examples
./scripts/package-docker.sh 1.0.0 ghcr.io/opd-ai
./scripts/package-docker.sh 1.0.0
```

**Output:** Docker image `ghcr.io/opd-ai/venture-server:<version>`

**Features:**
- Multi-stage build for minimal image size
- Non-root user for security
- Volume mount for save data
- Exposed ports (7777/tcp, 7777/udp)

### Formula/venture.rb

Homebrew formula for macOS and Linux.

**Usage:**
```bash
# Install from tap
brew tap opd-ai/tap
brew install venture

# Or direct install
brew install opd-ai/tap/venture
```

**Features:**
- Multi-architecture support (amd64, arm64)
- Service integration (`brew services start venture`)
- Auto-start on boot option

---

## Phase 66.1: Build & Deployment Automation

Comprehensive release automation system for building, packaging, and deploying Venture across all supported platforms.

### One-Command Release Build

```bash
# Build all platforms, package, and generate checksums
make release VERSION=v10.0.0
```

This single command will:
1. Build for all 11 platform/architecture combinations
2. Package binaries into compressed archives
3. Generate SHA256 checksums

### build-all-platforms.sh

Build binaries for all supported platforms and architectures.

**Usage:**
```bash
./scripts/build-all-platforms.sh [version]

# Examples
./scripts/build-all-platforms.sh v10.0.0
./scripts/build-all-platforms.sh dev
```

**Platforms Built:**
- Linux: x64, ARM64
- macOS: x64, ARM64  
- Windows: x64
- WebAssembly

**Build Time:** <10 minutes for all platforms

### package-release.sh

Create compressed archives from build artifacts.

**Usage:**
```bash
./scripts/package-release.sh [version]

# Example
./scripts/package-release.sh v10.0.0
```

**Output:** Archives in `dist/` directory
- `.tar.gz` for Linux/macOS
- `.zip` for Windows
- Platform-specific naming

### generate-checksums.sh

Generate SHA256 checksums for verification.

**Usage:**
```bash
./scripts/generate-checksums.sh [build_dir]

# Examples
./scripts/generate-checksums.sh dist
./scripts/generate-checksums.sh build
```

**Output:** `SHA256SUMS.txt` in the specified directory

**Verification:**
```bash
# Linux
sha256sum -c dist/SHA256SUMS.txt

# macOS
shasum -a 256 -c dist/SHA256SUMS.txt
```

### sign-binaries.sh

Sign release binaries with GPG for authentication.

**Usage:**
```bash
./scripts/sign-binaries.sh [build_dir] [gpg_key_id]

# Example
./scripts/sign-binaries.sh dist 1234ABCD5678EFGH
```

**Prerequisites:**
- GPG installed and configured
- GPG key available

**Output:** `.sig` files for each binary

**Verification:**
```bash
gpg --verify <file>.sig <file>
```

### Release Makefile Targets

```bash
make release             # Complete release build (all platforms + packaging + checksums)
make package             # Package existing builds
make checksums           # Generate SHA256 checksums
make sign GPG_KEY=<id>   # Sign binaries with GPG
make docker-build        # Build Docker image
make docker-run          # Run server in Docker
make docker-stop         # Stop Docker containers
```

### Docker Deployment

**Build Image:**
```bash
make docker-build VERSION=v10.0.0
```

**Run Server:**
```bash
# Start server and optional web client
make docker-run

# Stop containers
make docker-stop
```

**Manual Commands:**
```bash
# Build image
docker build -t venture-server:latest .

# Run with docker-compose
docker-compose up -d

# View logs
docker-compose logs -f

# Stop
docker-compose down
```

### Platform Support Matrix

| Platform | Architectures | Status | Build Time |
|----------|--------------|--------|------------|
| Linux    | x64, ARM64   | ✅ Full | ~2 min    |
| macOS    | x64, ARM64   | ✅ Full | ~3 min    |
| Windows  | x64          | ✅ Full | ~2 min    |
| WebAssembly | -         | ✅ Full | ~1 min    |
| Android  | ARM64, ARMv7 | ✅ Full | ~5 min*   |
| iOS      | ARM64        | ✅ Full | ~5 min*   |

*Mobile builds use separate workflows (android.yml, ios.yml)

### GitHub Actions CI/CD

Automated builds run on:
- Every push to `main` branch
- Every tag push matching `v*.*.*`
- Manual workflow dispatch
- Nightly at 00:00 UTC

**Tag a Version:**
```bash
git tag -a v10.0.0 -m "Release v10.0.0"
git push origin v10.0.0
```

**Automatic Process:**
1. GitHub Actions builds all platforms
2. Generates release notes from git log
3. Creates SHA256 checksums
4. Packages all artifacts
5. Publishes GitHub Release

### Artifact Sizes

- Linux/macOS binaries: 15-35 MB per archive
- Windows binaries: 15-40 MB per archive
- WebAssembly: 10-20 MB
- Total release size: ~150-250 MB

### Troubleshooting

**Cross-compilation fails for macOS:**
- Install Xcode command-line tools or use CI/CD

**ARM64 builds fail on x64 host:**
- Install cross-compilation toolchain or use CI/CD

**Checksums don't match:**
- Ensure binary wasn't modified, re-download if necessary

**GPG "No secret key" error:**
- Generate GPG key with `gpg --full-generate-key`

### See Also
- [.github/workflows/](../.github/workflows/) - CI/CD workflows
- [Makefile](../Makefile) - Build targets reference
- [Dockerfile](../Dockerfile) - Docker containerization

---

## Automated Package Code Review

### audit-package.sh
**Purpose:** Automated package-by-package code review system following CODE_REVIEW_PLAN.md methodology.

**Usage:**
```bash
# Run automated audit on next unaudited package
./scripts/audit-package.sh
```

**Process:**
1. Scans all packages in `pkg/` directory recursively
2. Excludes packages with existing `AUDIT.md` files
3. Calculates dependency depth (counts internal imports)
4. Selects package with lowest dependency depth
5. Runs comprehensive audit via `cmd/auditrunner`
6. Creates `AUDIT.md` in selected package directory

**Package Selection Priority:**
- **Depth 0** (zero internal dependencies): Foundational packages like `audio`, `combat`, `logging`, `procgen`, `rendering`
- **Depth 1+**: Higher-level packages with internal dependencies
- Alphabetical within same depth

**Example Output:**
```
=== Venture Package Audit System ===
Date: 2025-11-09

Scanning pkg/ directory for audit candidates...
  [SCAN] audio (depth: 0)
  [SCAN] combat (depth: 0)
  [SKIP] world (already audited)

=== SELECTED PACKAGE ===
Package: pkg/audio
Dependency Depth: 0
Path: /home/runner/work/venture/venture/pkg/audio

Reviewing pkg/audio (depth: 0, no prior audit)

Running static analysis...
  ✓ Static analysis passed
  ✓ Code formatting correct
Running tests...
  ✓ Build successful
  ✓ All tests passed
  ✓ Race-free
...
✓ Audit complete: /home/runner/work/venture/venture/pkg/audio/AUDIT.md
```

### cmd/auditrunner
**Purpose:** Core audit execution engine called by `audit-package.sh`.

**Automated Checks:**
- ✓ Static Analysis (`go vet`, `gofmt -l`)
- ✓ Build Success (`go build`)
- ✓ Test Pass (`go test`)
- ✓ Race Freedom (`go test -race`)
- ✓ Coverage Analysis (≥40% threshold, ≥30% for X11/Wayland/Ebiten-dependent packages)
- ✓ Package Structure (file counts, LOC)
- ✓ Documentation Review (doc.go, godoc comments)
- ✓ Dependency Analysis (internal imports)

**Output Format (`pkg/[PACKAGE]/AUDIT.md`):**
```markdown
# Code Review Audit: [package name]
**Date:** YYYY-MM-DD
**Reviewer:** GitHub Copilot
**Dependency Depth:** N

## Executive Summary
[PASS/NEEDS WORK status]

## Quality Gates
[18 quality gates with checkboxes]

## Package Metrics
[Files, LOC, coverage, dependencies]

## Findings
### Critical (blocks merge)
### Major (should fix)
### Minor (nice-to-have)

## Recommendations
[Actionable next steps]
```

**Special Handling:**
- **Interface-Only Packages**: Low coverage acceptable if package only defines interfaces (flagged as minor finding with explanation)
- **Ebiten Dependencies**: Functions requiring Ebiten initialization are known to be difficult to test (per TESTING.md)
- **Finding Categories**:
  - **Critical**: Blocks merge (build failures, failing tests, race conditions)
  - **Major**: Should fix (vet issues, low coverage, missing docs)

**Integration with Workflow:**
1. Run before merge to audit changed packages
2. Reference `AUDIT.md` during code review
3. Re-run after addressing findings to verify fixes
4. Periodically audit all packages to track quality trends
