// Package main provides a CLI tool for testing companion housing integration.
package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/opd-ai/venture/pkg/integration/companion_housing"
)

var (
	verbose = flag.Bool("verbose", false, "Enable verbose output")
	mode    = flag.String("mode", "demo", "Test mode: demo, loyalty, training, storage, all")
)

func main() {
	flag.Parse()

	fmt.Println("=== Companion Housing Integration Test Tool ===\n")

	switch *mode {
	case "demo":
		runDemo()
	case "loyalty":
		testLoyaltySystem()
	case "training":
		testTrainingSystem()
	case "storage":
		testStorageSystem()
	case "all":
		runDemo()
		testLoyaltySystem()
		testTrainingSystem()
		testStorageSystem()
	default:
		fmt.Printf("Unknown mode: %s\n", *mode)
		fmt.Println("Available modes: demo, loyalty, training, storage, all")
	}
}

func runDemo() {
	fmt.Println("## Demo: Complete Companion Housing Setup ##\n")

	manager := companion_housing.NewPetHomeManager()

	// Setup house with all furniture types
	fmt.Println("Setting up house with companion furniture:")
	manager.AddBedding("demo_house", "luxury_bed", companion_housing.BeddingLuxury)
	manager.AddTrainingArea("demo_house", "combat_dummy", companion_housing.TrainingCombat)
	manager.AddTrainingArea("demo_house", "magic_crystal", companion_housing.TrainingMagic)
	manager.AddStorageChest("demo_house", "shared_chest", 75, true)
	manager.AddStorageChest("demo_house", "private_chest", 50, false)

	fmt.Printf("  - Luxury bedding (0.2 loyalty/day)\n")
	fmt.Printf("  - Combat training dummy (1.5x XP)\n")
	fmt.Printf("  - Magic focus crystal (1.5x XP)\n")
	fmt.Printf("  - Shared chest (75 slots)\n")
	fmt.Printf("  - Private chest (50 slots)\n\n")

	// Assign companion
	companionID := uint64(1001)
	err := manager.AssignCompanionToBed(companionID, "luxury_bed")
	if err != nil {
		fmt.Printf("Error assigning companion: %v\n", err)
		return
	}

	fmt.Printf("Companion %d assigned to luxury bedding\n", companionID)

	// Show loyalty bonus
	bonus := manager.GetLoyaltyBonus(companionID, "demo_house")
	fmt.Printf("Daily loyalty bonus: +%.2f (vs +0.05 without housing)\n", bonus)
	fmt.Printf("Loyalty improvement: %.0f%%\n\n", (bonus/0.05-1)*100)

	// Start training
	err = manager.StartTrainingSession(companionID, "combat_dummy")
	if err != nil {
		fmt.Printf("Error starting training: %v\n", err)
		return
	}

	trainingBonus := manager.GetTrainingBonus(companionID, "demo_house")
	fmt.Printf("Training started in combat area\n")
	fmt.Printf("XP multiplier: %.2fx (50%% faster skill progression)\n\n", trainingBonus)

	// Storage info
	sharedCapacity := manager.GetSharedStorageCapacity("demo_house")
	fmt.Printf("Shared storage capacity: %d slots\n", sharedCapacity)
	fmt.Printf("Companion can access shared chest for item storage\n\n")

	fmt.Println("Demo complete!\n")
}

func testLoyaltySystem() {
	fmt.Println("## Testing Loyalty Bonus System ##\n")

	manager := companion_housing.NewPetHomeManager()

	qualities := []companion_housing.BeddingQuality{
		companion_housing.BeddingBasic,
		companion_housing.BeddingStandard,
		companion_housing.BeddingAdvanced,
		companion_housing.BeddingLuxury,
	}

	names := []string{"Basic", "Standard", "Advanced", "Luxury"}

	fmt.Println("Loyalty bonuses by bedding quality:")
	for i, quality := range qualities {
		houseID := fmt.Sprintf("house_%d", i)
		bedID := fmt.Sprintf("bed_%d", i)
		companionID := uint64(2000 + i)

		manager.AddBedding(houseID, bedID, quality)
		manager.AssignCompanionToBed(companionID, bedID)

		bonus := manager.GetLoyaltyBonus(companionID, houseID)
		baseline := 0.05

		fmt.Printf("  %s: +%.2f/day (%.0fx baseline)\n", names[i], bonus, bonus/baseline)
	}

	// Simulate rest
	companionID := uint64(2001)
	fmt.Printf("\nSimulating rest for companion %d...\n", companionID)
	err := manager.RecordRest(companionID)
	if err != nil {
		fmt.Printf("Error recording rest: %v\n", err)
		return
	}

	fmt.Println("Rest recorded successfully")

	// Simulate multi-day loyalty calculation
	fmt.Println("\nProjected loyalty gains (30 days):")
	bonus := manager.GetLoyaltyBonus(companionID, "house_1")
	for days := 7; days <= 30; days += 7 {
		totalGain := bonus * float64(days)
		fmt.Printf("  Day %2d: +%.1f loyalty (%.1f%% toward max)\n", days, totalGain, totalGain)
	}

	fmt.Println()
}

