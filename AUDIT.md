# IMPLEMENTATION GAP AUDIT — 2026-04-25

> **Audit revision**: Rev 3. Supersedes the previous AUDIT.md (ROADMAP.md form,
> 2026-04-24). All prior findings and road-map priorities from that file are
> preserved in the **Legacy Gap ID Compatibility** section at the bottom.

---

## Project Architecture Overview

Venture is a fully procedural multiplayer action-RPG built on Go 1.24.5 and
Ebiten v2.9.3. **All graphics, audio, and gameplay content are generated at
runtime from a single binary; no external asset files exist.** Key architectural
pillars:

- **ECS**: `pkg/engine/` — ~400 files, 100+ systems, 183 `world.AddSystem` calls
  in `pkg/engine/system_init.go`.
- **Procedural generation**: `pkg/procgen/` — 25+ generators (terrain, entities,
  items, quests, spells, dialog, narrative).
- **Runtime rendering**: `pkg/rendering/` — sprites, tiles, lighting, particles,
  post-processing.
- **Procedural audio**: `pkg/audio/` — synthesis engine, adaptive music, SFX.
- **Multiplayer**: `pkg/network/` — TCP-based with prediction, lag compensation,
  federation, voice chat.
- **Platforms**: Linux, macOS, Windows, WebAssembly, iOS, Android.

---

## Build Status

`go build ./...` was attempted but failed due to missing X11 / OpenGL headers
(`X11/Xlib.h`) in the CI runner — a runner environment issue, not a code defect.
The failure is confined to Ebiten's GLFW C bindings and does not reflect any
Go-level error. `go vet` could not be run for the same reason. All static
analysis in this audit is based on source-level inspection.

---

## Gap Summary

| Category          | Count | Critical | High | Medium | Low |
|-------------------|-------|----------|------|--------|-----|
| Stubs / Simulated |   2   |    0     |  1   |   1    |  0  |
| Dead Code         |   1   |    0     |  0   |   0    |  1  |
| Partially Wired   |   1   |    0     |  0   |   0    |  1  |
| Interface Gaps    |   0   |    0     |  0   |   0    |  0  |
| Dependency Gaps   |   0   |    0     |  0   |   0    |  0  |
| **Total**         | **4** |  **0**   |**1** | **1**  |**2**|

---

## Implementation Completeness by Package

| Package | Assessment | Notes |
|---------|-----------|-------|
| `pkg/engine` | ✅ Complete | 183 systems registered; all major constructors called |
| `pkg/procgen` | ✅ Complete | All generators dispatch correctly by genre |
| `pkg/rendering` | ✅ Complete | Headless stubs are intentional (build-tag guarded) |
| `pkg/audio` | ✅ Complete | 91–98% coverage per prior benchmarks |
| `pkg/network` | ⚠️ Partial | TCP federation complete; WebRTC transport is simulated (G17) |
| `pkg/network/federation/webrtc` | 🔴 Stub | `Peer.Connect()` calls `simulateConnection()` (G17) |
| `pkg/world` | ✅ Complete | Housing, economy, raids, territory all wired |
| `pkg/mobile` | ✅ Complete | `MobileInputAdapter` wired in `cmd/mobile/mobile.go:308` |
| `pkg/modding` | ✅ Complete | FileSystemModRepository + HotReloadSystem both wired (G3, G15, G16 resolved) |
| `pkg/vr` | ✅ Complete (experimental) | OpenXR + WebXR adapters present; deliberately labeled experimental |
| `pkg/memprofile` | ✅ Complete | `fmt.Printf` exemption formally documented at `profile.go:195` |
| `pkg/engine/class_progression_system.go` | ⚠️ Partial | `Update()` is a no-op; progression via explicit API only (G18) |
| `pkg/engine/companion_system.go` | ⚠️ Stub | `executeScout()` uses hardcoded velocity instead of pathfinding (G19) |
| `pkg/engine/behavior_tree_advanced_nodes.go` | ⚠️ Stub | Ambush node uses random offset, not cover-based pathfinding (G20) |
| `cmd/client` | ✅ Complete | All v1–v8 systems wired via init_versions.go + handlers.go |
| `cmd/server` | ✅ Complete | All systems wired via v4–v8 system files |
| `cmd/mobile` | ✅ Complete | MobileInputAdapter, engine systems, lifecycle all present |

