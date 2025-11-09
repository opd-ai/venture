# Venture Build and Validation Scripts

This directory contains scripts for building, testing, and validating the Venture project.

## Code Quality

### validate-code-review.sh
**Purpose:** Automated validation of code review quality gates as defined in `docs/CODE_REVIEW_PLAN.md`.

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
- ✓ Code coverage (≥65% per package)
- ✓ Static analysis (go vet)
- ✓ Code formatting (gofmt)
- ✓ Documentation completeness
- ✓ Package docs present (doc.go files)
- ✓ No circular dependencies

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

- [CODE_REVIEW_PLAN.md](../docs/CODE_REVIEW_PLAN.md) - Detailed code review methodology
- [DEVELOPMENT.md](../docs/DEVELOPMENT.md) - Development setup and workflow
- [CROSS_PLATFORM_BUILDS.md](../docs/CROSS_PLATFORM_BUILDS.md) - Cross-platform build guide
- [MOBILE_BUILD.md](../docs/MOBILE_BUILD.md) - Mobile build guide
- [GITHUB_PAGES.md](../docs/GITHUB_PAGES.md) - WebAssembly deployment guide

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
- ✓ Coverage Analysis (≥65% threshold, with interface-only exceptions)
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
  - **Minor**: Nice-to-have (enhancements, optimizations)

**Integration with Workflow:**
1. Run before merge to audit changed packages
2. Reference `AUDIT.md` during code review
3. Re-run after addressing findings to verify fixes
4. Periodically audit all packages to track quality trends
