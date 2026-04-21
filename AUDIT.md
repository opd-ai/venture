# RESOURCE AUDIT — 2026-04-21

## Project Resource Profile

Venture is a long-running multiplayer action-RPG server/client binary with no external asset files. The server process runs continuously (no fixed uptime bound), handling player connections, world persistence, guild management, federation peering, and observability. The client similarly runs for the duration of a play session and embeds a local server in solo-play mode.

**Stated reliability goals (from README)**:
- High-latency multiplayer designed for 200–5000ms latency, supporting Tor/onion services
- Persistent world state with backup rotation and incremental saves
- Single distributable binary with no external asset files
- Long-running server suitable for dedicated hosting

**Resource types used**:
- **File handles**: Save/load pipeline (world state, housing, blueprints, guild bank, fleet, choice history, terrain cache), character portrait PNG import, profanity word-list loading, visual-regression snapshots
- **Compression resources**: `compress/gzip.Writer` and `compress/gzip.Reader` throughout the persistence layer; `encoding/gob` for terrain cache
- **Network connections**: TCP connections for multiplayer (`net.Listener`, `net.Conn`), UDP PacketConn for LAN discovery, WebRTC relay connections, HTTP server for Prometheus metrics (`/metrics`)
- **Custom closers**: `ConnectionPool`, `DiscoverySystem`, `MetricsExporter` — all implement `Close()` or `Stop()` methods
- **Goroutines / tickers**: Background cleanup goroutines in `ConnectionPool`, `DiscoverySystem`, `MetricsExporter`
- **Temp files**: None in production paths; `os.MkdirTemp` used only in `examples/file_watcher_demo/main.go` with correct `defer os.RemoveAll`
- **Database resources**: None — no `database/sql`, no ORM
- **Child processes**: None — `exec.Command` is not used
- **CGo**: None in the main module (CGo is used transitively by Ebiten's GLFW, not by application code)

**Lifecycle patterns observed**:
- The save/load layer consistently uses `os.Create` → write → `defer f.Close()` or explicit close
- gzip wrapping is used for all persistent storage; close-error handling is **inconsistent** (see Findings)
- Network connections are wrapped in structs that implement cleanup methods
- The `compress/gzip.Writer` pattern is the dominant resource management concern

---

## Resource Inventory

| Package | File Handles | DB Resources | Net Connections | Child Processes | Custom Closers | Temp Files |
|---------|-------------|-------------|-----------------|----------------|---------------|------------|
| `pkg/world` | 6 | 0 | 0 | 0 | `WorldPersistence` | 0 |
| `pkg/world/housing` | 6 | 0 | 0 | 0 | `Manager`, `GuildHallManager`, `BlueprintStore` | 0 |
| `pkg/world/economy` | 2 | 0 | 0 | 0 | `GuildBankManager` | 0 |
| `pkg/network` | 1 | 0 | 6+ | 0 | `TCPServer`, `TCPClient` | 0 |
| `pkg/network/federation` | 0 | 0 | 4+ | 0 | `ConnectionPool`, `DiscoverySystem`, `FederationProtocol` | 0 |
| `pkg/network/federation/webrtc` | 0 | 0 | 2 | 0 | `STUNClient`, `RelayManager` | 0 |
| `pkg/saveload` | 4 | 0 | 0 | 0 | `SaveManager` | 0 |
| `pkg/integration/choice_consequences` | 2 | 0 | 0 | 0 | `ChoiceTracker` | 0 |
| `pkg/integration/guild_vehicle` | 2 | 0 | 0 | 0 | `FleetManager` | 0 |
| `pkg/procgen/terrain` | 2 | 0 | 0 | 0 | `TerrainCache` | 0 |
| `pkg/observability` | 0 | 0 | 1 | 0 | `MetricsExporter` | 0 |
| `pkg/engine/character_creation` | 1 | 0 | 0 | 0 | — | 0 |
| `pkg/visualtest` | 4 | 0 | 0 | 0 | — | 0 |
| `examples/file_watcher_demo` | 0 | 0 | 0 | 0 | — | 1 (cleaned) |

---

## Findings

### HIGH

- [ ] **gzip.Writer close error silently dropped in `GuildBankManager.Save`** — `pkg/world/economy/guild_bank.go:546` — **Evidence**: `gzipWriter := gzip.NewWriter(file)` is created at line 545, then `defer gzipWriter.Close()` is placed at line 546. `json.NewEncoder(gzipWriter).Encode(m.vaults)` runs successfully, and the function returns `nil`. At that point the deferred `gzipWriter.Close()` executes; `gzip.Writer.Close()` flushes all internally buffered compressed data and writes the gzip stream footer (8 bytes). If `Close()` fails (full disk, I/O error, network filesystem disconnect), the return value is discarded, the function already returned `nil` to the caller, and the file on disk is a truncated, structurally invalid gzip stream. **Impact**: Silent data loss — guild vault gold balances and item inventories are saved to a file that cannot be decompressed on the next load. The `Load()` path will fail with "gzip: invalid header" or similar on restart. In a long-running server this is a CRITICAL data loss path. **Remediation**: Replace the deferred close with an explicit checked close before `return nil`:
  ```go
  // After encoder.Encode succeeds:
  if err := gzipWriter.Close(); err != nil {
      file.Close()
      return fmt.Errorf("failed to flush gzip writer: %w", err)
  }
  // Close the file explicitly and check the error for the write path:
  if err := file.Close(); err != nil {
      return fmt.Errorf("failed to close file: %w", err)
  }
  return nil
  ```
  Remove `defer file.Close()` and `defer gzipWriter.Close()`. Validate with: inspect the function returns `error` on a mock `io.Writer` that returns an error during `Close()`.

- [ ] **gzip.Writer close error silently dropped in `Blueprint.Export`** — `pkg/world/housing/blueprint.go:248` — **Evidence**: `gzipWriter := gzip.NewWriter(file)` at line 247 with `defer gzipWriter.Close()` at line 248. `json.NewEncoder(gzipWriter).Encode(bp)` executes and the function returns `nil`. The deferred `gzipWriter.Close()` then runs; if it fails, the exported blueprint file is a truncated gzip stream. **Impact**: Exported blueprint files are silently written as corrupt gzip streams — attempting to re-import them (via `ImportBlueprint`) will fail with a gzip decode error. Players lose blueprint data without any error feedback. **Remediation**: Apply the same pattern as `pkg/world/housing/persistence.go:61–70` (named return + deferred close with error capture):
  ```go
  func (bp *Blueprint) Export(filepath string) (err error) {
      file, err := os.Create(filepath)
      if err != nil { return fmt.Errorf("failed to create file: %w", err) }
      defer func() {
          if closeErr := file.Close(); closeErr != nil && err == nil {
              err = fmt.Errorf("failed to close blueprint file: %w", closeErr)
          }
      }()
      gzipWriter := gzip.NewWriter(file)
      defer func() {
          if closeErr := gzipWriter.Close(); closeErr != nil && err == nil {
              err = fmt.Errorf("failed to flush gzip writer: %w", closeErr)
          }
      }()
      encoder := json.NewEncoder(gzipWriter)
      encoder.SetIndent("", "  ")
      return encoder.Encode(bp)
  }
  ```
  Validate: write a test that passes a failing `io.Writer` and asserts a non-nil error is returned.

---

### MEDIUM

- [ ] **gzip.Writer close error silently dropped in `ChoiceTracker.SaveTo`** — `pkg/integration/choice_consequences/choice_tracker.go:595` — **Evidence**: `gzWriter := gzip.NewWriter(w)` at line 594, `defer gzWriter.Close()` at line 595. `SaveTo` is called from `Save(filename string)` at line 629, which passes an `*os.File`. The `defer gzWriter.Close()` runs after `return nil`; any flush error is discarded. `Save(filename)` sees nil from `SaveTo` and returns nil to its own caller. **Impact**: Player choice history and NPC relationship data can be silently written as an incomplete gzip stream. On next load, `LoadFrom` will fail to decompress the file, and the choice history is lost. **Remediation**: Change `SaveTo` to return the gzip close error:
  ```go
  func (ct *ChoiceTracker) SaveTo(w io.Writer) error {
      // ... existing lock + encode logic ...
      if err := encoder.Encode(ct.players); err != nil {
          return fmt.Errorf("failed to encode data: %w", err)
      }
      if err := gzWriter.Close(); err != nil {
          return fmt.Errorf("failed to flush gzip writer: %w", err)
      }
      return nil
  }
  ```
  Remove `defer gzWriter.Close()`. Validate with `go vet ./pkg/integration/choice_consequences/...` and a unit test that uses a failing writer.

- [ ] **gzip.Writer close error silently dropped in `GuildHallManager.Save`** — `pkg/world/housing/guildhall_manager.go:235` — **Evidence**: `gzWriter := gzip.NewWriter(w)` at line 234, `defer gzWriter.Close()` at line 235. `encoder.Encode(m.guildHalls)` succeeds and `return encoder.Encode(...)` returns nil, then the deferred `gzWriter.Close()` runs with its error discarded. No production callers were found in this audit (only a test using `bytes.Buffer`), but the method signature accepts `io.Writer`, allowing callers to pass file handles in the future. **Impact**: If called with a file-backed writer (e.g., during a save-to-disk integration), guild hall progression data (construction phases, resource pools) can be silently written as a corrupt gzip stream. **Remediation**:
  ```go
  func (m *GuildHallManager) Save(w io.Writer) error {
      m.mu.RLock()
      defer m.mu.RUnlock()
      gzWriter := gzip.NewWriter(w)
      encoder := json.NewEncoder(gzWriter)
      if err := encoder.Encode(m.guildHalls); err != nil {
          gzWriter.Close() // best-effort on error path
          return fmt.Errorf("encode guild halls: %w", err)
      }
      if err := gzWriter.Close(); err != nil {
          return fmt.Errorf("flush gzip writer: %w", err)
      }
      return nil
  }
  ```
  Validate by running `go vet ./pkg/world/housing/...` and confirming test `TestGuildHallManagerSaveLoad` still passes.

---

### LOW

- [ ] **`copyFile` in `WorldPersistence` drops close error on backup copy** — `pkg/world/persistence.go:297` — **Evidence**: In `copyFile`, the destination file is closed via a deferred anonymous function: `defer func() { dstFile.Close(); if copyErr != nil { os.Remove(dst) } }()`. The `dstFile.Close()` return value is discarded. `dstFile.Sync()` is called before the deferred return (which mitigates most data loss risk), but the close error is not propagated to the caller. This is used to create backup save files during `rotateBackups()`. **Impact**: A failed `Close()` (e.g., NFS write error) is not surfaced; the backup rotation proceeds as if the backup is valid. If the primary save is later lost and the backup cannot be read, world data is unrecoverable. **Remediation**: Assign the close error to `copyErr` if it is currently nil:
  ```go
  defer func() {
      if closeErr := dstFile.Close(); closeErr != nil && copyErr == nil {
          copyErr = fmt.Errorf("failed to close destination: %w", closeErr)
          os.Remove(dst)
      } else if copyErr != nil {
          dstFile.Close()
          os.Remove(dst)
      }
  }()
  ```
  Validate with `go vet ./pkg/world/...`.

- [ ] **`createBackup` in `SaveManager` drops backup file close error** — `pkg/saveload/recovery.go:62` — **Evidence**: `defer backup.Close()` at line 62 discards the close error. Unlike `copyFile` in the world package, this backup path does not call `Sync()` before returning. If the OS write buffer is not flushed (which is the common case on Linux with writeback caching), the file descriptor `Close()` is where the final flush occurs, and any error from it is lost. **Impact**: The `.bak` backup may be silently incomplete. If the primary save is later corrupted and `recoverFromBackup` is invoked, it reads the backup with `os.ReadFile` and tries to unmarshal it. A truncated backup will fail to unmarshal, leaving the player with an unrecoverable save. **Remediation**: Use a named return to capture and return the close error:
  ```go
  func (m *SaveManager) createBackup(name string) (backupPath string, err error) {
      // ... existing code up through io.Copy ...
      defer func() {
          if closeErr := backup.Close(); closeErr != nil && err == nil {
              err = errors.FileSystemWrap(closeErr, "failed to close backup file").
                  WithContext("backupPath", backupPath)
          }
      }()
      // ... rest of function
  }
  ```
  Validate with `go vet ./pkg/saveload/...`.

- [ ] **`writeWorldData` double-closes the temporary save file** — `pkg/world/persistence.go:147,172` — **Evidence**: The deferred anonymous function at line 147 always calls `f.Close()`. The explicit `f.Close()` at line 172 is also called and its error is checked. On the happy path, `f.Close()` succeeds at line 172, the function returns nil, and then the deferred `f.Close()` runs again, receiving `os.ErrClosed`. The error from the second close is silently ignored in the defer. **Impact**: Benign — `*os.File.Close()` on an already-closed file returns `os.ErrClosed` but does not panic and does not double-free the file descriptor. The `saveErr != nil` guard prevents spurious cleanup. **Remediation**: Add a sentinel guard to avoid the second close, or restructure to not call `f.Close()` explicitly and instead use only the deferred close with a named return:
  ```go
  func (w *WorldPersistence) writeWorldData(...) (err error) {
      f, openErr := os.Create(tempPath)
      if openErr != nil { return fmt.Errorf(...) }
      defer func() {
          if closeErr := f.Close(); closeErr != nil && err == nil {
              err = fmt.Errorf("failed to close temp file: %w", closeErr)
          }
          if err != nil { os.Remove(tempPath) }
      }()
      // ... gz.Close() with error check, return gz close error
  }
  ```
  Validate with `go vet ./pkg/world/...`.

- [ ] **`TerrainCache.saveToDisk` drops file close error** — `pkg/procgen/terrain/cache.go:289` — **Evidence**: `defer file.Close()` at line 289 after `os.Create` discards the close error. The terrain cache is a non-critical, reproducible disk cache; the file is a `.gob` file used to avoid re-generating terrain on server restart. **Impact**: A failed close leaves a potentially truncated `.gob` file on disk. On next load, `loadFromDisk` will detect the checksum mismatch and delete the file, then regenerate from seed — no data loss, only a missed cache hit. **Remediation**: Given the self-healing nature of the cache (corrupted files are detected and regenerated), this is acceptable as-is. If stricter integrity is desired, add a log warning when the close fails:
  ```go
  defer func() {
      if err := file.Close(); err != nil {
          logrus.WithFields(logrus.Fields{"key": key, "filename": filename, "error": err}).
              Warn("Failed to close terrain cache file; cache entry may be corrupt")
      }
  }()
  ```
  Validate: no action required unless stricter cache guarantees are desired.

---

## False Positives Considered and Rejected

| Candidate Finding | Reason Rejected |
|-------------------|----------------|
| `pkg/world/housing/persistence.go` `Save()` `defer gzWriter.Close()` | Uses named return `err` with a deferred func that assigns `closeErr` to the named return when the encode succeeded. This is the correct pattern — no error is dropped. |
| `pkg/world/housing/persistence.go` `SavePlayerData()` `defer gzWriter.Close()` | Same named-return pattern as `Save()`. Error is properly propagated. |
| `pkg/world/persistence.go:381` `defer gz.Close()` in `openAndDecodeState` | This is a `gzip.Reader` (not Writer). Closing a gzip reader only releases its internal decompressor state; no buffered writes can be lost. Suppressing the error is standard for readers. |
| `pkg/social/persistence/trust_manager.go:194` | Writes to `bytes.Buffer` via `gzip.NewWriter(&buf)`. `bytes.Buffer.Write` never returns an error; `gzip.Writer.Close()` writing to a `bytes.Buffer` cannot fail. The `if err := gzipWriter.Close(); err != nil` check is already present and correct. |
| `pkg/social/persistence/reputation_manager.go:208` | Same as above — writes to `bytes.Buffer`, close error is already checked. |
| `pkg/integration/guild_vehicle/fleet_manager.go` `Save()` | Explicitly calls `gzWriter.Close()` and checks the error before `file.Close()`. Correctly implemented. |
| `pkg/network/federation/connectionpool.go` `Close()` | `ConnectionPool.Close()` iterates all connections and calls `conn.Conn.Close()` on each. The cleanup goroutine also closes stale connections. Proper resource lifecycle. |
| `pkg/network/server.go` `disconnect()` | `clientConnection.disconnect()` sets connected=false and calls `c.conn.Close()` with error logging via `logrus.WithError(err).Warn(...)`. Proper resource lifecycle. |
| `pkg/hostplay/host_and_play.go` `FindAvailablePort()` | Opens a `net.Listener` to probe a port, then immediately calls `listener.Close()` before returning the port number. No leak — the listener lifetime is scoped to the probe. |
| `examples/file_watcher_demo/main.go` `os.MkdirTemp` | Uses `defer os.RemoveAll(tmpDir)` in `main()`. Correct cleanup — this is an example/demo binary. |
| `pkg/procgen/terrain/cache.go` `loadFromDisk` | `file.Close()` is called on all three code paths: (a) decode error → `file.Close()` then return nil, (b) decode success → `file.Close()` before checksum validation, (c) checksum mismatch → file is already closed at step (b). No leak. |
| `pkg/network/federation/webrtc/relay.go` `pingRelay` | Dials a UDP connection with `defer conn.Close()`. Connection is closed when the probe function returns. Proper lifecycle. |
| `pkg/world/housing/blueprint.go` `ImportBlueprint` | Opens file with `defer file.Close()` and creates gzip reader with `defer gzipReader.Close()`. Both are readers — no write buffers to flush. Correct. |
| `pkg/world/economy/guild_bank.go` `Load()` | Opens file with `defer file.Close()` and gzip reader with `defer gzipReader.Close()`. Read path — no buffered writes. Correct. |
