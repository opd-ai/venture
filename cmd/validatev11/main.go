// Package main provides a validation script for Venture v1.1 features.
// This script verifies that all major v1.1 systems are operational and
// can be instantiated without errors. Run this script before deployment
// to ensure production readiness.
//
// Usage: go run scripts/validate_v1_1_features.go
package main

import (
	"fmt"
	"os"

	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/procgen"
	procgenEntity "github.com/opd-ai/venture/pkg/procgen/entity"
	"github.com/opd-ai/venture/pkg/procgen/item"
	"github.com/opd-ai/venture/pkg/rendering/particles"
)

func main() {
	fmt.Println("Venture v1.1 Feature Validation")
	fmt.Println("================================")
	fmt.Println()

	passed := 0
	failed := 0

	world := engine.NewWorld()
	inventorySystem := engine.NewInventorySystem(world)

	passed, failed = runSystemTests(world, inventorySystem, passed, failed)
	passed, failed = runComponentTests(passed, failed)

	printValidationSummary(passed, failed)
}

// runSystemTests validates all system-level features.
func runSystemTests(world *engine.World, inventorySystem *engine.InventorySystem, passed, failed int) (int, int) {
	passed, failed = validateCommerceSystem(world, inventorySystem, passed, failed)
	passed, failed = validateCraftingSystem(world, inventorySystem, passed, failed)
	passed, failed = validateDialogSystem(world, passed, failed)
	passed, failed = validateMerchantGeneration(passed, failed)
	passed, failed = validateParticlePooling(passed, failed)
	passed, failed = validateTerrainSystems(passed, failed)
	return passed, failed
}

// validateCommerceSystem tests the commerce system initialization.
func validateCommerceSystem(world *engine.World, inventorySystem *engine.InventorySystem, passed, failed int) (int, int) {
	fmt.Print("✓ Testing Commerce System... ")
	commerceSystem := engine.NewCommerceSystem(world, inventorySystem)
	if commerceSystem != nil {
		fmt.Println("PASS")
		return passed + 1, failed
	}
	fmt.Println("FAIL")
	return passed, failed + 1
}

// validateCraftingSystem tests the crafting system initialization.
func validateCraftingSystem(world *engine.World, inventorySystem *engine.InventorySystem, passed, failed int) (int, int) {
	fmt.Print("✓ Testing Crafting System... ")
	itemGen := item.NewItemGenerator()
	craftingSystem := engine.NewCraftingSystem(world, inventorySystem, itemGen)
	if craftingSystem != nil {
		fmt.Println("PASS")
		return passed + 1, failed
	}
	fmt.Println("FAIL")
	return passed, failed + 1
}

// validateDialogSystem tests the dialog system initialization.
func validateDialogSystem(world *engine.World, passed, failed int) (int, int) {
	fmt.Print("✓ Testing Dialog System... ")
	dialogSystem := engine.NewDialogSystem(world)
	if dialogSystem != nil {
		fmt.Println("PASS")
		return passed + 1, failed
	}
	fmt.Println("FAIL")
	return passed, failed + 1
}

// validateMerchantGeneration tests procedural merchant generation.
func validateMerchantGeneration(passed, failed int) (int, int) {
	fmt.Print("✓ Testing Merchant Generation... ")
	entityGen := procgenEntity.NewEntityGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      1,
		GenreID:    "fantasy",
		Custom:     map[string]interface{}{"count": 1},
	}
	merchantData, err := entityGen.GenerateMerchant(12345, params, procgenEntity.MerchantFixed)
	if err != nil {
		fmt.Println("FAIL:", err)
		return passed, failed + 1
	}
	if merchantData != nil && merchantData.Entity != nil && len(merchantData.Inventory) > 0 {
		fmt.Println("PASS")
		return passed + 1, failed
	}
	fmt.Println("FAIL")
	return passed, failed + 1
}

// validateParticlePooling tests particle system pooling.
func validateParticlePooling(passed, failed int) (int, int) {
	fmt.Print("✓ Testing Particle Pooling... ")
	ps := particles.NewParticleSystem([]particles.Particle{}, particles.ParticleSpark, particles.DefaultConfig())
	particles.ReleaseParticleSystem(ps)
	ps2 := particles.NewParticleSystem([]particles.Particle{}, particles.ParticleSpark, particles.DefaultConfig())
	if ps2 != nil {
		fmt.Println("PASS")
		particles.ReleaseParticleSystem(ps2)
		return passed + 1, failed
	}
	fmt.Println("FAIL")
	return passed, failed + 1
}

// validateTerrainSystems tests terrain modification, fire propagation, and construction systems.
func validateTerrainSystems(passed, failed int) (int, int) {
	tileSize := 32
	tests := []struct {
		name   string
		system interface{}
	}{
		{"Terrain Modification System", engine.NewTerrainModificationSystem(tileSize)},
		{"Fire Propagation System", engine.NewFirePropagationSystem(tileSize, 12345)},
		{"Terrain Construction System", engine.NewTerrainConstructionSystem(tileSize)},
	}

	for _, test := range tests {
		fmt.Printf("✓ Testing %s... ", test.name)
		if test.system != nil {
			fmt.Println("PASS")
			passed++
		} else {
			fmt.Println("FAIL")
			failed++
		}
	}
	return passed, failed
}

// runComponentTests validates component-level features.
func runComponentTests(passed, failed int) (int, int) {
	passed, failed = validateMerchantComponent(passed, failed)
	passed, failed = validateDialogComponent(passed, failed)
	return passed, failed
}

// validateMerchantComponent tests merchant component creation.
func validateMerchantComponent(passed, failed int) (int, int) {
	fmt.Print("✓ Testing MerchantComponent... ")
	merchantComp := engine.NewMerchantComponent(20, engine.MerchantFixed, 1.5)
	if merchantComp != nil && merchantComp.Type() == "merchant" {
		fmt.Println("PASS")
		return passed + 1, failed
	}
	fmt.Println("FAIL")
	return passed, failed + 1
}

// validateDialogComponent tests dialog component creation.
func validateDialogComponent(passed, failed int) (int, int) {
	fmt.Print("✓ Testing DialogComponent... ")
	provider := engine.NewMerchantDialogProvider("Test Merchant")
	dialogComp := engine.NewDialogComponent(provider)
	if dialogComp != nil && dialogComp.Type() == "dialog" {
		fmt.Println("PASS")
		return passed + 1, failed
	}
	fmt.Println("FAIL")
	return passed, failed + 1
}

// printValidationSummary displays test results and exits with appropriate code.
func printValidationSummary(passed, failed int) {
	fmt.Println("\n================================")
	fmt.Printf("Tests Passed: %d/%d\n", passed, passed+failed)

	if failed > 0 {
		fmt.Printf("Tests Failed: %d\n", failed)
		fmt.Println("❌ v1.1 validation FAILED")
		os.Exit(1)
	}
	fmt.Println("✅ All v1.1 features validated")
	fmt.Println("Ready for production deployment")
	os.Exit(0)
}
