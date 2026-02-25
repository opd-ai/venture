// Package memory provides memory benchmark tests for validating the <500MB client memory usage target.
// These tests simulate realistic game scenarios without requiring graphics initialization,
// making them suitable for CI/CD environments without display servers.
//
// Usage:
//
//	go test ./pkg/benchmark/memory/
//	go test -v ./pkg/benchmark/memory/ -run=TestMemoryHighEntityCount
//
// # Custom Memory Profiling
//
// To create custom memory benchmarks using the memprofile package:
//
//	import "github.com/opd-ai/venture/pkg/memprofile"
//
//	func TestMyCustomMemoryBenchmark(t *testing.T) {
//	    profile := memprofile.StartMemoryProfile("MyFeature")
//
//	    // Allocate your data structures
//	    data := make([]byte, 1024*1024) // 1MB allocation
//	    profile.Snapshot()
//
//	    // Perform operations that may allocate memory
//	    moreData := make([]byte, 2*1024*1024) // 2MB more
//	    profile.Snapshot()
//
//	    profile.End()
//
//	    // Check results
//	    peak := profile.GetPeakAllocation()
//	    peakMB := float64(peak) / (1024 * 1024)
//	    t.Logf("Peak allocation: %.2fMB", peakMB)
//
//	    // Assert against thresholds
//	    if peak > uint64(500*1024*1024) {
//	        t.Errorf("Memory exceeded 500MB threshold: %.2fMB", peakMB)
//	    }
//	}
//
// The memprofile package tracks heap allocations between snapshots and can identify
// memory growth patterns and peak usage, making it ideal for validating memory-intensive
// game subsystems like procedural generation, sprite caching, and world loading.
package memory
