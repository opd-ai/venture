TASK: Conduct a systematic UI audit of Venture, a procedural multiplayer action-RPG built with Go/Ebiten, document all discovered issues with actionable fixes, and output findings to AUDIT.md in the project root.

CONTEXT:
- **Game**: Venture - Fully procedural multiplayer action-RPG
- **Technology Stack**: Go 1.24+ with Ebiten 2.9.3 game engine
- **Architecture**: Entity-Component-System (ECS) with deterministic procedural generation
- **Audit Scope**: All UI systems including game HUD, menus, inventory, character progression, crafting, quests, and multiplayer interfaces
- **Genre Support**: Fantasy, sci-fi, horror, cyberpunk, post-apocalyptic themes with cross-genre blending

REQUIREMENTS:

**Phase 1: Interface Discovery**
1. Map all UI components in `pkg/rendering/ui/`:
    - Main menu, pause menu, settings, genre selection
    - Game HUD (health, mana, minimap, hotbar, chat)
    - Character systems (inventory, skills, progression, stats)
    - World interaction (crafting stations, merchant trading, quest log)
    - Multiplayer interfaces (lobby, player list, connection status)

2. Document navigation flow using dual-exit pattern (ESC + dedicated back buttons)
3. Test procedural UI adaptation across all five genre themes
4. Verify deterministic UI state with same world seeds

**Phase 2: Systematic Testing**
1. Test each UI component for:
    - Visual clarity across all genre color palettes (`pkg/rendering/palette/`)
    - 60 FPS performance target with complex procedural content
    - Responsive feedback (visual/audio through `pkg/audio/synthesis/`)
    - Cross-platform input handling (keyboard, gamepad, touch)
    - Network latency tolerance (200-5000ms multiplayer scenarios)

2. Evaluate procedural generation integration:
    - Dynamic item tooltips with generated stats and descriptions
    - Procedural quest text rendering and formatting
    - Generated skill tree visualization and navigation
    - Real-time entity health bars and status effects
    - Merchant inventory with procedural items and pricing

3. Test ECS integration:
    - UI component updates through entity systems
    - Performance with 2000+ entities (target: 106 FPS achieved)
    - Component state synchronization in multiplayer
    - Memory usage (target: <500MB, achieved: 73MB)

**Phase 3: Venture-Specific Validation**
1. **Deterministic Testing**: Verify UI displays identical content with same seed across clients
2. **Genre Consistency**: Test UI theming across fantasy/sci-fi/horror/cyberpunk/post-apocalyptic
3. **Multiplayer Sync**: Validate UI state synchronization with client-side prediction
4. **Performance Benchmarks**: Ensure UI doesn't impact core performance targets
5. **Mobile Adaptation**: Test touch interface scaling and responsiveness
6. **Accessibility**: Verify keyboard navigation and screen reader compatibility

**Phase 4: Issue Classification**
Categorize by impact on Venture's core experience:
- **Critical**: Prevents core gameplay loop (exploration, combat, progression, multiplayer)
- **High**: Significantly degrades procedural content experience or genre immersion
- **Medium**: Affects usability but doesn't break deterministic generation
- **Low**: Polish opportunities that enhance but don't impair functionality

OUTPUT FORMAT:

Create `/AUDIT.md` in the project root with this structure:

