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

**Status**: ✅ MVP Complete - Main menu system functional with settings menu

**Current Behavior**:
- Game starts in AppStateMainMenu showing main menu
- Single-Player directly transitions to gameplay (New Game)
- Multi-Player uses existing CLI flags for server connection
- **Settings menu fully functional with interactive controls** ✅ **(NEW)**
- Quit option ready for implementation

**Future Enhancements** (Phase 1.1):
- Single-player submenu with "New Game" / "Load Game" / "Back"
- Multi-player submenu with server address input field

**Settings Menu Implementation** ✅ **COMPLETE** (October 27, 2025):
- ✅ Created `pkg/engine/settings.go` with SettingsManager for persistent storage
- ✅ Created `pkg/engine/settings_ui.go` with interactive controls (keyboard/mouse navigation)
- ✅ Settings stored in `~/.venture/settings.json` with automatic validation
- ✅ Options: Master/Music/SFX Volume, Graphics Quality, VSync, Show FPS, Fullscreen
- ✅ Dual-exit navigation (ESC to save and return)
- ✅ Comprehensive test suites: 90.7% coverage on settings.go, 95%+ on settings_ui.go
- ✅ Integrated with AppState system (AppStateSettings)
- ✅ Automatic save on hide, graceful degradation on errors
- ✅ Real-time adjustment with visual feedback (< value >)

**Settings Application Integration** ✅ **COMPLETE** (October 27, 2025):
- ✅ Added `ApplySettings()` method to `pkg/engine/game.go` for real-time application
- ✅ Volume control: applies Master/Music/SFX volumes to AudioManager (master * specific)
- ✅ Display control: applies VSync and Fullscreen settings to Ebiten window
- ✅ Callback system: SettingsUI.SetApplyCallback() → game.ApplySettings()
- ✅ AudioManager wiring: SetAudioManager() auto-applies settings on initialization
- ✅ Client integration: game.SetAudioManager(audioManager) in `cmd/client/main.go`
- ✅ Comprehensive test suite: `settings_integration_test.go` (13 test functions, 2 benchmarks)
  * TestApplySettings_AudioVolumes: Validates volume application with master multiplier
  * TestApplySettings_NoAudioManager: Graceful handling when AudioManager is nil
  * TestApplySettings_NoSettingsManager: Graceful handling when SettingsManager is nil
  * TestSetAudioManager: Validates auto-apply on manager initialization
  * TestSettingsUI_ApplyCallback: Verifies callback triggers on Hide()
  * TestApplySettings_MasterVolumeZero: Confirms zero master mutes all audio
  * TestApplySettings_MaxVolumes: Validates maximum volume settings
  * BenchmarkApplySettings: Performance baseline for settings application
  * BenchmarkSetAudioManager: Performance baseline for manager initialization
- ✅ Coverage: ApplySettings 78.6%, SetAudioManager 100%, SetApplyCallback 100%
- ✅ All tests passing (13/13)
- ✅ Zero build errors, production-ready

**Technical Notes**:
- AppState separate from GameState (input filtering) to avoid namespace collision
- State machine enforces valid transitions
- Callbacks allow client to control world initialization timing
- Menu rendered before world generation for fast startup
- **Settings Application**: Changes apply immediately to running game (no restart required)
- **Master Volume**: Acts as global multiplier (actualVolume = masterVolume * specificVolume)
- **Callback Chain**: SettingsUI.Hide() → onSettingsApplied → game.ApplySettings()
- **Graceful Degradation**: ApplySettings handles nil AudioManager/SettingsManager safely
- **Initialization Order**: SetAudioManager() automatically applies settings on first call

---

## 🚀 Future Enhancements

These are optional features that would add polish but are not required for core functionality:

### Genre System Enhancements
- [ ] Dynamic genre transitions during gameplay
- [ ] Genre-specific world events and weather effects
- [ ] Cross-genre legendary items with unique mechanics

### Multiplayer Enhancements  
- [ ] Voice chat support
- [ ] Guild/clan systems
- [ ] PvP arenas
- [ ] Ranked leaderboards

### Performance & Polish
- [ ] Level streaming for massive worlds
- [ ] Texture atlasing for sprite batching
- [ ] Advanced particle pooling
- [ ] LOD system for distant entities

---

## 📋 Archived Phases

