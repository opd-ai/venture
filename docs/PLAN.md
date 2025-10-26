# Venture Game Mechanics Expansion Plan

**Status**: In Progress  
**Target Version**: 1.1+  
**Last Updated**: October 26, 2025

This document outlines the roadmap for expanding Venture's gameplay mechanics beyond the current 1.0 foundation.

## Completed Items ✅

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

## Phase 1: Menu System & Game Modes

**Goal**: Add splash screen menu for game mode selection.

**Components**:
- Main menu UI with "Single-Player" and "Multi-Player" options
- Single-player submenu: "New Game" / "Load Game"
- Multi-player submenu: server address field + "Connect" button
- Menu state management in client

**Technical Notes**:
- Extend `pkg/rendering/ui` for menu components
- Update `cmd/client/main.go` initialization flow
- Menu state stored in world manager

---

## Phase 2: Character Creation & Tutorial

**Goal**: Unified onboarding experience combining character creation with tutorial.

**Components**:
- Interactive character creation interface
- Name input, basic appearance selection
- Tutorial prompts during gameplay start
- Support for both single-player and multiplayer contexts

**Technical Notes**:
- New character creation system in `pkg/engine`
- Integrate with existing player entity generation
- Tutorial state tracked via quest system
- Network sync for multiplayer character data

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
