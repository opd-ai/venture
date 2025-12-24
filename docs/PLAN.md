# Integration Plan - Complete Resolution of All Integration Gaps

**Date:** 2025-12-22  
**Status:** Ready for Implementation  
**Total Gaps:** 34 integration gaps across 6 categories

---

## Executive Summary

This plan provides complete resolution steps for all 44 documented integration gaps in the Venture codebase. Each gap is categorized, prioritized, and includes specific code changes required for full integration.

### Integration Categories

| Category | Description | Count | Priority |
|----------|-------------|-------|----------|
| **A** | System Interface Implementation | 6 | High |
| **B** | UI Integration | 15 | Medium |
| **D** | Data Persistence | 5 | High |
| **E** | Reward System | 1 | Medium |
| **F** | Gameplay Integration | 4 | High |
| **G** | Federation | 3 | Medium |

---

## Core Principles

1. **No Feature Flags** — All features become unconditionally enabled
2. **Component Initialization During Creation** — Never lazy-initialize in Update()
3. **Defensive Nil Checks** — Always check `if ok && comp != nil`
4. **Platform Parity** — Integrate in both desktop and mobile clients
5. **System Registration Order** — Respect dependencies

---

## Phase 1: High Priority - System Interfaces (Week 1)

### Gap A1: AdaptiveMusicSystem Interface Implementation ✅ COMPLETED

**File:** `pkg/engine/audio_manager.go`  
**Gap:** AudioManager did not implement audio.AdaptiveMusicSystem interface required by MusicTriggerSystem

**Status:** RESOLVED (Pre-existing implementation verified 2025-12-22)

**Implementation:**
AudioManager already implements all required methods of `audio.AdaptiveMusicSystem`:
- `SetContext(context MusicContext) error`
- `UpdateIntensity(intensity float64) error`
- `AddLayer(layer MusicLayer) error`
- `RemoveLayer(layer MusicLayer) error`
- `Update(deltaTime float64)`
- `GenerateTrack(duration float64) *AudioSample`

**Verification:**
```bash
# Compile-time verification confirms interface implementation
var _ audio.AdaptiveMusicSystem = &engine.AudioManager{}
go test ./pkg/engine/... -run AudioManager  # All tests pass
```

---

### Gap A2: Consumable Spell Effect Activation ✅ COMPLETED

**File:** `pkg/engine/inventory_system.go`  
**Gap:** Using consumable items doesn't trigger spell effects (potions, scrolls)

**Status:** RESOLVED (2025-12-22)

**Implementation:**
1. Added `SpellEffectID` field to `item.Item` struct in `pkg/procgen/item/types.go`
2. Added `spellEffectSystem` field and `SetSpellEffectSystem()` method to InventorySystem
3. Implemented `applyScrollSpellEffect()` to trigger spell effects when scrolls are consumed
4. Implemented `mapSpellEffectID()` to convert spell IDs to EffectTypes
5. Implemented `calculateScrollMagnitude()` for rarity-based magnitude scaling

**Integration Points:**
1. `pkg/procgen/item/types.go` - Added `SpellEffectID` field to Item struct
2. `pkg/engine/inventory_system.go` - Scroll consumption now triggers spell effects
3. `pkg/engine/spell_effect_system.go` - Uses `ApplySpellEffect()` method

**Tests Added:**
- `TestInventorySystem_UseScrollWithSpellEffect` - Verifies spell effects are applied
- `TestInventorySystem_UseScrollWithoutSpellSystem` - Verifies graceful degradation
- `TestInventorySystem_UseScrollWithoutSpellEffectID` - Verifies handling of scrolls without spell IDs
- `TestInventorySystem_SetSpellEffectSystem` - Verifies setter
- `TestInventorySystem_MapSpellEffectID` - Tests all spell ID mappings
- `TestInventorySystem_CalculateScrollMagnitude` - Tests magnitude calculation by rarity

---

### Gap A3: Mini-Game Item Rewards ✅ COMPLETED

**File:** `pkg/engine/minigame_system.go`  
**Gap:** EndGame() doesn't add reward.Items to player inventory

**Status:** RESOLVED (2025-12-22)

**Implementation:**
The `awardReward` method now properly iterates through `reward.Items` (which contains entity IDs), retrieves the ItemComponent from each entity, and adds the items to the player's inventory:

```go
// In awardReward() method:
if len(reward.Items) > 0 {
    invCompRaw, hasInv := entity.GetComponent("inventory")
    if hasInv {
        invComp := invCompRaw.(*InventoryComponent)
        for _, itemEntityID := range reward.Items {
            itemEntity, exists := s.world.GetEntity(itemEntityID)
            if !exists || itemEntity == nil {
                continue
            }
            itemCompRaw, hasItem := itemEntity.GetComponent("item")
            if !hasItem {
                continue
            }
            itemComp, ok := itemCompRaw.(*ItemComponent)
            if !ok || itemComp == nil || itemComp.Item == nil {
                continue
            }
            invComp.AddItem(itemComp.Item)
        }
    }
}
```

**Tests Added:**
- `TestMiniGameSystem_AwardReward_WithItems` - Verifies items are properly added to inventory
- `TestMiniGameSystem_AwardReward_WithInvalidItemEntity` - Verifies graceful handling of invalid entity IDs
- `TestMiniGameSystem_AwardReward_EntityWithoutItemComponent` - Verifies graceful handling of entities without ItemComponent

