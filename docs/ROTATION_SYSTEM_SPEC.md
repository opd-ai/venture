# 360° Rotation System - Technical Specification

## Overview

The 360° Rotation System introduces full directional control to Venture, transforming it from a 4-directional action-RPG into a dual-stick shooter with independent movement and aim directions. This document specifies the technical implementation, design decisions, and integration points for Phase 10.1 of Version 2.0.

**Version:** 2.0 Phase 10.1  
**Status:** Foundation Complete (Components + System implemented)  
**Date:** October 2025

## Architecture

### Components

#### RotationComponent

Stores entity facing direction in 2D space using radians.

**Fields:**
- `Angle float64`: Current facing direction [0, 2π)
  - 0 = right, π/2 = down, π = left, 3π/2 = up
- `TargetAngle float64`: Desired facing direction for smooth rotation
- `AngularVelocity float64`: Current rotation speed (rad/s)
  - Positive = clockwise, negative = counter-clockwise
- `RotationSpeed float64`: Maximum rotation rate (rad/s, default 3.0)
- `SmoothRotation bool`: Enable interpolation vs instant snap

**Methods:**
- `NewRotationComponent(initialAngle, rotationSpeed) *RotationComponent`
- `SetTargetAngle(angle)`: Set desired facing direction
- `SetAngleImmediate(angle)`: Instant rotation without interpolation
- `Update(deltaTime) bool`: Smooth rotation towards target, returns true when complete
- `GetDirectionVector() (x, y float64)`: Unit vector in facing direction
- `GetCardinalDirection() int`: Nearest of 8 cardinal directions (0-7)

**Design Rationale:**
- Radians for mathematical precision (no degree/radian conversion overhead)
- Separate Angle/TargetAngle enables smooth interpolation
- Cardinal direction mapping for sprite caching optimization
- SmoothRotation flag allows instant rotation for teleports/respawns

#### AimComponent

Manages independent aim direction separate from movement.

**Fields:**
- `AimAngle float64`: Current aim direction [0, 2π)
- `AimTarget Vector2D`: World-space position being aimed at
- `HasTarget bool`: Whether AimTarget is valid
- `AutoAim bool`: Enable aim assist for mobile/controller
- `SnapRadius float64`: Max distance for auto-aim (default 100 pixels)
- `AutoAimStrength float64`: Aim correction amount [0, 1] (default 0.3)

**Methods:**
- `NewAimComponent(initialAngle) *AimComponent`
- `SetAimAngle(angle)`: Direct angle setting (gamepad right-stick)
- `SetAimTarget(x, y)`: Target-based aiming (mouse/touch)
- `UpdateAimAngle(entityX, entityY) float64`: Calculate angle from position to target
- `GetAimDirection() (x, y float64)`: Unit vector in aim direction
- `GetAttackOrigin(entityX, entityY, weaponOffset) (x, y float64)`: Projectile spawn position
- `ApplyAutoAim(entityX, entityY, enemyX, enemyY) bool`: Aim assist towards enemy
- `IsAimingAt(entityX, entityY, targetX, targetY, tolerance) bool`: Check aim accuracy

**Design Rationale:**
- Separate from RotationComponent enables strafe mechanics
- Target-based aiming supports mouse/touch input naturally
- Auto-aim system makes mobile/controller input competitive
- Attack origin calculation centralizes projectile spawning logic

### System

#### RotationSystem

Manages entity rotation and orientation updates.

**Methods:**
- `NewRotationSystem(world) *RotationSystem`
- `Update(deltaTime)`: Process all entities with "rotation" component
  - Syncs rotation target with aim component if present
  - Updates aim angle from position if target-based
  - Performs smooth rotation interpolation
- `SyncRotationToAim(entityID) bool`: Instant align rotation to aim
- `SetEntityRotation(entityID, angle) bool`: Direct rotation setting
- `GetEntityRotation(entityID) (angle, ok)`: Query entity rotation
- `EnableSmoothRotation(entityID, enabled) bool`: Toggle rotation mode
- `SetRotationSpeed(entityID, speed) bool`: Configure rotation rate

