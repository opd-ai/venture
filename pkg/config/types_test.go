package config

import (
	"testing"
)

func TestCharacterClass_String(t *testing.T) {
	tests := []struct {
		name  string
		class CharacterClass
		want  string
	}{
		{"Warrior", ClassWarrior, "Warrior"},
		{"Mage", ClassMage, "Mage"},
		{"Rogue", ClassRogue, "Rogue"},
		{"Ranger", ClassRanger, "Ranger"},
		{"Cleric", ClassCleric, "Cleric"},
		{"Necromancer", ClassNecromancer, "Necromancer"},
		{"Battlemage", ClassBattlemage, "Battlemage"},
		{"Spellblade", ClassSpellblade, "Spellblade"},
		{"Paladin", ClassPaladin, "Paladin"},
		{"Monk", ClassMonk, "Monk"},
		{"Death Knight", ClassDeathKnight, "Death Knight"},
		{"Witch Hunter", ClassWitchHunter, "Witch Hunter"},
		{"Beastlord", ClassBeastlord, "Beastlord"},
		{"Arcane Archer", ClassArcaneArcher, "Arcane Archer"},
		{"Shadow Priest", ClassShadowPriest, "Shadow Priest"},
		{"Druid", ClassDruid, "Druid"},
		{"Inquisitor", ClassInquisitor, "Inquisitor"},
		{"Blood Knight", ClassBloodKnight, "Blood Knight"},
		{"Mystic", ClassMystic, "Mystic"},
		{"Warlock", ClassWarlock, "Warlock"},
		{"Ninja", ClassNinja, "Ninja"},
		{"Unknown", CharacterClass(999), "Unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.class.String(); got != tt.want {
				t.Errorf("CharacterClass.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCharacterClass_Description(t *testing.T) {
	tests := []struct {
		name  string
		class CharacterClass
	}{
		{"Warrior", ClassWarrior},
		{"Mage", ClassMage},
		{"Rogue", ClassRogue},
		{"Ranger", ClassRanger},
		{"Cleric", ClassCleric},
		{"Necromancer", ClassNecromancer},
		{"Battlemage", ClassBattlemage},
		{"Spellblade", ClassSpellblade},
		{"Paladin", ClassPaladin},
		{"Monk", ClassMonk},
		{"Death Knight", ClassDeathKnight},
		{"Witch Hunter", ClassWitchHunter},
		{"Beastlord", ClassBeastlord},
		{"Arcane Archer", ClassArcaneArcher},
		{"Shadow Priest", ClassShadowPriest},
		{"Druid", ClassDruid},
		{"Inquisitor", ClassInquisitor},
		{"Blood Knight", ClassBloodKnight},
		{"Mystic", ClassMystic},
		{"Warlock", ClassWarlock},
		{"Ninja", ClassNinja},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			desc := tt.class.Description()
			if desc == "" {
				t.Errorf("CharacterClass.Description() returned empty string for %v", tt.class)
			}
			if len(desc) < 10 {
				t.Errorf("CharacterClass.Description() too short: %v", desc)
			}
		})
	}
}

func TestCharacterClass_UnknownDescription(t *testing.T) {
	unknown := CharacterClass(999)
	if got := unknown.Description(); got != "Unknown class" {
		t.Errorf("CharacterClass.Description() for unknown = %v, want 'Unknown class'", got)
	}
}
