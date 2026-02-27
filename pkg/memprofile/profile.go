package memprofile

import (
	"fmt"
	"runtime"
	"time"

	log "github.com/sirupsen/logrus"
)

// MemorySnapshot captures memory statistics at a point in time.
type MemorySnapshot struct {
	Timestamp    time.Time `json:"timestamp"`
	Alloc        uint64    `json:"alloc"`          // Bytes allocated and still in use
	TotalAlloc   uint64    `json:"total_alloc"`    // Bytes allocated (cumulative)
	Sys          uint64    `json:"sys"`            // Bytes obtained from system
	Mallocs      uint64    `json:"mallocs"`        // Number of mallocs
	Frees        uint64    `json:"frees"`          // Number of frees
	LiveObjects  uint64    `json:"live_objects"`   // Number of live objects (Mallocs - Frees)
	NumGC        uint32    `json:"num_gc"`         // Number of GC runs
	PauseTotalNs uint64    `json:"pause_total_ns"` // Total GC pause time
}

// MemoryProfile tracks memory usage over time.
type MemoryProfile struct {
	Name         string           `json:"name"`
	Snapshots    []MemorySnapshot `json:"snapshots"`
	StartTime    time.Time        `json:"start_time"`
	EndTime      time.Time        `json:"end_time"`
	LeakDetected bool             `json:"leak_detected"`
	LeakRate     float64          `json:"leak_rate"` // Bytes per second
}

// CaptureMemorySnapshot captures current memory statistics.
func CaptureMemorySnapshot() MemorySnapshot {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return MemorySnapshot{
		Timestamp:    time.Now(),
		Alloc:        m.Alloc,
		TotalAlloc:   m.TotalAlloc,
		Sys:          m.Sys,
		Mallocs:      m.Mallocs,
		Frees:        m.Frees,
		LiveObjects:  m.Mallocs - m.Frees,
		NumGC:        m.NumGC,
		PauseTotalNs: m.PauseTotalNs,
	}
}

// StartMemoryProfile begins tracking memory usage.
func StartMemoryProfile(name string) *MemoryProfile {
	runtime.GC() // Collect garbage before starting

	profile := &MemoryProfile{
		Name:      name,
		Snapshots: make([]MemorySnapshot, 0, 10),
		StartTime: time.Now(),
	}

	// Take initial snapshot
	profile.Snapshots = append(profile.Snapshots, CaptureMemorySnapshot())

	log.WithFields(log.Fields{
		"profile_name":    name,
		"initial_alloc":   profile.Snapshots[0].Alloc,
		"initial_objects": profile.Snapshots[0].LiveObjects,
	}).Debug("memory profile started")

	return profile
}

// Snapshot adds a memory snapshot to the profile.
func (p *MemoryProfile) Snapshot() {
	p.Snapshots = append(p.Snapshots, CaptureMemorySnapshot())
}

// End finalizes the memory profile and performs leak detection.
func (p *MemoryProfile) End() {
	runtime.GC()                      // Collect garbage before final snapshot
	time.Sleep(10 * time.Millisecond) // Allow GC to complete

	p.EndTime = time.Now()
	p.Snapshots = append(p.Snapshots, CaptureMemorySnapshot())

	// Detect memory leaks
	p.detectLeaks()

	finalSnapshot := p.Snapshots[len(p.Snapshots)-1]
	log.WithFields(log.Fields{
		"profile_name":  p.Name,
		"duration":      p.EndTime.Sub(p.StartTime),
		"final_alloc":   finalSnapshot.Alloc,
		"final_objects": finalSnapshot.LiveObjects,
		"leak_detected": p.LeakDetected,
		"leak_rate":     p.LeakRate,
	}).Debug("memory profile ended")
}

// detectLeaks analyzes snapshots to identify memory leaks.
func (p *MemoryProfile) detectLeaks() {
	if len(p.Snapshots) < 2 {
		return
	}

	first := p.Snapshots[0]
	last := p.Snapshots[len(p.Snapshots)-1]

	// Calculate allocation growth
	allocGrowth := int64(last.Alloc) - int64(first.Alloc)
	objectGrowth := int64(last.LiveObjects) - int64(first.LiveObjects)
	duration := last.Timestamp.Sub(first.Timestamp)

	// If allocation and object count both grew significantly, suspect leak
	// Allow 10% growth as normal variation
	// Guard against division by zero when first.Alloc or first.LiveObjects is 0
	var allocGrowthPercent, objectGrowthPercent float64
	if first.Alloc > 0 {
		allocGrowthPercent = float64(allocGrowth) / float64(first.Alloc) * 100.0
	} else if allocGrowth > 0 {
		// Started at zero allocation, any growth from zero is significant
		allocGrowthPercent = 100.0
	}

	if first.LiveObjects > 0 {
		objectGrowthPercent = float64(objectGrowth) / float64(first.LiveObjects) * 100.0
	} else if objectGrowth > 0 {
		// Started with zero objects, any growth from zero is significant
		objectGrowthPercent = 100.0
	}

	// Consider it a leak if:
	// 1. Both alloc and objects grew by >10%
	// 2. Growth is sustained (not just a spike)
	if allocGrowthPercent > 10.0 && objectGrowthPercent > 10.0 {
		p.LeakDetected = true
		if duration.Seconds() > 0 {
			p.LeakRate = float64(allocGrowth) / duration.Seconds()
		}

		log.WithFields(log.Fields{
			"profile_name":          p.Name,
			"alloc_growth_percent":  allocGrowthPercent,
			"object_growth_percent": objectGrowthPercent,
			"alloc_growth_bytes":    allocGrowth,
			"object_growth_count":   objectGrowth,
			"leak_rate_bytes_sec":   p.LeakRate,
		}).Warn("memory leak detected")
	}
}

