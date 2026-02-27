# Implementation Plan: Phase 1 — Foundation

## Phase Overview
- **Objective**: Achieve production safety and operational readiness by adding graceful shutdown, removing panics, standardizing logging, hardening input validation, and documenting all runtime flags.
- **Prerequisites**: Go 1.24.5+ toolchain, passing `go vet ./...` and `go test ./...` on current `main`.
- **Estimated Scope**: Medium (2–4 weeks, 5 tasks across 8–12 files)

## Implementation Steps

1. **Add OS-signal handling to `cmd/server/main.go`**
   - Deliverable: Server listens for `SIGINT`/`SIGTERM` via `signal.Notify`, cancels a root `context.Context`, and triggers orderly shutdown of all subsystems (metrics exporter, stability monitor, network server) before exiting with code 0.
   - Dependencies: None.
   - Files to modify: `cmd/server/main.go`.
   - Approach:
     1. Create a root context with `signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)`.
     2. Replace `context.Background()` passed to long-lived goroutines (e.g., stability monitor at line 972) with the cancellable context.
     3. Replace the `defer shutdownServer(...)` pattern with an explicit shutdown sequence triggered by context cancellation.
     4. Add a 5-second deadline for graceful shutdown; log and force-exit if exceeded.
   - Acceptance criteria: `kill -TERM <pid>` results in logged graceful shutdown with exit code 0; integration test validates the behavior.

2. **Replace panic with error handling in display config**
   - Deliverable: `NewConfigDefault()` in `pkg/rendering/display/config.go` returns `(*Config, error)` instead of panicking; all callers propagate or handle the error.
   - Dependencies: None.
   - Files to modify: `pkg/rendering/display/config.go` and all callers of `NewConfigDefault()`.
   - Approach:
     1. Change signature from `func NewConfigDefault() *Config` to `func NewConfigDefault() (*Config, error)`.
     2. Return the error from `NewConfig(1920, 1080, false)` instead of panicking.
     3. Find all callers with `grep -rn 'NewConfigDefault' pkg/ cmd/` and update them to handle the error (log and exit at application boundaries, propagate in libraries).
   - Acceptance criteria: `go vet ./pkg/rendering/display/...` passes; `grep -rn 'panic(' pkg/rendering/display/` returns zero matches in non-test code.

3. **Audit and migrate unstructured logging to Logrus**
   - Deliverable: All `fmt.Printf`/`fmt.Println`/`log.Print`/`log.Printf` calls in non-test, non-doc, non-example production code are replaced with `logrus.WithFields(...)` calls using standard field names.
   - Dependencies: None.
   - Files to modify (current known violations):
     - `pkg/memprofile/profile.go` (~6 `fmt.Printf` calls in `Print()` method).
     - `pkg/engine/mail_doc.go` (1 `fmt.Printf` in non-comment code — verify if this is example-only).
   - Approach:
     1. Run `grep -rn 'fmt\.Print\|log\.Print' pkg/ cmd/ --include='*.go' | grep -v _test.go | grep -v doc.go | grep -v examples/` to find remaining violations.
     2. For each violation, replace with the equivalent `logrus` call using structured fields.
     3. For `pkg/memprofile/profile.go`, if the `Print()` method is intended for CLI/debug output, consider gating it behind a `logrus.Debug` call or keeping it as-is if it is only called from CLI tools (document the exception).
   - Acceptance criteria: `grep -rn 'fmt\.Print\|log\.Print' pkg/ cmd/ --include='*.go' | grep -v _test.go | grep -v doc.go | grep -v examples/ | grep -v '//'` returns zero matches (excluding in-comment references).

4. **Harden input validation (chat byte-length and trade quantities)**
   - Deliverable: Chat validator rejects messages exceeding a max byte length (e.g., 2000 bytes) in addition to the existing rune-length check; trade validator rejects zero-quantity and negative-quantity item trades.
   - Dependencies: None.
   - Files to modify:
     - `pkg/validation/chat.go` — add `MaxChatMessageBytes` constant and byte-length check in `ValidateMessage()`.
     - `pkg/validation/trade.go` — add `ValidateTradeQuantity(quantity int)` method rejecting quantity ≤ 0.
     - Corresponding `_test.go` files for table-driven boundary tests.
   - Approach:
     1. In `chat.go`, add `const MaxChatMessageBytes = 2000` and check `len([]byte(message)) > MaxChatMessageBytes` before the rune-length check.
     2. In `trade.go`, add a `ValidateTradeQuantity` method: reject `quantity <= 0` with descriptive errors.
     3. Write table-driven tests covering: empty message, max-rune message, oversized UTF-8 message (e.g., 500 4-byte emoji = 2000 bytes but 500 runes), zero quantity, negative quantity, valid quantity.
   - Acceptance criteria: `go test -race ./pkg/validation/...` passes; new tests cover boundary values.

5. **Create `docs/CONFIGURATION.md` documenting all runtime flags**
   - Deliverable: A single Markdown file listing every `flag.*` definition from `cmd/server/main.go` (24 flags) and `cmd/client/util.go` (35 flags) with defaults, valid ranges, types, and corresponding environment variables.
   - Dependencies: Steps 1 (may add new flags or change defaults).
   - Files to create: `docs/CONFIGURATION.md`.
   - Approach:
     1. Extract all `flag.String`, `flag.Bool`, `flag.Int`, `flag.Int64`, `flag.Float64` declarations from `cmd/server/main.go` and `cmd/client/util.go`.
     2. Cross-reference with README.md tables (Client Flags, Server Flags, Environment Variables) and fill in any missing entries.
     3. Organize by: Server Flags, Client Flags, Environment Variables, with columns: Flag, Type, Default, Valid Range, Description.
   - Acceptance criteria: Every `flag.*` call in `cmd/` has a corresponding entry; document is cross-referenced against `grep -rn 'flag\.' cmd/`.

## Technical Specifications
- **Signal handling**: Use `signal.NotifyContext` (Go 1.16+) for clean integration with the existing `context.Context` usage pattern.
- **Panic removal strategy**: Staged migration — add the error-returning variant alongside the panic variant first, then switch callers, then remove the panic variant. This avoids a single large breaking change.
- **Byte-length validation**: Check raw `len(message)` (which is byte length in Go) rather than `len([]byte(message))` to avoid an unnecessary allocation.
- **Structured logging fields**: Use existing standard field names from `pkg/logging`: `system_name`, `entityID`, `seed`, `component_type`.
- **Configuration doc format**: Use Markdown tables matching the style already established in README.md.

## Validation Criteria
- [ ] `kill -TERM <pid>` on a running server produces a logged graceful shutdown and exit code 0
- [ ] `grep -rn 'panic(' pkg/rendering/display/ | grep -v _test.go` returns zero results
- [ ] `go vet ./...` passes with no new warnings
- [ ] `go test -race ./pkg/validation/...` passes with new boundary-value tests
- [ ] `grep -rn 'fmt\.Print\|log\.Print' pkg/ cmd/ --include='*.go' | grep -v _test.go | grep -v doc.go | grep -v examples/ | grep -v '//'` returns zero matches in production code
- [ ] `docs/CONFIGURATION.md` exists and contains an entry for every `flag.*` declaration in `cmd/`
- [ ] All existing tests continue to pass: `go test ./...`

## Known Gaps
See [GAPS.md](GAPS.md) for detailed gap analysis.
