// Package main provides a headless performance profiling tool for the Venture game engine.
// This tool focuses on non-graphical systems: ECS, entity queries, procedural generation.
package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/opd-ai/venture/pkg/procgen/entity"
	"github.com/opd-ai/venture/pkg/procgen/item"
	"github.com/opd-ai/venture/pkg/procgen/magic"
	"github.com/opd-ai/venture/pkg/procgen/terrain"
	"github.com/sirupsen/logrus"
)

var (
	cpuProfile   = flag.String("cpuprofile", "", "Write CPU profile to file")
	memProfile   = flag.String("memprofile", "", "Write memory profile to file")
	duration     = flag.Duration("duration", 10*time.Second, "Duration to run profiling")
	entityCount  = flag.Int("entities", 1000, "Number of entities to simulate")
	reportOutput = flag.String("output", "docs/profiling/assessment_report.md", "Output file for profiling report")
	verbose      = flag.Bool("v", false, "Verbose output")
)

func main() {
	flag.Parse()

	// Configure logging
	logger := logrus.New()
	if *verbose {
		logger.SetLevel(logrus.DebugLevel)
	} else {
		logger.SetLevel(logrus.InfoLevel)
	}

	logger.Info("Starting headless performance profiling")
	logger.Infof("Configuration: entities=%d, duration=%v", *entityCount, *duration)

	// Start CPU profiling if requested
	if *cpuProfile != "" {
		f, err := os.Create(*cpuProfile)
		if err != nil {
			logger.Fatalf("Failed to create CPU profile: %v", err)
		}
		defer f.Close()

		if err := pprof.StartCPUProfile(f); err != nil {
			logger.Fatalf("Failed to start CPU profiling: %v", err)
		}
		defer pprof.StopCPUProfile()
		logger.Infof("CPU profiling enabled: %s", *cpuProfile)
	}

	// Run performance tests
	report := runProfileTests(logger)

	// Write memory profile if requested
	if *memProfile != "" {
		f, err := os.Create(*memProfile)
		if err != nil {
			logger.Fatalf("Failed to create memory profile: %v", err)
		}
		defer f.Close()

		runtime.GC() // Force GC before memory profile
		if err := pprof.WriteHeapProfile(f); err != nil {
			logger.Fatalf("Failed to write memory profile: %v", err)
		}
		logger.Infof("Memory profiling complete: %s", *memProfile)
	}

	// Write report
	if err := writeReport(report, *reportOutput); err != nil {
		logger.Fatalf("Failed to write report: %v", err)
	}

	logger.Infof("Profiling complete. Report written to: %s", *reportOutput)
}

// ProfileReport contains comprehensive profiling data.
type ProfileReport struct {
	Duration    time.Duration
	EntityCount int

	// ECS Performance
	WorldUpdateTime     time.Duration
	EntityQueryTime     time.Duration
	ComponentAccessTime time.Duration

	// Memory statistics
	InitialMemory runtime.MemStats
	FinalMemory   runtime.MemStats
	MemoryGrowth  uint64
	GCCount       uint32
	GCPauseTotal  time.Duration

	// Generation performance
	TerrainGenTime time.Duration
	EntityGenTime  time.Duration
	ItemGenTime    time.Duration
	SpellGenTime   time.Duration

	// Query cache performance
	QueryCacheHits   int
	QueryCacheMisses int
}

func runProfileTests(logger *logrus.Logger) *ProfileReport {
	report := &ProfileReport{
		Duration:    *duration,
		EntityCount: *entityCount,
	}

	// Record initial memory state
	runtime.ReadMemStats(&report.InitialMemory)

	// Test 1: ECS World Performance
	logger.Info("Testing ECS world performance...")
	report.WorldUpdateTime, report.EntityQueryTime, report.ComponentAccessTime = benchmarkECS(logger)

	// Test 2: Procedural Generation
	logger.Info("Testing procedural generation...")
	report.TerrainGenTime = benchmarkTerrain()
	report.EntityGenTime = benchmarkEntity()
	report.ItemGenTime = benchmarkItem()
	report.SpellGenTime = benchmarkSpell()

	// Record final memory state
	runtime.ReadMemStats(&report.FinalMemory)
	report.MemoryGrowth = report.FinalMemory.Alloc - report.InitialMemory.Alloc
	report.GCCount = report.FinalMemory.NumGC - report.InitialMemory.NumGC
	report.GCPauseTotal = time.Duration(report.FinalMemory.PauseTotalNs - report.InitialMemory.PauseTotalNs)

	logger.Info("Profiling tests complete")
	return report
}

