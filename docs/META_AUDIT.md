TASK: Audit one Go sub-package for implementation completeness; produce structured findings.

EXECUTION MODE: Autonomous action — select package, audit, write files, report.

---

## Phase 0: Pre-Audit Preparation

1. **Read root `AUDIT.md`** to identify already-audited packages.
2. **Read `pkg/engine/interfaces.go`** to understand all defined `System`, `Component`, and `Input` interfaces.

## Phase 0.5: Full-Stack Integration Baseline

Before beginning the per-package audit, verify that every major subsystem is **on by default** — meaning it is initialized, registered, and reachable without requiring manual flags, hidden toggles, or developer-only configuration. A feature that exists in code but is not wired into the default startup path is treated as a **High Severity** integration gap.

Check each row and mark ✅ (on by default), ⚠️ (present but requires opt-in), or ❌ (missing/dead code):

| Subsystem | Default Entry Point | What to verify |
|---|---|---|
| **Main Menu** | `cmd/client/` startup | Main menu is the first screen shown; all sub-menus (New Game, Continue, Settings, Multiplayer, Quit) are reachable without code modification; seed entry and genre selector present at game start |
| **Tutorial / Onboarding** | First launch or New Game flow | Tutorial system initialized and shown by default on first run; step progression, input-device-adaptive prompts, and dismiss/re-open all functional; not gated behind a debug flag |
| **Character Creation** | New Game → Character Creation | Character creation screen reachable from main menu; class selection, appearance, name entry, and starting stat allocation wired; genre-specific creation options present; feeds correctly into ECS entity initialization |
| **AI Systems** | Server/client startup | `ai_system.go`, `behavior_tree_system.go`, `squad_system.go`, `companion_ai_system.go` all registered in default system list; NPCs exhibit behavior on game start without manual enable; AI ticks correctly relative to movement and combat systems |
| **Procedural Generation** | New Game / zone load | Terrain, entity, item, quest, dialog, and narrative generators all invoked on new game creation; all use seed from CLI/config (not hardcoded); genre parameter propagated to every generator; async terrain loading active by default |
| **Networking (Client/Server)** | `cmd/server/` startup, `cmd/client/` join flow | Server starts and accepts connections by default on configured port; client multiplayer join flow reachable from main menu; high-latency mode activatable via `-high-latency` flag without code changes |
| **Federation** | Server startup with federation config | Cross-server discovery, handshake, sync, and portal systems initialize when federation config present; circuit breaker and retry logic active; guild and market federation on by default when enabled |
| **WebRTC** | WASM client build | WebRTC signaling path compiled and reachable in WASM build; peer connection initialized on multiplayer join; fallback to standard networking on non-WASM platforms |
| **Housing System** | Interact with house entity | Housing UI (`HousingUIProvider`) registered and reachable in-game; blueprint selection, furniture placement, and permissions functional by default; guildhall tab present for guild officers |
| **Guild System** | Social menu or guild NPC | Guild creation, member management, bank, log, and cross-server federation all reachable via default UI path; rank-gated controls enforced |
| **Economy / Marketplace** | Shop NPC interaction | Pricing engine active; buy/sell UI reachable from vendor interaction; cross-server market available when federation enabled; transaction logging via structured logging |
| **Weather & World Events** | Gameplay state | `weather_system.go` and `world_events_system.go` registered and ticking; weather visually present; world events trigger according to procedural schedule |
| **Progression Systems** | Gameplay state | `progression_system.go`, `skill_progression_system.go`, `class_progression_system.go`, `reputation_system.go`, `achievement.go` all registered; XP gain, level-up, and skill unlock functional without manual wiring |
| **Combat Systems** | Gameplay state | `combat_system.go`, `player_combat_system.go`, `spell_casting.go`, `status_effect_system.go` all registered; player can attack, cast, and receive status effects on game start |
| **Crafting System** | Crafting station interaction | Recipe list populated from `pkg/procgen/recipe/`; craft queue (`pkg/engine/qol/`) active; station types from `pkg/procgen/station/` generated and placeable |
| **Save / Load** | Pause menu → Save; Continue on main menu | Save writes all persistent components via `pkg/saveload/`; load restores full game state; WASM storage path used on browser builds; migration path active for version changes |
| **Mod System** | Startup (if mods directory present) | `pkg/modding/` loader scans mod directory on startup; JSON rule mods applied before first generate call; sandboxing prevents executable code; invalid mods rejected with structured log error |
| **Audio** | Gameplay state | Adaptive music (`audio/music/`) and sound effects (`audio/sfx/`) initialized on game start; genre-based motifs generated from seed; volume respects settings |
| **Chat** | HUD / keybind | Chat system and network chat channels initialized; global, party, guild, and whisper channels selectable; validation (rate limit, profanity) active by default |
| **QoL Systems** | Gameplay state | Auto-loot, craft queue, mount whistle, recipe tracker, and storage sorter (`pkg/engine/qol/`) all registered and active by default; togglable via settings |
| **Physics Subsystems** | Gameplay state | Fluid simulation, vehicle physics, and environmental destruction systems registered when relevant entities exist; buoyancy and flooding active in appropriate terrain tiles |
| **VR / Stereoscopic** | `-vr` flag or auto-detect | VR mode activates stereoscopic rendering and controller input when VR headset detected or flag set; no crash or undefined behavior when VR hardware absent |
| **Prestige / New Game+** | Post-completion flow | Prestige and New Game+ systems (`pkg/engine/prestige/`) reachable after game completion; carry-over rules applied; not accessible before completion condition met |

