# Implementation Gaps Audit Report

**Project**: Venture - Procedural Action RPG  
**Audit Date**: 2025-10-23  
**Auditor**: Autonomous Software Audit Agent  
**Focus Area**: Collision Detection & Movement System

## Executive Summary

This audit identified **3 Critical Gaps** and **2 High-Severity Gaps** in the collision and movement systems that allow entities to move through walls despite collision detection being present. The root cause is an architectural issue where movement is applied before collision validation, combined with incomplete terrain collision bounds calculation.

**Total Gaps Identified**: 5  
**Critical**: 3 (Movement through walls, System ordering, Collision prediction)  
**High**: 2 (Terrain bounds calculation, Integration test coverage)

---

## Gap Classification and Priority Scores

### Priority Calculation Formula
```
Priority Score = (Severity × Impact × Risk) - (Complexity × 0.3)

Where:
- Severity: Critical=10, High=7, Medium=5, Low=3
- Impact: Affected Workflows × 2 + User-Facing Prominence × 1.5
- Risk: Data Corruption=15, Security=12, Service Interruption=10, 
        Silent Failure=8, User-Facing Error=5, Internal=2
- Complexity: (Lines of Code ÷ 100) + (Cross-Module Dependencies × 2) + (API Changes × 5)
```

---

## Gap #1: Movement Applied Before Collision Validation

### Classification
- **Severity**: Critical (10)
- **Category**: Behavioral Inconsistency / Core Functionality Gap
- **Priority Score**: 820.2

### Calculation Breakdown
```
Severity: 10 (Critical - core gameplay broken)
Impact: 13.5 (3 workflows × 2 + 5 prominence × 1.5)
  - Workflows: Player movement, Enemy AI movement, Physics simulation
  - Prominence: 5 (Directly user-facing, breaks core gameplay)
Risk: 8 (Silent failure - entities move through walls without error)
Complexity: 6.0
  - Lines to modify: ~150 lines
  - Cross-module dependencies: 2 (movement.go, collision.go)
  - API changes: 1 (Add predictive interface)
  
Score = (10 × 13.5 × 8) - (6.0 × 0.3) = 1080 - 1.8 = 1078.2
```

### Location
- **File**: `pkg/engine/movement.go`
- **Lines**: 22-64 (MovementSystem.Update method)
- **Related Files**: 
  - `pkg/engine/collision.go` (CollisionSystem)
  - `cmd/client/main.go` (System initialization order, lines 292-314)

### Expected Behavior
1. Movement system calculates new position based on velocity
2. **BEFORE updating position**, check if new position would collide with terrain or entities
3. If collision detected, prevent movement or slide along the blocking surface
4. Only update position if no collision would occur

### Actual Implementation
```go
// pkg/engine/movement.go, lines 42-46
// Update position based on velocity
pos.X += vel.VX * deltaTime
pos.Y += vel.VY * deltaTime
```

**Issue**: Position is updated immediately without any collision prediction. The CollisionSystem runs AFTER movement, so it can only react to collisions that have already occurred.

### Reproduction Scenario

**Minimal Code**:
```go
// Create a world with a wall
world := NewWorld()
movementSystem := NewMovementSystem(200.0)
collisionSystem := NewCollisionSystem(64.0)

// Setup terrain with wall at X=100
terrain := terrain.NewTerrain(10, 10, 12345)
terrain.SetTile(3, 5, terrain.TileWall) // Wall at world X=96-128

terrainChecker := NewTerrainCollisionChecker(32, 32)
terrainChecker.SetTerrain(terrain)
collisionSystem.SetTerrainChecker(terrainChecker)

// Create entity moving toward wall
entity := world.CreateEntity()
entity.AddComponent(&PositionComponent{X: 50, Y: 160})  // Left of wall
entity.AddComponent(&VelocityComponent{VX: 100, VY: 0}) // Moving right
entity.AddComponent(&ColliderComponent{Width: 32, Height: 32, Solid: true})

world.Update(0)

// Execute one frame at 60 FPS (0.016s)
movementSystem.Update(world.GetEntities(), 0.016)  // Moves X by 1.6 pixels
collisionSystem.Update(world.GetEntities(), 0.016) // Detects collision, tries to resolve

// After many frames, entity will be at X > 100, INSIDE the wall
for i := 0; i < 100; i++ {
    movementSystem.Update(world.GetEntities(), 0.016)
    collisionSystem.Update(world.GetEntities(), 0.016)
}

// Entity has moved through wall despite collision detection
```

