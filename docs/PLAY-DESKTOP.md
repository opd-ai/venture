# Venture Desktop Bug Audit & Remediation

## Objective
Identify and fix all gameplay-blocking bugs and critical defects in the Venture desktop builds (Linux/macOS/Windows) by methodically testing the complete player journey from launch to endgame, prioritizing issues that prevent normal gameplay progression.

## Execution Mode
**Autonomous Action** - Automatically detect and fix all discovered bugs without requiring approval.

## Platform Specifications

### Target Platforms
- **Linux**: Ubuntu 20.04+, Fedora 35+, Arch Linux (x64/ARM64)
- **macOS**: macOS 11.0+ Big Sur (x64/ARM64 Apple Silicon)
- **Windows**: Windows 10/11 (x64)

### Build Verification
```bash
# Linux/macOS/Windows
go build ./cmd/client
go build ./cmd/server

# Cross-platform builds
make build-linux      # Linux x64
make build-linux-arm  # Linux ARM64
make build-darwin     # macOS x64
make build-darwin-arm # macOS ARM64 (Apple Silicon)
make build-windows    # Windows x64
```

### Display Requirements
- **Default Resolution**: 1920x1080 (windowed mode)
- **Supported Resolutions**: 1280x720, 1920x1080, 2560x1440, 3840x2160
- **Multi-Monitor**: Primary monitor detection with manual selection support
- **Window Modes**: Windowed, fullscreen, borderless windowed

## Testing Methodology: Player Journey Order

### Phase 1: Launch & Main Menu (5 minutes)

1. **Application Launch**
   - Build verification: `go build ./cmd/client`
   - Launch succeeds without crashes
   - Window renders at correct resolution (1920x1080 default)
   - Window title displays "Venture" or "Venture v[version]"
   - Window icon renders correctly (if implemented)

2. **Main Menu Navigation**
   - All menu buttons visible and functional (New Game, Load Game, Settings, Exit)
   - **ESC key** returns to previous menu or exits application
   - Genre selection UI accessible and functional
   - Settings menu: graphics, audio, controls adjustable
   - Exit paths work from all submenu states
   - Mouse click and keyboard navigation (arrow keys + Enter) both functional

3. **Desktop-Specific Display Tests**
   - **Window Resize**: Drag window edges, verify UI scales correctly
   - **Multi-Monitor**: Move window between monitors, verify rendering persists
   - **High-DPI Scaling**: Test on 4K displays (3840x2160), verify UI readability
   - **Fullscreen Toggle**: Alt+Enter or F11 switches modes without artifacts
   - **Native Dialogs**: Zenity integration for file dialogs (Linux) works correctly

### Phase 2: Tutorial & First-Time Experience (5 minutes)

4. **Tutorial System**
   - Tutorial triggers on first launch or new game
   - All tutorial prompts display correctly
   - Tutorial controls functional (can skip, advance, close)
   - **ESC exits tutorial** without trapping player
   - Tutorial completion flag persists correctly in save data

5. **Initial Spawn**
   - Player entity spawns in walkable terrain
   - Camera centers on player
   - HUD elements render (health, mana, XP bar)
   - **Keyboard controls responsive**: WASD movement, mouse aim

### Phase 3: Core Gameplay Loop (10 minutes)

6. **Movement & Collision**
   - **WASD movement** in all 8 directions (W, WA, A, AS, S, SD, D, DW)
   - **Arrow keys** as alternative movement input
   - Collision detection prevents wall clipping
   - **Mouse aim** rotates player sprite (360° rotation system)
   - **Mouse cursor** displays correctly (crosshair or custom cursor)
   - Frame rate stable at 60+ FPS during movement

7. **Combat System**
   - **Space bar or Left Mouse Click** triggers attack animation
   - **Right Mouse Click** for secondary attack (if implemented)
   - Damage calculation applies to enemies
   - Player takes damage from enemy attacks
   - Health depletes correctly, death/revival system activates
   - Status effects apply and expire (visual indicators present)
   - **Number keys 1-5** trigger ability hotbar

8. **Inventory System**
   - **I key** opens inventory menu
   - Items display with sprites and stats
   - **Mouse drag-and-drop** or click-to-equip works
   - **ESC closes inventory** (dual-exit pattern)
   - Equipment slots update character stats in real-time
   - Inventory grid scrolls with mouse wheel

9. **Item Pickup & Loot**
   - **E key** interacts with chests/drops
   - Items added to inventory with pickup animation
   - Inventory full condition handled gracefully (notification displayed)
   - Loot rarity colors display correctly (Common/Uncommon/Rare/Epic/Legendary)
   - **Mouse hover** shows item tooltips

### Phase 4: Progression Systems (10 minutes)

10. **Character Sheet**
    - **C key** opens character menu
    - Stats display current values (STR, DEX, INT, VIT, etc.)
    - Level-up increases stats correctly
    - **ESC closes menu**
    - **Tab key** cycles between character sheet tabs (if implemented)

