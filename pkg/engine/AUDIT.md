# Audit: github.com/opd-ai/venture/pkg/engine
**Date**: 2026-02-14
**Status**: Needs Work

## Summary
The engine package is the 240K+ LOC core of the ECS system containing 382 production Go files and 100+ systems. **Build currently fails** with 11+ compile errors across 4 files due to undefined types/constants and API mismatches. Package exhibits extensive ECS component method violations (200+ instances) and time-based non-determinism in 15+ systems. Critical build errors prevent all compilation and testing.

## AUDIT SUMMARY
| Category | Count |
|----------|-------|
| CRITICAL BUG | 7 |
| FUNCTIONAL MISMATCH | 2 |
| MISSING FEATURE | 3 |
| EDGE CASE BUG | 0 |
| PERFORMANCE ISSUE | 0 |
| Previously Identified (pending) | 6 |

## DETAILED FINDINGS

### CRITICAL BUG: Undefined EquipmentSlot Constant SlotHelmet
**File:** combat_equipment_durability_particle_system.go:107
**Severity:** High
**Description:** The system references `SlotHelmet` which does not exist. The actual constant is `SlotHead` as defined in inventory_components.go:134.
**Expected Behavior:** Code should compile using valid EquipmentSlot constants
**Actual Behavior:** Build fails with `undefined: SlotHelmet`
**Impact:** Package cannot compile, blocking all dependent packages
**Reproduction:** Run `go build ./pkg/engine/...`
**Code Reference:**
```go
armorSlots := []EquipmentSlot{SlotChest, SlotHelmet, SlotBoots, SlotGloves, SlotLegs}
// SlotHelmet does not exist - should be SlotHead
```

~~~~

### CRITICAL BUG: Undefined particles.SpawnConfig Type
**File:** combat_equipment_durability_particle_system.go:190,212
**Severity:** High
**Description:** The system uses `particles.SpawnConfig` struct which does not exist in the `pkg/rendering/particles` package. The particles package uses different configuration types for spawning particles.
**Expected Behavior:** Particle spawning should use the correct type from the particles package
**Actual Behavior:** Build fails with `undefined: particles.SpawnConfig`
**Impact:** Package cannot compile
**Reproduction:** Run `go build ./pkg/engine/...`
**Code Reference:**
```go
config := particles.SpawnConfig{
    Type:           particles.ParticleSparkle,
    Count:          count,
    // ... SpawnConfig does not exist in particles package
}
```

~~~~

### CRITICAL BUG: Undefined Entity.GetCachedPosition Method
**File:** combat_equipment_durability_particle_system.go:382
**Severity:** High
**Description:** The system calls `entity.GetCachedPosition()` but `Entity` type has no such method. Position retrieval should use the component system via `entity.GetComponent("position")`.
**Expected Behavior:** Entity should provide position via component system
**Actual Behavior:** Build fails with `entity.GetCachedPosition undefined (type *Entity has no field or method GetCachedPosition)`
**Impact:** Package cannot compile
**Reproduction:** Run `go build ./pkg/engine/...`
**Code Reference:**
```go
func (s *CombatEquipmentDurabilityParticleSystem) getEntityPosition(entity *Entity) (float64, float64) {
    if pos := entity.GetCachedPosition(); pos != nil {  // Method does not exist
        return pos.X, pos.Y
    }
    return 0, 0
}
```

~~~~

### CRITICAL BUG: Undefined particles.ZLayerEffects Constant
**File:** fishing_line_tension_particle_system.go:307
**Severity:** High
**Description:** The system references `particles.ZLayerEffects` constant which does not exist in the particles package.
**Expected Behavior:** Z-layer constant should exist or different API should be used
**Actual Behavior:** Build fails with `undefined: particles.ZLayerEffects`
**Impact:** Package cannot compile
**Reproduction:** Run `go build ./pkg/engine/...`
**Code Reference:**
```go
ZLayer:   particles.ZLayerEffects,  // Constant does not exist
```

