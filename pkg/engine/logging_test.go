package engine

import (
	"bytes"
	"testing"

	"github.com/opd-ai/venture/pkg/logging"
	"github.com/opd-ai/venture/pkg/procgen/item"
	"github.com/sirupsen/logrus"
)

// TestProgressionSystemLogging verifies that logging works in progression system.
func TestProgressionSystemLogging(t *testing.T) {
	// Create a logger that writes to a buffer
	logger := logging.NewLogger(logging.Config{
		Level:       logging.DebugLevel,
		Format:      logging.TextFormat,
		AddCaller:   false,
		EnableColor: false,
	})

	var buf bytes.Buffer
	logger.SetOutput(&buf)

	// Create world and entity
	world := NewWorld()
	entity := world.CreateEntity()
	entity.AddComponent(&ExperienceComponent{
		Level:      1,
		CurrentXP:  0,
		RequiredXP: 100,
	})
	entity.AddComponent(&LevelScalingComponent{
		HealthPerLevel:     10,
		MagicPowerPerLevel: 5,
	})
	entity.AddComponent(&HealthComponent{
		Current: 100,
		Max:     100,
	})

	// Create progression system with logger
	ps := NewProgressionSystemWithLogger(world, logger)

	// Award XP
	err := ps.AwardXP(entity, 150) // Should trigger level up
	if err != nil {
		t.Fatalf("AwardXP failed: %v", err)
	}

	output := buf.String()

	// Verify expected log messages
	expectedPhrases := []string{
		"awarding XP to entity",
		"entity leveled up",
		"newLevel=2",
	}

	for _, phrase := range expectedPhrases {
		if !bytes.Contains([]byte(output), []byte(phrase)) {
			t.Errorf("Log output missing expected phrase: %q\nOutput:\n%s", phrase, output)
		}
	}
}

// TestInventorySystemLogging verifies that logging works in inventory system.
func TestInventorySystemLogging(t *testing.T) {
	// Create a logger that writes to a buffer
	logger := logging.NewLogger(logging.Config{
		Level:       logging.DebugLevel,
		Format:      logging.TextFormat,
		AddCaller:   false,
		EnableColor: false,
	})

	var buf bytes.Buffer
	logger.SetOutput(&buf)

	// Create world and entity
	world := NewWorld()
	entity := world.CreateEntity()
	entity.AddComponent(&InventoryComponent{
		Items:    make([]*item.Item, 0),
		MaxItems: 20,
	})

	// Create inventory system with logger
	is := NewInventorySystemWithLogger(world, logger)

	// Create a test item
	testItem := &item.Item{
		Name: "Test Sword",
		Type: item.TypeWeapon,
	}

	// Add item
	success, err := is.AddItemToInventory(entity.ID, testItem)
	if err != nil {
		t.Fatalf("AddItemToInventory failed: %v", err)
	}
	if !success {
		t.Fatal("AddItemToInventory returned false")
	}

	output := buf.String()

	// Verify expected log messages
	expectedPhrases := []string{
		"adding item to inventory",
		"itemName=\"Test Sword\"",
		"itemType=weapon",
		"success=true",
	}

	for _, phrase := range expectedPhrases {
		if !bytes.Contains([]byte(output), []byte(phrase)) {
			t.Errorf("Log output missing expected phrase: %q\nOutput:\n%s", phrase, output)
		}
	}
}

// TestLoggingDoesNotBreakDeterminism verifies that systems with logging
// behave identically to systems without logging.
func TestLoggingDoesNotBreakDeterminism(t *testing.T) {
	// Test progression system
	t.Run("ProgressionSystem", func(t *testing.T) {
		// Without logging
		world1 := NewWorld()
		entity1 := world1.CreateEntity()
		entity1.AddComponent(&ExperienceComponent{
			Level:      1,
			CurrentXP:  0,
			RequiredXP: 100,
		})
		ps1 := NewProgressionSystem(world1)
		ps1.AwardXP(entity1, 250)

		comp1, _ := entity1.GetComponent("experience")
		exp1 := comp1.(*ExperienceComponent)

		// With logging
		logger := createTestLogger()
		world2 := NewWorld()
		entity2 := world2.CreateEntity()
		entity2.AddComponent(&ExperienceComponent{
			Level:      1,
			CurrentXP:  0,
			RequiredXP: 100,
		})
		ps2 := NewProgressionSystemWithLogger(world2, logger)
		ps2.AwardXP(entity2, 250)

		comp2, _ := entity2.GetComponent("experience")
		exp2 := comp2.(*ExperienceComponent)

		// Compare results
		if exp1.Level != exp2.Level {
			t.Errorf("Level mismatch: %d (no log) vs %d (with log)", exp1.Level, exp2.Level)
		}
		if exp1.CurrentXP != exp2.CurrentXP {
			t.Errorf("XP mismatch: %d (no log) vs %d (with log)", exp1.CurrentXP, exp2.CurrentXP)
		}
	})
	t.Run("InventorySystem", func(t *testing.T) {
		testItem := &item.Item{
			Name: "Test Item",
			Type: item.TypeConsumable,
		}

		// Without logging
		world1 := NewWorld()
		entity1 := world1.CreateEntity()
		entity1.AddComponent(&InventoryComponent{
			Items:    make([]*item.Item, 0),
			MaxItems: 20,
		})
		is1 := NewInventorySystem(world1)
		success1, _ := is1.AddItemToInventory(entity1.ID, testItem)

		comp1, _ := entity1.GetComponent("inventory")
		inv1 := comp1.(*InventoryComponent)

		// With logging
		logger := createTestLogger()
		world2 := NewWorld()
		entity2 := world2.CreateEntity()
		entity2.AddComponent(&InventoryComponent{
			Items:    make([]*item.Item, 0),
			MaxItems: 20,
		})
		is2 := NewInventorySystemWithLogger(world2, logger)
		success2, _ := is2.AddItemToInventory(entity2.ID, testItem)

		comp2, _ := entity2.GetComponent("inventory")
		inv2 := comp2.(*InventoryComponent)

		// Compare results
		if success1 != success2 {
			t.Errorf("Success mismatch: %v (no log) vs %v (with log)", success1, success2)
		}
		if len(inv1.Items) != len(inv2.Items) {
			t.Errorf("Item count mismatch: %d (no log) vs %d (with log)", len(inv1.Items), len(inv2.Items))
		}
	})
}

// createTestLogger creates a logger for testing that discards output.
func createTestLogger() *logrus.Logger {
	logger := logging.NewLogger(logging.Config{
		Level:       logging.DebugLevel,
		Format:      logging.TextFormat,
		AddCaller:   false,
		EnableColor: false,
	})
	var buf bytes.Buffer
	logger.SetOutput(&buf)
	return logger
}
