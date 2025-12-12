package main

import (
	"flag"
	"fmt"
	"math/rand"
	"time"

	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/logging"
	"github.com/sirupsen/logrus"
)

func main() {
	// Parse command-line flags
	count := flag.Int("count", 10, "Number of entities to simulate")
	duration := flag.Float64("duration", 5.0, "Simulation duration in seconds")
	verbose := flag.Bool("verbose", false, "Show detailed output")
	seed := flag.Int64("seed", time.Now().UnixNano(), "Random seed")
	flag.Parse()

	// Initialize logger
	logger := logging.TestUtilityLogger("movementtest")
	if *verbose {
		logger.SetLevel(logrus.DebugLevel)
	}

	logger.WithFields(logrus.Fields{
		"entities": *count,
		"duration": *duration,
		"seed":     *seed,
	}).Info("Movement Test Tool started")

	fmt.Println("=== Movement and Collision System Demo ===")
	fmt.Printf("Entities: %d, Duration: %.1fs, Seed: %d\n\n", *count, *duration, *seed)

	// Create random number generator
	rng := rand.New(rand.NewSource(*seed))

	// Create world and systems
	world := engine.NewWorld()
	movementSystem := engine.NewMovementSystem(200.0)  // Max speed of 200 units/s
	collisionSystem := engine.NewCollisionSystem(64.0) // 64 unit cell size

	// Track collisions
	collisionCount := 0
	collisionSystem.SetCollisionCallback(func(e1, e2 *engine.Entity) {
		collisionCount++
		if *verbose {
			x1, y1, _ := engine.GetPosition(e1)
			x2, y2, _ := engine.GetPosition(e2)
			fmt.Printf("  Collision: Entity %d at (%.1f, %.1f) <-> Entity %d at (%.1f, %.1f)\n",
				e1.ID, x1, y1, e2.ID, x2, y2)
			logger.WithFields(logrus.Fields{
				"entity1": e1.ID,
				"entity2": e2.ID,
			}).Debug("collision detected")
		}
	})

	world.AddSystem(movementSystem)
	world.AddSystem(collisionSystem)

	logger.Info("systems initialized")

	// Create entities with random positions and velocities
	fmt.Println("Creating entities...")
	for i := 0; i < *count; i++ {
		entity := world.CreateEntity()

		// Random position within 800x600 area
		x := rng.Float64() * 800
		y := rng.Float64() * 600

		// Random velocity
		vx := (rng.Float64()*2 - 1) * 100 // -100 to +100
		vy := (rng.Float64()*2 - 1) * 100

		// Random size
		size := 10 + rng.Float64()*30 // 10-40 units

		entity.AddComponent(&engine.PositionComponent{X: x, Y: y})
		entity.AddComponent(&engine.VelocityComponent{VX: vx, VY: vy})
		entity.AddComponent(&engine.ColliderComponent{
			Width:  size,
			Height: size,
			Solid:  true,
		})
		entity.AddComponent(&engine.BoundsComponent{
			MinX: 0,
			MinY: 0,
			MaxX: 800,
			MaxY: 600,
			Wrap: false,
		})

		if *verbose {
			fmt.Printf("  Entity %d: pos=(%.1f, %.1f), vel=(%.1f, %.1f), size=%.1f\n",
				entity.ID, x, y, vx, vy, size)
		}
	}

	logger.WithField("count", *count).Info("entities created")

	world.Update(0) // Process entity additions

	fmt.Printf("\nSimulating movement and collisions...\n\n")

	// Simulation loop
	const targetFPS = 60
	const deltaTime = 1.0 / targetFPS
	steps := int(*duration / deltaTime)

	startTime := time.Now()

	for step := 0; step < steps; step++ {
		world.Update(deltaTime)

		// Print progress every second
		if *verbose && step%(targetFPS) == 0 {
			elapsed := float64(step) * deltaTime
			fmt.Printf("Time: %.1fs, Collisions so far: %d\n", elapsed, collisionCount)
		}
	}

	simulationTime := time.Since(startTime)

	// Display final statistics
	fmt.Println("=== Simulation Complete ===")
	fmt.Printf("Total time simulated: %.1fs\n", *duration)
	fmt.Printf("Actual computation time: %v\n", simulationTime)
	fmt.Printf("Total collisions detected: %d\n", collisionCount)
	fmt.Printf("Average collisions/second: %.1f\n", float64(collisionCount)/(*duration))
	fmt.Printf("Simulation speed: %.1fx real-time\n\n", *duration/simulationTime.Seconds())

	logger.WithFields(logrus.Fields{
		"collisions":      collisionCount,
		"computationTime": simulationTime.String(),
		"speedRatio":      *duration / simulationTime.Seconds(),
	}).Info("simulation completed")

	// Show final entity positions
	if *verbose {
		fmt.Println("Final entity positions:")
		entities := world.GetEntities()
		for _, entity := range entities {
			x, y, ok := engine.GetPosition(entity)
			if ok {
				fmt.Printf("  Entity %d: (%.1f, %.1f)\n", entity.ID, x, y)
			}
		}
	}

	// Performance analysis
	fmt.Println("=== Performance Analysis ===")
	updatesPerSecond := float64(steps) / simulationTime.Seconds()
	fmt.Printf("World updates/second: %.0f\n", updatesPerSecond)
	fmt.Printf("Entity updates/second: %.0f\n", updatesPerSecond*float64(*count))
	fmt.Printf("Target FPS achieved: %v\n", updatesPerSecond >= targetFPS*0.95)

	// Benchmark example
	fmt.Println("\n=== Integration Example ===")
	fmt.Println("This demo shows the movement and collision systems working together:")
	fmt.Println("1. MovementSystem updates entity positions based on velocity")
	fmt.Println("2. Entities bounce off world boundaries")
	fmt.Println("3. CollisionSystem detects overlaps using spatial partitioning")
	fmt.Println("4. Solid colliders push entities apart to resolve collisions")
	fmt.Println("\nUsage in your game:")
	fmt.Println("  world := engine.NewWorld()")
	fmt.Println("  world.AddSystem(engine.NewMovementSystem(maxSpeed))")
	fmt.Println("  world.AddSystem(engine.NewCollisionSystem(cellSize))")
	fmt.Println("  // In game loop:")
	fmt.Println("  world.Update(deltaTime)")
}