The following phases represent completed major milestones in the project:

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
- ✅ Comprehensive test suite with 100% coverage on testable functions (22 test functions, 52+ test cases)
- ✅ Tutorial integration: class descriptions teach gameplay during selection
- ✅ **Custom Defaults Feature**: Press F2 to save preferred name/class for faster repeated character creation
- ✅ **Custom Portrait Feature**: User-provided .png images (up to 512x512, auto-downscaled) for visual customization *(NEW)*

**Status**: ✅ Complete - Production-ready with full integration

**Current Behavior**:
- Main Menu → Single-Player → Character Creation (4 steps) → Gameplay
- Main Menu → Multi-Player → Character Creation (4 steps) → Connect to Server
- Step 1: Name input with keyboard entry (alphanumeric + spaces, max 20 characters)
- Step 2: Class selection via arrow keys or number keys (1-3)
- Step 3: Portrait selection - **press SPACE/B to open native file picker dialog** (optional, press TAB to skip) *(UPDATED)*
- Step 4: Confirmation showing name, class, portrait preview, and stats
- Tutorial prompts embedded in class descriptions
- Stats applied immediately on game start (both single-player and multiplayer)
- Character data ready for network sync to server
- **Custom Defaults**: Press F2 on name/class/portrait screens to save as defaults; Reset() automatically applies saved defaults
- **Portrait Validation**: Auto-downscales images >512x512 while preserving aspect ratio; only .png files accepted
- **Native File Dialog**: Uses platform-native file pickers (Windows Explorer, macOS Finder, Linux file dialogs) starting in user's Pictures directory *(NEW)*

**Class Stats**:
- **Warrior**: HP 150, Mana 50, Attack 12, Defense 8, Crit Damage 2.0x
- **Mage**: HP 80, Mana 150, Attack 6, Defense 3, Crit Chance 10%, Mana Regen 8/s
- **Rogue**: HP 100, Mana 80, Attack 10, Defense 5, Crit 15%, Evasion 15%, Fast Attacks (0.3s cooldown)

**Technical Notes**:
- Character data stored in pending state during transition
- isMultiplayerMode flag tracks whether creating character for single-player or multiplayer
- Single-player: triggers onNewGame() callback after character creation
- Multiplayer: triggers onMultiplayerConnect() callback after character creation
- ApplyClassStats() modifies health, mana, stats, and attack components
- Validation ensures names 1-20 characters, valid class selection
- UI uses Ebiten drawing with keyboard navigation (no mouse required)
- Character data automatically synced to server when connecting (multiplayer)
- **Custom Defaults**: CharacterCreationDefaults struct stores default name/class/portrait; SetDefaults()/GetDefaults() for configuration; F2 handlers save current selection; displayed in gray text
- **Portrait System**: LoadPortrait() loads .png from user's local filesystem; validates extension and downscales >512x512 using bilinear interpolation; preserves aspect ratio; Portrait field in CharacterData holds *ebiten.Image; user-provided images not considered "game assets"
- **File Dialog System**: OpenPortraitDialog() uses github.com/ncruces/zenity for native platform dialogs without GTK dependencies; GetDefaultPicturesDirectory() detects OS-specific Pictures folder (Windows: %USERPROFILE%\Pictures, macOS: ~/Pictures, Linux: ~/Pictures, Mobile: app directory); cross-platform support for Windows, macOS, Linux, Android, iOS *(NEW)*
- Future: Server validates and stores character data, broadcasts to other players; portrait images can be synced via base64 encoding or hash-based caching

---

## Phase 3: Commerce & NPC Interaction ✅ COMPLETE (October 26, 2025)

**Goal**: Shop system with merchant NPCs and dialog interface.

