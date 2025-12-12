package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/logging"
	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/opd-ai/venture/pkg/procgen/entity"
	"github.com/opd-ai/venture/pkg/procgen/item"
	"github.com/opd-ai/venture/pkg/procgen/magic"
	"github.com/opd-ai/venture/pkg/procgen/terrain"
	"github.com/opd-ai/venture/pkg/rendering/cache"
	"github.com/opd-ai/venture/pkg/rendering/sprites"
	"github.com/sirupsen/logrus"
)

var (
	entityCount  = flag.Int("entities", 2000, "Number of entities to spawn")
	duration     = flag.Int("duration", 300, "Test duration in seconds (for leak detection)")
	memProfile   = flag.String("memprofile", "", "Write memory profile to file")
	checkLeaks   = flag.Bool("check-leaks", true, "Monitor for memory leaks")
	targetMemory = flag.Int64("target-memory", 500, "Target memory limit in MB")
	verbose      = flag.Bool("verbose", false, "Enable verbose logging")
	outputReport = flag.String("output", "", "Output memory report to file")
)

// MemoryStats tracks memory usage over time
type MemoryStats struct {
	Timestamp   time.Time
	Alloc       uint64 // Allocated memory in bytes
	TotalAlloc  uint64 // Total allocated memory
	Sys         uint64 // System memory obtained from OS
	NumGC       uint32 // Number of GC runs
	HeapAlloc   uint64 // Heap allocated memory
	HeapSys     uint64 // Heap system memory
	HeapObjects uint64 // Number of heap objects
	StackInuse  uint64 // Stack memory in use
	StackSys    uint64 // Stack system memory
}

// collectMemoryStats captures current memory statistics
func collectMemoryStats() MemoryStats {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return MemoryStats{
		Timestamp:   time.Now(),
		Alloc:       m.Alloc,
		TotalAlloc:  m.TotalAlloc,
		Sys:         m.Sys,
		NumGC:       m.NumGC,
		HeapAlloc:   m.HeapAlloc,
		HeapSys:     m.HeapSys,
		HeapObjects: m.HeapObjects,
		StackInuse:  m.StackInuse,
		StackSys:    m.StackSys,
	}
}

// formatBytes converts bytes to human-readable format
func formatBytes(bytes uint64) string {
	if bytes < 1024 {
		return fmt.Sprintf("%d B", bytes)
	} else if bytes < 1024*1024 {
		return fmt.Sprintf("%.2f KB", float64(bytes)/1024)
	} else if bytes < 1024*1024*1024 {
		return fmt.Sprintf("%.2f MB", float64(bytes)/(1024*1024))
	}
	return fmt.Sprintf("%.2f GB", float64(bytes)/(1024*1024*1024))
}

func main() {
	flag.Parse()

	logger := initializeLogger()
	logTestConfiguration(logger)

	initialStats := captureInitialMemory(logger)
	memorySamples := []MemoryStats{initialStats}

	spriteCache := createSpriteCache(logger)
	world := setupWorld(logger)
	generators := initializeGenerators(logger)

	seed := int64(12345)
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
	}

	generateTerrain(generators.terrainGen, seed, params, logger)
	generateEntities(world, generators.entityGen, generators.spriteGen, spriteCache, seed, params, logger)
	generateItems(generators.itemGen, seed, params, logger)
	generateMagic(generators.magicGen, seed, params, logger)

	afterGenerationStats := capturePostGenerationMemory(logger)
	memorySamples = append(memorySamples, afterGenerationStats)
	logCacheStatistics(spriteCache, logger)

	if *checkLeaks {
		runLeakDetectionTest(world, initialStats, &memorySamples, logger)
	}

	finalStats := captureFinalMemory(logger)
	memorySamples = append(memorySamples, finalStats)

	printMemoryReport(initialStats, afterGenerationStats, finalStats, memorySamples, spriteCache)
	writeOptionalOutputs(logger, initialStats, afterGenerationStats, finalStats, spriteCache.Stats(), memorySamples)

	logger.Info("memory profiling complete!")
}

// initializeLogger creates and configures the logger based on verbose flag.
func initializeLogger() *logrus.Logger {
	logger := logging.TestUtilityLogger("memorytest")
	if *verbose {
		logger.SetLevel(logrus.DebugLevel)
	}
	return logger
}