---

### Gap A4: Terrain Manipulation Spell Effects ✅ COMPLETED

**File:** `pkg/engine/spell_effect_system.go`  
**Gap:** TerrainManipulation effects lack terrain system integration

**Status:** RESOLVED (2025-12-22)

**Implementation:**
1. Added `terrainModificationSystem` field to `SpellEffectSystem` struct
2. Added `SetTerrainModificationSystem()` setter method for dependency injection
3. Added `TerrainModifierType` enum with `TerrainModifierCreateWall`, `TerrainModifierDigTunnel`, `TerrainModifierCreatePit`
4. Implemented `executeTerrainManipulation()` to call terrain modification system methods based on effect parameters
5. Added new methods to `TerrainModificationSystem`: `SetTileAtWorldPosition()`, `SetTilesInArea()`, `setTile()`

**Integration Points:**
1. `pkg/engine/spell_effect_system.go` - SpellEffectSystem now integrates with TerrainModificationSystem
2. `pkg/engine/terrain_modification_system.go` - New methods for setting tile types at positions

**Tests Added:**
- `TestSpellEffectSystem_SetTerrainModificationSystem` - Verifies setter
- `TestSpellEffectSystem_ExecuteTerrainManipulation_NoTerrainSystem` - Verifies graceful degradation
- `TestSpellEffectSystem_ExecuteTerrainManipulation_CreateWall` - Verifies wall creation
- `TestSpellEffectSystem_ExecuteTerrainManipulation_DigTunnel` - Verifies tunnel digging
- `TestSpellEffectSystem_ExecuteTerrainManipulation_CreatePit` - Verifies pit creation
- `TestSpellEffectSystem_ExecuteTerrainManipulation_OnlyExecutesOnce` - Verifies single execution
- `TestSpellEffectSystem_ExecuteTerrainManipulation_AreaEffect` - Verifies area effects
- `TestTerrainModifierType_Values` - Verifies enum values
- `TestTerrainModificationSystem_SetTileAtWorldPosition` - Verifies single tile setting
- `TestTerrainModificationSystem_SetTilesInArea` - Verifies area tile setting
- `TestTerrainModificationSystem_setTile` - Verifies tile type mapping

---

### Gap A5: Vehicle Weapon Projectile Spawning ✅ COMPLETED

**File:** `pkg/engine/vehicle_combat_system.go`  
**Gap:** Vehicle mounted weapons should spawn projectile entities

**Status:** RESOLVED (2025-12-23)

**Implementation:**
1. Added `WeaponProjectileSpeed` field to `VehicleCombatComponent` in `pkg/engine/vehicle_combat_component.go`
2. Updated `NewVehicleCombatComponent()` to set default projectile speed (300.0 pixels/second)
3. Implemented `spawnWeaponProjectile()` method to create actual projectile entities using `ProjectileSystem.SpawnProjectile()`
4. Added `weaponTypeToProjectileType()` helper to map weapon types ("Cannon", "MachineGun", "Laser", "Magic", "Ballista") to projectile visual types

**Integration Points:**
1. `pkg/engine/vehicle_combat_component.go` - Added `WeaponProjectileSpeed` field
2. `pkg/engine/vehicle_combat_system.go` - Implemented projectile spawning with proper velocity, damage, and lifetime
3. `pkg/engine/projectile_system.go` - Uses existing `SpawnProjectile()` method

**Tests Added:**
- `TestVehicleCombatSystem_NewVehicleCombatSystem` - Verifies system creation
- `TestVehicleCombatSystem_NewVehicleCombatSystem_NilWorld` - Verifies graceful handling of nil world
- `TestVehicleCombatSystem_SetProjectileSystem` - Verifies setter
- `TestVehicleCombatSystem_SetCombatSystem` - Verifies setter
- `TestVehicleCombatSystem_SpawnWeaponProjectile_WithProjectileSystem` - Verifies projectile creation
- `TestVehicleCombatSystem_SpawnWeaponProjectile_NoProjectileSystem` - Verifies graceful degradation
- `TestVehicleCombatSystem_SpawnWeaponProjectile_VelocityCalculation` - Tests all 4 cardinal directions
- `TestVehicleCombatSystem_WeaponTypeToProjectileType` - Tests all weapon type mappings
- `TestVehicleCombatSystem_SpawnWeaponProjectile_ProjectileProperties` - Verifies damage, type, owner, lifetime
- `TestVehicleCombatComponent_WeaponProjectileSpeed` - Tests new field
- `TestVehicleCombatSystem_SpawnWeaponProjectile_DefaultSpeed` - Tests default speed fallback

---

### Gap A6: Vehicle Terrain Hazard Damage ✅ COMPLETED

**File:** `pkg/engine/vehicle_durability_system.go`  
**Gap:** Vehicle durability system doesn't check terrain hazards

**Status:** RESOLVED (2025-12-23)

