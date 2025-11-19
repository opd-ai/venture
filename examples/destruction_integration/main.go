package main

import (
	"fmt"
	"log"

	"github.com/opd-ai/venture/pkg/engine/physics/destruction"
	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/opd-ai/venture/pkg/procgen/building"
)

// This example demonstrates integration of the building generator
// with the structural destruction system.
func main() {
	// BUG FIX: Static Analysis - Redundant newline in fmt.Println
	// Resolution: Removed \n from fmt.Println as it already adds a newline
	fmt.Println("=== Building Generation & Destruction Integration ===")

	seed := int64(12345)
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      1,
		GenreID:    "fantasy",
		Custom: map[string]interface{}{
			"buildingType": building.TypeManor,
			"floors":       3,
		},
	}

	gen := building.NewGenerator()

	fmt.Println("Generating procedural building...")
	result, err := gen.Generate(seed, params)
	if err != nil {
		log.Fatalf("Building generation failed: %v", err)
	}

	bldg := result.(*building.Building)
	fmt.Printf("Generated: %s (%s style)\n", bldg.Type, bldg.Style)
	fmt.Printf("Dimensions: %dx%d tiles, %d floors\n", bldg.Width, bldg.Height, bldg.Floors)
	fmt.Printf("Rooms: %d, Doors: %d, Windows: %d\n",
		len(bldg.Rooms), len(bldg.Doors), len(bldg.Windows))
	fmt.Printf("Roof Type: %s\n\n", bldg.RoofType)

	material := destruction.MaterialStone
	if bldg.Style == building.StyleElven {
		material = destruction.MaterialWood
	} else if bldg.Style == building.StyleCrystalline {
		material = destruction.MaterialGlass
	}

	config := destruction.DefaultConfig()
	config.MaxDebrisParticles = 1000
	config.DebrisLifetime = 15.0
	destructionSys := destruction.NewSystem(config)

	buildingID := "manor_001"
	fmt.Printf("Registering building with destruction system (material: %s)...\n", material)
	destructionSys.RegisterBuilding(buildingID, bldg.Width, bldg.Height, bldg.Floors, material)

	integrity, _ := destructionSys.GetIntegrity(buildingID)
	fmt.Printf("Initial integrity: %s (%.0f health points)\n",
		integrity.State, integrity.TotalHealth)
	fmt.Printf("Structural supports: %d\n\n", len(integrity.Supports))

	damageX := bldg.Width / 2
	damageY := bldg.Height / 2
	damageAmount := 0.4
	damageRadius := 6.0

	fmt.Printf("Simulating explosion at center (%d, %d)...\n", damageX, damageY)
	fmt.Printf("Damage: amount=%.1f, radius=%.1f tiles\n\n", damageAmount, damageRadius)

	err = destructionSys.ApplyDamage(buildingID, damageX, damageY, 0, damageAmount, damageRadius)
	if err != nil {
		log.Fatalf("ApplyDamage failed: %v", err)
	}

	integrity, _ = destructionSys.GetIntegrity(buildingID)
	fmt.Printf("After explosion:\n")
	fmt.Printf("  State: %s\n", integrity.State)
	fmt.Printf("  Health: %.1f%%\n", integrity.CurrentHealth*100)
	fmt.Printf("  Collapse Risk: %.1f%%\n", integrity.CollapseRisk*100)
	fmt.Printf("  Damaged Areas: %d\n\n", len(integrity.DamagedAreas))

	fmt.Println("Simulating 10 seconds of damage propagation...")
	for i := 0; i < 100; i++ {
		destructionSys.Update(0.1)
	}

	integrity, _ = destructionSys.GetIntegrity(buildingID)
	fmt.Printf("\nAfter propagation:\n")
	fmt.Printf("  State: %s\n", integrity.State)
	fmt.Printf("  Health: %.1f%%\n", integrity.CurrentHealth*100)
	fmt.Printf("  Collapse Risk: %.1f%%\n\n", integrity.CollapseRisk*100)

	if integrity.State == destruction.IntegrityCollapsed {
		fmt.Println("!!! BUILDING COLLAPSED !!!")
		fmt.Printf("Debris generated: %d particles\n", destructionSys.GetDebrisCount())
		fmt.Printf("Falling objects: %d\n", destructionSys.GetFallingObjectCount())
	} else {
		fmt.Println("Building structure is holding...")

		damagedSupports := 0
		for _, support := range integrity.Supports {
			if support.Health < 0.5 {
				damagedSupports++
			}
		}
		fmt.Printf("Damaged supports: %d/%d\n", damagedSupports, len(integrity.Supports))
	}

	fmt.Println("\n=== Integration Complete ===")
	fmt.Println("\nKey Integration Points:")
	fmt.Println("1. Building dimensions (width, height, floors) → RegisterBuilding()")
	fmt.Println("2. Building style/genre → Material selection")
	fmt.Println("3. Room positions → Targeted damage application")
	fmt.Println("4. Support points → Structural integrity calculation")
	fmt.Println("5. Collapse state → Trigger building removal/replacement")
	fmt.Println("\nPotential Enhancements:")
	fmt.Println("- Map room types to material types (workshop=metal, kitchen=brick)")
	fmt.Println("- Use window/door positions as weak points")
	fmt.Println("- Generate debris matching building materials")
	fmt.Println("- Apply damage based on room inhabitants or contents")
	fmt.Println("- Visual damage indicators (cracks, missing walls)")
}