11. **Skills & Abilities**
    - **K key** opens skill tree menu
    - Skill nodes display with prerequisites (locked/unlocked visual states)
    - Skill points allocation works (mouse click to unlock)
    - **Number keys 1-5** trigger learned spells
    - Mana consumption and cooldowns function (visual indicators)
    - **Mouse hover** shows skill descriptions

12. **Quest System**
    - **J key** opens quest log
    - Active quests display objectives with progress tracking
    - Quest progress updates on completion
    - Rewards granted (XP, items, gold) with notification
    - Completed quests marked correctly (visual checkmark or greyed out)
    - **Mouse click** expands/collapses quest details

### Phase 5: World Interaction (5 minutes)

13. **NPC Interaction**
    - **F key** triggers merchant/NPC dialog
    - Dialog UI displays text correctly (readable font, no overflow)
    - Shop interface allows buy/sell (mouse click on items)
    - Prices scale by rarity
    - Transaction validation works (insufficient gold prevents purchase)
    - **ESC exits dialog** without locking player

14. **Crafting System**
    - **R key** opens crafting menu
    - Recipes display with requirements (materials + crafting station)
    - Item crafting consumes materials correctly
    - Crafted items added to inventory
    - Station proximity validation works (error message if too far)
    - **Mouse click** on recipe initiates crafting

15. **Map & Navigation**
    - **M key** opens map overlay
    - Explored areas visible, unexplored dark (fog of war)
    - Player position marker accurate and updates in real-time
    - **ESC closes map**
    - **Mouse drag** pans map (if implemented)
    - **Mouse wheel** zooms map (if implemented)

### Phase 6: Multiplayer (5 minutes)

16. **Connection & Sync**
    - Server starts: `./venture-server -port 8080`
    - Client connects: `./venture-client -multiplayer -server localhost:8080`
    - Player entities spawn for all clients
    - Movement synchronizes across clients (tested with 2+ instances)
    - Combat damage replicates correctly
    - Network latency display (ping indicator) accurate

17. **Social Systems (V5+)**
    - **Enter key** or **/chat** opens chat input
    - Text messages send/receive in chat window
    - Trade system initiates (if implemented)
    - Player names display above entities

### Phase 7: Save/Load (5 minutes)

18. **Save System**
    - **F5 quick save** triggers without errors
    - Save file created in `saves/` directory (verify file exists)
    - Save includes all state (position, inventory, quests, skills)
    - Multiple save slots supported (named saves)
    - Save timestamp displays correctly

19. **Load System**
    - **F9 quick load** restores saved state
    - Load Game menu lists save files with metadata (date, time, level)
    - Load preserves world seed (deterministic generation verified)
    - No state corruption on load (inventory, quests, skills intact)
    - **ESC exits load menu** without loading

## Desktop-Specific Tests

### Window Management
- **Minimize/Restore**: Window minimizes to taskbar, restores without black screen
- **Focus Loss**: Alt+Tab away and back, game pauses/resumes correctly
- **System Sleep**: Computer sleep/wake cycle preserves game state
- **Screen Lock**: Lock screen and unlock, game resumes without crash

### Input Edge Cases
- **Keyboard Shortcuts**: Ctrl+C/V/X in text fields (if implemented)
- **Alt+F4**: Closes application cleanly (saves if auto-save enabled)
- **Windows Key**: Game pauses when Start menu opens (Windows)
- **Cmd+Q**: Quits application cleanly (macOS)
- **Multi-Key Combos**: WASD + Space + Shift simultaneous input works

### Performance
- **CPU Usage**: <30% on mid-range desktop (Intel i5/Ryzen 5)
- **Memory Usage**: <500MB (achieved 73MB in current builds)
- **Frame Rate**: 60+ FPS sustained (achieved 106 FPS with 2000 entities)
- **Loading Times**: <2s for area generation

### Platform-Specific
- **Linux**: X11 and Wayland compatibility, no Xorg crashes
- **macOS**: Retina display support, menu bar integration, Gatekeeper approval
- **Windows**: UAC prompts (if required), no SmartScreen warnings for signed builds

## Bug Categories & Priority

### Critical (P1 - Gameplay Blockers)
- **Menu Traps:** UI states where ESC fails to exit
- **Broken Controls:** Non-functional keybinds (WASD, Space, E, F, etc.)
- **Progression Blockers:** Infinite loops, deadlocks, unexitable states
- **Missing Menus:** Advertised features without accessible UI
- **Spawn Failures:** Player spawns in walls or out-of-bounds
- **Window Crashes:** Black screens on resize, minimize, or multi-monitor move