// logTestConfiguration logs the test parameters and configuration.
func logTestConfiguration(logger *logrus.Logger) {
	logger.Info("=== V3.0 Memory Profiling Test ===")
	logger.WithFields(logrus.Fields{
		"entities":     *entityCount,
		"duration":     *duration,
		"targetMemory": *targetMemory,
		"checkLeaks":   *checkLeaks,
	}).Info("test configuration")
}

// captureInitialMemory forces GC and captures the initial memory state.
func captureInitialMemory(logger *logrus.Logger) MemoryStats {
	runtime.GC()
	time.Sleep(100 * time.Millisecond)

	initialStats := collectMemoryStats()
	logger.WithFields(logrus.Fields{
		"alloc":     formatBytes(initialStats.Alloc),
		"heapAlloc": formatBytes(initialStats.HeapAlloc),
		"sys":       formatBytes(initialStats.Sys),
		"numGC":     initialStats.NumGC,
	}).Info("initial memory state")

	return initialStats
}

// createSpriteCache initializes the sprite cache with a 100 MB limit.
func createSpriteCache(logger *logrus.Logger) *cache.SpriteCache {
	spriteCache := cache.NewSpriteCache(100 * 1024 * 1024)
	logger.Info("sprite cache initialized (100 MB)")
	return spriteCache
}

// generatorSet holds all procedural content generators.
type generatorSet struct {
	terrainGen *terrain.BSPGenerator
	entityGen  *entity.EntityGenerator
	itemGen    *item.ItemGenerator
	magicGen   *magic.SpellGenerator
	spriteGen  *sprites.Generator
}

// setupWorld creates and initializes the ECS world with all required systems.
func setupWorld(logger *logrus.Logger) *engine.World {
	world := engine.NewWorld()

	movementSystem := &engine.MovementSystem{}
	collisionSystem := &engine.CollisionSystem{}
	spatialSystem := engine.NewSpatialPartitionSystem(10000, 10000)

	world.AddSystem(movementSystem)
	world.AddSystem(collisionSystem)
	world.AddSystem(spatialSystem)

	logger.Info("ECS systems initialized")
	return world
}

// initializeGenerators creates all V3.0 procedural generators.
func initializeGenerators(logger *logrus.Logger) generatorSet {
	gens := generatorSet{
		terrainGen: terrain.NewBSPGenerator(),
		entityGen:  entity.NewEntityGenerator(),
		itemGen:    item.NewItemGenerator(),
		magicGen:   magic.NewSpellGenerator(),
		spriteGen:  sprites.NewGenerator(),
	}

	logger.Info("V3.0 generators initialized")
	logger.Info("generating V3.0 procedural content...")
	return gens
}

// generateTerrain generates and logs terrain data.
func generateTerrain(terrainGen *terrain.BSPGenerator, seed int64, params procgen.GenerationParams, logger *logrus.Logger) {
	logger.Info("generating terrain...")
	terrainResult, err := terrainGen.Generate(seed, params)
	if err != nil {
		logger.WithError(err).Fatal("terrain generation failed")
	}
	terrain := terrainResult.(*terrain.Terrain)
	logger.WithFields(logrus.Fields{
		"width":  terrain.Width,
		"height": terrain.Height,
		"rooms":  len(terrain.Rooms),
	}).Info("terrain generated")
}

// generateEntities creates entities with components and tests sprite caching.
func generateEntities(world *engine.World, entityGen *entity.EntityGenerator, spriteGen *sprites.Generator, spriteCache *cache.SpriteCache, seed int64, params procgen.GenerationParams, logger *logrus.Logger) {
	logger.WithField("count", *entityCount).Info("generating entities...")
	entityStartTime := time.Now()

	for i := 0; i < *entityCount; i++ {
		entitySeed := seed + int64(i)
		entityResult, err := entityGen.Generate(entitySeed, params)
		if err != nil {
			logger.WithError(err).Warn("entity generation failed")
			continue
		}

		entities := entityResult.([]*entity.Entity)
		if len(entities) == 0 {
			continue
		}
		entityData := entities[0]

		e := world.CreateEntity()
		x := float64(i%100) * 10.0
		y := float64(i/100) * 10.0

		e.AddComponent(&engine.PositionComponent{X: x, Y: y})
		e.AddComponent(&engine.VelocityComponent{
			VX: float64((i%3)-1) * 20.0,
			VY: float64((i%5)-2) * 20.0,
		})
		e.AddComponent(&engine.HealthComponent{
			Current: float64(entityData.Stats.Health),
			Max:     float64(entityData.Stats.Health),
		})

		if i%10 == 0 {
			testSpriteCaching(spriteGen, spriteCache, entitySeed)
		}
	}

	entityDuration := time.Since(entityStartTime)
	logger.WithFields(logrus.Fields{
		"count":    *entityCount,
		"duration": float64(entityDuration.Milliseconds()),
	}).Info("entities generated")
}

