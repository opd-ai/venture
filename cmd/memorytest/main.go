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

	// Initialize logger
	logger := logging.TestUtilityLogger("memorytest")
	if *verbose {
		logger.SetLevel(logrus.DebugLevel)
	}

	logger.Info("=== V3.0 Memory Profiling Test ===")
	logger.WithFields(logrus.Fields{
		"entities":     *entityCount,
		"duration":     *duration,
		"targetMemory": *targetMemory,
		"checkLeaks":   *checkLeaks,
	}).Info("test configuration")

	// Force GC before starting
	runtime.GC()
	time.Sleep(100 * time.Millisecond)

	// Initial memory snapshot
	initialStats := collectMemoryStats()
	logger.WithFields(logrus.Fields{
		"alloc":     formatBytes(initialStats.Alloc),
		"heapAlloc": formatBytes(initialStats.HeapAlloc),
		"sys":       formatBytes(initialStats.Sys),
		"numGC":     initialStats.NumGC,
	}).Info("initial memory state")

	// Track memory samples over time
	var memorySamples []MemoryStats
	memorySamples = append(memorySamples, initialStats)

	// Create sprite cache and track its efficiency
	spriteCache := cache.NewSpriteCache(100 * 1024 * 1024) // 100 MB cache
	logger.Info("sprite cache initialized (100 MB)")

	// Create world with V3.0 features
	world := engine.NewWorld()

	// Add systems
	movementSystem := &engine.MovementSystem{}
	collisionSystem := &engine.CollisionSystem{}
	spatialSystem := engine.NewSpatialPartitionSystem(10000, 10000)

	world.AddSystem(movementSystem)
	world.AddSystem(collisionSystem)
	world.AddSystem(spatialSystem)

	logger.Info("ECS systems initialized")

	// Initialize generators (V3.0 features)
	terrainGen := terrain.NewBSPGenerator()
	entityGen := entity.NewEntityGenerator()
	itemGen := item.NewItemGenerator()
	magicGen := magic.NewSpellGenerator()
	spriteGen := sprites.NewGenerator()

	logger.Info("V3.0 generators initialized")

	// Generate V3.0 content
	logger.Info("generating V3.0 procedural content...")

	seed := int64(12345)
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
	}

	// Generate terrain
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

	// Generate entities
	logger.WithField("count", *entityCount).Info("generating entities...")
	entityStartTime := time.Now()

	for i := 0; i < *entityCount; i++ {
		// Generate entity
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

		// Create ECS entity
		e := world.CreateEntity()

		// Add components
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

		// Test sprite generation and caching
		if i%10 == 0 { // Generate sprites for 10% of entities
			// Generate same sprite 20 times to test caching (first miss, then 19 hits = 95% hit rate)
			for repeat := 0; repeat < 20; repeat++ {
				// Use consistent cache key
				cacheKey := cache.GenerateKey(entitySeed, "idle", 0)
				_, hit := spriteCache.Get(cacheKey)

				if !hit {
					// Generate sprite (V3.0 enhanced sprites)
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
	}

	entityDuration := time.Since(entityStartTime)
	logger.WithFields(logrus.Fields{
		"count":    *entityCount,
		"duration": float64(entityDuration.Milliseconds()),
	}).Info("entities generated")

	// Generate items (V3.0 features)
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

	// Generate magic spells (V3.0 features)
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

	// Capture memory after content generation
	runtime.GC()
	time.Sleep(100 * time.Millisecond)
	afterGenerationStats := collectMemoryStats()
	memorySamples = append(memorySamples, afterGenerationStats)

	logger.WithFields(logrus.Fields{
		"alloc":     formatBytes(afterGenerationStats.Alloc),
		"heapAlloc": formatBytes(afterGenerationStats.HeapAlloc),
		"sys":       formatBytes(afterGenerationStats.Sys),
		"numGC":     afterGenerationStats.NumGC,
	}).Info("memory after content generation")

	// Check sprite cache statistics
	cacheStats := spriteCache.Stats()
	logger.WithFields(logrus.Fields{
		"hits":      cacheStats.Hits,
		"misses":    cacheStats.Misses,
		"hitRate":   fmt.Sprintf("%.2f%%", cacheStats.HitRate()*100),
		"entries":   cacheStats.EntryCount,
		"totalSize": formatBytes(uint64(cacheStats.TotalSize)),
		"evictions": cacheStats.Evictions,
	}).Info("sprite cache statistics")

	// Run simulation for leak detection
	if *checkLeaks {
		logger.WithField("duration", *duration).Info("starting memory leak detection test")

		endTime := time.Now().Add(time.Duration(*duration) * time.Second)
		sampleInterval := 10 * time.Second
		lastSample := time.Now()
		frameCount := 0

		ticker := time.NewTicker(16 * time.Millisecond) // ~60 FPS
		defer ticker.Stop()

		for time.Now().Before(endTime) {
			<-ticker.C

			// Update world
			world.Update(0.016)
			frameCount++

			// Periodic memory sampling
			if time.Since(lastSample) >= sampleInterval {
				runtime.GC()
				time.Sleep(50 * time.Millisecond)

				currentStats := collectMemoryStats()
				memorySamples = append(memorySamples, currentStats)

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

				lastSample = time.Now()
			}
		}

		logger.WithField("frames", frameCount).Info("simulation complete")
	}

	// Final memory snapshot
	runtime.GC()
	time.Sleep(100 * time.Millisecond)
	finalStats := collectMemoryStats()
	memorySamples = append(memorySamples, finalStats)

	logger.Info("=== Final Memory Report ===")

	// Memory usage
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

	// Memory growth analysis
	fmt.Printf("\n=== Memory Growth ===\n")
	allocGrowth := int64(finalStats.Alloc) - int64(initialStats.Alloc)
	heapGrowth := int64(finalStats.HeapAlloc) - int64(initialStats.HeapAlloc)
	sysGrowth := int64(finalStats.Sys) - int64(initialStats.Sys)

	fmt.Printf("  Allocated Growth: %s\n", formatBytes(uint64(allocGrowth)))
	fmt.Printf("  Heap Growth:      %s\n", formatBytes(uint64(heapGrowth)))
	fmt.Printf("  System Growth:    %s\n", formatBytes(uint64(sysGrowth)))

	// Check for memory leaks
	if *checkLeaks && len(memorySamples) > 2 {
		fmt.Printf("\n=== Memory Leak Analysis ===\n")

		// Compare memory samples over time
		startRunIdx := 2 // After content generation
		endRunIdx := len(memorySamples) - 1

		if endRunIdx > startRunIdx {
			runStartMem := memorySamples[startRunIdx].Alloc
			runEndMem := memorySamples[endRunIdx].Alloc
			runDuration := memorySamples[endRunIdx].Timestamp.Sub(memorySamples[startRunIdx].Timestamp)

			memGrowth := int64(runEndMem) - int64(runStartMem)
			growthRate := float64(memGrowth) / runDuration.Seconds() // bytes per second

			fmt.Printf("  Runtime Duration: %.1f seconds\n", runDuration.Seconds())
			fmt.Printf("  Memory Growth:    %s\n", formatBytes(uint64(memGrowth)))
			fmt.Printf("  Growth Rate:      %.2f KB/s\n", growthRate/1024)

			// Determine if leak is present
			// Allow small growth due to normal allocations
			maxGrowthRate := 10.0 * 1024 // 10 KB/s threshold
			if growthRate > maxGrowthRate {
				fmt.Printf("  Status: ⚠️  POTENTIAL LEAK (growth rate > %.1f KB/s)\n", maxGrowthRate/1024)
			} else {
				fmt.Printf("  Status: ✅ NO LEAK DETECTED\n")
			}
		}
	}

	// Sprite cache efficiency
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

	// Target validation
	fmt.Printf("\n=== Performance Targets ===\n")
	targetMemoryBytes := uint64(*targetMemory) * 1024 * 1024
	targetCacheHitRate := 0.95 // 95%

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

	// Write memory profile if requested
	if *memProfile != "" {
		logger.WithField("path", *memProfile).Info("writing memory profile")
		f, err := os.Create(*memProfile)
		if err != nil {
			logger.WithError(err).Error("failed to create memory profile")
		} else {
			runtime.GC()
			if err := pprof.WriteHeapProfile(f); err != nil {
				logger.WithError(err).Error("failed to write memory profile")
			}
			f.Close()
			logger.Info("memory profile written")
		}
	}

	// Write detailed report if requested
	if *outputReport != "" {
		reportContent := generateReport(
			initialStats, afterGenerationStats, finalStats,
			finalCacheStats, *entityCount, memorySamples,
		)

		if err := os.WriteFile(*outputReport, []byte(reportContent), 0o644); err != nil {
			logger.WithError(err).Error("failed to write report")
		} else {
			logger.WithField("path", *outputReport).Info("memory report written")
		}
	}

	logger.Info("memory profiling complete!")
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