**Implementation Progress**:
- ✅ Created `pkg/engine/commerce_components.go` with MerchantComponent, DialogComponent, and related types (320 lines, 87.4% coverage)
- ✅ Created `pkg/engine/dialog_system.go` with DialogSystem for NPC interactions (205 lines, 77.7% coverage)
- ✅ Created `pkg/engine/commerce_system.go` with atomic transaction logic (370 lines, 85.3% coverage)
- ✅ Created `pkg/engine/shop_ui.go` with ShopUI for merchant interaction interface (490 lines, 92.1% coverage)
- ✅ Added Shop key (S) to MenuKeys in `pkg/engine/menu_keys.go` for standardized navigation
- ✅ Comprehensive test suites: `commerce_components_test.go` (390+ lines, 24 test cases), `dialog_system_test.go` (420+ lines, 13 test functions), `commerce_system_test.go` (520+ lines, 13 test functions with 52+ test cases), `shop_ui_test.go` (490+ lines, 16 test functions)
- ✅ MerchantDialogProvider for simple buy/sell/leave dialogs
- ✅ DefaultTransactionValidator for extensible transaction validation
- ✅ BuyItem() and SellItem() methods with atomic rollback on failure
- ✅ Shop UI with dual-mode (buy/sell), keyboard/mouse navigation, dual-exit (S key + ESC), transaction feedback
- ✅ Merchant NPC generation in `pkg/procgen/entity/merchant.go` (282 lines, 92.0% coverage)
  - GenerateMerchant() creates merchants with genre-appropriate names and inventory
  - MerchantData struct holds entity, type (fixed/nomadic), inventory, pricing parameters
  - GenerateMerchantSpawnPoints() provides deterministic spawn locations
  - Added NPC templates for all 5 genres (scifi, horror, cyberpunk, postapoc)
  - MerchantNameTemplates with genre-specific merchant names
  - Comprehensive test suite with 9 test functions + 2 benchmarks (merchant_test.go)
- ✅ Network protocol support in `pkg/network/protocol.go` (6 new message types)
  - OpenShopMessage: Client request to interact with merchant
  - ShopInventoryMessage: Server response with merchant stock and pricing
  - BuyItemMessage: Client request to purchase item
  - SellItemMessage: Client request to sell item
  - TransactionResultMessage: Server response with transaction outcome
  - CloseShopMessage: Client notification of shop UI closure
  - Comprehensive test suite with 8 test functions covering all message types (protocol_test.go)
  - Full workflow test demonstrating client-server commerce interaction
  - Error scenario tests for transaction failures
- ✅ Client integration in `cmd/client/main.go` (October 26, 2025)
  - SpawnMerchantsInTerrain() spawns 2 merchants per dungeon level
  - CommerceSystem and DialogSystem initialized and wired to game systems
  - ShopUI integrated into game rendering and update loop
  - F key for NPC interaction with proximity detection (64 pixel range)
  - FindClosestMerchant() and GetNearbyMerchants() for merchant discovery
  - GetMerchantInteractionPrompt() for UI feedback
  - Shop UI blocks game input when visible (similar to inventory/quest UI)
- ✅ Helper functions in `pkg/engine/merchant_spawn.go` (312 lines)
  - SpawnMerchantFromData() converts procgen MerchantData to engine entities
  - SpawnMerchantsInTerrain() spawns merchants at deterministic locations
  - GetNearbyMerchants() returns merchants within radius
  - FindClosestMerchant() finds nearest merchant for interaction
  - GetMerchantInteractionPrompt() generates UI text
  - Comprehensive test suite with 5 test functions (merchant_spawn_test.go)

**Technical Notes**:
- Merchant Generation: docs/IMPLEMENTATION_MERCHANT_GENERATION.md
- Network Protocol: docs/IMPLEMENTATION_COMMERCE_PROTOCOL.md
- All merchants have dialog components with MerchantDialogProvider
- Merchants spawn with full inventories (15-24 items, genre-appropriate)
- Fixed merchants for dungeon shops, nomadic merchants for future wandering NPCs
- F key interaction with proximity detection (no need to aim)
- Shop UI follows same patterns as Inventory/Quest UI for consistency
- Transaction validation extensible via TransactionValidator interface
- Server-authoritative commerce prevents multiplayer exploits

**Components**:
- Fixed-location shopkeepers (towns/settlements)
- Nomadic merchants with procedural spawn logic
- Dialog system with extensible provider interface
- Buy/sell transactions integrated with inventory system
- Network protocol for multiplayer commerce synchronization
- Proximity-based NPC interaction system

**Known Limitations**:
- Merchants currently have infinite gold (future: limited merchant funds)
- No merchant reputation system yet
- Dialog system is simple (future: branching conversations, quest dialogs)
- No merchant restocking yet (RestockTimeSec field prepared for future use)

