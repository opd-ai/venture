# FPS Benchmark Suite

This package provides dedicated FPS (frames per second) benchmarks for the Venture game engine, validating the **60 FPS with 2000 entities** performance target specified in `docs/PERFORMANCE.md`.

## Performance Target

**60 FPS = 16.67ms per frame = 16,666,666 ns/op**

## Benchmarks

### Core Benchmarks

| Benchmark | Entity Count | Purpose |
|-----------|--------------|---------|
| `BenchmarkFPS2000Entities` | 2000 | Primary 60 FPS validation (baseline components) |
| `BenchmarkFPS2000EntitiesRealistic` | 2000 | Real-world scenario with varied components |
| `BenchmarkFPS500Entities` | 500 | Validate performance headroom |
| `BenchmarkFPS5000Entities` | 5000 | Stress test for degradation patterns |

### Unit Tests

| Test | Purpose |
|------|---------|
| `TestFPS60TargetWith2000Entities` | CI/CD pass/fail validation of 60 FPS target |
| `TestFPS60TargetLightWeight` | Validate 60 FPS with light loads (500 entities) |

## Usage

### Run All Benchmarks
```bash
go test -bench=. ./pkg/benchmark/fps/
```

### Run Specific Benchmark
```bash
go test -bench=BenchmarkFPS2000Entities ./pkg/benchmark/fps/
```

### Extended Benchmark Run
```bash
go test -bench=BenchmarkFPS2000Entities -benchtime=10s ./pkg/benchmark/fps/
```

### Run Tests Only
```bash
go test ./pkg/benchmark/fps/
```

### Memory Profiling
```bash
go test -bench=BenchmarkFPS2000Entities -benchmem ./pkg/benchmark/fps/
```

## Sample Output

```
BenchmarkFPS2000Entities-8              50000    16234 ns/op     0 B/op    0 allocs/op
BenchmarkFPS2000EntitiesRealistic-8     40000    18123 ns/op     0 B/op    0 allocs/op
BenchmarkFPS500Entities-8              200000     4056 ns/op     0 B/op    0 allocs/op
BenchmarkFPS5000Entities-8              20000    40567 ns/op     0 B/op    0 allocs/op
```

## CI Integration

These benchmarks are designed for continuous integration:

1. **Unit Tests** provide pass/fail thresholds for automated validation
2. **Benchmarks** track performance regression over time
3. **Separate from Memory Tests** for cleaner reporting and focused metrics

## Relationship to Other Tests

- **pkg/benchmark/memory/**: Memory usage validation (<500MB target)
- **pkg/engine/performance_targets_test.go**: Legacy combined FPS+memory tests (still maintained for backwards compatibility)
- **This package**: Dedicated FPS-only benchmarks for cleaner CI reporting

## Design Notes

- Uses `MovementSystem` only to isolate core ECS overhead
- Collision detection excluded (depends on spatial partitioning configuration)
- Entity component variety matches real-world usage patterns
- Delta time fixed at 0.016 (60 FPS) for consistent measurements