~~~~

### CRITICAL BUG: World.GetEntity Return Value Mismatch
**File:** terrain_companion_health_regen_system.go:312, weather_block_chance_system.go:237
**Severity:** High
**Description:** Multiple systems assign `World.GetEntity()` return value to a single variable, but the method returns 2 values `(*Entity, bool)`. This causes assignment mismatch errors.
**Expected Behavior:** Systems should handle both return values: `entity, exists := s.world.GetEntity(entityID)`
**Actual Behavior:** Build fails with `assignment mismatch: 1 variable but s.world.GetEntity returns 2 values`
**Impact:** Package cannot compile
**Reproduction:** Run `go build ./pkg/engine/...`
**Code Reference:**
```go
// terrain_companion_health_regen_system.go:312
entity := s.world.GetEntity(companionID)  // Should be: entity, _ := s.world.GetEntity(companionID)

// weather_block_chance_system.go:237
entity := s.world.GetEntity(entityID)     // Should be: entity, _ := s.world.GetEntity(entityID)
```

~~~~

### CRITICAL BUG: Undefined Weather Type Constants
**File:** weather_block_chance_system.go:187,196,199
**Severity:** High
**Description:** The system references weather type constants (`WeatherBlizzard`, `WeatherStorm`, `WeatherClear`) that do not exist in the particles package. The particles package only defines: WeatherRain, WeatherSnow, WeatherFog, WeatherDust, WeatherAsh, WeatherNeonRain, WeatherSmog, WeatherRadiation, WeatherSandstorm, WeatherBloodRain.
**Expected Behavior:** Code should use valid WeatherType constants from particles package
**Actual Behavior:** Build fails with `undefined: particles.WeatherBlizzard`, `undefined: particles.WeatherStorm`, `undefined: particles.WeatherClear`
**Impact:** Package cannot compile
**Reproduction:** Run `go build ./pkg/engine/...`
**Code Reference:**
```go
case particles.WeatherBlizzard:  // Does not exist
    return -0.20
case particles.WeatherStorm:     // Does not exist
    return -0.18
case particles.WeatherClear:     // Does not exist
    return 0.0
```

~~~~

### FUNCTIONAL MISMATCH: Equipment Slot Naming Inconsistency
**File:** combat_equipment_durability_particle_system.go:107
**Severity:** Medium
**Description:** Documentation and some code references use "Helmet" terminology while the actual constant is `SlotHead`. This creates confusion and inconsistency across the codebase.
**Expected Behavior:** Consistent naming convention across all references
**Actual Behavior:** Mixed "Helmet" and "Head" terminology for the same slot
**Impact:** Code readability and maintainability issues; direct build failure in this case
**Reproduction:** Search for "Helmet" vs "SlotHead" across codebase

~~~~

### FUNCTIONAL MISMATCH: Particle Spawn API Mismatch
**File:** combat_equipment_durability_particle_system.go, fishing_line_tension_particle_system.go
**Severity:** Medium
**Description:** Several systems were written expecting a different particle spawning API (`SpawnConfig` struct, `ZLayerEffects` constant) than what exists in the particles package. This suggests either the particles API changed without updating dependent systems, or systems were written against planned but unimplemented API.
**Expected Behavior:** Systems should use the actual particles package API
**Actual Behavior:** Systems reference non-existent types and constants
**Impact:** Multiple systems cannot compile
**Reproduction:** Compare system code against `pkg/rendering/particles` exports

~~~~

### MISSING FEATURE: particles.SpawnConfig Struct
**File:** pkg/rendering/particles (expected)
**Severity:** High
**Description:** The `SpawnConfig` struct is used by multiple engine systems but does not exist in the particles package. Either the type needs to be added to particles, or systems need refactoring.
**Expected Behavior:** Particles package should provide a unified spawn configuration type
**Actual Behavior:** Type does not exist, causing build failures
**Impact:** Multiple systems cannot compile
**Reproduction:** Check particles package exports