// GetPeakAllocation returns the maximum allocation seen.
func (p *MemoryProfile) GetPeakAllocation() uint64 {
	var peak uint64
	for _, snapshot := range p.Snapshots {
		if snapshot.Alloc > peak {
			peak = snapshot.Alloc
		}
	}
	return peak
}

// GetAverageAllocation returns the average allocation across snapshots.
func (p *MemoryProfile) GetAverageAllocation() uint64 {
	if len(p.Snapshots) == 0 {
		return 0
	}

	var total uint64
	for _, snapshot := range p.Snapshots {
		total += snapshot.Alloc
	}
	return total / uint64(len(p.Snapshots))
}

// GetAllocationGrowth returns bytes allocated between first and last snapshot.
func (p *MemoryProfile) GetAllocationGrowth() int64 {
	if len(p.Snapshots) < 2 {
		return 0
	}
	return int64(p.Snapshots[len(p.Snapshots)-1].Alloc) - int64(p.Snapshots[0].Alloc)
}

// GetObjectGrowth returns object count change between first and last snapshot.
func (p *MemoryProfile) GetObjectGrowth() int64 {
	if len(p.Snapshots) < 2 {
		return 0
	}
	return int64(p.Snapshots[len(p.Snapshots)-1].LiveObjects) - int64(p.Snapshots[0].LiveObjects)
}

// PrintProfile prints a formatted memory profile report to stdout.
// NOTE: This function intentionally uses fmt.Printf for CLI/debug output.
// It is exempt from the structured logging guideline (Coding Guideline #3)
// as it's designed for human-readable console output in testing and debugging tools.
func (p *MemoryProfile) PrintProfile() {
	fmt.Printf("\n=== Memory Profile: %s ===\n", p.Name)
	fmt.Printf("Duration: %v\n", p.EndTime.Sub(p.StartTime))
	fmt.Printf("Snapshots: %d\n\n", len(p.Snapshots))

	if len(p.Snapshots) == 0 {
		fmt.Println("No snapshots captured")
		return
	}

	// Print summary statistics
	fmt.Printf("Peak Allocation:    %s\n", formatBytes(p.GetPeakAllocation()))
	fmt.Printf("Average Allocation: %s\n", formatBytes(p.GetAverageAllocation()))
	fmt.Printf("Allocation Growth:  %s\n", formatBytesWithSign(p.GetAllocationGrowth()))
	fmt.Printf("Object Growth:      %+d objects\n", p.GetObjectGrowth())

	if p.LeakDetected {
		fmt.Printf("\n⚠ LEAK DETECTED: %.2f bytes/sec\n", p.LeakRate)
	} else {
		fmt.Printf("\n✓ No leaks detected\n")
	}

	// Print snapshot table
	fmt.Println("\nSnapshots:")
	fmt.Printf("%-20s %15s %15s %15s\n", "Time", "Alloc", "Live Objects", "GC Runs")
	for i, snapshot := range p.Snapshots {
		timestamp := "Initial"
		if i > 0 {
			duration := snapshot.Timestamp.Sub(p.Snapshots[0].Timestamp)
			timestamp = fmt.Sprintf("+%v", duration)
		}
		fmt.Printf("%-20s %15s %15d %15d\n",
			timestamp,
			formatBytes(snapshot.Alloc),
			snapshot.LiveObjects,
			snapshot.NumGC,
		)
	}
}

// formatBytes formats bytes into a human-readable string.
func formatBytes(bytes uint64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

// formatBytesWithSign formats bytes with +/- sign for growth.
func formatBytesWithSign(bytes int64) string {
	sign := ""
	absBytes := uint64(bytes)
	if bytes > 0 {
		sign = "+"
	} else if bytes < 0 {
		sign = "-"
		absBytes = uint64(-bytes)
	}

	return sign + formatBytes(absBytes)
}

// ProfileFunction profiles memory usage of a function.
func ProfileFunction(name string, iterations int, fn func()) *MemoryProfile {
	profile := StartMemoryProfile(name)

	// Take snapshots during execution
	snapshotInterval := iterations / 5
	if snapshotInterval < 1 {
		snapshotInterval = 1
	}

	for i := 0; i < iterations; i++ {
		fn()

		if i%snapshotInterval == 0 {
			profile.Snapshot()
		}
	}

	profile.End()
	return profile
}

// DetectLeaksInBenchmark runs a function multiple times and checks for leaks.
func DetectLeaksInBenchmark(name string, warmupRuns, testRuns int, fn func()) bool {
	// Warmup to stabilize allocations
	for i := 0; i < warmupRuns; i++ {
		fn()
	}

	// Profile the actual test
	profile := ProfileFunction(name, testRuns, fn)
	return profile.LeakDetected
}

// MemoryTest represents a memory test configuration.
type MemoryTest struct {
	Name       string
	WarmupRuns int
	TestRuns   int
	MaxGrowth  int64 // Maximum allowed allocation growth in bytes
}

// RunMemoryTest executes a memory test and returns whether it passed.
func RunMemoryTest(test MemoryTest, fn func()) (bool, *MemoryProfile) {
	profile := ProfileFunction(test.Name, test.TestRuns, fn)

	growth := profile.GetAllocationGrowth()
	passed := !profile.LeakDetected && (test.MaxGrowth == 0 || growth <= test.MaxGrowth)

	return passed, profile
}
