package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/opd-ai/venture/pkg/engine/physics/fluids"
)

func main() {
	// Command-line flags
	gridWidth := flag.Int("width", 100, "Grid width")
	gridHeight := flag.Int("height", 100, "Grid height")
	duration := flag.Int("duration", 10, "Simulation duration in seconds")
	fluidType := flag.String("fluid", "water", "Fluid type (water, lava, oil, acid, poison)")
	testBuoyancy := flag.Bool("buoyancy", false, "Test buoyancy calculations")
	testSwimming := flag.Bool("swimming", false, "Test swimming mechanics")
	testFlooding := flag.Bool("flooding", false, "Test flooding system")
	flag.Parse()

	fmt.Println("=== Venture Fluid Dynamics Test ===")
	fmt.Println()

	if *testBuoyancy {
		runBuoyancyTest()
		return
	}

	if *testSwimming {
		runSwimmingTest()
		return
	}

	if *testFlooding {
		runFloodingTest(*gridWidth, *gridHeight)
		return
	}

	// Default: run fluid simulation
	runFluidSimulation(*gridWidth, *gridHeight, *duration, *fluidType)
}

func runFluidSimulation(width, height, duration int, fluidTypeName string) {
	fmt.Printf("Running fluid simulation: %dx%d grid, %d seconds\n", width, height, duration)
	fmt.Printf("Fluid type: %s\n\n", fluidTypeName)

	// Parse fluid type
	fluidType := parseFluidType(fluidTypeName)
	props := fluids.GetFluidProperties(fluidType)

	fmt.Printf("Fluid properties:\n")
	fmt.Printf("  Viscosity: %.2f\n", props.Viscosity)
	fmt.Printf("  Density: %.1f kg/m³\n", props.Density)
	fmt.Printf("  Flow rate: %.2f\n", props.FlowRate)
	fmt.Printf("  Damage: %.1f/sec\n", props.Damage)
	fmt.Println()

	// Create simulator
	config := fluids.DefaultSimulationConfig()
	config.GridWidth = width
	config.GridHeight = height
	simulator := fluids.NewSimulator(config)

	// Add fluid sources
	simulator.AddFluid(width/4, 10, 1.0, fluidType)
	simulator.AddFluid(width/2, 10, 1.0, fluidType)
	simulator.AddFluid(3*width/4, 10, 1.0, fluidType)

	fmt.Println("Starting simulation...")
	start := time.Now()

	// Run simulation
	updateRate := 30.0 // 30 FPS
	deltaTime := 1.0 / updateRate
	updates := 0
	maxUpdates := int(float64(duration) * updateRate)

	ticker := time.NewTicker(time.Duration(float64(time.Second) / updateRate))
	defer ticker.Stop()

	for range ticker.C {
		simulator.Update(deltaTime)
		updates++

		// Print progress every second
		if updates%(int(updateRate)) == 0 {
			elapsed := time.Since(start).Seconds()
			fmt.Printf("  Time: %.1fs | Updates: %d | Grid updates/sec: %.1f\n",
				elapsed, updates, float64(updates)/elapsed)
		}

		if updates >= maxUpdates {
			break
		}
	}

	elapsed := time.Since(start)
	fmt.Println()
	fmt.Printf("Simulation complete!\n")
	fmt.Printf("Total time: %v\n", elapsed)
	fmt.Printf("Total updates: %d\n", updates)
	fmt.Printf("Average update time: %.3fms\n", elapsed.Seconds()*1000/float64(updates))
	fmt.Printf("Target: <5ms for 100x100 grid\n")
	fmt.Println()

	// Sample fluid at various points
	fmt.Println("Final fluid distribution (sample points):")
	samplePoints := []struct{ x, y int }{
		{width / 4, height / 2},
		{width / 2, height / 2},
		{3 * width / 4, height / 2},
		{width / 2, height - 10},
	}

	for _, pt := range samplePoints {
		amount, _, err := simulator.GetFluidAt(pt.x, pt.y)
		if err == nil {
			fmt.Printf("  Position (%d, %d): %.2f fluid\n", pt.x, pt.y, amount)
		}
	}
}

