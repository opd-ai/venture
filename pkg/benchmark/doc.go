// Package benchmark provides performance validation infrastructure for the Venture game engine.
//
// This package contains two sub-packages that validate critical performance targets
// from docs/PERFORMANCE.md:
//
// Sub-packages:
//
//   - fps: Frame rate benchmarks validating the 60 FPS target with various entity counts
//     (500, 2000, 5000 entities). Includes realistic system load simulation with
//     AI pathfinding, combat processing, and physics updates.
//
//   - memory: Memory usage tests validating the <500MB client memory target across
//     4 scenarios: baseline world, high entity count (2000 entities), procgen stress,
//     and rendering stress. Uses pkg/memprofile for allocation tracking.
//
// CI/CD Integration:
//
// Both sub-packages are integrated with GitHub Actions:
//   - .github/workflows/benchmark.yml runs fps benchmarks with 3s runtime
//   - scripts/benchmark-memory.sh executes memory tests and generates reports
//
// The separation between fps and memory tests enables cleaner CI reporting and
// focused performance regression detection.
//
// Usage:
//
// Run all benchmarks:
//
//	go test -bench=. ./pkg/benchmark/...
//
// Run only FPS benchmarks:
//
//	go test -bench=. ./pkg/benchmark/fps/
//
// Run only memory tests:
//
//	go test -v ./pkg/benchmark/memory/
//
// Performance Targets:
//
//   - FPS: 60 FPS minimum (16.67ms per frame / 16,666,666 ns/op)
//   - Memory: <500MB client memory under typical load
package benchmark
