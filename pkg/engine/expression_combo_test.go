package engine

import (
	"testing"
)

func TestComboType_String(t *testing.T) {
	tests := []struct {
		name     string
		combo    ComboType
		expected string
	}{
		{"Synchronized", ComboSynchronized, "Synchronized"},
		{"Wave", ComboWave, "Wave"},
		{"Circle", ComboCircle, "Circle"},
		{"Unknown", ComboType(999), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.combo.String()
			if result != tt.expected {
				t.Errorf("ComboType.String() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestExpressionComboComponent_Type(t *testing.T) {
	comp := ExpressionComboComponent{}
	if comp.Type() != "expressioncombo" {
		t.Errorf("ExpressionComboComponent.Type() = %q, want %q", comp.Type(), "expressioncombo")
	}
}

func TestNewExpressionComboSystem(t *testing.T) {
	world := NewWorld()
	system := NewExpressionComboSystem(world)

	if system == nil {
		t.Errorf("NewExpressionComboSystem() returned nil")
		return
	}
	if system.world != world {
		t.Errorf("NewExpressionComboSystem() world not set correctly")
	}
	if system.comboWindow != 2.0 {
		t.Errorf("NewExpressionComboSystem() comboWindow = %f, want 2.0", system.comboWindow)
	}
	if system.pendingCombos == nil {
		t.Errorf("NewExpressionComboSystem() pendingCombos not initialized")
	}
}

func TestExpressionComboSystem_SingleEntity(t *testing.T) {
	world := NewWorld()
	system := NewExpressionComboSystem(world)

	// Create entity with expression
	entity := world.CreateEntity()
	world.Update(0) // Process entity addition
	expComp := &ExpressionComponent{
		ActiveExpression: ExpressionDance,
		ExpressionTime:   3.0,
		Cooldown:         0,
	}
	entity.AddComponent(expComp)

	// Update system - single entity shouldn't create combo
	system.Update(0.1)

	// Check no active combo
	activeCombo := system.GetActiveCombo(entity.ID)
	if activeCombo != nil {
		t.Errorf("Single entity should not create combo")
	}

	// Check pending combos
	if len(system.pendingCombos) != 1 {
		t.Errorf("Should have 1 pending combo, got %d", len(system.pendingCombos))
	}
}

func TestExpressionComboSystem_TwoEntitiesSynchronized(t *testing.T) {
	world := NewWorld()
	system := NewExpressionComboSystem(world)

	// Create first entity with expression component before flushing to world
	// so the query cache is populated with the correct component set on flush.
	entity1 := world.CreateEntity()
	expComp1 := &ExpressionComponent{
		ActiveExpression: ExpressionDance,
		ExpressionTime:   3.0,
		Cooldown:         0,
	}
	entity1.AddComponent(expComp1)
	world.Update(0) // Process entity addition (entity1 now visible to queries)

	// First update - entity1 starts pending combo
	system.Update(0.1)

	// Create second entity with same expression (within sync window)
	entity2 := world.CreateEntity()
	expComp2 := &ExpressionComponent{
		ActiveExpression: ExpressionDance,
		ExpressionTime:   3.0,
		Cooldown:         0,
	}
	entity2.AddComponent(expComp2)
	world.Update(0) // Process entity addition (entity2 now visible to queries)

	// Second update - entity2 joins combo
	system.Update(0.1)

	// Check both entities have active combo
	activeCombo1 := system.GetActiveCombo(entity1.ID)
	activeCombo2 := system.GetActiveCombo(entity2.ID)

	if activeCombo1 == nil {
		t.Errorf("Entity1 should have active combo")
	}
	if activeCombo2 == nil {
		t.Errorf("Entity2 should have active combo")
	}
	if activeCombo1 != activeCombo2 {
		t.Errorf("Both entities should share the same combo")
	}

	// Check combo details
	if activeCombo1 != nil {
		if len(activeCombo1.ParticipantIDs) != 2 {
			t.Errorf("Combo should have 2 participants, got %d", len(activeCombo1.ParticipantIDs))
		}
		if activeCombo1.ExpressionType != ExpressionDance {
			t.Errorf("Combo expression type = %v, want %v", activeCombo1.ExpressionType, ExpressionDance)
		}
		if !activeCombo1.Active {
			t.Errorf("Combo should be active")
		}
	}
}

func TestExpressionComboSystem_DifferentExpressions(t *testing.T) {
	world := NewWorld()
	system := NewExpressionComboSystem(world)

	// Create entity with Wave expression
	entity1 := world.CreateEntity()
	world.Update(0) // Process entity addition
	expComp1 := &ExpressionComponent{
		ActiveExpression: ExpressionWave,
		ExpressionTime:   3.0,
		Cooldown:         0,
	}
	entity1.AddComponent(expComp1)

	// Create entity with Dance expression
	entity2 := world.CreateEntity()
	world.Update(0) // Process entity addition
	expComp2 := &ExpressionComponent{
		ActiveExpression: ExpressionDance,
		ExpressionTime:   3.0,
		Cooldown:         0,
	}
	entity2.AddComponent(expComp2)

	// Update system
	system.Update(0.1)

	// Different expressions shouldn't create combo
	activeCombo1 := system.GetActiveCombo(entity1.ID)
	activeCombo2 := system.GetActiveCombo(entity2.ID)

	if activeCombo1 != nil && activeCombo1.Active {
		t.Errorf("Different expressions shouldn't create active combo")
	}
	if activeCombo2 != nil && activeCombo2.Active {
		t.Errorf("Different expressions shouldn't create active combo")
	}

	// Should have 2 separate pending combos
	if len(system.pendingCombos) != 2 {
		t.Errorf("Should have 2 pending combos, got %d", len(system.pendingCombos))
	}
}

func TestExpressionComboSystem_ComboExpiration(t *testing.T) {
	world := NewWorld()
	system := NewExpressionComboSystem(world)

	// Create entity with expression
	entity := world.CreateEntity()
	world.Update(0) // Process entity addition
	expComp := &ExpressionComponent{
		ActiveExpression: ExpressionCheer,
		ExpressionTime:   3.0,
		Cooldown:         0,
	}
	entity.AddComponent(expComp)

	// Update system
	system.Update(0.1)

	// Check pending combo exists
	if len(system.pendingCombos) != 1 {
		t.Errorf("Should have 1 pending combo, got %d", len(system.pendingCombos))
	}

	// Wait past sync window (2.0 seconds + buffer)
	system.Update(2.5)

	// Pending combo should be cleaned up (single entity, not enough participants)
	if len(system.pendingCombos) != 0 {
		t.Errorf("Expired pending combo should be cleaned up, got %d", len(system.pendingCombos))
	}
}

func TestExpressionComboSystem_ComboFinalization(t *testing.T) {
	world := NewWorld()
	system := NewExpressionComboSystem(world)

	// Create two entities with same expression
	entity1 := world.CreateEntity()
	world.Update(0) // Process entity addition
	expComp1 := &ExpressionComponent{
		ActiveExpression: ExpressionLaugh,
		ExpressionTime:   3.0,
		Cooldown:         0,
	}
	entity1.AddComponent(expComp1)

	entity2 := world.CreateEntity()
	world.Update(0) // Process entity addition
	expComp2 := &ExpressionComponent{
		ActiveExpression: ExpressionLaugh,
		ExpressionTime:   3.0,
		Cooldown:         0,
	}
	entity2.AddComponent(expComp2)

	// Update to create combo
	system.Update(0.1)
	system.Update(0.1)

	// Wait past sync window to finalize
	system.Update(2.5)

	// Check that combo was recorded in history
	comboCount1 := system.GetComboCount(entity1.ID)
	comboCount2 := system.GetComboCount(entity2.ID)

	if comboCount1 != 1 {
		t.Errorf("Entity1 combo count = %d, want 1", comboCount1)
	}
	if comboCount2 != 1 {
		t.Errorf("Entity2 combo count = %d, want 1", comboCount2)
	}
}

func TestExpressionComboSystem_GetComboCount_NoComponent(t *testing.T) {
	world := NewWorld()
	system := NewExpressionComboSystem(world)

	// Create entity without combo component
	entity := world.CreateEntity()
	world.Update(0) // Process entity addition

	count := system.GetComboCount(entity.ID)
	if count != 0 {
		t.Errorf("GetComboCount() for entity without component = %d, want 0", count)
	}
}

func TestExpressionComboSystem_GetActiveCombo_InvalidEntity(t *testing.T) {
	world := NewWorld()
	system := NewExpressionComboSystem(world)

	combo := system.GetActiveCombo(999)
	if combo != nil {
		t.Errorf("GetActiveCombo() for invalid entity should return nil")
	}
}

func BenchmarkExpressionComboSystem_Update(b *testing.B) {
	world := NewWorld()
	system := NewExpressionComboSystem(world)

	// Create 10 entities with expressions
	for i := 0; i < 10; i++ {
		entity := world.CreateEntity()
		world.Update(0) // Process entity addition
		expComp := &ExpressionComponent{
			ActiveExpression: ExpressionDance,
			ExpressionTime:   3.0,
			Cooldown:         0,
		}
		entity.AddComponent(expComp)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		system.Update(0.016) // ~60 FPS
	}
}
