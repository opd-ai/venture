package learning

import (
	"testing"
	"time"
)

func TestNewCompanionLearningSystem(t *testing.T) {
	system := NewCompanionLearningSystem(time.Second)

	if system == nil {
		t.Fatal("NewCompanionLearningSystem returned nil")
	}
	if system.manager == nil {
		t.Error("manager not initialized")
	}
	if system.updateInterval != time.Second {
		t.Errorf("expected update interval 1s, got %v", system.updateInterval)
	}
}

func TestSystemUpdate(t *testing.T) {
	mockTime := &MockTimeProvider{CurrentTime: time.Unix(1000000, 0)}
	manager := NewManagerWithOptions(mockTime, nil)
	system := &CompanionLearningSystem{
		manager:        manager,
		updateInterval: 10 * time.Millisecond,
		timeProvider:   mockTime,
		lastUpdate:     time.Time{},
	}
	comp := manager.AddCompanion("test", 1.0)

	// Add some XP to a skill
	_ = comp.SkillTree.AddExperience("Basic Attack", 50.0, 1.0)

	// Mark skill as used recently
	comp.LastSkillUse["Basic Attack"] = mockTime.Now()

	// Update system
	time.Sleep(15 * time.Millisecond)
	system.Update(0.015)

	// Skill should not decay significantly since it was used recently
	skill := comp.SkillTree.Skills["Basic Attack"]
	if skill.Experience < 49.0 {
		t.Errorf("expected minimal decay, got XP %f", skill.Experience)
	}
}

func TestSystemUpdateSkillDecay(t *testing.T) {
	mockTime := &MockTimeProvider{CurrentTime: time.Unix(1000000, 0)}
	manager := NewManagerWithOptions(mockTime, nil)
	system := &CompanionLearningSystem{
		manager:        manager,
		updateInterval: 10 * time.Millisecond,
		timeProvider:   mockTime,
		lastUpdate:     time.Time{},
	}
	comp := manager.AddCompanion("test", 1.0)

	// Add some XP to a skill
	_ = comp.SkillTree.AddExperience("Basic Attack", 50.0, 1.0)

	// Mark skill as used long ago
	comp.LastSkillUse["Basic Attack"] = mockTime.Now().Add(-48 * time.Hour)

	// Update system
	time.Sleep(15 * time.Millisecond)
	system.Update(0.015)

	// Skill should decay slightly
	skill := comp.SkillTree.Skills["Basic Attack"]
	if skill.Experience >= 50.0 {
		t.Errorf("expected decay, but XP is still %f", skill.Experience)
	}
}

func TestGetManager(t *testing.T) {
	system := NewCompanionLearningSystem(time.Second)
	manager := system.GetManager()

	if manager == nil {
		t.Error("GetManager returned nil")
	}
	if manager != system.manager {
		t.Error("GetManager returned different manager instance")
	}
}

func TestRecordSkillUse(t *testing.T) {
	mockTime := &MockTimeProvider{CurrentTime: time.Unix(1000000, 0)}
	manager := NewManagerWithOptions(mockTime, nil)
	comp := manager.AddCompanion("test", 1.0)

	before := mockTime.Now()
	RecordSkillUse(comp, "Basic Attack", mockTime)
	after := mockTime.Now()

	lastUse, ok := comp.LastSkillUse["Basic Attack"]
	if !ok {
		t.Fatal("skill use not recorded")
	}

	if lastUse.Before(before) || lastUse.After(after) {
		t.Error("skill use timestamp out of expected range")
	}
}

func TestGetSkillBonus(t *testing.T) {
	manager := NewManager()
	comp := manager.AddCompanion("test", 1.0)

	// Level 0 skill
	bonus := GetSkillBonus(comp, "Basic Attack")
	if bonus != 1.0 {
		t.Errorf("expected bonus 1.0 for level 0, got %f", bonus)
	}

	// Level up the skill
	skill := comp.SkillTree.Skills["Basic Attack"]
	skill.Level = 5

	bonus = GetSkillBonus(comp, "Basic Attack")
	expected := 1.0 + (5.0 * 0.1)
	if bonus != expected {
		t.Errorf("expected bonus %f for level 5, got %f", expected, bonus)
	}
}