**Technical Notes**:
- Merchant entity type in `pkg/procgen/entity`
- Dialog component + system in `pkg/engine`
- Shop UI in `pkg/rendering/ui`
- Transaction validation on server for multiplayer
- **Detailed Documentation**: 
  - Merchant Generation: `docs/IMPLEMENTATION_MERCHANT_GENERATION.md`
  - Network Protocol: `docs/IMPLEMENTATION_COMMERCE_PROTOCOL.md`

---

## Phase 4: Environmental Manipulation ✅ **COMPLETE** (October 26, 2025)

**Goal**: Destructible and constructible terrain.

**Implementation Progress**:
- ✅ Created `pkg/engine/terrain_components.go` with terrain modification components (October 26, 2025)
  - MaterialType enum with 6 material types (Stone, Wood, Earth, Metal, Glass, Ice)
  - DestructibleComponent for tile health, damage tracking, and destruction
  - FireComponent for fire propagation with intensity, duration, and spread mechanics
  - BuildableComponent for construction progress tracking
  - Comprehensive test suite: `terrain_components_test.go` (24 test functions, 95%+ coverage)
  - All components follow ECS pattern (data-only, no behavior)
- ✅ Created `pkg/engine/terrain_modification_system.go` with destruction system (October 26, 2025)
  - TerrainModificationSystem handles weapon/spell-based tile destruction
  - ProcessWeaponAttack() for weapon-based terrain damage
  - Automatic tile replacement: wall → floor when destroyed
  - Destructible entity management (creation, damage tracking, removal)
  - Helper methods: DamageTileAtWorldPosition(), DamageTilesInArea() for area effects
  - Server-authoritative design (requires world, terrain, worldMap references)
  - Comprehensive test suite: `terrain_modification_system_test.go` (18 test functions, 3 benchmarks, all passing)
  - 343 lines with comprehensive logging support
- ✅ Created `pkg/engine/fire_propagation_system.go` with fire spread system (October 26, 2025)
  - FirePropagationSystem implements cellular automata for fire propagation
  - Checks 4-connected neighbors (up, down, left, right) for spread
  - Spread chance: 0.3 * intensity per second to adjacent tiles
  - Flammability based on material type (only wood is flammable)
  - Fire burns for 10-15 seconds (default: 12, configurable)
  - Helper methods: IgniteTile(), IgniteTilesInArea() for external systems
  - GetActiveFireCount() for performance monitoring
  - Comprehensive test suite: 10 test functions, all passing
  - 348 lines with full ECS integration
- ✅ Created `pkg/engine/terrain_construction_system.go` with wall building system (October 26, 2025)
  - TerrainConstructionSystem handles wall placement with inventory material consumption
  - StartConstruction() validates placement and creates buildable entity
  - Material requirements: Default 10 stone per wall (configurable via BuildableComponent)
  - Material sources: Inventory items (name-based detection) + gold (10 gold = 1 stone equivalent)
  - Construction time: 3 seconds default (configurable)
  - Update() processes construction progress with timer
  - completeConstruction() places wall tile and removes buildable entity
  - GetConstructionProgress() returns 0.0-1.0 for UI feedback
  - Placement validation: Tile must be walkable (TileFloor), not occupied
  - Comprehensive test suite: 12 test functions, all passing, 2 benchmarks
  - 356 lines with full ECS integration
- ✅ Network protocol support in `pkg/network/protocol.go` (8 new message types) (October 26, 2025)
  - TileDamageMessage: Client→Server damage request (PlayerID, TileX/Y, Damage, WeaponID, SequenceNumber)
  - TileDestroyedMessage: Server→Clients destruction notification (TileX/Y, TimeOfDestruction, DestroyedByPlayerID, SequenceNumber)
  - TileConstructMessage: Client→Server build request (PlayerID, TileX/Y, TileType, SequenceNumber)
  - ConstructionStartedMessage: Server→Clients start notification (TileX/Y, BuilderPlayerID, TileType, ConstructionTime, TimeStarted, SequenceNumber)
  - ConstructionCompletedMessage: Server→Clients completion notification (TileX/Y, TileType, TimeCompleted, SequenceNumber)
  - FireIgniteMessage: Client→Server ignition request (PlayerID, TileX/Y, Intensity, SourceType, SequenceNumber)
  - FireSpreadMessage: Server→Clients spread notification (TileX/Y, Intensity, Duration, TimeIgnited, SequenceNumber)
  - FireExtinguishedMessage: Server→Clients extinguish notification (TileX/Y, TimeExtinguished, Reason, SequenceNumber)
  - Comprehensive test suite: `protocol_test.go` (11 test functions: 8 structure tests, 2 workflow tests, 1 validation test, all passing)
  - Coverage: 57.1% (protocol.go is simple data structures, tests verify struct initialization and workflows)
  - Server-authoritative design: clients request, server validates and broadcasts
