# Performance Benchmarks Workflow

This document describes the automated performance validation workflow that ensures Venture meets its documented performance targets.

## Overview

The benchmark infrastructure uses shell scripts to validate two critical performance targets documented in `docs/ARCHITECTURE.md`:

1. **60 FPS minimum framerate** with 2000 entities
2. **<500MB client memory** usage

## Benchmark Scripts

### 1. FPS Target Validation

**Script:** Run manually or integrate into CI

**Benchmark:** `BenchmarkFPS2000Entities` in `pkg/benchmark/fps/`
- Creates 2000 entities with Position, Velocity, Health, and Stats components
- Uses MovementSystem for realistic update overhead
- Measures nanoseconds per frame update

**Target:** ≤16,666,666 ns/op (60 FPS)
**Current Performance:** ~130,000 ns/op (~7,800 FPS) ✅

**Execution:**
```bash
go test -bench=BenchmarkFPS2000Entities ./pkg/benchmark/fps/
```

**Failure Condition:** If ns/op exceeds 16,666,666 (below 60 FPS)

### 2. Memory Target Validation

**Script:** `scripts/benchmark-memory.sh`

**Test:** `TestMemoryBounds2000Entities` in `pkg/benchmark/memory/`
- Creates 2000 entities with multiple components
- Measures heap allocation after stabilization
- Compares against 500MB target

**Target:** <500MB heap allocation
**Current Performance:** ~2.5MB (0.5% of budget) ✅

**Execution:**
```bash
./scripts/benchmark-memory.sh
```

**Failure Condition:** If heap allocation ≥500MB

### 3. Regression Detection

**Script:** `scripts/benchmark-regression.sh`

**Purpose:** Detects performance regressions across all benchmarks.

**Action:** Compares current benchmarks against baseline stored in `scripts/benchmark-baseline.json`
- Captures all benchmark results for historical comparison
- Detects regressions >10% slower than baseline
- Helps identify performance degradation across all systems

**Execution:**
```bash
./scripts/benchmark-regression.sh
```

## CI/CD Integration

The benchmark infrastructure can be integrated into CI/CD pipelines:

**Recommended Approach:**
1. Run `scripts/benchmark-memory.sh` on each PR to validate memory targets
2. Run FPS benchmarks periodically (weekly/monthly) due to X11 display requirement
3. Use `scripts/benchmark-regression.sh` to detect gradual performance degradation

**Display Server Requirement:**
- FPS benchmarks require X11/Wayland display (Ebiten dependency)
- Memory benchmarks are headless-compatible and suitable for standard CI
- For headless FPS testing, use Xvfb: `xvfb-run go test -bench=. ./pkg/benchmark/fps/`

## Interpreting Results

### FPS Calculation

```bash
FPS = 1,000,000,000 ns/s ÷ ns/op
```

Example:
- 16,666,666 ns/op = 60 FPS (minimum target)
- 130,000 ns/op = 7,692 FPS (current performance)

### Memory Measurement

The test uses `runtime.MemStats.Alloc` after `runtime.GC()` to measure heap allocation:

```go
runtime.GC()
var mem runtime.MemStats
runtime.ReadMemStats(&mem)
totalMemoryMB := float64(mem.Alloc) / (1024 * 1024)
```

## Performance Targets Documentation

The targets validated by this workflow are documented in:
- `docs/ARCHITECTURE.md` - Performance requirements (60 FPS, <500MB)
- Project README.md context section - Performance guidelines

## Benchmark Implementation

### Test File: `pkg/engine/performance_targets_test.go`

Contains three key functions:

1. **`BenchmarkWorldUpdateWith2000Entities`**
   - Validates core ECS update performance
   - Uses MovementSystem only (lightweight)
   - Measures frame update time with 2000 entities

2. **`TestMemoryBounds2000Entities`**
   - Validates memory usage under target
   - Creates realistic entity mix
   - Measures heap after stabilization

3. **`BenchmarkWorldUpdateWith2000EntitiesRealistic`**
   - More comprehensive benchmark with varied components
   - Includes multiple component types per entity
   - Useful for regression detection

### Design Decisions

**Why MovementSystem only?**
- Collision detection performance depends heavily on spatial partitioning configuration
- MovementSystem represents core ECS overhead without spatial complexity
- Collision is tested separately with specific spatial grid setups

**Why 2000 entities?**
- Documented target in ARCHITECTURE.md
- Represents realistic gameplay scenario
- Historical performance benchmarks used this count

**Why separate memory test?**
- Memory measurement requires different tooling (`runtime.MemStats`)
- Benchmarks focus on time, tests validate absolute bounds
- Allows clear pass/fail criteria for CI

## Related Files

- `pkg/benchmark/fps/` - FPS benchmark tests (see fps/README.md)
- `pkg/benchmark/memory/` - Memory validation tests (see memory/README.md)
- `docs/ARCHITECTURE.md` - Performance target documentation
- `scripts/benchmark-memory.sh` - Memory validation script
- `scripts/benchmark-baseline.json` - Benchmark baselines for regression detection
- `scripts/benchmark-regression.sh` - Regression detection script

## Future Enhancements

Potential improvements to the benchmark infrastructure:

1. **GitHub Actions Workflow:** Create `.github/workflows/benchmark.yml` to automate benchmark execution on PR/push events
2. **Trend Analysis:** Enhanced historical comparison beyond current baseline approach
3. **Platform Matrix:** Run on multiple OS/architectures
4. **Collision Benchmarks:** Add optimized collision system benchmarks
5. **Rendering Benchmarks:** Add sprite generation/rendering performance tests
6. **Network Benchmarks:** Add packet serialization performance tests

## Maintenance

When updating performance targets:

1. Update targets in `docs/ARCHITECTURE.md`
2. Update constants in `pkg/benchmark/fps/fps_test.go` (MAX_NS_PER_FRAME)
3. Update threshold in `pkg/benchmark/memory/memory_test.go`
4. Update `scripts/benchmark-baseline.json` baseline values
5. Update this documentation

When adding new benchmarks:

1. Add to `pkg/benchmark/fps/` or `pkg/benchmark/memory/` with appropriate prefix
2. Update `scripts/benchmark-baseline.json` with new baseline
3. Document in respective README.md files
4. Update this documentation
