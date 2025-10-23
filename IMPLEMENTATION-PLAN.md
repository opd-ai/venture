# Implementation Action Plan - Phase 8 Completion
Generated: 2025-10-22  
Based on: GAPS-AUDIT.md  
Target: True Beta Release Status

## Overview
This document provides a prioritized, actionable plan to address the 12 implementation gaps identified in the audit. Each task includes specific files to create/modify, implementation approach, and estimated effort.

---

## Critical Path (Blockers for Beta) - 10-12 days

### Task 1.1: Implement Network Server TCP Listener
**Priority:** CRITICAL (Gap #1, Score: 162.5)  
**Effort:** 2-3 days  
**Dependencies:** None

**Files to Create:**
- `pkg/network/server.go` - TCP server implementation
- `pkg/network/client_connection.go` - Per-client connection handler
- `pkg/network/protocol.go` - Network protocol message types

**Files to Modify:**
- `cmd/server/main.go` - Add actual TCP listener and client handling

**Implementation Steps:**
1. Create `Server` struct with `net.Listener` for TCP connections
2. Implement `AcceptClients()` goroutine to handle incoming connections
3. Add client authentication handshake (player name, version check)
4. Implement state broadcasting at configured tick rate (20 Hz)
5. Add graceful shutdown handling (cleanup connections)
6. Update server logging to show actual connection status

**Testing:**
- Unit tests for protocol serialization/deserialization
- Integration test: start server, connect client, verify handshake
- Load test: connect 4 clients simultaneously
- Network test: verify state broadcast frequency

**Acceptance Criteria:**
- Server listens on configured port
- Clients can connect and authenticate
- Server broadcasts world state at 20 Hz
- `netstat` shows port in LISTEN state
- No "stub" warnings in logs

---

### Task 1.2: Build Inventory UI System
**Priority:** HIGH (Gap #8, Score: 40.0)  
**Effort:** 2 days  
**Dependencies:** Task 2.1 (keyboard shortcuts)

**Files to Create:**
- `pkg/engine/inventory_ui.go` - Inventory rendering and interaction
- `pkg/rendering/ui/inventory_window.go` - Inventory window widget

**Files to Modify:**
- `pkg/engine/game.go` - Add inventory UI rendering
- `pkg/engine/input_system.go` - Add 'I' key handling

**Implementation Steps:**
1. Create `InventoryUI` struct with grid layout (8x4 item slots)
2. Implement item icon rendering using procedural sprites
3. Add equipment slots display (weapon, armor, accessory)
4. Implement item tooltip on hover (stats, description, value)
5. Add drag-and-drop functionality for item management
6. Implement use/equip/drop actions via right-click or hotkeys
7. Add gold and weight capacity display bars
8. Connect to existing `InventorySystem` backend

**Testing:**
- Test item grid rendering with various item counts
- Test drag-and-drop between slots
- Test equip/unequip actions
- Test tooltip display
- Verify weight limit enforcement

**Acceptance Criteria:**
- 'I' key opens/closes inventory window
- All inventory items displayed in grid
- Equipment slots show equipped items
- Can drag-drop items to reorder/equip
- Tooltips show item details on hover

---

### Task 1.3: Build Quest Tracking and UI System
**Priority:** HIGH (Gap #9, Score: 38.5)  
**Effort:** 2-3 days  
**Dependencies:** Task 2.1 (keyboard shortcuts)

**Files to Create:**
- `pkg/engine/quest_tracker.go` - Quest state tracking (active/completed)
- `pkg/engine/quest_ui.go` - Quest log UI rendering
- `pkg/rendering/ui/quest_window.go` - Quest window widget

**Files to Modify:**
- `pkg/engine/game.go` - Add quest tracking and UI rendering
- `pkg/engine/input_system.go` - Add 'J' key handling
- `pkg/procgen/quest/generator.go` - Add integration helpers

**Implementation Steps:**
1. Create `QuestTracker` component to store active/completed quests
2. Implement quest acceptance system (triggered by NPC interaction)
3. Add objective progress tracking (kill counts, item collection, etc.)
4. Create quest log UI with tabs: Active / Completed
5. Implement quest detail view (description, objectives, rewards)
6. Add quest notification system (new quest, objective complete, quest complete)
7. Integrate with NPC system for quest givers
8. Add quest marker rendering on HUD (optional)

**Testing:**
- Test quest acceptance and tracking
- Test objective progress updates
- Test quest completion and rewards
- Test UI navigation between active/completed tabs
- Verify notifications appear correctly

**Acceptance Criteria:**
- 'J' key opens quest log
- Active quests show progress bars
- Completed quests archived separately
- Notifications appear for quest events
- Quest objectives update in real-time

---

### Task 1.4: Complete Menu System Implementation
**Priority:** HIGH (Gap #3, Score: 67.8)  
**Effort:** 2 days  
**Dependencies:** Tasks 1.2, 1.3 (inventory and quest UIs)

**Files to Modify:**
- `pkg/engine/menu_system.go` - Add Update() and Draw() methods
- `pkg/engine/game.go` - Integrate menu system rendering
- `pkg/engine/input_system.go` - Update ESC key handling

**Implementation Steps:**
1. Implement `Update()` method for menu navigation
2. Implement `Draw()` method for menu rendering
3. Add menu items: Resume, Inventory, Character, Skills, Quests, Settings, Save, Load, Quit
4. Create sub-menus for Settings (graphics, audio, controls)
5. Add settings persistence (write to `config.json`)
6. Integrate with existing Save/Load functionality
7. Add character stats screen
8. Add skills tree visualization screen
9. Update ESC key to open menu (not just help)

**Testing:**
- Test menu navigation with keyboard
- Test each menu action
- Test settings persistence
- Test menu state stack (nested menus)
- Verify ESC key behavior (close menu, back to previous)

**Acceptance Criteria:**
- ESC opens game menu
- All menu items functional
- Can access inventory/character/skills/quests from menu
- Settings saved to `config.json`
- Menu navigation smooth and responsive

---

## High-Value Polish - 5-7 days

### Task 2.1: Implement All Keyboard Shortcuts
**Priority:** MEDIUM (Gap #4, Score: 52.5)  
**Effort:** 1 day  
**Dependencies:** Tasks 1.2, 1.3, 1.5 (UIs must exist first)

**Files to Modify:**
- `pkg/engine/input_system.go` - Add missing key bindings

**Implementation Steps:**
1. Add `KeyInventory` (I), `KeyCharacter` (C), `KeySkills` (K)
2. Add `KeyQuests` (J), `KeyMap` (M), `KeyCycleTargets` (Tab)
3. Implement key handling in `Update()` method
4. Connect keys to respective UI systems
5. Implement target cycling system for Tab key
6. Update help system with new shortcuts

**Testing:**
- Test each shortcut key
- Verify no key conflicts
- Test target cycling functionality

**Acceptance Criteria:**
- All 12 documented shortcuts work
- Shortcuts open correct UIs
- Tab cycles through nearby enemies

---

### Task 2.2: Integrate Audio System into Game Loop
**Priority:** MEDIUM (Gap #6, Score: 45.0)  
**Effort:** 2-3 days  
**Dependencies:** None

**Files to Create:**
- `pkg/engine/audio_system.go` - Audio playback system
- `pkg/audio/player.go` - Ebiten audio player wrapper

**Files to Modify:**
- `cmd/client/main.go` - Initialize audio system
- `pkg/engine/combat_system.go` - Trigger attack/damage sounds
- `pkg/engine/inventory_system.go` - Trigger item pickup sounds
- `pkg/engine/progression_system.go` - Trigger level up sounds

**Implementation Steps:**
1. Create `AudioSystem` with Ebiten audio context
2. Implement music player with looping and crossfade
3. Implement SFX player with multiple channels
4. Generate background music based on genre and context (combat/exploration)
5. Add SFX triggers for: attacks, damage, item pickup, level up, quest complete
6. Implement volume controls (music/SFX separate)
7. Add audio mute option
8. Ensure deterministic generation (same seed = same music)

**Testing:**
- Test music playback and looping
- Test SFX triggering for various events
- Test volume controls
- Verify no audio glitches or clicking
- Test determinism (same seed = same audio)

**Acceptance Criteria:**
- Background music plays continuously
- Genre-appropriate music themes
- SFX trigger for all major actions
- Volume controls work
- Audio setting persists in config

---

### Task 2.3: Integrate Particle Effects into Gameplay
**Priority:** MEDIUM (Gap #7, Score: 42.0)  
**Effort:** 2 days  
**Dependencies:** None

**Files to Create:**
- `pkg/engine/particle_system.go` - Particle spawning and management

**Files to Modify:**
- `pkg/engine/combat_system.go` - Spawn particles for attacks/damage
- `pkg/engine/inventory_system.go` - Spawn particles for item pickup
- `pkg/engine/progression_system.go` - Spawn particles for level up
- `pkg/engine/game.go` - Render particle system

**Implementation Steps:**
1. Create `ParticleSystem` that manages active particle emitters
2. Integrate with `pkg/rendering/particles/` generators
3. Add particle spawn triggers for: melee hit, projectile hit, spell cast, critical hit
4. Add particles for: item pickup (sparkles), level up (burst), death (smoke)
5. Implement particle rendering in game draw loop
6. Add particle pooling for performance
7. Ensure genre-appropriate particle colors

**Testing:**
- Test particle spawning for each trigger
- Test particle rendering performance (1000+ particles)
- Verify particles follow entities correctly
- Test particle cleanup (no memory leaks)

**Acceptance Criteria:**
- Particles appear for combat actions
- Particles appear for item interactions
- Level up creates visual burst effect
- Particles are genre-appropriate colors
- No performance impact (<1ms per frame)

---

### Task 2.4: Implement Developer Console System
**Priority:** MEDIUM (Gap #2, Score: 89.6)  
**Effort:** 2-3 days  
**Dependencies:** None (but useful for all other tasks)

**Files to Create:**
- `pkg/engine/console.go` - Console UI and command parser
- `pkg/engine/console_commands.go` - Command implementations
- `pkg/engine/achievements.go` - Achievement tracking (optional)

**Files to Modify:**
- `pkg/engine/input_system.go` - Add tilde (~) key handling
- `pkg/engine/game.go` - Integrate console rendering

**Implementation Steps:**
1. Create `ConsoleSystem` with text input overlay
2. Implement command parser (slash-prefix commands)
3. Implement commands: `/tp X Y`, `/give ITEM_ID`, `/level N`
4. Implement commands: `/god`, `/noclip`, `/spawn ENTITY_TYPE`
5. Implement command: `/kill_all`
6. Add command history (up/down arrows)
7. Add auto-complete for commands
8. Add "cheats enabled" flag that disables achievements
9. Implement basic achievement system (optional)

**Testing:**
- Test each console command
- Test command parsing with various inputs
- Test error handling for invalid commands
- Verify command history works
- Test cheat detection flag

**Acceptance Criteria:**
- Tilde (~) opens/closes console
- All 7 documented commands work
- Commands show clear error messages
- Command history accessible
- Console overlay doesn't block gameplay view

---

## Quality of Life - 3-4 days

### Task 3.1: Implement Map System
**Priority:** LOW (Gap #10, Score: 35.0)  
**Effort:** 2 days  
**Dependencies:** Task 2.1 (M key shortcut)

**Files to Create:**
- `pkg/engine/map_system.go` - Map rendering and fog of war
- `pkg/rendering/ui/map_window.go` - Map window widget

**Files to Modify:**
- `pkg/engine/game.go` - Add map rendering
- `pkg/engine/input_system.go` - Add M key handling (already in 2.1)

**Implementation Steps:**
1. Create `MapSystem` that tracks explored tiles
2. Implement fog of war (unexplored/explored/visible)
3. Create minimap renderer (top-right corner of HUD)
4. Create full map overlay (M key)
5. Add player position indicator
6. Add entity markers (enemies as red dots, NPCs as blue)
7. Add quest objective markers (yellow stars)
8. Implement zoom controls for full map

**Testing:**
- Test fog of war updates as player moves
- Test minimap accuracy
- Test full map overlay
- Verify markers show correct positions

**Acceptance Criteria:**
- Minimap always visible in corner
- M key opens full-screen map
- Fog of war updates correctly
- Entity markers accurate
- Map updates in real-time

---

### Task 3.2: Implement Configuration Persistence
**Priority:** LOW (Gap #5, Score: 48.0)  
**Effort:** 1 day  
**Dependencies:** Task 1.4 (menu system with settings)

**Files to Create:**
- `pkg/engine/config.go` - Configuration management

**Files to Modify:**
- `cmd/client/main.go` - Load config on startup
- `pkg/engine/menu_system.go` - Save config when settings change

**Implementation Steps:**
1. Create `Config` struct with all settings
2. Implement `LoadConfig()` from `./config.json`
3. Implement `SaveConfig()` to `./config.json`
4. Add default config values
5. Merge command-line flags with config file
6. Settings include: resolution, fullscreen, vsync, music volume, SFX volume, key bindings

**Testing:**
- Test config loading on startup
- Test config saving from menu
- Test default config creation
- Verify command-line flags override config

**Acceptance Criteria:**
- Config file created on first run
- Settings persist across restarts
- Config file is human-readable JSON
- Command-line flags take precedence

---

### Task 3.3: Implement Structured Logging
**Priority:** LOW (Gap #5, Score: 48.0)  
**Effort:** 0.5 days  
**Dependencies:** None

**Files to Create:**
- `pkg/engine/logger.go` - Logging system

**Files to Modify:**
- `cmd/client/main.go` - Initialize logger
- `cmd/server/main.go` - Initialize logger
- Replace all `log.Printf()` calls with structured logger

**Implementation Steps:**
1. Create `Logger` struct with file output
2. Implement log levels (DEBUG, INFO, WARN, ERROR)
3. Create `./logs/` directory on startup
4. Write logs to `./logs/venture.log`
5. Implement log rotation (max 10MB, keep 5 files)
6. Add structured fields (timestamp, level, component, message)

**Testing:**
- Verify log file created
- Test log rotation
- Verify all log levels work

**Acceptance Criteria:**
- Logs written to `./logs/venture.log`
- Logs include timestamps and levels
- Log rotation works correctly
- Console output still shows logs

---

### Task 3.4: Implement Screenshot Feature
**Priority:** LOW (Gap #5, Score: 48.0)  
**Effort:** 0.5 days  
**Dependencies:** Task 2.1 (add key binding)

**Files to Create:**
- `pkg/engine/screenshot.go` - Screenshot system

**Files to Modify:**
- `pkg/engine/input_system.go` - Add F12 key for screenshots
- `pkg/engine/game.go` - Integrate screenshot capture

**Implementation Steps:**
1. Create `./screenshots/` directory on startup
2. Implement `CaptureScreen()` using Ebiten image capture
3. Save as PNG with timestamp filename
4. Add F12 key binding
5. Show notification when screenshot saved

**Testing:**
- Test screenshot capture
- Verify PNG format
- Verify directory creation

**Acceptance Criteria:**
- F12 captures screenshot
- Screenshots saved to `./screenshots/`
- Filename includes timestamp
- Notification shows save location

---

## Minor Fixes - 1 day

### Task 4.1: Fix Server Log Messages
**Priority:** LOW (Gap #11, Score: 32.0)  
**Effort:** 0.25 days  
**Dependencies:** Task 1.1 (network server implementation)

**Files to Modify:**
- `cmd/server/main.go` - Update logging after network implementation

**Implementation Steps:**
1. Remove "(not accepting connections yet - network layer stub)" message
2. Add clear "Server listening on port X" message
3. Log successful client connections
4. Add error logging for port binding failures

**Acceptance Criteria:**
- Server logs accurately reflect network state
- No misleading "stub" messages
- Connection events logged

---

### Task 4.2: Update Documentation
**Priority:** LOW (Gap #12, Score: 24.0)  
**Effort:** 0.5 days  
**Dependencies:** All other tasks (document what's actually implemented)

**Files to Modify:**
- `docs/USER_MANUAL.md` - Remove achievement references or mark as planned
- `README.md` - Update feature status
- `docs/ROADMAP.md` - Update phase completion status

**Implementation Steps:**
1. Review all documentation for accuracy
2. Remove or mark "planned" for achievements
3. Update feature lists to match implementation
4. Add notes about which features are complete vs. planned
5. Update keyboard shortcuts table
6. Update file locations section

**Acceptance Criteria:**
- Documentation matches implementation
- No false claims about features
- Phase status accurate
- User Manual accurate

---

## Implementation Order

### Week 1: Core Networking and UIs
- Day 1-3: Task 1.1 (Network Server) - CRITICAL
- Day 4-5: Task 1.2 (Inventory UI)

### Week 2: Quest and Menu Systems
- Day 6-8: Task 1.3 (Quest Tracking & UI)
- Day 9-10: Task 1.4 (Complete Menu System)

### Week 3: Polish and Integration
- Day 11: Task 2.1 (Keyboard Shortcuts)
- Day 12-14: Task 2.2 (Audio Integration)
- Day 15-16: Task 2.3 (Particle Integration)

### Week 4: Developer Tools and QoL
- Day 17-19: Task 2.4 (Console System)
- Day 20-21: Task 3.1 (Map System)
- Day 22: Tasks 3.2, 3.3, 3.4 (Config, Logging, Screenshots)

### Week 5: Final Polish
- Day 23: Task 4.1, 4.2 (Fixes and Documentation)
- Day 24-25: Integration testing and bug fixes

---

## Testing Strategy

### Unit Tests (Ongoing)
- Test each new system in isolation
- Maintain 80%+ coverage target
- Use `-tags test` flag for headless testing

### Integration Tests (After Each Task)
- Test system integration with existing code
- Verify no regressions in existing features
- Test multiplayer scenarios (for networking tasks)

### Performance Tests (End of Week)
- Profile after major changes
- Verify 60 FPS target maintained
- Check memory usage (<500MB)
- Measure network bandwidth (<100KB/s per player)

### User Acceptance Tests (End of Implementation)
- Complete gameplay session from start to save/load
- Test all keyboard shortcuts
- Test multiplayer with 4 clients
- Test all menu options
- Verify documentation accuracy

---

## Success Metrics

### Feature Completion
- [ ] All 12 documented gaps resolved
- [ ] All keyboard shortcuts functional
- [ ] All menu options working
- [ ] Network server accepts connections
- [ ] Audio plays during gameplay
- [ ] Particle effects visible
- [ ] Developer console functional

### Quality Metrics
- [ ] Test coverage ≥ 80% for all packages
- [ ] Network package coverage ≥ 75% (improved from 66.8%)
- [ ] Performance: 60 FPS with 2000 entities
- [ ] Memory: <500MB client, <1GB server
- [ ] Network: <100KB/s per player at 20 Hz

### Documentation
- [ ] README.md reflects actual features
- [ ] USER_MANUAL.md 100% accurate
- [ ] ROADMAP.md updated with actual status
- [ ] API documentation complete

---

## Risk Assessment

### High Risk Items
1. **Network Server** - Complex integration, potential protocol issues
   - Mitigation: Start early, thorough testing, protocol versioning
2. **Audio Integration** - Timing issues, Ebiten audio quirks
   - Mitigation: Test on multiple platforms, add mute fallback
3. **UI Systems** - Complex state management, input handling
   - Mitigation: Start with simple layouts, iterate

### Medium Risk Items
1. **Console System** - Command parsing edge cases
2. **Particle Effects** - Performance impact with many particles
3. **Quest Tracking** - Complex objective condition checking

### Low Risk Items
1. **Keyboard Shortcuts** - Simple key binding additions
2. **Config Persistence** - Standard JSON serialization
3. **Logging** - Standard file I/O
4. **Screenshots** - Built-in Ebiten functionality

---

## Post-Implementation Tasks

### Beta Release Preparation
1. Create installation packages (Linux, macOS, Windows)
2. Write beta testing guide
3. Set up bug reporting process
4. Create demo video/screenshots
5. Prepare marketing materials

### Beta Testing Phase
1. Recruit 10-20 beta testers
2. Collect feedback on all systems
3. Monitor performance metrics
4. Fix critical bugs
5. Iterate on UI/UX based on feedback

### Release Candidate
1. Final documentation review
2. Performance optimization pass
3. Security audit (especially networking)
4. Create release notes
5. Tag v0.9.0-beta1

---

## Conclusion

This plan addresses all 12 identified gaps with a realistic 23-25 day timeline. The critical path focuses on user-blocking issues (networking, UIs), followed by polish (audio, particles, console), and finally quality-of-life improvements (map, config, logging).

**Key Priorities:**
1. Network server (enables multiplayer)
2. Inventory/Quest UIs (enables player access to backends)
3. Menu system completion (enables settings/navigation)
4. Audio/Particle integration (polish for beta)

**Estimated Total Effort:** 20-25 development days (4-5 weeks)

After completion, the project will genuinely meet "Beta Release" status with all documented features functional and ready for external testing.
