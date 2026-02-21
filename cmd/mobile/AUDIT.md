# cmd/mobile/ Audit — 2026-02-16

## Summary

Audited all source files in `cmd/mobile/` (3 source files, ~400 LOC).
All issues resolved.

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

### Resolved Issues

1. **[LOW] Nil terrain access** (`mobile.go:165`) — **FIXED 2026-02-21**
   - Added nil check for `generatedTerrain` before accessing `Rooms`
   - Now safely falls back to default position (400, 300) if terrain is nil

2. **[LOW] Hardcoded screen dimensions** (`mobile.go:67-68`) — **FIXED 2026-02-21**
   - Extracted to named constants `DefaultScreenWidth` (720) and `DefaultScreenHeight` (1280)
   - Constants are now used throughout the package for consistency

3. **[LOW] Silent item generation failures** (`mobile.go:326, 334, 347`) — **FIXED 2026-02-21**
   - Changed from `if err == nil` pattern to explicit error logging with `logger.WithError(err).Debug()`
   - Item generation failures are now logged for observability while maintaining non-blocking behavior

## Test Coverage

- `config/seed_test.go` — Unit tests for `GetSeedFromEnv`, `GetGenreFromEnv`
- `config/integration_test.go` — Integration tests for full config flow
- `mobile.go` — No unit tests (requires ebitenmobile runtime, not testable in CI)