**Implementation:**
1. Added `terrainCollisionChecker` field to `VehicleDurabilitySystem` struct
2. Added `SetTerrainCollisionChecker()` setter method for dependency injection
3. Added terrain hazard damage constants: `LavaFlowDamagePerSecond` (10), `DeepWaterDamagePerSecond` (5), `PitDamagePerSecond` (15)
4. Implemented `calculateTerrainHazardDamage()` to determine damage based on tile type and vehicle capabilities
5. Updated `checkEnvironmentalDamage()` to query terrain tiles and apply appropriate damage

**Integration Points:**
1. `pkg/engine/vehicle_durability_system.go` - VehicleDurabilitySystem now integrates with TerrainCollisionChecker
2. `pkg/engine/terrain_collision_system.go` - Uses existing `worldToTileCoords()` and terrain data access

**Vehicle Immunities:**
- Gliders are immune to pit damage (they fly over)
- Boats and gliders are immune to deep water damage
- All vehicles take lava damage (no immunity)

**Tests Added:**
- `TestVehicleDurabilitySystem_NewVehicleDurabilitySystem` - Verifies system creation
- `TestVehicleDurabilitySystem_NewVehicleDurabilitySystem_NilWorld` - Verifies graceful handling of nil world
- `TestVehicleDurabilitySystem_SetTerrainCollisionChecker` - Verifies setter
- `TestVehicleDurabilitySystem_SetTerrainCollisionChecker_Nil` - Verifies nil handling
- `TestVehicleDurabilitySystem_CheckEnvironmentalDamage_LavaFlow` - Verifies lava damage
- `TestVehicleDurabilitySystem_CheckEnvironmentalDamage_Pit` - Verifies pit damage
- `TestVehicleDurabilitySystem_CheckEnvironmentalDamage_Pit_GliderImmune` - Verifies glider immunity to pits
- `TestVehicleDurabilitySystem_CheckEnvironmentalDamage_DeepWater` - Verifies deep water damage
- `TestVehicleDurabilitySystem_CheckEnvironmentalDamage_DeepWater_BoatImmune` - Verifies boat immunity
- `TestVehicleDurabilitySystem_CheckEnvironmentalDamage_NoTerrainChecker` - Verifies graceful degradation
- `TestVehicleDurabilitySystem_CheckEnvironmentalDamage_NoPosition` - Verifies handling of entities without position
- `TestVehicleDurabilitySystem_CheckEnvironmentalDamage_SafeTerrain` - Verifies no damage on safe tiles
- `TestVehicleDurabilitySystem_CalculateTerrainHazardDamage` - Tests all terrain/vehicle combinations
- `TestVehicleDurabilitySystem_Update_WithTerrainHazards` - Tests Update loop with hazards
- `TestVehicleDurabilitySystem_VehicleDestruction_FromTerrainHazard` - Verifies vehicle destruction
- `TestVehicleDurabilitySystem_TerrainHazardDamageConstants` - Verifies damage constants

---

## Phase 2: UI Integration (Week 2)

### Gap B1-B4: Gallery UI Complete Integration ✅ COMPLETED

**Files:** `pkg/engine/gallery_ui.go`, `pkg/engine/game.go`, `pkg/engine/input_system.go`

**Gaps:**
- B1: UI defined but gallery manager never connected
- B2: No input handling for image navigation
- B3: No visual representation of image gallery
- B4: Gallery passed but never stored or used

**Status:** RESOLVED (2025-12-23)

**Implementation:**
1. Added `galleryUI` field to `uiComponents` struct in `pkg/engine/game.go`
2. Created GalleryUI in `initializeUIComponents()` using `NewGalleryUI(screenWidth, screenHeight)`
3. Wired GalleryUI in `buildGameInstance()` to set `GalleryUI: ui.galleryUI`
4. Input handling was already implemented via `SetGalleryCallback()` in input_system.go
5. Visual representation already implemented in `gallery_ui.go` Draw() method
6. Gallery storage already implemented via `SetGallery()` method

**Integration Points:**
1. `pkg/engine/game.go` - GalleryUI now properly initialized and wired during game creation
2. `pkg/engine/input_system.go` - Keyboard 'G' key triggers gallery toggle (pre-existing)
3. `pkg/engine/gallery_ui.go` - Complete UI with navigation, metadata display, touch support (pre-existing)

---

### Gap B5-B9: Housing UI Complete Integration ✅ COMPLETED

**Files:** `pkg/world/housing/ui.go`, `pkg/engine/game.go`, `pkg/engine/input_system.go`, `pkg/engine/interfaces.go`

**Gaps:**
- B5: UI defined but managers never connected
- B6: No input handling for housing management
- B7: No visual representation of housing system
- B8: Player ID setter was empty stub
- B9: Managers passed but never stored or used

**Status:** RESOLVED (2025-12-23)

**Implementation:**
1. Added `housingUI` field to `uiComponents` struct in `pkg/engine/game.go`
2. Created HousingUI in `initializeUIComponents()` using `housing.NewHousingUI(screenWidth, screenHeight)`
3. Wired HousingUI in `buildGameInstance()` to set `HousingUI: ui.housingUI`
4. Added `HousingUIProvider` interface to `pkg/engine/interfaces.go` to avoid import cycles
5. Added `IsVisible()` method to `HousingUI` to satisfy interface requirements
6. Added `housingUI` field to `InputSystem` struct for ESC key handling
7. Added `SetHousingUI()` method to `InputSystem` for dependency injection
8. Added ESC key handling in `handleShopUIEscapeActions()` to close housing UI
9. SetManagers and SetPlayerID already implemented in ui.go (verified pre-existing)
10. Input handling already implemented via SetHousingCallback() (pre-existing)
11. Visual representation already implemented in Draw() method (pre-existing)