- ✅ Additional test suites (terrain_modification_system_test.go) (October 26, 2025)
  - Comprehensive test coverage for TerrainModificationSystem
  - 18 test functions: reference setting, weapon checks, damage calculation, direction detection, tile damage, destruction flow, weapon attack processing, world coordinate damage, area damage, entity lookup, entity creation, tile replacement
  - 3 benchmark functions: DamageTile (396ns/op), DamageTilesInArea (5.6μs/op), ProcessWeaponAttack (479ns/op)
  - All tests passing, covers success and error paths
  - Performance validated: <1ms per operation (meets Phase 4 target)

**Components**:
- **Destruction**: Wall breaking via weapons/spells, fire propagation
- **Construction**: Wall building with materials, magic terrain creation
- Environmental damage system
- Visual effects for terrain changes

**Technical Notes**:
- Terrain modification in `pkg/world` with network sync
- Material types: Stone (100 HP, not flammable), Wood (50 HP, flammable), Earth (30 HP), Metal (200 HP), Glass (20 HP), Ice (40 HP)
- Fire burns for 10-15 seconds (configurable), spreads to adjacent flammable tiles at 30% * intensity chance/second
- Construction requires materials (default: 10 stone per wall), takes 3 seconds (configurable)
- Weapon damage to terrain: weapon.Stats.Damage * 0.5 (terrain multiplier)
- Attack direction determined by AnimationComponent.Facing (DirUp/Down/Left/Right)
- Destructible entities automatically created for damaged walls, removed when destroyed
- Fire spread uses cellular automata: checks 4-connected neighbors each update
- Fire tracking via internal map for performance (<2ms per frame for 100 fires)
- Client prediction for instant feedback
- Server-authoritative for multiplayer synchronization

**Usage Example - Terrain Destruction**:
```go
// Setup
system := engine.NewTerrainModificationSystem(tileSize)
system.SetWorld(world)
system.SetTerrain(terrain)
system.SetWorldMap(worldMap)

// Process weapon attack
system.ProcessWeaponAttack(playerEntity, equippedWeapon)

// Explosion damage
system.DamageTilesInArea(centerX, centerY, radius, damage)
```

**Usage Example - Fire Propagation**:
```go
// Setup
system := engine.NewFirePropagationSystem(tileSize, seed)
system.SetWorld(world)
system.SetTerrain(terrain)

// Ignite single tile
system.IgniteTile(tileX, tileY, intensity)

// Ignite area (explosions, fire spells)
system.IgniteTilesInArea(centerX, centerY, radius, intensity)

// Update in game loop
system.Update(entities, deltaTime)

// Monitor fire count
activeCount := system.GetActiveFireCount()
```

**Usage Example - Terrain Construction**:
```go
// Setup
system := engine.NewTerrainConstructionSystem(tileSize)
system.SetWorld(world)
system.SetTerrain(terrain)
system.SetWorldMap(worldMap)

// Start building a wall
err := system.StartConstruction(builderEntity, tileX, tileY, world.TileWall)
if err != nil {
    // Handle error: invalid placement, insufficient materials, etc.
}

// Update in game loop
entities := world.GetEntities()
system.Update(entities, deltaTime)

// Check construction progress for UI feedback
progress := system.GetConstructionProgress(tileX, tileY) // 0.0-1.0
```

**Status**: ✅ Phase 4 Complete (Components, Modification System + Tests, Fire System, Construction System, Network Protocol)

---

## Phase 5: Crafting Systems ✅ **CORE COMPLETE** (October 26, 2025)

**Goal**: Potion brewing, enchanting, and magic item crafting.

