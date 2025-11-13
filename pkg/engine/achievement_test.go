package engine

import (
	"testing"
)

func TestAchievementType_String(t *testing.T) {
	tests := []struct {
		name     string
		achType  AchievementType
		expected string
	}{
		{"FirstExpression", AchievementFirstExpression, "First Expression"},
		{"ExpressionMaster", AchievementExpressionMaster, "Expression Master"},
		{"ComboStarter", AchievementComboStarter, "Combo Starter"},
		{"ComboExpert", AchievementComboExpert, "Combo Expert"},
		{"ComboLegend", AchievementComboLegend, "Combo Legend"},
		{"SocialButterfly", AchievementSocialButterfly, "Social Butterfly"},
		{"RareExpression", AchievementRareExpression, "Rare Expression"},
		{"GroupPerformer", AchievementGroupPerformer, "Group Performer"},
		{"Unknown", AchievementType(999), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.achType.String()
			if result != tt.expected {
				t.Errorf("AchievementType.String() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestAchievementComponent_Type(t *testing.T) {
	comp := AchievementComponent{}
	if comp.Type() != "achievement" {
		t.Errorf("AchievementComponent.Type() = %q, want %q", comp.Type(), "achievement")
	}
}

func TestAchievementComponent_HasAchievement(t *testing.T) {
	comp := &AchievementComponent{
		Achievements: []Achievement{
			{Type: AchievementFirstExpression, UnlockedAt: 12345, Description: "Test"},
			{Type: AchievementComboStarter, UnlockedAt: 12346, Description: "Test2"},
		},
	}

	// Test existing achievement
	if !comp.HasAchievement(AchievementFirstExpression) {
		t.Errorf("HasAchievement() should return true for existing achievement")
	}

	// Test non-existing achievement
	if comp.HasAchievement(AchievementExpressionMaster) {
		t.Errorf("HasAchievement() should return false for non-existing achievement")
	}
}

func TestNewAchievementSystem(t *testing.T) {
	world := NewWorld()
	system := NewAchievementSystem(world)

	if system == nil {
		t.Errorf("NewAchievementSystem() returned nil")
		return
	}
	if system.world != world {
		t.Errorf("NewAchievementSystem() world not set correctly")
	}
}

func TestAchievementSystem_OnExpressionUsed_FirstExpression(t *testing.T) {
	world := NewWorld()
	system := NewAchievementSystem(world)

	entity := world.CreateEntity()
	world.Update(0) // Process entity addition

	// Use first expression
	system.OnExpressionUsed(entity.ID, ExpressionWave)

	// Check achievement unlocked
	achievements := system.GetAchievements(entity.ID)
	if len(achievements) != 1 {
		t.Errorf("Should have 1 achievement, got %d", len(achievements))
		return
	}
	if achievements[0].Type != AchievementFirstExpression {
		t.Errorf("Achievement type = %v, want %v", achievements[0].Type, AchievementFirstExpression)
	}
}

func TestAchievementSystem_OnExpressionUsed_ExpressionMaster(t *testing.T) {
	world := NewWorld()
	system := NewAchievementSystem(world)

	entity := world.CreateEntity()
	world.Update(0) // Process entity addition

	// Use all 12 expressions
	expressions := []ExpressionType{
		ExpressionWave, ExpressionCheer, ExpressionDance, ExpressionLaugh,
		ExpressionCry, ExpressionSit, ExpressionPoint, ExpressionSalute,
		ExpressionShrug, ExpressionThumbsUp, ExpressionFacepalm, ExpressionSleep,
	}

	for _, exp := range expressions {
		system.OnExpressionUsed(entity.ID, exp)
	}

	// Check Expression Master unlocked
	achCompRaw, _ := entity.GetComponent("achievement")
	achComp := achCompRaw.(*AchievementComponent)

	if !achComp.HasAchievement(AchievementExpressionMaster) {
		t.Errorf("Should have Expression Master achievement")
	}
}

func TestAchievementSystem_OnExpressionUsed_SocialButterfly(t *testing.T) {
	world := NewWorld()
	system := NewAchievementSystem(world)

	entity := world.CreateEntity()
	world.Update(0) // Process entity addition

	// Use 50 expressions
	for i := 0; i < 50; i++ {
		system.OnExpressionUsed(entity.ID, ExpressionWave)
	}

	// Check Social Butterfly unlocked
	achCompRaw, _ := entity.GetComponent("achievement")
	achComp := achCompRaw.(*AchievementComponent)

	if !achComp.HasAchievement(AchievementSocialButterfly) {
		t.Errorf("Should have Social Butterfly achievement")
	}
}

func TestAchievementSystem_OnComboCompleted_ComboStarter(t *testing.T) {
	world := NewWorld()
	system := NewAchievementSystem(world)

	entity := world.CreateEntity()
	world.Update(0) // Process entity addition
	
	// Add combo component with 1 combo
	comboComp := &ExpressionComboComponent{
		TotalCombos: 1,
	}
	entity.AddComponent(comboComp)

	// Trigger combo completed
	combo := &ExpressionCombo{
		ComboType:      ComboSynchronized,
		ParticipantIDs: []uint64{entity.ID, 2},
		ExpressionType: ExpressionDance,
	}
	system.OnComboCompleted(entity.ID, combo)

	// Check Combo Starter unlocked
	achCompRaw, _ := entity.GetComponent("achievement")
	achComp := achCompRaw.(*AchievementComponent)

	if !achComp.HasAchievement(AchievementComboStarter) {
		t.Errorf("Should have Combo Starter achievement")
	}
}

func TestAchievementSystem_OnComboCompleted_ComboExpert(t *testing.T) {
	world := NewWorld()
	system := NewAchievementSystem(world)

	entity := world.CreateEntity()
	world.Update(0) // Process entity addition
	
	// Add combo component with 10 combos
	comboComp := &ExpressionComboComponent{
		TotalCombos: 10,
	}
	entity.AddComponent(comboComp)

	// Trigger combo completed
	combo := &ExpressionCombo{
		ComboType:      ComboSynchronized,
		ParticipantIDs: []uint64{entity.ID, 2},
		ExpressionType: ExpressionDance,
	}
	system.OnComboCompleted(entity.ID, combo)

	// Check Combo Expert unlocked
	achCompRaw, _ := entity.GetComponent("achievement")
	achComp := achCompRaw.(*AchievementComponent)

	if !achComp.HasAchievement(AchievementComboExpert) {
		t.Errorf("Should have Combo Expert achievement")
	}
}

func TestAchievementSystem_OnComboCompleted_ComboLegend(t *testing.T) {
	world := NewWorld()
	system := NewAchievementSystem(world)

	entity := world.CreateEntity()
	world.Update(0) // Process entity addition
	
	// Add combo component with 100 combos
	comboComp := &ExpressionComboComponent{
		TotalCombos: 100,
	}
	entity.AddComponent(comboComp)

	// Trigger combo completed
	combo := &ExpressionCombo{
		ComboType:      ComboSynchronized,
		ParticipantIDs: []uint64{entity.ID, 2},
		ExpressionType: ExpressionDance,
	}
	system.OnComboCompleted(entity.ID, combo)

	// Check Combo Legend unlocked
	achCompRaw, _ := entity.GetComponent("achievement")
	achComp := achCompRaw.(*AchievementComponent)

	if !achComp.HasAchievement(AchievementComboLegend) {
		t.Errorf("Should have Combo Legend achievement")
	}
}

func TestAchievementSystem_OnComboCompleted_GroupPerformer(t *testing.T) {
	world := NewWorld()
	system := NewAchievementSystem(world)

	entity := world.CreateEntity()
	world.Update(0) // Process entity addition
	
	// Add combo component
	comboComp := &ExpressionComboComponent{
		TotalCombos: 1,
	}
	entity.AddComponent(comboComp)

	// Trigger combo with 5 participants
	combo := &ExpressionCombo{
		ComboType:      ComboSynchronized,
		ParticipantIDs: []uint64{entity.ID, 2, 3, 4, 5},
		ExpressionType: ExpressionDance,
	}
	system.OnComboCompleted(entity.ID, combo)

	// Check Group Performer unlocked
	achCompRaw, _ := entity.GetComponent("achievement")
	achComp := achCompRaw.(*AchievementComponent)

	if !achComp.HasAchievement(AchievementGroupPerformer) {
		t.Errorf("Should have Group Performer achievement")
	}
}

func TestAchievementSystem_GetAchievements_NoComponent(t *testing.T) {
	world := NewWorld()
	system := NewAchievementSystem(world)

	entity := world.CreateEntity()
	world.Update(0) // Process entity addition

	achievements := system.GetAchievements(entity.ID)
	if achievements != nil {
		t.Errorf("GetAchievements() should return nil for entity without component")
	}
}

func TestAchievementSystem_GetAchievementCount(t *testing.T) {
	world := NewWorld()
	system := NewAchievementSystem(world)

	entity := world.CreateEntity()
	world.Update(0) // Process entity addition

	// Add some achievements
	system.OnExpressionUsed(entity.ID, ExpressionWave)
	system.OnExpressionUsed(entity.ID, ExpressionCheer)

	count := system.GetAchievementCount(entity.ID)
	if count < 1 {
		t.Errorf("GetAchievementCount() = %d, want >= 1", count)
	}
}

func TestAchievementSystem_NoDoubleUnlock(t *testing.T) {
	world := NewWorld()
	system := NewAchievementSystem(world)

	entity := world.CreateEntity()
	world.Update(0) // Process entity addition

	// Use expression twice
	system.OnExpressionUsed(entity.ID, ExpressionWave)
	system.OnExpressionUsed(entity.ID, ExpressionCheer)
	
	// Should only have First Expression once
	achCompRaw, _ := entity.GetComponent("achievement")
	achComp := achCompRaw.(*AchievementComponent)
	
	firstExpCount := 0
	for _, ach := range achComp.Achievements {
		if ach.Type == AchievementFirstExpression {
			firstExpCount++
		}
	}

	if firstExpCount != 1 {
		t.Errorf("First Expression should only be unlocked once, got %d times", firstExpCount)
	}
}

func BenchmarkAchievementSystem_OnExpressionUsed(b *testing.B) {
	world := NewWorld()
	system := NewAchievementSystem(world)

	entity := world.CreateEntity()
	world.Update(0) // Process entity addition

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		system.OnExpressionUsed(entity.ID, ExpressionWave)
	}
}