**Integration Points:**
1. `pkg/engine/game.go` - HousingUI now properly initialized and wired during game creation
2. `pkg/engine/input_system.go` - Keyboard 'H' key triggers toggle, ESC key closes panel
3. `pkg/engine/interfaces.go` - HousingUIProvider interface for input system integration
4. `pkg/world/housing/ui.go` - Complete UI with plot display, menu navigation, manager connections

---

### Gap B10-B13: V8.0 UI System Fields and Callbacks ✅ COMPLETED (Partial)

**Files:** `pkg/engine/game.go`, `pkg/engine/input_system.go`

**Status:** RESOLVED (2025-12-23) - GuildUI integration completed, BlueprintUI deferred (not yet created)

**Implementation:**
1. Added `KeyGuild ebiten.Key` field to InputSystem struct (mapped to 'O' key, changed from 'U' to avoid Achievements conflict)
2. Added `onGuildOpen func()` callback to InputSystem struct
3. Added `SetGuildCallback()` method to InputSystem for dependency injection
4. Added `guildUI *GuildUI` field to uiComponents struct
5. Initialized GuildUI in `initializeUIComponents()` with nil GuildSystem (defensive programming)
6. Wired GuildUI to game instance in `buildGameInstance()`
7. Added guild callback to `setupOptionalUICallbacks()`
8. Added GuildUI.Draw() call in Draw() method
9. Added defensive nil check in GuildUI.Draw() for missing GuildSystem
10. Guild key ('O') handling already implemented via callback system

**Integration Points:**
1. `pkg/engine/input_system.go` - KeyGuild field, onGuildOpen callback, SetGuildCallback() method
2. `pkg/engine/game.go` - GuildUI initialization and wiring to game loop
3. `pkg/engine/guild_ui.go` - Defensive nil check for GuildSystem

**Tests Added:**
- `TestInputSystem_SetGuildCallback` - Verifies callback setter
- `TestInputSystem_SetGuildCallback_NilCallback` - Verifies nil rejection
- `TestInputSystem_KeyGuild_Initialization` - Verifies KeyGuild initialization
- `TestGuildUI_NilGuildSystem` - Verifies graceful handling of nil GuildSystem
- `TestGuildUI_Toggle` - Verifies toggle functionality
- `TestGuildUI_IsVisible` - Verifies visibility checking

**Completed Fields:**
```go
// In EbitenGame struct: ✅
HousingUI  *housing.HousingUI  // ✅ Already completed in Gap B5-B9
GalleryUI  *GalleryUI          // ✅ Already completed in Gap B1-B4
GuildUI    *GuildUI            // ✅ Completed in this gap
TerritoryUI *TerritoryUI       // ✅ Already existed in struct

// In InputSystem struct: ✅
onHousingOpen   func() // ✅ Already completed in Gap B5-B9
onGalleryOpen   func() // ✅ Already completed in Gap B1-B4
onGuildOpen     func() // ✅ Completed in this gap
onTerritoryOpen func() // ✅ Already existed
```

**Deferred:**
- `BlueprintUI` field and callback - BlueprintUI struct does not exist yet, would require creation
- `onBlueprintOpen` callback - Deferred until BlueprintUI is created

---

### Gap B14: Bookshelf Dialog Provider ✅ COMPLETED (Pre-existing)

**File:** `pkg/engine/book_spawning.go`

**Gap:** Bookshelves reuse merchant dialog instead of dedicated bookshelf dialog

**Status:** RESOLVED (Pre-existing implementation verified 2025-12-23)

**Implementation:**
The `NewBookshelfDialogProvider()` function already exists and creates book-specific dialog providers using Markov chain generation instead of static dialog entries. This provides dynamic, genre-appropriate book descriptions rather than fixed text.

```go
func NewBookshelfDialogProvider(bookCount int, seed int64, genreID string) DialogProvider {
    bookshelfName := fmt.Sprintf("Bookshelf (%d books)", bookCount)
    bookPersonality := dialog.NewPersonality(dialog.PersonalityScholarly)
    provider := NewMarkovDialogProvider(seed, genreID, bookshelfName, bookPersonality)
    return provider
}
```

**Verification:**
- Function exists at `pkg/engine/book_spawning.go:473`
- Used in `createBookshelfEntity()` at line 325
- Creates DialogProvider with scholarly personality for bookshelf interactions
- Provides procedurally generated dialog text based on genre and seed

**Note:** The implementation uses MarkovDialogProvider for dynamic text generation rather than static DialogEntry structs suggested in the original plan. This provides richer, more varied interactions.

---

### Gap B15: Story Fragment Quest Unlocking ✅ COMPLETED

**File:** `pkg/engine/discovery_system.go`

**Gap:** DiscoverySystem needs hook to QuestGeneration for unlocking story-based quests

**Status:** RESOLVED (2025-12-24)

