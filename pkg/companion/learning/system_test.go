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
	system := NewCompanionLearningSystem(10 * time.Millisecond)
	manager := system.GetManager()
	comp := manager.AddCompanion("test", 1.0)

	// Add some XP to a skill
	_ = comp.SkillTree.AddExperience("Basic Attack", 50.0, 1.0)

	// Mark skill as used recently
	comp.LastSkillUse["Basic Attack"] = time.Now()

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
	system := NewCompanionLearningSystem(10 * time.Millisecond)
	manager := system.GetManager()
	comp := manager.AddCompanion("test", 1.0)

	// Add some XP to a skill
	_ = comp.SkillTree.AddExperience("Basic Attack", 50.0, 1.0)

	// Mark skill as used long ago
	comp.LastSkillUse["Basic Attack"] = time.Now().Add(-48 * time.Hour)

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
	manager := NewManager()
	comp := manager.AddCompanion("test", 1.0)

	before := time.Now()
	RecordSkillUse(comp, "Basic Attack")
	after := time.Now()

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

	comp.Memory.AddEvent(MemorableEvent{
		Type:      EventCombat,
		Timestamp: time.Now(),
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