~~~~

### MISSING FEATURE: Weather Type Constants (Blizzard, Storm, Clear)
**File:** pkg/rendering/particles/weather.go
**Severity:** Medium
**Description:** Weather system code expects additional weather types that aren't defined. The particles package has 10 weather types but is missing Blizzard, Storm, and Clear.
**Expected Behavior:** All referenced weather types should exist
**Actual Behavior:** Missing 3 weather type constants
**Impact:** weather_block_chance_system.go cannot compile
**Reproduction:** Check WeatherType constants in particles/weather.go

~~~~

### MISSING FEATURE: Entity Position Caching API
**File:** pkg/engine/ecs.go
**Severity:** Medium
**Description:** The `Entity` type does not provide a `GetCachedPosition()` method for hot-path position access. Systems requiring fast position access must use the full component system.
**Expected Behavior:** Entity provides convenient position accessor for hot paths
**Actual Behavior:** Method does not exist, systems fail to compile
**Impact:** combat_equipment_durability_particle_system.go cannot compile
**Reproduction:** Check Entity methods in ecs.go

~~~~

## Previously Identified Issues (Still Pending)

The following issues from the 2026-02-13 audit remain unresolved:

- [ ] **high** **ECS compliance** — 200+ component behavior methods violating ECS purity (components must be pure data, systems own behavior)
- [ ] **high** **deterministic procgen** — BountySystem uses `time.Now()` for expiration/timestamps (`bounty_system.go`)
- [ ] **high** **deterministic procgen** — ChallengeSystem uses `time.Now()` for time-based game state (`challenge_system.go`)
- [ ] **high** **deterministic procgen** — ConversationManager generates UUIDs via `crypto/rand.Reader` (`conversation_manager.go`)
- [ ] **high** **error handling** — Enhanced chat system hardcodes placeholder encryption key (`enhanced_chat_system.go:76`)
- [ ] **med** **stub/incomplete** — Multiple placeholder implementations throughout package

## Test Coverage
Build failed - coverage unmeasurable. Target: 65%+. Estimated: 80%+ based on 354 `*_test.go` files.

## Integration Status
The engine package is the ECS core - all other packages depend on it. **Package currently cannot be imported due to build failures.** Systems are registered via `system_init.go` initialization. Components lacking Serialize/Deserialize for persistence: `AimComponent`, `ScreenShakeComponent`, `HitStopComponent`, `BehaviorTreeComponent`, `ExpressionComponent`, and 20+ others.

## Recommendations

### Priority 1: Fix Build Errors (Blocking)
1. **combat_equipment_durability_particle_system.go:107** — Change `SlotHelmet` to `SlotHead`
2. **combat_equipment_durability_particle_system.go:190,212** — Refactor to use existing particles package API or add `SpawnConfig` to particles
3. **combat_equipment_durability_particle_system.go:382** — Replace `GetCachedPosition()` with component-based position retrieval
4. **fishing_line_tension_particle_system.go:307** — Remove or replace `ZLayerEffects` reference
5. **terrain_companion_health_regen_system.go:312** — Change to `entity, _ := s.world.GetEntity(companionID)`
6. **weather_block_chance_system.go:187,196,199** — Either add missing weather constants to particles package or remove unsupported cases
7. **weather_block_chance_system.go:237** — Change to `entity, _ := s.world.GetEntity(entityID)`

### Priority 2: API Alignment
1. Decide on particles spawning API — either add `SpawnConfig` to particles or refactor all consuming systems
2. Add missing weather type constants if needed for game design
3. Consider adding Entity position caching for hot-path access

### Priority 3: Previously Identified Issues
1. Refactor component behavior methods to systems (ECS compliance)
2. Replace `time.Now()` with deterministic game clock
3. Replace placeholder encryption key
4. Complete placeholder implementations
