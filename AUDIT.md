# RESOURCE AUDIT — 2026-04-22

## Project Resource Profile
Venture is a long-running multiplayer client/server game with explicit reliability goals (high-latency multiplayer, persistent world save/load, and single-binary runtime generation). Resource lifecycles matter most on server/client runtime paths (`pkg/network`, `pkg/world`, `pkg/saveload`, persistence-heavy integration packages).

Baseline evidence:
- `go-stats-generator` (skip tests): 215,873 LOC, 12,878 methods, 107 packages.
- `go vet ./...` could not fully execute in this environment due missing system headers for Ebiten/GLFW (`X11/Xlib.h`), so static-check confirmation is partial.

Repository conventions observed:
- Strong use of `defer` immediately after acquire on read paths.
- Many write paths explicitly close gzip writers and files with error checks (`pkg/world/economy/guild_bank.go`, `pkg/integration/guild_vehicle/fleet_manager.go`, `pkg/world/persistence.go`).
- Ownership transfer patterns are generally explicit for in-memory readers/writers and manager-level `Stop`/`Shutdown` methods.

Online research (focused):
- No open GitHub issues were returned for `opd-ai/venture` leak/close/FD/connection cleanup keywords.
- Relevant Go pitfalls confirmed: unchecked `Close()` on writable files can hide flush/data-loss errors; `StdoutPipe`/`StderrPipe` must be drained concurrently before/around `Wait()` to avoid deadlocks/truncated reads.

## Resource Inventory
| Package | File Handles | DB Resources | Net Connections | Child Processes | Custom Closers | Temp Files |
|---------|-------------:|-------------:|----------------:|----------------:|---------------:|-----------:|
| `pkg/world/persistence` | 4 | 0 | 0 | 0 | 0 | 0 |
| `pkg/world/housing` | 5 | 0 | 0 | 0 | 0 | 0 |
| `pkg/integration/choice_consequences` | 2 | 0 | 0 | 0 | 0 | 0 |
| `pkg/integration/guild_vehicle` | 2 | 0 | 0 | 0 | 0 | 0 |
| `pkg/world/economy` | 2 | 0 | 0 | 0 | 0 | 0 |
| `pkg/procgen/terrain` | 2 | 0 | 0 | 0 | 0 | 0 |
| `pkg/network` | 1 | 0 | 1 listener + accepted conns | 0 | 1 (`disconnect`) | 0 |
| `pkg/network/federation` | 0 | 0 | 1 UDP packet conn | 0 | 1 (`Stop`) | 0 |
| `pkg/hostplay` | 0 | 0 | 1 probe listener | 0 | 1 (`Shutdown`) | 0 |
| `cmd/client` tests | 0 | 0 | 0 | `exec.Command`, pipes | 0 | 0 |

## Findings
### CRITICAL
- [x] Writable file close errors are suppressed in housing persistence save paths — `/home/runner/work/venture/venture/pkg/world/housing/persistence.go:51-58`, `/home/runner/work/venture/venture/pkg/world/housing/persistence.go:179-187` — `os.Create` is used for persistent save output, but deferred `file.Close()` only logs and never propagates close failure; function can return `nil` after a failed close (possible partial flush / data loss). — **Remediation:** make `Save` and `SavePlayerData` use named return `err` and set `err` on `file.Close()` failure when no prior error (same pattern already used for gzip close in this file). Validate with `go test ./pkg/world/housing` plus a fault-injection test that simulates close failure. — **COMPLETED 2026-04-22**: added named returns and close-error propagation in both save methods; added injected close-failure tests in `pkg/world/housing/persistence_test.go`.
- [x] Writable file close error is ignored in choice consequence tracker save path — `/home/runner/work/venture/venture/pkg/integration/choice_consequences/choice_tracker.go:631` — `os.Create` result is deferred with `defer file.Close()` without checking close error; save may report success while final flush fails. — **Remediation:** convert `Save` to named return and check deferred `file.Close()` error (or explicitly close and return close error after `SaveTo`). Validate with `go test ./pkg/integration/choice_consequences` and an injected writer/FS error path test. — **COMPLETED 2026-04-22**: added named-return close error propagation in `Save` and added injected close-failure test in `manager_test.go`.

### HIGH
- [ ] `exec.Cmd` stdout/stderr pipes are drained sequentially (not concurrently), creating deadlock risk in integration tests — `/home/runner/work/venture/venture/cmd/client/integration_test.go:101-104`, `/home/runner/work/venture/venture/cmd/client/integration_test.go:181-184` — goroutine performs `io.ReadAll(stdout)` then `io.ReadAll(stderr)` then `cmd.Wait()`. If child fills stderr while stdout read blocks, process can hang and retain process/pipe resources until timeout termination. — **Remediation:** drain stdout and stderr in separate goroutines, synchronize both drains, then call `cmd.Wait()` and join drains before returning. Validate by running `go test ./cmd/client -run TestHostAndPlayStartup -count=20` and confirming no timeout/hang.

### LOW
- [ ] Writable image save path does not surface `Close()` errors — `/home/runner/work/venture/venture/pkg/visualtest/snapshot.go:221` — `saveImage` uses `defer file.Close()` after `os.Create` and returns only `png.Encode` error; close failures can be missed. This is non-production tooling but still a correctness gap. — **Remediation:** use named return and propagate `file.Close()` error when encode succeeds. Validate with `go test ./pkg/visualtest` and a close-failure test double.

## False Positives Considered and Rejected
| Candidate Finding | Reason Rejected |
|-------------------|----------------|
| `pkg/network/server.go` accepted `net.Conn` appears unclosed at accept site | Rejected: connection ownership is transferred to `clientConnection`; closed via `client.disconnect()` (`/home/runner/work/venture/venture/pkg/network/server.go:1089-1101`) and on shutdown path. |
| `pkg/network/federation/discovery.go` UDP packet conn might leak | Rejected: `Stop()` closes both tickers and `ds.conn` (`/home/runner/work/venture/venture/pkg/network/federation/discovery.go:150-173`), matching lifecycle owner. |
| `pkg/integration/guild_vehicle/fleet_manager.go` and `pkg/world/economy/guild_bank.go` save/load file handling | Rejected: both implementations explicitly close gzip and file handles and return close errors on write paths. |
| `pkg/procgen/terrain/cache.go` read-path close handling | Rejected as critical leak: file is closed on all paths; only close-error reporting is weak on read path, with limited OS-resource impact versus write-path data-loss risks. |
| SQL resource leaks | Rejected: no runtime `database/sql` query/transaction usage found in compiled Go code paths audited. |
