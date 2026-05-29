# UNIVERSAL BUG AUDIT — 2026-05-29

## Project Profile

**Venture** is a fully procedural, multiplayer action-RPG written in Go (1.24.5) on the
Ebiten game engine. All graphics, audio, items, quests, NPCs, terrain, and dialog are
generated at runtime from a deterministic, seed-based procedural-generation pipeline; the
project ships as a single binary per platform with **no external asset files**.

- **Target users:** players (desktop native + WebAssembly builds) and self-hosters running
  the dedicated server (`cmd/server`), including over high-latency / Tor onion routing.
- **Deployment model:** single distributable binary. Desktop clients auto-start a localhost
  server for solo play; WASM clients must connect to an external server (embedded server is
  compiled out).
- **Critical paths (primary stated goals):**
  1. **Deterministic seed-based procedural generation** (`pkg/procgen/**`) — the central
     correctness contract. The project documents that *save format and network protocol are
     stable within a MAJOR version* (`pkg/version/version.go:23` → v1.0.0), which makes
     generator output stability a hard requirement.
  2. **Multiplayer networking & federation** (`pkg/network/**`) — untrusted input from remote
     peers/clients; the main trust boundary.
  3. **Save/load integrity** (`pkg/saveload/**`) — data-corruption surface.
  4. **Modding sandbox** (`pkg/modding/**`) — loads user-supplied JSON; a trust boundary.

**Trust boundaries:** untrusted input enters at (a) federation handshakes/messages from
remote servers (`pkg/network/federation`), (b) client→server gameplay messages
(`pkg/network`), (c) mod files loaded from disk (`pkg/modding`), and (d) save names / save
data (`pkg/saveload`). Each was traced for validation depth (see findings).

## Audit Scope

- **Packages audited:** 109 packages / 1,299 source files. Deep scrutiny on the critical-path
  packages above; structural/automated scan across the remainder.
- **Tooling:** `go-stats-generator v1.0.0` (functions, packages, documentation, duplication,
  patterns, interfaces, structs), `go vet ./...`, and `go test -race ./...` (run under
  `xvfb-run` because the client links CGO/X11 via Ebiten).
- **Methodology:** README/version claims extracted as acceptance criteria → baseline metrics →
  systematic bug-class checklist (logic, nil/boundary, error handling, resource lifecycle,
  concurrency, security, aliasing, init order, API contracts, performance) → Phase 3l
  false-positive screening (data-flow trace, upstream-guard check, comment review,
  exploitability assessment, test evidence) on every candidate.

## Goal-Achievement Summary

| Stated Goal | Status | Blocking Findings |
|-------------|--------|-------------------|
| Deterministic, seed-based generation reproducible for a given seed | ⚠️ | C-1 (output is stable per build but has drifted from the recorded v1.0.0 baseline; `go test` is red) |
| Save format stable within a MAJOR version (v1.x) | ⚠️ | C-1 (Item generator output changed without a MAJOR bump) |
| Single binary, no external asset files (all content generated at runtime) | ✅ | None — no `//go:embed` of assets; generation is fully procedural |
| Multiplayer over high-latency / Tor links; replay-protected federation | ✅ (with caveat) | H-1 (per-handshake goroutine spawn on the untrusted path) |
| Save/load integrity validation | ✅ (with caveat) | M-1 (WASM "sha256" key actually stores a non-crypto FNV-1a digest) |
| Modding sandbox prevents path traversal / resource abuse | ✅ | None — `ValidatePath` resolves symlinks, enforces root prefix, rejects `..`; called before every read |
| WASM build disables embedded server | ✅ | None — verified in `cmd/client/main.go:48`,`:73` |

## Findings

### CRITICAL

