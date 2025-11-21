package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/opd-ai/venture/pkg/engine/physics/destruction"
)

func main() {
	buildingType := flag.String("type", "house", "Building type: house, manor, tower")
	material := flag.String("material", "stone", "Material: wood, stone, metal, glass, concrete, brick")
	damageX := flag.Int("x", 8, "Damage X position")
	damageY := flag.Int("y", 8, "Damage Y position")
	damageAmount := flag.Float64("amount", 0.5, "Damage amount (0.0-1.0)")
	damageRadius := flag.Float64("radius", 5.0, "Damage radius (tiles)")
	simulate := flag.Bool("simulate", false, "Run simulation")
	verbose := flag.Bool("verbose", false, "Verbose output")

	flag.Parse()

	mat := parseMaterialType(*material)
	width, height, floors := getBuildingDimensions(*buildingType)

	config := destruction.DefaultConfig()
	sys := destruction.NewSystem(config)

	buildingID := "test_building"
	sys.RegisterBuilding(buildingID, width, height, floors, mat)

	fmt.Printf("=== Building Destruction Test ===\n")
	fmt.Printf("Type: %s (%dx%d, %d floors)\n", *buildingType, width, height, floors)
	fmt.Printf("Material: %s\n", *material)
	fmt.Printf("\n")

	integrity, _ := sys.GetIntegrity(buildingID)
	fmt.Printf("Initial State:\n")
	printIntegrity(integrity, *verbose)

	fmt.Printf("\nApplying damage at (%d, %d) with amount=%.2f, radius=%.1f\n",
		*damageX, *damageY, *damageAmount, *damageRadius)

	err := sys.ApplyDamage(buildingID, *damageX, *damageY, 0, *damageAmount, *damageRadius)
	if err != nil {
		log.Fatalf("ApplyDamage failed: %v", err)
	}

	integrity, _ = sys.GetIntegrity(buildingID)
	fmt.Printf("\nAfter Damage:\n")
	printIntegrity(integrity, *verbose)

	if *simulate {
		runCollapseSimulation(sys, buildingID)
	}

	runPhysicsSimulation(sys, mat)

	fmt.Printf("\n=== Test Complete ===\n")
}

// parseMaterialType converts material string to MaterialType enum.
func parseMaterialType(material string) destruction.MaterialType {
	materialMap := map[string]destruction.MaterialType{
		"wood":     destruction.MaterialWood,
		"stone":    destruction.MaterialStone,
		"metal":    destruction.MaterialMetal,
		"glass":    destruction.MaterialGlass,
		"concrete": destruction.MaterialConcrete,
		"brick":    destruction.MaterialBrick,
	}

	mat, ok := materialMap[material]
	if !ok {
		log.Fatalf("Unknown material: %s", material)
	}
	return mat
}

// getBuildingDimensions returns width, height, and floors for a building type.
func getBuildingDimensions(buildingType string) (int, int, int) {
	switch buildingType {
	case "house":
		return 16, 16, 2
	case "manor":
		return 24, 24, 3
	case "tower":
		return 8, 8, 5
	default:
		return 16, 16, 2
	}
}

// runCollapseSimulation executes the building collapse simulation.
func runCollapseSimulation(sys *destruction.System, buildingID string) {
	fmt.Printf("\n=== Running Simulation ===\n")

	for i := 0; i < 100; i++ {
		sys.Update(0.1)

		if (i+1)%20 == 0 {
			integrity, _ := sys.GetIntegrity(buildingID)
			fmt.Printf("\nStep %d (%.1fs):\n", i+1, float64(i+1)*0.1)
			printIntegrity(integrity, false)

			if integrity.State == destruction.IntegrityCollapsed {
				fmt.Printf("\n!!! BUILDING COLLAPSED !!!\n")
				fmt.Printf("Debris particles: %d\n", sys.GetDebrisCount())
				fmt.Printf("Falling objects: %d\n", sys.GetFallingObjectCount())
				break
			}
		}
	}
}

// runPhysicsSimulation demonstrates falling object physics.
func runPhysicsSimulation(sys *destruction.System, mat destruction.MaterialType) {
	fmt.Printf("\n=== Physics Simulation ===\n")
	fmt.Printf("Spawning falling object...\n")
	sys.SpawnFallingObject(100, 100, 200, mat, 16, 16)

	for i := 0; i < 50; i++ {
		sys.Update(0.016)

		objs := sys.GetFallingObjects()
		if len(objs) > 0 {
			obj := objs[0]
			if (i+1)%10 == 0 {
				fmt.Printf("Step %d: Z=%.1f, VelZ=%.1f, Bounces=%d, Grounded=%v\n",
					i+1, obj.Z, obj.VelZ, obj.Bounces, obj.IsGrounded())
			}

			if obj.IsGrounded() {
				fmt.Printf("Object settled after %d steps\n", i+1)
				break
			}
		}
	}
}

func printIntegrity(integrity *destruction.StructuralIntegrity, verbose bool) {
	fmt.Printf("  State: %s\n", integrity.State)
	fmt.Printf("  Health: %.1f%% (%.0f/%.0f)\n",
		integrity.CurrentHealth*100,
		integrity.CurrentHealth*integrity.TotalHealth,
		integrity.TotalHealth)
	fmt.Printf("  Collapse Risk: %.1f%%\n", integrity.CollapseRisk*100)
	fmt.Printf("  Supports: %d total\n", len(integrity.Supports))
	fmt.Printf("  Damaged Areas: %d\n", len(integrity.DamagedAreas))

	if verbose {
		loadBearing := 0
		damaged := 0
		for _, support := range integrity.Supports {
			if support.LoadBearing {
				loadBearing++
				if support.Health < 0.5 {
					damaged++
				}
			}
		}
		fmt.Printf("  Load-Bearing Supports: %d (%d damaged)\n", loadBearing, damaged)

		supportTypes := make(map[destruction.SupportType]int)
		for _, support := range integrity.Supports {
			supportTypes[support.Type]++
		}
		fmt.Printf("  Support Distribution:\n")
		for typ, count := range supportTypes {
			fmt.Printf("    %s: %d\n", typ, count)
		}
	}
}
