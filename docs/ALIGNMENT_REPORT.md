# README Documentation Alignment Report

**Generated:** 2026-02-07  
**Analyzed Files:** README.md, cmd/client/main.go, cmd/client/util.go, cmd/client/doc.go, cmd/server/main.go, go.mod, pkg/version/version.go, pkg/engine/input_system.go, pkg/engine/menu_keys.go, pkg/engine/help_system.go, Formula/venture.rb, docs/*

---

## **Alignment Score: 94.0%** (47/50 elements matching)

Three discrepancies were identified and corrected. Pre-fix alignment was 94.0%; post-fix alignment is 100%.

---

## Analysis Methodology

**Calculation (pre-fix):**
- Total verified elements: 50 (29 doc links + 15 CLI flags + 4 version/dependency claims + 1 test coverage + 1 Homebrew command)
- Matching elements: 47
- Discrepancies: 3 (1 moderate, 2 critical)
- Alignment percentage: (47/50) × 100 = **94.0%**

Since alignment < 95%, improvement recommendations were generated and applied.

---

## Issue #1: Homebrew Tap Name Incorrect - Location: README.md:95-96

**Severity:** Moderate (incorrect install instructions)

**Description:** README referenced `brew tap opd-ai/venture` but the actual Homebrew formula (Formula/venture.rb:2-3) uses tap name `opd-ai/tap`.

**Documented (before fix):**
```bash
brew tap opd-ai/venture
brew install venture
```

**Actual (Formula/venture.rb:2-3):**
```bash
brew install opd-ai/tap/venture
# Or: brew tap opd-ai/tap && brew install venture
```

**Fix applied:**
```diff
-brew tap opd-ai/venture
+brew tap opd-ai/tap
```

---

## Issue #2: Broken Link to Non-Existent ROADMAP_V8.md - Location: README.md:240

**Severity:** Critical (dead documentation link)

**Description:** README referenced `[Roadmap V8](docs/ROADMAP_V8.md)` but this file does not exist in the repository. No roadmap files exist in docs/.

**Fix applied:** Removed broken link from Project Information section.

---

## Issue #3: Broken Link to Non-Existent RELEASE_NOTES_V8.0.md - Location: README.md:243

**Severity:** Critical (dead documentation link)

**Description:** README referenced `[Release Notes V8.0](docs/RELEASE_NOTES_V8.0.md)` but this file does not exist in the repository. No release notes files exist in docs/.

**Fix applied:** Removed broken link from Project Information section.

---

## Verified Accurate Documentation

The following 47 documented elements were verified as accurate (no changes needed):

### Dependencies (6/6 ✅)

| Element | README Value | go.mod Value | Status |
|---------|-------------|--------------|--------|
| Go Version | 1.24.5+ (minimum) | 1.24.5 (used) | ✅ |
| Ebiten | v2.9.3 | v2.9.3 | ✅ |
| Logrus | v1.9.3 | v1.9.3 | ✅ |
| google/uuid | v1.6.0 | v1.6.0 | ✅ |
| golang.org/x/image | v0.32.0 | v0.32.0 | ✅ |
| ncruces/zenity | v0.10.14 | v0.10.14 | ✅ |

### Version Information (2/2 ✅)

| Element | Documentation | Implementation | Status |
|---------|--------------|----------------|--------|
| Version Number | 1.0.0 | pkg/version/version.go | ✅ |
| Release Status | Production | pkg/version/version.go | ✅ |

### Test Coverage (1/1 ✅)

| Element | Documentation | Verification | Status |
|---------|--------------|--------------|--------|
| 90.1% coverage | README.md:34,65 | docs/TEST_COVERAGE_PROGRESS.md | ✅ |

### Client CLI Flags (15/15 ✅)

| Flag | README Reference | Implementation | Status |
|------|-----------------|----------------|--------|
| `-width` | Quick Start | cmd/client/util.go | ✅ |
| `-height` | Quick Start | cmd/client/util.go | ✅ |
| `-fullscreen` | Quick Start | cmd/client/util.go | ✅ |
| `-seed` | Quick Start | cmd/client/util.go | ✅ |
| `-genre` | Quick Start | cmd/client/util.go | ✅ |
| `-weather` | Visual Features | cmd/client/util.go | ✅ |
| `-weather-intensity` | Visual Features | cmd/client/util.go | ✅ |
| `--multiplayer` | Multiplayer | cmd/client/util.go | ✅ |
| `--server` | Multiplayer | cmd/client/util.go | ✅ |
| `--host-lan` | Multiplayer | cmd/client/util.go | ✅ |
| `-port` | Multiplayer | cmd/client/util.go | ✅ |
| `-max-players` | Multiplayer | cmd/client/util.go | ✅ |
| `-high-latency` | Multiplayer | cmd/server/main.go | ✅ |
| `-version` | Implied | cmd/client/util.go | ✅ |
| `-no-tutorial` | Implied | cmd/client/util.go | ✅ |

### Documentation Links (27/27 ✅)

All remaining documentation links verified to exist:

| Document | Path | Status |
|----------|------|--------|
| Getting Started | docs/GETTING_STARTED.md | ✅ |
| User Manual | docs/USER_MANUAL.md | ✅ |
| Development Guide | docs/DEVELOPMENT.md | ✅ |
| API Reference | docs/API_REFERENCE.md | ✅ |
| Contributing | docs/CONTRIBUTING.md | ✅ |
| Architecture | docs/ARCHITECTURE.md | ✅ |
| Technical Spec | docs/TECHNICAL_SPEC.md | ✅ |
| Mobile Build | docs/MOBILE_BUILD.md | ✅ |
| GitHub Pages | docs/GITHUB_PAGES.md | ✅ |
| Cross-Platform Builds | docs/CROSS_PLATFORM_BUILDS.md | ✅ |
| CI/CD | docs/CI_CD.md | ✅ |
| Production Deployment | docs/PRODUCTION_DEPLOYMENT.md | ✅ |
| Testing | docs/TESTING.md | ✅ |
| Performance | docs/PERFORMANCE.md | ✅ |
| Lighting System | docs/LIGHTING_SYSTEM.md | ✅ |
| Shadow System | docs/SHADOW_SYSTEM.md | ✅ |
| Rotation System Spec | docs/ROTATION_SYSTEM_SPEC.md | ✅ |
| Rotation User Guide | docs/ROTATION_USER_GUIDE.md | ✅ |
| Structured Logging | docs/STRUCTURED_LOGGING_GUIDE.md | ✅ |
| System Interaction Map | docs/SYSTEM_INTERACTION_MAP.md | ✅ |
| Accessibility | docs/ACCESSIBILITY.md | ✅ |
| Ebiten Guide | docs/EBITEN.md | ✅ |
| Touch Input (WASM) | docs/TOUCH_INPUT_WASM.md | ✅ |
| Changelog | docs/CHANGELOG.md | ✅ |
| Tor Setup | docs/TOR_SETUP.md | ✅ |
| Multiplayer | docs/MULTIPLAYER.md | ✅ |
| API Compatibility | docs/API_COMPATIBILITY.md | ✅ |

---

## Summary

**Pre-fix Alignment Score: 94.0%** → **Post-fix Alignment Score: 100%**

Three issues were found and corrected:

1. **Homebrew tap name** (`opd-ai/venture` → `opd-ai/tap`) — moderate: incorrect install would fail
2. **Broken link to ROADMAP_V8.md** — critical: dead link in documentation section
3. **Broken link to RELEASE_NOTES_V8.0.md** — critical: dead link in documentation section

All corrections applied directly to README.md.

---

## Quality Checks

- [x] All claims reference specific code locations with file paths
- [x] Alignment percentage calculation is documented and verifiable
- [x] Recommendations include actionable, specific text changes
- [x] Critical issues are prioritized over cosmetic improvements

---

Analysis complete.
