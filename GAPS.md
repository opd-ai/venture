# Implementation Gaps — 2026-05-30

Detailed gap analysis and implementation roadmap, derived from `AUDIT.md`. Every gap below
was verified against the committed tree (`go build`/`go vet` clean; go-stats-generator
metrics). Gaps are ordered by the audit tiebreaker: partially-wired features → dead code on
maintained paths → interface/contract gaps. There are **no stubs on critical paths** and
**no dependency gaps**.

---

## Gap 1 — Legendary quests do not verify raid completion against the raid system

- **Intended Behavior**: `pkg/procgen/legendary/doc.go:26-31` states legendary quests
  "integrate with V9 Phase 59.1 raid system" — raid-gated quest phases should be validated
  against actual raid clears. `QuestManager` is constructed with a `*raids.Manager`
  (`manager.go:55-66`).
- **Current State**: The `raidManager` field (`manager.go:22`) is stored at `:66` and
  **never read** anywhere in the package. `ValidateRaidCompletion` (`manager.go:208`)
  validates only against in-memory quest phase data and trusts the caller-supplied
  `raidID`/`tier`. The dependency is plumbed and discarded.
- **Blocked Goal**: Raid System + Legendary content — verified raid-gated progression.
- **Implementation Path**: In `ValidateRaidCompletion`, query `qm.raidManager` to confirm
  the `raidID` corresponds to a completed instance/lockout for the player (matching the
  required `tier`) before calling `tracker.AddRaidCompleted`. Return an error when the raid
  manager has no record of the clear. Add a unit test that rejects an unverified completion
  and accepts a recorded one.
- **Dependencies**: Requires Gap 2 (an actual caller) to exercise the path end-to-end.
- **Effort**: small.

---

## Gap 2 — `ValidateRaidCompletion` is exported but never called in production

- **Intended Behavior**: When a player clears a raid encounter that a legendary quest
  requires, the quest's raid phase should advance.
- **Current State**: `pkg/procgen/legendary/manager.go:208` `ValidateRaidCompletion` has
  **no production callers** (verified by call-graph search; only its own definition matches).
  No code path records raid completion for legendary quests.
- **Blocked Goal**: Completing `PhaseRaid` phases of legendary quests.
- **Implementation Path**: Invoke `ValidateRaidCompletion` from
  `pkg/engine/legendary_quest_system.go` when a raid encounter completes (the system already
  holds a real `*raids.Manager` via `engine.NewLegendaryQuestSystem`, wired at
  `cmd/client/handlers.go:1123` and `cmd/server/v4_systems.go:191`). Add an integration test
  driving a raid-phase quest to completion.
- **Dependencies**: Pairs with Gap 1; close Gap 1's verification logic first.
- **Effort**: small.

---

## Gap 3 — Orphan client legendary manager initialized with a nil dependency

- **Intended Behavior**: A single, correctly-wired legendary quest manager per client.
- **Current State**: `cmd/client/init_versions.go:532` creates
  `sys.legendaryManager = legendary.NewQuestManager(nil)` ("nil raid manager for now"). The
  field (`cmd/client/handlers.go:584`) is **never read or registered** afterward. The real
  legendary feature runs through `sys.legendaryQuestSystem`
  (`engine.NewLegendaryQuestSystem(..., raidManager)`, `handlers.go:1123`).
- **Blocked Goal**: None functionally; this is redundant dead state that passes a `nil`
  dependency and invites confusion/maintenance burden.
- **Implementation Path**: Remove the `sys.legendaryManager` field and its initialization,
  or replace its uses (if any are later added) with `LegendaryQuestSystem`'s embedded
  manager.
- **Dependencies**: None.
- **Effort**: small.

---

## Gap 4 — Federated marketplace hop count is a constant, not real federation distance

- **Intended Behavior**: Cross-server "shared marketplaces" should price fees and delivery
  times by actual network distance between servers (README: federation enables "shared
  marketplaces").
- **Current State**: `pkg/world/economy/marketplace.go:374-380` `UpdateRemoteCache`
  hardcodes `EstimatedHops = 1` (same server) or `2` (remote) with the comment "For now,
  assume 1 hop per remote server." `EstimatedHops` drives real logic: transaction fees
  (`marketplace.go:325` and `types.go:176` via `CalculateTransactionFee`) and delivery time
  (`types.go:162`: `hopTime := l.EstimatedHops * 300`).
- **Blocked Goal**: Accurate federated-marketplace economics across multi-hop topologies.
- **Implementation Path**: Thread the real hop/route distance from the federation
  discovery/handshake layer (which already tracks peers and routes) into `UpdateRemoteCache`,
  and set `EstimatedHops` from it. Add a test asserting fee and delivery time scale with hop
  count rather than the 1/2 constant.
- **Dependencies**: Needs a federation API exposing per-server hop/route distance.
- **Effort**: medium.

---

## Gap 5 — Political-warfare reputation penalties never reach the faction system

- **Intended Behavior**: Aggressive guild actions (war declarations, embargoes) should lower
  the acting guild's standing with affected factions (`pkg/integration/political_warfare`
  models "guild wars, treaties, embargoes, and diplomacy").
