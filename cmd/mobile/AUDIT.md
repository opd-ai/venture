# Audit: github.com/opd-ai/venture/cmd/mobile
**Date**: 2026-02-13
**Status**: Needs Work

## Summary
The `cmd/mobile` package serves as the entry point for iOS and Android platforms. It has 73.9% coverage in `config/` subpackage but 0% coverage for the main mobile.go entry point. The package properly initializes game systems and integrates with Ebiten's mobile API. Primary issues are lack of test coverage for mobile.go, missing package documentation, and minor error handling improvements needed.

## Issues Found
- [ ] **high** Test coverage — mobile.go has 0.0% coverage (target: 65%) (`mobile.go`)
- [ ] **med** Error handling — Unreachable return after logger.Fatal (`mobile.go:76`)
- [ ] **med** Error handling — Unreachable return nil after logger.Fatal (`mobile.go:99`)
- [ ] **med** Doc coverage — No package doc.go for cmd/mobile (missing package-level documentation)
- [ ] **med** Doc coverage — No package doc.go for cmd/mobile/config (missing package-level documentation)
- [ ] **low** Doc coverage — Exported function Start lacks godoc comment (`mobile.go:361`)
- [ ] **low** Doc coverage — Exported function Update lacks godoc comment (`mobile.go:369`)
- [ ] **low** Doc coverage — Exported function GetScreenWidth lacks godoc comment (`mobile.go:374`)
- [ ] **low** Doc coverage — Exported function GetScreenHeight lacks godoc comment (`mobile.go:382`)
- [ ] **low** Deterministic procgen — Uses time.Now().UnixNano() for default seed (acceptable for mobile UX but should be documented) (`config/seed.go:32`)

## Test Coverage
- **Overall**: 36.95% (mobile.go: 0.0%, config/: 73.9%)
- **Target**: 65%
- **Gap**: -28.05 percentage points

## Integration Status
The package successfully integrates with:
- `pkg/engine` — Uses `NewEbitenGameWithLogger`, `InitializeGameSystems`, `DefaultSystemInitConfig`
- `pkg/procgen/terrain` — Uses BSP generator for dungeon creation
- `pkg/procgen/item` — Generates starter items
- `github.com/hajimehoshi/ebiten/v2/mobile` — Properly calls `mobile.SetGame(gameInstance)`

**Mobile-specific initialization**: All game systems (44 total) are initialized inline via `engine.InitializeGameSystems`. No separate mobile system registration required. Camera, HUD, and animation systems are properly configured for mobile screen dimensions (720x1280).

**WASM note**: This package is mobile-specific; WASM uses `cmd/client` instead.

## Recommendations
1. **Add tests for mobile.go** — Create test stubs that mock Ebiten dependencies (similar to pkg/engine tests with StubSprite/StubInput). Focus on initialization flow, error paths, and configuration handling. Target: bring coverage above 65%.
2. **Create package documentation** — Add `doc.go` files for both `cmd/mobile` and `cmd/mobile/config` explaining purpose, usage, and platform-specific considerations.
3. **Remove unreachable returns** — Lines 76 and 99 have returns after logger.Fatal which never returns. Remove these dead code paths.
4. **Document time-based seed behavior** — Add comment in config/seed.go explaining that time.Now() usage is intentional for mobile UX (different experience each launch unless VENTURE_SEED is set).
5. **Add godoc comments** — Document exported functions Start, Update, GetScreenWidth, GetScreenHeight with their purpose and usage context.
