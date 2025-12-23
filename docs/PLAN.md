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
1. Added `KeyGuild ebiten.Key` field to InputSystem struct (mapped to 'U' key)
2. Added `onGuildOpen func()` callback to InputSystem struct
3. Added `SetGuildCallback()` method to InputSystem for dependency injection
4. Added `guildUI *GuildUI` field to uiComponents struct
5. Initialized GuildUI in `initializeUIComponents()` with nil GuildSystem (defensive programming)
6. Wired GuildUI to game instance in `buildGameInstance()`
7. Added guild callback to `setupOptionalUICallbacks()`
8. Added GuildUI.Draw() call in Draw() method
9. Added defensive nil check in GuildUI.Draw() for missing GuildSystem
10. Guild key ('U') handling already implemented via callback system

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

### Gap B14: Bookshelf Dialog Provider

**File:** `pkg/engine/book_spawning.go`

**Gap:** Bookshelves reuse merchant dialog instead of dedicated bookshelf dialog

**Resolution:**
```go
// Create dedicated bookshelf dialog:
func NewBookshelfDialogProvider() *DialogProvider {
    return &DialogProvider{
        Type: DialogTypeBookshelf,
        Entries: []DialogEntry{
            {ID: "browse", Text: "Browse books...", Action: ActionBrowseBooks},
            {ID: "read", Text: "Read selected book", Action: ActionReadBook},
            {ID: "take", Text: "Take book", Action: ActionTakeBook},
            {ID: "close", Text: "Close bookshelf", Action: ActionClose},
        },
    }
}

// Use in bookshelf entity creation:
func SpawnBookshelf(world *World, x, y float64, books []BookComponent) *Entity {
    bookshelf := world.CreateEntity()
    bookshelf.AddComponent(&PositionComponent{X: x, Y: y})
    bookshelf.AddComponent(&InteractableComponent{
        Type: InteractableBookshelf,
        DialogProvider: NewBookshelfDialogProvider(),
    })
    bookshelf.AddComponent(&BookshelfComponent{Books: books})
    return bookshelf
}
```

---

### Gap B15: Story Fragment Quest Unlocking

**File:** `pkg/engine/discovery_system.go`

**Gap:** DiscoverySystem needs hook to QuestGeneration for unlocking story-based quests

**Resolution:**
```go
// Add quest system reference:
type DiscoverySystem struct {
    world       *World
    questSystem *QuestSystem // NEW
    // ...
}

// In discoverFragment():
func (ds *DiscoverySystem) discoverFragment(entity *Entity, fragment StoryFragment) {
    // ... existing discovery logic ...
    
    // NEW: Unlock related quests
    if fragment.UnlocksQuestID != "" {
        if ds.questSystem != nil {
            ds.questSystem.UnlockQuest(entity.ID, fragment.UnlocksQuestID)
            ds.logger.WithFields(logrus.Fields{
                "fragment": fragment.ID,
                "quest":    fragment.UnlocksQuestID,
            }).Info("story fragment unlocked quest")
        }
    }
}
```

---

## Phase 3: Data Persistence (Week 3)

### Gap D1-D4: V8/V9 Feature Persistence

**File:** `pkg/saveload/types.go`

**Gap:** Housing, trust scores, reputation, guild membership, vehicles, companions not persisted

**Resolution - Step 1:** Define persistence types
```go
// V8/V9 Player Data Types
type HousingPlotData struct {
    PlotID     string    `json:"plot_id"`
    OwnerID    string    `json:"owner_id"`
    Position   [2]float64 `json:"position"`
    Size       int       `json:"size"`
    BuildingID string    `json:"building_id"`
}

type TrustScoreData struct {
    PlayerID   string    `json:"player_id"`
    TargetID   string    `json:"target_id"`
    Score      float64   `json:"score"`
    LastUpdate time.Time `json:"last_update"`
}

type GuildMembershipData struct {
    GuildID    string    `json:"guild_id"`
    PlayerID   string    `json:"player_id"`
    Rank       string    `json:"rank"`
    JoinedAt   time.Time `json:"joined_at"`
}

type VehicleData struct {
    VehicleID  uint64    `json:"vehicle_id"`
    OwnerID    string    `json:"owner_id"`
    Type       string    `json:"type"`
    Durability float64   `json:"durability"`
    Position   [2]float64 `json:"position"`
}

type CompanionData struct {
    CompanionID uint64    `json:"companion_id"`
    OwnerID     string    `json:"owner_id"`
    Name        string    `json:"name"`
    Species     string    `json:"species"`
    Level       int       `json:"level"`
    Skills      []string  `json:"skills"`
}
```

