package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/opd-ai/venture/pkg/engine"
)

var (
	entityCount = flag.Int("entities", 1000, "Number of entities to spawn")
	duration    = flag.Int("duration", 10, "Test duration in seconds")
	verbose     = flag.Bool("verbose", false, "Enable verbose logging")
)

func main() {
	flag.Parse()

	log.Printf("Performance Test - Spawning %d entities for %d seconds", *entityCount, *duration)

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

	log.Println("Systems initialized: Movement, Collision, Spatial Partitioning")

	// Spawn entities
	log.Printf("Spawning %d entities...", *entityCount)
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

	log.Printf("Spawned %d entities in %.2fms", *entityCount, float64(spawnDuration.Microseconds())/1000.0)

	// Run simulation
	log.Println("Starting performance test...")
	log.Printf("Target: 60 FPS (16.67ms per frame)")

	targetFPS := 60.0
	frameDuration := time.Second / time.Duration(targetFPS)
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
	log.Println("\n=== Performance Test Complete ===")
	metrics := monitor.GetMetrics()

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
	fmt.Printf("\nPerformance Target (60 FPS): ")
	if metrics.IsPerformanceTarget() {
		fmt.Printf("✅ MET (%.2f FPS)\n", metrics.FPS)
	} else {
		fmt.Printf("❌ NOT MET (%.2f FPS)\n", metrics.FPS)
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

	log.Println("\nPerformance test complete!")
}