func TestGetSkillBonusNonexistent(t *testing.T) {
	manager := NewManager()
	comp := manager.AddCompanion("test", 1.0)

	bonus := GetSkillBonus(comp, "Nonexistent Skill")
	if bonus != 1.0 {
		t.Errorf("expected default bonus 1.0, got %f", bonus)
	}
}

func TestGetPersonalityInfluence(t *testing.T) {
	manager := NewManager()
	comp := manager.AddCompanion("test", 1.0)

	comp.Personality.Traits[TraitBrave] = 0.8

	influence := GetPersonalityInfluence(comp, TraitBrave)
	if influence != 0.8 {
		t.Errorf("expected influence 0.8, got %f", influence)
	}
}

func TestIsSkillMaxed(t *testing.T) {
	manager := NewManager()
	comp := manager.AddCompanion("test", 1.0)

	skill := comp.SkillTree.Skills["Basic Attack"]
	skill.Level = 0

	if IsSkillMaxed(comp, "Basic Attack") {
		t.Error("expected skill not maxed at level 0")
	}

	skill.Level = skill.MaxLevel

	if !IsSkillMaxed(comp, "Basic Attack") {
		t.Error("expected skill maxed at max level")
	}
}

func TestGetTotalSkillPoints(t *testing.T) {
	manager := NewManager()
	comp := manager.AddCompanion("test", 1.0)

	total := GetTotalSkillPoints(comp)
	if total != 0 {
		t.Errorf("expected 0 skill points spent, got %d", total)
	}

	// Learn a skill (costs 1 point)
	comp.SkillTree.AvailablePoints = 5
	_ = comp.SkillTree.LearnSkill("Basic Attack")
	comp.SkillTree.Skills["Basic Attack"].Level = 1

	total = GetTotalSkillPoints(comp)
	if total != 1 {
		t.Errorf("expected 1 skill point spent, got %d", total)
	}
}

func TestGetSkillsByType(t *testing.T) {
	manager := NewManager()
	comp := manager.AddCompanion("test", 1.0)

	combatSkills := GetSkillsByType(comp, SkillCombat)
	if len(combatSkills) == 0 {
		t.Error("expected combat skills, got none")
	}

	for _, skill := range combatSkills {
		if skill.Type != SkillCombat {
			t.Errorf("expected SkillCombat, got %s", skill.Type.String())
		}
	}
}

func TestGetMemorySummary(t *testing.T) {
	manager := NewManager()
	comp := manager.AddCompanion("test", 1.0)

	summary := GetMemorySummary(comp)
	if len(summary) == 0 {
		t.Error("expected non-empty summary")
	}

	mockTime := &MockTimeProvider{CurrentTime: time.Unix(1000000, 0)}
	comp.Memory.AddEvent(MemorableEvent{
		Type:      EventCombat,
		Timestamp: mockTime.Now(),
	})

	summary = GetMemorySummary(comp)
	if len(summary) == 0 {
		t.Error("expected non-empty summary with events")
	}
}

func TestCalculateLearningProgress(t *testing.T) {
	manager := NewManager()
	comp := manager.AddCompanion("test", 1.0)

	progress := CalculateLearningProgress(comp)
	if progress != 0.0 {
		t.Errorf("expected 0.0 progress initially, got %f", progress)
	}

	// Level up one skill
	skill := comp.SkillTree.Skills["Basic Attack"]
	skill.Level = 5

	progress = CalculateLearningProgress(comp)
	if progress <= 0.0 || progress >= 1.0 {
		t.Errorf("expected progress between 0 and 1, got %f", progress)
	}
}

func TestShouldLearnNewSkill(t *testing.T) {
	manager := NewManager()
	comp := manager.AddCompanion("test", 1.0)

	// No points available
	if ShouldLearnNewSkill(comp, "Basic Attack") {
		t.Error("expected false with no skill points")
	}

	// Add points and set aggressive personality
	comp.SkillTree.AvailablePoints = 5
	comp.Personality.Traits[TraitAggressive] = 0.9

	if !ShouldLearnNewSkill(comp, "Basic Attack") {
		t.Error("expected true for combat skill with aggressive personality")
	}
}