**Implementation:**
1. Added `questGenerator` field of type `QuestGeneratorInterface` to `DiscoverySystem` struct
2. Added `genreID` and `seed` fields to support quest generation parameters
3. Added `SetQuestGenerator()` method for dependency injection with genreID and seed parameters
4. Implemented actual quest generation in `unlockStoryQuests()` method using quest generator callback
5. Quest generation uses series-specific seed and generates exploration quests with custom titles/descriptions
6. Graceful degradation when quest generator is not configured or player lacks quest tracker

**Integration Points:**
1. `pkg/engine/discovery_system.go` - DiscoverySystem now integrates with QuestGenerator
2. `pkg/engine/quest_tracker.go` - Uses existing `UnlockStoryQuests()` callback mechanism
3. `pkg/procgen/quest/generator.go` - Uses existing `Generate()` method with proper parameters

**Quest Generation Details:**
- Generates quests with ID format: `story-{seriesID}`
- Quest name: `Investigate: {seriesID}`
- Quest description: Custom text referencing completed story fragments
- Quest type: Exploration quests (TypeExplore)
- Difficulty: Medium (0.5) for story quests
- Depth: 1 for all story quests

**Tests Added:**
- `TestDiscoverySystem_SetQuestGenerator` - Verifies setter with valid generator
- `TestDiscoverySystem_SetQuestGenerator_NilGenerator` - Verifies nil generator handling
- `TestDiscoverySystem_UnlockStoryQuests_WithGenerator` - Verifies quest generation and unlocking
- `TestDiscoverySystem_UnlockStoryQuests_NoGenerator` - Verifies graceful degradation without generator
- `TestDiscoverySystem_UnlockStoryQuests_NoQuestTracker` - Verifies handling of missing quest tracker
- `TestDiscoverySystem_UnlockStoryQuests_GeneratorError` - Verifies error handling from generator
- `TestDiscoverySystem_UnlockStoryQuests_EmptyQuestList` - Verifies handling of empty quest results
- `TestDiscoverySystem_UnlockStoryQuests_IntegrationWithDiscovery` - Full integration test with discovery flow
- `TestDiscoverySystem_UnlockStoryQuests_DuplicateUnlock` - Verifies no duplicate quest unlocking

---

## Phase 3: Data Persistence (Week 3)

### Gap D1-D4: V8/V9 Feature Persistence ✅ COMPLETED (Pre-existing)

**File:** `pkg/saveload/types.go`

**Gap:** Housing, trust scores, reputation, guild membership, vehicles, companions not persisted

**Status:** RESOLVED (Pre-existing implementation verified 2025-12-24)

**Implementation:**
The following types and fields were already implemented in `pkg/saveload/types.go`:

1. **HousingPlotData** (lines 287-295) - Complete housing plot persistence with furniture
2. **GuildMembershipData** (lines 307-313) - Guild membership with rank and permissions
3. **VehicleData** (lines 316-328) - Vehicle state with health, fuel, durability
4. **CompanionData** (lines 331-346) - Companion stats, loyalty, equipment, skills

**PlayerState Fields Added** (lines 74-87):
- `OwnedPlots []HousingPlotData` - Housing ownership
- `TrustScores map[string]float64` - Player trust scores
- `ReputationScores map[string]int` - Reputation by category
- `GuildMembership *GuildMembershipData` - Guild membership
- `OwnedVehicles []VehicleData` - Owned vehicles
- `ActiveCompanions []CompanionData` - Active companions

**Verification:**
- All types properly tagged with JSON serialization
- Integration fix comments present in code
- Tests pass for PlayerState serialization

---

### Gap D5-D6: World State Persistence ✅ COMPLETED (Pre-existing)

**File:** `pkg/saveload/types.go`

**Gap:** Guild halls, territory control, global events not persisted in world state

**Status:** RESOLVED (Pre-existing implementation verified 2025-12-24)

**Implementation:**
The following types and fields were already implemented in `pkg/saveload/types.go`:

1. **GuildHallData** (lines 471-480) - Guild hall placement with rooms and tier
2. **RoomData** (lines 483-491) - Individual room data within guild halls
3. **WorldEventData** (lines 494-500) - Active meta-game events with state

**WorldState Fields Added** (lines 438-444):
- `GuildHalls []GuildHallData` - Guild hall placements
- `TerritoryControl map[string]string` - Zone ownership (ZoneID -> GuildID)
- `ActiveEvents []WorldEventData` - Active world events

**Additional Features:**
- `LivingWorldMemory *LivingWorldMemoryData` (line 448) - Persistent world memory for cities and NPCs
- Integration fix comments present for Phase 51.2 and Phase 42

**Verification:**
- All types properly tagged with JSON serialization
- WorldState structure includes all V8/V9 persistence features
- Tests pass for saveload package

---

## Phase 4: Gameplay Integration (Week 4)

### Gap F1: Party Chat Delivery ✅ COMPLETED

**File:** `pkg/engine/chat_system.go`

**Gap:** Party channel requires PartyComponent to filter recipients

**Status:** RESOLVED (2025-12-24)

**Implementation:**
1. Added `PartyComponent` struct to `pkg/engine/chat_trade_components.go` with:
   - `PartyID` - unique party identifier
   - `MemberIDs` - slice of entity IDs for all party members
   - `LeaderID` - entity ID of the party leader