**Implementation Progress**:
- ✅ Created `pkg/engine/crafting_components.go` with comprehensive component system (October 26, 2025)
  - RecipeKnowledgeComponent for recipe discovery (unlimited or slot-limited)
  - CraftingSkillComponent with XP progression (100 XP per level, scaling requirements)
  - CraftingStationComponent for bonus success/speed (5% success, 25% faster)
  - CraftingProgressComponent for tracking ongoing crafts
  - Recipe struct with materials, costs, skill requirements, success chances
  - MaterialRequirement with name, quantity, optional flag, item type filtering
  - RecipeType enum: Potion, Enchanting, MagicItem
  - RecipeRarity enum: Common, Uncommon, Rare, Epic, Legendary
  - Comprehensive test suite: 100% coverage on components (22 test functions, 60+ test cases)
- ✅ Created `pkg/engine/crafting_system.go` with full crafting workflow (October 26, 2025)
  - StartCraft() validates recipe knowledge, materials, skill, station requirements
  - Update() processes crafting progress with deltaTime integration
  - completeCraft() rolls for success, generates items, awards XP
  - Material consumption with atomic rollback on failure
  - Station reservation system prevents concurrent use
  - Skill-based success scaling: BaseChance + (0.05 * (skillLevel - required)), capped at 95%
  - XP rewards: 10 * (rarity + 1), half XP on failure
  - Integration with item generator for deterministic output
  - Server-authoritative design for multiplayer support
  - 645 lines with comprehensive logging support
- ✅ Created `pkg/procgen/recipe/generator.go` with genre-themed recipes (October 26, 2025)
  - RecipeGenerator with deterministic seed-based generation
  - Template system for all 5 genres (fantasy, sci-fi, horror, cyberpunk, post-apocalyptic)
  - Three recipe types: potions (50%), enchanting (30%), magic items (20%)
  - Rarity distribution: Common 50%, Uncommon 30%, Rare 15%, Epic 4%, Legendary 1%
  - Depth and difficulty modify rarity chances (deeper = rarer recipes)
  - Genre-specific material names and crafting themes
  - Fantasy: Healing Herbs, Mana Crystals, Enchantment Scrolls
  - Sci-Fi: Nano-Gel, Circuit Boards, Plasma Cores
  - Horror: Dried Blood, Bone Dust, Soul Fragments
  - Cyberpunk: Synth-Chem, Neural Links, Titanium Alloy
  - Post-Apocalyptic: Purified Water, Scrap Metal, Duct Tape
  - Comprehensive validation: checks recipe ID, name, materials, success chances
  - 550+ lines with full genre template registration

**Status**: ✅ Core System Complete (Components, CraftingSystem, RecipeGenerator with tests)

**Current Behavior**:
- Recipe knowledge tracked per entity (unlimited or slot-limited)
- Crafting skill progression: 0-100 levels, XP-based with scaling requirements
- Recipe validation: checks knowledge, skill, materials, gold, inventory space
- Material consumption: atomic operation with rollback on validation failure
- Crafting progress: real-time tracking with station bonuses
- Success rolling: deterministic (recipe seed + entity ID), skill-scaled, capped at 95%
- Item generation: uses item generator with recipe parameters
- XP rewards: scaled by rarity, half XP on failure
- Station bonuses: 5% success chance, 25% speed boost when using correct station type
- Multiplayer-ready: server-authoritative validation, client progress tracking

**Recipe Examples**:
- **Healing Potion** (Common): 2 Healing Herb + 1 Water Flask, 10 gold, skill 0, 75% base success, 4s craft
- **Mana Elixir** (Uncommon): 2 Mana Crystal + 1 Arcane Dust, 25 gold, skill 3, 65% base success, 6s craft
- **Minor Enchantment** (Common): 2 Enchantment Scroll + 1 Magic Ink, 30 gold, skill 2, 70% base success, 10s craft
- **Apprentice Wand** (Common): 3 Oak Branch + 1 Magic Crystal, 45 gold, skill 5, 60% base success, 12s craft

**Technical Notes**:
- Components follow ECS pattern: data-only, no behavior
- CraftingSystem integrates with inventory, skills, item generation
- Recipe generation deterministic: same seed + params = same recipes
- Material matching: name-based with optional item type filtering
- Gold can substitute for materials (10 gold = 1 stone equivalent, future enhancement)
- Failed crafts consume 100% of materials (risk/reward), award 50% XP
- Station availability managed automatically (reserve on start, release on complete)
- Crafting skill separate from combat skills (dedicated progression path)
- Recipe discovery: world drops, quest rewards, NPC teaching (future integration)
- Network protocol: client sends StartCraft, server validates and broadcasts progress/result