**Resolution - Step 2:** Add to SaveData struct
```go
type SaveData struct {
    // ... existing fields ...
    
    // V8/V9 Features
    HousingPlots     []HousingPlotData     `json:"housing_plots"`
    TrustScores      []TrustScoreData      `json:"trust_scores"`
    GuildMemberships []GuildMembershipData `json:"guild_memberships"`
    Vehicles         []VehicleData         `json:"vehicles"`
    Companions       []CompanionData       `json:"companions"`
    Reputation       map[string]float64    `json:"reputation"`
}
```

---

### Gap D5-D6: World State Persistence

**File:** `pkg/saveload/types.go`

**Gap:** Guild halls, territory control, global events not persisted in world state

**Resolution:**
```go
// V8/V9 World Data Types
type GuildHallData struct {
    HallID     string     `json:"hall_id"`
    GuildID    string     `json:"guild_id"`
    Floors     int        `json:"floors"`
    Position   [2]float64 `json:"position"`
    Decorations []string  `json:"decorations"`
}

type TerritoryData struct {
    TerritoryID string     `json:"territory_id"`
    OwnerGuild  string     `json:"owner_guild"`
    ControlPct  float64    `json:"control_percentage"`
    Contested   bool       `json:"contested"`
    Resources   []string   `json:"resources"`
}

type GlobalEventData struct {
    EventID     string    `json:"event_id"`
    Type        string    `json:"type"`
    StartTime   time.Time `json:"start_time"`
    EndTime     time.Time `json:"end_time"`
    Participants []string `json:"participants"`
    Progress    float64   `json:"progress"`
}

type WorldData struct {
    // ... existing fields ...
    
    // V8/V9 Features
    GuildHalls    []GuildHallData    `json:"guild_halls"`
    Territories   []TerritoryData    `json:"territories"`
    GlobalEvents  []GlobalEventData  `json:"global_events"`
}
```

---

## Phase 4: Gameplay Integration (Week 4)

### Gap F1: Party Chat Delivery

**File:** `pkg/engine/chat_system.go`

**Gap:** Party channel requires PartyComponent to filter recipients

**Resolution:**
```go
// Add party member lookup:
func (cs *ChatSystem) deliverPartyMessage(sender *Entity, message ChatMessage) {
    partyComp, ok := sender.GetComponent("party")
    if !ok || partyComp == nil {
        cs.logger.Warn("party chat failed: sender not in party")
        return
    }
    party := partyComp.(*PartyComponent)
    
    // Deliver to all party members
    for _, memberID := range party.MemberIDs {
        member := cs.world.GetEntity(memberID)
        if member != nil {
            if chatComp, ok := member.GetComponent("chat_inbox"); ok && chatComp != nil {
                inbox := chatComp.(*ChatInboxComponent)
                inbox.AddMessage(message)
            }
        }
    }
}
```

---

### Gap F2: Lock-Picking Mini-Game Integration

**File:** `pkg/engine/interaction_system.go`

**Gap:** ActionOpenLocked requires MiniGameSystem.StartGame() integration

**Resolution:**
```go
// In handleLockedInteraction():
func (is *InteractionSystem) handleLockedInteraction(entity *Entity, target *Entity) {
    lockComp, ok := target.GetComponent("lockable")
    if !ok || lockComp == nil {
        return
    }
    lock := lockComp.(*LockableComponent)
    
    // Check for key in inventory first
    if is.playerHasKey(entity, lock.KeyID) {
        lock.IsLocked = false
        is.logger.Info("lock opened with key")
        return
    }
    
    // Start lock-picking mini-game
    if is.miniGameSystem != nil {
        difficulty := float64(lock.Difficulty) / 100.0
        is.miniGameSystem.StartGame(entity, MiniGameLockPick, MiniGameParams{
            Difficulty: difficulty,
            TimeLimit:  30.0,
            OnComplete: func(success bool) {
                if success {
                    lock.IsLocked = false
                    is.logger.Info("lock picked successfully")
                }
            },
        })
    }
}
```

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

