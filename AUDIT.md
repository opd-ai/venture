# Venture Codebase Audit Plan

**Version:** 2.0  
**Last Updated:** 2026-02-07  
**Purpose:** Step-by-step plan for conducting a package-by-package audit of the Venture codebase

---

## How to Use This Plan

This document is an actionable, step-by-step guide for auditing every package in the Venture codebase. Follow the steps in order. Each step tells you exactly what to do, what commands to run, and what to record. The plan is organized into three phases:

1. **Preparation** — Set up your environment and gather baseline data.
2. **Package-by-Package Audit** — Audit each package using a standard checklist, ordered from foundational packages to dependent ones.
3. **Cross-Cutting Verification** — Validate integration, performance, and documentation across the whole codebase.

Each package audit produces or updates an `AUDIT.md` file within the package directory. There are currently 110 such files across the codebase.

---

## Table of Contents

1. [Phase 1: Preparation](#phase-1-preparation)
2. [Phase 2: Package-by-Package Audit](#phase-2-package-by-package-audit)
3. [Phase 3: Cross-Cutting Verification](#phase-3-cross-cutting-verification)
4. [Appendix A: Per-Package Audit Checklist](#appendix-a-per-package-audit-checklist)
5. [Appendix B: Audit Command Reference](#appendix-b-audit-command-reference)
6. [Appendix C: System Catalog](#appendix-c-system-catalog)
7. [Appendix D: Access Requirement Matrix](#appendix-d-access-requirement-matrix)
8. [Appendix E: Performance Baselines](#appendix-e-performance-baselines)

---

## Phase 1: Preparation ✅ COMPLETED (2026-02-07)

**Status:** All preparation steps completed. Baseline metrics collected and documented in `/tmp/audit_phase1_report.md`.

Complete every step below before starting any package audit.

### Step 1.1: Set Up the Environment ✅

**Completed:** 2026-02-07

**Results:**
- ✅ Go version: go1.24.5 linux/amd64 (meets requirement)
- ✅ Client build: SUCCESS
- ✅ Server build: SUCCESS
- ✅ Environment: LOG_LEVEL=debug, LOG_FORMAT=json

1. Ensure Go 1.24.5+ is installed:
   ```bash
   go version
   ```
2. Clone and verify the repository builds:
   ```bash
   go build ./cmd/client && go build ./cmd/server
   ```
3. Set logging to debug for full audit visibility:
   ```bash
   export LOG_LEVEL=debug
   export LOG_FORMAT=json
   ```

### Step 1.2: Enumerate All Packages ✅

**Completed:** 2026-02-07

**Results:**
- Total packages: 113
- AUDIT.md files: 110/113 (97.3% coverage)
- Output saved to `/tmp/all_packages.txt` and `/tmp/audit_files.txt`

Run these commands and save the output for reference throughout the audit:

```bash
# List every Go package in the project
go list ./... > /tmp/all_packages.txt

# Count packages
wc -l /tmp/all_packages.txt

# List all existing AUDIT.md files
find . -name "AUDIT.md" -type f | sort > /tmp/audit_files.txt
```

### Step 1.3: Gather Baseline Metrics ✅

**Completed:** 2026-02-07

**Results:**
- Systems: 141 (as documented)
- Components: 39 types
- Interfaces: 32 core interfaces
- Test coverage: 82.4% average (from passing packages)
- Coverage baseline saved to `/tmp/coverage_baseline.txt`

**Note:** Some tests fail in headless environment due to X11/Ebiten requirements. Core logic packages pass.

Record current test coverage and build status before making any changes:

```bash
# Full test suite (record pass/fail and time)
go test ./... 2>&1 | tail -5

# Per-package coverage summary
go test -cover ./pkg/... 2>&1 | grep -E "^ok|^FAIL" > /tmp/coverage_baseline.txt

# Count systems, components, and interfaces
grep -r "type.*System struct" pkg/engine/*.go | grep -v test | wc -l
grep -r "func.*Type() string" pkg/engine/*_components.go | wc -l
grep -c "type.*interface" pkg/engine/interfaces.go
```

### Step 1.4: Map System Registrations ✅

**Completed:** 2026-02-07

**Results:**
- Client registrations: 113 `AddSystem` calls
- Server registrations: 61 across v4/v8/v9
- Note: 141 systems defined, 113 registered (client) - 28 may be server-only, conditional, or deprecated

Identify how systems are wired together at the application level:

```bash
# Client system registrations
grep "AddSystem" cmd/client/handlers.go | wc -l

# Server system registrations (each version)
grep "AddSystem\|RegisterSystem" cmd/server/v4_systems.go cmd/server/v8_systems.go cmd/server/v9_systems.go | wc -l

# UI keyboard bindings
grep "ebiten.Key" pkg/engine/input_system.go

# Network packet types
grep -r "type.*Packet" pkg/network/*.go
```

### Step 1.5: Review Entry Points ✅

**Completed:** 2026-02-07

**Status:** All entry point files reviewed and documented.

Read these files to understand how the application initializes:

| File | What to look for |
|------|------------------|
| `cmd/client/main.go` | Client startup, Ebiten game loop |
| `cmd/client/handlers.go` | System registration order (~100+ `AddSystem` calls) |
| `cmd/server/main.go` | Server startup, connection handling |
| `cmd/server/v9_systems.go` | Latest server system set |
| `pkg/engine/ecs.go` | World/Entity/Component framework |
| `pkg/engine/interfaces.go` | 32 core interfaces |

### Phase 1 Completion Summary

**Completion Date:** 2026-02-07  
**Full Report:** `/tmp/audit_phase1_report.md`

**Baseline Metrics:**
- Total Packages: 113
- AUDIT.md Coverage: 110/113 (97.3%)
- Systems: 141 defined, 113 registered (client), 61 (server)
- Components: 39 types
- Interfaces: 32
- Test Coverage: 82.4% average
- Build Status: ✅ Client and Server both build successfully

**Issues Identified:**
- 8 packages with failing tests (mostly X11/Ebiten-related in headless environment)
- 3 packages missing AUDIT.md files
- 28 system registration discrepancy to investigate

**Ready for Phase 2:** ✅ All preparation steps complete

---

## Phase 2: Package-by-Package Audit

Audit packages in the order listed below. This order ensures that foundational packages are audited before the packages that depend on them. For each package, follow the [Per-Package Audit Checklist](#appendix-a-per-package-audit-checklist) in Appendix A.

### Audit Group 1: Foundation Packages ✅ COMPLETED (2026-02-07)

These have no internal dependencies. Audit them first.

| # | Package | AUDIT.md | What to focus on | Status |
|---|---------|----------|------------------|--------|
| 1 | `pkg/errors` | `pkg/errors/AUDIT.md` | Error types, correlation IDs, helpers | ✅ Complete |
| 2 | `pkg/logging` | `pkg/logging/AUDIT.md` | Structured logging, logrus field conventions | ✅ Complete |
| 3 | `pkg/config` | `pkg/config/AUDIT.md` | Configuration types, validation | ✅ Complete |
| 4 | `pkg/version` | `pkg/version/AUDIT.md` | Version management | ✅ Complete |
| 5 | `pkg/recovery` | `pkg/recovery/AUDIT.md` | Panic recovery handlers | ✅ Complete |
| 6 | `pkg/stability` | `pkg/stability/AUDIT.md` | Stability monitoring | ✅ Complete |
| 7 | `pkg/observability` | `pkg/observability/AUDIT.md` | Metrics and observability | ✅ Complete |
| 8 | `pkg/validation` | `pkg/validation/AUDIT.md` | Input validation (chat, rate limiting, trade) | ✅ Complete |
| 9 | `pkg/security` | `pkg/security/AUDIT.md` | Security audit and persistence | ✅ Complete |

**Audit Summary:**
- All 9 packages: ✅ Builds passing, tests passing
- Average coverage: 96.6% (exceeds 65% minimum)
- All packages have completed audit checklists
- No TODOs/FIXMEs/HACKs in production code

**Steps for each package in this group:**
1. Run `go test -v -cover ./pkg/<name>/...` and record coverage.
2. Run `go vet ./pkg/<name>/...` and fix any issues.
3. Complete the [Per-Package Audit Checklist](#appendix-a-per-package-audit-checklist).
4. Update or create the package's `AUDIT.md` with findings.

### Audit Group 2: Core Engine ✅ COMPLETED (2026-02-07)

The ECS framework and all 141 game systems.

| # | Package | AUDIT.md | What to focus on |
|---|---------|----------|------------------|
| 10 | `pkg/engine` | `pkg/engine/AUDIT.md` | ECS core (ecs.go, interfaces.go, components.go), 141 systems, spatial partitioning |
| 11 | `pkg/engine/performance` | `pkg/engine/performance/AUDIT.md` | Performance monitoring |
| 12 | `pkg/engine/physics` | `pkg/engine/physics/AUDIT.md` | Physics subsystem coordinator |
| 13 | `pkg/engine/physics/fluids` | `pkg/engine/physics/fluids/AUDIT.md` | Buoyancy, swimming, flooding |
| 14 | `pkg/engine/physics/destruction` | `pkg/engine/physics/destruction/AUDIT.md` | Environmental destruction, debris |
| 15 | `pkg/engine/physics/vehicle` | `pkg/engine/physics/vehicle/AUDIT.md` | Suspension, weight transfer, collision |
| 16 | `pkg/engine/prestige` | `pkg/engine/prestige/AUDIT.md` | New Game+ and prestige progression |
| 17 | `pkg/engine/qol` | `pkg/engine/qol/AUDIT.md` | Auto-loot, craft queue, mount whistle, etc. |

**Audit Summary:**
- All 8 packages: ✅ AUDIT.md files present and complete
- `pkg/engine`: 65.3% coverage, 141 systems, interface consolidation complete
- `pkg/engine/performance`: Production ready, efficient implementation
- `pkg/engine/physics`: Parent package coordinator, sub-packages audited separately
- `pkg/engine/physics/fluids`: 95.3% coverage, comprehensive fluid simulation
- `pkg/engine/physics/destruction`: 81.4% coverage, 21/21 tests passing
- `pkg/engine/physics/vehicle`: Grade A+, exemplary implementation
- `pkg/engine/prestige`: Optimal structure, well-documented
- `pkg/engine/qol`: 94.6% coverage, production-ready

**Steps for `pkg/engine` (the largest package):**
1. Run `go test -v -cover ./pkg/engine/` and record the coverage (target: ≥65%).
2. Verify every `*_system.go` file defines a struct implementing `System` interface:
   ```bash
   grep -r "type.*System struct" pkg/engine/*.go | grep -v test
   ```
3. Verify every system is registered in `cmd/client/handlers.go` or `cmd/server/v*_systems.go`:
   ```bash
   grep "AddSystem" cmd/client/handlers.go
   ```
4. Check system initialization order respects dependencies (see System Initialization Sequence below).
5. Verify component definitions are pure data with only `Type() string`:
   ```bash
   grep -r "func.*Type() string" pkg/engine/*_components.go
   ```
6. Run the per-package audit checklist and update `pkg/engine/AUDIT.md`.

#### System Initialization Sequence

Systems must be registered in this order in `cmd/client/handlers.go`:

1. **Core Systems** — Performance monitoring, Input processing, Camera, Rotation
2. **Player Systems** — Player combat, item use, spell casting
3. **Physics Systems** — Movement, Collision (pixel-perfect), Projectile, Vehicle, Fluid, Destruction
4. **Combat Systems** — Combat resolution, Status effects, Revival
5. **AI Systems** — AI decisions, Behavior trees, Squad coordination, Companion AI
6. **Progression Systems** — XP/leveling, Prestige, Objective tracking, Class progression
7. **Rendering Systems** — Animation, Equipment visuals, Particles, Lighting, Post-processing
8. **Social Systems** — Chat, Guild management, Trading, Social persistence

### Audit Group 3: Procedural Generation (26 sub-packages)

All generators must implement `procgen.Generator` interface and use deterministic seed-based RNG.

| # | Package | AUDIT.md | What to focus on | Status |
|---|---------|----------|------------------|--------|
| 18 | `pkg/procgen` | `pkg/procgen/AUDIT.md` | Root generator interface, shared types | ✅ Complete (2026-02-07) |
| 19 | `pkg/procgen/terrain` | `pkg/procgen/terrain/AUDIT.md` | BSP, cellular automata, L-system, Voronoi, city | ✅ Complete (2026-02-07) |
| 20 | `pkg/procgen/entity` | `pkg/procgen/entity/AUDIT.md` | NPC/creature generation, templates | ✅ Complete (2026-02-07) |
| 21 | `pkg/procgen/item` | `pkg/procgen/item/AUDIT.md` | Item generation, rarity, class restrictions | ✅ Complete (2026-02-07) |
| 22 | `pkg/procgen/quest` | `pkg/procgen/quest/AUDIT.md` | Quest objectives, rewards, progression | ✅ Complete (2026-02-07) |
| 23 | `pkg/procgen/magic` | `pkg/procgen/magic/AUDIT.md` | Spell generation, balance calculations | ✅ Complete (2026-02-07) |
| 24 | `pkg/procgen/skills` | `pkg/procgen/skills/AUDIT.md` | Skill trees, templates, progression | ✅ Complete (2026-02-07) |
| 25 | `pkg/procgen/dialog` | `pkg/procgen/dialog/AUDIT.md` | Markov chains, personality, corpus | ✅ Complete (2026-02-07) |
| 26 | `pkg/procgen/narrative` | `pkg/procgen/narrative/AUDIT.md` | Story beats, narrative arcs | ✅ Complete (2026-02-07) |
| 27 | `pkg/procgen/story` | `pkg/procgen/story/AUDIT.md` | Archaeology, branching, cross-dungeon, timelines | ✅ Complete (2026-02-07) |
| 28 | `pkg/procgen/faction` | `pkg/procgen/faction/AUDIT.md` | Faction generation, relationships | ✅ Complete (2026-02-07) |
| 29 | `pkg/procgen/companion` | `pkg/procgen/companion/AUDIT.md` | Companion/pet generation | ✅ Complete (2026-02-07) |
| 30 | `pkg/procgen/environment` | `pkg/procgen/environment/AUDIT.md` | Environmental detail generation | ✅ Complete (2026-02-07) |
| 31 | `pkg/procgen/vehicle` | `pkg/procgen/vehicle/AUDIT.md` | Vehicle generation, combat variants |
| 32 | `pkg/procgen/legendary` | `pkg/procgen/legendary/AUDIT.md` | Legendary items and quests |
| 33 | `pkg/procgen/minigame` | `pkg/procgen/minigame/AUDIT.md` | Mini-game generation, state machine |
| 34 | `pkg/procgen/minigame/games` | `pkg/procgen/minigame/games/AUDIT.md` | Individual game implementations |
| 35 | `pkg/procgen/puzzle` | `pkg/procgen/puzzle/AUDIT.md` | Puzzle generation and solver |
| 36 | `pkg/procgen/class` | `pkg/procgen/class/AUDIT.md` | Class/multiclass generation |
| 37 | `pkg/procgen/book` | `pkg/procgen/book/AUDIT.md` | In-game book content |
| 38 | `pkg/procgen/building` | `pkg/procgen/building/AUDIT.md` | Building/structure generation |
| 39 | `pkg/procgen/furniture` | `pkg/procgen/furniture/AUDIT.md` | Furniture generation, placement |
| 40 | `pkg/procgen/station` | `pkg/procgen/station/AUDIT.md` | Crafting station generation |
| 41 | `pkg/procgen/recipe` | `pkg/procgen/recipe/AUDIT.md` | Recipe generation |
| 42 | `pkg/procgen/genre` | `pkg/procgen/genre/AUDIT.md` | Genre blending, registry |
| 43 | `pkg/procgen/audit` | `pkg/procgen/audit/AUDIT.md` | Procgen audit tests |

**Steps for each procgen sub-package:**
1. Run `go test -v -cover ./pkg/procgen/<name>/...` and record coverage (target: ≥65%).
2. Verify the generator implements `procgen.Generator` interface:
   ```bash
   grep -l "Generate(" pkg/procgen/<name>/generator.go
   ```
3. Verify deterministic generation — ensure `rand.New(rand.NewSource(seed))` is used, not global `rand`:
   ```bash
   grep -n "rand\." pkg/procgen/<name>/*.go | grep -v "rand.New\|rand.NewSource\|_test.go"
   ```
4. Verify `Validate()` method exists and is tested.
5. Complete the per-package audit checklist and update the `AUDIT.md`.

### Audit Group 4: Rendering Pipeline (16 sub-packages)

All rendering must be procedural — no external asset files.

| # | Package | AUDIT.md | What to focus on |
|---|---------|----------|------------------|
| 44 | `pkg/rendering` | `pkg/rendering/AUDIT.md` | Pipeline coordinator |
| 45 | `pkg/rendering/sprites` | `pkg/rendering/sprites/AUDIT.md` | Anatomy templates, equipment overlays, caching |
| 46 | `pkg/rendering/animation` | `pkg/rendering/animation/AUDIT.md` | Articulation, caching, directional variants |
| 47 | `pkg/rendering/tiles` | `pkg/rendering/tiles/AUDIT.md` | Parallax, wall variants, transitions |
| 48 | `pkg/rendering/lighting` | `pkg/rendering/lighting/AUDIT.md` | Bloom, ambient occlusion, dynamic lights |
| 49 | `pkg/rendering/postprocess` | `pkg/rendering/postprocess/AUDIT.md` | Chromatic aberration, color grading, blur, vignette |
| 50 | `pkg/rendering/particles` | `pkg/rendering/particles/AUDIT.md` | Behaviors, physics, LOD, weather, pooling |
| 51 | `pkg/rendering/ui` | `pkg/rendering/ui/AUDIT.md` | Chat, decorations, hierarchy, notifications |
| 52 | `pkg/rendering/palette` | `pkg/rendering/palette/AUDIT.md` | Color palettes, gradients, time-of-day |
| 53 | `pkg/rendering/patterns` | `pkg/rendering/patterns/AUDIT.md` | Texture pattern generation |
| 54 | `pkg/rendering/cache` | `pkg/rendering/cache/AUDIT.md` | Sprite caching, predictive warming, memory monitoring |
| 55 | `pkg/rendering/pool` | `pkg/rendering/pool/AUDIT.md` | Resource pooling for sprites/images |
| 56 | `pkg/rendering/parallel` | `pkg/rendering/parallel/AUDIT.md` | Parallel rendering utilities |
| 57 | `pkg/rendering/quality` | `pkg/rendering/quality/AUDIT.md` | Quality settings, LOD management |
| 58 | `pkg/rendering/display` | `pkg/rendering/display/AUDIT.md` | Display configuration |
| 59 | `pkg/rendering/shapes` | `pkg/rendering/shapes/AUDIT.md` | Shape rendering |

**Steps for each rendering sub-package:**
1. Run `go test -v -cover ./pkg/rendering/<name>/...` and record coverage.
2. Verify no external asset files are loaded (no `os.Open` or `embed` for images/audio):
   ```bash
   grep -rn "os.Open\|os.ReadFile\|embed" pkg/rendering/<name>/*.go | grep -v _test.go
   ```
3. Verify sprite caching and object pooling are used where applicable.
4. Complete the per-package audit checklist and update the `AUDIT.md`.

### Audit Group 5: Audio Pipeline (4 sub-packages)

All audio must be synthesized at runtime.

| # | Package | AUDIT.md | What to focus on |
|---|---------|----------|------------------|
| 60 | `pkg/audio` | `pkg/audio/AUDIT.md` | Audio coordinator |
| 61 | `pkg/audio/music` | `pkg/audio/music/AUDIT.md` | Adaptive soundtrack, motifs, theory |
| 62 | `pkg/audio/synthesis` | `pkg/audio/synthesis/AUDIT.md` | Oscillators, envelopes, engine |
| 63 | `pkg/audio/sfx` | `pkg/audio/sfx/AUDIT.md` | Sound effects generator, variety |

**Steps for each audio sub-package:**
1. Run `go test -v -cover ./pkg/audio/<name>/...` and record coverage.
2. Verify no external audio files are loaded.
3. Complete the per-package audit checklist and update the `AUDIT.md`.

### Audit Group 6: Network Layer (9 sub-packages)

Network variables must use interface types (e.g. `net.Addr` not `net.UDPAddr`).

| # | Package | AUDIT.md | What to focus on |
|---|---------|----------|------------------|
| 64 | `pkg/network` | `pkg/network/AUDIT.md` | Protocol, packets, compression, crypto, prediction, snapshots |
| 65 | `pkg/network/federation` | `pkg/network/federation/AUDIT.md` | Discovery, auth, handshake, sync, transfer, portal |
| 66 | `pkg/network/federation/guild` | `pkg/network/federation/guild/AUDIT.md` | Cross-server guild management |
| 67 | `pkg/network/federation/mobile` | `pkg/network/federation/mobile/AUDIT.md` | Mobile-specific federation |
| 68 | `pkg/network/federation/webrtc` | `pkg/network/federation/webrtc/AUDIT.md` | WebRTC peer connections |
| 69 | `pkg/network/federation/market` | — | Cross-server marketplace |
| 70 | `pkg/network/chat` | `pkg/network/chat/AUDIT.md` | Chat channels, E2E encryption |
| 71 | `pkg/network/trade` | `pkg/network/trade/AUDIT.md` | Trade system |
| 72 | `pkg/network/resilience` | `pkg/network/resilience/AUDIT.md` | Resilience metrics, simulator |

**Steps for each network sub-package:**
1. Run `go test -v -cover ./pkg/network/<name>/...` and record coverage.
2. Verify interface-based network variable usage — flag any concrete types:
   ```bash
   grep -rn "net.UDPAddr\|net.TCPAddr\|net.UDPConn\|net.TCPConn\|net.TCPListener\|net.UDPListener" pkg/network/<name>/*.go | grep -v _test.go
   ```
3. Verify E2E encryption is applied to chat messages.
4. Verify high-latency support (200–5000ms) in relevant packages.
5. Complete the per-package audit checklist and update the `AUDIT.md`.

### Audit Group 7: World Management (5 sub-packages)

| # | Package | AUDIT.md | What to focus on |
|---|---------|----------|------------------|
| 73 | `pkg/world` | `pkg/world/AUDIT.md` | State, persistence, chunk loading/compression |
| 74 | `pkg/world/housing` | `pkg/world/housing/AUDIT.md` | Blueprints, guildhalls, spatial management, UI |
| 75 | `pkg/world/economy` | `pkg/world/economy/AUDIT.md` | Marketplace, pricing engine, guild bank |
| 76 | `pkg/world/territory` | `pkg/world/territory/AUDIT.md` | Territory control, siege mechanics |
| 77 | `pkg/world/raids` | `pkg/world/raids/AUDIT.md` | Raid generation, instances, lockouts |

**Steps for each world sub-package:**
1. Run `go test -v -cover ./pkg/world/<name>/...` and record coverage.
2. Verify persistence: check that `Serialize`/`Deserialize` methods exist for stateful data.
3. Verify chunk loading/unloading does not leak memory.
4. Complete the per-package audit checklist and update the `AUDIT.md`.

### Audit Group 8: Gameplay Support Packages

| # | Package | AUDIT.md | What to focus on |
|---|---------|----------|------------------|
| 78 | `pkg/combat` | `pkg/combat/AUDIT.md` | Damage calculation, interfaces, validation |
| 79 | `pkg/class` | `pkg/class/AUDIT.md` | Class system definitions |
| 80 | `pkg/class/advanced` | `pkg/class/advanced/AUDIT.md` | Advanced multiclassing |
| 81 | `pkg/companion` | `pkg/companion/AUDIT.md` | Companion system |
| 82 | `pkg/companion/learning` | `pkg/companion/learning/AUDIT.md` | Companion learning system |
| 83 | `pkg/narrative` | `pkg/narrative/AUDIT.md` | Branching narrative types |
| 84 | `pkg/narrative/branching` | `pkg/narrative/branching/AUDIT.md` | Branching narrative implementation |
| 85 | `pkg/balance` | `pkg/balance/AUDIT.md` | Combat and economic balance |
| 86 | `pkg/social` | `pkg/social/AUDIT.md` | Social system persistence |
| 87 | `pkg/social/persistence` | `pkg/social/persistence/AUDIT.md` | Social data persistence |

**Steps for each package:**
1. Run `go test -v -cover ./pkg/<name>/...` and record coverage.
2. Verify the package uses interface abstractions (not concrete types) for cross-system calls.
3. Complete the per-package audit checklist and update the `AUDIT.md`.

### Audit Group 9: Infrastructure Packages

| # | Package | AUDIT.md | What to focus on |
|---|---------|----------|------------------|
| 88 | `pkg/saveload` | `pkg/saveload/AUDIT.md` | Save/load manager, migrator, recovery, WASM |
| 89 | `pkg/migration` | `pkg/migration/AUDIT.md` | Data migration, validation |
| 90 | `pkg/modding` | `pkg/modding/AUDIT.md` | Mod loader, manager, sandboxed execution |
| 91 | `pkg/mobile` | `pkg/mobile/AUDIT.md` | Controls, touch input, dual joystick, keyboard |
| 92 | `pkg/hostplay` | `pkg/hostplay/AUDIT.md` | Host-and-play (local server + client) |

**Steps for each package:**
1. Run `go test -v -cover ./pkg/<name>/...` and record coverage.
2. For `pkg/saveload`: verify WASM storage compatibility and migration support.
3. For `pkg/modding`: verify sandboxing — no executable code allowed in mods.
4. Complete the per-package audit checklist and update the `AUDIT.md`.

### Audit Group 10: Integration Packages (10 sub-packages)

These depend on multiple other packages. Audit them after their dependencies.

| # | Package | AUDIT.md | What to focus on |
|---|---------|----------|------------------|
| 93 | `pkg/integration` | `pkg/integration/AUDIT.md` | Integration overview |
| 94 | `pkg/integration/companion_housing` | `pkg/integration/companion_housing/AUDIT.md` | Companion home, bedding, training |
| 95 | `pkg/integration/guild_housing` | `pkg/integration/guild_housing/AUDIT.md` | Guild halls, permissions, upgrades |
| 96 | `pkg/integration/guild_vehicle` | `pkg/integration/guild_vehicle/AUDIT.md` | Guild fleet management |
| 97 | `pkg/integration/housing_crafting` | `pkg/integration/housing_crafting/AUDIT.md` | Housing + crafting |
| 98 | `pkg/integration/choice_consequences` | `pkg/integration/choice_consequences/AUDIT.md` | Narrative choice tracking |
| 99 | `pkg/integration/narrative_world` | `pkg/integration/narrative_world/AUDIT.md` | Narrative + world state |
| 100 | `pkg/integration/political_warfare` | `pkg/integration/political_warfare/AUDIT.md` | Political/faction warfare |
| 101 | `pkg/integration/trade_routes` | `pkg/integration/trade_routes/AUDIT.md` | Trade route management |
| 102 | `pkg/integration/world_events` | `pkg/integration/world_events/AUDIT.md` | World event management |

**Steps for each integration package:**
1. Run `go test -v -cover ./pkg/integration/<name>/...` and record coverage.
2. Verify cross-system interactions — confirm that each integration bridges the packages it claims to.
3. Verify no circular dependencies:
   ```bash
   go vet ./pkg/integration/<name>/...
   ```
4. Complete the per-package audit checklist and update the `AUDIT.md`.

### Audit Group 11: Command Packages (Application Entry Points)

Audit these last since they depend on everything above.

| # | Package | AUDIT.md | What to focus on |
|---|---------|----------|------------------|
| 103 | `cmd/client` | `cmd/client/AUDIT.md` | Desktop client, system registration, UI, WASM |
| 104 | `cmd/server` | `cmd/server/AUDIT.md` | Dedicated server, player connections, validation |
| 105 | `cmd/mobile` | `cmd/mobile/AUDIT.md` | Mobile entry point (iOS/Android) |

**Steps for each command package:**
1. Verify the package builds: `go build ./cmd/<name>`.
2. For `cmd/client`: confirm every system in `pkg/engine` is registered in `handlers.go`.
3. For `cmd/server`: confirm server system versions (v4/v8/v9) register the correct systems.
4. Complete the per-package audit checklist and update the `AUDIT.md`.

### Audit Group 12: Testing & Quality Packages

| # | Package | AUDIT.md | What to focus on |
|---|---------|----------|------------------|
| 106 | `pkg/audit` | `pkg/audit/AUDIT.md` | Audit tools |
| 107 | `pkg/audit/features` | `pkg/audit/features/AUDIT.md` | Feature audit tests |
| 108 | `pkg/visualtest` | `pkg/visualtest/AUDIT.md` | Visual regression testing |
| 109 | `pkg/visualtest/parity` | `pkg/visualtest/parity/AUDIT.md` | Cross-platform parity tests |
| 110 | `pkg/ux` | `pkg/ux/AUDIT.md` | UX validation, user journeys |

**Steps for each package:**
1. Run the package's tests and record results.
2. Verify the audit/test utilities themselves are well-tested.
3. Complete the per-package audit checklist and update the `AUDIT.md`.

---

## Phase 3: Cross-Cutting Verification

After all packages have been individually audited, perform these cross-cutting checks.

### Step 3.1: Integration Validation

Verify that all system categories are fully connected end-to-end.

#### 3.1.1 Procedural Generation Integration
- [ ] All 26 generators implement `procgen.Generator` interface
- [ ] All generators use seed-based deterministic RNG
- [ ] Terrain → Entity → Item → Quest generation pipeline is connected
- [ ] Generated content renders correctly via the rendering pipeline
- [ ] Generated content persists correctly via the save/load system

**Validation command:**
```bash
grep -r "func.*Generate(seed" pkg/procgen/*/generator.go | wc -l  # Should be 26
```

#### 3.1.2 Physics Integration
- [ ] Movement → Collision → Terrain interaction chain works
- [ ] Vehicle physics (suspension, weight transfer) integrated
- [ ] Fluid dynamics (buoyancy, swimming, flooding) integrated
- [ ] Environmental destruction (debris, propagation) integrated

#### 3.1.3 Combat Integration
- [ ] Player combat → Damage calculation → Status effects → Revival chain works
- [ ] Combat triggers screen shake (camera), hit particles, and sound effects
- [ ] Spell casting (keys 1–8) resolves through the combat pipeline

#### 3.1.4 Character Progression Integration
- [ ] XP → Level up → Skill unlock → Class progression chain works
- [ ] 15 base classes + 20 prestige classes are all accessible
- [ ] Prestige / New Game+ carries over correct bonuses

#### 3.1.5 Social Systems Integration
- [ ] Chat (Enter key) is E2E encrypted end-to-end
- [ ] Guild system (G key) integrates with guild housing and cross-server sync
- [ ] Trading is secure with proper validation
- [ ] Mail system (L key) delivers messages

#### 3.1.6 World Interaction Integration
- [ ] NPC dialog (F key) triggers correctly
- [ ] Crafting (B key) connects to recipe and station systems
- [ ] Housing (H key) connects to persistence and economy
- [ ] Vehicles (V key) connect to physics engine
- [ ] Companions (P key) connect to AI and learning systems

#### 3.1.7 Network Integration
- [ ] Client-server entity synchronization works
- [ ] Client-side prediction and lag compensation (200–5000ms) work
- [ ] Federation discovery, handshake, and portal transfer work
- [ ] WebRTC falls back to TCP/UDP correctly

### Step 3.2: UI/Control Access Verification

Verify every in-game UI system is accessible via its documented control:

| Key | System | Implementation File | Status |
|-----|--------|---------------------|--------|
| I | Inventory | `inventory_ui.go` | [ ] |
| C | Character Sheet | `hud_system.go` | [ ] |
| K | Skills | `skill_progression_system.go` | [ ] |
| J | Quest Log | `quest_ui.go` | [ ] |
| M | Map | `render_system.go` | [ ] |
| B | Crafting | `crafting_ui.go` | [ ] |
| L | Mail | `mail_system.go` | [ ] |
| G | Guild | `guild_ui.go` | [ ] |
| V | Vehicle | `vehicle_system.go` | [ ] |
| P | Companion | `companion_ai_system.go` | [ ] |
| H | Housing | `pkg/world/housing/ui.go` | [ ] |
| Enter | Chat | `chat_system.go` | [ ] |
| F1 | Help | `help_system.go` | [ ] |
| Esc | Menu | `menu_system.go` | [ ] |

**Validation command:**
```bash
grep -A 5 "case ebiten.Key" pkg/engine/input_system.go
```

### Step 3.3: Persistence Verification

Verify that all stateful systems persist and restore correctly:

```bash
# Find all Serialize/Deserialize implementations
grep -rn "func.*Serialize\|func.*Deserialize" pkg/engine/*.go pkg/world/*.go | grep -v _test.go

# Check save/load integration
find pkg/saveload -name "*.go" | xargs grep "Serialize\|Deserialize"

# Verify migration support
ls pkg/migration/*_test.go
```

### Step 3.4: Performance Verification

Run benchmarks and compare against baselines in Appendix E:

```bash
# Run benchmarks
go test -bench=. -benchmem ./pkg/engine/ ./pkg/rendering/... ./pkg/procgen/...

# CPU profiling
./scripts/profile_cpu.sh

# Memory profiling
go test -memprofile=mem.prof ./pkg/engine/
go tool pprof mem.prof
```

**Performance targets:**
- 60 FPS minimum with 2000 entities
- <500MB client memory
- <1GB server memory (4 players)
- Sprite cache hit rate >95%

### Step 3.5: Documentation Completeness

Verify all packages have complete documentation:

```bash
# Check for missing AUDIT.md files
go list ./pkg/... | while read pkg; do
  dir=$(echo "$pkg" | sed 's|.*venture/||')
  if [ ! -f "$dir/AUDIT.md" ]; then
    echo "MISSING: $dir/AUDIT.md"
  fi
done

# Check for missing doc.go files
find pkg -type d | while read dir; do
  if ls "$dir"/*.go &>/dev/null && [ ! -f "$dir/doc.go" ]; then
    echo "MISSING: $dir/doc.go"
  fi
done

# Verify exported symbols have godoc comments
go vet ./pkg/...
```

### Step 3.6: Compile the Final Audit Report

After completing all package audits and cross-cutting checks:

1. Aggregate per-package coverage numbers into a summary table.
2. List any packages below the 65% coverage minimum.
3. List any unresolved findings (bugs, missing implementations, interface violations).
4. List any performance regressions compared to baselines.
5. Update this file's "Last Updated" date.

---

## Appendix A: Per-Package Audit Checklist

Use this checklist for **every** package audit. Copy it into the package's `AUDIT.md` and fill it in.

### 1. Build & Test
- [ ] Package builds: `go build ./path/to/pkg/...`
- [ ] Package passes vet: `go vet ./path/to/pkg/...`
- [ ] All tests pass: `go test -v ./path/to/pkg/...`
- [ ] Test coverage recorded: `go test -cover ./path/to/pkg/...`
- [ ] Coverage meets minimum (≥65%)

### 2. Code Quality
- [ ] No TODO/FIXME/HACK in production code
- [ ] All exported symbols have godoc comments
- [ ] Errors are handled (no ignored return values)
- [ ] Structured logging with `logrus.Fields` used (not `fmt.Printf`)
- [ ] No dead code or unused imports

### 3. System Initialization (for `pkg/engine` systems only)
- [ ] System struct implements `System` interface (`Update(entities, deltaTime)`)
- [ ] Constructor exists (`NewXXXSystem(...)`)
- [ ] System registered in `cmd/client/handlers.go` or `cmd/server/v*_systems.go`
- [ ] Dependencies injected via setters or constructor
- [ ] Initialization order respects dependencies

### 4. Deterministic Generation (for `pkg/procgen` packages only)
- [ ] Generator implements `procgen.Generator` interface
- [ ] Uses `rand.New(rand.NewSource(seed))`, not global `rand`
- [ ] Same seed produces identical output
- [ ] `Validate()` method exists and is tested

### 5. Network Compliance (for `pkg/network` packages only)
- [ ] Uses `net.Addr` (not `net.UDPAddr`/`net.TCPAddr`)
- [ ] Uses `net.PacketConn` (not `net.UDPConn`)
- [ ] Uses `net.Conn` (not `net.TCPConn`)
- [ ] Uses `net.Listener` (not concrete listener types)
- [ ] No type switches/assertions to concrete network types

### 6. No External Assets (all packages)
- [ ] No external image/audio/data files loaded at runtime
- [ ] All content generated procedurally

### 7. Data Persistence (if stateful)
- [ ] Component serialization implemented
- [ ] Save/load integration with `pkg/saveload`
- [ ] Migration support for version changes
- [ ] WASM storage compatibility (if applicable)

### 8. Resource Management
- [ ] Object pooling used where applicable
- [ ] Cache integration where applicable
- [ ] Cleanup on entity removal
- [ ] No memory leaks

### 9. Cross-System Interactions
- [ ] Dependencies documented
- [ ] Interface abstractions used for testability
- [ ] No circular dependencies
- [ ] Integration tests exist (if multi-system)

### 10. Security
- [ ] Input validation on all user-supplied data
- [ ] No secrets in source code
- [ ] Encryption used for sensitive network traffic
- [ ] Mod system sandboxing enforced (if applicable)

---

## Appendix B: Audit Command Reference

### Discovery Commands
```bash
# Count all systems
grep -r "type.*System struct" pkg/engine/*.go | grep -v test | wc -l

# Count all AUDIT.md files
find . -name "AUDIT.md" -type f | wc -l

# Find all generators
find pkg/procgen -name "generator.go"

# List all UI files
find pkg/engine -name "*_ui.go"

# Check system registration
grep "AddSystem" cmd/client/handlers.go

# Find keyboard bindings
grep "case ebiten.Key" pkg/engine/input_system.go

# Find all interfaces
grep -A 10 "^type.*interface" pkg/engine/interfaces.go
```

### Validation Commands
```bash
# Run all tests
go test ./...

# Check test coverage (all packages)
go test -cover ./pkg/...

# Run specific package tests
go test -v ./pkg/engine/

# Run benchmarks
go test -bench=. ./pkg/...

# Profile CPU
./scripts/profile_cpu.sh

# Profile memory
go test -memprofile=mem.prof ./pkg/engine/
go tool pprof mem.prof
```

### Build Commands
```bash
# Desktop builds
go build ./cmd/client
go build ./cmd/server

# Mobile builds
./scripts/build-android.sh
./scripts/build-ios.sh

# WASM build
GOOS=js GOARCH=wasm go build -o web/venture.wasm ./cmd/client
```

---

## Appendix C: System Catalog

### Complete System Inventory (141 Systems)

All systems located in `pkg/engine/` unless otherwise specified. See `pkg/engine/AUDIT.md` for the detailed breakdown.

| Category | Count | Examples |
|----------|-------|---------|
| Core Systems | 10 | Input, Camera, Rotation, Performance |
| Combat Systems | 10 | Player/NPC combat, spells, status effects |
| AI Systems | 12 | AI, behavior trees, companions, squads |
| Progression Systems | 15 | XP, skills, classes, achievements |
| Inventory Systems | 8 | Inventory, equipment, loot |
| Crafting & Economy | 8 | Crafting, commerce, economy |
| Social Systems | 12 | Chat, guilds, trading, mail |
| World Systems | 16 | Weather, cities, territories, events |
| Vehicle & Mount Systems | 8 | Vehicles, mounts, customization |
| Housing Systems | 6 | Housing, furniture, blueprints |
| Narrative Systems | 10 | Quests, branching narratives, books |
| Rendering Systems | 12 | Render, animation, lighting, particles |
| UI Systems | 10 | Menu, HUD, various UI screens |
| Audio Systems | 3 | Audio manager, music, sound effects |
| **Total** | **141** | |

---

## Appendix D: Access Requirement Matrix

| System | Keyboard | Touch | Level Req | Quest Req | Tutorial |
|--------|----------|-------|-----------|-----------|----------|
| Inventory | I | Bag Icon | 1 | None | ✅ Basic |
| Character Sheet | C | Stats Icon | 1 | None | ✅ Basic |
| Skills | K | Skills Icon | 1 | None | ✅ Basic |
| Quest Log | J | Quest Icon | 1 | None | ✅ Basic |
| Map | M | Map Icon | 1 | None | ✅ Basic |
| Crafting | B | Craft Icon | 5 | "Crafting 101" | ✅ Advanced |
| Mail | L | Mail Icon | 1 | None | ✅ Basic |
| Guild | G | Guild Icon | 10 | None | ✅ Advanced |
| Housing | H | Home Icon | 10 | Optional | ✅ Advanced |
| Vehicle | V | — | 15 | "Driver's License" | ✅ Advanced |
| Companion | P | Pet Icon | 5 | "First Friend" | ✅ Advanced |
| Chat | Enter | Chat Icon | 1 | None | ✅ Basic |
| Help | F1 | ? Icon | 1 | None | ✅ Basic |

**Legend:** ✅ Basic = Tutorial in first 10 minutes. ✅ Advanced = Tutorial on first access.

---

## Appendix E: Performance Baselines

### System Update Performance (v8.0)

| System | Entities | Time/Frame | % of 16.67ms |
|--------|----------|------------|--------------|
| MovementSystem | 2000 | 0.8ms | 4.8% |
| CollisionSystem | 2000 | 2.1ms | 12.6% |
| AISystem | 500 | 1.2ms | 7.2% |
| CombatSystem | 100 | 0.3ms | 1.8% |
| RenderSystem | 2000 | 4.5ms | 27.0% |
| AnimationSystem | 2000 | 1.9ms | 11.4% |
| ParticleSystem | 1000 | 0.7ms | 4.2% |
| UISystem | 1 | 0.2ms | 1.2% |
| **Total** | — | **11.7ms** | **70.2%** |

**Target:** <16.67ms per frame (60 FPS)  
**Actual:** 11.7ms average (89 FPS with 2000 entities)

### Memory Usage (v8.0)

| Component | Memory | % of 500MB Budget |
|-----------|--------|-------------------|
| Entity Pool | 45MB | 9.0% |
| Component Data | 30MB | 6.0% |
| Sprite Cache | 35MB | 7.0% |
| Audio Buffers | 5MB | 1.0% |
| Network Buffers | 3MB | 0.6% |
| UI Textures | 2MB | 0.4% |
| **Total** | **120MB** | **24.0%** |

---

## References

- `pkg/engine/AUDIT.md` — Engine package audit (65.3% coverage, 141 systems)
- `pkg/procgen/AUDIT.md` — Procedural generation audit (92.1% coverage, 26 generators)
- `pkg/integration/AUDIT.md` — Integration package audit (11 integrations)
- `docs/SYSTEM_INTERACTION_MAP.md` — System dependency graph
- `docs/CONTROLS.md` — Complete control reference
- `docs/ARCHITECTURE.md` — ECS architecture overview
- `docs/TECHNICAL_SPEC.md` — Technical specifications
- `README.md` — Project overview and features

---

**Last Updated:** 2026-02-07  
**Next Review:** On major version release (v9.0)  
**Maintainer:** See CONTRIBUTING.md