**Usage Example - Basic Crafting**:
```go
// Setup
craftingSystem := engine.NewCraftingSystem(world, inventorySystem, itemGen)

// Entity must have components
entity.AddComponent(engine.NewRecipeKnowledgeComponent(0)) // Unlimited slots
entity.AddComponent(engine.NewCraftingSkillComponent())     // Start at level 0

// Learn a recipe
recipeGen := recipe.NewRecipeGenerator()
params := procgen.GenerationParams{Difficulty: 0.5, Depth: 1, GenreID: "fantasy"}
result, _ := recipeGen.Generate(12345, params)
recipes := result.([]*engine.Recipe)

knowledgeComp := entity.GetComponent("recipe_knowledge").(*engine.RecipeKnowledgeComponent)
knowledgeComp.LearnRecipe(recipes[0])

// Start crafting (no station)
result, err := craftingSystem.StartCraft(entity.ID, recipes[0], 0)
if err != nil || !result.Success {
    log.Printf("Craft failed: %s", result.ErrorMessage)
}

// Update in game loop
craftingSystem.Update(entities, deltaTime)
```

**Usage Example - With Crafting Station**:
```go
// Create crafting station
station := engine.NewEntity(world.NextEntityID())
station.AddComponent(engine.NewCraftingStationComponent(engine.RecipePotion))
world.AddEntity(station)

// Start crafting at station
result, err := craftingSystem.StartCraft(entity.ID, potionRecipe, station.ID)
// Station provides: +5% success chance, 25% faster craft time
```

**Remaining Work** (Phase 5.1 - UI & Integration):
- [x] **Crafting UI with recipe list, material display, craft button** (October 27, 2025)
  - ✅ Created `pkg/engine/crafting_ui.go` (499 lines) following ShopUI patterns
  - ✅ Recipe list with scrolling support for large recipe collections
  - ✅ Material requirement display with availability checking
  - ✅ Skill requirement and success chance visualization
  - ✅ Crafting progress display with percentage completion
  - ✅ Station bonus indicators (+5% success, 25% faster)
  - ✅ Dual-exit navigation (R key + ESC)
  - ✅ Keyboard and mouse navigation support
  - ✅ Comprehensive test suite: 13 test functions, 95.6% coverage on testable functions
  - ✅ All tests passing (13/13)
- [x] **Client integration: wire system, add R key binding** (October 27, 2025)
  - ✅ Added CraftingUI field to EbitenGame struct in `pkg/engine/game.go`
  - ✅ Initialized CraftingSystem with itemGen in `cmd/client/main.go` (line ~547)
  - ✅ Initialized CraftingUI and connected to player entity and CraftingSystem (line ~910)
  - ✅ Added CraftingUI.Update() call to game loop (updates R key toggle)
  - ✅ Added CraftingUI.Draw() call to rendering pipeline
  - ✅ Added CraftingUI visibility check to input blocking logic
  - ✅ Updated SetPlayerEntity() to wire CraftingUI
  - ✅ Client builds successfully, R key toggle functional (R for Recipe)
- [x] **Recipe discovery integration with loot tables** (October 27, 2025)
  - ✅ Created `RecipeEntityComponent` in `pkg/engine/item_spawning.go` marking recipe entities
  - ✅ Implemented `SpawnRecipeInWorld()` creating pickupable recipe entities with physics
  - ✅ Implemented `GenerateRecipeDrop()` with deterministic generation (5% base, 20% bosses)
  - ✅ Extended `ItemPickupSystem` to handle recipe pickups and auto-learn
  - ✅ Recipe colors based on type: purple (potions), blue (enchanting), gold (magic items)
  - ✅ Rarity affects brightness (1.0x to 2.0x multiplier)
  - ✅ Integration in `cmd/client/main.go` death callback alongside item drops
  - ✅ Comprehensive test suite: 10 new test functions covering spawn, generation, pickup
  - ✅ Recipe drop rate testing: validates 5% normal, 20% boss drop rates
  - ✅ Coverage: SpawnRecipeInWorld 100%, GenerateRecipeDrop 90.9%, getRecipeColor 90.9%
  - ✅ All tests passing (10/10 recipe tests)
