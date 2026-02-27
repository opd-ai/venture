# Implementation Gaps

## Phase 1: Foundation

### Gap 1: Signal handler integration test infrastructure
- **Gap**: The acceptance criteria for Task 1 (OS-signal handling) requires an integration test that validates exit code 0 on `SIGTERM`. No integration test harness currently exists for the server binary's lifecycle (start → signal → graceful shutdown).
- **Impact**: Without an integration test, graceful shutdown can only be validated manually. Regressions could go undetected.
- **Resolution needed**: Create a test helper (or `TestMain`-based integration test in `cmd/server/`) that starts the server as a subprocess, sends `SIGTERM`, and asserts exit code 0 within a 5-second timeout.

### Gap 2: `NewConfigDefault()` caller inventory incomplete
- **Gap**: The panic removal in `pkg/rendering/display/config.go` requires updating all callers of `NewConfigDefault()`. The full caller set has not been enumerated; callers may exist in client initialization, test helpers, or rendering subsystems.
- **Impact**: Changing the function signature without updating all callers will cause compile errors. Some callers may be in conditionally-compiled files (build tags for WASM, mobile).
- **Resolution needed**: Run `grep -rn 'NewConfigDefault' --include='*.go'` across the entire repository and enumerate every call site, including those behind build tags (`//go:build wasm`, `//go:build android`).

### Gap 3: Trade validation lacks quantity-per-item concept
- **Gap**: The current `TradeValidator` in `pkg/validation/trade.go` validates item IDs and item counts but has no concept of per-item quantities. The ROADMAP specifies rejecting "negative-quantity and zero-value trades," but the existing trade data model passes item IDs as string slices with no associated quantity field.
- **Impact**: Adding `ValidateTradeQuantity` is straightforward, but it will only be useful if the trade system's data model is updated to include quantities. Without this, the validation method exists but is never called from the trade flow.
- **Resolution needed**: Determine whether the trade system should support per-item quantities (e.g., "5x Iron Ore") or if the current model (each item ID is unique) is intentional. If quantities are needed, update `pkg/network/trade/` and `pkg/engine/trade_system.go` data structures first.

### Gap 4: `pkg/memprofile/profile.go` logging exemption decision
- **Gap**: The `Print()` method in `pkg/memprofile/profile.go` uses 6+ `fmt.Printf` calls to output a human-readable memory profile report. This may be intentional CLI/debug output rather than structured logging.
- **Impact**: Migrating these to `logrus` would change the output format of a debugging tool. Keeping them as `fmt.Printf` would violate the "zero unstructured logging" acceptance criteria.
- **Resolution needed**: Decide whether `pkg/memprofile` is classified as a CLI debugging tool (exempt from structured logging) or production code (must use `logrus`). Document the decision.

### Gap 5: No automated flag-to-documentation synchronization
- **Gap**: Task 5 creates `docs/CONFIGURATION.md` as a manual document. There is no mechanism to detect when new flags are added to `cmd/server/main.go` or `cmd/client/util.go` without a corresponding documentation update.
- **Impact**: Configuration documentation will drift from code over time as new flags are added.
- **Resolution needed**: Consider adding a CI check (e.g., a script that parses `flag.*` declarations and compares against `docs/CONFIGURATION.md` entries) or a `go generate` step to auto-generate the documentation from flag definitions.
