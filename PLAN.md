# Game Repair Plan

## Critical Issues (Blocks Playability)

### Progression System — No XP Awarded on Enemy Kill
- [ ] Kill callback missing main character XP grant - Location: `pkg/engine/system_init.go:971-980` - Impact: The `SetKillCallback` only calls `weaponMasterySystem.OnKill()` and `classAffinitySystem.OnKill()` but never calls `progressionSystem.AwardXP()`. Players never gain experience or level up from combat. The entire progression system is disconnected from the core gameplay loop. `AwardXP()` is defined at `pkg/engine/progression_system.go:97` but only called from `pkg/engine/faction_xp_bonus_system.go:81` (for faction bonuses), never from the kill callback.

### GPU Memory Leak in Loading Screen
- [ ] `drawRect()` creates and leaks `ebiten.Image` objects every frame - Location: `pkg/engine/loading_ui.go:86-92` - Impact: `ebiten.NewImage()` is called 3 times per frame (lines 75, 76, 81) during the loading screen without ever calling `Dispose()`. Over a typical 5-30 second terrain generation, this leaks thousands of GPU-backed images. On low-memory devices (mobile, WASM), this can cause out-of-memory crashes before the game even starts. Fix: add `defer img.Dispose()` after line 87, or pool/reuse the rectangle images.

### Mobile Platform: CraftingSystem Initialized with nil ItemGenerator
- [ ] CraftingSystem created with nil itemGen on mobile path - Location: `pkg/engine/system_init.go:1168` - Impact: `InitializeGameSystems()` (used by `cmd/mobile/mobile.go:164`) creates the CraftingSystem with `nil` for itemGen: `NewCraftingSystem(game.World, inventorySystem, nil)`. Comment says "itemGen set later" but no subsequent code on the mobile path ever sets it, and the current implementation does not expose a setter for `itemGenerator`. The desktop client (`cmd/client/handlers.go:988`) correctly creates it with `sys.itemGen`. This means crafting on mobile will fail/crash when attempting to generate crafted items. Dependency: Requires passing a non-nil `item.ItemGenerator` into `InitializeGameSystems` so `NewCraftingSystem` is constructed correctly, or first adding an explicit `CraftingSystem` setter and updating the mobile path to call it.

## High Priority (Degrades Experience)

### Division by Zero Risk in Terrain Renderer
- [ ] No validation that tileWidth/tileHeight are non-zero before division - Location: `pkg/engine/terrain_render_system.go:109-112` - Impact: `NewTerrainRenderSystem()` (line 36) accepts `tileWidth, tileHeight int` without validating they are positive. Lines 109-112 divide viewport coordinates by `float64(t.tileWidth)` and `float64(t.tileHeight)`. If either is 0 (e.g., from misconfiguration or default zero-value struct), this produces `+Inf`/`-Inf`; converting those values to `int` is implementation-dependent and can yield invalid render bounds. In practice this is more likely to break tile culling, render nothing, or trigger a downstream panic than to cause an actual infinite loop. Fix: add `if t.tileWidth <= 0 || t.tileHeight <= 0 { return }` guard at start of `Draw()`.

### Scene Buffer Size Mismatch on Window Resize
- [ ] `litBuffer` not disposed/recreated in `Layout()` on resize - Location: `pkg/engine/game.go:1723-1742` - Impact: When the window is resized, `Layout()` properly disposes and recreates `sceneBuffer` (line 1733-1736) but does NOT handle `litBuffer`. The `litBuffer` is only recreated inside `drawLitScene()` (lines 1546-1551) when a size mismatch is detected. This means the first frame after resize renders lighting to a stale-sized buffer, causing visual clipping or misalignment. Fix: set `g.litBuffer = nil` in `Layout()` after screen dimension change, so `drawLitScene()` recreates it with correct dimensions.

### Voice Chat Network Transport Not Implemented
- [ ] Voice network packets never sent over network - Location: `pkg/audio/voice.go` (entire file), documented in `GAPS.md` Gap 1 - Impact: Voice codec (ADPCM) exists and can encode/decode audio, but the `VoiceTransport` interface that would send voice packets over the network is not implemented. `TODO(integration)` at `pkg/audio/manager.go:310` confirms this. Voice chat is completely non-functional in multiplayer.

### UI Visibility Checks Inconsistent for World Pause
- [ ] Different UI elements checked in different code paths - Location: `pkg/engine/game.go:1320,1354-1370` - Impact: `shouldUpdateWorld()` and the virtual controls visibility check (`anyUIOpen`) do not consistently check all open UI panels. Some UIs (e.g., QuestUI, ShopUI) may not properly pause world updates when opened, allowing enemies to attack the player while browsing menus. Related comments in the code say "BUG FIX" indicating this has been a recurring issue.

### Item Pickup Bypasses Inventory System
- [ ] Direct array append instead of using `InventorySystem.AddItemToInventory()` - Location: `pkg/engine/item_spawning.go:714-716` - Impact: `ItemPickupSystem.attemptItemPickup()` directly appends to `inventory.Items` instead of going through the normal inventory add path (`InventorySystem.AddItemToInventory()` / `InventoryComponent.AddItem`). Today this creates duplicated item-insertion logic and risks diverging behavior if capacity checks, stacking rules, logging, or error handling are centralized in the inventory add path. Result: pickups can behave inconsistently with other inventory additions and are harder to maintain safely.

