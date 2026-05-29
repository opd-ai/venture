# Implementation Gaps — 2026-05-29

This document records gaps between what Venture *claims* (README, `pkg/version/version.go`
GoDoc, and `pkg/saveload/doc.go`) and what the code *actually does*. Each gap is reproducible
from the committed tree.

## Gap 1 — Item generation is no longer reproducible against the declared v1.0.0 baseline

- **Stated Goal:** The README opens with "deterministic, seed-based procedural generation to
  create all game content at runtime," and `pkg/version/version.go` (v1.0.0) documents that
  "Save file format is backward-compatible within a MAJOR version" and "Network protocol is
  stable within a MAJOR version." The audit harness encodes this as Phase 62.1 requirement #5:
  "v1.0.0 output matches baseline for same seed."
- **Current State:** `TestDeterminism_VersionStability/Item` **fails** on the committed code.
  For `BaselineSeed = 99999`, the item generator now produces an output whose SHA256 is
  `afba81832bf6ffe8…`, while the recorded v1.0.0 baseline
  (`pkg/procgen/audit/baseline.go:25`) is `2b36ce659bf7c7b6…`. The new hash is **stable across
  repeated runs** (verified 3× under `xvfb`), so generation is still internally deterministic
  — its output has simply *drifted* from the baseline. The other 13 audited generators still
  match their baselines; only `Item` regressed. `BaselineVersion` and the binary version both
  remain "1.0.0".
- **Impact:** Two builds both labelled `1.0.0` generate different items for the same seed,
  violating the headline determinism promise and the documented intra-MAJOR save/protocol
  stability guarantee. This can desynchronize multiplayer snapshots (items differ between
  peers) and break save reproducibility. Because the test is red, CI can no longer tell this
  regression apart from any *future* determinism break.
- **Closing the Gap:** In `pkg/procgen/item`, determine whether the output change was
  intentional. If unintentional, locate and revert the generation-logic change that moved the
  hash off `2b36ce659bf7c7b6`. If intentional, bump the version in
  `pkg/version/version.go` per the project's own semver rules (a save/protocol-affecting change
  is at least MINOR), then regenerate and commit the new baseline (`baseline.go:25`,
  `baseline_hashes.json`, and `BaselineVersion`). Re-run
  `xvfb-run -a go test ./pkg/procgen/audit/ -run TestDeterminism_VersionStability -count=1`
  until green.

## Gap 2 — Federation handshake cleanup scales with attacker traffic, not with time

- **Stated Goal:** The README markets multiplayer/federation that is robust over high-latency
  Tor links (200–5000 ms) — an environment where reconnects and handshake retries are frequent.
- **Current State:** `ProcessHandshake` launches a new goroutine — `go hm.cleanupNonces()` —
  on **every** inbound handshake (`pkg/network/federation/handshake.go:294`), and each
  goroutine acquires the manager's exclusive lock. The cleanup cadence is therefore driven by
  remote peer traffic rather than by a timer.
- **Impact:** A peer can force one goroutine creation + one exclusive-lock acquisition per
  handshake, producing scheduler/GC churn and lock contention on the core networking path — a
  mild DoS-amplification vector precisely on the high-churn Tor path the project targets.
  (No data race: `go test -race` is clean; the map is correctly guarded.)
- **Closing the Gap:** Replace the per-call goroutine with a single periodic sweeper owned by
  `HandshakeManager` (a `time.Ticker` started in `NewHandshakeManager`, stopped via `Close()`),
  or gate cleanup behind a last-swept timestamp so it runs at most once per interval. Validate
  with `go test -race ./pkg/network/federation/...` plus a bounded-goroutine benchmark.

## Gap 3 — WASM save "sha256" integrity files actually hold a non-cryptographic FNV-1a digest

- **Stated Goal:** `pkg/saveload/doc.go` promises "integrity validation" for saves on both
  desktop (SHA256) and WASM (FNV-1a).
- **Current State:** The documentation is accurate, but the **implementation naming** is
  misleading: on WASM the digest is computed with FNV-1a (`storage_wasm.go:782`,
  `computeChecksum`) yet stored under keys suffixed `".sha256"` (`:474`,`:493`,`:742`,`:774`).
  A comment at `:781` further claims `crypto/sha256` "may have WASM compatibility issues,"
  which is outdated (`crypto/sha256` works under `GOOS=js`).
- **Impact:** FNV-1a catches accidental corruption but offers no tamper resistance, and the
  `.sha256` suffix invites a maintainer (or cross-platform recovery code) to assume a
  cryptographic checksum is present. In single-player WASM the save is not a security boundary,
  so the practical risk is maintainability/correctness rather than exploit.
- **Closing the Gap:** Either rename the WASM key suffix to reflect the algorithm
  (e.g. `.fnv1a` / `.checksum`, updating the matching reads) or switch WASM to `crypto/sha256`
  for parity with desktop (`recovery.go:27`) and delete the outdated comment. Validate with
  `GOOS=js GOARCH=wasm go build ./pkg/saveload/...` and the existing checksum tests.

## Gap 4 — A 16 MB local analysis artifact is committed and untracked by `.gitignore`

- **Stated Goal:** The README presents Venture as "a single distributable binary per platform
  with no external asset files" — i.e. a lean, asset-free repository.
- **Current State:** `post.json` (16,526,322 bytes) sits in the repository root. It is a
  go-stats-generator analysis dump (its `metadata.repository` field embeds a developer's local
  `/home/user/go/src/github.com/opd-ai/venture` path), is referenced by zero Go source files,
  and is absent from `.gitignore`.
- **Impact:** Every clone carries ~16 MB of dead weight and a leaked local path string. No
  functional/runtime impact.
- **Closing the Gap:** `git rm post.json` and add it (or the tool's output pattern) to
  `.gitignore`. No build/test validation needed (non-code change).