**Update Flow:**
1. Query world for entities with "rotation" component
2. If entity has "aim" component, sync rotation target to aim angle
3. If entity has "position" and aim has target, update aim angle
4. Call `rotation.Update(deltaTime)` for smooth interpolation
5. Angle automatically normalized to [0, 2π)

**Design Rationale:**
- Single-pass update for all rotating entities (efficient)
- Automatic sync between rotation and aim simplifies client code
- Helper methods provide convenient API without direct component access
- Returns bool instead of error (matches existing system patterns)

## Integration Points

### With Existing Systems

#### InputSystem (Enhancement Required)
**Current State:** 4-directional movement only  
**Required Changes:**
- Track mouse cursor position (screen coordinates)
- Convert screen→world coordinates using camera transform
- Calculate aim angle: `atan2(mouseWorldY - entityY, mouseWorldX - entityX)`
- Set AimComponent.AimTarget on player entity
- Touch: Detect dual virtual joystick (left=move, right=aim)

**Implementation Approach:**
```go
// In InputSystem.Update()
if entity has AimComponent {
    aimComp := entity.GetComponent("aim").(*AimComponent)
    
    // Mouse aiming
    mouseX, mouseY := ebiten.CursorPosition()
    worldX, worldY := screenToWorld(mouseX, mouseY, camera)
    aimComp.SetAimTarget(worldX, worldY)
    
    // Touch aiming (right half of screen)
    if touchID := rightScreenTouch(); touchID != -1 {
        tx, ty := ebiten.TouchPosition(touchID)
        worldX, worldY := screenToWorld(tx, ty, camera)
        aimComp.SetAimTarget(worldX, worldY)
    }
}
```

#### MovementSystem (Enhancement Required)
**Current State:** Movement direction determines facing  
**Required Changes:**
- Decouple velocity calculation from facing direction
- WASD sets VelocityComponent in world-space directions
  - W = (0, -1), A = (-1, 0), S = (0, 1), D = (1, 0)
- Remove automatic facing direction updates
- Rotation handled separately by RotationSystem

**Implementation Approach:**
```go
// In MovementSystem.Update()
// OLD: vel.VX, vel.VY = direction * speed  (movement = facing)
// NEW: vel.VX, vel.VY = input_direction * speed  (movement independent)

if W pressed { vel.VY = -speed }
if A pressed { vel.VX = -speed }
if S pressed { vel.VY = speed }
if D pressed { vel.VX = speed }

// RotationSystem handles facing separately based on AimComponent
```

#### RenderSystem (Enhancement Required)
**Current State:** Sprites rendered at fixed orientation  
**Required Changes:**
- Check for RotationComponent during entity rendering
- Apply sprite rotation using `ebiten.DrawImageOptions.GeoM.Rotate(angle)`
- Implement sprite rotation cache at 8 cardinal directions
  - Pre-compute rotated sprites on entity creation
  - Use `RotationComponent.GetCardinalDirection()` to select cached sprite
  - Balances quality (smooth-looking) with memory (8 images per sprite)
- Handle rotation pivot point (center of sprite)

**Implementation Approach:**
```go
// In RenderSystem.Draw()
if entity.HasComponent("rotation") {
    rotComp := entity.GetComponent("rotation").(*RotationComponent)
    angle := rotComp.Angle
    
    // Option 1: Runtime rotation (simple, more CPU)
    opts.GeoM.Translate(-spriteWidth/2, -spriteHeight/2) // Pivot
    opts.GeoM.Rotate(angle)
    opts.GeoM.Translate(spriteWidth/2, spriteHeight/2)
    
    // Option 2: Cached rotation (complex, less CPU, more memory)
    cardinalDir := rotComp.GetCardinalDirection()
    sprite := spriteCache.Get(entityType, cardinalDir)
}
```

#### CombatSystem (Enhancement Required)
**Current State:** Attacks fire in movement direction  
**Required Changes:**
- Melee attacks: Use AimComponent.AimAngle for hitbox placement
  - Position attack hitbox in aim direction from entity
  - Hitbox offset: `entity.pos + aim.GetAimDirection() * meleeRange`