### High Priority (P2 - Visual/UX Defects)
- **Rendering Failures:** Black screens, missing sprites, broken animations
- **UI Layout Issues:** Overlapping elements, off-screen buttons, unreadable text
- **Input Edge Cases:** Multi-key conflicts, mouse button failures
- **Audio Failures:** Missing SFX, music playback errors, no audio on specific platforms
- **High-DPI Bugs:** Blurry UI on 4K displays, incorrect scaling

### Medium Priority (P3 - System Bugs)
- **Memory Leaks:** Unbounded cache growth during long sessions
- **Race Conditions:** Detected by `go test -race`
- **Save Corruption:** State persistence failures across sessions
- **Network Edge Cases:** Packet loss, reconnection failures, firewall blocks
- **Performance Degradation:** FPS drops over time, memory bloat

## Automated Detection

### Static Analysis
```bash
go vet ./...                           # Suspicious constructs
go test -race ./pkg/engine/... ./pkg/rendering/ui/...  # Race conditions
golangci-lint run                      # Comprehensive linting (optional)
```

### Pattern-Based Search
```bash
grep -rn "panic\|TODO.*block\|FIXME.*critical" pkg/ cmd/
grep -rn "for \{" pkg/engine/ pkg/rendering/     # Infinite loops
grep -rn "time\.Sleep.*second" pkg/engine/        # Blocking in game loop
grep -rn "ebiten\.NewImage" pkg/engine/ pkg/procgen/  # Ebiten calls in non-rendering code
```

### Desktop-Specific Checks
```bash
# Check for platform-specific imports
grep -rn "golang.org/x/sys/windows" pkg/
grep -rn "github.com/ncruces/zenity" pkg/  # Linux dialog library

# Verify window management
grep -rn "ebiten\.SetWindowSize\|SetFullscreen\|SetWindowTitle" cmd/client/
```

### ECS System Verification
- Check `cmd/client/main.go` for all system registrations
- Verify systems in execution order (Input → Movement → Combat → Render)
- Confirm component serialization for multiplayer sync

## Fix Requirements

### Code Quality
- Maintain ECS architecture (no logic in components)
- Preserve deterministic generation (seed-based RNG only)
- Follow dual-exit UI pattern (ESC + back button/mouse click)
- Ensure ≥40% test coverage after fixes (≥30% for packages depending on X11/Wayland/Ebiten)
- Pass `go test ./...` without errors

### Fix Documentation
Add inline comment for each fix:
```go
// BUG FIX: [Phase] - [Issue Description]
// Resolution: [What changed and why]
// Platform: Desktop (Linux/macOS/Windows)
// Example:
// BUG FIX: Phase 3.7 - Inventory ESC key not bound to Hide()
// Resolution: Added inpututil.IsKeyJustPressed(ebiten.KeyEscape) check in Update()
// Platform: Desktop (all)
```

## Success Criteria

- ✅ Complete player journey tested (Phase 1-7) without blocks on all desktop platforms
- ✅ All builds compile: `go build ./cmd/client ./cmd/server`
- ✅ Full test suite passes: `go test ./...`
- ✅ No race conditions: `go test -race ./...`
- ✅ All UI menus have functional exit paths (ESC + close button)
- ✅ Core gameplay loop playable: launch → tutorial → combat → loot → progression → save/load
- ✅ Multiplayer connects and synchronizes (2+ clients on localhost and LAN)
- ✅ Window management robust (resize, minimize, multi-monitor, fullscreen)
- ✅ High-DPI displays render correctly without blurring
- ✅ Platform-specific features work (Zenity dialogs on Linux, Retina on macOS, UAC on Windows)

## Constraints

- **Do Not Break:** Deterministic generation (same seed = same world)
- **Do Not Modify:** Public API signatures without version bump justification
- **Do Not Skip:** Test validation for each fix on all three platforms
- **Do Not Add:** New features or refactoring beyond bug fixes
- **Do Not Hardcode:** Platform paths or display settings (use runtime detection)

## Execution Order

1. Fix Phase 1 (Launch & Main Menu) blockers on all platforms
2. Fix Phase 2 (Tutorial) blockers
3. Fix Phase 3 (Core Gameplay) blockers
4. Fix Phase 4 (Progression) issues
5. Fix Phase 5 (World Interaction) issues
6. Fix Phase 6 (Multiplayer) issues
7. Fix Phase 7 (Save/Load) issues
8. Fix Desktop-Specific Tests (window management, high-DPI, platform quirks)
9. Validate all fixes with test suite on Linux, macOS, and Windows
10. Report summary of bugs found and fixed by phase and platform

## Cross-References

- **Build Process**: See `docs/CROSS_PLATFORM_BUILDS.md` for detailed build instructions
- **Development Setup**: See `docs/DEVELOPMENT.md` for environment configuration
- **Performance Optimization**: See `docs/PERFORMANCE.md` for profiling and optimization
- **Architecture**: See `docs/ARCHITECTURE.md` for ECS system design