- [ ] **C-1: Item generator output has drifted from the v1.0.0 determinism baseline** —
  `pkg/procgen/audit/baseline.go:25` (baseline) ⇄ `pkg/procgen/item/generator.go:79`
  (`ItemGenerator.Generate`) — **bug class: API/behavioral contract + logic regression** —
  The project's own version-stability guard `TestDeterminism_VersionStability/Item`
  (`pkg/procgen/audit/determinism_test.go:336`) **fails**: for the fixed `BaselineSeed`
  (99999), the SHA256 of the JSON-serialized item output is `afba81832bf6ffe8…` but the
  recorded v1.0.0 baseline is `2b36ce659bf7c7b6…`. I ran the test three times under `xvfb`:
  the new hash is **stable across runs** (`afba81832bf6ffe8` every time), so generation is
  still internally deterministic — but its output has *changed* since the baseline was
  recorded. Only the `Item` generator drifted; the other 13 registered generators
  (Book, Building, Entity, Magic, Quest, …) all still match their baselines.
  **Concrete consequence:** `pkg/version/version.go` pins the release at **1.0.0** and states
  "Save file format is backward-compatible within a MAJOR version" and "Network protocol is
  stable within a MAJOR version." A changed item-generation output under the *same* MAJOR
  version means two builds both labelled `1.0.0` produce different items for the same seed —
  breaking save reproducibility and multiplayer snapshot determinism for items, the exact
  guarantee the README headlines ("deterministic, seed-based procedural generation"). The
  committed test suite is also red, so CI cannot distinguish this regression from future ones.
  **Remediation:** Decide intent and act in `pkg/procgen/item`:
  - If the item-output change was **unintentional**, `git bisect` the `pkg/procgen/item`
    change that moved the hash off `2b36ce659bf7c7b6` and restore the prior generation logic.
  - If the change was **intentional**, bump the version (`pkg/version/version.go`) per the
    project's own semver policy (a save/protocol-affecting change warrants at least a MINOR,
    arguably MAJOR, bump), then regenerate and commit the new baseline (`baseline.go:25` and
    `baseline_hashes.json`, updating `BaselineVersion`).
  Validate with: `xvfb-run -a go test ./pkg/procgen/audit/ -run TestDeterminism_VersionStability -count=1`.

### HIGH

- [ ] **H-1: Unbounded goroutine spawn per inbound federation handshake** —
  `pkg/network/federation/handshake.go:294` (`ProcessHandshake` → `go hm.cleanupNonces()`) —
  **bug class: concurrency / performance on an untrusted-input path** — Every call to
  `ProcessHandshake` launches a brand-new goroutine solely to sweep the `seenNonces` map, and
  each such goroutine immediately takes `hm.mu.Lock()` (`handshake.go:303`). `ProcessHandshake`
  is reached from `protocol.go:154` when a *remote* peer sends a handshake/response, so the
  spawn rate is controlled by external traffic. Under a burst of handshakes (the README
  explicitly targets federation over Tor, where reconnect storms are common) this produces a
  goroutine per message, all contending for the same write lock, amplifying load and lock
  contention. `go vet` is clean and `go test -race` reports **no data race** here (the map is
  correctly guarded), so this is a resource/perf defect, not a race.
  **Concrete consequence:** a peer can cheaply force the server to create and schedule one
  goroutine + one exclusive-lock acquisition per handshake — a mild DoS amplification and
  needless GC/scheduler churn on the primary networking path.
  **Remediation:** Replace the per-call `go hm.cleanupNonces()` with a single periodic sweeper
  owned by `HandshakeManager` (e.g., a `time.Ticker` started in `NewHandshakeManager`, stopped
  via a `Close()`), or gate cleanup behind a `sync.Once`-guarded ticker / a "last swept"
  timestamp so cleanup runs at most once per interval. Keep the existing lock discipline.
  Validate with: `go test -race ./pkg/network/federation/...` and a benchmark that calls
  `ProcessHandshake` in a tight loop asserting bounded `runtime.NumGoroutine()` growth.

### MEDIUM

- [ ] **M-1: WASM save integrity uses a non-cryptographic digest stored under a `.sha256`
  key** — `pkg/saveload/storage_wasm.go:782` (`computeChecksum`) and its callers at
  `:474`,`:493`,`:742`,`:774` — **bug class: API/behavioral contract (misleading invariant)**
  — On WASM the integrity value is computed with **FNV-1a** (`computeChecksum`) but persisted
  under keys suffixed `".sha256"` (e.g. `key+".sha256"`). `pkg/saveload/doc.go:63` does honestly
  document WASM as using "FNV-1a checksums", so the *documentation* is consistent; the defect is
  the **internal naming** — a `.sha256` key holding a 64-bit FNV digest is a trap for future
  maintainers and any cross-platform recovery code that assumes the suffix names the algorithm.
  The desktop path (`pkg/saveload/recovery.go:27`) correctly uses `crypto/sha256`.
  **Concrete consequence:** FNV-1a detects accidental corruption but provides **no tamper
  resistance**; combined with the misleading suffix, a maintainer could wrongly assume WASM
  saves are cryptographically integrity-checked. In single-player WASM the save is not a
  security boundary, so impact is limited to correctness/maintainability — hence MEDIUM, not
  HIGH.
  **Remediation:** Either (a) rename the WASM key suffix to reflect the algorithm (e.g.
  `".fnv1a"` / `".checksum"`) and update the matching reads at `:644`,`:696`,`:742`,`:774`, or
  (b) use `crypto/sha256` on WASM too (it is supported under `GOOS=js`; the comment at
  `storage_wasm.go:781` that "crypto/sha256 may have WASM compatibility issues" is outdated).
  Validate with: `GOOS=js GOARCH=wasm go build ./pkg/saveload/...` and the existing
  `pkg/saveload` checksum tests.

### LOW