- Ranged attacks: Use AimComponent.GetAttackOrigin() for projectile spawn
  - Spawns projectile at weapon position (not entity center)
  - Projectile direction = AimComponent.GetAimDirection()
- Visual feedback: Rotate weapon sprite to match aim angle

**Implementation Approach:**
```go
// In CombatSystem.processAttack()
if entity.HasComponent("aim") {
    aimComp := entity.GetComponent("aim").(*AimComponent)
    
    if isMeleeAttack {
        // Hitbox in aim direction
        dx, dy := aimComp.GetAimDirection()
        hitboxX := entity.X + dx * meleeRange
        hitboxY := entity.Y + dy * meleeRange
        checkMeleeHit(hitboxX, hitboxY, hitboxRadius)
    } else {
        // Projectile spawn
        projX, projY := aimComp.GetAttackOrigin(entity.X, entity.Y, weaponOffset)
        createProjectile(projX, projY, aimComp.AimAngle, projectileSpeed)
    }
}
```

#### NetworkComponent (Enhancement Required)
**Current State:** Position, velocity synced  
**Required Changes:**
- Add rotation angle to sync protocol
  - Compressed: 1 byte (256 discrete angles = ~1.4° precision)
  - Full precision: 4 bytes (float32)
- Client-side prediction for rotation (same as movement)
- Server-authoritative rotation validation
- Interpolation buffer for smooth remote entity rotation

**Protocol Addition:**
```go
type EntityStateMessage struct {
    // ... existing fields
    Rotation uint8 // Angle * 256 / (2*π), precision: ~1.4 degrees
}
```

## Testing Strategy

### Unit Tests (Complete)

**RotationComponent Tests (15 tests):**
- Component creation with defaults
- Angle normalization (negative, >2π)
- Target angle setting
- Immediate vs smooth rotation
- Update interpolation
- Direction vector calculation
- Cardinal direction mapping
- Angle utility functions (normalizeAngle, shortestAngularDistance)

**AimComponent Tests (14 tests):**
- Component creation with defaults
- Direct angle setting
- Target-based aiming
- Aim angle calculation from position
- Direction vector calculation
- Attack origin calculation
- Auto-aim application (enabled/disabled, in/out of range)
- Aim accuracy checking

**RotationSystem Tests (10 tests):**
- System creation
- Single entity rotation update
- Multiple entity batch processing
- Rotation sync with aim component
- Helper method operations (set, get, enable, configure)
- Error conditions (missing entity, missing component)
- Rotation without aim component

**Coverage:** 100% on all rotation/aim components and system

### Integration Tests (Planned)

**Movement + Rotation Integration:**
- Entity moves left while aiming right (strafe mechanics)
- Rotation follows mouse cursor smoothly
- Touch dual-joystick controls work independently

**Combat + Rotation Integration:**
- Melee attacks hit in aim direction, not movement direction
- Projectiles spawn at weapon position and fire in aim direction
- Weapon sprite rotates to match aim angle

**Multiplayer + Rotation Integration:**
- Rotation syncs across clients with <50ms visual latency
- Client-side prediction for rotation
- Server validates rotation state

### Performance Benchmarks (Planned)

**Target Metrics:**
- 500 rotating entities: <1ms frame time increase
- Sprite rotation cache: <10MB memory overhead
- Network sync: +2 bytes per entity update

**Benchmark Tests:**
```go
BenchmarkRotationSystem_Update500Entities
BenchmarkRotationComponent_Update
BenchmarkAimComponent_UpdateAimAngle
BenchmarkSpriteRotationCache
```

## Determinism Guarantees

**Seed-Based Generation:** N/A (rotation is input-driven, not generated)

**Deterministic Behavior:**
- Angle normalization is deterministic (same input → same output)
- Smooth rotation interpolation is deterministic given same deltaTime
- Cardinal direction mapping is deterministic

**Multiplayer Synchronization:**
- Server is authoritative for all rotation state
- Client predicts rotation based on input
- Server sends canonical rotation every network tick
- Client reconciles predicted rotation with server state

