# System Interaction Map

System categories, execution flow, and dependencies in Venture's ECS architecture.

## System Categories (38 Total)

**Core ECS (3):** Movement, Collision, Spatial  
**Combat (4):** Combat, Player Combat, Status, Revival  
**AI (2):** AI, Fire Propagation*  
**Progression (3):** XP, Skills, Objectives  
**Inventory (4):** Inventory, Pickup, Use, Visuals*  
**Magic (3):** Casting, Player Casting, Mana  
**Input (1):** Input System  
**Rendering (5):** Render, Terrain, Animation, Particles, Feedback  
**Camera (1):** Camera System  
**Audio (1):** Audio Manager  
**UI (4):** HUD, Menu, Tutorial, Help  
**Terrain (2):** Construction*, Modification*  
**Commerce (2):** Commerce, Dialog  
**Crafting (1):** Crafting*  
**Lighting (1):** Dynamic Lighting*  
**Weather (1):** Weather Effects*

\* = Future feature (implemented but not integrated)

## Execution Order (Per Frame)

1. **Input:** Keyboard/mouse capture
2. **Player Actions:** Attack/item use/spell cast
3. **Physics:** Movement → collision detection
4. **Combat:** Damage calculation, status effects
5. **AI:** Revival, enemy decisions
6. **Progression:** XP, skill bonuses
7. **Visual:** Hit flash, tints
8. **Audio:** Music context updates
9. **Quests:** Quest progress, item pickup
10. **Magic:** Spell effects, mana regen
11. **Inventory:** Item management
12. **Animation:** Sprite frame updates
13. **UI:** Tutorial, help, dialog
14. **Particles:** Effect rendering
15. **Spatial:** Quadtree updates (every 60 frames)

## Critical Dependencies

**Movement ↔ Collision (Bidirectional):**
```go
MovementSystem.SetCollisionSystem(collisionSystem)
// Enables predictive collision detection
```

**Collision → Terrain:**
```go
collisionSystem.SetTerrainChecker(NewTerrainCollisionChecker())
// Entity-terrain collision
```

**Combat → Camera + Particles:**
```go
combatSystem.SetCamera(cameraSystem)           // Screen shake
combatSystem.SetParticleSystem(particleSystem) // Hit effects
```

**Render → Spatial (Culling):**
```go
renderSystem.SetSpatialPartition(spatialSystem)
renderSystem.EnableCulling(true)  // Viewport optimization
```

**Input → UI + Callbacks:**
```go
inputSystem.SetHelpSystem(helpSystem)
inputSystem.SetTutorialSystem(tutorialSystem)
inputSystem.SetQuickSaveCallback(saveFunc)
inputSystem.SetInteractCallback(merchantFunc)
// Hotkey dispatch
```

## Component Query Patterns

**Pattern 1: Simple Query**
```go
entities := world.GetEntitiesWith("position", "velocity")
// Used by: Movement, Collision, AI
```

**Pattern 2: Component Filtering**
```go
for _, entity := range entities {
    if pos, ok := entity.GetComponent("position"); ok {
        if vel, ok := entity.GetComponent("velocity"); ok {
            // Process movement
        }
    }
}
// Used by: Most ECS systems
```

**Pattern 3: Type Checking**
```go
if !entity.HasComponent("input") {
    continue  // Skip non-player entities
}
// Used by: Player-specific systems
```

## System Integration

**Add New System:**
1. Implement `Update(deltaTime float64)` method
2. Query entities with required components
3. Operate on component data only (no entity logic)
4. Register in world: `world.RegisterSystem(system)`
5. Set dependencies via setter methods

**System Registration Order:**
Critical systems (movement, collision) → Gameplay systems (combat, AI) → Rendering systems (particles, animation)

## Data Flow

**Input → Game State:**
```
Keyboard/Mouse → InputComponent → Player*Systems → Components → World State
```

**Game State → Rendering:**
```
Components → RenderSystem → SpriteCache → Screen
           → ParticleSystem → Effects
           → AnimationSystem → Frames
```

**Multiplayer Sync:**
```
Server: World State → Snapshots → Network
Client: Network → Snapshots → Interpolation → Prediction → Render
```

## Performance Characteristics

**Quadtree Spatial Partition:**
- Update frequency: Every 60 frames (~1s at 60 FPS)
- Query time: O(log n) for collision/culling
- Entities per query: ~5-50 (typical viewport)

**Viewport Culling:**
- Speedup: 1,635x over naive rendering
- Entities rendered: ~200 of 2000 (10% typical)

**Batch Rendering:**
- Speedup: 1,667x over individual draws
- Batches per frame: <50 (from ~500 unbatched)

**Sprite Cache:**
- Hit rate: 95.9%
- Miss penalty: ~1-2ms generation time

---

**Last Updated:** November 14, 2025