```markdown
# Venture UI Audit Report
**Game**: Venture v3.0.0 Production - Procedural Multiplayer Action-RPG
**Audit Date**: [ISO 8601 format]
**Auditor**: GitHub Copilot
**Technology**: Go 1.24+ / Ebiten 2.9.3 / ECS Architecture
**Total Issues Found**: [Number]

## Executive Summary
[2-3 sentence overview focusing on procedural generation UI integration and multiplayer synchronization]

## Issues by Severity

### Critical Issues
#### Issue #[N]: [Descriptive Title]
- **Component**: [Specific UI system in pkg/rendering/ui/]
- **Genre Impact**: [Which genres affected: fantasy/sci-fi/horror/cyberpunk/post-apocalyptic]
- **Platform**: [Desktop/Web/Mobile or All]
- **Description**: [Problem explanation with ECS/procedural context]
- **Steps to Reproduce**:
  1. [Include seed value for deterministic testing]
  2. [Specify genre and difficulty parameters]
  3. [Detail multiplayer context if applicable]
- **Expected Behavior**: [What should happen per Venture's design patterns]
- **Actual Behavior**: [Current problematic behavior]
- **Suggested Fix**: [Actionable solution with Go/Ebiten code examples]
- **ECS Integration**: [How fix affects entity/component/system architecture]
- **Performance Impact**: [Frame rate or memory considerations]
- **Testing Verification**: [How to verify fix maintains determinism]

### High Priority Issues
[Same structure, focusing on genre immersion and procedural content display]

### Medium Priority Issues  
[Same structure, emphasizing usability and polish]

### Low Priority Issues
[Same structure, covering enhancement opportunities]

## Procedural Generation UI Integration
- **Genre Theming**: [How well UI adapts to different genres]
- **Dynamic Content**: [Procedural item/quest/skill display quality]
- **Performance**: [UI impact on generation and rendering performance]
- **Determinism**: [UI state consistency across clients with same seed]

## Multiplayer UI Assessment
- **State Synchronization**: [UI updates with network prediction]
- **Lag Tolerance**: [Interface responsiveness at 200-5000ms latency]
- **Player Communication**: [Chat, status indicators, lobby functionality]
- **Connection Feedback**: [Network status and error reporting]

## Cross-Platform Compatibility
- **Desktop**: [Windows/macOS/Linux keyboard and mouse experience]
- **WebAssembly**: [Browser deployment UI considerations]
- **Mobile**: [Touch interface adaptation and responsiveness]

## Positive Observations
[Highlight excellent UI/procedural integration examples and ECS patterns]

## Recommendations Summary
1. [Prioritized fixes with estimated development effort]
2. [Architecture improvements for better procedural integration]
3. [Performance optimizations for UI rendering]
4. [Multiplayer UI enhancements]

## Technical Implementation Notes
- **Ebiten Version**: 2.9.3
- **Go Version**: 1.24+
- **Resolution Tested**: [Multiple resolutions including mobile viewports]
- **Seed Used**: [For deterministic reproduction]
- **Network Conditions**: [Latency scenarios tested]
- **Entity Count**: [UI performance with various entity loads]
```

**Quality Standards Specific to Venture**:
- ✓ Every issue tested with deterministic seeds for reproducibility
- ✓ Fixes consider ECS architecture and component design patterns
- ✓ Performance impact assessed against 60 FPS / <500MB targets
- ✓ Multiplayer implications evaluated for all UI changes
- ✓ Genre theming compatibility verified across all five themes
  - _(... 2 more items ...)_

**Testing Methodology for Venture**:
1. **Deterministic Exploration**: Use fixed seeds (e.g., 12345) for consistent UI state testing
2. **Genre Rotation**: Test each UI component across all five supported genres
3. **Multiplayer Simulation**: Test UI with local server and multiple clients
4. **Performance Monitoring**: Use `go test -bench` patterns to verify UI performance
5. **Cross-Platform Validation**: Test desktop, WebAssembly, and mobile builds
6. **ECS Integration**: Verify UI components properly interact with entity systems
7. **Save/Load Consistency**: Test UI state preservation through save/load cycles

**Venture-Specific Considerations**:
- Procedural content must render consistently with same generation parameters
- UI theming should seamlessly adapt to genre blending scenarios
- Network UI must handle Venture's high-latency tolerance (onion services)
- Touch controls must maintain action-RPG responsiveness on mobile
- Entity UI (health bars, status effects) must perform with 2000+ entities
  - _(... 4 more items ...)_

## Visual Layering and Collision Detection Audit

**Purpose**: Systematically verify that all visual elements render in correct z-order and collision detection works accurately across all game layers. This ensures proper gameplay mechanics where entities interact correctly based on their layer positions.

### Layer System Architecture

Venture implements a multi-layer terrain and rendering system with three primary layer types:

1. **Terrain Layers** (`pkg/engine/layer_component.go`):
   - Layer 0: Ground layer (default, floor, corridors)
   - Layer 1: Water/Pit layer (deep water, pits, chasms)
   - Layer 2: Platform layer (elevated platforms, bridges)

2. **Sprite Render Layers** (`pkg/engine/render_system.go`):
   - Controlled by `EbitenSprite.Layer` integer field
   - Higher values render on top of lower values
   - Sorted via `sortEntitiesByLayer()` before rendering
   - Typical values: 0=terrain, 1=ground items, 2=entities, 3=particles, 4=UI overlays

