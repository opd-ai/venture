# Status Effect Component Object Pooling

## Overview

This document describes the object pooling implementation for `StatusEffectComponent` to reduce GC pressure during combat.

## Problem

During combat, status effects (DoT, buffs, debuffs) are allocated and freed constantly:
- 10-20 allocations per second during active combat
- Causes GC pauses that contribute to frame stuttering
- Each allocation adds overhead even though `StatusEffectComponent` is small

## Solution

Implemented object pooling using Go's `sync.Pool`:
- Reuses `StatusEffectComponent` instances instead of allocating new ones
- Thread-safe concurrent access
- Automatic memory management (pool can shrink during GC if needed)
- Zero-allocation in the hot path

## API Changes

### Before (Direct Allocation)
```go
effect := &StatusEffectComponent{
    EffectType:   "poison",
    Duration:     10.0,
    Magnitude:    5.0,
    TickInterval: 1.0,
    NextTick:     1.0,
}
entity.AddComponent(effect)
```

### After (Pooled Allocation)
```go
effect := NewStatusEffectComponent("poison", 5.0, 10.0, 1.0)
entity.AddComponent(effect)
// ... later when effect expires ...
ReleaseStatusEffect(effect)
```

## Implementation Details

### Files Modified/Created

1. **pkg/engine/status_effect_pool.go** (NEW)
   - `statusEffectPool` - Global sync.Pool instance
   - `NewStatusEffectComponent()` - Acquire from pool
   - `ReleaseStatusEffect()` - Return to pool
   - Pool statistics tracking (optional)

2. **pkg/engine/combat_components.go** (MODIFIED)
   - Added `Reset()` method to `StatusEffectComponent`
   - Clears all fields to prevent memory leaks

3. **pkg/engine/status_effect_system.go** (MODIFIED)
   - Changed `ApplyStatusEffect()` to use `NewStatusEffectComponent()`
   - Added `ReleaseStatusEffect()` call when removing expired effects

4. **pkg/engine/combat_system.go** (MODIFIED)
   - Changed `ApplyStatusEffect()` to use `NewStatusEffectComponent()`

5. **pkg/engine/status_effect_pool_test.go** (NEW)
   - Comprehensive test suite with 13 test functions
   - 6 benchmarks comparing pooled vs direct allocation
   - Concurrent access tests
   - Integration tests

### Usage Pattern

The typical lifecycle of a pooled status effect:

1. **Acquire**: `effect := NewStatusEffectComponent(...)`
2. **Use**: Effect is active on entity, updated each frame
3. **Expire**: Effect duration reaches zero
4. **Release**: `ReleaseStatusEffect(effect)` returns to pool
5. **Reuse**: Next `NewStatusEffectComponent()` call may reuse this instance

### Reset Method

The `Reset()` method clears all fields:
```go
func (s *StatusEffectComponent) Reset() {
    s.EffectType = ""
    s.Duration = 0
    s.Magnitude = 0
    s.TickInterval = 0
    s.NextTick = 0
}
```

This is critical to prevent:
- Memory leaks (dangling string references)
- State contamination (old effect data in new effect)
- Bugs from non-zero initial values

## Performance Impact

### Expected Improvements

Based on PLAN.md targets:
- **Allocations eliminated**: 10-20 per second during combat
- **GC pause reduction**: 30-40% (fewer objects to scan)
- **Frame time improvement**: 0.2-0.5ms (fewer GC pauses)

### Benchmark Results

To measure actual improvement, run:
```bash
# Direct allocation baseline
go test -bench=BenchmarkStatusEffectPool_DirectAllocation -benchmem ./pkg/engine

# Pooled allocation
go test -bench=BenchmarkStatusEffectPool_NewAndRelease -benchmem ./pkg/engine

# Typical combat pattern
go test -bench=BenchmarkStatusEffectPool_TypicalCombatPattern -benchmem ./pkg/engine
```

Expected results:
- Direct allocation: ~48 B/op, 1 alloc/op
- Pooled allocation: ~0 B/op, 0 allocs/op (after warmup)
- 100% allocation reduction in steady state

## Testing

### Unit Tests

Run all pool tests:
```bash
go test -v -run TestStatusEffectPool ./pkg/engine
```

Tests cover:
- Basic acquire/release cycle
- State reset on release
- Object reuse
- Concurrent access (100 goroutines, 100 iterations each)
- Nil safety
- Integration patterns

### Integration Testing

The system was designed to be drop-in compatible. Existing tests should pass without modification because:
- `NewStatusEffectComponent()` creates a fully initialized component
- `ReleaseStatusEffect()` is called automatically when effects expire
- Direct allocation still works (for tests that need it)

## Memory Safety

### Preventing Leaks

1. **Always call Release**: Every `NewStatusEffectComponent()` must be paired with `ReleaseStatusEffect()`
2. **Reset before release**: The `Reset()` method clears references
3. **Idempotent release**: Safe to call `ReleaseStatusEffect()` multiple times
4. **Nil safe**: `ReleaseStatusEffect(nil)` is a no-op

### Avoiding Use-After-Release

Once `ReleaseStatusEffect()` is called, the component should not be used:
```go
effect := NewStatusEffectComponent("poison", 5.0, 10.0, 1.0)
entity.AddComponent(effect)
// ... effect expires ...
entity.RemoveComponent(effect.Type())
ReleaseStatusEffect(effect)
// ❌ DON'T use effect here - it may be reused by another entity
```

## Future Work

### Additional Pooling Candidates

Based on PLAN.md Phase 3 priorities:

1. **ParticleComponent** (Phase 3.2)
   - Similar pattern to StatusEffectComponent
   - High allocation rate (100+ per second with effects)
   - Expected: 20-30% GC pause reduction

2. **Network Buffers** (Phase 3.3)
   - Packet serialization/deserialization
   - Large allocations (4KB per packet)
   - Expected: 10-15% network processing improvement

### Pool Statistics

Currently, pool statistics are tracked but disabled for performance. To enable:
```go
// In status_effect_pool.go
func NewStatusEffectComponent(...) *StatusEffectComponent {
    trackAcquire() // Uncomment to enable tracking
    effect := statusEffectPool.Get().(*StatusEffectComponent)
    // ... rest of function
}
```

Statistics include:
- Total acquired (lifetime)
- Total released (lifetime)
- Active (approximate current usage)

## Compatibility

### Backward Compatibility

The implementation is fully backward compatible:
- Direct allocation still works: `&StatusEffectComponent{...}`
- No changes to component interface
- Existing tests don't need modification
- Gradual migration possible (mix pooled and direct allocation)

### Determinism

Object pooling does not affect determinism:
- `NewStatusEffectComponent()` fully initializes all fields
- `Reset()` clears all state before reuse
- Same seed produces same game state regardless of pooling

## References

- **PLAN.md**: Performance Optimization Plan, Phase 3.1
- **docs/profiling/optimization_progress.md**: Performance tracking
- **Go sync.Pool documentation**: https://pkg.go.dev/sync#Pool

## Author

Performance Optimization Team, based on PLAN.md specifications

## Date

2025-10-28
