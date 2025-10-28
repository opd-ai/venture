# Performance Profiling Guide

**Version:** 1.0  
**Last Updated:** October 28, 2025  
**Status:** Production Ready

## Overview

This guide explains how to use Venture's performance profiling tools to identify bottlenecks, validate optimizations, and ensure the game meets its 60 FPS performance target.

## Quick Start

### Running a Basic Assessment

```bash
# Run 30-second assessment with 2000 entities
go run ./cmd/perfprofile -duration=30s -entities=2000

# Output: docs/profiling/assessment_report.md
```

### Generating CPU/Memory Profiles

```bash
# Generate profiling data
go run ./cmd/perfprofile \
  -cpuprofile=cpu.prof \
  -memprofile=mem.prof \
  -duration=60s \
  -entities=2000

# Analyze CPU profile
go tool pprof cpu.prof
(pprof) top20          # Show top 20 CPU consumers
(pprof) list FuncName  # Show annotated source
(pprof) web            # Generate call graph (requires graphviz)

# Analyze memory profile
go tool pprof mem.prof
(pprof) top20 -cum     # Show top allocators
(pprof) list FuncName  # Show allocation sites
```

## Performance Targets

The profiling tool compares measurements against these targets:

| Metric | Target | Rationale |
|--------|--------|-----------|
| World Update | < 16.67 ms | 60 FPS frame budget |
| Entity Query | < 100 μs | Hot path efficiency |
| Component Access | < 5 ns | Cache-optimized access |
| Terrain Generation | < 2000 ms | Level load time |
| Entity Generation | < 1 ms | Per-entity spawn time |
| GC Pause | < 5 ms | Avoid frame drops |

## Using the Profiling Tool

### Command-Line Options

```
-duration duration
    Duration to run profiling (default 10s)
    Example: -duration=30s, -duration=1m

-entities int
    Number of entities to simulate (default 1000)
    Example: -entities=2000, -entities=500

-cpuprofile string
    Write CPU profile to file
    Example: -cpuprofile=cpu.prof

-memprofile string
    Write memory profile to file
    Example: -memprofile=mem.prof

-output string
    Output file for profiling report (default docs/profiling/assessment_report.md)
    Example: -output=my_report.md

-v  Verbose output (default false)
    Shows progress during profiling
```

### Example Workflows

#### Scenario 1: Quick Performance Check

```bash
# Fast assessment to verify targets are met
go run ./cmd/perfprofile -duration=5s -entities=500

# Check report
cat docs/profiling/assessment_report.md
```

#### Scenario 2: Detailed Performance Analysis

```bash
# Full assessment with profiling data
go run ./cmd/perfprofile \
  -duration=60s \
  -entities=2000 \
  -cpuprofile=docs/profiling/cpu.prof \
  -memprofile=docs/profiling/mem.prof \
  -v

# Analyze hotspots
go tool pprof -top docs/profiling/cpu.prof

# Find allocation sites
go tool pprof -alloc_space docs/profiling/mem.prof
```

#### Scenario 3: Before/After Optimization Comparison

```bash
# Before optimization
go run ./cmd/perfprofile \
  -duration=30s \
  -output=before_optimization.md

# Make code changes...

# After optimization
go run ./cmd/perfprofile \
  -duration=30s \
  -output=after_optimization.md

# Compare reports manually or with diff
diff before_optimization.md after_optimization.md
```

## Understanding the Report

### ECS Performance Section

```markdown
| Operation | Time | Target | Status |
|-----------|------|--------|--------|
| World Update | 0.823 ms | <16.67 ms | ✅ |
| Entity Query | 12.3 μs | <100 μs | ✅ |
| Component Access | 0.4 ns | <5 ns | ✅ |
```

**Interpretation:**
- ✅ = Meets target (good performance)
- ⚠️ = Exceeds target (needs optimization)

**World Update**: Time to process one game loop iteration (60 per second at 60 FPS)
**Entity Query**: Time to find entities with specific components (GetEntitiesWith)
**Component Access**: Time to retrieve a component from an entity (GetPosition)

### Memory Analysis Section

```markdown
| Metric | Initial | Final | Change |
|--------|---------|-------|--------|
| Allocated | 5.23 MB | 12.45 MB | 7.22 MB |
| System | 15.67 MB | 18.92 MB | 3.25 MB |
| GC Count | 0 | 3 | 3 |
| Avg GC Pause | - | - | 1.23 ms |
```

**Key Metrics:**
- **Allocated**: Heap memory used by application
- **System**: Total memory requested from OS
- **GC Count**: Number of garbage collection cycles
- **Avg GC Pause**: Average stop-the-world GC duration

**Red Flags:**
- Memory growing >1 MB/s (may indicate leak)
- GC pause >5ms (causes frame drops)
- High GC frequency (>10 per second)

### Procedural Generation Section

```markdown
| Generator | Avg Time | Target | Status |
|-----------|----------|--------|--------|
| Terrain | 456.78 ms | <2000 ms | ✅ |
| Entity | 0.23 ms | <1 ms | ✅ |
| Item | 0.15 ms | <1 ms | ✅ |
| Spell | 0.18 ms | <1 ms | ✅ |
```

**Interpretation:**
- Times shown are per-generation averages
- Terrain generation runs once per level
- Entity/Item/Spell generation happens frequently during gameplay

### Recommendations Section

