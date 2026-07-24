# IMPLEMENTATION GAP AUDIT — 2026-05-30

> Scope: implementation-completeness audit of the `github.com/opd-ai/venture` Go module.
> Report only — no source code was modified. Findings are evaluated against the
> project's **own stated goals and architecture**, not aspirational completeness.

## Project Architecture Overview

Venture is a fully procedural multiplayer action-RPG built with Go and Ebiten. All
graphics, audio, and gameplay content are generated at runtime from a single binary with
no external asset files. Stated architecture (README.md, ROADMAP.md):

- `cmd/client/` — desktop game client; auto-starts a localhost server for solo play on native builds.
- `cmd/server/` — dedicated authoritative multiplayer server.
- `cmd/mobile/` — iOS/Android entry point.
- `pkg/engine/` — ECS core and the large majority of game systems (~5,500 exported funcs).
- `pkg/procgen/` — 25+ seed-based deterministic content generators.
- `pkg/rendering/` — runtime graphics pipeline (sprites, tiles, lighting, particles, post-processing).
- `pkg/audio/` — procedural music/SFX synthesis and ADPCM voice codec.
- `pkg/network/` — multiplayer networking, federation, chat, trade, resilience.
- `pkg/world/` — persistent world: housing, economy, territory, raids.
- `pkg/integration/` — cross-system feature integrations (guild wars, narrative-world, etc.).

Stated headline goals: 100% procedural content, single-binary distribution, deterministic
seed-based generation, ECS architecture, high-latency multiplayer (200–5000 ms / Tor),
cross-server federation (shared marketplaces, multi-server guilds), procedural audio +
voice chat, modding, mobile, and WebAssembly.

### Baseline health (Phase 2 evidence)

| Check | Result |
|-------|--------|
| `go build ./...` | **clean** (after installing X11 dev headers for cgo/Ebiten) |
| `go vet ./...` | **clean** (no diagnostics) |
| go-stats-generator overall doc coverage | **90.0%** (functions 96.6%, packages 99.1%) |
| TODO / FIXME code annotations | **0** |
| `Deprecated:` godoc markers | 50 (all with documented replacements) |
| LOC (non-test) / files / packages | 217,275 / 1,299 / 109 |

The codebase is mature, builds and vets cleanly, and carries **no** `TODO`/`FIXME`
markers. go-stats-generator's `BUG=98 / HACK=2 / XXX=5` annotation counts were verified to
be **false positives** sourced from game-content strings (e.g. cyberpunk "Hack" item and
skill names) rather than code annotations. Consequently there are **no CRITICAL or HIGH**
implementation gaps: the project fulfills its core stated purpose. The remaining gaps are
non-critical **partially-wired integrations** and a **test-only simulation layer**.

## Gap Summary

| Category | Count | Critical | High | Medium | Low |
|----------|-------|----------|------|--------|-----|
| Stubs/TODOs | 0 | 0 | 0 | 0 | 0 |
| Dead Code | 2 | 0 | 0 | 2 | 0 |
| Partially Wired | 4 | 0 | 0 | 3 | 1 |
| Interface/Contract Gaps | 1 | 0 | 0 | 1 | 0 |
| Dependency Gaps | 0 | 0 | 0 | 0 | 0 |
| Tracked/Minor | 1 | 0 | 0 | 0 | 1 |
| **Total** | **8** | **0** | **0** | **6** | **2** |

(Several findings span two categories; each is counted once under its primary category.)

## Implementation Completeness by Package