func benchmarkECS(logger *logrus.Logger) (worldUpdate, entityQuery, componentAccess time.Duration) {
	world := engine.NewWorld()

	// Create entities
	for i := 0; i < *entityCount; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&engine.PositionComponent{X: float64(i), Y: float64(i)})
		if i%2 == 0 {
			entity.AddComponent(&engine.VelocityComponent{VX: 1.0, VY: 1.0})
		}
		if i%3 == 0 {
			entity.AddComponent(&engine.HealthComponent{Current: 100, Max: 100})
		}
	}

	// Process pending entity additions (no systems needed for this test)
	world.Update(0)

	// Benchmark world update (without systems - just entity management)
	start := time.Now()
	iterations := int((*duration).Seconds() * 60) // 60 updates per second
	for i := 0; i < iterations; i++ {
		world.Update(1.0 / 60.0)
	}
	worldUpdate = time.Since(start) / time.Duration(iterations)

	// Benchmark entity queries
	start = time.Now()
	for i := 0; i < 1000; i++ {
		_ = world.GetEntitiesWith("position", "velocity")
	}
	entityQuery = time.Since(start) / 1000

	// Benchmark component access
	entities := world.GetEntitiesWith("position")
	start = time.Now()
	for i := 0; i < 10000; i++ {
		for _, e := range entities {
			_ = e.GetPosition()
		}
	}
	componentAccess = time.Since(start) / 10000

	return
}

func benchmarkTerrain() time.Duration {
	gen := terrain.NewBSPGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
	}

	start := time.Now()
	for i := 0; i < 10; i++ {
		_, _ = gen.Generate(int64(i), params)
	}
	return time.Since(start) / 10
}

func benchmarkEntity() time.Duration {
	gen := entity.NewEntityGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
	}

	start := time.Now()
	for i := 0; i < 100; i++ {
		_, _ = gen.Generate(int64(i), params)
	}
	return time.Since(start) / 100
}

func benchmarkItem() time.Duration {
	gen := item.NewItemGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
	}

	start := time.Now()
	for i := 0; i < 100; i++ {
		_, _ = gen.Generate(int64(i), params)
	}
	return time.Since(start) / 100
}

func benchmarkSpell() time.Duration {
	gen := magic.NewSpellGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
	}

	start := time.Now()
	for i := 0; i < 100; i++ {
		_, _ = gen.Generate(int64(i), params)
	}
	return time.Since(start) / 100
}

func writeReport(report *ProfileReport, filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	// Write markdown report
	fmt.Fprintf(f, "# Performance Assessment Report\n\n")
	fmt.Fprintf(f, "**Generated:** %s\n", time.Now().Format(time.RFC3339))
	fmt.Fprintf(f, "**Duration:** %v\n", report.Duration)
	fmt.Fprintf(f, "**Entity Count:** %d\n\n", report.EntityCount)

	fmt.Fprintf(f, "## ECS Performance\n\n")
	fmt.Fprintf(f, "| Operation | Time | Target | Status |\n")
	fmt.Fprintf(f, "|-----------|------|--------|--------|\n")

	worldUpdateMs := float64(report.WorldUpdateTime.Microseconds()) / 1000.0
	fmt.Fprintf(f, "| World Update | %.3f ms | <16.67 ms | %s |\n",
		worldUpdateMs,
		statusIcon(worldUpdateMs < 16.67))

	queryUs := float64(report.EntityQueryTime.Nanoseconds()) / 1000.0
	fmt.Fprintf(f, "| Entity Query | %.1f μs | <100 μs | %s |\n",
		queryUs,
		statusIcon(queryUs < 100))

	accessNs := float64(report.ComponentAccessTime.Nanoseconds())
	fmt.Fprintf(f, "| Component Access | %.1f ns | <5 ns | %s |\n\n",
		accessNs,
		statusIcon(accessNs < 5))

	fmt.Fprintf(f, "## Memory Analysis\n\n")
	fmt.Fprintf(f, "| Metric | Initial | Final | Change |\n")
	fmt.Fprintf(f, "|--------|---------|-------|--------|\n")
	fmt.Fprintf(f, "| Allocated | %.2f MB | %.2f MB | %.2f MB |\n",
		float64(report.InitialMemory.Alloc)/(1024*1024),
		float64(report.FinalMemory.Alloc)/(1024*1024),
		float64(report.MemoryGrowth)/(1024*1024))
	fmt.Fprintf(f, "| System | %.2f MB | %.2f MB | %.2f MB |\n",
		float64(report.InitialMemory.Sys)/(1024*1024),
		float64(report.FinalMemory.Sys)/(1024*1024),
		float64(report.FinalMemory.Sys-report.InitialMemory.Sys)/(1024*1024))
	fmt.Fprintf(f, "| GC Count | %d | %d | %d |\n",
		report.InitialMemory.NumGC,
		report.FinalMemory.NumGC,
		report.GCCount)

	if report.GCCount > 0 {
		avgPause := float64(report.GCPauseTotal.Microseconds()) / float64(report.GCCount) / 1000.0
		fmt.Fprintf(f, "| Avg GC Pause | - | - | %.2f ms |\n\n", avgPause)

		fmt.Fprintf(f, "**GC Pause Assessment:** %s (target: <5ms)\n\n",
			statusIcon(avgPause < 5.0))
	} else {
		fmt.Fprintf(f, "| Avg GC Pause | - | - | N/A |\n\n")
	}

	fmt.Fprintf(f, "## Procedural Generation Performance\n\n")
	fmt.Fprintf(f, "| Generator | Avg Time | Target | Status |\n")
	fmt.Fprintf(f, "|-----------|----------|--------|--------|\n")

	terrainMs := float64(report.TerrainGenTime.Microseconds()) / 1000.0
	fmt.Fprintf(f, "| Terrain | %.2f ms | <2000 ms | %s |\n",
		terrainMs,
		statusIcon(terrainMs < 2000))

	entityMs := float64(report.EntityGenTime.Microseconds()) / 1000.0
	fmt.Fprintf(f, "| Entity | %.2f ms | <1 ms | %s |\n",
		entityMs,
		statusIcon(entityMs < 1.0))

	itemMs := float64(report.ItemGenTime.Microseconds()) / 1000.0
	fmt.Fprintf(f, "| Item | %.2f ms | <1 ms | %s |\n",
		itemMs,
		statusIcon(itemMs < 1.0))

	spellMs := float64(report.SpellGenTime.Microseconds()) / 1000.0
	fmt.Fprintf(f, "| Spell | %.2f ms | <1 ms | %s |\n\n",
		spellMs,
		statusIcon(spellMs < 1.0))

	fmt.Fprintf(f, "## Summary\n\n")

	// Assess overall performance
	recs := generateRecommendations(report)
	if len(recs) == 0 {
		fmt.Fprintf(f, "✅ **All Performance Targets Met**\n\n")
		fmt.Fprintf(f, "The system is performing well across all metrics. ")
		fmt.Fprintf(f, "Continue monitoring for performance regressions.\n\n")
	} else {
		fmt.Fprintf(f, "### Recommendations\n\n")
		for i, rec := range recs {
			fmt.Fprintf(f, "%d. **%s**\n", i+1, rec.Title)
			fmt.Fprintf(f, "   - %s\n\n", rec.Description)
		}
	}

	fmt.Fprintf(f, "## Usage\n\n")
	fmt.Fprintf(f, "To re-run this assessment:\n")
	fmt.Fprintf(f, "```bash\n")
	fmt.Fprintf(f, "go run ./cmd/perfprofile -duration=%v -entities=%d\n", report.Duration, report.EntityCount)
	fmt.Fprintf(f, "```\n\n")

	fmt.Fprintf(f, "To profile with CPU/memory traces:\n")
	fmt.Fprintf(f, "```bash\n")
	fmt.Fprintf(f, "go run ./cmd/perfprofile -cpuprofile=cpu.prof -memprofile=mem.prof\n")
	fmt.Fprintf(f, "go tool pprof cpu.prof\n")
	fmt.Fprintf(f, "```\n")

	return nil
}