func TestBalanceTraits(t *testing.T) {
	system := NewCompanionLearningSystem(time.Second)
	pe := NewPersonalityEvolution()

	// Set unbalanced opposing traits
	pe.Traits[TraitCautious] = 0.9
	pe.Traits[TraitBrave] = 0.8

	system.balanceTraits(pe, TraitCautious, TraitBrave)

	sum := pe.Traits[TraitCautious] + pe.Traits[TraitBrave]
	if sum < 0.8 || sum > 1.2 {
		t.Errorf("expected sum between 0.8 and 1.2, got %f", sum)
	}
}

func TestNormalizeOpposingTraits(t *testing.T) {
	system := NewCompanionLearningSystem(time.Second)
	pe := NewPersonalityEvolution()

	// Set all opposing pairs to extreme values
	pe.Traits[TraitCautious] = 1.0
	pe.Traits[TraitBrave] = 0.9
	pe.Traits[TraitShy] = 1.0
	pe.Traits[TraitOutgoing] = 0.9

	system.normalizeOpposingTraits(pe)

	// All opposing pairs should be more balanced
	cautious := pe.Traits[TraitCautious]
	brave := pe.Traits[TraitBrave]
	if cautious+brave < 0.8 || cautious+brave > 1.2 {
		t.Errorf("expected balanced cautious/brave, got sum %f", cautious+brave)
	}
}

// Benchmarks

func BenchmarkSystemUpdate(b *testing.B) {
	system := NewCompanionLearningSystem(time.Millisecond)
	manager := system.GetManager()
	for i := 0; i < 10; i++ {
		manager.AddCompanion(string(rune(i)), 1.0)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		system.Update(0.016)
	}
}

func BenchmarkGetSkillBonus(b *testing.B) {
	manager := NewManager()
	comp := manager.AddCompanion("test", 1.0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GetSkillBonus(comp, "Basic Attack")
	}
}

func BenchmarkCalculateLearningProgress(b *testing.B) {
	manager := NewManager()
	comp := manager.AddCompanion("test", 1.0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CalculateLearningProgress(comp)
	}
}

func BenchmarkShouldLearnNewSkill(b *testing.B) {
	manager := NewManager()
	comp := manager.AddCompanion("test", 1.0)
	comp.SkillTree.AvailablePoints = 10

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ShouldLearnNewSkill(comp, "Basic Attack")
	}
}

// === Edge case tests for improved coverage ===

// TestRecordSkillUse_NilComponent tests RecordSkillUse with nil component.
func TestRecordSkillUse_NilComponent(t *testing.T) {
	// Should not panic with nil component
	RecordSkillUse(nil, "Basic Attack", nil)
}

// TestRecordSkillUse_NilLastSkillUse tests RecordSkillUse with nil map.
func TestRecordSkillUse_NilLastSkillUse(t *testing.T) {
	comp := &CompanionLearningComponent{
		LastSkillUse: nil,
	}
	// Should not panic with nil LastSkillUse map
	RecordSkillUse(comp, "Basic Attack", nil)
}

// TestGetPersonalityInfluence_NilComponent tests with nil component.
func TestGetPersonalityInfluence_NilComponent(t *testing.T) {
	influence := GetPersonalityInfluence(nil, TraitBrave)
	if influence != 0.5 {
		t.Errorf("expected default 0.5 for nil component, got %f", influence)
	}
}

// TestGetPersonalityInfluence_NilPersonality tests with nil personality.
func TestGetPersonalityInfluence_NilPersonality(t *testing.T) {
	comp := &CompanionLearningComponent{
		Personality: nil,
	}
	influence := GetPersonalityInfluence(comp, TraitBrave)
	if influence != 0.5 {
		t.Errorf("expected default 0.5 for nil personality, got %f", influence)
	}
}

// TestGetPersonalityInfluence_MissingTrait tests with trait not in map.
func TestGetPersonalityInfluence_MissingTrait(t *testing.T) {
	manager := NewManager()
	comp := manager.AddCompanion("test", 1.0)

	// Remove the trait from the map to test missing trait case
	delete(comp.Personality.Traits, TraitPractical)

	influence := GetPersonalityInfluence(comp, TraitPractical)
	if influence != 0.5 {
		t.Errorf("expected default 0.5 for missing trait, got %f", influence)
	}
}

