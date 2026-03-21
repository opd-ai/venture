# Performance Tuning Guide

This guide helps optimize Venture's performance across different hardware configurations. Venture targets 60 FPS on mid-range hardware with <500MB memory usage.

## Minimum Hardware Requirements

| Component | Minimum | Recommended |
|-----------|---------|-------------|
| CPU | Dual-core 2.0 GHz | Quad-core 3.0 GHz |
| RAM | 4 GB | 8 GB |
| GPU | OpenGL 3.0 / WebGL 2.0 | OpenGL 4.0 |
| Storage | 100 MB (single binary) | 100 MB |

## Performance Targets

- **Frame Rate**: 60 FPS minimum
- **Frame Budget**: 16.6ms per frame
- **Memory**: <500MB client, <1GB server (8 players)
- **Entity Count**: 2000+ entities at 60 FPS

## Command-Line Flags for Performance

### Display Settings

```bash
# Resolution (lower = better performance)
-width=1920 -height=1080   # Full HD (default)
-width=1280 -height=720    # HD (for weaker hardware)
-width=2560 -height=1440   # QHD (requires powerful GPU)
-width=3840 -height=2160   # 4K (requires high-end GPU)

# Fullscreen mode (may improve performance)
-fullscreen=true
```

### Visual Effects (Disable for Performance)

```bash
# Disable post-processing for maximum FPS
-postprocess-color-grading=false
-postprocess-vignette=false
-postprocess-chromatic=false

# Reduce post-processing intensity
-postprocess-saturation=1.0
-postprocess-contrast=1.0
-postprocess-vignette-intensity=0.3
-postprocess-chromatic-intensity=0.1
```

### Memory Management

```bash
# Sprite cache size (MB)
# Default: 400 desktop, 150 WASM
# Reduce for low-memory systems, increase for fast load times
-sprite-cache-mb=200   # Moderate memory usage
-sprite-cache-mb=100   # Low memory usage
```

### Network Performance (High-Latency)

```bash
# For Tor/onion services or satellite internet (200-5000ms latency)
-high-latency=true

# Server configuration
-tick-rate=30          # Updates per second (lower = less bandwidth)
-max-players=4         # Player cap affects server memory
```

### Profiling

```bash
# Enable frame time tracking
-profile=true

# Verbose logging (disable for production)
-verbose=false
```

## Performance Profiles

### Low-End Hardware Profile

```bash
./venture-client \
  -width=1280 -height=720 \
  -postprocess-color-grading=false \
  -postprocess-vignette=false \
  -postprocess-chromatic=false \
  -sprite-cache-mb=100 \
  -verbose=false
```

### High-End Hardware Profile

```bash
./venture-client \
  -width=2560 -height=1440 \
  -fullscreen=true \
  -postprocess-preset=cinematic \
  -sprite-cache-mb=400
```

### Low-Bandwidth Network Profile

```bash
./venture-client \
  -multiplayer=true \
  -server=example.onion:8080 \
  -high-latency=true \
  -tick-rate=20
```

## Interpreting Profiling Output

When `-profile=true` is enabled, the game logs frame timing information:

```
Frame: 1234, DeltaTime: 16.2ms, FPS: 61.7
  Update: 4.1ms (25%)
  Render: 10.8ms (67%)
  Other: 1.3ms (8%)
```

### What to Look For

| Metric | Healthy | Warning | Critical |
|--------|---------|---------|----------|
| DeltaTime | <16.6ms | 16.6-33.3ms | >33.3ms |
| FPS | >60 | 30-60 | <30 |
| Update | <6ms | 6-10ms | >10ms |
| Render | <10ms | 10-14ms | >14ms |

### Common Bottlenecks

**High Update Time (>6ms)**
- Too many AI entities active
- Complex collision detection
- Network synchronization overhead

Solutions:
- Reduce entity count in world generation
- Increase spatial partition cell size
- Use `-high-latency=true` for network issues

**High Render Time (>10ms)**
- Too many particles active
- High-resolution sprites being generated
- Post-processing effects

