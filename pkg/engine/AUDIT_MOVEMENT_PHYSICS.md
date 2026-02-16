# Engine Movement & Physics Sub-Audit

**Scope**: Movement, collision, projectile, mounting, and fluid physics systems
**Date**: 2026-02-16
**Status**: Complete

## Files Audited

| File | Lines | Description |
|------|-------|-------------|
| `movement.go` | 888 | MovementSystem: velocity, friction, bounds, wall sliding, animation state |
| `collision.go` | 692 | CollisionSystem: spatial grid, quadtree, entity/terrain collision detection |
| `collision_precise.go` | 742 | Pixel-perfect collision: sub-pixel precision, shapes, wall sliding |
| `collision_pool.go` | 46 | Memory pooling for collision query results |
| `projectile_system.go` | 616 | ProjectileSystem: physics, hit detection, bouncing, explosions, trails |
| `projectile_component.go` | 124 | ProjectileComponent: pure data struct with pierce/bounce/explosion |
| `projectile_pool.go` | 222 | Object pools for projectile, velocity, and position components |
| `mounting_system.go` | 338 | MountingSystem: rider-vehicle mounting, position/rotation sync |
| `fluid_physics_system.go` | 264 | FluidPhysicsSystem: buoyancy, swimming, flooding mechanics |

## Issues Found: 4 (0 high, 1 medium, 3 low)

### MED-1: Variable shadowing in `GetCollisionAlignment` — FIXED

**File**: `collision_precise.go:699`
**Description**: The `else if` clause `cComp, hasCollider := entity.GetComponent("collider")` used `:=` which created a new `hasCollider` variable in the inner scope, shadowing the outer `hasCollider` declared on line 684. This meant when an entity had a standard `ColliderComponent` (but no `PreciseColliderComponent`), the outer `hasCollider` was never set to `true`, causing the function to incorrectly return 0 alignment error instead of computing the actual alignment.
**Fix**: Renamed inner variable to `hasStandard` to avoid shadowing.
**Test**: `TestGetCollisionAlignment_StandardCollider`, `TestGetCollisionAlignment_StandardColliderOffset`

### LOW-1: `applySpeedLimit` bypassed cached debug flag — FIXED

**File**: `movement.go:314`
**Description**: `applySpeedLimit` called `log.GetLevel() >= log.DebugLevel` directly instead of using the cached `movementDebugEnabled` flag. The flag was specifically designed (and documented) to avoid per-frame mutex acquisition in `GetLevel()`, which costs ~1µs/frame. This single method bypassed the optimization on every speed-limited entity per frame.
**Fix**: Changed to use `movementDebugEnabled`.

### LOW-2: Multiple movement collision methods bypassed cached debug flag — FIXED

**File**: `movement.go:334,380,394,405,415,436,461,491,543,559,694,708,725,742,774,788`
**Description**: 16 locations in movement.go used `log.GetLevel() >= log.DebugLevel` instead of the cached `movementDebugEnabled` flag. These include `calculateValidPosition`, `handleTerrainCollision`, `tryTerrainWallSlide`, `handleEntityCollisions`, `tryEntityWallSlide`, `applyBoundsConstraints`, `applyFriction`, `updateAnimationState`, `updateFacingDirection`, `SetVelocity`, `GetPosition`, `SetPosition`, `GetDistance`, and `MoveTowards`.
**Fix**: Replaced all direct `log.GetLevel()` calls with `movementDebugEnabled`, preserving the two initialization lines that set the flag.

### LOW-3: Unconditional debug logging in `collision_precise.go` — FIXED

**File**: `collision_precise.go` (multiple functions)
**Description**: Five functions (`ResolveWallCollision`, `CheckPreciseCollision`, `checkPreciseColliders`, `checkStandardColliders`, `GetCollisionAlignment`) logged unconditionally without checking the `collisionDebugEnabled` flag, unlike all other functions in the file that properly guarded debug logging. This caused unnecessary allocations from `logrus.Fields` map creation on every call, even at Info/Warn log levels.
**Fix**: Added `collisionDebugEnabled` guards around all Debug-level log calls. Kept Error-level logs unconditional (they indicate real bugs). Kept Warn-level logs for type assertion failures unconditional (they indicate corrupted state).

## No Issues Found

The following files were audited and found to have no issues:

- **`collision.go`**: Well-structured with proper pooling, quadtree integration, and defensive nil checks.
- **`collision_pool.go`**: Clean pooling implementation with proper cleanup and memory bloat prevention.
- **`projectile_system.go`**: Correct lifecycle management, proper nil checks, good use of typed getters.
- **`projectile_component.go`**: Pure data component with correct pierce/bounce decrement logic.
- **`projectile_pool.go`**: Standard sync.Pool usage with proper reset and nil guards.
- **`mounting_system.go`**: Proper error handling, defensive nil checks, clean mount/dismount lifecycle.
- **`fluid_physics_system.go`**: Correct physics integration, proper error handling on fluid operations.
