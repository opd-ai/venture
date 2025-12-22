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

### Gap A1: AdaptiveMusicSystem Interface Implementation

**File:** `pkg/engine/audio_manager.go`  
**Gap:** AudioManager did not implement audio.AdaptiveMusicSystem interface required by MusicTriggerSystem

**Resolution:**
```go
// Add to AudioManager struct:
func (am *AudioManager) GetCurrentIntensity() int {
    if am.soundtrackSystem != nil {
        return am.soundtrackSystem.GetIntensity()
    }
    return 0
}

func (am *AudioManager) SetIntensity(intensity int) {
    if am.soundtrackSystem != nil {
        am.soundtrackSystem.SetIntensity(intensity)
    }
}

func (am *AudioManager) TriggerMusicEvent(event string) {
    if am.soundtrackSystem != nil {
        am.soundtrackSystem.HandleEvent(event)
    }
}
```

**Verification:**
```bash
grep -n "AdaptiveMusicSystem" pkg/engine/audio_manager.go
go test ./pkg/engine/... -run AudioManager
```

---

### Gap A2: Consumable Spell Effect Activation

**File:** `pkg/engine/inventory_system.go`  
**Gap:** Using consumable items doesn't trigger spell effects (potions, scrolls)

**Resolution:**
```go
// In UseItem() method, after consuming the item:
func (is *InventorySystem) UseItem(entity *Entity, itemID uint64) error {
    // ... existing validation ...
    
    // NEW: Trigger spell effect for consumables
    if item.Type == ItemTypeConsumable && item.SpellID != "" {
        if spellComp, ok := entity.GetComponent("spell_caster"); ok && spellComp != nil {
            caster := spellComp.(*SpellCasterComponent)
            is.world.GetSystem("spell_effect").(*SpellEffectSystem).CastSpell(
                entity, caster, item.SpellID, item.SpellLevel,
            )
        }
    }
    
    // ... rest of consumption logic ...
}
```

**Integration Points:**
1. `pkg/engine/inventory_system.go` - Add spell effect trigger
2. `pkg/engine/spell_effect_system.go` - Ensure CastSpell is public

---

### Gap A3: Mini-Game Item Rewards

**File:** `pkg/engine/minigame_system.go`  
**Gap:** EndGame() doesn't add reward.Items to player inventory

**Resolution:**
```go
// In endGame() method:
func (ms *MiniGameSystem) endGame(entity *Entity, success bool) {
    // ... existing gold/XP reward logic ...
    
    // NEW: Add item rewards to inventory
    if len(reward.Items) > 0 {
        if invComp, ok := entity.GetComponent("inventory"); ok && invComp != nil {
            inv := invComp.(*InventoryComponent)
            for _, itemID := range reward.Items {
                inv.AddItem(itemID)
            }
            ms.logger.WithField("items", len(reward.Items)).Debug("mini-game items awarded")
        }
    }
}
```

---

### Gap A4: Terrain Manipulation Spell Effects

**File:** `pkg/engine/spell_effect_system.go`  
**Gap:** TerrainManipulation effects lack terrain system integration

**Resolution:**
```go
// Add terrain system reference:
type SpellEffectSystem struct {
    world         *World
    terrainSystem *TerrainSystem // NEW
    // ...
}

// In handleTerrainManipulation():
func (s *SpellEffectSystem) handleTerrainManipulation(effect SpellEffect, caster *Entity) {
    pos, _ := caster.GetComponent("position")
    if pos == nil {
        return
    }
    position := pos.(*PositionComponent)
    
    switch effect.SubType {
    case "create_wall":
        s.terrainSystem.SetTile(int(position.X), int(position.Y), TileWall)
    case "dig_tunnel":
        s.terrainSystem.SetTile(int(position.X), int(position.Y), TileFloor)
    case "create_pit":
        s.terrainSystem.SetTile(int(position.X), int(position.Y), TilePit)
    }
}
```

---

### Gap A5: Vehicle Weapon Projectile Spawning

**File:** `pkg/engine/vehicle_combat_system.go`  
**Gap:** Vehicle mounted weapons should spawn projectile entities

**Resolution:**
```go
// Replace placeholder with actual projectile spawning:
func (vcs *VehicleCombatSystem) fireWeapon(vehicle *Entity, weapon *VehicleWeaponComponent) {
    pos, _ := vehicle.GetComponent("position")
    if pos == nil {
        return
    }
    position := pos.(*PositionComponent)
    
    // Create projectile entity
    projectile := vcs.world.CreateEntity()
    projectile.AddComponent(&PositionComponent{X: position.X, Y: position.Y})
    projectile.AddComponent(&VelocityComponent{
        X: weapon.Direction.X * weapon.ProjectileSpeed,
        Y: weapon.Direction.Y * weapon.ProjectileSpeed,
    })
    projectile.AddComponent(&ProjectileComponent{
        Damage:   weapon.Damage,
        OwnerID:  vehicle.ID,
        Lifetime: 5.0,
    })
    projectile.AddComponent(&ColliderComponent{
        Width:  8,
        Height: 8,
    })
}
```

---

### Gap A6: Vehicle Terrain Hazard Damage

**File:** `pkg/engine/vehicle_durability_system.go`  
**Gap:** Vehicle durability system doesn't check terrain hazards

