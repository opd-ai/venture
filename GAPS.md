# Resource Management Gaps — 2026-04-22

## Gap 1: Persistent save APIs do not uniformly propagate writable file close failures
- **Stated Goal**: Reliable long-running multiplayer operation with persistent world/player data integrity.
- **Current State**: Some save paths rigorously propagate close/flush errors (e.g., guild bank, fleet manager, world persistence), but housing and choice consequence save paths can return success after `file.Close()` failure.
- **Risk**: Silent data loss or corrupted save artifacts under disk-full/NFS-writeback/failure-on-close conditions.
- **Closing the Gap**: Standardize a single write-path idiom for all persistence APIs: check encoder/gzip close, then check file close and return failure if it fails. Add failure-injection tests for close errors in each persistence package.

## Gap 2: Child-process pipe handling pattern in integration tests is not robust against deadlock
- **Stated Goal**: Stable CI/testing quality gates for a large codebase.
- **Current State**: `cmd/client` integration tests read `StdoutPipe` then `StderrPipe` sequentially before `Wait()` in shared goroutines.
- **Risk**: Pipe-buffer deadlocks and intermittent test hangs/timeouts; can leave process/pipe resources active until timeout handling executes.
- **Closing the Gap**: Adopt a reusable helper that concurrently drains stdout/stderr, waits for both drains, then performs/validates `Wait()`. Use this helper in all process-based integration tests.

## Gap 3: Resource lifecycle policy is implicit rather than centrally documented/enforced
- **Stated Goal**: Maintainability and reliability across 100+ packages and long-lived server/client runtimes.
- **Current State**: Strong local patterns exist but there is no single root policy/checklist that enforces write-close error propagation, pipe-drain ordering, and ownership-transfer documentation.
- **Risk**: Regressions as new persistence and process-invocation code is added, with inconsistent handling across subsystems.
- **Closing the Gap**: Add a root-level resource-lifecycle checklist to contributor docs and CI review templates; include explicit rules for writable `Close()` error propagation, process pipe draining, and owner-responsible close semantics.