2. Added `NewPartyComponent()` constructor with leader initialization
3. Added member management methods: `AddMember()`, `RemoveMember()`, `IsMember()`, `IsLeader()`
4. Updated `ChatSystem.deliverToParty()` to use PartyComponent for filtering:
   - Gets sender's PartyComponent
   - Iterates through MemberIDs only (not all entities)
   - Delivers message only to party members with active ChatParty channel
   - Gracefully handles missing party component (sender not in party)
5. Removed INTEGRATION FIX comment from deliverToParty implementation

**Integration Points:**
1. `pkg/engine/chat_trade_components.go` - Added PartyComponent with member management
2. `pkg/engine/chat_system.go` - deliverToParty() now filters by party membership
3. Party chat now properly scoped to party members instead of all subscribed entities

**Tests Added:**
- `TestNewPartyComponent` - Verifies component creation with leader
- `TestPartyComponentAddMember` - Tests member addition with idempotency
- `TestPartyComponentRemoveMember` - Tests member removal including non-existent members
- `TestPartyComponentIsMember` - Tests membership checking
- `TestPartyComponentIsLeader` - Tests leader checking
- `TestPartyComponentMultipleMemberOperations` - Tests multiple member operations
- `TestChatSystem_DeliverMessage_Party` - Verifies party-only message delivery
- `TestChatSystem_DeliverMessage_Party_NoPartyComponent` - Verifies graceful handling without party
- `TestChatSystem_DeliverMessage_Party_MemberNotSubscribed` - Verifies channel subscription check
- `TestChatSystem_DeliverMessage_Party_MemberMissingChatComponent` - Verifies missing chat component handling
- `TestChatSystem_DeliverMessage_Party_MemberEntityNotFound` - Verifies non-existent entity handling

---

### Gap F2: Lock-Picking Mini-Game Integration ✅ COMPLETED

**File:** `pkg/engine/interaction_system.go`

**Gap:** ActionOpenLocked requires MiniGameSystem.StartGame() integration

**Status:** RESOLVED (2025-12-24)

**Implementation:**
1. Added `miniGameSystem` field to `InteractionSystem` struct
2. Added `SetMiniGameSystem()` setter method for dependency injection
3. Updated `handleOpenAction()` to accept player entity parameter
4. Implemented lock-picking mini-game start logic in `handleOpenAction()`:
   - Checks `RequiresLockPicking` flag on ContextActionComponent
   - Starts MiniGameLockPicking with normalized difficulty (0.0-1.0)
   - Stores locked entity ID in mini-game state for completion callback
   - Logs mini-game start with player ID and difficulty
5. Implemented `ProcessLockPickingCompletion()` method to handle game results:
   - Retrieves locked entity ID from mini-game state
   - On success: opens door/chest by changing ActionType to ActionClose and disabling RequiresLockPicking
   - On failure: door/chest remains locked
   - Graceful handling of missing entities or invalid state
6. Graceful degradation when mini-game system is not available

**Integration Points:**
1. `pkg/engine/interaction_system.go` - InteractionSystem now integrates with MiniGameSystem
2. `pkg/engine/minigame_system.go` - Uses existing `StartGame()`, `GetGameComponent()` methods
3. `pkg/engine/carriable_component.go` - Uses ContextActionComponent fields (RequiresLockPicking, LockDifficulty)

**Tests Added:**
- `TestInteractionSystem_SetMiniGameSystem` - Verifies setter
- `TestInteractionSystem_SetMiniGameSystem_Nil` - Verifies nil handling
- `TestInteractionSystem_HandleOpenAction_WithLockPicking` - Verifies mini-game start with proper state storage
- `TestInteractionSystem_HandleOpenAction_NoMiniGameSystem` - Verifies graceful degradation
- `TestInteractionSystem_HandleOpenAction_NormalDoor` - Verifies normal door opening (no lock-picking)
- `TestInteractionSystem_HandleOpenAction_DifficultyNormalization` - Tests difficulty clamping to 0.0-1.0 range
- `TestInteractionSystem_ProcessLockPickingCompletion_Success` - Verifies door opens on success
- `TestInteractionSystem_ProcessLockPickingCompletion_Failure` - Verifies door remains locked on failure
- `TestInteractionSystem_ProcessLockPickingCompletion_NoMiniGameSystem` - Verifies graceful degradation
- `TestInteractionSystem_ProcessLockPickingCompletion_NoGameComponent` - Verifies handling when no game active
- `TestInteractionSystem_ProcessLockPickingCompletion_WrongGameType` - Verifies handling of non-lock-picking games
- `TestInteractionSystem_ProcessLockPickingCompletion_InvalidEntityID` - Verifies handling of invalid entities

---

### Gap F3: Bookshelf Key Requirement ✅ COMPLETED

**File:** `pkg/engine/interaction_system.go`

**Gap:** Locked bookshelf interaction needs inventory check for key items

**Status:** RESOLVED (2025-12-23)

**Implementation:**
1. Added `playerHasItemByID()` helper method to check if player has item with specific ID
2. Updated `handleBookshelfRead()` to check for required key before allowing access
3. When bookshelf is locked and requires a key:
   - Checks player inventory for matching KeyItemID
   - If key found: unlocks bookshelf and allows browsing
   - If key not found: bookshelf remains locked, access denied