**Flag as High Severity if:**
- Any subsystem in this table requires editing source code (not config/flags) to enable.
- A subsystem is initialized but not connected to the ECS update loop (registered but never ticked).
- A subsystem's UI is reachable but its backing system is not registered, causing silent no-ops.
- Procedural generators are called without a genre parameter, producing genre-blind output.
- Federation, WebRTC, or networking systems panic or log fatal errors on startup in single-player mode (non-fatal absence of peers must be handled gracefully).
- Tutorial or character creation can be bypassed in a way that leaves required components uninitialized on the player entity.

4. **Read `cmd/client/` entry point** to understand `EbitenGame` state machine, system initialization order, and lazy-init patterns.
4. **Read `cmd/client/` entry point** to understand `EbitenGame` state machine, system initialization order, and lazy-init patterns.
5. **Read `cmd/server/` entry point** to understand server-side system registration and validation layers.
6. **Catalog all existing input providers and menu/UI systems** (see Phase 2 and Phase 3 below) so audits can verify integration.

---

## Phase 1: Package Selection

1. Pick **ONE** un-audited sub-package from `pkg/` or `cmd/`. Prefer packages that:
   - Are listed in root `AUDIT.md` but unchecked.
   - Have high integration surface (many imports/importers).
   - Fall under core domains: `engine`, `procgen`, `rendering`, `network`, `world`.
   - Contain input handling, UI/menu systems, or system-initialization code (high-risk for missed wiring).
2. State the chosen package **and rationale** before proceeding.

---

## Phase 2: Input Integration Audit

For every input subsystem the codebase supports, verify the audited package correctly integrates with it **if the package's domain requires input**. Check each row; mark N/A if the package has no input responsibilities.

