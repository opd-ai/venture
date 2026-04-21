# Resource Management Gaps — 2026-04-21

> This file documents gaps in resource lifecycle management relative to the project's
> stated goals of long-running reliability, persistent world state, and data integrity.
> All prior non-resource-management gaps previously recorded here have been migrated to
> the repository issue tracker and are not reproduced below.

---

## Gap R1 — Inconsistent gzip.Writer Close-Error Handling Across Persistence Layer

- **Stated Goal**: Venture targets continuous-uptime server operation with "persistent world state with backup rotation and incremental saves" (README). The save/load pipeline is on the critical path for player data integrity — guild vault contents, blueprints, choice histories, and guild hall state are all persisted via gzip-compressed JSON files.
- **Current State**: The project has **two coexisting patterns** for closing a `gzip.Writer` to a file:
  1. **Correct pattern** (used in `pkg/world/persistence.go`, `pkg/world/housing/persistence.go`, `pkg/integration/guild_vehicle/fleet_manager.go`, `pkg/network/federation/guild/persistence.go`, `pkg/integration/political_warfare/manager.go`, `pkg/social/persistence/trust_manager.go`, `pkg/engine/prestige/manager.go`): explicit `if err := gzWriter.Close(); err != nil { return err }` or a named-return deferred func that captures the close error.
  2. **Incorrect pattern** (used in `pkg/world/economy/guild_bank.go`, `pkg/world/housing/blueprint.go`, `pkg/world/housing/guildhall_manager.go`, `pkg/integration/choice_consequences/choice_tracker.go`): `defer gzWriter.Close()` whose return value is discarded. Because the `gzip.Writer` buffers compressed output internally, a failed `Close()` leaves the destination without the gzip stream footer (and potentially missing the last compressed block), producing a file that cannot be decompressed.