- [x] **Network protocol messages for multiplayer** (October 27, 2025)
  - ✅ Created `StartCraftMessage` in `pkg/network/protocol.go` for craft requests (client → server)
  - ✅ Created `CraftProgressMessage` for progress updates (server → clients, periodic sync)
  - ✅ Created `CraftCompleteMessage` for completion/failure notifications (server → clients)
  - ✅ Created `RecipeLearnMessage` for recipe discovery synchronization (server → clients)
  - ✅ All messages follow existing protocol patterns with SequenceNumber, timestamps, entity IDs
  - ✅ Comprehensive test suite in `protocol_test.go`:
    * TestStartCraftMessage_Structure (3 test cases)
    * TestCraftProgressMessage_Structure (3 test cases: start, halfway, nearly done)
    * TestCraftCompleteMessage_Structure (2 test cases: success, failure)
    * TestRecipeLearnMessage_Structure (4 test cases: drop, quest, npc, pickup)
    * TestCraftingWorkflow (complete success flow)
    * TestCraftingWorkflow_Failure (failure flow)
  - ✅ All tests passing (6 test functions, 15+ test cases)
  - ✅ Coverage: 57.1% for network package (protocol.go is data structures)
  - ✅ Server-authoritative design: clients request, server validates and broadcasts
  - ✅ Documentation: comprehensive GoDoc comments for all message types
- [x] **Crafting stations in world generation** (October 27, 2025)
  - ✅ Created `pkg/procgen/station/generator.go` with StationGenerator (308 lines)
  - ✅ Three station types: Alchemy Table (potions), Forge (enchanting), Workbench (magic items)
  - ✅ Genre-specific station names for all 5 genres (fantasy, sci-fi, horror, cyberpunk, post-apocalyptic)
  - ✅ Deterministic generation: same seed always produces same stations
  - ✅ Comprehensive test suite: 11 test functions + 2 benchmarks (94.2% coverage)
  - ✅ Updated `pkg/engine/station_spawn.go` to use new generator (proper type conversion)
  - ✅ Integrated into `cmd/client/main.go`: spawns 3 stations per dungeon level
  - ✅ Station bonuses: +5% success chance, 25% faster crafting for matching recipe types
  - ✅ Crafting UI enhancements:
    * Shows station bonuses in title bar
    * Displays success chance increase when at compatible station (e.g., "60% → 65% (station +5%)")
    * Shows nearby station proximity indicators (within 100 units)
    * Station-aware craft time display with bonus indication
  - ✅ Stations spawn on walkable tiles with 200-pixel minimum separation
  - ✅ Collision detection: stations are solid entities (32×32 pixels)
  - ✅ Visual distinction: stations have AnimationComponent with pulsing effect
  - ✅ Full ECS integration: position, sprite, collider, animation, crafting station components
  - ✅ All tests passing, zero build regressions

**Status**: ✅ **Phase 5.1 COMPLETE** - All core crafting functionality implemented and tested!

**Note**: Crafting system is fully production-ready with stations providing optional bonuses (+5% success, 25% faster craft time). Core system works perfectly without stations; stations enhance the experience.

**Components**:
- Recipe system for potions, enchantments, magic items ✅
- Crafting skill progression ✅
- Material validation and consumption ✅
- Success chance calculation ✅
- Crafting station bonuses ✅
- Deterministic item generation ✅
- Crafting UI with ingredient slots ✅
- Recipe discovery via loot drops ✅
- Network protocol for multiplayer ✅
- Station spawning in world generation ✅

**Technical Notes**:
- Recipe definitions in `pkg/procgen/recipe` ✅
- Crafting system in `pkg/engine` ✅
- Recipe discovery via skill progression ✅
- Deterministic crafting results (seed-based) ✅
- Server-authoritative validation ✅
- Client-side progress tracking ✅
- Network protocol in `pkg/network/protocol.go`:
  * StartCraftMessage: Client requests crafting with recipe ID and optional station
  * CraftProgressMessage: Server broadcasts progress (elapsed/total time, percent complete)
  * CraftCompleteMessage: Server broadcasts result (success/failure, item, XP, materials lost)
  * RecipeLearnMessage: Server broadcasts recipe discovery (source: drop, quest, npc, pickup, level_up)
  * All messages include SequenceNumber for ordering, timestamps for synchronization
  * Server validates recipe knowledge, materials, skill, station before starting craft
  * Clients update UI progress bars and components from CraftProgressMessage
  * Client-side prediction possible but server has final authority ✅

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