3. **Equipment Visual Layers** (`pkg/engine/equipment_visual_component.go`):
   - LayerBody: Base character sprite
   - LayerHead: Head equipment overlay
   - LayerWeapon: Weapon sprite overlay
   - LayerArmor: Armor visual overlay
   - LayerAccessory: Accessory overlays (multiple)

### Phase 1: Visual Layering Audit

**1.1 Sprite Layer Verification**

Test that entities render in correct z-order:

```go
// Test Steps:
// 1. Create entities with different sprite layers
entity1 := world.CreateEntity() // Layer 0
entity1.AddComponent("sprite", NewSpriteComponent(32, 32, color.RGBA{255, 0, 0, 255}))
sprite1 := entity1.GetComponent("sprite").(*EbitenSprite)
sprite1.Layer = 0

entity2 := world.CreateEntity() // Layer 2
entity2.AddComponent("sprite", NewSpriteComponent(32, 32, color.RGBA{0, 255, 0, 255}))
sprite2 := entity2.GetComponent("sprite").(*EbitenSprite)
sprite2.Layer = 2

// 2. Place entities at same position
entity1.AddComponent("position", &PositionComponent{X: 400, Y: 300})
entity2.AddComponent("position", &PositionComponent{X: 400, Y: 300})

// 3. Expected: entity2 (green, Layer 2) renders on top of entity1 (red, Layer 0)
// 4. Verify via screenshot or visual inspection
```

**Issues to Check**:
- Entities with higher `Layer` values always render on top of lower values
- Sorting is stable and consistent across frames
- Layer changes take effect immediately (no frame delay)
- Negative layer values work correctly (if used)
- Layer 0 entities don't obscure higher-layer UI elements

**1.2 Terrain Layer Verification**

Test multi-layer terrain collision and rendering:

```go
// Test Steps:
// 1. Create entity with LayerComponent
entity := world.CreateEntity()
layerComp := NewLayerComponent() // Starts on Layer 0 (ground)
entity.AddComponent("layer", layerComp)

// 2. Test layer transition
layerComp.StartTransition(2) // Transition to platform layer
// Expected: CanTransitionTo(2) returns true if entity has CanClimb=true

// 3. Update transition progress
for progress := 0.0; progress < 1.0; progress += 0.1 {
    layerComp.UpdateTransition(0.1)
    // Verify GetEffectiveLayer() returns correct layer during transition
}

// 4. Verify layer-based collision filtering
entity2 := world.CreateEntity()
entity2.AddComponent("layer", NewLayerComponent())
entity2.GetComponent("layer").(*LayerComponent).CurrentLayer = 1
// Expected: OnSameLayer(layerComp, entity2.GetComponent("layer")) returns false
```

**Issues to Check**:
- `CurrentLayer` correctly determines collision eligibility
- `GetEffectiveLayer()` returns target layer when transition progress > 0.5
- `OnSameLayer()` properly handles nil layer components (defaults to Layer 0)
- Flying entities (`CanFly=true`) can move between any layers
- Swimming entities can only enter water layer (Layer 1)

**1.3 Equipment Visual Layer Verification**

Test equipment overlay rendering:

```go
// Test Steps:
// 1. Create entity with equipment visual component
entity := world.CreateEntity()
equipVisual := NewEquipmentVisualComponent()
entity.AddComponent("equipment_visual", equipVisual)
entity.AddComponent("sprite", NewSpriteComponent(64, 64, color.White))

// 2. Add equipment layers
equipVisual.SetWeaponEquipped(true)
equipVisual.SetArmorEquipped(true)
equipVisual.AddAccessorySlot()

// 3. Regenerate equipment layers via EquipmentVisualSystem
system := NewEquipmentVisualSystem(spriteGen)
system.Update([]*Entity{entity}, 0.016)

// 4. Expected: WeaponLayer, ArmorLayer, AccessoryLayers[0] are non-nil
// 5. Verify layers composite correctly on base sprite
```

**Issues to Check**:
- Equipment layers render on top of base character sprite
- Layer order: Body < Head < Armor < Weapon < Accessory
- Removing equipment clears corresponding layer (sets to nil)
- Equipment sprites match genre palette and character size
- Equipment visibility respects sprite `Visible` flag

### Phase 2: Collision Detection Audit