4. Proper logging for all scenarios (key found, key missing, already unlocked)

**Integration Points:**
1. `pkg/engine/interaction_system.go` - Added key checking logic in handleBookshelfRead
2. `pkg/engine/bookshelf_component.go` - Uses existing RequiresKey, KeyItemID, IsLocked fields
3. `pkg/engine/inventory_components.go` - Queries InventoryComponent.Items for matching item ID

**Tests Added:**
- `TestPlayerHasItemByID_WithMatchingItem` - Verifies item found when present
- `TestPlayerHasItemByID_WithoutMatchingItem` - Verifies item not found when absent
- `TestPlayerHasItemByID_EmptyInventory` - Tests empty inventory handling
- `TestPlayerHasItemByID_NoInventoryComponent` - Tests missing inventory component
- `TestPlayerHasItemByID_NilItem` - Tests nil item handling
- `TestHandleBookshelfRead_UnlockedBookshelf` - Tests normal unlocked access
- `TestHandleBookshelfRead_LockedWithoutKey` - Tests locked bookshelf blocks access without key
- `TestHandleBookshelfRead_LockedWithKey` - Tests locked bookshelf unlocks with correct key
- `TestHandleBookshelfRead_LockedWithWrongKey` - Tests wrong key doesn't unlock
- `TestHandleBookshelfRead_LockedNoKeyRequired` - Tests edge case of locked without key mechanism
- `TestHandleBookshelfRead_EmptyBookshelf` - Tests empty bookshelf handling
- `TestHandleBookshelfRead_UnlockThenBrowse` - Tests full unlock and browse flow
- `TestHandleBookshelfRead_MultipleKeys` - Tests finding correct key among multiple keys

---

### Gap F4: Radial Gradient Lighting ✅ COMPLETED

**File:** `pkg/engine/lighting_system.go`

**Gap:** Light rendering uses simple circle fill instead of proper radial gradients

**Status:** RESOLVED (2025-12-24)

**Implementation:**
1. Added `lightCacheKey` struct to cache gradients by both diameter and falloff type
2. Modified `getCachedLightCircle()` to accept falloff parameter and generate proper radial gradients
3. Implemented `calculateFalloffIntensity()` method supporting all falloff types:
   - `FalloffLinear`: intensity = 1 - distance (smooth linear decrease)
   - `FalloffQuadratic`: intensity = (1 - distance)^2 (sharper falloff)
   - `FalloffInverseSquare`: intensity = 1 / (1 + distance^2 * 4) (realistic physics-based)
   - `FalloffConstant`: intensity = 1.0 until edge (hard cutoff)
4. Updated `applyPointLight()` to pass falloff type to cache lookup
5. Gradient uses white color with alpha falloff, allowing ColorScale modulation at draw time

**Integration Points:**
1. `pkg/engine/lighting_system.go` - Gradient generation with proper falloff curves
2. Cache key includes both diameter and falloff type for efficient reuse
3. Maintains performance optimization with cached gradient textures

**Performance:**
- Zero performance degradation - still uses cached images
- Benchmark: ApplyPointLight ~504 ns/op (same as before)
- Multiple lights: ~6125 ns/op for 10 lights (same as before)
- Cache efficiency maintained with (diameter, falloff) composite key

**Tests Added:**
- `TestGetCachedLightCircle_Caching` - Verifies caching with falloff types
- `TestGetCachedLightCircle_GradientGeneration` - Verifies gradient image creation
- `TestGetCachedLightCircle_AllFalloffTypes` - Tests all 4 falloff types
- `TestCalculateFalloffIntensity_Linear` - Tests linear falloff calculation
- `TestCalculateFalloffIntensity_Quadratic` - Tests quadratic falloff calculation
- `TestCalculateFalloffIntensity_InverseSquare` - Tests inverse square falloff
- `TestCalculateFalloffIntensity_Constant` - Tests constant falloff
- `TestCalculateFalloffIntensity_UnknownType` - Tests fallback behavior
- `TestLightCacheKey` - Tests cache key struct equality

---

## Phase 5: Federation & Rewards (Week 5)

### Gap G1-G3: Guild Federation Package

**Files:** `pkg/network/federation/guild/`

**Gaps:**
- G1: Package did not exist
- G2: Guild manager missing for cross-server coordination
- G3: Guild federation types missing

**Resolution - Already Created:** These files already exist with stub implementations. Complete them:

```go
// manager.go - Add full implementation
func (gm *GuildManager) SyncGuildState(guildID string) error {
    guild, ok := gm.guilds[guildID]
    if !ok {
        return fmt.Errorf("guild not found: %s", guildID)
    }
    
    // Broadcast to all federated servers
    for _, serverID := range gm.federatedServers {
        msg := GuildSyncMessage{
            Type:    MsgTypeGuildSync,
            GuildID: guildID,
            Guild:   guild,
        }
        gm.sendToServer(serverID, msg)
    }
    return nil
}

func (gm *GuildManager) HandleGuildMessage(msg GuildMessage) error {
    switch msg.Type {
    case MsgTypeGuildSync:
        return gm.handleGuildSync(msg)
    case MsgTypeMemberJoin:
        return gm.handleMemberJoin(msg)
    case MsgTypeMemberLeave:
        return gm.handleMemberLeave(msg)
    case MsgTypeTerritoryChange:
        return gm.handleTerritoryChange(msg)
    default:
        return fmt.Errorf("unknown message type: %v", msg.Type)
    }
}
```