---

## Findings

### HIGH

- [ ] **[G17] WebRTC Browser-to-Browser Federation is Simulated, Not Real**
  — `pkg/network/federation/webrtc/peer.go:4,77` —
  The package header states "This is a stub implementation for testing; real
  WebRTC integration requires `github.com/pion/webrtc/v3`." The `Connect()`
  method calls `simulateConnection()` (line 77), which queues fake
  `StateConnected` transitions without any real ICE/DTLS negotiation. The
  signaling server is similarly simulated (`signaling.go:76,299`). The
  WASM-specific initializer `initWebRTCFederation()` at
  `cmd/client/webrtc_wasm.go:20` is defined but never called from any entry
  point. `NewWebRTCTransport()` (`transport_webrtc.go:31`) is never instantiated
  outside of `_test.go` files.

  **Impact**: Browser-to-browser federation (WASM client ↔ WASM server without a
  dedicated TCP host) is advertised — README line 60 lists "federation/WebRTC,
  portals" — but non-functional. Desktop TCP-based federation is fully
  operational; this gap only blocks the WebRTC code path for WASM peers.

  **Remediation**:
  1. Add `github.com/pion/webrtc/v3` (WASM-safe, no CGo) to `go.mod`.
  2. In `pkg/network/federation/webrtc/peer.go`, replace `simulateConnection()`
     body with `pion.NewPeerConnection`, data-channel creation, and SDP
     offer/answer exchange via the existing signaling transport.
  3. In `cmd/client/webrtc_wasm.go`, call `initWebRTCFederation(clientID)` and
     wire the returned `*Peer` into `sys.federationProtocol` via
     `NewWebRTCTransport()`.

  **Validation**: WASM build; two browser tabs can exchange a "ping" via the
  WebRTC data channel without a dedicated TCP server.

---

### MEDIUM

- [ ] **[G19] `CompanionSystem.executeScout()` Uses Hardcoded Velocity**
  — `pkg/engine/companion_system.go:496–502` —
  The comment reads "This is a stub - full implementation would use pathfinding".
  The body unconditionally sets `VX = 80.0`, `VY = 80.0`. A companion in Scout
  mode moves diagonally north-east indefinitely; it never returns to the owner,
  never avoids walls, and never changes direction.

  **Impact**: Companions assigned the Scout role (set via
  `CompanionComponent.BehaviorMode = BehaviorScout`) exhibit broken movement.
  Any player who relies on scout companions for exploration or detection will see
  nonsensical behavior.

  **Remediation**: Either (a) implement expanding-circle motion using the entity's
  `SpatialPartition` to select walkable cells at increasing radii, or (b) as a
  minimal fix, add angular variation using a per-companion counter so scouts move
  in four cardinal directions sequentially rather than one fixed diagonal.

  **Validation**: Spawn a companion with `BehaviorScout` mode; confirm it moves
  away from owner, turns at obstacles, and returns after a configurable radius.

---

### LOW

- [ ] **[G18] `ClassProgressionSystem.Update()` is a No-op**
  — `pkg/engine/class_progression_system.go:18–20` —
  The `Update` method body is a single comment: "Currently a stub - progression
  happens through LevelUp() calls / This system could be extended to apply
  passive effects." The system is registered (`cmd/client/handlers.go:2170`,
  `cmd/server/v4_systems.go:99`) and called every frame, consuming a scheduler
  slot without producing any output.

  **Impact**: Any time-based passive class effects (mana regen rate modifiers,
  class-specific buff cooldowns, etc.) cannot be applied automatically. Game
  designers must wire every class effect through explicit `LevelUp()` callsites.
  The core leveling path is unaffected.

  **Remediation**: Implement at minimum an empty-but-documented contract: if
  passive effects are genuinely out of scope, add a `_ = deltaTime` comment and
  a `// COMPLEXITY JUSTIFICATION: passive effects are event-driven` note to
  satisfy the quality gate. If passive effects are desired, iterate entities with
  `class_progression` component and apply class-specific regen ticks.

  **Validation**: `go test ./pkg/engine/ -run TestClassProgressionSystem`; all
  cases pass; no regression in leveling tests.

