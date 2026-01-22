package learning

import (
	"testing"
)

// TestCompanionLearningComponent_Type tests the ECS component Type method.
func TestCompanionLearningComponent_Type(t *testing.T) {
	comp := &CompanionLearningComponent{}
	got := comp.Type()
	want := "companion_learning"
	if got != want {
		t.Errorf("CompanionLearningComponent.Type() = %q, want %q", got, want)
	}
}

// TestEventType_String tests all EventType String() method cases.
func TestEventType_String(t *testing.T) {
	tests := []struct {
		name     string
		event    EventType
		expected string
	}{
		{"EventCombat", EventCombat, "Combat"},
		{"EventDialog", EventDialog, "Dialog"},
		{"EventTrade", EventTrade, "Trade"},
		{"EventQuest", EventQuest, "Quest"},
		{"EventExploration", EventExploration, "Exploration"},
		{"EventCrafting", EventCrafting, "Crafting"},
		{"EventDeath", EventDeath, "Death"},
		{"EventRevival", EventRevival, "Revival"},
		{"EventGift", EventGift, "Gift"},
		{"EventBetrayal", EventBetrayal, "Betrayal"},
		{"Unknown", EventType(999), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.event.String()
			if got != tt.expected {
				t.Errorf("EventType(%d).String() = %q, want %q", tt.event, got, tt.expected)
			}
		})
	}
}

// TestSkillType_String tests all SkillType String() method cases.
func TestSkillType_String(t *testing.T) {
	tests := []struct {
		name     string
		skill    SkillType
		expected string
	}{
		{"SkillCombat", SkillCombat, "Combat"},
		{"SkillUtility", SkillUtility, "Utility"},
		{"SkillSocial", SkillSocial, "Social"},
		{"SkillCrafting", SkillCrafting, "Crafting"},
		{"SkillMagic", SkillMagic, "Magic"},
		{"SkillDefense", SkillDefense, "Defense"},
		{"SkillHealing", SkillHealing, "Healing"},
		{"SkillStealth", SkillStealth, "Stealth"},
		{"Unknown", SkillType(999), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.skill.String()
			if got != tt.expected {
				t.Errorf("SkillType(%d).String() = %q, want %q", tt.skill, got, tt.expected)
			}
		})
	}
}

// TestPersonalityTrait_String tests all PersonalityTrait String() method cases.
func TestPersonalityTrait_String(t *testing.T) {
	tests := []struct {
		name     string
		trait    PersonalityTrait
		expected string
	}{
		{"TraitCautious", TraitCautious, "Cautious"},
		{"TraitBrave", TraitBrave, "Brave"},
		{"TraitShy", TraitShy, "Shy"},
		{"TraitOutgoing", TraitOutgoing, "Outgoing"},
		{"TraitAggressive", TraitAggressive, "Aggressive"},
		{"TraitPacifist", TraitPacifist, "Pacifist"},
		{"TraitLoyal", TraitLoyal, "Loyal"},
		{"TraitIndependent", TraitIndependent, "Independent"},
		{"TraitCurious", TraitCurious, "Curious"},
		{"TraitPractical", TraitPractical, "Practical"},
		{"Unknown", PersonalityTrait(999), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.trait.String()
			if got != tt.expected {
				t.Errorf("PersonalityTrait(%d).String() = %q, want %q", tt.trait, got, tt.expected)
			}
		})
	}
}
