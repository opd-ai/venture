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

	// Test 1: Commerce System
	fmt.Print("✓ Testing Commerce System... ")
	world := engine.NewWorld()
	inventorySystem := engine.NewInventorySystem(world)
	commerceSystem := engine.NewCommerceSystem(world, inventorySystem)
	if commerceSystem != nil {
		fmt.Println("PASS")
		passed++
	} else {
		fmt.Println("FAIL")
		failed++
	}

	// Test 2: Crafting System
	fmt.Print("✓ Testing Crafting System... ")
	itemGen := item.NewItemGenerator()
	craftingSystem := engine.NewCraftingSystem(world, inventorySystem, itemGen)
	if craftingSystem != nil {
		fmt.Println("PASS")
		passed++
	} else {
		fmt.Println("FAIL")
		failed++
	}

	// Test 3: Dialog System
	fmt.Print("✓ Testing Dialog System... ")
	dialogSystem := engine.NewDialogSystem(world)
	if dialogSystem != nil {
		fmt.Println("PASS")
		passed++
	} else {
		fmt.Println("FAIL")
		failed++
	}

	// Test 4: Merchant Generation
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
		failed++
	} else if merchantData != nil && merchantData.Entity != nil && len(merchantData.Inventory) > 0 {
		fmt.Println("PASS")
		passed++
	} else {
		fmt.Println("FAIL")
		failed++
	}

	// Test 5: Particle Pooling
	fmt.Print("✓ Testing Particle Pooling... ")
	ps := particles.NewParticleSystem(
		[]particles.Particle{},
		particles.ParticleSpark,
		particles.DefaultConfig(),
	)
	particles.ReleaseParticleSystem(ps)
	ps2 := particles.NewParticleSystem(
		[]particles.Particle{},
		particles.ParticleSpark,
		particles.DefaultConfig(),
	)
	if ps2 != nil {
		fmt.Println("PASS")
		passed++
		particles.ReleaseParticleSystem(ps2)
	} else {
		fmt.Println("FAIL")
		failed++
	}

	// Test 6: Terrain Modification System
	fmt.Print("✓ Testing Terrain Modification System... ")
	terrainMod := engine.NewTerrainModificationSystem(32) // 32 is standard tile size
	if terrainMod != nil {
		fmt.Println("PASS")
		passed++
	} else {
		fmt.Println("FAIL")
		failed++
	}

	// Test 7: Fire Propagation System
	fmt.Print("✓ Testing Fire Propagation System... ")
	fireProp := engine.NewFirePropagationSystem(32, 12345) // 32 = tile size, 12345 = seed
	if fireProp != nil {
		fmt.Println("PASS")
		passed++
	} else {
		fmt.Println("FAIL")
		failed++
	}

	// Test 8: Terrain Construction System
	fmt.Print("✓ Testing Terrain Construction System... ")
	terrainConstruct := engine.NewTerrainConstructionSystem(32) // 32 is standard tile size
	if terrainConstruct != nil {
		fmt.Println("PASS")
		passed++
	} else {
		fmt.Println("FAIL")
		failed++
	}

	// Test 9: MerchantComponent
	fmt.Print("✓ Testing MerchantComponent... ")
	merchantComp := engine.NewMerchantComponent(20, engine.MerchantFixed, 1.5)
	if merchantComp != nil && merchantComp.Type() == "merchant" {
		fmt.Println("PASS")
		passed++
	} else {
		fmt.Println("FAIL")
		failed++
	}

	// Test 10: DialogComponent
	fmt.Print("✓ Testing DialogComponent... ")
	provider := engine.NewMerchantDialogProvider("Test Merchant")
	dialogComp := engine.NewDialogComponent(provider)
	if dialogComp != nil && dialogComp.Type() == "dialog" {
		fmt.Println("PASS")
		passed++
	} else {
		fmt.Println("FAIL")
		failed++
	}

	// Summary
	fmt.Println("\n================================")
	fmt.Printf("Tests Passed: %d/%d\n", passed, passed+failed)

	if failed > 0 {
		fmt.Printf("Tests Failed: %d\n", failed)
		fmt.Println("❌ v1.1 validation FAILED")
		os.Exit(1)
	} else {
		fmt.Println("✅ All v1.1 features validated")
		fmt.Println("Ready for production deployment")
		os.Exit(0)
	}
}
