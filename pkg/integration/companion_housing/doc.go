// Package companion_housing provides integration between companion AI systems and
// player housing, enabling companions to interact with furniture, gain loyalty bonuses,
// train with XP multipliers, and access shared storage.
//
// # Overview
//
// The companion_housing package bridges the companion AI system with the housing system,
// allowing players to assign their companions to houses where they can rest, train, and
// access shared storage. This integration provides gameplay benefits through loyalty
// bonuses and experience multipliers based on housing quality and furniture.
//
// # Core Components
//
// ## CompanionHousingComponent
//
// ECS component that tracks a companion's housing integration state:
//
//	component := &companion_housing.CompanionHousingComponent{
//		OwnerHouseID:      playerHouseID,
//		BeddingID:         "bedding_001",
//		LastRestTime:      gameTime, // Use game time from TimeProvider for determinism
//		ActiveTraining:    "training_001",
//		TrainingBonus:     1.5,
//		SharedChestAccess: []string{"chest_001", "chest_002"},
//	}
//
// The component stores pure data. Use CompanionHousingSystem methods to:
//   - Check if companion is assigned to a house (OwnerHouseID != "")
//   - Check if companion has bedding (BeddingID != "")
//   - Check if companion is training (ActiveTraining != "")
//   - Check if companion has storage access (len(SharedChestAccess) > 0)
//   - Calculate days since rest using LastRestTime
//
// ## PetHomeManager
//
// Thread-safe manager for all companion-housing interactions:
//
//	manager := companion_housing.NewPetHomeManager()
//
//	// Register furniture
//	manager.AddBedding(houseID, furnitureID, companion_housing.BeddingQualityHigh)
//	manager.AddTrainingArea(houseID, furnitureID, companion_housing.TrainingTypeCombat)
//	manager.AddStorageChest(houseID, furnitureID, 50) // 50 slot capacity
//
//	// Assign companion
//	manager.AssignCompanionToHouse(companionID, houseID)
//	manager.AssignCompanionToBed(companionID, beddingID)
//
//	// Calculate bonuses
//	loyaltyBonus := manager.GetLoyaltyBonus(companionID)
//	xpMultiplier := manager.GetTrainingBonus(companionID)
//
// # Bedding System
//
// Companions assigned to bedding receive daily loyalty bonuses:
//
//	const (
//		BeddingQualityLow    = 0.5  // +0.5 loyalty per day
//		BeddingQualityMedium = 1.0  // +1.0 loyalty per day
//		BeddingQualityHigh   = 2.0  // +2.0 loyalty per day
//		BeddingQualityLuxury = 3.0  // +3.0 loyalty per day
//	)
//
// Loyalty bonus calculation:
//   - Base bonus from bedding quality
//   - Scales with days since last rest
//   - Encourages regular companion care
//   - Affects companion effectiveness in combat
//
// # Training System
//
// Companions training in designated areas receive XP multipliers:
//
//	const (
//		TrainingTypeCombat   = "combat"   // 1.5x combat XP
//		TrainingTypeMagic    = "magic"    // 1.5x magic XP
//		TrainingTypeCrafting = "crafting" // 1.5x crafting XP
//		TrainingTypeGeneral  = "general"  // 1.2x all XP
//	)
//
// Training mechanics:
//   - 1.5x XP multiplier for specialized training
//   - 1.2x XP multiplier for general training
//   - Training session management (start/end)
//   - Only one active training session per companion
//   - Training areas can support multiple companions
//
// # Storage System
//
// Companions can access shared storage chests in their assigned house:
//
//	manager.AddStorageChest(houseID, chestID, 50)         // 50 slots
//	manager.GrantStorageAccess(companionID, chestID)      // Grant access
//	slots := manager.GetAvailableStorage(companionID)     // Get capacity
//	manager.AddItemToStorage(chestID, itemID, quantity)   // Store item
//	manager.RemoveItemFromStorage(chestID, itemID, qty)   // Retrieve item
//
// Storage features:
//   - Shared inventory between player and companions
//   - Capacity management (slot-based)
//   - Item stacking support
//   - Thread-safe concurrent access
//
// # Thread Safety
//
// All manager operations are thread-safe using sync.RWMutex:
//
//	// Multiple goroutines can safely access manager
//	go manager.GetLoyaltyBonus(companion1)
//	go manager.GetLoyaltyBonus(companion2)
//	go manager.RecordRest(companion3)
//
// Read operations use RLock() for concurrent reads:
//   - GetLoyaltyBonus
//   - GetTrainingBonus
//   - GetAvailableStorage
//   - IsTrainingActive
//
// Write operations use Lock() for exclusive access:
//   - AssignCompanionToBed
//   - StartTrainingSession
//   - AddItemToStorage
//   - RecordRest
//
// # Usage Example
//
// Complete companion housing integration:
//
//	// Initialize manager
//	manager := companion_housing.NewPetHomeManager()
//
//	// Set up player house with companion furniture
//	houseID := "house_123"
//	manager.AddBedding(houseID, "bed_001", companion_housing.BeddingQualityHigh)
//	manager.AddTrainingArea(houseID, "train_001", companion_housing.TrainingTypeCombat)
//	manager.AddStorageChest(houseID, "chest_001", 50)
//
//	// Assign companion to house
//	companionID := uint64(42)
//	manager.AssignCompanionToHouse(companionID, houseID)
//	manager.AssignCompanionToBed(companionID, "bed_001")
//	manager.GrantStorageAccess(companionID, "chest_001")
//
//	// Start training session
//	manager.StartTrainingSession(companionID, "train_001")
//
//	// Calculate bonuses during gameplay
//	loyaltyBonus := manager.GetLoyaltyBonus(companionID)
//	xpMultiplier := manager.GetTrainingBonus(companionID)
//
//	// Apply bonuses to companion
//	companion.Loyalty += loyaltyBonus * deltaTime
//	companion.XP = earnedXP * xpMultiplier
//
//	// Record rest when companion uses bed
//	manager.RecordRest(companionID)
//
//	// End training session
//	manager.EndTrainingSession(companionID)
//
// # ECS Integration
//
// The CompanionHousingComponent integrates with the ECS architecture:
//
//	// Add component to companion entity
//	companion := world.CreateEntity()
//	companion.AddComponent(&engine.AIComponent{...})
//	companion.AddComponent(&companion_housing.CompanionHousingComponent{
//		CompanionID: companion.ID,
//	})
//
//	// Update component from manager state
//	if comp, ok := companion.GetComponent("companion_housing"); ok {
//		housingComp := comp.(*companion_housing.CompanionHousingComponent)
//		housingComp.UpdateFromManager(manager, companion.ID)
//	}
//
// # Performance Considerations
//
// The system is optimized for real-time gameplay:
//
//   - Hash map lookups for O(1) companion/furniture access
//   - Read-write locks minimize contention
//   - Bonus calculations are cached where possible
//   - Storage operations use efficient map structures
//
// Typical performance:
//   - GetLoyaltyBonus: <1 microsecond
//   - GetTrainingBonus: <1 microsecond
//   - AssignCompanionToBed: ~2 microseconds
//   - AddItemToStorage: ~3 microseconds
//
// # Testing
//
// The package includes comprehensive tests with >70% coverage:
//
//	func TestLoyaltyBonus(t *testing.T) {
//		manager := companion_housing.NewPetHomeManager()
//		manager.AddBedding("house1", "bed1", companion_housing.BeddingQualityHigh)
//		manager.AssignCompanionToBed(1, "bed1")
//		bonus := manager.GetLoyaltyBonus(1)
//		// Verify bonus calculation
//	}
//
// Benchmark tests validate performance:
//
//	BenchmarkGetLoyaltyBonus
//	BenchmarkGetTrainingBonus
//	BenchmarkStorageOperations
//
// # Integration
//
// This package integrates with other game systems:
//
//   - companion/learning: Training XP multipliers affect skill progression
//   - world/housing: Furniture placement and house management
//   - engine (ECS): Component-based architecture for companion entities
//   - saveload: Housing assignments persist across save/load
//
// # References
//
// For more information:
//   - Companion AI: pkg/companion/learning/doc.go
//   - Housing System: pkg/world/housing/doc.go
//   - Integration Systems: docs/INTEGRATION.md
package companion_housing
