package engine

import (
	"testing"
)

func TestSkillInheritanceComponent_Type(t *testing.T) {
	comp := NewSkillInheritanceComponent(5, 0.1)
	if comp.Type() != "skillinheritance" {
		t.Errorf("expected 'skillinheritance', got '%s'", comp.Type())
	}
}

func TestNewSkillInheritanceComponent(t *testing.T) {
	comp := NewSkillInheritanceComponent(5, 0.15)
	
	if comp.MaxSkills != 5 {
		t.Errorf("expected MaxSkills 5, got %d", comp.MaxSkills)
	}
	if comp.LearningRate != 0.15 {
		t.Errorf("expected LearningRate 0.15, got %f", comp.LearningRate)
	}
	if comp.RequiredLoyalty != 50.0 {
		t.Errorf("expected RequiredLoyalty 50.0, got %f", comp.RequiredLoyalty)
	}
	if len(comp.LearnedSkills) != 0 {
		t.Errorf("expected empty learned skills, got %d", len(comp.LearnedSkills))
	}
}

func TestSkillInheritanceComponent_CanLearnSkill(t *testing.T) {
	tests := []struct {
		name          string
		maxSkills     int
		existingCount int
		skillID       string
		want          bool
	}{
		{"can learn first skill", 5, 0, "fireball", true},
		{"can learn within limit", 5, 3, "ice_shard", true},
		{"cannot learn at max", 3, 3, "lightning", false},
		{"can continue learning existing", 3, 3, "existing_skill", true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := NewSkillInheritanceComponent(tt.maxSkills, 0.1)
			
			// Add existing skills
			for i := 0; i < tt.existingCount; i++ {
				if i == tt.existingCount-1 && tt.skillID == "existing_skill" {
					comp.LearnedSkills["existing_skill"] = 0.5
				} else {
					comp.LearnedSkills[string(rune('a'+i))] = 0.3
				}
			}
			
			got := comp.CanLearnSkill(tt.skillID)
			if got != tt.want {
				t.Errorf("CanLearnSkill() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSkillInheritanceComponent_AddSkillProgress(t *testing.T) {
	tests := []struct {
		name            string
		initialProgress float64
		addAmount       float64
		wantProgress    float64
		wantFullyLearned bool
	}{
		{"learn from zero", 0.0, 0.3, 0.3, false},
		{"continue learning", 0.5, 0.3, 0.8, false},
		{"reach 1.0 exactly", 0.7, 0.3, 1.0, true},
		{"exceed 1.0 clamped", 0.9, 0.5, 1.0, true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := NewSkillInheritanceComponent(5, 0.1)
			
			if tt.initialProgress > 0 {
				comp.LearnedSkills["test_skill"] = tt.initialProgress
			}
			
			fullyLearned := comp.AddSkillProgress("test_skill", tt.addAmount)
			gotProgress := comp.GetSkillProgress("test_skill")
			
			if fullyLearned != tt.wantFullyLearned {
				t.Errorf("AddSkillProgress() returned %v, want %v", fullyLearned, tt.wantFullyLearned)
			}
			
			if gotProgress < tt.wantProgress-0.01 || gotProgress > tt.wantProgress+0.01 {
				t.Errorf("progress = %f, want %f", gotProgress, tt.wantProgress)
			}
		})
	}
}

func TestSkillInheritanceComponent_IsSkillActive(t *testing.T) {
	comp := NewSkillInheritanceComponent(5, 0.1)
	
	// Learn a skill fully
	comp.AddSkillProgress("fireball", 1.0)
	
	if !comp.IsSkillActive("fireball") {
		t.Error("expected fireball to be active after full learning")
	}
	
	if comp.IsSkillActive("ice_shard") {
		t.Error("expected ice_shard to not be active")
	}
}

func TestSkillInheritanceComponent_GetCounts(t *testing.T) {
	comp := NewSkillInheritanceComponent(5, 0.1)
	
	// Add partial learning
	comp.AddSkillProgress("skill1", 0.5)
	comp.AddSkillProgress("skill2", 0.8)
	
	// Add full learning
	comp.AddSkillProgress("skill3", 1.0)
	comp.AddSkillProgress("skill4", 1.0)
	
	if comp.GetLearnedSkillCount() != 4 {
		t.Errorf("expected 4 learned skills, got %d", comp.GetLearnedSkillCount())
	}
	
	if comp.GetActiveSkillCount() != 2 {
		t.Errorf("expected 2 active skills, got %d", comp.GetActiveSkillCount())
	}
}

func TestBondingPerk_String(t *testing.T) {
	tests := []struct {
		perk BondingPerk
		want string
	}{
		{PerkExtraHealth, "Extra Health"},
		{PerkExtraDamage, "Extra Damage"},
		{PerkFasterLearning, "Faster Learning"},
		{PerkLoyalGuard, "Loyal Guard"},
		{PerkSharedExperience, "Shared Experience"},
		{PerkAutoRevive, "Auto Revive"},
		{PerkNone, "None"},
	}
	
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.perk.String()
			if got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCompanionComponent_HasPerk(t *testing.T) {
	comp := &CompanionComponent{
		BondingPerks: []BondingPerk{PerkExtraHealth, PerkFasterLearning},
	}
	
	if !comp.HasPerk(PerkExtraHealth) {
		t.Error("expected HasPerk(PerkExtraHealth) to be true")
	}
	
	if !comp.HasPerk(PerkFasterLearning) {
		t.Error("expected HasPerk(PerkFasterLearning) to be true")
	}
	
	if comp.HasPerk(PerkExtraDamage) {
		t.Error("expected HasPerk(PerkExtraDamage) to be false")
	}
}

func TestCompanionComponent_AddPerk(t *testing.T) {
	comp := &CompanionComponent{
		BondingPerks: []BondingPerk{},
	}
	
	// Add first perk
	comp.AddPerk(PerkExtraHealth)
	if !comp.HasPerk(PerkExtraHealth) {
		t.Error("expected perk to be added")
	}
	if len(comp.BondingPerks) != 1 {
		t.Errorf("expected 1 perk, got %d", len(comp.BondingPerks))
	}
	
	// Try to add same perk again
	comp.AddPerk(PerkExtraHealth)
	if len(comp.BondingPerks) != 1 {
		t.Errorf("expected 1 perk (duplicate rejected), got %d", len(comp.BondingPerks))
	}
	
	// Add different perk
	comp.AddPerk(PerkFasterLearning)
	if len(comp.BondingPerks) != 2 {
		t.Errorf("expected 2 perks, got %d", len(comp.BondingPerks))
	}
}

func TestCompanionComponent_Permadeath(t *testing.T) {
	// Test permadeath flag
	comp := &CompanionComponent{
		Permadeath: true,
	}
	
	if !comp.Permadeath {
		t.Error("expected Permadeath to be true")
	}
	
	comp.Permadeath = false
	if comp.Permadeath {
		t.Error("expected Permadeath to be false after setting")
	}
}