- **Risk**: Silent data corruption on write. `GuildBankManager.Save` and `Blueprint.Export` write to an `*os.File` on the critical path. If the file system returns an I/O error during `gzip.Writer.Close()` (disk full, NFS disconnect, hardware error), the function returns `nil` to the caller, the caller assumes success, and the resulting file is a structurally invalid gzip stream. On the next server start or player import, the load path returns a hard error and the data cannot be recovered. Guild vault items and gold balances, or player-shared blueprint files, are permanently lost without any error ever being surfaced.
- **Closing the Gap**:
  1. Adopt the named-return deferred pattern used in `pkg/world/housing/persistence.go` as the project standard for all `gzip.Writer`-to-file operations:
     ```go
     func ExampleSave(w io.Writer) (err error) {
         gz := gzip.NewWriter(w)
         defer func() {
             if closeErr := gz.Close(); closeErr != nil && err == nil {
                 err = fmt.Errorf("flush gzip: %w", closeErr)
             }
         }()
         return json.NewEncoder(gz).Encode(data)
     }
     ```
  2. Audit and fix the four identified sites: `guild_bank.go:546`, `blueprint.go:248`, `guildhall_manager.go:235`, `choice_tracker.go:595` (see AUDIT.md for exact remediation per site).
  3. Add a project-wide CI check or linter rule (e.g., a `grep` in the Makefile's `lint` target) that flags `defer.*gzip.*\.Close()` without a named-return variable, preventing regression.

---

## Gap R2 — Backup File Close Errors Not Propagated, Undermining Recovery Guarantee

- **Stated Goal**: `SaveManager.SaveGameWithBackup` and `SaveManager.LoadGameWithRecovery` document a recovery guarantee: "automatic backup and checksum validation" and "automatic corruption detection and recovery." The recovery path relies on a valid `.bak` file being present after a successful `createBackup` call.
- **Current State**: `pkg/saveload/recovery.go:62` uses `defer backup.Close()` which discards the close error. The `createBackup` function does not call `backup.Sync()` before returning, so for write-back caches (the default on Linux ext4) the OS may not have flushed the write buffer to disk when `Close()` is called. If `Close()` fails, the `.bak` file is silently incomplete; `createBackup` returns the backup path and a nil error, causing the caller to believe the backup is valid. When `recoverFromBackup` later reads the file, it will fail to unmarshal the truncated data, and the recovery path returns `false, nil` — leaving the player with an unrecoverable, corrupted primary save.
- **Risk**: The recovery guarantee is illusory in the presence of I/O errors during backup creation. Players who trust the recovery path may lose save data despite the feature existing.
- **Closing the Gap**:
  1. Add `Sync()` before `Close()` in `createBackup` (mirroring `copyFile` in `pkg/world/persistence.go:305`):
     ```go
     if _, err := io.Copy(backup, source); err != nil { return "", ... }
     if err := backup.Sync(); err != nil { return "", ... }
     // defer backup.Close() or explicit close with error check
     ```
  2. Capture and return the close error using a named return (see Gap R1 pattern).
  3. Add a test that injects a failing `Close()` on the backup file and asserts that `createBackup` returns a non-nil error and `SaveGameWithBackup` does NOT proceed.

---

## Gap R3 — No Automated Resource Leak Detection in CI

- **Stated Goal**: The `Makefile` defines `make quality` as running all quality validations, and CI enforces build + test pass. The project explicitly calls out high-latency, long-running server operation as a design target, where resource leaks accumulate over time and cause failures.
- **Current State**: The CI pipeline (`scripts/validate-network-types.sh`) validates network interface types. There is no analogous check for resource lifecycle:
  - No `go vet` shadow or `staticcheck` pass that detects `defer io.Closer.Close()` with a discarded return value
  - No test that exercises the persistence layer with a slow/failing writer and asserts that errors are surfaced
  - No `lsof`-based FD count regression test for the server process
- **Risk**: The pattern of `defer gzip.Writer.Close()` that silently drops errors recurred in four separate files written at different times, suggesting it is an easy mistake to make without automated detection. Without a gate, the pattern will recur as the persistence layer grows. For a game server intended to run continuously, even infrequent I/O errors become likely over long uptime; unreported gzip close failures lead to the data loss scenarios described in Gaps R1 and R2.
- **Closing the Gap**:
  1. Add to `Makefile.quality`:
     ```makefile
     vet-closer-errors:
         @echo "Checking for deferred gzip.Writer.Close() with discarded error..."
         @if grep -rn --include="*.go" 'defer.*\.Close()' pkg/ | \
            grep -E 'gz|gzip' | grep -v '_test.go' | grep -v 'err.*Close\|Close.*err'; then \
            echo "ERROR: Found deferred gzip Close() calls with potentially discarded errors"; exit 1; \
         fi
     ```
  2. Add integration tests that use `iotest.ErrWriter` (or a custom `failingWriter`) as the backing writer for each `Save()` function and assert the returned error is non-nil.
  3. Consider adding `staticcheck` (`honnef.co/go/tools/cmd/staticcheck`) to the CI pipeline, which detects discarded error returns from `io.Closer.Close()` (rule `SA1004` family).

---

## Gap R4 — `WorldPersistence.writeWorldData` Double-Close Creates Misleading Code Pattern

- **Stated Goal**: The save pipeline uses a temporary file + atomic rename pattern (`writeWorldData` → temp file → `atomicReplace`) to ensure the world save is never partially written. This is a correct reliability technique.
- **Current State**: `writeWorldData` (`pkg/world/persistence.go:135–175`) both explicitly calls `f.Close()` (with error checking, at line 172) and also closes `f` via a deferred anonymous function (line 147). On the happy path, `f` is closed twice. The second close in the defer returns `os.ErrClosed` but the error is discarded. The `saveErr != nil` guard in the defer prevents accidental file deletion after a successful close, so there is no functional impact today. However, the pattern misleads future contributors: it implies a "close in defer is the primary cleanup mechanism" idiom, which (as seen in Gaps R1 and R2) causes actual bugs when the close error is discarded.
- **Risk**: A future contributor following the double-close pattern in a write-path context could produce a silent data loss bug by analogy. The pattern also wastes a system call (the second `f.Close()` on an already-closed descriptor) in the hot save path.
- **Closing the Gap**: Refactor `writeWorldData` to use a single close mechanism — either all-explicit (current, but remove the deferred `f.Close()`) or all-deferred using the named-return pattern (see Gap R1). The explicit approach is already present and correct; simply remove the redundant `f.Close()` from the defer:
  ```go
  defer func() {
      // Note: f.Close() is handled explicitly below to capture its error.
      if saveErr != nil {
          os.Remove(tempPath)
      }
  }()
  ```