func testTrainingSystem() {
	fmt.Println("## Testing Training System ##\n")

	manager := companion_housing.NewPetHomeManager()

	areas := []companion_housing.TrainingAreaType{
		companion_housing.TrainingCombat,
		companion_housing.TrainingMagic,
		companion_housing.TrainingAgility,
		companion_housing.TrainingStrength,
		companion_housing.TrainingObedience,
		companion_housing.TrainingEndurance,
	}

	fmt.Println("Training area XP multipliers:")
	for i, areaType := range areas {
		furnitureID := fmt.Sprintf("training_%d", i)
		manager.AddTrainingArea("training_house", furnitureID, areaType)

		companionID := uint64(3000 + i)
		manager.StartTrainingSession(companionID, furnitureID)

		bonus := manager.GetTrainingBonus(companionID, "training_house")
		increase := (bonus - 1.0) * 100

		fmt.Printf("  %s: %.2fx (+%.0f%% faster)\n", areaType.String(), bonus, increase)
	}

	// Simulate training session duration
	fmt.Println("\nSimulating 1-hour training session...")
	time.Sleep(50 * time.Millisecond) // Simulate passage of time

	companionID := uint64(3000)
	manager.EndTrainingSession(companionID, "training_0")
	fmt.Println("Training session ended")

	bonus := manager.GetTrainingBonus(companionID, "training_house")
	if bonus == 1.0 {
		fmt.Println("XP multiplier returned to baseline (1.0x) after session end")
	}

	fmt.Println()
}

func testStorageSystem() {
	fmt.Println("## Testing Storage System ##\n")

	manager := companion_housing.NewPetHomeManager()

	// Add multiple chests with different configurations
	manager.AddStorageChest("storage_house", "shared_1", 50, true)
	manager.AddStorageChest("storage_house", "shared_2", 75, true)
	manager.AddStorageChest("storage_house", "private_1", 100, false)

	fmt.Println("Storage chest configuration:")
	fmt.Println("  - Shared chest 1: 50 slots (accessible by companions)")
	fmt.Println("  - Shared chest 2: 75 slots (accessible by companions)")
	fmt.Println("  - Private chest: 100 slots (player only)")

	sharedCapacity := manager.GetSharedStorageCapacity("storage_house")
	fmt.Printf("\nTotal shared storage: %d slots\n", sharedCapacity)

	// Test chest operations
	chest := manager.GetStorageChest("shared_1")
	if chest == nil {
		fmt.Println("Error: chest not found")
		return
	}

	fmt.Println("\nTesting chest operations:")
	fmt.Printf("Initial available slots: %d\n", chest.AvailableSlots())

	// Add items
	for i := 0; i < 5; i++ {
		itemID := fmt.Sprintf("item_%d", i)
		if chest.AddItem(itemID) {
			fmt.Printf("  + Added %s\n", itemID)
		}
	}

	fmt.Printf("Available slots after adding: %d\n", chest.AvailableSlots())

	// Remove item
	if chest.RemoveItem("item_2") {
		fmt.Println("  - Removed item_2")
	}

	fmt.Printf("Final available slots: %d\n", chest.AvailableSlots())

	// Test capacity limits
	fmt.Println("\nTesting capacity limits...")
	maxCapacity := 50
	filled := len(chest.Items)
	for i := 0; i < maxCapacity; i++ {
		itemID := fmt.Sprintf("filler_%d", i)
		if !chest.AddItem(itemID) {
			fmt.Printf("Chest full at %d/%d items\n", len(chest.Items), maxCapacity)
			break
		}
		filled++
	}

	fmt.Println()
}
