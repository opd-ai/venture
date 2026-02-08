// Package fps provides dedicated FPS (frames per second) benchmarks for the Venture game engine.
//
// This package validates the 60 FPS performance target specified in docs/PERFORMANCE.md
// with various entity counts and system loads.
//
// Performance Target: 60 FPS = 16.67ms per frame = 16,666,666 ns/op
//
// The benchmarks in this package are separated from memory benchmarks to enable cleaner
// CI reporting and focused performance regression detection.
//
// Usage:
//
// go test -bench=. ./pkg/benchmark/fps/
// go test -bench=BenchmarkFPS2000Entities -benchtime=10s ./pkg/benchmark/fps/
package fps