- [ ] **L-1: 16 MB analysis artifact `post.json` committed at the repo root** —
  `post.json` (repository root) — **bug class: repository hygiene** — `post.json` is a
  16,526,322-byte go-stats-generator output dump (its `metadata.repository` field points at a
  developer's local `/home/user/go/src/...` path). It is not referenced by any Go source
  (`grep` finds zero references) and is not in `.gitignore`.
  **Concrete consequence:** bloats every clone by ~16 MB and leaks a local filesystem path; no
  functional impact. **Remediation:** `git rm post.json` and add `post.json` (or the tool's
  output glob) to `.gitignore`. No build/test validation required (non-code change).

- [ ] **L-2: Outdated comment claims `crypto/sha256` is WASM-incompatible** —
  `pkg/saveload/storage_wasm.go:781` — **bug class: documentation** — The comment justifying
  the FNV-1a fallback ("Uses a simple hash since crypto/sha256 may have WASM compatibility
  issues") is inaccurate; `crypto/sha256` compiles and runs under `GOOS=js`. **Remediation:**
  correct or remove the comment (tie this to M-1's resolution). Non-code/comment change.

- [ ] **L-3: Extreme single-function size on a critical init path** —
  `pkg/engine/system_init.go` `InitializeGameSystems` (~1,808 lines, cyclomatic 20, the
  longest function in the repo per go-stats-generator) — **bug class: maintainability / code
  smell** — This is the boot path that wires all 66 systems; its length plus the many
  near-identical 27-line duplicate blocks flagged by go-stats-generator
  (`system_init.go:650,1272,1273,1279,1444,1452,1453,…`) make ordering bugs and drift between
  the duplicated registration blocks easy to introduce. No defect confirmed (`go vet` clean,
  client tests pass), so this is LOW/theoretical. **Remediation (proportionate):** extract the
  repeated system-registration block into a small helper and call it per system; do not
  restructure initialization order. Validate with `xvfb-run -a go test ./cmd/client/... ./pkg/engine/...`.

## Metrics Snapshot

| Metric | Value |
|--------|-------|
| Total functions (free functions) | 4,127 |
| Total methods | 12,924 |
| Functions above complexity 15 | ~6 (top: `pionConnect` 31.4, `updatePoseFromFrame` 30.6, `Tick` 29.8) |
| Functions with complexity > 10 | 95 |
| Avg cyclomatic complexity | 4.0 |
| Functions > 50 lines | 681 (4.0%) |
| Doc coverage (overall) | 90.0% (pkg 99.1%, func 96.6%, type 89.3%, method 88.4%) |
| Duplication ratio | 5.39% (635 clone pairs, 21,936 dup lines) |
| Circular dependencies | 0 |
| Test pass rate (packages) | 113 passed / 114 testable (1 FAIL: `pkg/procgen/audit` — see C-1); 13 pkgs have no tests |
| `go test -race` data races | 0 |
| `go vet ./...` warnings | 0 |

## False Positives Considered and Rejected

| Candidate | Reason Rejected |
|-----------|----------------|
| `math/rand` used for guild identity (`pkg/network/federation/guild/identity.go:26`) | Intentional **deterministic** procgen seeded by `int64` — not a security context. Matches the project's seed-based generation design. |
| Federation tokens/nonces use weak randomness | They use `crypto/rand` (`auth.go:196,216`); the handshake nonce uses `crypto/rand` (`handshake.go:112`). Correct. |
| Missing handshake replay protection | Implemented: ed25519 signature, public-key↔ServerID fingerprint binding, 60 s timestamp window, and a seen-nonce set (`handshake.go:155-257,283-291`). |
| Mod loader path traversal | `LoadFromFile` calls `validateSandboxPath` *before* `os.ReadFile` (`loader.go:40-44`); `Sandbox.ValidatePath` resolves symlinks, enforces the mods-dir prefix, and rejects `..` (`sandbox.go:92-133`). |
| Save name path traversal | `ValidateSaveName` rejects `/`, `\`, and shell/Windows metacharacters (`validation.go:28-37`); a bare `..` cannot traverse without a separator. |
| `InsecureSkipVerify` / hardcoded secrets / `os/exec` injection | None present — grep across non-test `.go` returns no matches. |
| Division-by-zero in `RotationComponent.Update` | Guarded: `if deltaTime <= 0 { return false }` before the division (`rotation_component.go:104-112`). |
| `recover()` swallowing panics | The `pkg/recovery` helpers log via logrus and run cleanup; they do not silently discard (`panic_recovery.go:60-100`). |
| Loop-variable capture in goroutine closures | Module is Go 1.24 (`go.mod`), where per-iteration loop variables are the default; pre-1.22 capture bug does not apply. |
| `//go:embed` of `.env`/keys | No `//go:embed` directives exist anywhere in the tree; all content is generated at runtime. |