// TestIsSkillMaxed_NilComponent tests with nil component.
func TestIsSkillMaxed_NilComponent(t *testing.T) {
	result := IsSkillMaxed(nil, "Basic Attack")
	if result {
		t.Error("expected false for nil component")
	}
}

// TestIsSkillMaxed_NilSkillTree tests with nil skill tree.
func TestIsSkillMaxed_NilSkillTree(t *testing.T) {
	comp := &CompanionLearningComponent{
		SkillTree: nil,
	}
	result := IsSkillMaxed(comp, "Basic Attack")
	if result {
		t.Error("expected false for nil skill tree")
	}
}

// TestIsSkillMaxed_NonexistentSkill tests with skill not in tree.
func TestIsSkillMaxed_NonexistentSkill(t *testing.T) {
	manager := NewManager()
	comp := manager.AddCompanion("test", 1.0)

	result := IsSkillMaxed(comp, "NonexistentSkill")
	if result {
		t.Error("expected false for non-existent skill")
	}
}

// TestGetSkillBonus_NilComponent tests with nil component.
func TestGetSkillBonus_NilComponent(t *testing.T) {
	bonus := GetSkillBonus(nil, "Basic Attack")
	if bonus != 1.0 {
		t.Errorf("expected default 1.0 for nil component, got %f", bonus)
	}
}

// TestGetSkillBonus_NilSkillTree tests with nil skill tree.
func TestGetSkillBonus_NilSkillTree(t *testing.T) {
	comp := &CompanionLearningComponent{
		SkillTree: nil,
	}
	bonus := GetSkillBonus(comp, "Basic Attack")
	if bonus != 1.0 {
		t.Errorf("expected default 1.0 for nil skill tree, got %f", bonus)
	}
}

// TestGetTotalSkillPoints_NilComponent tests with nil component.
func TestGetTotalSkillPoints_NilComponent(t *testing.T) {
	total := GetTotalSkillPoints(nil)
	if total != 0 {
		t.Errorf("expected 0 for nil component, got %d", total)
	}
}

// TestGetTotalSkillPoints_NilSkillTree tests with nil skill tree.
func TestGetTotalSkillPoints_NilSkillTree(t *testing.T) {
	comp := &CompanionLearningComponent{
		SkillTree: nil,
	}
	total := GetTotalSkillPoints(comp)
	if total != 0 {
		t.Errorf("expected 0 for nil skill tree, got %d", total)
	}
}

// TestGetSkillsByType_NilComponent tests with nil component.
func TestGetSkillsByType_NilComponent(t *testing.T) {
	skills := GetSkillsByType(nil, SkillCombat)
	if skills != nil {
		t.Error("expected nil for nil component")
	}
}

// TestGetSkillsByType_NilSkillTree tests with nil skill tree.
func TestGetSkillsByType_NilSkillTree(t *testing.T) {
	comp := &CompanionLearningComponent{
		SkillTree: nil,
	}
	skills := GetSkillsByType(comp, SkillCombat)
	if skills != nil {
		t.Error("expected nil for nil skill tree")
	}
}

// TestGetMemorySummary_NilComponent tests with nil component.
func TestGetMemorySummary_NilComponent(t *testing.T) {
	summary := GetMemorySummary(nil)
	expected := "No companion data"
	if summary != expected {
		t.Errorf("expected %q, got %q", expected, summary)
	}
}

// TestGetMemorySummary_NilMemory tests with nil memory.
func TestGetMemorySummary_NilMemory(t *testing.T) {
	comp := &CompanionLearningComponent{
		Memory: nil,
	}
	summary := GetMemorySummary(comp)
	expected := "No companion data"
	if summary != expected {
		t.Errorf("expected %q, got %q", expected, summary)
	}
}

// TestCalculateLearningProgress_NilComponent tests with nil component.
func TestCalculateLearningProgress_NilComponent(t *testing.T) {
	progress := CalculateLearningProgress(nil)
	if progress != 0.0 {
		t.Errorf("expected 0.0 for nil component, got %f", progress)
	}
}