**2.1 Layer-Based Collision Filtering**

Verify entities on different layers don't collide:

```go
// Test Steps:
// 1. Create two entities on different terrain layers
entity1 := world.CreateEntity()
entity1.AddComponent("position", &PositionComponent{X: 100, Y: 100})
entity1.AddComponent("collider", NewColliderComponent(32, 32))
entity1.AddComponent("layer", NewLayerComponent()) // Layer 0

entity2 := world.CreateEntity()
entity2.AddComponent("position", &PositionComponent{X: 100, Y: 100})
entity2.AddComponent("collider", NewColliderComponent(32, 32))
layer2 := NewLayerComponent()
layer2.CurrentLayer = 2 // Platform layer
entity2.AddComponent("layer", layer2)

// 2. Check collision
collisionSystem := NewCollisionSystem(64.0)
collisionSystem.Update([]*Entity{entity1, entity2}, 0.016)

// 3. Expected: No collision because entities are on different layers
// 4. Verify via collision callback (should not be invoked)
```

**Issues to Check**:
- Entities on different terrain layers don't trigger collisions
- `OnSameLayer()` is checked before collision resolution
- Collider `Layer` field (0 = all layers) works correctly
- Flying entities collide with entities on all layers
- Layer transitions handle collision state changes smoothly

**2.2 Collider Type Verification**

Test solid vs. trigger collider behavior:

```go
// Test Steps:
// 1. Create solid collider entity
solid := world.CreateEntity()
solid.AddComponent("position", &PositionComponent{X: 100, Y: 100})
solidCollider := NewColliderComponent(32, 32)
solidCollider.Solid = true
solidCollider.IsTrigger = false
solid.AddComponent("collider", solidCollider)

// 2. Create trigger collider entity
trigger := world.CreateEntity()
trigger.AddComponent("position", &PositionComponent{X: 100, Y: 100})
triggerCollider := NewColliderComponent(32, 32)
triggerCollider.Solid = true
triggerCollider.IsTrigger = true
trigger.AddComponent("collider", triggerCollider)

// 3. Test collision behavior
collisionSystem := NewCollisionSystem(64.0)
var collisionTriggered bool
collisionSystem.SetCollisionCallback(func(e1, e2 *Entity) {
    collisionTriggered = true
})
collisionSystem.Update([]*Entity{solid, trigger}, 0.016)

// 4. Expected: Callback invoked, but no position resolution (trigger doesn't block)
```

**Issues to Check**:
- `IsTrigger=true` colliders invoke callbacks but don't block movement
- `Solid=false` colliders don't resolve overlap
- Trigger-trigger collision still triggers callbacks
- Solid-solid collision resolves position overlap correctly
- Collider bounds (`GetBounds()`) accurately represent entity size

**2.3 Terrain Collision Verification**

Test wall collision detection and resolution:

```go
// Test Steps:
// 1. Set up terrain collision checker with wall at (100, 100)
checker := NewTerrainCollisionChecker()
checker.SetWallAt(100, 100)

collisionSystem := NewCollisionSystem(64.0)
collisionSystem.SetTerrainChecker(checker)

// 2. Create entity moving toward wall
entity := world.CreateEntity()
entity.AddComponent("position", &PositionComponent{X: 90, Y: 100})
entity.AddComponent("collider", NewColliderComponent(16, 16))
entity.AddComponent("velocity", &VelocityComponent{VX: 20, VY: 0})

// 3. Update collision system
collisionSystem.Update([]*Entity{entity}, 0.016)

// 4. Expected: Entity position pushed away from wall, velocity.VX set to 0
pos := entity.GetComponent("position").(*PositionComponent)
vel := entity.GetComponent("velocity").(*VelocityComponent)
// Verify pos.X < 100 and vel.VX == 0
```

**Issues to Check**:
- `WouldCollideWithTerrain()` accurately predicts collisions
- `CheckEntityCollision()` detects when entity overlaps walls
- `resolveTerrainCollision()` pushes entity to valid position
- Collision resolution checks all 8 directions for valid positions
- Velocity is zeroed in blocked direction only (preserves perpendicular movement)

**2.4 Predictive Collision Checks**

Verify predictive collision methods:

```go
// Test Steps:
// 1. Create two entities at safe distance
entity1 := world.CreateEntity()
entity1.AddComponent("position", &PositionComponent{X: 100, Y: 100})
entity1.AddComponent("collider", NewColliderComponent(32, 32))

entity2 := world.CreateEntity()
entity2.AddComponent("position", &PositionComponent{X: 200, Y: 100})
entity2.AddComponent("collider", NewColliderComponent(32, 32))

collisionSystem := NewCollisionSystem(64.0)

// 2. Test predictive collision at overlapping position
wouldCollide := collisionSystem.WouldCollideWithEntity(entity1, 200, 100, entity2)
// Expected: wouldCollide = true

// 3. Test predictive collision at safe position
wouldCollide = collisionSystem.WouldCollideWithEntity(entity1, 100, 100, entity2)
// Expected: wouldCollide = false
```

**Issues to Check**:
- `WouldCollideWithTerrain()` doesn't modify entity state
- `WouldCollideWithEntity()` correctly predicts overlap at given position
- Predictive checks respect layer compatibility
- Predictive checks handle trigger colliders correctly (return false)
- Bounds calculations at predicted positions are accurate

### Phase 3: Integration Testing

**3.1 Rendering-Collision Consistency**

Verify visual position matches collision bounds:

```markdown
**Test Procedure**:
1. Enable debug rendering: `renderSystem.SetShowColliders(true)`
2. Create entity with sprite and collider
3. Screenshot entity with collider overlay visible
4. Expected: Green collider rectangle perfectly outlines visible sprite
5. Test with various sprite sizes, rotations, and positions
```

**Issues to Check**:
- Collider bounds match sprite visual bounds
- Rotation doesn't desync collider from sprite
- Position updates sync between sprite and collider
- Collider offset (if any) matches sprite anchor point

**3.2 Layer Transition Visual Feedback**

Test visual feedback during layer transitions:

```markdown
**Test Procedure**:
1. Create entity with LayerComponent on Layer 0
2. Start transition to Layer 2 via StartTransition(2)
3. During transition (TransitionProgress 0.0 to 1.0):
   - Entity should remain visible
   - Collision should use GetEffectiveLayer() (switches at 0.5)
   - Visual position should smoothly interpolate (if using y-offset for depth)
4. After transition completes (CurrentLayer = 2):
   - Entity fully on new layer
   - Collides only with Layer 2 entities
```

**Issues to Check**:
- Visual feedback shows entity is transitioning (optional visual effect)
- Collision layer switches at TransitionProgress > 0.5
- Transition can be cancelled without visual artifacts
- Multiple simultaneous transitions don't interfere

**3.3 Performance Under Load**

Verify layering/collision performance with 2000+ entities:

```markdown
**Test Procedure**:
1. Create 2000 entities across all 3 terrain layers
2. Assign sprite layers 0-4 randomly
3. Enable collision system with spatial partitioning
4. Measure frame time with go test -bench
5. Target: <16.67ms per frame (60 FPS)
6. Current benchmark: ~0.02ms with viewport culling
```

**Issues to Check**:
- Layer sorting doesn't degrade performance (O(n log n) is acceptable)
- Collision checks scale with spatial partitioning, not entity count
- GetEffectiveLayer() doesn't add measurable overhead
- Batch rendering works correctly across layer boundaries

### Phase 4: Common Issues and Fixes

**Issue #1: Entities Render in Wrong Order**
```markdown
**Symptom**: Entity A should appear behind Entity B but renders on top
**Root Cause**: Incorrect sprite Layer values
**Fix**: 
```go
// Verify layer assignments
spriteA := entityA.GetComponent("sprite").(*EbitenSprite)
spriteB := entityB.GetComponent("sprite").(*EbitenSprite)
fmt.Printf("Entity A layer: %d, Entity B layer: %d\n", spriteA.Layer, spriteB.Layer)

// Correct if needed
spriteA.Layer = 1  // Lower value = renders behind
spriteB.Layer = 2  // Higher value = renders on top
```

**Issue #2: Entities Collide When On Different Terrain Layers**
```markdown
**Symptom**: Platform entity (Layer 2) collides with ground entity (Layer 0)
**Root Cause**: LayerComponent missing or not checked in collision
**Fix**:
```go
// Ensure both entities have LayerComponent
if !entity.HasComponent("layer") {
    entity.AddComponent("layer", NewLayerComponent()) // Defaults to Layer 0
}

// Verify OnSameLayer check in collision system
// collision.go should check this before resolving collision
```