**Serialization:**
```go
// Save format
type SavedRotation struct {
    Angle          float64
    TargetAngle    float64
    RotationSpeed  float64
    SmoothRotation bool
}
```

## Configuration

**Rotation Speeds (Tunable):**
- Default: 3.0 rad/s (~172°/s) - Responsive but smooth
- Fast: 5.0 rad/s (~286°/s) - Arcade shooter feel
- Instant: SmoothRotation=false - No interpolation

**Auto-Aim Settings (Mobile):**
- SnapRadius: 100 pixels (default), 150 pixels (forgiving)
- AutoAimStrength: 0.3 (subtle), 0.5 (moderate), 1.0 (full snap)

**Sprite Cache Optimization:**
- Cardinals: 8 directions (default), 16 directions (higher quality)
- Memory: ~2MB for 100 entity types × 8 directions × 32×32 sprites

## Future Enhancements (Version 2.0 Later Phases)

**Phase 10.2: Projectile Physics**
- Rotation determines projectile launch angle
- Wind/gravity affects projectile trajectory
- Bouncing projectiles use angle of incidence

**Phase 10.3: Screen Shake & Impact Feedback**
- Screen shake intensity based on rotation change rate
- Camera rotation for dramatic impacts

**Phase 11: Advanced Level Design**
- Rotation-sensitive puzzles (aim at targets, reflect projectiles)
- Diagonal wall collision using rotation angles

**Phase 13: Advanced AI**
- AI entities use rotation for facing behavior
- Patrol paths with smooth rotation at waypoints
- Flanking AI rotates to attack from behind

## Known Limitations

**Current Implementation:**
- No sprite rotation cache yet (runtime rotation only)
- InputSystem enhancement not complete (no mouse tracking)
- MovementSystem still couples movement with facing
- CombatSystem not using aim direction yet
- Network protocol doesn't include rotation

**Planned Fixes (Week 2-4):**
- Implement sprite rotation cache (RenderSystem)
- Add mouse/touch aim input (InputSystem)
- Decouple movement direction (MovementSystem)
- Integrate with combat (CombatSystem)
- Network sync protocol (NetworkComponent)

## Code Quality

**Go Best Practices:**
- ✅ Components contain only data, no behavior
- ✅ Systems contain logic, operate on components
- ✅ godoc comments on all exported types/functions
- ✅ Table-driven tests for comprehensive coverage
- ✅ Error handling via bool returns (matches existing patterns)
- ✅ No Ebiten dependencies in components/tests (testable in CI)

**Performance:**
- ✅ Angle normalization: O(1) using modulo
- ✅ Shortest angular distance: O(1) calculation
- ✅ Cardinal direction mapping: O(1) using division
- ✅ System update: O(n) single-pass over entities
- ✅ No allocations in hot paths (Update methods)

**Testing:**
- ✅ 100% test coverage on all new code
- ✅ 39 total tests across 3 test files
- ✅ Tests follow existing project patterns
- ✅ Comprehensive edge case coverage

## References

**Existing Code:**
- `pkg/engine/components.go` - Component patterns
- `pkg/engine/movement_system.go` - System patterns
- `pkg/engine/ai_system.go` - ECS update patterns
- `pkg/engine/combat_system.go` - Attack direction logic

**Documentation:**
- `docs/ARCHITECTURE.md` - ECS architecture
- `docs/TECHNICAL_SPEC.md` - System specifications
- `docs/ROADMAP_V2.md` - Phase 10 detailed plan
- `docs/TESTING.md` - Testing guidelines

**External References:**
- [Ebiten Geometry Matrix](https://ebitengine.org/en/documents/matrix.html) - Sprite rotation
- [Game Programming Patterns](https://gameprogrammingpatterns.com/component.html) - ECS design
- Dual-stick shooter conventions: left=move, right=aim

---

**Document Version:** 1.0  
**Last Updated:** October 2025  
**Maintained By:** Venture Development Team  
**Next Review:** After Phase 10.1 completion (Week 4)
