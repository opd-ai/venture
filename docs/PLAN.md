# Venture Game Mechanics Expansion Plan

**Status**: In Progress  
**Target Version**: 1.1+  
**Last Updated**: October 26, 2025

This document outlines the roadmap for expanding Venture's gameplay mechanics beyond the current 1.0 foundation.

## Completed Items ✅

### Dynamic Music Context Switching (January 2026)
**Goal**: Implement adaptive music system that changes based on game context.

**Implementation**:
- Created `pkg/engine/music_context.go` with context detection system
- MusicContext enum: Exploration, Combat, Boss, Danger, Victory, Death
- Proximity-based enemy detection (300px radius)
- Priority-based transition manager with cooldown (10 seconds)
- Integrated into AudioManagerSystem for automatic context switching
- 96.9% test coverage with 24 table-driven test cases

**Status**: ✅ Complete - Production-ready with comprehensive testing

**Usage**: AudioManagerSystem automatically detects context changes and switches music accordingly.

---

### Host-and-Play Mode (October 26, 2025)
**Goal**: Single-command LAN party mode for easy multiplayer hosting.

**Implementation**:
- Created `pkg/hostplay` package with server lifecycle management
- Added `--host-and-play` and `--host-lan` flags to client
- Port fallback mechanism (8080-8089) with automatic detection
- 96% test coverage with comprehensive error handling
- Documentation updated in README.md and ROADMAP.md

**Usage**: `./venture-client --host-and-play` starts server and connects automatically.

---

## Phase 1: Menu System & Game Modes ✅ **COMPLETED (MVP)** (October 26, 2025)

**Goal**: Add splash screen menu for game mode selection.

**Implementation**:
- ✅ Created `pkg/engine/app_state.go` with AppStateManager for application state management (100% test coverage)
- ✅ Created `pkg/engine/main_menu_ui.go` with keyboard/mouse navigation (92.3% test coverage on testable functions)
- ✅ Integrated into EbitenGame with state-aware Update()/Draw() methods
- ✅ Main menu options: Single-Player, Multi-Player, Settings, Quit
- ✅ Callbacks for state transitions (onNewGame, onMultiplayerConnect, onQuitToMenu)

**Status**: ✅ MVP Complete - Main menu system functional with simplified flow

**Current Behavior**:
- Game starts in AppStateMainMenu showing main menu
- Single-Player directly transitions to gameplay (New Game)
- Multi-Player uses existing CLI flags for server connection
- Settings shows "not implemented" message (future feature)
- Quit option ready for implementation

**Future Enhancements** (Phase 1.1):
- Single-player submenu with "New Game" / "Load Game" / "Back"
- Multi-player submenu with server address input field
- Settings menu implementation
- Character creation integration before gameplay start

**Technical Notes**:
- AppState separate from GameState (input filtering) to avoid namespace collision
- State machine enforces valid transitions
- Callbacks allow client to control world initialization timing
- Menu rendered before world generation for fast startup

---

## Phase 1 Implementation Notes

The Phase 1 MVP implements a working main menu system following the SIMPLICITY RULE from copilot-instructions. The implementation:

1. **Uses existing patterns**: Leverages Ebiten's Update()/Draw() flow with state checking
2. **Minimizes abstraction**: Simple state enum + manager, no complex state pattern
3. **Defers complexity**: Submenus planned for Phase 1.1 after MVP validation
4. **Maintains testability**: 100% coverage on state machine, 92%+ on UI logic

This approach allows immediate user feedback while keeping the codebase maintainable.

---

## Phase 2: Character Creation & Tutorial ✅ **COMPLETED** (October 26, 2025)

**Goal**: Unified onboarding experience combining character creation with tutorial.

**Implementation**:
- ✅ Created `pkg/engine/character_creation.go` with interactive UI (626 lines)
- ✅ Character class system: Warrior (high HP/defense), Mage (high mana/magic), Rogue (balanced/agility)
- ✅ Three-step creation flow: Name input → Class selection → Confirmation
- ✅ Integrated with AppState system (AppStateCharacterCreation)
- ✅ Character data applied to player entity via ApplyClassStats()
- ✅ Comprehensive test suite with 100% coverage on testable functions (16 test functions, 42 test cases)
- ✅ Tutorial integration: class descriptions teach gameplay during selection