Counts are exported functions/methods per package from go-stats-generator. "Stub/Dead"
columns list verified non-functional or unconsumed elements found by this audit. Test
coverage percentages are documented per package in `ROADMAP.md` (all audited packages
exceed the project's ≥40% floor); the JSON run used `--skip-tests`, so per-package test %
is not reproduced here.

| Package | Exported Fns | Verified Stub/Dead Elements | Notes |
|---------|-------------:|-----------------------------|-------|
| `pkg/network/federation/webrtc` | 72 | `STUNClient.querySTUNServer`, `NATTraversal.tryDirectConnection`/`tryTURNConnection`, native `Peer.Connect/Send`, native `SignalingClient` transport | Documented test-only simulation; real path is pion (WASM only) |
| `pkg/procgen/legendary` | 42 | `QuestManager.raidManager` field (never read); `ValidateRaidCompletion` (no production callers) | Raid integration documented but not consumed |
| `pkg/world/economy` | 58 | `System.world` field (never read) | Entity-based economy integration not wired |
| `pkg/integration/political_warfare` | 30 | faction effect of `applyReputationPenaltyInternal` | Penalties recorded/persisted but never applied to factions |
| `pkg/engine` | ~5,541 | `FindClosestMerchant` returned distance (placeholder) | Distance value used only for a debug log |

## Findings

### CRITICAL
*(none — the project builds, vets clean, has no TODO/FIXME markers, and fulfills its core stated purpose.)*

### HIGH
*(none — no partially-implemented feature was found to produce incorrect behavior on a critical execution path.)*

### MEDIUM

- [ ] **Legendary raid integration is plumbed but never consumed** — `pkg/procgen/legendary/manager.go:22` (`raidManager` field), set at `:66`, never read anywhere in the package (`ValidateRaidCompletion` at `:208` validates only against in-memory quest phase data and never queries the raid manager) — **What is missing:** the documented integration with the raid system (`pkg/procgen/legendary/doc.go:26-31`: "Quests integrate with V9 Phase 59.1 raid system") is not implemented; the `*raids.Manager` dependency is stored and discarded. **Which stated goal is blocked:** legendary quests requiring verified raid clears (a "Raid System" + "Legendary" feature). **Remediation:** In `legendary`, have `ValidateRaidCompletion` consult `qm.raidManager` (e.g. confirm the `raidID`/`tier` corresponds to a completed instance/lockout via the raids API) before calling `tracker.AddRaidCompleted`. Add a unit test that rejects an unverified raid completion. Validate with `go test ./pkg/procgen/legendary/...`.

- [ ] **`ValidateRaidCompletion` is exported dead code with no callers** — `pkg/procgen/legendary/manager.go:208` — **What is missing:** the function is never invoked from within the project (verified by call-graph search); legendary quests with `PhaseRaid` phases therefore have no production path that records raid completion. **Which stated goal is blocked:** completing raid-gated legendary quest phases. **Remediation:** Wire a call from `engine.LegendaryQuestSystem` (`pkg/engine/legendary_quest_system.go`) when a raid encounter completes, or document it as a server/external API. Add an integration test that drives a raid-phase quest to completion. Validate with `go test ./pkg/engine/... ./pkg/procgen/legendary/...`.

- [ ] **Standalone client legendary manager is a dead field** — `cmd/client/init_versions.go:532` assigns `sys.legendaryManager = legendary.NewQuestManager(nil)` ("nil raid manager for now"); the field (declared `cmd/client/handlers.go:584`) is never read or registered after assignment. Real legendary functionality runs through `sys.legendaryQuestSystem` (`engine.NewLegendaryQuestSystem(..., raidManager)`, `handlers.go:1123`) which already receives a real raid manager. **What is missing:** the orphan instance serves no purpose and passes a `nil` dependency. **Remediation:** Remove `sys.legendaryManager` and its initialization, or route it through `LegendaryQuestSystem`. Validate with `go build ./cmd/client/...`.

- [ ] **Federated marketplace hop count is hardcoded, not derived from federation topology** — `pkg/world/economy/marketplace.go:374-380` (`UpdateRemoteCache`: "For now, assume 1 hop per remote server"; sets `EstimatedHops = 1`, else `2`). `EstimatedHops` feeds real business logic: transaction fees (`marketplace.go:325`, `types.go:176` `CalculateTransactionFee`) and delivery time (`types.go:162` `hopTime := l.EstimatedHops * 300`). **Which stated goal is blocked:** correct cross-server "shared marketplace" pricing/delivery in federation. **Remediation:** Compute `EstimatedHops` from actual federation routing distance (the federation discovery/handshake layer already tracks peers) instead of the 1/2 constant, threading the real hop count into `UpdateRemoteCache`. Add a test asserting fee/delivery scales with hop count. Validate with `go test ./pkg/world/economy/...`.

- [ ] **Political-warfare reputation penalties are recorded but never applied to factions** — `pkg/integration/political_warfare/manager.go:418-432` (`applyReputationPenaltyInternal`: "In a full implementation, this would interact with the faction system / For now, we just record the penalty"; `FactionID` hardcoded `"all"`). Called on real war declarations (`:119`) and embargoes (`:223`); the recorded `m.penalties` slice is only persisted/returned (`:708,:721`), never consumed to alter faction reputation. **Which stated goal is blocked:** guild diplomacy/"political warfare" actually affecting faction standing. **Remediation:** Inject a faction-reputation interface into `Manager` and apply the penalty to the relevant faction(s) in `applyReputationPenaltyInternal`, replacing the `"all"` placeholder with the resolved faction ID. Add a test asserting a war declaration lowers the attacker's faction reputation. Validate with `go test ./pkg/integration/political_warfare/...`.

- [ ] **Economy system's ECS `world` dependency is never consumed (interface/contract gap)** — `pkg/world/economy/system.go:13` (`world World` field), interface defined at `pkg/world/economy/interfaces.go:9`; the field is never read in the package, and the client wires it `nil` (`cmd/client/init_versions.go:537`: "nil world for now, entity interface"). The `System` doc claims it "handles ... entity-based transactions," but `Update` only performs marketplace cleanup. **Which stated goal is blocked:** entity/inventory-aware marketplace transactions. **Remediation:** Either implement the entity-based transaction path that reads from `world` (e.g. debit/credit player inventory entities on listing/purchase) or remove the unused `World` interface and field and correct the doc comment. Validate with `go test ./pkg/world/economy/...` and `go vet ./...`.

### LOW

- [x] **WebRTC native simulation layer is unused on every production path + self-contradictory docs** — `pkg/network/federation/webrtc/stun.go:132-178` (`querySTUNServer` dials UDP then returns a hardcoded `203.0.113.42` / `localPort+10000` instead of an RFC 5389 binding), `nat_traversal.go:178-187` (`tryDirectConnection` always fails) and `:213-238` (`tryTURNConnection` never establishes a TURN allocation). On native builds `wireWebRTCFederation` is a no-op (`cmd/client/webrtc_native.go:10-14`, federation uses TCP); on WASM the real pion path (`peer_wasm.go`) configures STUN via pion directly and bypasses `STUNClient`/`NATTraversal`. Thus the simulation is exercised only by tests. `doc.go:170-180` simultaneously lists `STUNClient`/NAT traversal as **"Production-ready (fully functional)"** and as **"Stub behavior … Returns simulated IP/port"** — a direct contradiction. **Remediation:** Reconcile `doc.go` so the simulated components are listed only under stub behavior; if a real native STUN/NAT path is intended, implement `querySTUNServer` against RFC 5389 and route `peer_native.go` through it. This is LOW because native federation already works over TCP and the gap is documented. Validate with `go test ./pkg/network/federation/webrtc/...`. **RESOLVED 2026-07-23**: Updated `doc.go` to clearly separate production-ready coordination logic from stub implementations; added note that native builds use TCP federation and WASM uses real pion/webrtc.

- [x] **`FindClosestMerchant` returns a placeholder distance** — `pkg/engine/merchant_spawn.go:341-348` computes the returned distance with a crude `for i := 0.0; i*i < minDistSq; i += 0.1` loop ("Placeholder - in real impl would use math.Sqrt(minDistSq) … Avoiding math import to keep file simple"). The value is consumed only by a debug log (`cmd/client/handlers.go:3316`); merchant selection itself uses `minDistSq` correctly. **Remediation:** Return `math.Sqrt(minDistSq)` (import `math`) for an accurate distance; the change is behavior-neutral for selection and only improves the logged value. Validate with `go test ./pkg/engine/...`. **RESOLVED**: Code already uses `math.Sqrt(minDistSq)` with `math` imported (verified at `pkg/engine/merchant_spawn.go:8,344`).

## False Positives Considered and Rejected

| Candidate Finding | Reason Rejected |
|-------------------|----------------|
| go-stats-generator reports `BUG=98`, `HACK=2`, `XXX=5` annotations | Verified false positives: matches are game-content strings (e.g. cyberpunk "Hack" item/skill names in `pkg/procgen/item/templates.go:335`, `pkg/procgen/minigame/generator.go:372`), not code annotations. `rg -w` for `TODO/FIXME/HACK/XXX` in non-content code returns 0. |
| 20+ `Gap:` comments in `cmd/client/init_versions.go` and `cmd/server/v4_systems.go` | These are historical **"INTEGRATION FIX … Fix: Added …"** records; each is immediately followed by the system's initialization (e.g. `init_versions.go:152-205`). They document **resolved** gaps, not open work. |
| `"not implemented"` string in `pkg/audit/features/feature_completeness.go:76` | It is a validation **message** appended to a feature's issue list, not a stub body. |
| 0-function packages `pkg/benchmark`, `pkg/engine/physics`, `pkg/integration` | `doc.go`-only namespace/umbrella packages by design; real code lives in subpackages. Not scaffolding stubs. |
| 50 `Deprecated:` godoc methods (e.g. `pkg/procgen/quest/types.go:139`) | Each names a documented replacement and is retained for backward compatibility — a tracked deprecation, not an implementation gap. |
| `audio.Manager.InitializeVoice` accepts a `nil` transport (`pkg/audio/manager.go:305-314`) | The doc explicitly states `nil` is for tests and documents the real production wiring (`TCPServer.routeVoiceCommand`, `TCPVoiceTransport`, `cmd/client/handlers.go` `initializeVoiceTransport`). Voice **is** wired in production. |
| `worldEconomySystem` created with `nil` world (`cmd/client/init_versions.go:537`) | Recorded as a finding for the unused `world` field, but the `nil` itself is harmless (field never dereferenced) — not a crash/latent-nil bug. |
| Server-side voice channel membership filtering "future enhancement" (`pkg/network/server.go:801`) | Documented design: filtering is performed client-side by `TCPVoiceTransport.HandleReceivedPacket`. Intentional, with rationale — not an undiscovered gap. |

## Methodology Notes

- go-stats-generator v(latest) JSON metrics (`functions`, `documentation`, `packages`,
  `interfaces`, `structs`, `duplication`) were the primary quantitative evidence.
- `go build ./...` and `go vet ./...` were run after installing X11 dev headers
  (`libx11-dev` et al.) required by Ebiten's cgo glfw backend; both are clean.
- Every finding was checked against the Phase 3f false-positive rules: documented
  minimalism, external/public-API callers, intentional simplicity, and call-graph reachability.
