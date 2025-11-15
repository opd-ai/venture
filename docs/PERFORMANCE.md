# Performance Optimization Task

## Objective
Identify and implement ONE safe, high-impact performance optimization in the Venture codebase.

## Execution Mode
**Autonomous Action** - Implement the optimization directly after analysis.

## Scope & Constraints
- **Target**: Focus on hot paths: game loop, rendering (viewport culling, batch rendering, sprite caching, pooling), entity queries, or network serialization
- **Safety Requirements**:
  - Zero API changes (no signature modifications)
  - No behavioral regressions (output must remain identical)
  - Maintain deterministic generation (preserve seed-based reproducibility)
  - Preserve test coverage (all existing tests must pass)
- **Priority Areas** (in order):
  1. Allocation reduction in game loop/render paths
  2. Redundant computation elimination (caching opportunities)
  3. Algorithmic improvements (O(n²) → O(n log n))
  4. Object pooling for frequently-allocated types

## Analysis Process
1. Review recent profiling data: `cov_images.out`, `coverage.out`
2. Examine rendering systems: `pkg/rendering/cache/`, `pkg/rendering/pool/`
3. Check ECS hot paths: `pkg/engine/` (Update methods, entity queries)
4. Scan for allocation patterns: `make([]T, 0)` in loops, repeated allocations

## Implementation Steps
1. Create benchmark test for target code path
2. Record baseline performance (time/op, allocs/op)
3. Implement optimization
4. Verify improvement with benchmark (minimum 15% improvement)
5. Run full test suite: `go test ./...`
6. Verify determinism if generation code modified

## Success Criteria
- Benchmark shows ≥15% improvement (time or allocations)
- All tests pass: `go test ./...`
- No race conditions: `go test -race ./...`
- Code coverage maintained or improved
- Deterministic behavior preserved (if applicable)

## Output Format
Provide:
1. **Optimization identified**: File, function, specific issue
2. **Benchmark results**: Before/after comparison
3. **Implementation summary**: What changed and why
4. **Verification**: Test results confirming safety

## Notes
- Current performance: 106 FPS with 2000 entities, 73MB memory
- Existing optimizations: viewport culling (1,635x), batch rendering (1,667x), sprite caching (37x), pooling (2x)
- Avoid premature optimization - profile before acting