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

	logger := setupLogger()
	configureTestMode(logger)

	world, monitor := initializeWorld(logger)
	spawnTestEntities(world, logger)

	metrics, frameCount := runPerformanceTest(world, monitor, logger)

	displayResults(metrics, frameCount, logger)
	writeReportIfRequested(metrics, frameCount, logger)
	displaySpatialPartitionStats(world)

	logger.Info("performance test complete!")
}

// setupLogger initializes the test logger with appropriate level.
func setupLogger() *logrus.Logger {
	logger := logging.TestUtilityLogger("perftest")
	if *verbose {
		logger.SetLevel(logrus.DebugLevel)
	}
	return logger
}

// configureTestMode adjusts settings for validation mode if requested.
func configureTestMode(logger *logrus.Logger) {
	if *validate2k {
		logger.Info("running validation test for README performance claim (2000 entities)")
		*entityCount = 2000
		*duration = 30
		*targetFPS = 106.0
	}

	logger.WithFields(logrus.Fields{
		"entities": *entityCount,
		"duration": *duration,
	}).Info("performance test starting")

	if *targetFPS != 60.0 {
		logger.WithField("targetFPS", *targetFPS).Info("custom target FPS")
	}
}

// initializeWorld creates the world and adds required systems.
func initializeWorld(logger *logrus.Logger) (*engine.World, *engine.PerformanceMonitor) {
	world := engine.NewWorld()
	monitor := engine.NewPerformanceMonitor(world)

	spatialSystem := engine.NewSpatialPartitionSystem(10000, 10000)
	world.AddSystem(spatialSystem)
	world.AddSystem(&engine.MovementSystem{})
	world.AddSystem(&engine.CollisionSystem{})

	logger.Info("systems initialized: Movement, Collision, Spatial Partitioning")
	return world, monitor
}

// spawnTestEntities creates and configures test entities in the world.
func spawnTestEntities(world *engine.World, logger *logrus.Logger) {
	logger.WithField("entityCount", *entityCount).Info("spawning entities...")
	startSpawn := time.Now()

	for i := 0; i < *entityCount; i++ {
		createTestEntity(world, i)
	}

	world.Update(0)
	spawnDuration := time.Since(startSpawn)

	logger.WithFields(logrus.Fields{
		"entityCount": *entityCount,
		"duration":    float64(spawnDuration.Microseconds()) / 1000.0,
	}).Info("entities spawned")
}

// createTestEntity creates a single test entity with position, velocity, and optional collider.
func createTestEntity(world *engine.World, index int) {
	entity := world.CreateEntity()

	x := float64(index%100) * 100.0
	y := float64(index/100) * 100.0

	entity.AddComponent(&engine.PositionComponent{X: x, Y: y})
	entity.AddComponent(&engine.VelocityComponent{
		VX: float64((index%3)-1) * 50.0,
		VY: float64((index%5)-2) * 50.0,
	})

	if index%10 == 0 {
		entity.AddComponent(&engine.ColliderComponent{
			Width:  16,
			Height: 16,
			Solid:  true,
			Layer:  1,
		})
	}
}

// runPerformanceTest executes the simulation loop and returns metrics.
func runPerformanceTest(world *engine.World, monitor *engine.PerformanceMonitor, logger *logrus.Logger) (*engine.PerformanceMetrics, int) {
	logger.Info("starting performance test")
	logger.WithFields(logrus.Fields{
		"targetFPS":  *targetFPS,
		"msPerFrame": 1000.0 / *targetFPS,
	}).Info("performance targets")

	frameDuration := time.Second / time.Duration(*targetFPS)
	endTime := time.Now().Add(time.Duration(*duration) * time.Second)
	frameCount := 0

	ticker := time.NewTicker(frameDuration)
	defer ticker.Stop()

	lastReport := time.Now()
	reportInterval := 1 * time.Second
	spatialSystem := getSpatialSystem(world)

	for time.Now().Before(endTime) {
		<-ticker.C
		monitor.Update(frameDuration.Seconds())
		frameCount++

		if time.Since(lastReport) >= reportInterval {
			reportProgress(monitor, spatialSystem)
			lastReport = time.Now()
		}
	}

	logger.Info("performance test complete")
	return monitor.GetMetrics(), frameCount
}

// getSpatialSystem retrieves the spatial partition system from world.
func getSpatialSystem(world *engine.World) *engine.SpatialPartitionSystem {
	for _, sys := range world.GetSystems() {
		if spatial, ok := sys.(*engine.SpatialPartitionSystem); ok {
			return spatial
		}
	}
	return nil
}

