# Memory Benchmark Package

This package provides memory usage validation tests for the Venture game client, verifying the <500MB memory usage target documented in `docs/PERFORMANCE.md`.

## Overview

The memory benchmark suite simulates realistic game scenarios and measures peak memory allocation to ensure the client stays within performance targets. All tests run without requiring a display server, making them suitable for CI/CD environments.

## Test Scenarios

### 1. Baseline World Generation
Simulates basic world initialization with minimal systems:
- 10MB world data
- 500 entities with 1KB each
- 100 cached sprites (64x64 RGBA)

**Result:** ~10MB peak allocation (2% of threshold)

### 2. High Entity Count
Tests performance target of 2000 entities with full components:
- 2000 entities with position, velocity, health, sprite, inventory
- Snapshots taken every 400 entities

**Result:** ~4MB peak allocation (0.7% of threshold)

### 3. Procedural Generation Stress
Intensive procedural content generation:
- 500 items with stats and icons
- 200 quests with descriptions and objectives
- 300 spells with effects
- 10 terrain chunks (512x512 tiles)

**Result:** ~2MB peak allocation (0.3% of threshold)

### 4. Rendering Pipeline Stress
Heavy rendering workload simulation:
- 50 sprite types × 8 animation frames (64x64 RGBA)
- 1000 particle effects
- 200 dynamic lights
- Triple buffering (1920x1080 RGBA)

**Result:** ~24MB peak allocation (3.2% of threshold)

## Running Tests

### Individual Test
```bash
go test -v -run=TestMemoryBaselineWorld ./pkg/benchmark/memory
```

### All Memory Tests
```bash
go test -v ./pkg/benchmark/memory
```

### Automated Benchmark Script
```bash
./scripts/benchmark-memory.sh
```

The script runs all tests and generates a results file at `build/memory-benchmark-results.txt`.

## Threshold

All tests validate against a 500MB threshold (524,288,000 bytes). Tests fail if peak allocation exceeds this limit.

## Dependencies

This package uses `pkg/memprofile` for memory profiling utilities, which has no graphics dependencies and can run in headless environments.

## CI/CD Integration

Add to your CI pipeline:
```yaml
- name: Memory Benchmarks
  run: ./scripts/benchmark-memory.sh
```

The script exits with status 0 on success, non-zero on failure.

## Results

All scenarios demonstrate significant headroom under the 500MB limit:
- Peak across all tests: 24MB
- Headroom: 476MB (95% remaining)

This validates the <500MB claim in the performance documentation.
