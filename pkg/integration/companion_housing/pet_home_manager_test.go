package companion_housing

import (
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

func TestBeddingQuality_LoyaltyBonus(t *testing.T) {
	tests := []struct {
		name    string
		quality BeddingQuality
		want    float64
	}{
		{"basic bedding", BeddingBasic, 0.05},
		{"standard bedding", BeddingStandard, 0.1},
		{"advanced bedding", BeddingAdvanced, 0.15},
		{"luxury bedding", BeddingLuxury, 0.2},
		{"custom quality 0.75", BeddingQuality(0.75), 0.075},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bedding := &CompanionBedding{Quality: tt.quality}
			got := bedding.LoyaltyBonus()
			// Use approximate comparison for floating point
			if diff := got - tt.want; diff < -0.0001 || diff > 0.0001 {
				t.Errorf("LoyaltyBonus() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTrainingAreaType_String(t *testing.T) {
	tests := []struct {
		areaType TrainingAreaType
		want     string
	}{
		{TrainingCombat, "Combat Training Dummy"},
		{TrainingAgility, "Agility Obstacle Course"},
		{TrainingMagic, "Magic Focus Crystal"},
		{TrainingObedience, "Obedience Training Post"},
		{TrainingStrength, "Strength Training Rack"},
		{TrainingEndurance, "Endurance Training Wheel"},
		{TrainingAreaType("unknown"), "Unknown Training Area"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.areaType.String()
			if got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTrainingAreaType_XPMultiplier(t *testing.T) {
	tests := []struct {
		name     string
		areaType TrainingAreaType
		want     float64
	}{
		{"combat (high)", TrainingCombat, 1.5},
		{"magic (high)", TrainingMagic, 1.5},
		{"agility (medium)", TrainingAgility, 1.35},
		{"strength (medium)", TrainingStrength, 1.35},
		{"obedience (support)", TrainingObedience, 1.25},
		{"endurance (support)", TrainingEndurance, 1.25},
		{"unknown", TrainingAreaType("invalid"), 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.areaType.XPMultiplier()
			if got != tt.want {
				t.Errorf("XPMultiplier() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTrainingArea_XPBonus(t *testing.T) {
	area := &TrainingArea{Type: TrainingCombat}
	got := area.XPBonus()
	want := 1.5

	if got != want {
		t.Errorf("XPBonus() = %v, want %v", got, want)
	}
}

func TestStorageChest_AddItem(t *testing.T) {
	chest := &StorageChest{
		Capacity: 5,
		Items:    []string{},
	}

	// Add items until full
	for i := 0; i < 5; i++ {
		if !chest.AddItem("item_" + string(rune('A'+i))) {
			t.Errorf("AddItem() failed at item %d, expected success", i)
		}
	}

	// Verify capacity reached
	if len(chest.Items) != 5 {
		t.Errorf("len(Items) = %d, want 5", len(chest.Items))
	}

	// Adding to full chest should fail
	if chest.AddItem("overflow") {
		t.Error("AddItem() succeeded on full chest, expected false")
	}
}

func TestStorageChest_RemoveItem(t *testing.T) {
	chest := &StorageChest{
		Capacity: 10,
		Items:    []string{"item_A", "item_B", "item_C"},
	}

	// Remove existing item
	if !chest.RemoveItem("item_B") {
		t.Error("RemoveItem() = false, want true for existing item")
	}

	// Verify item removed
	if len(chest.Items) != 2 {
		t.Errorf("len(Items) = %d, want 2 after removal", len(chest.Items))
	}

	// Check items are item_A and item_C
	found := make(map[string]bool)
	for _, item := range chest.Items {
		found[item] = true
	}
	if !found["item_A"] || !found["item_C"] {
		t.Errorf("Items = %v, want [item_A, item_C]", chest.Items)
	}

	// Remove non-existent item should fail
	if chest.RemoveItem("item_Z") {
		t.Error("RemoveItem() = true for non-existent item, want false")
	}
}

func TestStorageChest_AvailableSlots(t *testing.T) {
	tests := []struct {
		name     string
		capacity int
		items    int
		want     int
	}{
		{"empty chest", 50, 0, 50},
		{"half full", 100, 50, 50},
		{"nearly full", 75, 70, 5},
		{"full chest", 60, 60, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chest := &StorageChest{
				Capacity: tt.capacity,
				Items:    make([]string, tt.items),
			}
			got := chest.AvailableSlots()
			if got != tt.want {
				t.Errorf("AvailableSlots() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestNewPetHomeManager(t *testing.T) {
	manager := NewPetHomeManager()

	if manager == nil {
		t.Fatal("NewPetHomeManager() returned nil")
	}

	// Verify all maps initialized
	if manager.bedding == nil || manager.trainingAreas == nil || manager.storageChests == nil {
		t.Error("NewPetHomeManager() did not initialize maps")
	}
}

func TestNewPetHomeManagerWithLogger(t *testing.T) {
	logger := logrus.New().WithField("test", "companion_housing")
	manager := NewPetHomeManagerWithLogger(logger)

	if manager == nil {
		t.Fatal("NewPetHomeManagerWithLogger() returned nil")
	}

	// Verify logger is set
	if manager.logger != logger {
		t.Error("NewPetHomeManagerWithLogger() did not set logger")
	}

	// Verify all maps initialized
	if manager.bedding == nil || manager.trainingAreas == nil || manager.storageChests == nil {
		t.Error("NewPetHomeManagerWithLogger() did not initialize maps")
	}
}

func TestPetHomeManager_AddRemoveBedding(t *testing.T) {
	manager := NewPetHomeManager()

	manager.AddBedding("house_1", "bed_1", BeddingStandard)

	// Verify bedding added
	if len(manager.bedding) != 1 {
		t.Errorf("len(bedding) = %d, want 1", len(manager.bedding))
	}

	bedding, ok := manager.bedding["bed_1"]
	if !ok || bedding.HouseID != "house_1" || bedding.Quality != BeddingStandard {
		t.Errorf("Bedding not added correctly: %+v", bedding)
	}

	// Remove bedding
	manager.RemoveBedding("bed_1")

	if len(manager.bedding) != 0 {
		t.Errorf("len(bedding) = %d, want 0 after removal", len(manager.bedding))
	}
}

func TestPetHomeManager_AssignCompanionToBed(t *testing.T) {
	manager := NewPetHomeManager()
	manager.AddBedding("house_1", "bed_1", BeddingStandard)

	// Assign companion to bed
	err := manager.AssignCompanionToBed(100, "bed_1")
	if err != nil {
		t.Fatalf("AssignCompanionToBed() error = %v", err)
	}

	// Verify assignment
	bedding := manager.bedding["bed_1"]
	if bedding.CompanionID != 100 {
		t.Errorf("CompanionID = %d, want 100", bedding.CompanionID)
	}

	// Try to assign different companion to same bed (should fail)
	err = manager.AssignCompanionToBed(200, "bed_1")
	if err == nil {
		t.Error("AssignCompanionToBed() succeeded for occupied bed, expected error")
	}

	// Re-assign same companion (should succeed)
	err = manager.AssignCompanionToBed(100, "bed_1")
	if err != nil {
		t.Errorf("AssignCompanionToBed() failed for same companion: %v", err)
	}

	// Assign to non-existent bed (should fail)
	err = manager.AssignCompanionToBed(300, "bed_999")
	if err == nil {
		t.Error("AssignCompanionToBed() succeeded for non-existent bed, expected error")
	}
}

func TestPetHomeManager_GetLoyaltyBonus(t *testing.T) {
	manager := NewPetHomeManager()
	manager.AddBedding("house_1", "bed_1", BeddingAdvanced)
	manager.AssignCompanionToBed(100, "bed_1")

	bonus := manager.GetLoyaltyBonus(100, "house_1")
	want := 0.15 // BeddingAdvanced = 1.5 * 0.1

	// Use approximate comparison for floating point
	if diff := bonus - want; diff < -0.0001 || diff > 0.0001 {
		t.Errorf("GetLoyaltyBonus() = %v, want %v", bonus, want)
	}

	// No bonus for unassigned companion
	bonus = manager.GetLoyaltyBonus(200, "house_1")
	if bonus != 0.0 {
		t.Errorf("GetLoyaltyBonus() = %v for unassigned companion, want 0.0", bonus)
	}

	// No bonus for wrong house
	bonus = manager.GetLoyaltyBonus(100, "house_2")
	if bonus != 0.0 {
		t.Errorf("GetLoyaltyBonus() = %v for wrong house, want 0.0", bonus)
	}
}

func TestPetHomeManager_RecordRest(t *testing.T) {
	manager := NewPetHomeManager()
	manager.AddBedding("house_1", "bed_1", BeddingStandard)
	manager.AssignCompanionToBed(100, "bed_1")

	before := manager.bedding["bed_1"].LastRestTime

	// Use explicit time for determinism
	restTime := time.Date(2026, 2, 13, 12, 0, 0, 0, time.UTC)

	err := manager.RecordRest(100, restTime)
	if err != nil {
		t.Fatalf("RecordRest() error = %v", err)
	}

	after := manager.bedding["bed_1"].LastRestTime

	if after != restTime {
		t.Errorf("RecordRest() did not set LastRestTime correctly, got %v, want %v", after, restTime)
	}

	if !after.After(before) {
		t.Error("RecordRest() did not update LastRestTime")
	}

	// Recording rest for unassigned companion should fail
	err = manager.RecordRest(999, restTime)
	if err == nil {
		t.Error("RecordRest() succeeded for unassigned companion, expected error")
	}
}

func TestPetHomeManager_AddRemoveTrainingArea(t *testing.T) {
	manager := NewPetHomeManager()
	manager.AddTrainingArea("house_1", "training_1", TrainingCombat)

	// Verify training area added
	if len(manager.trainingAreas) != 1 {
		t.Errorf("len(trainingAreas) = %d, want 1", len(manager.trainingAreas))
	}

	area, ok := manager.trainingAreas["training_1"]
	if !ok || area.Type != TrainingCombat {
		t.Errorf("Training area not added correctly: %+v", area)
	}

	// Remove training area
	manager.RemoveTrainingArea("training_1")

	if len(manager.trainingAreas) != 0 {
		t.Errorf("len(trainingAreas) = %d, want 0 after removal", len(manager.trainingAreas))
	}
}

func TestPetHomeManager_TrainingSession(t *testing.T) {
	manager := NewPetHomeManager()
	manager.AddTrainingArea("house_1", "training_1", TrainingMagic)

	// Use explicit time for determinism
	sessionTime := time.Date(2026, 2, 13, 14, 30, 0, 0, time.UTC)

	// Start training session
	err := manager.StartTrainingSession(100, "training_1", sessionTime)
	if err != nil {
		t.Fatalf("StartTrainingSession() error = %v", err)
	}

	// Verify session active
	area := manager.trainingAreas["training_1"]
	startTime, ok := area.ActiveSessions[100]
	if !ok {
		t.Error("Training session not started")
	}
	if startTime != sessionTime {
		t.Errorf("Training session time = %v, want %v", startTime, sessionTime)
	}

	// End training session
	manager.EndTrainingSession(100, "training_1")

	// Verify session ended
	if _, ok := area.ActiveSessions[100]; ok {
		t.Error("Training session not ended")
	}

	// Starting session on non-existent area should fail
	err = manager.StartTrainingSession(100, "training_999", sessionTime)
	if err == nil {
		t.Error("StartTrainingSession() succeeded for non-existent area, expected error")
	}
}

func TestPetHomeManager_GetTrainingBonus(t *testing.T) {
	manager := NewPetHomeManager()
	manager.AddTrainingArea("house_1", "training_1", TrainingCombat)

	// Use explicit time for determinism
	sessionTime := time.Date(2026, 2, 13, 14, 30, 0, 0, time.UTC)
	manager.StartTrainingSession(100, "training_1", sessionTime)

	bonus := manager.GetTrainingBonus(100, "house_1")
	want := 1.5 // TrainingCombat XP multiplier

	if bonus != want {
		t.Errorf("GetTrainingBonus() = %v, want %v", bonus, want)
	}

	// No bonus for inactive companion
	bonus = manager.GetTrainingBonus(200, "house_1")
	if bonus != 1.0 {
		t.Errorf("GetTrainingBonus() = %v for inactive companion, want 1.0", bonus)
	}

	// End session, bonus should revert to 1.0
	manager.EndTrainingSession(100, "training_1")
	bonus = manager.GetTrainingBonus(100, "house_1")
	if bonus != 1.0 {
		t.Errorf("GetTrainingBonus() = %v after session ended, want 1.0", bonus)
	}
}

func TestPetHomeManager_StorageChest(t *testing.T) {
	manager := NewPetHomeManager()
	manager.AddStorageChest("house_1", "chest_1", 50, true)

	// Verify chest added
	if len(manager.storageChests) != 1 {
		t.Errorf("len(storageChests) = %d, want 1", len(manager.storageChests))
	}

	chest := manager.GetStorageChest("chest_1")
	if chest == nil || chest.Capacity != 50 || !chest.SharedWithPets {
		t.Errorf("Storage chest not added correctly: %+v", chest)
	}

	// Remove chest
	manager.RemoveStorageChest("chest_1")

	if len(manager.storageChests) != 0 {
		t.Errorf("len(storageChests) = %d, want 0 after removal", len(manager.storageChests))
	}
}

func TestPetHomeManager_GetSharedStorageCapacity(t *testing.T) {
	manager := NewPetHomeManager()
	manager.AddStorageChest("house_1", "chest_1", 50, true)
	manager.AddStorageChest("house_1", "chest_2", 75, true)
	manager.AddStorageChest("house_1", "chest_3", 100, false) // Not shared

	total := manager.GetSharedStorageCapacity("house_1")
	want := 125 // chest_1 + chest_2 only

	if total != want {
		t.Errorf("GetSharedStorageCapacity() = %d, want %d", total, want)
	}

	// Empty house
	total = manager.GetSharedStorageCapacity("house_999")
	if total != 0 {
		t.Errorf("GetSharedStorageCapacity() = %d for empty house, want 0", total)
	}
}

func TestPetHomeManager_GetCompanionHome(t *testing.T) {
	manager := NewPetHomeManager()
	manager.AddBedding("house_1", "bed_1", BeddingStandard)
	manager.AssignCompanionToBed(100, "bed_1")

	houseID := manager.GetCompanionHome(100)
	if houseID != "house_1" {
		t.Errorf("GetCompanionHome() = %s, want house_1", houseID)
	}

	// Unassigned companion
	houseID = manager.GetCompanionHome(999)
	if houseID != "" {
		t.Errorf("GetCompanionHome() = %s for unassigned companion, want empty string", houseID)
	}
}

func TestPetHomeManager_GetHouseCounts(t *testing.T) {
	manager := NewPetHomeManager()
	manager.AddBedding("house_1", "bed_1", BeddingStandard)
	manager.AddBedding("house_1", "bed_2", BeddingLuxury)
	manager.AddTrainingArea("house_1", "training_1", TrainingCombat)

	beddingCount := manager.GetHouseBeddingCount("house_1")
	if beddingCount != 2 {
		t.Errorf("GetHouseBeddingCount() = %d, want 2", beddingCount)
	}

	trainingCount := manager.GetHouseTrainingCount("house_1")
	if trainingCount != 1 {
		t.Errorf("GetHouseTrainingCount() = %d, want 1", trainingCount)
	}

	// Empty house
	beddingCount = manager.GetHouseBeddingCount("house_999")
	if beddingCount != 0 {
		t.Errorf("GetHouseBeddingCount() = %d for empty house, want 0", beddingCount)
	}
}

func TestPetHomeManager_RemoveBedding_UpdatesHouseSlice(t *testing.T) {
	manager := NewPetHomeManager()
	manager.AddBedding("house_1", "bed_1", BeddingStandard)
	manager.AddBedding("house_1", "bed_2", BeddingLuxury)

	if len(manager.houseBedding["house_1"]) != 2 {
		t.Fatalf("houseBedding slice = %d, want 2", len(manager.houseBedding["house_1"]))
	}

	manager.RemoveBedding("bed_1")

	if len(manager.houseBedding["house_1"]) != 1 {
		t.Errorf("houseBedding slice = %d after removal, want 1", len(manager.houseBedding["house_1"]))
	}
	if manager.houseBedding["house_1"][0] != "bed_2" {
		t.Errorf("houseBedding[0] = %s, want bed_2", manager.houseBedding["house_1"][0])
	}

	manager.RemoveBedding("bed_2")

	if len(manager.houseBedding["house_1"]) != 0 {
		t.Errorf("houseBedding slice = %d after removing all, want 0", len(manager.houseBedding["house_1"]))
	}
}

func TestPetHomeManager_RemoveTrainingArea_UpdatesHouseSlice(t *testing.T) {
	manager := NewPetHomeManager()
	manager.AddTrainingArea("house_1", "train_1", TrainingCombat)
	manager.AddTrainingArea("house_1", "train_2", TrainingMagic)

	if len(manager.houseTraining["house_1"]) != 2 {
		t.Fatalf("houseTraining slice = %d, want 2", len(manager.houseTraining["house_1"]))
	}

	manager.RemoveTrainingArea("train_1")

	if len(manager.houseTraining["house_1"]) != 1 {
		t.Errorf("houseTraining slice = %d after removal, want 1", len(manager.houseTraining["house_1"]))
	}
	if manager.houseTraining["house_1"][0] != "train_2" {
		t.Errorf("houseTraining[0] = %s, want train_2", manager.houseTraining["house_1"][0])
	}
}

func TestPetHomeManager_RemoveStorageChest_UpdatesHouseSlice(t *testing.T) {
	manager := NewPetHomeManager()
	manager.AddStorageChest("house_1", "chest_1", 50, true)
	manager.AddStorageChest("house_1", "chest_2", 75, false)

	if len(manager.houseStorage["house_1"]) != 2 {
		t.Fatalf("houseStorage slice = %d, want 2", len(manager.houseStorage["house_1"]))
	}

	manager.RemoveStorageChest("chest_1")

	if len(manager.houseStorage["house_1"]) != 1 {
		t.Errorf("houseStorage slice = %d after removal, want 1", len(manager.houseStorage["house_1"]))
	}
	if manager.houseStorage["house_1"][0] != "chest_2" {
		t.Errorf("houseStorage[0] = %s, want chest_2", manager.houseStorage["house_1"][0])
	}

	// Verify capacity reflects removal
	total := manager.GetSharedStorageCapacity("house_1")
	if total != 0 { // chest_2 is not shared
		t.Errorf("GetSharedStorageCapacity() = %d after removing shared chest, want 0", total)
	}
}

func TestPetHomeManager_ConcurrentAccess(t *testing.T) {
	manager := NewPetHomeManager()
	manager.AddBedding("house_1", "bed_1", BeddingStandard)
	manager.AddTrainingArea("house_1", "training_1", TrainingCombat)

	// Concurrent read operations (should not panic)
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			_ = manager.GetLoyaltyBonus(100, "house_1")
			_ = manager.GetTrainingBonus(100, "house_1")
			_ = manager.GetCompanionHome(100)
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}