// testSpriteCaching generates sprites 20 times to validate cache hit rate.
func testSpriteCaching(spriteGen *sprites.Generator, spriteCache *cache.SpriteCache, entitySeed int64) {
	for repeat := 0; repeat < 20; repeat++ {
		cacheKey := cache.GenerateKey(entitySeed, "idle", 0)
		_, hit := spriteCache.Get(cacheKey)

		if !hit {
			spriteConfig := sprites.Config{
				Type:       sprites.SpriteEntity,
				Width:      32,
				Height:     32,
				Seed:       entitySeed,
				GenreID:    "fantasy",
				Complexity: 0.7,
				Variation:  0,
			}

			spriteImg, err := spriteGen.Generate(spriteConfig)
			if err == nil {
				spriteCache.Put(cacheKey, spriteImg)
			}
		}
	}
}

// generateItems generates item content and logs results.
func generateItems(itemGen *item.ItemGenerator, seed int64, params procgen.GenerationParams, logger *logrus.Logger) {
	logger.Info("generating items...")
	itemCount := 100
	for i := 0; i < itemCount; i++ {
		itemSeed := seed + int64(i*1000)
		_, err := itemGen.Generate(itemSeed, params)
		if err != nil {
			logger.WithError(err).Warn("item generation failed")
		}
	}
	logger.WithField("count", itemCount).Info("items generated")
}

// generateMagic generates magic spell content and logs results.
func generateMagic(magicGen *magic.SpellGenerator, seed int64, params procgen.GenerationParams, logger *logrus.Logger) {
	logger.Info("generating magic spells...")
	spellCount := 50
	for i := 0; i < spellCount; i++ {
		spellSeed := seed + int64(i*2000)
		_, err := magicGen.Generate(spellSeed, params)
		if err != nil {
			logger.WithError(err).Warn("spell generation failed")
		}
	}
	logger.WithField("count", spellCount).Info("spells generated")
}

// capturePostGenerationMemory captures and logs memory state after content generation.
func capturePostGenerationMemory(logger *logrus.Logger) MemoryStats {
	runtime.GC()
	time.Sleep(100 * time.Millisecond)
	afterGenerationStats := collectMemoryStats()

	logger.WithFields(logrus.Fields{
		"alloc":     formatBytes(afterGenerationStats.Alloc),
		"heapAlloc": formatBytes(afterGenerationStats.HeapAlloc),
		"sys":       formatBytes(afterGenerationStats.Sys),
		"numGC":     afterGenerationStats.NumGC,
	}).Info("memory after content generation")

	return afterGenerationStats
}

// logCacheStatistics logs sprite cache performance metrics.
func logCacheStatistics(spriteCache *cache.SpriteCache, logger *logrus.Logger) {
	cacheStats := spriteCache.Stats()
	logger.WithFields(logrus.Fields{
		"hits":      cacheStats.Hits,
		"misses":    cacheStats.Misses,
		"hitRate":   fmt.Sprintf("%.2f%%", cacheStats.HitRate()*100),
		"entries":   cacheStats.EntryCount,
		"totalSize": formatBytes(uint64(cacheStats.TotalSize)),
		"evictions": cacheStats.Evictions,
	}).Info("sprite cache statistics")
}