## Medium Priority (Polish/Optimization)

### 43 Test Packages Panic Without Display Server
- [ ] Tests using `ebiten.NewImage()` panic without X11/GLFW context - Location: Multiple packages including `pkg/engine/`, `pkg/rendering/animation/`, `pkg/rendering/cache/`, `pkg/rendering/display/`, `pkg/rendering/lighting/`, `pkg/rendering/pool/`, `pkg/rendering/postprocess/`, `pkg/rendering/shapes/`, `pkg/rendering/sprites/`, `pkg/rendering/ui/`, `pkg/world/housing/`, `pkg/visualtest/` - Impact: These tests require `xvfb-run` to execute in CI. Tests panic with "GLFW library is not initialized" when run without a display. Tests that directly call `ebiten.NewImage()` should be refactored to use `StubImage`/`StubSprite` test doubles where possible, or properly guarded with build tags.

### FPS Benchmark Too Narrow
- [ ] FPS benchmark only tests 1 system out of 66+ - Location: `pkg/benchmark/fps/` (documented in `GAPS.md` Gap 5) - Impact: The FPS benchmark does not reflect real-world performance with all systems active. Performance regressions in other systems go undetected.

### BasicFont Only — No Scalable Text Rendering
- [ ] All text rendered with `basicfont.Face7x13` (7px wide, 13px tall) - Location: `pkg/engine/character_creation.go`, `pkg/engine/character_creation_tutorial.go`, `pkg/rendering/ui/chat.go`, `pkg/rendering/ui/notifications.go` - Impact: Text is very small on high-DPI displays and cannot scale. Font hierarchy in `pkg/rendering/ui/hierarchy.go` defines scale multipliers (0.8x-1.5x) but these are relative to the same tiny base font. This significantly degrades readability on modern displays (4K, Retina).

### MiniGameSystem Uses Custom Update Signature
- [ ] MiniGameSystem not added directly to World via standard System interface - Location: `pkg/engine/system_init.go:1178-1179` - Impact: `MiniGameSystem` has a custom `Update` signature that doesn't match the `System` interface. Both the client (`cmd/client/handlers.go:2150`) and server (`cmd/server/v4_systems.go:111`) work around this with a `miniGameSystemWrapper`. This is functional but adds complexity. The wrapper adapts the signature correctly, so this is a code quality concern, not a bug.

### Mobile Keyboard Integration Pending
- [ ] Native mobile keyboard APIs not integrated - Location: `pkg/mobile/keyboard_default.go:19` - Impact: `TODO: Integrate native mobile keyboard APIs (UIKeyboard on iOS, InputMethodManager on Android)`. Text input on mobile likely falls back to virtual keyboard which may not work correctly for all input scenarios (chat, character naming).

### Missing Gold Check for Respec
- [ ] Respec operation doesn't verify player has sufficient gold - Location: `cmd/client/handlers.go:3260` - Impact: `TODO: Check if player has enough gold for respec`. Players may be able to respec skills/attributes for free, bypassing intended gold sink.

## Resolution Notes

### Architecture Context
- The desktop client (`cmd/client/`) uses `setupAllGameSystems()` in `handlers.go` which is a different initialization path from `InitializeGameSystems()` in `pkg/engine/system_init.go` (used by mobile). Issues in `system_init.go` primarily affect mobile; issues in `handlers.go` affect desktop.
- Ebiten v2.9.3 automatically clears the screen before each `Draw()` call, so the absence of explicit `screen.Clear()` in `game.go:Draw()` is correct behavior, not a bug.
- The `SetTerrain()` chain IS properly wired in both desktop (`cmd/client/handlers.go:2320,2439,2545`) and mobile (`cmd/mobile/mobile.go:242,246`).
- The spatial quadtree IS properly wired to both collision and AI systems in the desktop client (`cmd/client/handlers.go`).
- `GenerateLootDrop()` IS called when enemies die (`cmd/client/util.go`), so loot drops are functional.

### Dependencies Between Fixes
1. **XP on Kill** (Critical #1) should be fixed first as it's the most impactful gameplay blocker — without it, no progression occurs.
2. **Loading UI Memory Leak** (Critical #2) should be fixed before extensive testing, as it can crash the game before reaching gameplay.
3. **Mobile CraftingSystem** (Critical #3) blocks mobile crafting but is independent of other fixes.
4. **Item Pickup bypass** (High #6) should be fixed alongside the XP fix, as both affect the core combat-reward gameplay loop.
5. **UI Visibility Checks** (High #5) should be audited comprehensively — a single pass through all UI panels to ensure consistent pause behavior.

### Verified Non-Issues (Initially Suspected)
- ~~Screen not cleared in Draw()~~ → Ebiten clears automatically; verified correct
- ~~SetTerrain never called~~ → Called in both handlers.go and mobile.go after generation
- ~~Quadtree not wired to AI/Collision~~ → Wired in cmd/client/handlers.go
- ~~Loot drops never generated~~ → Called from cmd/client/util.go
- ~~DiscoverySystem/PvPRatingSystem dangling~~ → Instantiated in cmd/server/v4_systems.go and cmd/client/init_versions.go
- ~~InventoryUI.inventorySystem never set~~ → Set via `game.SetInventorySystem()` at cmd/client/handlers.go:3243
