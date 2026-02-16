# Audit: pkg/narrative/ (branching)

**Date**: 2026-02-16
**Auditor**: Copilot
**Coverage**: 91.2%

## Package Overview

The `pkg/narrative/branching/` package implements procedural branching narrative generation and management for story arcs with player choices, alignment tracking, faction reputation, and consequence systems.

## Audit Summary

| Severity | Count | Fixed |
|----------|-------|-------|
| High     | 0     | 0     |
| Medium   | 1     | 1     |
| Low      | 1     | 0     |

## Issues Found

### Medium

#### 1. Missing nil checks for arc lookups after getProgress (FIXED)
- **Files**: `manager.go` (validateChoiceContext, GetCurrentNode, AdvanceStory)
- **Description**: After `getProgress()` succeeded, arc was fetched from `m.graph.Arcs[arcID]` without existence check. If the arc was removed between `StartArc` and subsequent calls, this would cause a nil pointer dereference panic.
- **Fix**: Added existence checks for both arc and current node lookups in all three methods. Added `TestManagerArcRemovedAfterStart` test covering all three code paths.

### Low

#### 1. CheckConsequences silently returns nil on error
- **File**: `manager.go:262-265`
- **Description**: `CheckConsequences()` returns `nil` when `getProgress()` fails instead of propagating the error. This can mask bugs during debugging. The method signature would need to change to return `([]string, error)` which is a broader API change.
- **Status**: Documented, not fixed (API change needed)

## Strengths

- ✅ Deterministic RNG: Uses `rand.New(rand.NewSource(seed))` correctly, never global rand
- ✅ Thread safety: Manager uses `sync.RWMutex` properly for concurrent access
- ✅ Pure data pattern: `NarrativeComponent.Type()` returns string only, no logic
- ✅ Comprehensive doc.go with usage examples
- ✅ Table-driven tests with good edge case coverage
- ✅ Benchmarks for performance-critical paths
- ✅ Well-factored code with small, focused functions