Solutions:
- Disable post-processing effects
- Lower resolution
- Reduce particle density in weather effects

## WebAssembly (Browser) Performance

WASM builds have additional constraints:

- Default sprite cache: 150MB (vs 400MB desktop)
- Single-threaded JavaScript environment
- Browser memory limits apply

### Browser Optimization Tips

1. Use Chrome/Edge for best WebAssembly performance
2. Close other tabs to free memory
3. Disable browser extensions during play
4. Use `-width=1280 -height=720` for smooth performance

## VR Mode Performance

VR mode requires maintaining 90+ FPS for comfort:

```bash
# VR with performance optimizations
./venture-client \
  -vr=true \
  -postprocess-color-grading=false \
  -postprocess-chromatic=false \
  -sprite-cache-mb=300
```

Note: VR mode is experimental. Use `-force-vr=true` for testing without hardware.

## Server Performance

### Dedicated Server Configuration

```bash
./venture-server \
  -port=8080 \
  -max-players=8 \
  -tick-rate=30
```

### Memory Usage Per Player

| Players | Estimated Memory |
|---------|-----------------|
| 1-4 | ~200MB |
| 5-8 | ~500MB |
| 9-16 | ~800MB |

### Tick Rate Guidelines

| Tick Rate | Latency | Bandwidth | Use Case |
|-----------|---------|-----------|----------|
| 60 | <50ms | High | LAN gaming |
| 30 | <100ms | Medium | Internet gaming |
| 20 | <200ms | Low | High-latency networks |
| 10 | <500ms | Minimal | Tor/satellite |

## Running Benchmarks

### FPS Benchmark

```bash
# Run FPS benchmark
xvfb-run go test -bench=BenchmarkFPS2000Entities ./pkg/benchmark/fps/

# Expected output:
# BenchmarkFPS2000Entities-8    7832    152331 ns/op    0 B/op    0 allocs/op
# Target: <200000 ns/op (5000+ FPS headroom)
```

### Memory Benchmark

```bash
# Run memory benchmark
xvfb-run go test -v -run=TestMemory ./pkg/benchmark/memory/

# Expected output:
# Peak allocation: 16.07MB
# Target: <500MB
```

### Full Regression Test

```bash
# Run all performance benchmarks
./scripts/benchmark-regression.sh
./scripts/benchmark-memory.sh
```

## Troubleshooting

### Game Stuttering

1. Check if frame time spikes correlate with garbage collection
2. Increase sprite cache size
3. Disable verbose logging

### High Memory Usage

1. Reduce sprite cache size
2. Lower resolution
3. Check for memory leaks with `-profile=true`

### Network Lag

1. Enable `-high-latency=true`
2. Lower tick rate
3. Check server region latency

### Black Screen or Crash

1. Update graphics drivers
2. Try lower resolution
3. Check OpenGL version support

## Platform-Specific Notes

### Linux

- Ensure OpenGL drivers are installed: `sudo apt install mesa-utils`
- Check OpenGL version: `glxinfo | grep "OpenGL version"`

### macOS

- Metal backend used by default (faster than OpenGL)
- Ensure System Preferences → Security allows the app

### Windows

- Update DirectX/OpenGL drivers
- Disable Windows Game Mode if experiencing stuttering

### Mobile (iOS/Android)

- Use lower resolution for battery life
- Reduce sprite cache to 100MB
- Disable post-processing for thermal management

## Monitoring Tools

```bash
# Real-time CPU/memory monitoring
top -p $(pgrep venture-client)

# GPU monitoring (NVIDIA)
nvidia-smi -l 1

# Memory profiling
go tool pprof http://localhost:6060/debug/pprof/heap
```

## Getting Help

If you experience performance issues not covered here:

1. Check [GitHub Issues](https://github.com/opd-ai/venture/issues) for known problems
2. Open a new issue with:
   - Hardware specifications
   - Operating system and version
   - Command-line flags used
   - Profiling output (`-profile=true -verbose=true`)
