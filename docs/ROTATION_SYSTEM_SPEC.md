# 360° Rotation System - Technical Specification

**Version:** 1.0.0 | **Status:** Production | **Date:** February 2026

## Overview

Full directional control with independent movement and aim (dual-stick shooter mechanics).

## Architecture

### RotationComponent

**Purpose:** Entity facing direction in 2D space (radians)

**Fields:**
- `Angle float64`: Current facing [0, 2π), 0=right, π/2=down, π=left, 3π/2=up
- `TargetAngle float64`: Desired facing for smooth rotation
- `AngularVelocity float64`: Rotation speed (rad/s), +clockwise, -counter-clockwise
- `RotationSpeed float64`: Max rotation rate (default 3.0 rad/s)
- `SmoothRotation bool`: Interpolate vs instant snap

**Methods:**
- `NewRotationComponent(angle, speed)` - Constructor
- `SetTargetAngle(angle)` - Smooth rotation target
- `SetAngleImmediate(angle)` - Instant rotation
- `Update(deltaTime) bool` - Interpolate towards target, returns true when complete
- `GetDirectionVector() (x, y)` - Unit vector
- `GetCardinalDirection() int` - Nearest of 8 directions (0-7)

**Rationale:** Radians for precision, separate Angle/TargetAngle for interpolation, cardinal mapping for sprite caching

### AimComponent

**Purpose:** Independent aim direction (separate from movement)

**Fields:**
- `AimAngle float64`: Current aim [0, 2π)
- `AimTarget Vector2D`: World-space aim position
- `HasTarget bool`: Target validity
- `AutoAim bool`: Aim assist (mobile/controller)
- `SnapRadius float64`: Auto-aim max distance (default 100px)
- `AutoAimStrength float64`: Correction amount [0, 1] (default 0.3)

**Methods:**
- `NewAimComponent(angle)` - Constructor
- `SetAimAngle(angle)` - Direct (gamepad right-stick)
- `SetAimTarget(x, y)` - Target-based (mouse/touch)
- `UpdateAimAngle(entityX, entityY) float64` - Calculate angle to target
- `GetAimDirection() (x, y)` - Unit vector
- `GetAttackOrigin(entityX, entityY, weaponOffset) (x, y)` - Projectile spawn
- `ApplyAutoAim(entityX, entityY, enemyX, enemyY) bool` - Aim assist
- `IsAimingAt(entityX, entityY, targetX, targetY, tolerance) bool` - Check accuracy

**Rationale:** Separate from rotation for strafe mechanics, target-based for mouse/touch, auto-aim for mobile/controller parity

### RotationSystem

**Methods:**
- `NewRotationSystem(world)` - Constructor
- `Update(deltaTime)` - Process all "rotation" entities
- `SyncRotationToAim(entityID) bool` - Instant align rotation to aim
- `SetEntityRotation(entityID, angle) bool` - Direct set
- `GetEntityRotation(entityID) (angle, ok)` - Query
- `EnableSmoothRotation(entityID, enabled) bool` - Toggle mode
- `SetRotationSpeed(entityID, speed) bool` - Configure rate

**Update Flow:**
1. Query entities with "rotation" component
2. If "aim" component exists, sync rotation target to aim angle
3. If "position" + aim target exist, update aim angle
4. Call `rotation.Update(deltaTime)` for interpolation
5. Auto-normalize angle to [0, 2π)

**Rationale:** Single-pass update, auto-sync rotation/aim, helper methods for convenience

## Integration

**Input Handling:**
```go
// Mouse aim
if mouseActive {
    aim.SetAimTarget(mouseX, mouseY)
}
// Gamepad aim
if rightStickActive {
    aimAngle := math.Atan2(rightStickY, rightStickX)
    aim.SetAimAngle(aimAngle)
}
```

**Combat System:**
```go
aim := entity.GetComponent("aim").(*AimComponent)
origin := aim.GetAttackOrigin(pos.X, pos.Y, weaponOffset)
direction := aim.GetAimDirection()
projectile := CreateProjectile(origin.X, origin.Y, direction.X, direction.Y)
```

**Animation Sync:**
```go
rotation := entity.GetComponent("rotation").(*RotationComponent)
cardinalDir := rotation.GetCardinalDirection()  // 0-7
sprite := spriteCache.Get(entityType, cardinalDir)
```

**Auto-Aim:**
```go
if aim.AutoAim {
    nearestEnemy := FindNearestEnemy(pos.X, pos.Y, aim.SnapRadius)
    if nearestEnemy != nil {
        aim.ApplyAutoAim(pos.X, pos.Y, nearestEnemy.X, nearestEnemy.Y)
    }
}
```

## Testing

**Unit Tests:** Component state transitions, angle normalization, interpolation accuracy  
**Integration Tests:** Rotation/aim sync, input → combat chain, auto-aim behavior  
**Performance:** <0.1ms per 100 entities, <1% CPU overhead

## Implementation Status

**Complete:**
- ✅ RotationComponent with smooth interpolation
- ✅ AimComponent with target/angle modes
- ✅ RotationSystem with auto-sync
- ✅ Input integration (keyboard, mouse, gamepad)
- ✅ Combat system updates (projectiles use aim direction)
- ✅ 8-direction sprite mapping

**Pending:**
- ⬜ Auto-aim tuning (strength, snap radius)
- ⬜ Mobile touch controls
- ⬜ Animation blend between directions

## Configuration

**Defaults:**
```go
RotationSpeed: 3.0 rad/s         // ~172°/s
AutoAimStrength: 0.3             // 30% correction
SnapRadius: 100px                // Auto-aim range
SmoothRotation: true             // Interpolate by default
```

**Performance Tuning:**
- Disable SmoothRotation for fast-paced enemies
- Reduce RotationSpeed for heavy/slow entities
- Adjust AutoAimStrength per difficulty (0.5 easy, 0.2 hard)

---

**Last Updated:** November 14, 2025