**Resolution:**
```go
// Add terrain hazard check in Update():
func (vds *VehicleDurabilitySystem) Update(entities []*Entity, deltaTime float64) {
    for _, entity := range entities {
        durability, ok := entity.GetComponent("vehicle_durability")
        if !ok || durability == nil {
            continue
        }
        dur := durability.(*VehicleDurabilityComponent)
        
        pos, _ := entity.GetComponent("position")
        if pos == nil {
            continue
        }
        position := pos.(*PositionComponent)
        
        // NEW: Check terrain hazards
        tile := vds.terrainSystem.GetTile(int(position.X), int(position.Y))
        switch tile {
        case TileLava:
            dur.Current -= 10 * deltaTime // 10 damage per second
        case TileSpikes:
            dur.Current -= 5 * deltaTime // 5 damage per second
        case TileAcid:
            dur.Current -= 15 * deltaTime // 15 damage per second
        }
        
        if dur.Current < 0 {
            dur.Current = 0
            vds.destroyVehicle(entity)
        }
    }
}
```

---

## Phase 2: UI Integration (Week 2)

### Gap B1-B4: Gallery UI Complete Integration

**Files:** `pkg/engine/gallery_ui.go`, `pkg/engine/game.go`, `pkg/engine/input_system.go`

**Gaps:**
- B1: UI defined but gallery manager never connected
- B2: No input handling for image navigation
- B3: No visual representation of image gallery
- B4: Gallery passed but never stored or used

**Resolution - Step 1:** Connect gallery manager in game.go
```go
// In EbitenGame initialization:
func (g *EbitenGame) initializeUI() {
    // ... existing UI init ...
    
    // V8.0 Gallery UI
    g.GalleryUI = NewGalleryUI(g.ScreenWidth, g.ScreenHeight)
    if g.SocialPersistence != nil {
        g.GalleryUI.SetGallery(g.SocialPersistence.GetImageGallery())
    }
}
```

**Resolution - Step 2:** Add input handling in input_system.go
```go
// In handleUIKeys():
func (is *InputSystem) handleUIKeys() {
    // ... existing key handlers ...
    
    if inpututil.IsKeyJustPressed(ebiten.KeyG) && is.onGalleryOpen != nil {
        is.onGalleryOpen()
    }
}

// Callback setter:
func (is *InputSystem) SetGalleryCallback(callback func()) {
    is.onGalleryOpen = callback
}
```

**Resolution - Step 3:** Wire callback in handlers.go
```go
// In initializeUICallbacks():
inputSystem.SetGalleryCallback(func() {
    game.GalleryUI.Toggle()
})
```

---

### Gap B5-B9: Housing UI Complete Integration

**Files:** `pkg/world/housing/ui.go`, `pkg/engine/game.go`, `pkg/engine/input_system.go`

**Gaps:**
- B5: UI defined but managers never connected
- B6: No input handling for housing management
- B7: No visual representation of housing system
- B8: Player ID setter was empty stub
- B9: Managers passed but never stored or used

**Resolution - Step 1:** Connect housing manager
```go
// In HousingUI:
func (h *HousingUI) SetManager(manager *Manager) {
    h.manager = manager
    h.refreshPlots()
}

func (h *HousingUI) SetPlayerID(playerID string) {
    h.playerID = playerID
    h.refreshPlots()
}

func (h *HousingUI) refreshPlots() {
    if h.manager != nil && h.playerID != "" {
        h.plots = h.manager.GetPlayerPlots(h.playerID)
    }
}
```

**Resolution - Step 2:** Add key binding and callback
```go
// In input_system.go:
func (is *InputSystem) SetHousingCallback(callback func()) {
    is.onHousingOpen = callback
}

// In handlers.go:
inputSystem.SetHousingCallback(func() {
    game.HousingUI.Toggle()
})
```

---

### Gap B10-B13: V8.0 UI System Fields and Callbacks

**Files:** `pkg/engine/game.go`, `pkg/engine/input_system.go`

**Resolution:** Add all V8.0 UI field declarations
```go
// In EbitenGame struct:
type EbitenGame struct {
    // ... existing fields ...
    
    // V8.0 UI Systems
    HousingUI  *housing.HousingUI
    GalleryUI  *GalleryUI
    BlueprintUI *BlueprintUI
    GuildUI    *GuildUI
    TerritoryUI *TerritoryUI
}

// In InputSystem struct:
type InputSystem struct {
    // ... existing fields ...
    
    // V8.0 UI Callbacks
    onHousingOpen   func()
    onGalleryOpen   func()
    onBlueprintOpen func()
    onGuildOpen     func()
    onTerritoryOpen func()
}
```

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

### Gap F3: Bookshelf Key Requirement

**File:** `pkg/engine/interaction_system.go`

**Gap:** Locked bookshelf interaction needs inventory check for key items

**Resolution:**
```go
// In handleBookshelfInteraction():
func (is *InteractionSystem) handleBookshelfInteraction(entity *Entity, bookshelf *Entity) {
    bookComp, ok := bookshelf.GetComponent("bookshelf")
    if !ok || bookComp == nil {
        return
    }
    shelf := bookComp.(*BookshelfComponent)
    
    // Check if locked
    if shelf.RequiresKey != "" {
        if !is.playerHasItem(entity, shelf.RequiresKey) {
            is.showMessage(entity, "This bookshelf is locked. You need: "+shelf.RequiresKey)
            return
        }
    }
    
    // Open bookshelf dialog
    is.openBookshelfDialog(entity, shelf)
}

func (is *InteractionSystem) playerHasItem(entity *Entity, itemName string) bool {
    invComp, ok := entity.GetComponent("inventory")
    if !ok || invComp == nil {
        return false
    }
    inv := invComp.(*InventoryComponent)
    return inv.HasItem(itemName)
}
```

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
- [ ] All UI panels open with correct keys (I, C, K, J, M, R, G, H)
- [ ] Gallery shows images when available
- [ ] Housing UI shows player plots
- [ ] Mini-games award items to inventory
- [ ] Vehicles take terrain damage
- [ ] Lock-picking mini-game starts for locked containers
- [ ] Party chat delivers to party members
- [ ] Spell effects modify terrain

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
