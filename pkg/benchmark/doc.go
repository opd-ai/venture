// Package benchmark provides performance validation infrastructure for the Venture game engine.
//
// This is a pure test infrastructure package containing only test files and documentation.
// No runtime production code exists in this package - all functionality is in test files
// organized into subdirectories.
//
// This package contains two sub-packages that validate critical performance targets
// from docs/PERFORMANCE.md:
//
// Sub-packages:
//
//   - fps: Frame rate benchmarks validating the 60 FPS target with various entity counts
//     (500, 2000, 5000 entities). Includes realistic system load simulation with
//     AI pathfinding, combat processing, and physics updates.
//     See pkg/benchmark/fps/README.md for detailed benchmark documentation.
//
//   - memory: Memory usage tests validating the <500MB client memory target across
//     4 scenarios: baseline world, high entity count (2000 entities), procgen stress,
//     and rendering stress. Uses pkg/memprofile for allocation tracking.
//     See pkg/benchmark/memory/README.md for detailed test documentation.
//
// CI/CD Integration:
//
// Both sub-packages are integrated with CI/CD via shell scripts:
//   - scripts/benchmark-memory.sh executes memory tests and generates reports
//   - scripts/benchmark-regression.sh detects performance regressions
//   - scripts/benchmark-baseline.json stores historical benchmark data
//
// For workflow documentation, see docs/BENCHMARK_WORKFLOW.md.
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
