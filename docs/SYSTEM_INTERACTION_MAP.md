# System Interaction Map

This document visualizes how Venture's game systems interact and depend on each other.

## System Categories

```
┌─────────────────────────────────────────────────────────────────┐
│                        VENTURE GAME SYSTEMS                      │
│                         (38 Total Systems)                       │
└─────────────────────────────────────────────────────────────────┘
        │
        ├── Core ECS (3) ──────────────────────── Movement, Collision, Spatial
        │
        ├── Combat (4) ───────────────────────── Combat, Player Combat, Status, Revival
        │
        ├── AI & Behavior (2) ────────────────── AI, Fire Propagation*
        │
        ├── Progression (3) ──────────────────── XP, Skills, Objectives
        │
        ├── Inventory & Items (4) ────────────── Inventory, Pickup, Use, Visuals*
        │
        ├── Magic & Spells (3) ───────────────── Casting, Player Casting, Mana
        │
        ├── Input (1) ────────────────────────── Input System
        │
        ├── Rendering (5) ────────────────────── Render, Terrain, Animation, Particles, Feedback
        │
        ├── Camera (1) ───────────────────────── Camera System
        │
        ├── Audio (1) ────────────────────────── Audio Manager
        │
        ├── UI (4) ───────────────────────────── HUD, Menu, Tutorial, Help
        │
        ├── Terrain (2) ──────────────────────── Construction*, Modification*
        │
        ├── Commerce (2) ─────────────────────── Commerce, Dialog
        │
        ├── Crafting (1) ─────────────────────── Crafting*
        │
        ├── Lighting (1) ─────────────────────── Dynamic Lighting*
        │
        └── Weather (1) ──────────────────────── Weather Effects*

* = Future feature (implemented but not yet integrated)
```

## System Execution Flow

### Game Loop Execution Order

```
┌────────────────────────────────────────────────────────────────┐
│                      FRAME START                                │
└────────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌────────────────────────────────────────────────────────────────┐
│ Phase 1: INPUT PROCESSING                                       │
│   └─ InputSystem: Captures keyboard/mouse input                │
└────────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌────────────────────────────────────────────────────────────────┐
│ Phase 2: PLAYER ACTIONS                                         │
│   ├─ PlayerCombatSystem: Space key → attack                    │
│   ├─ PlayerItemUseSystem: E key → use item                     │
│   └─ PlayerSpellCastingSystem: 1-5 keys → cast spell           │
└────────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌────────────────────────────────────────────────────────────────┐
│ Phase 3: PHYSICS                                                │
│   ├─ MovementSystem: velocity → position                       │
│   └─ CollisionSystem: AABB collision detection                 │
└────────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌────────────────────────────────────────────────────────────────┐
│ Phase 4: COMBAT RESOLUTION                                      │
│   ├─ CombatSystem: Damage calculation                          │
│   └─ StatusEffectSystem: DoT, buffs, debuffs                   │
└────────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌────────────────────────────────────────────────────────────────┐
│ Phase 5: AI & BEHAVIOR                                          │
│   ├─ RevivalSystem: Player revival mechanics                   │
│   └─ AISystem: Enemy decision making                           │
└────────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌────────────────────────────────────────────────────────────────┐
│ Phase 6: PROGRESSION                                            │
│   ├─ ProgressionSystem: XP and leveling                        │
│   └─ SkillProgressionSystem: Skill bonuses                     │
└────────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌────────────────────────────────────────────────────────────────┐
│ Phase 7: VISUAL EFFECTS                                         │
│   └─ VisualFeedbackSystem: Hit flash, tints                    │
└────────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌────────────────────────────────────────────────────────────────┐
│ Phase 8: AUDIO                                                  │
│   └─ AudioManagerSystem: Music context updates                 │
└────────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌────────────────────────────────────────────────────────────────┐
│ Phase 9: QUESTS & ITEMS                                         │
│   ├─ ObjectiveTrackerSystem: Quest progress                    │
│   └─ ItemPickupSystem: Auto-pickup items                       │
└────────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌────────────────────────────────────────────────────────────────┐
│ Phase 10: MAGIC                                                 │
│   ├─ SpellCastingSystem: Execute spell effects                 │
│   └─ ManaRegenSystem: Regenerate mana                          │
└────────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌────────────────────────────────────────────────────────────────┐
│ Phase 11: INVENTORY                                             │
│   └─ InventorySystem: Item management                          │
└────────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌────────────────────────────────────────────────────────────────┐
│ Phase 12: ANIMATION                                             │
│   └─ AnimationSystem: Update sprite frames                     │
└────────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌────────────────────────────────────────────────────────────────┐
│ Phase 13: UI OVERLAYS                                           │
│   ├─ TutorialSystem: Tutorial steps                            │
│   ├─ HelpSystem: Help screen                                   │
│   └─ DialogSystem: Dialog state (FIXED in audit)               │
└────────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌────────────────────────────────────────────────────────────────┐
│ Phase 14: PARTICLES                                             │
│   └─ ParticleSystem: Particle effects                          │
└────────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌────────────────────────────────────────────────────────────────┐
│ Phase 15: SPATIAL OPTIMIZATION                                  │
│   └─ SpatialPartitionSystem: Quadtree updates (every 60 frames)│
└────────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌────────────────────────────────────────────────────────────────┐
│                      FRAME END                                  │
└────────────────────────────────────────────────────────────────┘
```