**Status**: ✅ Complete - Production-ready with full integration

**Current Behavior**:
- Main Menu → Single-Player → Character Creation (3 steps) → Gameplay
- Name input with keyboard entry (alphanumeric + spaces, max 20 characters)
- Class selection via arrow keys or number keys (1-3)
- Tutorial prompts embedded in class descriptions
- Stats applied immediately on game start
- Multiplayer-ready (character data prepared for network sync)

**Class Stats**:
- **Warrior**: HP 150, Mana 50, Attack 12, Defense 8, Crit Damage 2.0x
- **Mage**: HP 80, Mana 150, Attack 6, Defense 3, Crit Chance 10%, Mana Regen 8/s
- **Rogue**: HP 100, Mana 80, Attack 10, Defense 5, Crit 15%, Evasion 15%, Fast Attacks (0.3s cooldown)

**Technical Notes**:
- Character data stored in pending state during transition
- ApplyClassStats() modifies health, mana, stats, and attack components
- Validation ensures names 1-20 characters, valid class selection
- UI uses Ebiten drawing with keyboard navigation (no mouse required)
- Future: Add NameComponent for multiplayer player identification

---

## Phase 3: Commerce & NPC Interaction

**Goal**: Shop system with merchant NPCs and dialog interface.

**Components**:
- Fixed-location shopkeepers (towns/settlements)
- Nomadic merchants with procedural spawn logic
- Basic dialog system with extensible interface
- Buy/sell transactions integrated with inventory

**Technical Notes**:
- Merchant entity type in `pkg/procgen/entity`
- Dialog component + system in `pkg/engine`
- Shop UI in `pkg/rendering/ui`
- Transaction validation on server for multiplayer

---

## Phase 4: Environmental Manipulation

**Goal**: Destructible and constructible terrain.

**Components**:
- **Destruction**: Wall breaking via weapons/spells, fire propagation
- **Construction**: Wall building with materials, magic terrain creation
- Environmental damage system
- Visual effects for terrain changes

**Technical Notes**:
- Terrain modification in `pkg/world` with network sync
- New destructible terrain component
- Fire propagation system (cellular automata)
- Client prediction for instant feedback

---

## Phase 5: Crafting Systems

**Goal**: Potion brewing, enchanting, and magic item crafting.

**Components**:
- Recipe system for potions, enchantments, magic items
- Crafting UI with ingredient slots
- Integration with skill tree (crafting skills)
- Resource gathering from environment/enemies

**Technical Notes**:
- Recipe definitions in `pkg/procgen/item`
- Crafting system in `pkg/engine`
- Recipe discovery via skill progression
- Deterministic crafting results (seed-based)

---

## Cross-Cutting Concerns

**Multiplayer**: All features require server authority and client prediction where applicable.

**Determinism**: Procedural elements (merchant spawns, recipes) must use seed-based generation.

**Performance**: Target 60 FPS with <500MB memory for all features.

**Testing**: Maintain 65%+ coverage for new packages/systems.

**UI/UX**: Consistent with existing dual-exit menu pattern (key/ESC).

---

## Implementation Order

Phases designed for incremental delivery:
1. **Phase 1** - Foundation for improved UX (menu system)
2. **Phase 2** - Enhanced onboarding (character creation)
3. **Phase 3** - Gameplay depth (NPC interaction)
4. **Phase 4** - World interaction (terrain manipulation)
5. **Phase 5** - Player expression (crafting)

Each phase can be developed and released independently.

---

## Future Considerations

Beyond Phase 5, potential expansions include:
- Quest branching and consequences
- Base building mechanics
- Advanced AI behaviors
- Procedural storytelling
- Biome-specific mechanics

These will be planned based on community feedback and Phase 1-5 outcomes.