| Input Source | What to verify | Files to inspect |
|---|---|---|
| **Keyboard** | - Key bindings registered<br>- `ebiten.IsKeyPressed` / `inpututil.IsKeyJustPressed` routed through the `Input` interface (not called directly in systems)<br>- Rebindable key map honored | `pkg/engine/interfaces.go` (`Input` interface), `cmd/client/` key handling, `pkg/engine/input_*.go` |
| **Mouse** | Click, drag, scroll events consumed via `Input` interface; UI hit-testing uses screen-to-world coordinate conversion; cursor state (default, pointer, crosshair) set correctly | `cmd/client/`, `pkg/engine/menu_system.go`, `pkg/rendering/ui/` |
| **Gamepad / Controller** | `ebiten.GamepadIDs()` enumerated; axis dead-zones applied; button mappings abstracted behind `Input` interface; hot-plug handled (gamepad connect/disconnect) | `pkg/engine/interfaces.go`, `cmd/client/`, any `gamepad*.go` or `controller*.go` |
| **Touch (Mobile)** | Touch IDs tracked; dual-joystick virtual controls (`pkg/mobile/`) route through `Input` interface; gesture recognition (tap, long-press, swipe, pinch) present | `pkg/mobile/`, `examples/virtual_controls_wasm_demo/`, `cmd/mobile/` |
| **VR Controllers** | VR controller pose/button input routed via `Input` interface; stereoscopic rendering triggers on VR mode flag | Any `vr_*.go`, `pkg/engine/interfaces.go` |
| **WASM-specific input** | Browser keyboard/mouse events not lost; virtual on-screen controls shown when no gamepad detected; clipboard paste for chat | `web/`, `cmd/client/` WASM build tags |
| **Stub/Test Input** | `StubInput` implementation exists and covers every method of `Input` interface; used in all system unit tests that require input | `pkg/engine/*_test.go`, `StubInput` definition |

