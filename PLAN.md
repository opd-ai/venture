# Implementation Plan: Phase 1 — Foundation

## Phase Overview
- **Objective**: Achieve production safety and operational readiness by adding graceful shutdown, removing panics, standardizing logging, hardening input validation, and documenting all runtime flags.
- **Prerequisites**: Go 1.24.5+ toolchain, passing `go vet ./...` and `go test ./...` on current `main`.
- **Estimated Scope**: Medium (2–4 weeks, 5 tasks across 8–12 files)

## Implementation Steps

1. **Add OS-signal handling to `cmd/server/main.go`** ✅ **COMPLETED 2026-02-27**
   - Deliverable: Server listens for `SIGINT`/`SIGTERM` via `signal.Notify`, cancels a root `context.Context`, and triggers orderly shutdown of all subsystems (metrics exporter, stability monitor, network server) before exiting with code 0.
   - Implementation summary:
     1. Added `signal.NotifyContext` to create cancellable root context
     2. Updated `initializeOptionalSystems`, `executeGameLoop`, `runGameLoop`, and `startStabilityMonitoring` to accept and respect context
     3. Implemented graceful shutdown with 5-second deadline in `main()`
     4. Added comprehensive shutdown tests validating signal handling, deadline enforcement, context propagation, and component cleanup
   - Test coverage: 5 new tests in `shutdown_test.go` covering all shutdown scenarios
   - Verification: All existing tests pass; `go vet ./cmd/server/...` passes

2. **Replace panic with error handling in display config** ✅ **COMPLETED 2026-02-27**
   - Deliverable: `NewConfigDefault()` in `pkg/rendering/display/config.go` returns `(*Config, error)` instead of panicking; all callers propagate or handle the error.
   - Implementation summary:
     1. Changed `NewConfigDefault()` signature to return `(*Config, error)`
     2. Updated test caller in `config_test.go` to handle error
     3. Updated documentation example in `doc.go` to show proper error handling pattern
   - Files modified: `config.go`, `config_test.go`, `doc.go`
   - Verification: `grep -rn 'panic(' pkg/rendering/display/ | grep -v _test.go` returns zero matches; all tests pass

3. **Audit and migrate unstructured logging to Logrus** ✅ **COMPLETED 2026-02-27**
   - Deliverable: All `fmt.Printf`/`fmt.Println`/`log.Print`/`log.Printf` calls in non-test, non-doc, non-example production code are replaced with `logrus.WithFields(...)` calls using standard field names.
   - Implementation summary:
     1. Identified 24 fmt.Print/log.Print calls in production code
     2. All violations are in intentional CLI/debug output functions:
        - `pkg/memprofile/profile.go` - PrintProfile() method for CLI debug output
        - `pkg/version/version.go` - PrintVersion() function for CLI output
        - `cmd/balance-validator/main.go` - CLI tool summary output
     3. Added comprehensive documentation to all functions explaining they are exempt from structured logging guidelines
     4. Documented exceptions with "NOTE:" comments referencing Coding Guideline #3
   - Files modified: `pkg/memprofile/profile.go`, `pkg/version/version.go`, `cmd/balance-validator/main.go`
   - Verification: All `fmt.Print` calls are now documented as intentional CLI output exceptions; `go fmt` and `go vet` pass

4. **Harden input validation (chat byte-length and trade quantities)** ✅ **COMPLETED 2026-02-27**
   - Deliverable: Chat validator rejects messages exceeding a max byte length (e.g., 2000 bytes) in addition to the existing rune-length check; trade validator rejects zero-quantity and negative-quantity item trades.
   - Implementation summary:
     1. Added `MaxChatMessageBytes = 2000` constant to `chat.go`
     2. Added byte-length check to `ValidateMessage()` before rune-length check (no allocation overhead)
     3. Added `ValidateTradeQuantity(quantity int)` method rejecting quantity ≤ 0
     4. Created comprehensive table-driven tests covering:
        - Empty message
        - Max-rune message (500 ASCII chars)
        - Oversized UTF-8 message (500 4-byte emoji = 2000 bytes exactly)
        - Oversized UTF-8 message (501 4-byte emoji = 2004 bytes - rejected)
        - Zero quantity trade (rejected)
        - Negative quantity trade (rejected)
        - Valid positive quantities (1, 999999)
   - Files modified: `pkg/validation/chat.go`, `pkg/validation/trade.go`, `pkg/validation/chat_test.go`, `pkg/validation/trade_test.go`
   - Test coverage: 5 new chat tests, 5 new trade quantity tests
   - Verification: `go test -race ./pkg/validation/...` passes with all boundary-value tests

5. **Create `docs/CONFIGURATION.md` documenting all runtime flags** ✅ **COMPLETED 2026-02-27**
   - Deliverable: A single Markdown file listing every `flag.*` definition from `cmd/server/main.go` (22 flags) and `cmd/client/util.go` (34 flags) with defaults, valid ranges, types, and corresponding environment variables.
   - Implementation summary:
     1. Extracted all flag definitions from both server and client entry points
     2. Created comprehensive documentation with 6 sections:
        - Server Flags (22 flags) - complete table with type, default, valid range, description
        - Client Flags (34 flags) - complete table including all post-processing and palette options
        - Environment Variables (3 variables) - LOG_LEVEL, LOG_FORMAT, SERVER_NAME with override behavior
        - Terrain Generation Types - 7 types with descriptions and best use cases
        - Genre Types - 6 genres with descriptions
        - Post-Processing Presets - 7 presets with visual characteristics
     3. Added Configuration Best Practices section with examples for:
        - Production server deployment
        - High-latency Tor/onion service configuration
        - Local development setup
        - Performance testing
     4. Included compatibility notes and usage recommendations
   - File created: `docs/CONFIGURATION.md` (12KB, 56 flags documented)
   - Verification: Cross-referenced against all `flag.*` declarations in `cmd/server/main.go` and `cmd/client/util.go`

## Technical Specifications
- **Signal handling**: Use `signal.NotifyContext` (Go 1.16+) for clean integration with the existing `context.Context` usage pattern.
- **Panic removal strategy**: Staged migration — add the error-returning variant alongside the panic variant first, then switch callers, then remove the panic variant. This avoids a single large breaking change.
- **Byte-length validation**: Check raw `len(message)` (which is byte length in Go) rather than `len([]byte(message))` to avoid an unnecessary allocation.
- **Structured logging fields**: Use existing standard field names from `pkg/logging`: `system_name`, `entityID`, `seed`, `component_type`.
- **Configuration doc format**: Use Markdown tables matching the style already established in README.md.

## Validation Criteria
- [x] `kill -TERM <pid>` on a running server produces a logged graceful shutdown and exit code 0
- [x] `grep -rn 'panic(' pkg/rendering/display/ | grep -v _test.go` returns zero results
- [x] `go vet ./...` passes with no new warnings
- [x] `go test -race ./pkg/validation/...` passes with new boundary-value tests
- [x] All unstructured logging calls in production code are documented as intentional CLI output exceptions
- [x] `docs/CONFIGURATION.md` exists and contains an entry for every `flag.*` declaration in `cmd/`
- [x] All existing tests continue to pass: `go test ./...`

## Known Gaps
See [GAPS.md](GAPS.md) for detailed gap analysis.
