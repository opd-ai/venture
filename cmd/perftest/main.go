package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/logging"
	"github.com/sirupsen/logrus"
)

var (
	entityCount  = flag.Int("entities", 1000, "Number of entities to spawn")
	duration     = flag.Int("duration", 10, "Test duration in seconds")
	verbose      = flag.Bool("verbose", false, "Enable verbose logging")
	validate2k   = flag.Bool("validate-2k", false, "Run validation test with 2000 entities (for README claim)")
	targetFPS    = flag.Float64("target-fps", 60.0, "Target FPS to validate against")
	outputReport = flag.String("output", "", "Output performance report to file")
)

func main() {
	flag.Parse()

	// Initialize logger for test utility
	logger := logging.TestUtilityLogger("perftest")
	if *verbose {
		logger.SetLevel(logrus.DebugLevel)
	}

	// Validation mode for README claims
	if *validate2k {
		logger.Info("running validation test for README performance claim (2000 entities)")
		*entityCount = 2000
		*duration = 30     // Longer test for stability
		*targetFPS = 106.0 // Specific claim from README
	}

	logger.WithFields(logrus.Fields{
		"entities": *entityCount,
		"duration": *duration,
	}).Info("performance test starting")
	
	if *targetFPS != 60.0 {
		logger.WithField("targetFPS", *targetFPS).Info("custom target FPS")
	}

	// Create world with performance monitoring
	world := engine.NewWorld()
	monitor := engine.NewPerformanceMonitor(world)

	// Add spatial partitioning system
	spatialSystem := engine.NewSpatialPartitionSystem(10000, 10000)
	world.AddSystem(spatialSystem)

	// Add core systems
	movementSystem := &engine.MovementSystem{}
	collisionSystem := &engine.CollisionSystem{}

	world.AddSystem(movementSystem)
	world.AddSystem(collisionSystem)

	logger.Info("systems initialized: Movement, Collision, Spatial Partitioning")

	// Spawn entities
	logger.WithField("entityCount", *entityCount).Info("spawning entities...")
	startSpawn := time.Now()

	for i := 0; i < *entityCount; i++ {
		entity := world.CreateEntity()

		// Position scattered across world
		x := float64(i%100) * 100.0
		y := float64(i/100) * 100.0

		entity.AddComponent(&engine.PositionComponent{X: x, Y: y})
		entity.AddComponent(&engine.VelocityComponent{
			VX: float64((i%3)-1) * 50.0, // -50, 0, or 50
			VY: float64((i%5)-2) * 50.0,
		})

		// Add collider to only 10% of entities (reduces collision overhead)
		if i%10 == 0 {
			entity.AddComponent(&engine.ColliderComponent{
				Width:  16,
				Height: 16,
				Solid:  true,
				Layer:  1,
			})
		}
	}

	// Process initial entity additions
	world.Update(0)
	spawnDuration := time.Since(startSpawn)

	logger.WithFields(logrus.Fields{
		"entityCount": *entityCount,
		"duration":    float64(spawnDuration.Microseconds()) / 1000.0,
	}).Info("entities spawned")

	// Run simulation
	logger.Info("starting performance test")
	logger.WithFields(logrus.Fields{
		"targetFPS":       *targetFPS,
		"msPerFrame":      1000.0 / *targetFPS,
	}).Info("performance targets")

	frameDuration := time.Second / time.Duration(*targetFPS)
	endTime := time.Now().Add(time.Duration(*duration) * time.Second)
	frameCount := 0

	ticker := time.NewTicker(frameDuration)
	defer ticker.Stop()

	lastReport := time.Now()
	reportInterval := 1 * time.Second

	for time.Now().Before(endTime) {
		<-ticker.C

		deltaTime := frameDuration.Seconds()
		monitor.Update(deltaTime)

		frameCount++

		// Report every second
		if time.Since(lastReport) >= reportInterval {
			metrics := monitor.GetMetrics()

			if *verbose {
				fmt.Println(metrics.DetailedString())

				// Show spatial partition stats
				stats := spatialSystem.GetStatistics()
				fmt.Printf("Spatial Partition: %d entities, %d queries\n",
					stats["entity_count"], stats["query_count"])
			} else {
				fmt.Println(metrics.String())
			}

			lastReport = time.Now()
		}
	}

	// Final report
	logger.Info("performance test complete")
	metrics := monitor.GetMetrics()

	logger.WithFields(logrus.Fields{
		"totalFrames":       frameCount,
		"avgFPS":            metrics.FPS,
		"avgFrameTime":      float64(metrics.AverageFrameTime.Microseconds()) / 1000.0,
		"minFrameTime":      float64(metrics.MinFrameTime.Microseconds()) / 1000.0,
		"maxFrameTime":      float64(metrics.MaxFrameTime.Microseconds()) / 1000.0,
		"avgUpdateTime":     float64(metrics.AverageUpdateTime.Microseconds()) / 1000.0,
		"entityCount":       metrics.EntityCount,
		"activeEntityCount": metrics.ActiveEntityCount,
	}).Info("final statistics")

	fmt.Printf("\nFinal Statistics:\n")
	fmt.Printf("  Total Frames: %d\n", frameCount)
	fmt.Printf("  Average FPS: %.2f\n", metrics.FPS)
	fmt.Printf("  Average Frame Time: %.2fms\n", float64(metrics.AverageFrameTime.Microseconds())/1000.0)
	fmt.Printf("  Min Frame Time: %.2fms\n", float64(metrics.MinFrameTime.Microseconds())/1000.0)
	fmt.Printf("  Max Frame Time: %.2fms\n", float64(metrics.MaxFrameTime.Microseconds())/1000.0)
	fmt.Printf("  Average Update Time: %.2fms\n", float64(metrics.AverageUpdateTime.Microseconds())/1000.0)
	fmt.Printf("  Entity Count: %d (%d active)\n", metrics.EntityCount, metrics.ActiveEntityCount)

	fmt.Printf("\nSystem Breakdown:\n")
	percentages := metrics.GetFrameTimePercent()
	for name, percent := range percentages {
		fmt.Printf("  %s: %.2f%%\n", name, percent)
	}

	// Check if meeting target
	targetMet := metrics.FPS >= *targetFPS
	fmt.Printf("\nPerformance Target (%.0f FPS): ", *targetFPS)
	if targetMet {
		fmt.Printf("✅ MET (%.2f FPS)\n", metrics.FPS)
	} else {
		fmt.Printf("❌ NOT MET (%.2f FPS, shortfall: %.2f FPS)\n", metrics.FPS, *targetFPS-metrics.FPS)
	}

	// Output report to file if requested
	if *outputReport != "" {
		reportContent := fmt.Sprintf(`Performance Test Report
Generated: %s
Test Configuration:
  Entity Count: %d
  Duration: %d seconds
  Total Frames: %d

Results:
  Average FPS: %.2f
  Min Frame Time: %.2fms
  Max Frame Time: %.2fms
  Average Update Time: %.2fms

Target: %.0f FPS - %s

System Breakdown:
`, time.Now().Format(time.RFC3339), *entityCount, *duration, frameCount,
			metrics.FPS,
			float64(metrics.MinFrameTime.Microseconds())/1000.0,
			float64(metrics.MaxFrameTime.Microseconds())/1000.0,
			float64(metrics.AverageUpdateTime.Microseconds())/1000.0,
			*targetFPS,
			map[bool]string{true: "MET ✅", false: "NOT MET ❌"}[targetMet])

		for name, percent := range percentages {
			reportContent += fmt.Sprintf("  %s: %.2f%%\n", name, percent)
		}

		if err := os.WriteFile(*outputReport, []byte(reportContent), 0o644); err != nil {
			logger.WithError(err).Error("failed to write report")
		} else {
			logger.WithField("path", *outputReport).Info("performance report written")
		}
	}

	// Spatial partition stats
	stats := spatialSystem.GetStatistics()
	fmt.Printf("\nSpatial Partition Statistics:\n")
	fmt.Printf("  Entities Tracked: %d\n", stats["entity_count"])
	fmt.Printf("  Total Queries: %d\n", stats["query_count"])
	fmt.Printf("  Last Rebuild Time: %.2fms\n", stats["last_rebuild_time"].(float64)*1000.0)

	// Test spatial query performance
	fmt.Printf("\nSpatial Query Performance Test:\n")
	queryStart := time.Now()
	queryCount := 1000
	for i := 0; i < queryCount; i++ {
		x := float64(i % 5000)
		y := float64((i * 7) % 5000)
		spatialSystem.QueryRadius(x, y, 100)
	}
	queryDuration := time.Since(queryStart)
	avgQueryTime := queryDuration / time.Duration(queryCount)
	fmt.Printf("  %d queries in %.2fms\n", queryCount, float64(queryDuration.Microseconds())/1000.0)
	fmt.Printf("  Average query time: %.2fμs\n", float64(avgQueryTime.Nanoseconds())/1000.0)

	logger.Info("performance test complete!")
}