The tool automatically generates prioritized recommendations when metrics exceed targets:

```markdown
### Recommendations

1. **Optimize Entity Queries**
   - Entity queries taking 156.2 μs (target: <100μs). Verify query cache is working.

2. **Reduce GC Pauses**
   - Average GC pause: 7.23 ms (target: <5ms). Implement object pooling for hot paths.
```

## Advanced Profiling Techniques

### CPU Profiling with pprof

```bash
# Generate CPU profile
go run ./cmd/perfprofile -cpuprofile=cpu.prof -duration=60s

# Interactive analysis
go tool pprof cpu.prof
(pprof) top20          # Top 20 functions by CPU time
(pprof) top20 -cum     # Top 20 by cumulative time
(pprof) list GetEntitiesWith  # Source code with timing annotations
(pprof) web            # Visual call graph (requires graphviz)

# Generate flame graph (if installed)
go tool pprof -http=:8080 cpu.prof  # Opens web interface
```

### Memory Profiling

```bash
# Generate memory profile
go run ./cmd/perfprofile -memprofile=mem.prof -duration=60s

# Analyze allocations
go tool pprof mem.prof
(pprof) top20 -alloc_space    # Total allocations
(pprof) top20 -inuse_space    # Current heap usage
(pprof) list EntityGenerator  # Allocation sites in code

# Find memory leaks
go tool pprof -base=before.prof after.prof  # Compare snapshots
```

### Benchmark Comparison with benchstat

```bash
# Run benchmarks before optimization
go test -bench=. -benchmem ./pkg/engine > before.txt

# Make optimizations...

# Run benchmarks after optimization
go test -bench=. -benchmem ./pkg/engine > after.txt

# Statistical comparison
benchstat before.txt after.txt
```

## Integration with CI/CD

### GitHub Actions

**Note**: The perfprofile tool requires a display (X11) and won't run in standard CI environments. Use `go test -bench` instead:

```yaml
name: Performance Regression Tests

on: [push, pull_request]

jobs:
  benchmark:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.24'
      
      - name: Run benchmarks
        run: |
          go test -bench=. -benchmem ./pkg/engine > new.txt
          
      - name: Compare with baseline
        run: benchstat baseline.txt new.txt
```

### Local Pre-Commit Hook

Create `.git/hooks/pre-push` for automated profiling:

```bash
#!/bin/bash
echo "Running performance check..."
go run ./cmd/perfprofile -duration=10s -entities=500 -output=/tmp/perf_report.md

# Check if all targets met
if grep -q "⚠️" /tmp/perf_report.md; then
    echo "⚠️  Performance regression detected!"
    cat /tmp/perf_report.md
    read -p "Continue push? (y/n) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi
```

## Profiling Best Practices

### DO

✅ **Profile before optimizing** - Measure first, optimize second  
✅ **Use realistic workloads** - Profile with 1000-2000 entities  
✅ **Run multiple samples** - Take 3-5 measurements and average  
✅ **Compare before/after** - Validate optimization effectiveness  
✅ **Profile in release mode** - Use `-ldflags="-s -w"` for realistic timings  
✅ **Test on target hardware** - Profile on low-end machines too

### DON'T

❌ **Don't optimize without profiling** - Premature optimization is wasteful  
❌ **Don't profile debug builds** - Debug builds are 10x slower  
❌ **Don't trust single samples** - Variance can be misleading  
❌ **Don't ignore allocations** - GC pauses cause frame drops  
❌ **Don't micro-optimize cold paths** - Focus on hot paths (>1% CPU time)  
❌ **Don't break determinism** - Maintain seed-based generation

## Troubleshooting

### Tool Won't Run (X11 Error)

**Problem**: `glfw: X11: The DISPLAY environment variable is missing`

**Cause**: Ebiten requires a display, even for non-graphical code

**Solutions**:
1. Run on machine with display (local development)
2. Use `go test -bench` for CI environments
3. Set up Xvfb virtual framebuffer:
   ```bash
   Xvfb :99 -screen 0 1024x768x24 &
   export DISPLAY=:99
   go run ./cmd/perfprofile
   ```

### Inconsistent Results

**Problem**: Performance varies between runs

**Solutions**:
1. Increase `-duration` for more stable averages
2. Close background applications
3. Disable CPU frequency scaling: `sudo cpupower frequency-set -g performance`
4. Run multiple times and use median/average

### Memory Profiling Shows No Allocations

**Problem**: Memory profile is empty or shows minimal data

**Cause**: Short profiling duration or good memory efficiency

**Solutions**:
1. Increase `-duration` to capture more samples
2. Increase `-entities` for more allocation pressure
3. Force GC before profile: `runtime.GC()`

## Related Documentation

- [PLAN.md](../../PLAN.md) - Complete optimization plan with priorities
- [docs/PERFORMANCE.md](../PERFORMANCE.md) - Performance optimization guide
- [docs/profiling/optimization_progress.md](optimization_progress.md) - Completed optimizations
- [docs/TESTING.md](../TESTING.md) - Testing and benchmarking guide

## References

- [Go pprof documentation](https://pkg.go.dev/runtime/pprof)
- [Profiling Go Programs](https://go.dev/blog/pprof)
- [Go Performance Optimization Guide](https://github.com/dgryski/go-perfbook)
- [Benchstat tool](https://pkg.go.dev/golang.org/x/perf/cmd/benchstat)