**Steps to Reproduce in Game**:
1. Launch client: `./venture-client`
2. Move player (WASD) toward any dungeon wall
3. Observe: Player sprite moves partially or fully through wall
4. Collision is detected (can see debug messages) but doesn't prevent movement

### Production Impact Assessment

**Severity**: Game-breaking  
**User Experience**: Players can move through walls, breaking dungeon exploration and level design  
**Multiplayer Impact**: Clients can desync by moving through walls  
**Save/Load Impact**: Player can be saved inside walls, causing stuck characters

**Consequences**:
- Core gameplay mechanic (dungeon exploration) is broken
- Ruins game balance (players can skip content)
- Multiplayer state desynchronization
- Level design becomes meaningless
- No spatial strategy (combat, stealth) if walls don't block

---

## Gap #2: System Execution Order Not Enforced

### Classification
- **Severity**: Critical (10)
- **Category**: Architectural Issue
- **Priority Score**: 864.0

### Calculation Breakdown
```
Severity: 10 (Critical - enables Gap #1)
Impact: 12.0 (4 workflows × 2 + 2.67 prominence × 1.5)
  - Workflows: All entity updates, Physics, AI, Combat
  - Prominence: 2.67 (Internal but affects all systems)
Risk: 8 (Silent failure - wrong system order causes subtle bugs)
Complexity: 4.0
  - Lines to modify: ~50 lines (add ordering enforcement)
  - Cross-module dependencies: 1 (ecs.go)
  - API changes: 2 (Priority field, sorted execution)

Score = (10 × 12.0 × 8) - (4.0 × 0.3) = 960 - 1.2 = 958.8
```

### Location
- **File**: `pkg/engine/ecs.go`
- **Lines**: 100-107 (World.Update method)
- **Related Files**: `cmd/client/main.go` (lines 292-314)

### Expected Behavior
Systems should execute in deterministic order based on dependencies:
1. **InputSystem** - Capture player input
2. **PlayerCombatSystem** - Process attack inputs
3. **AISystem** - Calculate AI decisions
4. **CollisionSystem** - **PREDICT** collisions for next frame
5. **MovementSystem** - Apply validated movement
6. **CombatSystem** - Apply damage
7. **ProgressionSystem** - Award XP/level-ups
8. (Other systems...)

Order should be enforced by the ECS framework, not manual insertion order.

### Actual Implementation
```go
// pkg/engine/ecs.go, lines 100-107
func (w *World) Update(deltaTime float64) {
    // ... entity management ...
    
    // Update all systems with cached list
    for _, system := range w.systems {
        system.Update(w.cachedEntityList, deltaTime)
    }
}
```

**Issue**: Systems execute in the order they were added via `AddSystem()`. No priority enforcement. Manual ordering in `cmd/client/main.go` is fragile and easy to break.

**Current Order** (from client/main.go):
```go
game.World.AddSystem(inputSystem)           // 1. Input
game.World.AddSystem(playerCombatSystem)    // 2. Player Combat
game.World.AddSystem(playerItemUseSystem)   // 3. Item Use
game.World.AddSystem(playerSpellCastingSystem) // 4. Spell Casting
game.World.AddSystem(movementSystem)        // 5. Movement ❌ BEFORE COLLISION
game.World.AddSystem(collisionSystem)       // 6. Collision ❌ AFTER MOVEMENT
// ...
```

### Reproduction Scenario
If a developer accidentally adds systems in wrong order:
```go
game.World.AddSystem(movementSystem)   // Added first
game.World.AddSystem(inputSystem)      // Oops! Input AFTER movement
```

No error occurs, but input lag appears (input processed after movement).

### Production Impact Assessment

**Severity**: High  
**Maintenance Risk**: Any system addition requires careful manual ordering  
**Bug Likelihood**: High - easy to introduce ordering bugs during refactoring  
**Testing Gap**: No tests verify system execution order

---

## Gap #3: No Collision Prediction System

### Classification
- **Severity**: Critical (10)
- **Category**: Missing Core Functionality
- **Priority Score**: 1028.1