// TestCalculateLearningProgress_NilSkillTree tests with nil skill tree.
func TestCalculateLearningProgress_NilSkillTree(t *testing.T) {
	comp := &CompanionLearningComponent{
		SkillTree: nil,
	}
	progress := CalculateLearningProgress(comp)
	if progress != 0.0 {
		t.Errorf("expected 0.0 for nil skill tree, got %f", progress)
	}
}

// TestShouldLearnNewSkill_NilComponent tests with nil component.
func TestShouldLearnNewSkill_NilComponent(t *testing.T) {
	result := ShouldLearnNewSkill(nil, "Basic Attack")
	if result {
		t.Error("expected false for nil component")
	}
}

// TestShouldLearnNewSkill_NilSkillTree tests with nil skill tree.
func TestShouldLearnNewSkill_NilSkillTree(t *testing.T) {
	comp := &CompanionLearningComponent{
		SkillTree:   nil,
		Personality: NewPersonalityEvolution(),
	}
	result := ShouldLearnNewSkill(comp, "Basic Attack")
	if result {
		t.Error("expected false for nil skill tree")
	}
}

// TestShouldLearnNewSkill_NilPersonality tests with nil personality.
func TestShouldLearnNewSkill_NilPersonality(t *testing.T) {
	manager := NewManager()
	comp := manager.AddCompanion("test", 1.0)
	comp.Personality = nil
	comp.SkillTree.AvailablePoints = 5

	result := ShouldLearnNewSkill(comp, "Basic Attack")
	if result {
		t.Error("expected false for nil personality")
	}
}

// TestShouldLearnNewSkill_SkillNotInTree tests with non-existent skill.
func TestShouldLearnNewSkill_SkillNotInTree(t *testing.T) {
	manager := NewManager()
	comp := manager.AddCompanion("test", 1.0)
	comp.SkillTree.AvailablePoints = 5
	comp.Personality.Traits[TraitAggressive] = 0.9

	// Try to learn a skill that doesn't exist
	result := ShouldLearnNewSkill(comp, "NonexistentSkill")
	if result {
		t.Error("expected false for non-existent skill")
	}
}

// TestShouldLearnNewSkill_AllSkillTypes tests different skill type branches.
func TestShouldLearnNewSkill_AllSkillTypes(t *testing.T) {
	tests := []struct {
		name      string
		skillName string
		skillType SkillType
		trait     PersonalityTrait
		expected  bool
	}{
		{"Combat/Aggressive", "Basic Attack", SkillCombat, TraitAggressive, true},
		{"Combat/Brave", "Basic Attack", SkillCombat, TraitBrave, true},
		{"Defense/Cautious", "Block", SkillDefense, TraitCautious, true},
		{"Defense/Pacifist", "Block", SkillDefense, TraitPacifist, true},
		{"Social/Outgoing", "Persuasion", SkillSocial, TraitOutgoing, true},
		{"Social/Loyal", "Persuasion", SkillSocial, TraitLoyal, true},
		{"Utility/Curious", "Gather", SkillUtility, TraitCurious, true},
		{"Utility/Practical", "Gather", SkillUtility, TraitPractical, true},
		{"Stealth/Independent", "Sneak", SkillStealth, TraitIndependent, true},
		{"Healing/Pacifist", "First Aid", SkillHealing, TraitPacifist, true},
		{"Magic/Curious", "Mana Control", SkillMagic, TraitCurious, true},
		{"Crafting/Practical", "Apprentice Smith", SkillCrafting, TraitPractical, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewManager()
			comp := manager.AddCompanion("test", 1.0)
			comp.SkillTree.AvailablePoints = 10

			// Set the dominant trait
			for trait := range comp.Personality.Traits {
				comp.Personality.Traits[trait] = 0.1
			}
			comp.Personality.Traits[tt.trait] = 0.9

			result := ShouldLearnNewSkill(comp, tt.skillName)
			if result != tt.expected {
				t.Errorf("ShouldLearnNewSkill(%q) with trait %s = %v, want %v",
					tt.skillName, tt.trait.String(), result, tt.expected)
			}
		})
	}
}
