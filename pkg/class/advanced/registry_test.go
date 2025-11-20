package advanced

import (
	"testing"
)

func TestGetClassDefinition(t *testing.T) {
	tests := []struct {
		name    string
		classID ClassID
		wantErr bool
	}{
		{"valid warrior", ClassWarrior, false},
		{"valid mage", ClassMage, false},
		{"valid rogue", ClassRogue, false},
		{"invalid class", ClassID("invalid"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			def, err := GetClassDefinition(tt.classID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetClassDefinition() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if def.ID != tt.classID {
					t.Errorf("GetClassDefinition() ID = %v, want %v", def.ID, tt.classID)
				}
				if def.Name == "" {
					t.Error("GetClassDefinition() Name is empty")
				}
			}
		})
	}
}

func TestGetPrestigeClassDefinition(t *testing.T) {
	tests := []struct {
		name       string
		prestigeID PrestigeClassID
		wantErr    bool
	}{
		{"valid blade master", PrestigeBladeMaster, false},
		{"valid archmage", PrestigeArchmage, false},
		{"invalid prestige", PrestigeClassID("invalid"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			def, err := GetPrestigeClassDefinition(tt.prestigeID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetPrestigeClassDefinition() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if def.ID != tt.prestigeID {
					t.Errorf("GetPrestigeClassDefinition() ID = %v, want %v", def.ID, tt.prestigeID)
				}
				if def.Requirements.MinLevel < 20 {
					t.Errorf("GetPrestigeClassDefinition() MinLevel = %v, want >= 20", def.Requirements.MinLevel)
				}
			}
		})
	}
}

func TestGetAllClasses(t *testing.T) {
	classes := GetAllClasses()
	if len(classes) != 15 {
		t.Errorf("GetAllClasses() returned %d classes, want 15", len(classes))
	}

	for _, class := range classes {
		if class.ID == "" {
			t.Error("GetAllClasses() returned class with empty ID")
		}
		if class.Name == "" {
			t.Error("GetAllClasses() returned class with empty Name")
		}
	}
}

func TestGetAllPrestigeClasses(t *testing.T) {
	classes := GetAllPrestigeClasses()
	if len(classes) != 20 {
		t.Errorf("GetAllPrestigeClasses() returned %d classes, want 20", len(classes))
	}

	for _, class := range classes {
		if class.ID == "" {
			t.Error("GetAllPrestigeClasses() returned class with empty ID")
		}
		if class.Requirements.MinLevel < 20 {
			t.Errorf("Prestige class %s has MinLevel %d, want >= 20", class.ID, class.Requirements.MinLevel)
		}
	}
}

func TestStatBonuses_Add(t *testing.T) {
	a := StatBonuses{
		Health:     100,
		Strength:   10,
		CritChance: 0.1,
	}
	b := StatBonuses{
		Health:     50,
		Strength:   5,
		CritDamage: 0.2,
	}

	result := a.Add(b)

	if result.Health != 150 {
		t.Errorf("Add() Health = %v, want 150", result.Health)
	}
	if result.Strength != 15 {
		t.Errorf("Add() Strength = %v, want 15", result.Strength)
	}
	if result.CritChance != 0.1 {
		t.Errorf("Add() CritChance = %v, want 0.1", result.CritChance)
	}
	if result.CritDamage != 0.2 {
		t.Errorf("Add() CritDamage = %v, want 0.2", result.CritDamage)
	}
}

func TestStatBonuses_Scale(t *testing.T) {
	original := StatBonuses{
		Health:     100,
		Strength:   10,
		CritChance: 0.1,
	}

	scaled := original.Scale(0.5)

	if scaled.Health != 50 {
		t.Errorf("Scale() Health = %v, want 50", scaled.Health)
	}
	if scaled.Strength != 5 {
		t.Errorf("Scale() Strength = %v, want 5", scaled.Strength)
	}
	if scaled.CritChance != 0.05 {
		t.Errorf("Scale() CritChance = %v, want 0.05", scaled.CritChance)
	}
}

func TestAdvancedClassComponent_Type(t *testing.T) {
	component := AdvancedClassComponent{}
	if component.Type() != "advanced_class" {
		t.Errorf("Type() = %v, want advanced_class", component.Type())
	}
}

func TestClassDefinitions(t *testing.T) {
	for id, def := range classDefinitions {
		t.Run(string(id), func(t *testing.T) {
			if def.ID != id {
				t.Errorf("Class %s has mismatched ID: %s", id, def.ID)
			}
			if def.Name == "" {
				t.Errorf("Class %s has empty name", id)
			}
			if def.Description == "" {
				t.Errorf("Class %s has empty description", id)
			}
			if def.Color.A == 0 {
				t.Errorf("Class %s has invalid color (alpha = 0)", id)
			}
		})
	}
}

func TestPrestigeDefinitions(t *testing.T) {
	for id, def := range prestigeDefinitions {
		t.Run(string(id), func(t *testing.T) {
			if def.ID != id {
				t.Errorf("Prestige class %s has mismatched ID: %s", id, def.ID)
			}
			if def.Name == "" {
				t.Errorf("Prestige class %s has empty name", id)
			}
			if def.Requirements.MinLevel < 20 {
				t.Errorf("Prestige class %s has MinLevel %d, want >= 20", id, def.Requirements.MinLevel)
			}
			if len(def.Requirements.RequiredPrimary) == 0 && len(def.Requirements.RequiredSecondary) == 0 {
				t.Errorf("Prestige class %s has no class requirements", id)
			}
		})
	}
}

func BenchmarkGetClassDefinition(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GetClassDefinition(ClassWarrior)
	}
}

func BenchmarkStatBonusesAdd(b *testing.B) {
	a := StatBonuses{Health: 100, Strength: 10}
	bonus := StatBonuses{Health: 50, Strength: 5}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.Add(bonus)
	}
}

func BenchmarkStatBonusesScale(b *testing.B) {
	bonus := StatBonuses{Health: 100, Strength: 10}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bonus.Scale(0.5)
	}
}
