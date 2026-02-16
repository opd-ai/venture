# cmd/mobile/ Audit — 2026-02-16

## Summary

Audited all source files in `cmd/mobile/` (3 source files, ~400 LOC).
Found 0 actionable issues requiring fixes.

**Coverage**: Config sub-package has tests (`config/seed_test.go`, `config/integration_test.go`).
Main `mobile.go` is untestable without mobile device/emulator (ebitenmobile entry point).

## Files Reviewed

| File | LOC | Notes |
|------|-----|-------|
| `mobile.go` | ~350 | Mobile entry point, init-based initialization |
| `doc.go` | ~15 | Package documentation |
| `config/seed.go` | ~80 | Seed/genre environment variable handling |
| `config/doc.go` | ~10 | Config package documentation |

## Observations

### Architecture
- Clean separation of config logic into sub-package with tests
- Uses `init()` for immediate initialization (required by ebitenmobile)
- Global state is necessary for mobile framework integration
- Nil logger checks in config package follow defensive patterns

### Low-Risk Items (No Fix Required)

1. **[LOW] Nil terrain access** (`mobile.go:165`)
   - `generatedTerrain.Rooms` accessed without nil check
   - Low risk: terrain generation `Fatal()`s on error, so nil path is unreachable
   - Pattern mirrors server code where it was fixed

2. **[LOW] Hardcoded screen dimensions** (`mobile.go:67-68`)
   - 720x1280 hardcoded for mobile
   - Acceptable for mobile targets (standard portrait resolution)

3. **[LOW] Silent item generation failures** (`mobile.go:326, 334, 347`)
   - Uses `if err == nil` pattern (inverted) to skip failed items
   - Acceptable: items are optional content, failure is non-critical

## Test Coverage

- `config/seed_test.go` — Unit tests for `GetSeedFromEnv`, `GetGenreFromEnv`
- `config/integration_test.go` — Integration tests for full config flow
- `mobile.go` — No unit tests (requires ebitenmobile runtime, not testable in CI)