// runLeakDetectionTest runs the simulation loop to detect memory leaks.
func runLeakDetectionTest(world *engine.World, initialStats MemoryStats, memorySamples *[]MemoryStats, logger *logrus.Logger) {
	logger.WithField("duration", *duration).Info("starting memory leak detection test")

	endTime := time.Now().Add(time.Duration(*duration) * time.Second)
	sampleInterval := 10 * time.Second
	lastSample := time.Now()
	frameCount := 0

	ticker := time.NewTicker(16 * time.Millisecond)
	defer ticker.Stop()

	for time.Now().Before(endTime) {
		<-ticker.C
		world.Update(0.016)
		frameCount++

		if time.Since(lastSample) >= sampleInterval {
			sampleMemory(initialStats, memorySamples, logger, frameCount)
			lastSample = time.Now()
		}
	}

	logger.WithField("frames", frameCount).Info("simulation complete")
}

// sampleMemory collects and logs a memory sample during leak detection.
func sampleMemory(initialStats MemoryStats, memorySamples *[]MemoryStats, logger *logrus.Logger, frameCount int) {
	runtime.GC()
	time.Sleep(50 * time.Millisecond)

	currentStats := collectMemoryStats()
	*memorySamples = append(*memorySamples, currentStats)

	elapsed := currentStats.Timestamp.Sub(initialStats.Timestamp)

	logger.WithFields(logrus.Fields{
		"elapsed":   elapsed.Seconds(),
		"alloc":     formatBytes(currentStats.Alloc),
		"heapAlloc": formatBytes(currentStats.HeapAlloc),
		"sys":       formatBytes(currentStats.Sys),
		"heapObjs":  currentStats.HeapObjects,
		"numGC":     currentStats.NumGC,
		"frames":    frameCount,
	}).Info("memory sample")
}

// captureFinalMemory captures and logs the final memory state.
func captureFinalMemory(logger *logrus.Logger) MemoryStats {
	runtime.GC()
	time.Sleep(100 * time.Millisecond)
	finalStats := collectMemoryStats()

	logger.Info("=== Final Memory Report ===")
	return finalStats
}

// printMemoryReport prints the complete memory profiling report to stdout.
func printMemoryReport(initialStats, afterGenerationStats, finalStats MemoryStats, memorySamples []MemoryStats, spriteCache *cache.SpriteCache) {
	printMemoryUsage(initialStats, afterGenerationStats, finalStats)
	printMemoryGrowth(initialStats, finalStats)
	printLeakAnalysis(memorySamples)
	printCacheEfficiency(spriteCache)
	printPerformanceTargets(finalStats, spriteCache)
}

// printMemoryUsage prints memory usage at different stages.
func printMemoryUsage(initialStats, afterGenerationStats, finalStats MemoryStats) {
	fmt.Printf("\n=== Memory Usage ===\n")
	fmt.Printf("Initial Memory:\n")
	fmt.Printf("  Allocated: %s\n", formatBytes(initialStats.Alloc))
	fmt.Printf("  Heap:      %s\n", formatBytes(initialStats.HeapAlloc))
	fmt.Printf("  System:    %s\n", formatBytes(initialStats.Sys))

	fmt.Printf("\nAfter Content Generation:\n")
	fmt.Printf("  Allocated: %s\n", formatBytes(afterGenerationStats.Alloc))
	fmt.Printf("  Heap:      %s\n", formatBytes(afterGenerationStats.HeapAlloc))
	fmt.Printf("  System:    %s\n", formatBytes(afterGenerationStats.Sys))

	fmt.Printf("\nFinal Memory:\n")
	fmt.Printf("  Allocated: %s\n", formatBytes(finalStats.Alloc))
	fmt.Printf("  Heap:      %s\n", formatBytes(finalStats.HeapAlloc))
	fmt.Printf("  System:    %s\n", formatBytes(finalStats.Sys))
	fmt.Printf("  Heap Objects: %d\n", finalStats.HeapObjects)
	fmt.Printf("  Stack In Use: %s\n", formatBytes(finalStats.StackInuse))
	fmt.Printf("  GC Runs:   %d\n", finalStats.NumGC)
}

// printMemoryGrowth prints memory growth analysis.
func printMemoryGrowth(initialStats, finalStats MemoryStats) {
	fmt.Printf("\n=== Memory Growth ===\n")
	allocGrowth := int64(finalStats.Alloc) - int64(initialStats.Alloc)
	heapGrowth := int64(finalStats.HeapAlloc) - int64(initialStats.HeapAlloc)
	sysGrowth := int64(finalStats.Sys) - int64(initialStats.Sys)

	fmt.Printf("  Allocated Growth: %s\n", formatBytes(uint64(allocGrowth)))
	fmt.Printf("  Heap Growth:      %s\n", formatBytes(uint64(heapGrowth)))
	fmt.Printf("  System Growth:    %s\n", formatBytes(uint64(sysGrowth)))
}