---

- [ ] **[G20] BehaviorTree Ambush Node Uses Random Positional Offset**
  — `pkg/engine/behavior_tree_advanced_nodes.go:391–396` —
  The comment states "In a full implementation, this would use pathfinding data."
  The ambush position is computed as the enemy's current position plus a ±50-unit
  random X/Y offset (`rng.Float64() - 0.5) * 100`). This does not account for
  walls, cover, or line-of-sight; enemies may "ambush" from inside solid terrain.

  **Impact**: Enemy AI quality for stealth/ambush archetypes is reduced. No
  feature is hard-blocked; the node evaluates to `NodeRunning` and the enemy
  eventually moves somewhere nearby.

  **Remediation**: Query `SpatialPartition` for passable tiles near the current
  position; prefer cells with low `VisibilityComponent.Visibility` (if populated
  by the LightingSystem) as cover candidates. Falls back to the current random
  offset if no cover tiles are found.

  **Validation**: `go test ./pkg/engine/ -run TestBehaviorTreeAmbush`; enemy
  ambush positions resolve to walkable cells.

| Candidate | Reason Rejected |
|---|---|
| `NewMySystem` (dangling) | In a doc comment at `pkg/engine/statmod.go:17` — not a real constructor |
| `NewStaminaRegenSystem` (dangling) | In a doc comment at `pkg/engine/doc.go:213` — documentation example only |
| `pkg/memprofile/profile.go` using `fmt.Printf` | Formal exemption documented at line 195–197 with guideline reference |
| `tradeRouteManager` not added via `world.AddSystem` in client | `RouteManager` is a goroutine-based manager (`Start()`/`Stop()`), not an ECS System; goroutine-based pattern is intentional |
| `ClassProgressionSystem` no callers for `Update` side-effects | Documented design: all progression goes through explicit `LevelUp()` calls |
| VR adapters use stub headset/controller on non-`vr`-tagged builds | Intentional via build tags; `vr_stub_adapters.go` is production-correct fallback |
| GPU bloom / GPU post-process headless stubs | Build-tag guarded (`headless`); correct test-infrastructure pattern |
| WebRTC `signaling.go:76` "simulate" comment | Part of the same simulated-peer stack as G17 — same root gap |
| `initWebRTCFederation` not called in main.go (non-WASM) | WASM-only file (`//go:build js && wasm`); correctly excluded from desktop builds |

---

## Legacy Gap ID Compatibility

Prior audit IDs from the 2026-04-24 `GAPS.md` (G1–G16) are mapped below. Code
comments, CI scripts, or issue-tracker entries referencing these IDs remain valid
cross-references.