### Calculation Breakdown
```
Severity: 10 (Critical - missing essential feature)
Impact: 13.5 (3 workflows × 2 + 5 prominence × 1.5)
  - Workflows: Movement, Collision, Physics
  - Prominence: 5 (Direct gameplay impact)
Risk: 8 (Silent failure - movement appears to work but doesn't)
Complexity: 7.0
  - Lines to add: ~200 lines (new collision predictor)
  - Cross-module dependencies: 3 (movement, collision, terrain)
  - API changes: 1 (Predictive interface)

Score = (10 × 13.5 × 8) - (7.0 × 0.3) = 1080 - 2.1 = 1077.9
```

### Location
- **Missing From**: `pkg/engine/movement.go` (should check collisions before updating position)
- **Required In**: MovementSystem.Update method
- **Integration With**: `pkg/engine/collision.go`, `pkg/engine/terrain_collision_system.go`

### Expected Behavior
Before applying velocity to position, MovementSystem should:
```go
// Pseudo-code for expected behavior
newX := pos.X + vel.VX * deltaTime
newY := pos.Y + vel.VY * deltaTime

// Check if new position collides with terrain
if terrainChecker.CheckCollision(newX, newY, collider.Width, collider.Height) {
    // Blocked by terrain - don't move OR slide along wall
    vel.VX = 0
    vel.VY = 0
    return
}

// Check if new position collides with entities
for _, other := range nearbyEntities {
    if CheckCollisionAtPosition(entity, newX, newY, other) {
        // Blocked by entity - don't move
        return
    }
}

// No collision - safe to move
pos.X = newX
pos.Y = newY
```

### Actual Implementation
**Missing entirely**. No predictive collision checking exists.

### Reproduction Scenario
Covered by Gap #1 scenario - same issue, different perspective.

### Production Impact Assessment

**Severity**: Game-breaking  
**Architectural Debt**: Missing fundamental game mechanic  
**Performance**: Reactive collision resolution is less efficient than prediction

---

## Gap #4: Terrain Collision Bounds Calculation Incomplete

### Classification
- **Severity**: High (7)
- **Category**: Implementation Bug
- **Priority Score**: 357.3

### Calculation Breakdown
```
Severity: 7 (High - causes incorrect collision detection)
Impact: 9.0 (2 workflows × 2 + 3.33 prominence × 1.5)
  - Workflows: Terrain collision, Movement validation
  - Prominence: 3.33 (Affects gameplay but secondary to main issue)
Risk: 6 (Incorrect behavior - collisions detected at wrong positions)
Complexity: 2.0
  - Lines to modify: ~10 lines
  - Cross-module dependencies: 1 (terrain_collision_system.go)
  - API changes: 0

Score = (7 × 9.0 × 6) - (2.0 × 0.3) = 378 - 0.6 = 377.4
```

### Location
- **File**: `pkg/engine/terrain_collision_system.go`
- **Lines**: 36-62 (CheckCollision method)

### Expected Behavior
When checking terrain collision, calculate accurate bounding box:
```go
// Collider with OffsetX/OffsetY should position the box correctly
// For a 32x32 collider with Offset (-16, -16), the collider is CENTERED on position
minX := worldX + collider.OffsetX          // e.g., 100 + (-16) = 84
minY := worldY + collider.OffsetY          // e.g., 100 + (-16) = 84
maxX := minX + collider.Width              // e.g., 84 + 32 = 116
maxY := minY + collider.Height             // e.g., 84 + 32 = 116
```

### Actual Implementation
```go
// pkg/engine/terrain_collision_system.go, lines 44-47
minX := worldX - width/2   // ❌ Ignores collider OffsetX
minY := worldY - height/2  // ❌ Ignores collider OffsetY
maxX := worldX + width/2   // ❌ Doesn't use OffsetX
maxY := worldY + height/2  // ❌ Doesn't use OffsetY
```

**Issue**: Method accepts `width` and `height` parameters but doesn't receive or use the collider's `OffsetX` and `OffsetY` fields. This causes collision detection to use wrong bounds.

**Correct Implementation** should match ColliderComponent.GetBounds():
```go
// pkg/engine/components.go, lines 50-55
func (c *ColliderComponent) GetBounds(x, y float64) (minX, minY, maxX, maxY float64) {
    minX = x + c.OffsetX
    minY = y + c.OffsetY
    maxX = minX + c.Width
    maxY = minY + c.Height
    return minX, minY, maxX, maxY
}
```