// printLeakAnalysis analyzes memory samples for potential leaks.
func printLeakAnalysis(memorySamples []MemoryStats) {
	if !*checkLeaks || len(memorySamples) <= 2 {
		return
	}

	fmt.Printf("\n=== Memory Leak Analysis ===\n")

	startRunIdx := 2
	endRunIdx := len(memorySamples) - 1

	if endRunIdx <= startRunIdx {
		return
	}

	runStartMem := memorySamples[startRunIdx].Alloc
	runEndMem := memorySamples[endRunIdx].Alloc
	runDuration := memorySamples[endRunIdx].Timestamp.Sub(memorySamples[startRunIdx].Timestamp)

	memGrowth := int64(runEndMem) - int64(runStartMem)
	growthRate := float64(memGrowth) / runDuration.Seconds()

	fmt.Printf("  Runtime Duration: %.1f seconds\n", runDuration.Seconds())
	fmt.Printf("  Memory Growth:    %s\n", formatBytes(uint64(memGrowth)))
	fmt.Printf("  Growth Rate:      %.2f KB/s\n", growthRate/1024)

	maxGrowthRate := 10.0 * 1024
	if growthRate > maxGrowthRate {
		fmt.Printf("  Status: ⚠️  POTENTIAL LEAK (growth rate > %.1f KB/s)\n", maxGrowthRate/1024)
	} else {
		fmt.Printf("  Status: ✅ NO LEAK DETECTED\n")
	}
}

// printCacheEfficiency prints sprite cache statistics.
func printCacheEfficiency(spriteCache *cache.SpriteCache) {
	fmt.Printf("\n=== Sprite Cache Efficiency ===\n")
	finalCacheStats := spriteCache.Stats()
	fmt.Printf("  Total Accesses: %d\n", finalCacheStats.Hits+finalCacheStats.Misses)
	fmt.Printf("  Cache Hits:     %d\n", finalCacheStats.Hits)
	fmt.Printf("  Cache Misses:   %d\n", finalCacheStats.Misses)
	fmt.Printf("  Hit Rate:       %.2f%%\n", finalCacheStats.HitRate()*100)
	fmt.Printf("  Entries:        %d\n", finalCacheStats.EntryCount)
	fmt.Printf("  Cache Size:     %s / %s\n",
		formatBytes(uint64(finalCacheStats.TotalSize)),
		formatBytes(uint64(spriteCache.MaxSize())))
	fmt.Printf("  Evictions:      %d\n", finalCacheStats.Evictions)
}

// printPerformanceTargets validates and prints performance target results.
func printPerformanceTargets(finalStats MemoryStats, spriteCache *cache.SpriteCache) {
	fmt.Printf("\n=== Performance Targets ===\n")
	targetMemoryBytes := uint64(*targetMemory) * 1024 * 1024
	targetCacheHitRate := 0.95

	finalCacheStats := spriteCache.Stats()
	memoryMet := finalStats.Alloc < targetMemoryBytes
	cacheEfficiencyMet := finalCacheStats.HitRate() >= targetCacheHitRate

	fmt.Printf("  Memory Target (<500 MB): ")
	if memoryMet {
		fmt.Printf("✅ MET (%s)\n", formatBytes(finalStats.Alloc))
	} else {
		fmt.Printf("❌ NOT MET (%s, excess: %s)\n",
			formatBytes(finalStats.Alloc),
			formatBytes(finalStats.Alloc-targetMemoryBytes))
	}

	fmt.Printf("  Cache Hit Rate (≥95%%): ")
	if cacheEfficiencyMet {
		fmt.Printf("✅ MET (%.2f%%)\n", finalCacheStats.HitRate()*100)
	} else {
		fmt.Printf("❌ NOT MET (%.2f%%, shortfall: %.2f%%)\n",
			finalCacheStats.HitRate()*100,
			(targetCacheHitRate-finalCacheStats.HitRate())*100)
	}
}