| Prior ID | Title | Status in this audit |
|----------|-------|---------------------|
| G1 | OpenXR Controller / Headset Input Stubbed | ✅ RESOLVED (per rev-2 GAPS.md) |
| G2 | Eleven Engine Systems Never Registered | ✅ RESOLVED (per rev-2 GAPS.md) |
| G3 | Mod Browser Unreachable from Any Binary | ✅ RESOLVED (per rev-2 GAPS.md) |
| G4 | Seasonal Event Subsystem Has No Spawner | ✅ RESOLVED (per rev-2 GAPS.md) |
| G5 | Modding System Wired Server-Only | ✅ RESOLVED (per rev-2 GAPS.md) |
| G6 | `vr_webxr_adapters.go` Documented but Missing | ✅ RESOLVED (per rev-2 GAPS.md) |
| G7 | Client Has No Observability/Health Endpoint | ✅ RESOLVED (per rev-2 GAPS.md) |
| G8 | Dead `Server` Type in `pkg/hostplay` | ✅ RESOLVED (per rev-2 GAPS.md) |
| G9 | `EnableShadows` Deprecation Not Enforced | ✅ RESOLVED (per rev-2 GAPS.md) |
| G10 | `ExtendedAchievementSystem` Shadows Wired System | ✅ RESOLVED (per rev-2 GAPS.md) |
| G11 | Menu "Exit Game" Returns Error | ✅ RESOLVED (per rev-2 GAPS.md) |
| G12 | Mobile Portrait Picker Returns Error | ✅ RESOLVED (per rev-2 GAPS.md) |
| G13 | `pkg/companion` Namespace Undocumented | ✅ RESOLVED (per rev-2 GAPS.md) |
| G14-1 | `FederatedMarket.Stop` lacks `sync.Once` | ✅ RESOLVED (per rev-2 GAPS.md) |
| G14-2 | `FederatedMarket.Start` lacks `sync.Once` | ✅ RESOLVED (per rev-2 GAPS.md) |
| G14-3 | `TCPServer.Start` defer-unlock fragility | ✅ RESOLVED (per rev-2 GAPS.md) |
| G14-4 | Non-defer mutex unlock patterns | ✅ RESOLVED (per rev-2 GAPS.md) |
| G14-5 | `startLegacyMetricsMonitor` goroutine leaks | ✅ RESOLVED (per rev-2 GAPS.md) |
| G14-6 | `CleanupTask` stop channel undocumented | ✅ RESOLVED (per rev-2 GAPS.md) |
| G15 | `HotReloadSystem` never registered | ✅ RESOLVED (per rev-2 GAPS.md) |
| G16 | `FileSystemModRepository` unused in production | ✅ RESOLVED (per rev-2 GAPS.md) |
| **Prior Gap 4** | `pkg/memprofile` uses `fmt.Printf` (custom instructions) | ✅ RESOLVED — Exemption formally documented at `pkg/memprofile/profile.go:195–197` |

---

## Appendix: Verified Integration Chain Spot-Checks

| System | Constructor | Registered | Update Fires | Output Consumed |
|--------|------------|-----------|-------------|----------------|
| AchievementSystem | `init_versions.go:140` | `handlers.go:2198` | ✅ | EventBus handlers |
| ExtendedAchievementSystem | `system_init.go:2112` | `system_init.go:2114` | ✅ | Achievement components |
| CompanionAISystem | `init_versions.go:61` | `handlers.go` | ✅ | Companion entity velocity |
| ClassProgressionSystem | `init_versions.go:79` | `handlers.go:2170` | ✅ (no-op body) | via LevelUp() |
| CarryOverSystem | `init_versions.go:259` | `handlers.go:2214` | ✅ | CarryoverComponent |
| ChallengeSystem | `system_init.go:2041` | `system_init.go` | ✅ | DailyChallengeComponent |
| GuildSystem | `init_versions.go:607` | registered | ✅ | GuildComponent |
| GuildCombatBonusSystem | `system_init.go:458` | `system_init.go` | ✅ | GuildCombatBonusComponent |
| HotReloadSystem | `init_versions.go:752` | `init_versions.go:~810` | ✅ | HotReloadComponent |
| CommerceSystem | `system_init.go:1094` | `system_init.go:1095` | ✅ | Commerce events |
| TradeRouteManager (client) | `init_versions.go:635` | `Start()` goroutine | ✅ | Route updates via goroutine |
| TradeRouteManager (server) | `v4_systems.go:~280` | `AddSystem` wrapper | ✅ | ECS update |
| MobileInputAdapter | `mobile.go:308` | entity attachment | ✅ | InputSystem reads |
| EbitenHUDSystem | `game.go:188` | `game.go:363` | ✅ | `game.go:1616` Draw |
| WebRTCTransport | — **never instantiated** — | 🔴 N/A | 🔴 | 🔴 |