### Gap F4: Radial Gradient Lighting

**File:** `pkg/engine/lighting_system.go`

**Gap:** Light rendering uses simple circle fill instead of proper radial gradients

**Resolution:**
```go
// Replace simple fill with radial gradient:
func (ls *LightingSystem) renderLight(screen *ebiten.Image, light LightComponent, pos PositionComponent) {
    // Calculate light bounds
    radius := light.Radius
    centerX, centerY := pos.X, pos.Y
    
    // Create gradient image
    gradientImg := ebiten.NewImage(int(radius*2), int(radius*2))
    
    for y := 0; y < int(radius*2); y++ {
        for x := 0; x < int(radius*2); x++ {
            dx := float64(x) - radius
            dy := float64(y) - radius
            distance := math.Sqrt(dx*dx + dy*dy)
            
            if distance <= radius {
                // Radial falloff
                intensity := 1.0 - (distance / radius)
                intensity = math.Pow(intensity, light.Falloff)
                
                r := uint8(float64(light.Color.R) * intensity)
                g := uint8(float64(light.Color.G) * intensity)
                b := uint8(float64(light.Color.B) * intensity)
                a := uint8(float64(light.Color.A) * intensity)
                
                gradientImg.Set(x, y, color.RGBA{r, g, b, a})
            }
        }
    }
    
    // Draw gradient at light position
    op := &ebiten.DrawImageOptions{}
    op.GeoM.Translate(centerX-radius, centerY-radius)
    op.CompositeMode = ebiten.CompositeModeSourceOver
    screen.DrawImage(gradientImg, op)
}
```

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

### Gap E1-E2: Mini-Game Reward Items

**File:** `pkg/engine/minigame_system.go`

**Gaps:**
- E1: Mini-game rewards should include procedurally generated items
- E2: EndGame() doesn't add items to inventory

**Resolution:**
```go
// In calculateReward():
func (ms *MiniGameSystem) calculateReward(gameType MiniGameType, difficulty float64, success bool) MiniGameReward {
    reward := MiniGameReward{
        Gold: int(100 * difficulty * float64(boolToInt(success))),
        XP:   int(50 * difficulty * float64(boolToInt(success))),
    }
    
    if success && ms.itemGenerator != nil {
        // Generate reward items based on difficulty
        numItems := 1
        if difficulty > 0.7 {
            numItems = 2
        }
        if difficulty > 0.9 {
            numItems = 3
        }
        
        for i := 0; i < numItems; i++ {
            item, err := ms.itemGenerator.Generate(ms.seed+int64(i), procgen.GenerationParams{
                Difficulty: difficulty,
                GenreID:    ms.genreID,
            })
            if err == nil {
                reward.Items = append(reward.Items, item.ID)
            }
        }
    }
    
    return reward
}
```

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
- [ ] All UI panels open with correct keys (I, C, K, J, M, R, G, H, U)
- [x] Gallery shows images when available ✅ (Gap B1-B4 completed 2025-12-23)
- [x] Housing UI shows player plots ✅ (Gap B5-B9 completed 2025-12-23)
- [x] Guild UI opens with U key ✅ (Gap B10-B13 completed 2025-12-23)
- [x] Mini-games award items to inventory ✅ (Gap A3 completed 2025-12-22)
- [x] Vehicles take terrain damage ✅ (Gap A6 completed 2025-12-23)
- [x] Locked bookshelves require key items ✅ (Gap F3 completed 2025-12-23)
- [ ] Lock-picking mini-game starts for locked containers
- [ ] Party chat delivers to party members
- [x] Spell effects modify terrain ✅ (Gap A4 completed 2025-12-22)

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