func runBuoyancyTest() {
	fmt.Println("Testing buoyancy calculations")
	fmt.Println()

	calc := fluids.NewBuoyancyCalculator(9.81)

	tests := []struct {
		name   string
		mass   float64
		volume float64
		fluid  fluids.FluidType
	}{
		{"Wood in water", 500.0, 2.0, fluids.FluidWater},
		{"Steel in water", 7850.0, 1.0, fluids.FluidWater},
		{"Cork in water", 240.0, 1.0, fluids.FluidWater},
		{"Wood in lava", 500.0, 2.0, fluids.FluidLava},
		{"Steel in lava", 7850.0, 1.0, fluids.FluidLava},
	}

	for _, tt := range tests {
		fmt.Printf("%s:\n", tt.name)

		component := &fluids.BuoyancyComponent{
			Mass:   tt.mass,
			Volume: tt.volume,
		}
		fluids.UpdateDensity(component)

		calc.CalculateBuoyancy(component, 1.0, tt.fluid)

		weight := tt.mass * 9.81
		netForce := calc.GetNetForce(component)

		fmt.Printf("  Density: %.1f kg/m³\n", component.Density)
		fmt.Printf("  Weight: %.1f N\n", weight)
		fmt.Printf("  Buoyant force: %.1f N\n", component.BuoyantForce)
		fmt.Printf("  Net force: %.1f N (%s)\n", netForce, map[bool]string{true: "upward", false: "downward"}[netForce > 0])
		fmt.Printf("  Floats: %v\n", component.Buoyant)
		fmt.Println()
	}

	// Benchmark
	fmt.Println("Performance benchmark:")
	component := &fluids.BuoyancyComponent{
		Mass:   500.0,
		Volume: 2.0,
	}

	iterations := 100000
	start := time.Now()
	for i := 0; i < iterations; i++ {
		calc.CalculateBuoyancy(component, 1.0, fluids.FluidWater)
	}
	elapsed := time.Since(start)

	fmt.Printf("  %d calculations in %v\n", iterations, elapsed)
	fmt.Printf("  Average: %.3fµs per calculation\n", elapsed.Seconds()*1e6/float64(iterations))
	fmt.Printf("  Target: <100µs per entity\n")
}

func runSwimmingTest() {
	fmt.Println("Testing swimming mechanics")
	fmt.Println()

	mgr := fluids.NewSwimmingManager(9.81)

	swimming := &fluids.SwimmingComponent{
		Stamina:        100.0,
		MaxStamina:     100.0,
		StaminaDrain:   10.0,
		StaminaRegen:   20.0,
		SwimSpeed:      0.5,
		DrowningDamage: 10.0,
	}

	buoyancy := &fluids.BuoyancyComponent{
		Mass:   70.0, // 70 kg person
		Volume: 0.07, // ~70 liters
	}

	fmt.Println("Scenario: Person swimming in water")
	fmt.Println()

	deltaTime := 1.0 // 1 second updates

	// Swim for 10 seconds
	fmt.Println("Swimming (10 seconds):")
	for i := 0; i < 10; i++ {
		mgr.UpdateSwimming(swimming, buoyancy, 1.0, deltaTime)
		speedMult := mgr.GetSwimSpeedMultiplier(swimming)
		damage := mgr.GetDrowningDamage(swimming)

		fmt.Printf("  Second %d: Stamina: %.1f%%, Speed: %.1f%%, Drowning: %v, Damage: %.1f/s\n",
			i+1, swimming.Stamina, speedMult*100, swimming.Drowning, damage)
	}

	fmt.Println()

	// Tread water for 5 seconds
	fmt.Println("Treading water (5 seconds):")
	swimming.TreadingWater = true
	for i := 0; i < 5; i++ {
		mgr.UpdateSwimming(swimming, buoyancy, 1.0, deltaTime)
		fmt.Printf("  Second %d: Stamina: %.1f%%\n", i+1, swimming.Stamina)
	}

	fmt.Println()

	// Return to land
	fmt.Println("On land (5 seconds, regenerating):")
	for i := 0; i < 5; i++ {
		mgr.UpdateSwimming(swimming, buoyancy, 0.0, deltaTime)
		fmt.Printf("  Second %d: Stamina: %.1f%%\n", i+1, swimming.Stamina)
	}
}

func runFloodingTest(width, height int) {
	fmt.Printf("Testing flooding system: %dx%d grid\n", width, height)
	fmt.Println()

	// Create simulator
	config := fluids.DefaultSimulationConfig()
	config.GridWidth = width
	config.GridHeight = height
	simulator := fluids.NewSimulator(config)

	// Create flooding manager
	mgr := fluids.NewFloodingManager(simulator)

	// Create flooding component
	flooding := &fluids.FloodingComponent{
		AreaID:        "test_room",
		FloodLevel:    0.0,
		FloodRate:     0.1,
		MaxFloodLevel: 1.0,
	}

	// Add flood sources
	mgr.AddFloodSource(flooding, width/4, height/4, 0.05)
	mgr.AddFloodSource(flooding, 3*width/4, height/4, 0.05)

	fmt.Printf("Flood sources: %d\n", len(flooding.Sources))
	fmt.Printf("Total inflow: 0.1 units/sec\n")
	fmt.Println()

	// Run flooding simulation
	fmt.Println("Flooding progress:")
	deltaTime := 1.0
	for i := 0; i < 10; i++ {
		mgr.UpdateFlooding(flooding, deltaTime)
		percentage := mgr.GetFloodPercentage(flooding)
		fullyFlooded := mgr.IsFullyFlooded(flooding)

		fmt.Printf("  Second %d: %.1f%% flooded, Fully flooded: %v\n",
			i+1, percentage*100, fullyFlooded)

		if fullyFlooded {
			fmt.Println("  Area completely flooded!")
			break
		}
	}
}

func parseFluidType(name string) fluids.FluidType {
	switch name {
	case "water":
		return fluids.FluidWater
	case "lava":
		return fluids.FluidLava
	case "oil":
		return fluids.FluidOil
	case "acid":
		return fluids.FluidAcid
	case "poison":
		return fluids.FluidPoison
	default:
		fmt.Printf("Unknown fluid type '%s', using water\n", name)
		return fluids.FluidWater
	}
}
