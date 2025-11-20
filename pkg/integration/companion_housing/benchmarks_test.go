package companion_housing

import (
	"fmt"
	"testing"
)

func BenchmarkBedding_LoyaltyBonus(b *testing.B) {
	bedding := &CompanionBedding{Quality: BeddingStandard}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = bedding.LoyaltyBonus()
	}
}

func BenchmarkTrainingArea_XPBonus(b *testing.B) {
	area := &TrainingArea{Type: TrainingCombat}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = area.XPBonus()
	}
}

func BenchmarkStorageChest_AddRemove(b *testing.B) {
	chest := &StorageChest{
		Capacity: 100,
		Items:    make([]string, 0, 100),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		chest.AddItem(fmt.Sprintf("item_%d", i))
		if len(chest.Items) > 50 {
			chest.RemoveItem(chest.Items[0])
		}
	}
}

func BenchmarkPetHomeManager_GetLoyaltyBonus(b *testing.B) {
	manager := NewPetHomeManager()

	// Setup: 10 houses with 5 bedding each
	for h := 0; h < 10; h++ {
		houseID := fmt.Sprintf("house_%d", h)
		for bed := 0; bed < 5; bed++ {
			bedID := fmt.Sprintf("bed_%d_%d", h, bed)
			manager.AddBedding(houseID, bedID, BeddingStandard)
		}
	}

	// Assign companions
	companionID := uint64(100)
	manager.AssignCompanionToBed(companionID, "bed_0_0")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = manager.GetLoyaltyBonus(companionID, "house_0")
	}
}

func BenchmarkPetHomeManager_GetTrainingBonus(b *testing.B) {
	manager := NewPetHomeManager()

	// Setup: 10 houses with 3 training areas each
	for h := 0; h < 10; h++ {
		houseID := fmt.Sprintf("house_%d", h)
		for t := 0; t < 3; t++ {
			trainingID := fmt.Sprintf("training_%d_%d", h, t)
			manager.AddTrainingArea(houseID, trainingID, TrainingCombat)
		}
	}

	// Start training session
	companionID := uint64(100)
	manager.StartTrainingSession(companionID, "training_0_0")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = manager.GetTrainingBonus(companionID, "house_0")
	}
}

func BenchmarkPetHomeManager_AssignCompanion(b *testing.B) {
	manager := NewPetHomeManager()

	// Pre-create bedding
	for i := 0; i < 100; i++ {
		manager.AddBedding("house_1", fmt.Sprintf("bed_%d", i), BeddingStandard)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		companionID := uint64(i % 100)
		bedID := fmt.Sprintf("bed_%d", i%100)
		_ = manager.AssignCompanionToBed(companionID, bedID)
	}
}

func BenchmarkPetHomeManager_ConcurrentReads(b *testing.B) {
	manager := NewPetHomeManager()

	// Setup
	for h := 0; h < 5; h++ {
		houseID := fmt.Sprintf("house_%d", h)
		for bed := 0; bed < 10; bed++ {
			bedID := fmt.Sprintf("bed_%d_%d", h, bed)
			manager.AddBedding(houseID, bedID, BeddingStandard)
			manager.AssignCompanionToBed(uint64(h*10+bed), bedID)
		}
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		companionID := uint64(25) // Middle companion
		for pb.Next() {
			_ = manager.GetLoyaltyBonus(companionID, "house_2")
			_ = manager.GetCompanionHome(companionID)
		}
	})
}

func BenchmarkStorageChest_AvailableSlots(b *testing.B) {
	chest := &StorageChest{
		Capacity: 100,
		Items:    make([]string, 50),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = chest.AvailableSlots()
	}
}

func BenchmarkPetHomeManager_GetSharedStorageCapacity(b *testing.B) {
	manager := NewPetHomeManager()

	// Setup: multiple houses with chests
	for h := 0; h < 10; h++ {
		houseID := fmt.Sprintf("house_%d", h)
		for c := 0; c < 5; c++ {
			chestID := fmt.Sprintf("chest_%d_%d", h, c)
			manager.AddStorageChest(houseID, chestID, 50, c < 3) // 3 shared, 2 private
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = manager.GetSharedStorageCapacity("house_5")
	}
}