## System Dependencies

### Critical Dependencies

```
┌──────────────────────────────────────────────────────────┐
│                  MOVEMENT ↔ COLLISION                     │
│                 (Bidirectional Reference)                 │
│                                                           │
│  MovementSystem.SetCollisionSystem(collisionSystem)      │
│  → Enables predictive collision detection                │
└──────────────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────────────┐
│           COLLISION → TERRAIN COLLISION CHECKER           │
│                                                           │
│  terrainChecker := NewTerrainCollisionChecker()          │
│  collisionSystem.SetTerrainChecker(terrainChecker)       │
│  → Enables entity-terrain collision                      │
└──────────────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────────────┐
│                COMBAT → CAMERA, PARTICLES                 │
│                                                           │
│  combatSystem.SetCamera(cameraSystem)                    │
│  → Enables screen shake on hit                           │
│                                                           │
│  combatSystem.SetParticleSystem(particleSystem)          │
│  → Enables hit effect particles                          │
└──────────────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────────────┐
│          RENDER → SPATIAL PARTITION (Culling)             │
│                                                           │
│  renderSystem.SetSpatialPartition(spatialSystem)         │
│  renderSystem.EnableCulling(true)                        │
│  → Viewport culling optimization                         │
└──────────────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────────────┐
│              INPUT → UI SYSTEMS & CALLBACKS               │
│                                                           │
│  inputSystem.SetHelpSystem(helpSystem)                   │
│  inputSystem.SetTutorialSystem(tutorialSystem)           │
│  inputSystem.SetQuickSaveCallback(saveFunc)              │
│  inputSystem.SetQuickLoadCallback(loadFunc)              │
│  inputSystem.SetInteractCallback(merchantFunc)           │
│  → Hotkey dispatch to various systems                    │
└──────────────────────────────────────────────────────────┘
```

## Component Query Patterns

### Pattern 1: Simple Component Query
```go
// Get all entities with position and velocity
entities := world.GetEntitiesWith("position", "velocity")
```

**Used by**: MovementSystem, CollisionSystem, AISystem

### Pattern 2: Component Filtering
```go
// Process entities with specific components
for _, entity := range entities {
    if posComp, ok := entity.GetComponent("position"); ok {
        if velComp, ok := entity.GetComponent("velocity"); ok {
            pos := posComp.(*PositionComponent)
            vel := velComp.(*VelocityComponent)
            // Process movement
        }
    }
}
```

**Used by**: Most ECS systems

### Pattern 3: Component Type Checking
```go
// Check if entity has required components
if !entity.HasComponent("input") {
    continue // Skip non-player entities
}
```

**Used by**: Player-specific systems (PlayerCombatSystem, PlayerItemUseSystem)

### Pattern 4: Service Pattern (No Component Query)
```go
// Direct method calls, no Update() loop
result, err := commerceSystem.BuyItem(playerID, merchantID, itemIndex)
```

**Used by**: CommerceSystem (called from UI, not World.Update)

## Client vs Server Distribution

### Client-Only Systems (20 systems)
```
Input Processing:
  • InputSystem
  • PlayerCombatSystem
  • PlayerItemUseSystem
  • PlayerSpellCastingSystem

Rendering:
  • EbitenRenderSystem
  • TerrainRenderSystem
  • AnimationSystem
  • ParticleSystem
  • VisualFeedbackSystem
  • CameraSystem

UI:
  • EbitenHUDSystem
  • EbitenMenuSystem
  • EbitenTutorialSystem
  • EbitenHelpSystem

Game Features:
  • AudioManagerSystem
  • ObjectiveTrackerSystem
  • ItemPickupSystem
  • SpellCastingSystem
  • ManaRegenSystem
  • StatusEffectSystem
  • RevivalSystem
  • SkillProgressionSystem
  • SpatialPartitionSystem
  • DialogSystem
  • CommerceSystem
```

### Shared Systems (6 systems)
```
Both Client and Server:
  • MovementSystem
  • CollisionSystem
  • CombatSystem
  • AISystem
  • ProgressionSystem
  • InventorySystem
```

**Rationale**: Server needs authoritative gameplay logic. Client needs local prediction and rendering.

## Integration Issue Found (FIXED)

### Before Audit
```
DialogSystem instantiated ✅
    ↓
DialogSystem.StartDialog() called ✅
    ↓
DialogSystem.Update() ❌ NEVER CALLED
    │
    └─ Not registered with World.AddSystem()
```

### After Fix
```
DialogSystem instantiated ✅
    ↓
World.AddSystem(dialogSystem) ✅ FIXED
    ↓
DialogSystem.Update() called every frame ✅
    │
    └─ Enables timed dialogs, dynamic content
```

## System Status Summary

| Status | Count | Systems |
|--------|-------|---------|
| ✅ Fully Integrated | 32 | All core gameplay, rendering, and UI systems |
| 🔧 Fixed During Audit | 1 | DialogSystem |
| ⚠️ Future Features | 6 | Crafting, Terrain Mod, Equipment Visual, Fire, Lighting, Weather |

---

**Last Updated**: 2025-10-27  
**Audit Completion**: 100%  
**Critical Issues**: 0 remaining (1 fixed)