**Issue #3: Collider Doesn't Match Visual Sprite**
```markdown
**Symptom**: Entity appears to collide before sprites touch
**Root Cause**: Collider size doesn't match sprite size
**Fix**:
```go
sprite := entity.GetComponent("sprite").(*EbitenSprite)
collider := entity.GetComponent("collider").(*ColliderComponent)

// Sync collider to sprite dimensions
collider.Width = sprite.Width
collider.Height = sprite.Height

// Optional: Adjust for sprite padding
collider.Width = sprite.Width * 0.8  // 80% of sprite width for tighter collision
collider.Height = sprite.Height * 0.8
```

**Issue #4: Trigger Not Firing**
```markdown
**Symptom**: Trigger collider doesn't invoke callback when overlapped
**Root Cause**: Solid flag set to false
**Fix**:
```go
trigger := entity.GetComponent("collider").(*ColliderComponent)
trigger.Solid = true   // Must be true for collision detection
trigger.IsTrigger = true  // This prevents position resolution
```

**Issue #5: Equipment Layers Not Visible**
```markdown
**Symptom**: Equipped items don't show on character
**Root Cause**: Equipment visual system not updating
**Fix**:
```go
// Ensure EquipmentVisualSystem is registered and updating
equipSystem := NewEquipmentVisualSystem(spriteGenerator)
world.AddSystem(equipSystem)

// Verify equipment flags are set
equipVisual := entity.GetComponent("equipment_visual").(*EquipmentVisualComponent)
equipVisual.SetWeaponEquipped(true)

// Force regeneration
equipSystem.Update([]*Entity{entity}, 0.0)
```

### Testing Methodology

**Deterministic Testing**:
- Use fixed seeds for reproducible visual and collision states
- Test example: `seed := int64(12345)`
- Generate entities with same seed across test runs
- Compare screenshots or collision logs for consistency

**Cross-Genre Validation**:
- Test layer rendering with all five genre palettes
- Verify layer visibility with dark (horror) and bright (cyberpunk) themes
- Ensure collision detection is genre-agnostic

**Multiplayer Synchronization**:
- Verify layer state syncs between client and server
- Test that LayerComponent.CurrentLayer matches across clients
- Ensure collision resolution is deterministic with same inputs

**Performance Benchmarks**:
```bash
# Run collision system benchmark
go test -bench=BenchmarkCollisionSystem -benchmem ./pkg/engine/

# Run render system benchmark with layering
go test -bench=BenchmarkRenderSystem -benchmem ./pkg/engine/

# Target: <0.1ms per frame for collision, <1ms for rendering with 2000 entities
```

**Example Venture-Specific Issue**:

```markdown
#### Issue #1: Procedural Item Tooltip Overflow in Cyberpunk Genre
- **Component**: pkg/rendering/ui/inventory.go - item tooltip rendering
- **Genre Impact**: Cyberpunk (neon color palette causes text contrast issues)
- **Platform**: All platforms
- **Description**: When displaying procedural items with long generated names (e.g., "Quantum-Enhanced Tactical Neural Interface Mk VII"), tooltips extend beyond screen bounds in cyberpunk genre due to bright neon backgrounds reducing text contrast.
- **Steps to Reproduce**:
  1. Start with seed 12345, cyberpunk genre, difficulty 0.8
  2. Generate items until tooltip with 40+ character name appears
  3. Hover over item in inventory screen
- **Expected Behavior**: Tooltip should wrap text and maintain readability across all genres
- **Actual Behavior**: Text overflows, neon background makes text illegible
- **Suggested Fix**: Implement dynamic text wrapping in renderTooltip() function:
  ```go
  func (ui *InventoryUI) renderTooltip(item *procgen.Item, palette *palette.GenrePalette) {
        maxWidth := ui.screenWidth * 0.3
        wrappedText := ui.wrapText(item.GetDisplayName(), maxWidth)
        // Use genre-appropriate contrast for background
        bgColor := palette.GetContrastBackground()
        ui.renderTextWithBackground(wrappedText, bgColor)
  }
  ```
- **ECS Integration**: Tooltip rendering system should query ItemComponent and use cached display names
- **Performance Impact**: Text wrapping adds ~0.1ms per tooltip, negligible for 60 FPS target
```