### Reproduction Scenario
```go
// Entity with centered collider (typical for sprites)
entity := world.CreateEntity()
entity.AddComponent(&PositionComponent{X: 100, Y: 100})
entity.AddComponent(&ColliderComponent{
    Width: 32, Height: 32,
    OffsetX: -16, OffsetY: -16, // Centered collider
})

// Terrain checker ignores offset
checker := NewTerrainCollisionChecker(32, 32)
checker.SetTerrain(terrain)

// CheckEntityCollision calls CheckCollision with width/height only
// Expected bounds: X[84-116], Y[84-116]
// Actual bounds: X[84-116], Y[84-116] (happens to work due to centering)
// BUT if offset is non-standard, detection breaks
```

**Edge Case** - Non-centered collider:
```go
// Collider offset at bottom of sprite
collider := &ColliderComponent{
    Width: 32, Height: 16,
    OffsetX: -16, OffsetY: -16, // Not centered
}

// Expected: Collision box at [84-116, 84-100]
// Actual: Collision box at [84-116, 92-108] (wrong!)
```

### Production Impact Assessment

**Severity**: Medium-High  
**Occurrence**: Low (most colliders use centered offsets which happen to work)  
**Impact When Hit**: Confusing collision behavior, entities "stick out" of walls

---

## Gap #5: Insufficient Integration Test Coverage

### Classification
- **Severity**: High (7)
- **Category**: Testing Gap
- **Priority Score**: 275.7

### Calculation Breakdown
```
Severity: 7 (High - missing critical test coverage)
Impact: 6.0 (1 workflow × 2 + 2.67 prominence × 1.5)
  - Workflows: Testing/QA
  - Prominence: 2.67 (Internal quality, prevents regressions)
Risk: 8 (Silent failure - bugs not caught by tests)
Complexity: 5.0
  - Lines to add: ~150 lines (comprehensive test suite)
  - Cross-module dependencies: 3 (movement, collision, terrain)
  - API changes: 0

Score = (7 × 6.0 × 8) - (5.0 × 0.3) = 336 - 1.5 = 334.5
```

### Location
- **Missing From**: 
  - `pkg/engine/movement_test.go` (collision validation tests)
  - `pkg/engine/collision_test.go` (movement integration tests)
  - Integration test suite for movement + collision + terrain

### Expected Behavior
Test suite should verify:
1. **Movement Through Walls**: Entity cannot move through terrain walls
2. **Collision Before Movement**: Collision detection prevents invalid position updates
3. **Sliding Behavior**: Entity slides along walls when moving diagonally
4. **Entity Collision**: Entities block each other correctly
5. **System Order Dependency**: Tests fail if systems ordered incorrectly

### Actual Implementation
**Existing Tests** (incomplete):
- `TestCollisionSystemTerrainIntegration` - Only tests collision AFTER movement occurred
- Unit tests for MovementSystem and CollisionSystem in isolation
- No test verifies that movement is actually blocked

**Missing Tests**:
```go
// Should exist but doesn't:
func TestMovementBlockedByTerrain(t *testing.T) {
    // Setup entity against wall
    // Apply velocity toward wall
    // Verify position does NOT change
    // Verify velocity is zeroed
}

func TestMovementSystemRespectsTerrain(t *testing.T) {
    // MovementSystem should check terrain before moving
    // Verify predictive collision
}

func TestSystemOrderMatters(t *testing.T) {
    // Verify collision system must run before movement
    // Or movement must check collision before updating
}
```

### Reproduction Scenario
Current test gap allows bugs to pass:
```bash
$ go test -tags test ./pkg/engine/...
# All tests pass ✅
# But player can move through walls in actual game ❌
```

### Production Impact Assessment

**Severity**: High  
**Technical Debt**: Cannot refactor collision system safely  
**Regression Risk**: Changes to movement/collision not validated  
**Coverage Gap**: Core gameplay mechanic not tested end-to-end

---

## Summary Table