---

### Gap E1-E2: Mini-Game Reward Items ✅ COMPLETED

**File:** `pkg/engine/minigame_system.go`

**Gaps:**
- E1: Mini-game rewards should include procedurally generated items
- E2: EndGame() doesn't add items to inventory (already completed in Gap A3)

**Status:** RESOLVED (2025-12-24)

**Implementation:**
1. Added `itemGenerator *item.ItemGenerator` field to `MiniGameSystem` struct
2. Added `genreID string` field to support genre-specific item generation
3. Added `SetItemGenerator()` method for dependency injection
4. Added `SetGenreID()` method for genre configuration
5. Modified `generateReward()` to create procedural items:
   - Determines number of items based on difficulty (1 item for low, 2 for medium, 3 for high)
   - Uses ItemGenerator.Generate() with proper parameters
   - Creates item entities with ItemComponent for each generated item
   - Adds entity IDs to reward.Items array
   - Gracefully degrades when ItemGenerator is not configured

**Integration Points:**
1. `pkg/engine/minigame_system.go` - Added item generation to reward creation
2. `pkg/procgen/item/generator.go` - Uses existing Generate() method with difficulty scaling
3. `pkg/engine/ecs.go` - Creates entities for reward items
4. Gap A3 already handles adding reward.Items to player inventory via awardReward()

**Reward Scaling:**
- Difficulty 0.0-0.7: 1 item
- Difficulty 0.7-0.9: 2 items
- Difficulty 0.9-1.0: 3 items

**Tests Added:**
- `TestMiniGameSystem_SetItemGenerator` - Verifies item generator setter
- `TestMiniGameSystem_SetItemGenerator_Nil` - Verifies nil handling for graceful degradation
- `TestMiniGameSystem_SetGenreID` - Verifies genre ID setter
- `TestMiniGameSystem_GenerateReward_WithItemGenerator` - Tests item generation at all difficulty levels
- `TestMiniGameSystem_GenerateReward_WithoutItemGenerator` - Tests graceful degradation without generator
- `TestMiniGameSystem_GenerateReward_ItemsDeterministic` - Verifies deterministic item generation
- `TestMiniGameSystem_GenerateReward_ItemsCreatedAsEntities` - Verifies items are proper entities with ItemComponent
- `TestMiniGameSystem_EndGame_WithGeneratedItems` - Full integration test with reward items added to inventory


---

## Verification Checklist

After completing each phase, verify:

### Build Verification
```bash
go build ./cmd/client && go build ./cmd/server
go test ./pkg/... -cover
```

### Runtime Verification
- [ ] Client starts without panic
- [ ] All UI panels open with correct keys (I, C, K, J, M, R, G, H, O)
- [x] Gallery shows images when available ✅ (Gap B1-B4 completed 2025-12-23)
- [x] Housing UI shows player plots ✅ (Gap B5-B9 completed 2025-12-23)
- [x] Guild UI opens with O key ✅ (Gap B10-B13 completed 2025-12-23)
- [x] Mini-games award items to inventory ✅ (Gap A3 completed 2025-12-22)
- [x] Mini-games generate procedural reward items ✅ (Gap E1-E2 completed 2025-12-24)
- [x] Vehicles take terrain damage ✅ (Gap A6 completed 2025-12-23)
- [x] Locked bookshelves require key items ✅ (Gap F3 completed 2025-12-23)
- [x] Lock-picking mini-game starts for locked containers ✅ (Gap F2 completed 2025-12-24)
- [x] Party chat delivers to party members ✅ (Gap F1 completed 2025-12-24)
- [x] Spell effects modify terrain ✅ (Gap A4 completed 2025-12-22)
- [x] Radial gradient lighting with proper falloff curves ✅ (Gap F4 completed 2025-12-24)

### Persistence Verification
- [ ] Save game includes V8/V9 data
- [ ] Load game restores housing plots
- [ ] Guild membership persists across sessions
- [ ] Trust scores persist and decay correctly

---

## Timeline Summary

| Phase | Focus | Duration | Gaps Resolved |
|-------|-------|----------|---------------|
| 1 | System Interfaces | Week 1 | 6 (A1-A6) |
| 2 | UI Integration | Week 2 | 15 (B1-B15) |
| 3 | Data Persistence | Week 3 | 5 (D1-D5) |
| 4 | Gameplay Integration | Week 4 | 4 (F1-F4) |
| 5 | Federation & Rewards | Week 5 | 4 (G1-G3, E1) |

**Total:** 34 integration gaps resolved over 5 weeks

---

## Success Criteria

- [ ] All 34 `INTEGRATION FIX` comments addressed
- [ ] Zero runtime panics related to nil components
- [ ] All V8.0 features accessible via UI
- [ ] Complete save/load cycle for all V8/V9 features
- [ ] All mini-games award items correctly
- [ ] Guild federation synchronizes across servers
- [ ] 100% feature parity between desktop and mobile

---

*End of Integration Plan*