func statusIcon(passing bool) string {
	if passing {
		return "✅"
	}
	return "⚠️"
}

type Recommendation struct {
	Title       string
	Description string
}

func generateRecommendations(report *ProfileReport) []Recommendation {
	var recs []Recommendation

	// Check world update time
	worldUpdateMs := float64(report.WorldUpdateTime.Microseconds()) / 1000.0
	if worldUpdateMs >= 16.67 {
		recs = append(recs, Recommendation{
			Title:       "Optimize World Update",
			Description: fmt.Sprintf("World update taking %.2f ms (target: <16.67ms for 60 FPS)", worldUpdateMs),
		})
	}

	// Check entity query time
	queryUs := float64(report.EntityQueryTime.Nanoseconds()) / 1000.0
	if queryUs >= 100 {
		recs = append(recs, Recommendation{
			Title:       "Optimize Entity Queries",
			Description: fmt.Sprintf("Entity queries taking %.1f μs (target: <100μs). Verify query cache is working.", queryUs),
		})
	}

	// Check component access time
	accessNs := float64(report.ComponentAccessTime.Nanoseconds())
	if accessNs >= 5 {
		recs = append(recs, Recommendation{
			Title:       "Optimize Component Access",
			Description: fmt.Sprintf("Component access taking %.1f ns (target: <5ns). Verify fast-path getters are used.", accessNs),
		})
	}

	// Check GC pauses
	if report.GCCount > 0 {
		avgPause := float64(report.GCPauseTotal.Microseconds()) / float64(report.GCCount) / 1000.0
		if avgPause >= 5.0 {
			recs = append(recs, Recommendation{
				Title:       "Reduce GC Pauses",
				Description: fmt.Sprintf("Average GC pause: %.2f ms (target: <5ms). Implement object pooling for hot paths.", avgPause),
			})
		}
	}

	// Check terrain generation
	terrainMs := float64(report.TerrainGenTime.Microseconds()) / 1000.0
	if terrainMs >= 2000 {
		recs = append(recs, Recommendation{
			Title:       "Optimize Terrain Generation",
			Description: fmt.Sprintf("Terrain generation taking %.2f ms (target: <2000ms). Consider streaming generation.", terrainMs),
		})
	}

	return recs
}