**Flag as issues:**
- Any direct Ebiten input or `inpututil` calls (e.g., `ebiten.IsKeyPressed`, `ebiten.IsMouseButtonPressed`, `ebiten.Gamepad*`, `inpututil.IsKeyJustPressed`) inside `System.Update()` or other systems instead of going through the `Input` interface (violates coding guideline #2, CodingGuidelineID: 2, Input interface abstraction).
- Missing dead-zone or sensitivity configuration.
- Touch input not falling back gracefully on desktop builds.
- `StubInput` missing methods added to the `Input` interface after stub was written.

---

## Phase 3: Menu & UI Integration Audit

For every menu/UI screen the game exposes, verify it is **reachable from the game state machine**, **handles all input sources**, and **connects to the correct backing system**. Check each row:

| Menu / UI Screen | State machine trigger | Backing system(s) | What to verify |
|---|---|---|---|
| **Main Menu** | Game launch / ESC from gameplay | `menu_system.go` | New Game, Continue, Settings, Quit options wired; seed entry field present; genre selector present; multiplayer join/host option present |
| **Settings / Options** | Main menu → Settings, or in-game pause → Settings | `menu_system.go`, `pkg/config/` | Audio volume, resolution, keybindings, quality preset, VR toggle, accessibility options all read/write `Config`; changes persisted via `pkg/saveload/` |
| **Tutorial / Help** | First launch or main menu → Tutorial | `pkg/rendering/ui/tutorial*.go`, `pkg/engine/tutorial_system.go` (if exists) | Tutorial overlay renders; step progression works; input prompts adapt to current input device (keyboard glyph vs. gamepad glyph vs. touch icon); can be dismissed and re-opened |
| **HUD (Heads-Up Display)** | Active during gameplay state | `hud_system.go` | Health bar, mana/stamina bar, minimap, hotbar, chat input area, notification toasts all rendered; HUD elements toggle off in menus; HUD scales with resolution |
| **Inventory** | Keybind (default `I`) or HUD button | `inventory_ui.go`, `pkg/engine/inventory_*.go` | Grid/list renders items; drag-and-drop or click-to-equip; tooltip on hover/long-press; sorting; item comparison; context menu (use, drop, split stack); controller navigation with D-pad; touch scroll |
| **Character / Stats** | Keybind (default `C`) | `character_ui.go` or `stats_ui.go`, `components.go` (`StatsComponent`) | Displays all stats from `StatsComponent`; equipment slots visualized; class/level/XP bar; stat tooltips; attribute point allocation if applicable |
| **Skill Tree** | Keybind or Character menu tab | `skill_progression_system.go`, `class_progression_system.go` | Tree renders nodes; unlockable nodes highlighted; spent/available points shown; hover/tap shows description; undo/respec button if supported |
| **Quest Log / Tracker** | Keybind (default `J` or `L`) | `quest_ui.go`, `quest_tracker.go` | Active/completed/failed tabs; quest description, objectives with progress, rewards listed; click-to-track pins quest to HUD; world map marker integration |
| **Map / Minimap** | Keybind (default `M`) or HUD minimap click | `map_system.go`, `pkg/engine/minimap*.go` | Full map opens; fog of war respected; zoom/pan with mouse-drag or pinch; waypoints placeable; quest markers shown; player position indicated; chunk boundaries not visible as seams |
| **Shop / Vendor** | Interact with NPC | `shop_ui.go`, `economy_system.go` | Buy/sell tabs; item prices from `pricing_engine`; stack quantity selector; insufficient-funds feedback; transaction logged via structured logging |
| **Crafting** | Keybind or interact with crafting station | `crafting_ui.go`, `pkg/procgen/recipe/`, `pkg/procgen/station/` | Recipe list filtered by known recipes; ingredient requirements shown with have/need counts; craft button disabled when missing ingredients; craft queue (`pkg/engine/qol/`) integration; output preview |
| **Trade (Player-to-Player)** | Interact with player or trade request | `trade_ui.go`, `trade_system.go`, `pkg/network/trade/`, `pkg/validation/trade*.go` | Both players' offers shown; ready/confirm two-phase commit; validation prevents item duplication; cancel button; timeout handling |
| **Chat** | Keybind (default `Enter`) or HUD chat area | `chat_system.go`, `pkg/network/chat/`, `pkg/rendering/ui/chat*.go`, `pkg/validation/chat*.go` | Channels (global, party, guild, whisper) selectable; message input with profanity/rate-limit validation; scroll history; clickable player names; WASM clipboard paste |
| **Guild** | Keybind or social menu | `guild_ui.go`, `guild_system.go`, `pkg/network/federation/guild/` | Member roster, ranks, permissions; promote/demote/kick controls gated by rank; guild bank tab; guild log; cross-server guild indicator if federated |
| **Housing** | Keybind or interact with house | `pkg/world/housing/ui.go` via `HousingUIProvider`, `pkg/integration/housing_crafting/` | Blueprint selection; furniture placement (grid snap, rotation, collision); visitor permissions; guildhall tab for guild officers; storage access |
| **Mail** | Keybind or HUD icon | `mail_system.go` | Inbox/sent tabs; compose with recipient autocomplete; attachment (item/gold) support; send/receive acknowledgment; unread badge on HUD |
| **Pause Menu** | `ESC` during gameplay | `menu_system.go` | Resume, Settings, Save, Quit to Menu, Quit to Desktop; game systems paused (deltaTime = 0) while open; multiplayer: pause only affects local UI, not server tick |
| **Death / Respawn** | Player health ≤ 0 | `hud_system.go` or dedicated respawn UI | Respawn options (checkpoint, town, gravestone); XP/gold penalty shown; timer if applicable; input locked except respawn choice |
| **Loading Screen** | State transitions (new game, zone change, teleport) | `cmd/client/` state machine | Progress bar or procedural generation status; tips or lore text; async chunk loading (`pkg/procgen/terrain/` async); input suppressed until load complete |
| **Matchmaking / Lobby** | Multiplayer → Find Game | `matchmaking_system.go`, `pkg/network/` | Server browser or queue; ping display; player count; cancel button; auto-retry on timeout for high-latency mode |
| **PvP / Tournament** | Arena NPC or matchmaking | `pvp_rating_system.go`, `tournament_system.go` | Rating displayed; queue/bracket UI; match result screen; reward claim |

**Flag as issues:**
- Menu unreachable from any game state (dead code).
- Menu opens but has no controller/touch navigation (keyboard-only).
- Menu does not pause/unpause correctly or blocks game-critical updates.
- Menu reads component data directly instead of through a system query.
- Menu has hardcoded screen coordinates instead of resolution-relative layout.
- Missing back/close button or ESC-to-close handler.
- Menu not listed in `system_init.go` or lazy-init registry.

---

## Phase 4: Core Code Quality Audit

Evaluate each item below. **Cite `file.go:LINE` for every issue found.**

| Category | What to check |
|---|---|
| **Stub / incomplete code** | Functions returning only `nil`/zero with no real logic; `TODO`/`FIXME`/`HACK`/`PLACEHOLDER`/`XXX` comments; empty method bodies; commented-out logic blocks > 10 lines; functions that only call `log.Warn("not implemented")` |
| **ECS compliance** | Components must be pure data + `Type() string` only; no logic methods on components (no `Move()`, `TakeDamage()`, `Calculate()`, etc.); systems must own all behavior; `Entity` hot-path component cache fields populated for performance-critical components; no direct `World` mutation from inside components |
| **Deterministic procgen** | Verify all randomness follows Coding Guideline #2 (Deterministic Generation): seed-based `rand.New(rand.NewSource(seed))` usage only, no global or time-based seeding, and tests confirming same seed ⇒ same output. Do not restate the guideline here; treat it as the single source of truth. |
| **Network interfaces** | Variables declared as `net.Addr` / `net.PacketConn` / `net.Conn` / `net.Listener` — never `*net.UDPConn`, `*net.TCPConn`, `*net.UDPAddr`, `*net.TCPAddr`, `*net.IPAddr`; no type assertions or type switches to concrete `net.*` types; verify mock-ability in tests |
| **Error handling** | No swallowed errors (`_ = someFunc()` where error matters); all returned errors checked; no bare `fmt.Println` or `log.Println` on error paths — use `logrus.WithFields(logrus.Fields{...}).Error(...)` with standard field names (`entityID`, `system_name`, `seed`, `playerID`, `component_type`); `errors.Wrap` or `fmt.Errorf("context: %w", err)` for error chain preservation |
| **Concurrency safety** | Shared mutable state protected by `sync.Mutex` / `sync.RWMutex` or channels; no data races in system `Update()` (systems should not write to entities another system is reading in the same tick); `go vet -race` clean |
| **Test coverage** | Run `go test -cover -count=1 ./path/to/pkg/...`; flag if below 40% target (30% for X11/Wayland/Ebiten-dependent packages); note missing table-driven tests; note missing benchmarks for hot-path code (rendering, collision, spatial partition, packet serialization); verify `StubInput`/`StubSprite` used where Ebiten runtime is unavailable |
| **Doc coverage** | All exported types, functions, methods, and constants have godoc comments; package has `doc.go` with package-level overview; complex algorithms have inline comments explaining approach |
| **API consistency** | Constructor functions follow `NewXxx(params) *Xxx` pattern; system constructors log creation with `system_name` field; generator functions accept `seed int64` as first meaningful parameter; `Validate()` exists alongside `Generate()` |
| **Resource management** | Images and audio buffers released when no longer needed or pooled (`pkg/rendering/pool/`, `pkg/rendering/cache/`); no goroutine leaks (all spawned goroutines have shutdown path); file handles closed in defer; context.Context used for cancellation where appropriate |

---

## Phase 5: Integration Point Verification

| Integration Point | What to verify |
|---|---|
| **System registration** | System is registered in `system_init.go`, `cmd/client/` lazy-init, or `cmd/server/` system setup; system update order is correct relative to dependencies (e.g., input before movement, movement before collision, collision before rendering) |
| **Component registration** | All components the package defines/uses are registered in ECS world; component `Type()` strings are unique and not colliding with other packages |
| **Serialize / Deserialize** | Persistent components implement `Serialize() ([]byte, error)` and `Deserialize(data []byte) error`; format is versioned or migratable via `pkg/saveload/migrator`; WASM storage path tested if component is saved client-side |
| **Network sync** | Components that replicate across network are listed in snapshot system; delta compression handles them; client-side prediction is aware of them; desync detection covers them |
| **Genre theming** | If package generates content, verify it reads `GenreID` from `GenerationParams` and adapts output; genre blending supported if applicable (`pkg/procgen/genre/`) |
| **Event bus / messaging** | Package emits and/or listens to events via the engine event system; events are typed and documented; no tight coupling via direct function calls where events are more appropriate |
| **Mod compatibility** | If package defines data that mods can override (items, recipes, stats, dialog), verify mod loader (`pkg/modding/`) can inject overrides; sandboxing prevents executable code |
| **Accessibility** | UI elements have sufficient color contrast; text sizes respect quality/accessibility settings; screen-reader hints present if applicable; colorblind-safe palette option honored |

---

## Phase 6: Platform-Specific Checks

| Platform | What to verify |
|---|---|
| **Desktop (Linux/macOS/Windows)** | No platform-specific imports without build tags; keybindings assume standard keyboard layout; mouse coordinates in screen space; windowed/fullscreen toggle works |
| **WASM (Browser)** | No `os.Exit`, no filesystem writes outside `pkg/saveload/` WASM storage; `syscall/js` calls guarded by `//go:build js`; virtual controls shown; WebRTC signaling path reachable |
| **Mobile (iOS/Android)** | Touch input primary; no hover-dependent UI; `pkg/mobile/` dual-joystick routed through `Input` interface; screen rotation handled; safe area insets respected |
| **Build tags** | Files with `_wasm.go`, `_mobile.go`, `_desktop.go` suffixes or `//go:build` constraints compile cleanly on all target platforms; `go vet` passes with `GOOS=js GOARCH=wasm` |

---

## Phase 7: Run Automated Checks

Execute these commands and record results:

```bash
# 1. Vet the package
go vet ./path/to/pkg/...

# 2. Test with coverage
go test -cover -count=1 -timeout 120s ./path/to/pkg/...

# 3. Check for race conditions (if tests exist)
go test -race -count=1 -timeout 120s ./path/to/pkg/...

# 4. WASM vet (if package has WASM relevance)
GOOS=js GOARCH=wasm go vet ./path/to/pkg/...

# 5. Search for common anti-patterns
grep -rn 'TODO\|FIXME\|HACK\|PLACEHOLDER\|XXX\|not implemented' ./path/to/pkg/
grep -rn '\<rand\.\(Intn\|Float\|Int31\|Seed\)' ./path/to/pkg/ | grep -v '\<rand\.New\('
grep -rn 'time\.Now' ./path/to/pkg/
grep -rn 'net\.UDPConn\|net\.TCPConn\|net\.UDPAddr\|net\.TCPAddr\|net\.IPAddr' ./path/to/pkg/
grep -rn 'fmt\.Print\|log\.Print\|log\.Fatal' ./path/to/pkg/ | grep -v '_test\.go'
```

---

## Phase 8: Output Files

### 1. Create `<package-dir>/AUDIT.md`

Use this exact template:

```markdown
# Audit: <package-import-path>
**Date**: YYYY-MM-DD (ISO 8601, e.g., 2026-02-21)
**Auditor**: GitHub Copilot (META_AUDIT v2)
**Status**: Complete | Incomplete | Needs Work
<!--
Status criteria:
- Complete: All automated checks passed and fewer than 5 non-critical issues identified.
- Incomplete: Audit was stopped early or one or more required checks (e.g., go test, go vet, race) were not run.
- Needs Work: 5 or more issues identified, or any critical/priority-0 failure (e.g., panics, data corruption, security issues).
-->

## Summary
<2-3 sentences: scope, overall health, critical risk.>

## Automated Check Results
| Check | Result |
|---|---|
| `go vet` | ✅ Pass / ❌ Fail (details) |
| `go test -cover` | XX.X% (target: 40%, or 30% for X11/Wayland/Ebiten-dependent packages) |
| `go test -race` | ✅ Pass / ❌ Fail / ⚠️ No tests |
| WASM vet | ✅ Pass / ❌ Fail / N/A |
| TODO/FIXME count | N |
| Non-deterministic rand | N occurrences |
| Concrete net types | N occurrences |

## Issues Found

### High Severity
- [ ] **<category>** — <description> (`file.go:LINE`)

### Medium Severity
- [ ] **<category>** — <description> (`file.go:LINE`)

### Low Severity
- [ ] **<category>** — <description> (`file.go:LINE`)

## Input Integration
| Input Source | Status | Notes |
|---|---|---|
| Keyboard | ✅/❌/N/A | <notes> |
| Mouse | ✅/❌/N/A | <notes> |
| Gamepad | ✅/❌/N/A | <notes> |
| Touch | ✅/❌/N/A | <notes> |
| VR | ✅/❌/N/A | <notes> |
| Stub/Test | ✅/❌/N/A | <notes> |

## Menu/UI Integration
| Menu | Reachable | Input-Complete | Backing System Wired | Notes |
|---|---|---|---|---|
| <Menu Name> | ✅/❌ | ✅/❌ | ✅/❌ | <notes> |

## Test Coverage
**Coverage**: XX.X% (target: 40%, or 30% for X11/Wayland/Ebiten-dependent packages)
- Missing test areas: <list>
- Missing benchmarks: <list>
- Table-driven test compliance: ✅/❌

## Documentation Coverage
- Package `doc.go`: ✅/❌
- Exported symbols documented: XX/YY (ZZ%)
- Complex algorithms commented: ✅/❌

## Integration Status
<How this package connects to engine, client, server.>
- System registration: ✅/❌ — <details>
- Component registration: ✅/❌ — <details>
- Serialize/Deserialize: ✅/❌/N/A — <details>
- Network sync: ✅/❌/N/A — <details>
- Genre theming: ✅/❌/N/A — <details>
- Mod compatibility: ✅/❌/N/A — <details>

## Platform Status
| Platform | Status | Notes |
|---|---|---|
| Desktop | ✅/❌/N/A | |
| WASM | ✅/❌/N/A | |
| Mobile | ✅/❌/N/A | |

## Recommendations
1. **[HIGH]** <highest-priority fix>
2. **[HIGH]** <next high-priority fix>
3. **[MED]** <medium-priority fix>
4. **[LOW]** <low-priority improvement>
```

### 2. Update root `AUDIT.md`

Change the matching unchecked line to checked and append status:

```markdown
- [x] `path/to/AUDIT.md` — <Status> — <N> issues (<H> high, <M> med, <L> low) — Coverage: XX.X%
```

If the package is not yet listed, append a new checked entry.

---

## Phase 9: Final Report (print to chat)

After writing files, print:

1. **Package audited**: import path and rationale
2. **Path to created AUDIT.md**
3. **`go vet` result**: pass/fail
4. **Test coverage**: percentage and pass/fail vs. 40% target (30% for X11/Wayland/Ebiten-dependent packages)
5. **Top 5 critical findings**: each with `file.go:LINE`, severity, and category
6. **Input integration gaps**: any input source not properly wired
7. **Menu/UI integration gaps**: any menu unreachable, input-incomplete, or unwired
8. **Platform issues**: any platform-specific compilation or runtime concerns
9. **Recommended fix order**: prioritized list of what to address first

---

## Success Criteria

All of the following must be true for the audit to be considered complete:

- [ ] Sub-package `AUDIT.md` exists and follows the template exactly
- [ ] Every issue in `AUDIT.md` has a `file.go:LINE` citation
- [ ] Root `AUDIT.md` updated with checked entry including status, issue counts, and coverage
- [ ] `go vet` passes on audited package
- [ ] `go test` runs without failures (test failures are reported as high-severity issues)
- [ ] Input integration table is fully populated for all 6 input sources
- [ ] Menu/UI integration table is populated for all menus relevant to the package
- [x] Findings reference codebase-specific standards: ECS purity, deterministic seeds, interface networking, structured logging, `Input` interface abstraction [VERIFIED 2026-02-27: All AUDIT.md files reference these standards appropriately]
- [ ] Platform-specific checks completed (at minimum desktop and WASM vet)