- **Current State**: `manager.go:418-432` `applyReputationPenaltyInternal` only appends a
  `ReputationPenalty` record ("In a full implementation, this would interact with the faction
  system / For now, we just record the penalty"; `FactionID` hardcoded `"all"`). It is called
  on real war declarations (`:119`) and embargoes (`:223`). The recorded `m.penalties` slice
  is only persisted/returned (`:708`, `:721`) and **never consumed** to alter faction
  reputation.
- **Blocked Goal**: "Political warfare" diplomacy having gameplay consequences on faction
  standing.
- **Implementation Path**: Add a faction-reputation interface (dependency-injected into
  `Manager`) and, in `applyReputationPenaltyInternal`, resolve the affected faction(s) and
  apply the penalty through it, replacing the `"all"` placeholder with the resolved ID. Add a
  test asserting a war declaration lowers the attacker's faction reputation.
- **Dependencies**: Requires access to the reputation/faction system
  (`pkg/engine` reputation systems) via an interface to avoid import cycles.
- **Effort**: medium.

---

## Gap 6 — Economy ECS `world` dependency declared but never consumed

- **Intended Behavior**: `economy.System` documents that it "handles … entity-based
  transactions," implying it reads/writes player inventory entities through the ECS `world`.
  The `World` interface is defined for this purpose (`pkg/world/economy/interfaces.go:9`).
- **Current State**: `system.go:13` stores a `world World` field that is **never read** in
  the package; `Update` performs only marketplace cleanup. The client wires it `nil`
  (`cmd/client/init_versions.go:537`: "nil world for now, entity interface").
- **Blocked Goal**: Entity/inventory-aware marketplace transactions (debit/credit player
  items on listing/purchase).
- **Implementation Path**: Either (a) implement the entity-based transaction path that reads
  the ECS `world` (e.g. validate/transfer inventory entities on `CreateListing`/purchase), or
  (b) if entity-based transactions are out of scope, remove the unused `World` interface and
  field and correct the `System` doc comment. Add a test for whichever path is chosen.
- **Dependencies**: If implemented, depends on the ECS inventory component API.
- **Effort**: medium (implement) / small (remove + doc fix).

---

## Gap 7 — WebRTC native simulation layer is unused on production paths; docs contradict themselves

- **Intended Behavior**: `pkg/network/federation/webrtc/doc.go` describes a STUN client and
  NAT-traversal coordinator. README markets federation robust over high-latency/Tor links.
- **Current State**: The native simulation never runs in production:
  `stun.go:132-178` `querySTUNServer` dials UDP then returns a hardcoded `203.0.113.42` /
  `localPort+10000` instead of an RFC 5389 binding; `nat_traversal.go:178-187`
  `tryDirectConnection` always fails and `:213-238` `tryTURNConnection` never creates a TURN
  allocation. On native, `wireWebRTCFederation` is a no-op and federation uses TCP
  (`cmd/client/webrtc_native.go:10-14`); on WASM the real pion path (`peer_wasm.go`)
  configures STUN through pion and **bypasses** `STUNClient`/`NATTraversal`. So the code is
  exercised only by tests. `doc.go:170-180` lists `STUNClient`/NAT traversal as both
  "Production-ready (fully functional)" **and** "Stub behavior … Returns simulated IP/port."
- **Blocked Goal**: None currently — native federation works over TCP and WASM uses real
  pion. The issue is a documentation contradiction plus parallel code that is never on a
  production path (maintenance/clarity burden).
- **Implementation Path**: Reconcile `doc.go` so simulated components appear only under "Stub
  behavior." If real native WebRTC is intended, implement `querySTUNServer` per RFC 5389 and
  route `peer_native.go` through `STUNClient`/`NATTraversal`; otherwise mark the simulation
  clearly as test-only scaffolding.
- **Dependencies**: None for the doc fix; a real native implementation would depend on a
  WebRTC/ICE stack on native builds.
- **Effort**: small (doc reconciliation) / large (real native WebRTC).

---

## Gap 8 — `FindClosestMerchant` returns a placeholder distance

- **Intended Behavior**: Return the actual distance from the player to the closest merchant.
- **Current State**: `pkg/engine/merchant_spawn.go:341-348` computes the returned distance
  with a crude `for i := 0.0; i*i < minDistSq; i += 0.1` loop ("Placeholder - in real impl
  would use math.Sqrt(minDistSq) … Avoiding math import to keep file simple"). The value is
  consumed only by a debug log (`cmd/client/handlers.go:3316`); merchant **selection** uses
  `minDistSq` correctly, so gameplay is unaffected.
- **Blocked Goal**: None — cosmetic accuracy of a logged value only.
- **Implementation Path**: Import `math` and return `math.Sqrt(minDistSq)`; delete the loop.
  Behavior-neutral for selection.
- **Dependencies**: None.
- **Effort**: small.

---

## Closed / Non-Gaps (for reference)

The following were investigated and confirmed **not** to be implementation gaps:

- **Zero TODO/FIXME markers**; go-stats-generator's `BUG/HACK/XXX` annotation counts are
  false positives from game-content strings (e.g. cyberpunk "Hack" names).
- **`Gap:`/`Fix:` comments** in `cmd/client/init_versions.go` and `cmd/server/v4_systems.go`
  are historical **resolved** "INTEGRATION FIX" records, each followed inline by the system's
  initialization.
- **Voice chat is wired** in production (`audio.Manager.InitializeVoice` `nil` transport is a
  test affordance; real path documented at `pkg/audio/manager.go:305-314`).
- **`doc.go`-only umbrella packages** (`pkg/benchmark`, `pkg/engine/physics`,
  `pkg/integration`) are intentional namespace docs, not empty scaffolding.
- **50 `Deprecated:` methods** are tracked deprecations with documented replacements.

## Validation Summary

After closing any subset of the above, validate with:

```bash
go build ./...
go vet ./...
go test ./pkg/procgen/legendary/... ./pkg/world/economy/... \
        ./pkg/integration/political_warfare/... ./pkg/network/federation/webrtc/...
# Full suite (no display): xvfb-run -a go test ./...
```