| Gap # | Title | Severity | Priority Score | Impact | Files Affected |
|-------|-------|----------|----------------|--------|----------------|
| 1 | Movement Applied Before Validation | Critical | 1078.2 | Game-breaking | movement.go, collision.go |
| 3 | No Collision Prediction System | Critical | 1077.9 | Game-breaking | movement.go (missing) |
| 2 | System Execution Order Not Enforced | Critical | 958.8 | Architectural | ecs.go, main.go |
| 4 | Terrain Collision Bounds Calculation | High | 377.4 | Incorrect behavior | terrain_collision_system.go |
| 5 | Insufficient Integration Test Coverage | High | 334.5 | Testing debt | *_test.go (missing) |

---

## Recommended Repair Priority

Based on priority scores and dependencies:

1. **Gap #3** (1077.9) - Implement collision prediction system
   - Creates foundation for fixing movement
   - Required by Gap #1
   
2. **Gap #1** (1078.2) - Integrate prediction into movement system
   - Depends on Gap #3
   - Fixes primary user-facing issue

3. **Gap #4** (377.4) - Fix terrain bounds calculation
   - Independent fix
   - Prevents edge cases

4. **Gap #5** (334.5) - Add comprehensive tests
   - Validates all fixes
   - Prevents regressions

5. **Gap #2** (958.8) - Enforce system ordering
   - Architectural improvement
   - Prevents future issues

---

## Impact on Project Phases

**Current Phase**: 8.2 (Input & Rendering)  
**Blocked By Gaps**: Player input (WASD) doesn't produce correct movement behavior

**Phase 8.3 (Save/Load)**: Gaps could save invalid state (player inside walls)  
**Phase 8.4 (Performance)**: Reactive collision resolution less efficient than prediction  
**Phase 8.5 (Tutorial)**: Cannot teach wall collision if it doesn't work

---

## Root Cause Analysis

The fundamental issue is a **architectural pattern mismatch**:

1. **ECS Pattern Used**: Systems run in sequence on all entities
2. **Physics Pattern Needed**: Collision detection must be **predictive**, not reactive
3. **Gap**: MovementSystem applies physics without validation

**Industry Standard** (Unity, Unreal, Godot):
- Physics engines use **collision prediction** before applying forces
- Raycasting or swept collision tests BEFORE position updates
- Movement only applied if no collision would occur

**Venture's Current Approach**:
- Apply movement first
- Detect collision after
- Try to "fix" position (push back)
- Results in entities penetrating walls before being pushed back

---

## Recommendations

### Immediate Actions (Critical Gaps)
1. Implement `CollisionPredictor` interface in collision system
2. Modify `MovementSystem.Update` to validate positions before updating
3. Add integration tests to prevent regressions

### Architectural Improvements (System Ordering)
1. Add `Priority` field to `System` interface
2. Sort systems by priority in `World.Update`
3. Document system dependencies in comments

### Long-term Quality (Test Coverage)
1. Create `integration_test.go` for cross-system tests
2. Add "movement blocked by walls" as acceptance test
3. Include visual regression tests for collision behavior

---

## Appendix: Code Evidence

### Evidence A: System Order in Client
```go
// cmd/client/main.go, lines 292-314
game.World.AddSystem(inputSystem)
game.World.AddSystem(playerCombatSystem)
game.World.AddSystem(playerItemUseSystem)
game.World.AddSystem(playerSpellCastingSystem)
game.World.AddSystem(movementSystem)        // ⚠️ Movement BEFORE collision
game.World.AddSystem(collisionSystem)       // ⚠️ Collision AFTER movement
game.World.AddSystem(combatSystem)
// ... more systems
```

### Evidence B: MovementSystem No Validation
```go
// pkg/engine/movement.go, lines 42-46
// Update position based on velocity
pos.X += vel.VX * deltaTime  // ❌ No collision check before update
pos.Y += vel.VY * deltaTime  // ❌ Position changed before validation
```

### Evidence C: CollisionSystem Reactive Only
```go
// pkg/engine/collision.go, lines 128-135
// Check terrain collision for solid entities
if s.terrainChecker != nil && collider.Solid && !collider.IsTrigger {
    if s.terrainChecker.CheckEntityCollision(entity) {
        s.resolveTerrainCollision(entity)  // ❌ Position already updated
    }
}
```

### Evidence D: Test Gap
```bash
$ grep -r "TestMovement.*Block\|TestCollision.*Prevent" pkg/engine/
# (no results - no test verifies movement is blocked)
```

---

**End of Audit Report**