// reportProgress displays current performance metrics during test.
func reportProgress(monitor *engine.PerformanceMonitor, spatialSystem *engine.SpatialPartitionSystem) {
	metrics := monitor.GetMetrics()

	if *verbose {
		fmt.Println(metrics.DetailedString())
		if spatialSystem != nil {
			stats := spatialSystem.GetStatistics()
			fmt.Printf("Spatial Partition: %d entities, %d queries\n",
				stats["entity_count"], stats["query_count"])
		}
	} else {
		fmt.Println(metrics.String())
	}
}

// displayResults shows final performance statistics.
func displayResults(metrics *engine.PerformanceMetrics, frameCount int, logger *logrus.Logger) {
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

	printFinalStatistics(metrics, frameCount)
	printSystemBreakdown(metrics)
	printPerformanceTarget(metrics)
}

// printFinalStatistics outputs final performance numbers.
func printFinalStatistics(metrics *engine.PerformanceMetrics, frameCount int) {
	fmt.Printf("\nFinal Statistics:\n")
	fmt.Printf("  Total Frames: %d\n", frameCount)
	fmt.Printf("  Average FPS: %.2f\n", metrics.FPS)
	fmt.Printf("  Average Frame Time: %.2fms\n", float64(metrics.AverageFrameTime.Microseconds())/1000.0)
	fmt.Printf("  Min Frame Time: %.2fms\n", float64(metrics.MinFrameTime.Microseconds())/1000.0)
	fmt.Printf("  Max Frame Time: %.2fms\n", float64(metrics.MaxFrameTime.Microseconds())/1000.0)
	fmt.Printf("  Average Update Time: %.2fms\n", float64(metrics.AverageUpdateTime.Microseconds())/1000.0)
	fmt.Printf("  Entity Count: %d (%d active)\n", metrics.EntityCount, metrics.ActiveEntityCount)
}

// printSystemBreakdown displays time spent in each system.
func printSystemBreakdown(metrics *engine.PerformanceMetrics) {
	fmt.Printf("\nSystem Breakdown:\n")
	percentages := metrics.GetFrameTimePercent()
	for name, percent := range percentages {
		fmt.Printf("  %s: %.2f%%\n", name, percent)
	}
}

// printPerformanceTarget checks and displays whether target FPS was met.
func printPerformanceTarget(metrics *engine.PerformanceMetrics) {
	targetMet := metrics.FPS >= *targetFPS
	fmt.Printf("\nPerformance Target (%.0f FPS): ", *targetFPS)
	if targetMet {
		fmt.Printf("✅ MET (%.2f FPS)\n", metrics.FPS)
	} else {
		fmt.Printf("❌ NOT MET (%.2f FPS, shortfall: %.2f FPS)\n", metrics.FPS, *targetFPS-metrics.FPS)
	}
}

// writeReportIfRequested writes performance report to file if output path specified.
func writeReportIfRequested(metrics *engine.PerformanceMetrics, frameCount int, logger *logrus.Logger) {
	if *outputReport == "" {
		return
	}

	targetMet := metrics.FPS >= *targetFPS
	reportContent := formatReport(metrics, frameCount, targetMet)

	if err := os.WriteFile(*outputReport, []byte(reportContent), 0o644); err != nil {
		logger.WithError(err).Error("failed to write report")
	} else {
		logger.WithField("path", *outputReport).Info("performance report written")
	}
}

// formatReport creates the performance report content string.
func formatReport(metrics *engine.PerformanceMetrics, frameCount int, targetMet bool) string {
	report := fmt.Sprintf(`Performance Test Report
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

	percentages := metrics.GetFrameTimePercent()
	for name, percent := range percentages {
		report += fmt.Sprintf("  %s: %.2f%%\n", name, percent)
	}

	return report
}

// displaySpatialPartitionStats shows spatial partition statistics and query performance.
func displaySpatialPartitionStats(world *engine.World) {
	spatialSystem := getSpatialSystem(world)
	if spatialSystem == nil {
		return
	}

	stats := spatialSystem.GetStatistics()
	fmt.Printf("\nSpatial Partition Statistics:\n")
	fmt.Printf("  Entities Tracked: %d\n", stats["entity_count"])
	fmt.Printf("  Total Queries: %d\n", stats["query_count"])
	fmt.Printf("  Last Rebuild Time: %.2fms\n", stats["last_rebuild_time"].(float64)*1000.0)

	benchmarkSpatialQueries(spatialSystem)
}

// benchmarkSpatialQueries measures spatial query performance.
func benchmarkSpatialQueries(spatialSystem *engine.SpatialPartitionSystem) {
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
}