// writeOptionalOutputs writes memory profile and report files if requested.
func writeOptionalOutputs(logger *logrus.Logger, initialStats, afterGenerationStats, finalStats MemoryStats, cacheStats cache.Statistics, memorySamples []MemoryStats) {
	if *memProfile != "" {
		writeMemoryProfile(logger)
	}

	if *outputReport != "" {
		writeMemoryReport(logger, initialStats, afterGenerationStats, finalStats, cacheStats, memorySamples)
	}
}

// writeMemoryProfile writes the heap profile to the specified file.
func writeMemoryProfile(logger *logrus.Logger) {
	logger.WithField("path", *memProfile).Info("writing memory profile")
	f, err := os.Create(*memProfile)
	if err != nil {
		logger.WithError(err).Error("failed to create memory profile")
		return
	}
	defer f.Close()

	runtime.GC()
	if err := pprof.WriteHeapProfile(f); err != nil {
		logger.WithError(err).Error("failed to write memory profile")
		return
	}
	logger.Info("memory profile written")
}

// writeMemoryReport generates and writes the detailed memory report.
func writeMemoryReport(logger *logrus.Logger, initialStats, afterGenerationStats, finalStats MemoryStats, cacheStats cache.Statistics, memorySamples []MemoryStats) {
	reportContent := generateReport(
		initialStats, afterGenerationStats, finalStats,
		cacheStats, *entityCount, memorySamples,
	)

	if err := os.WriteFile(*outputReport, []byte(reportContent), 0o644); err != nil {
		logger.WithError(err).Error("failed to write report")
	} else {
		logger.WithField("path", *outputReport).Info("memory report written")
	}
}

// generateReport creates a detailed text report
func generateReport(initial, afterGen, final MemoryStats, cacheStats cache.Statistics, entityCount int, samples []MemoryStats) string {
	report := fmt.Sprintf(`V3.0 Memory Profiling Report
Generated: %s

=== Test Configuration ===
Entity Count: %d
Test Duration: %d samples
Cache Size: 100 MB

=== Memory Usage ===

Initial State:
  Allocated:    %s
  Heap:         %s
  System:       %s
  Heap Objects: %d

After Content Generation:
  Allocated:    %s (growth: %s)
  Heap:         %s (growth: %s)
  System:       %s (growth: %s)
  Heap Objects: %d

Final State:
  Allocated:    %s
  Heap:         %s
  System:       %s
  Heap Objects: %d
  Stack In Use: %s
  GC Runs:      %d

=== Sprite Cache Efficiency ===
Total Accesses: %d
Cache Hits:     %d
Cache Misses:   %d
Hit Rate:       %.2f%%
Entries:        %d
Cache Size:     %s
Evictions:      %d

=== Memory Samples Over Time ===
`,
		time.Now().Format(time.RFC3339),
		entityCount,
		len(samples),

		formatBytes(initial.Alloc),
		formatBytes(initial.HeapAlloc),
		formatBytes(initial.Sys),
		initial.HeapObjects,

		formatBytes(afterGen.Alloc),
		formatBytes(afterGen.Alloc-initial.Alloc),
		formatBytes(afterGen.HeapAlloc),
		formatBytes(afterGen.HeapAlloc-initial.HeapAlloc),
		formatBytes(afterGen.Sys),
		formatBytes(afterGen.Sys-initial.Sys),
		afterGen.HeapObjects,

		formatBytes(final.Alloc),
		formatBytes(final.HeapAlloc),
		formatBytes(final.Sys),
		final.HeapObjects,
		formatBytes(final.StackInuse),
		final.NumGC,

		cacheStats.Hits+cacheStats.Misses,
		cacheStats.Hits,
		cacheStats.Misses,
		cacheStats.HitRate()*100,
		cacheStats.EntryCount,
		formatBytes(uint64(cacheStats.TotalSize)),
		cacheStats.Evictions,
	)

	// Add memory sample timeline
	if len(samples) > 0 {
		startTime := samples[0].Timestamp
		for i, sample := range samples {
			elapsed := sample.Timestamp.Sub(startTime).Seconds()
			report += fmt.Sprintf("Sample %d (T+%.1fs): Alloc=%s, Heap=%s, GC=%d\n",
				i, elapsed,
				formatBytes(sample.Alloc),
				formatBytes(sample.HeapAlloc),
				sample.NumGC,
			)
		}
	}

	return report
}
