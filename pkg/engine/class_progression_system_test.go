package engine

import (
	"testing"
)

func TestClassProgressionSystem_LevelUp(t *testing.T) {
	system := NewClassProgressionSystem()

	tests := []struct {
		name          string
		setupEntity   func() *Entity
		expectLevel   int
		expectSuccess bool
	}{
		{
			name: "level up warrior increases level",
			setupEntity: func() *Entity {
				entity := NewEntity(1)
				entity.AddComponent(&ClassProgressionComponent{
					Class:     ClassWarrior,
					Level:     1,
					Abilities: GetClassAbilities(ClassWarrior),
				})
				entity.AddComponent(&HealthComponent{Current: 150, Max: 150})
				entity.AddComponent(&ManaComponent{Current: 50, Max: 50})
				entity.AddComponent(&StatsComponent{Attack: 12, Defense: 8})
				entity.AddComponent(&AttackComponent{Damage: 20})
				return entity
			},
			expectLevel:   2,
			expectSuccess: true,
		},
		{
			name: "entity without class component fails",
			setupEntity: func() *Entity {
				entity := NewEntity(2)
				return entity
			},
			expectLevel:   0,
			expectSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entity := tt.setupEntity()
			success := system.LevelUp(entity)

			if success != tt.expectSuccess {
				t.Errorf("LevelUp() success = %v, want %v", success, tt.expectSuccess)
			}

			if success {
				classComp, _ := entity.GetComponent("class_progression")
				progression := classComp.(*ClassProgressionComponent)
				if progression.Level != tt.expectLevel {
					t.Errorf("Level = %v, want %v", progression.Level, tt.expectLevel)
				}
			}
		})
	}
}

func TestClassProgressionSystem_ApplyStatGrowth(t *testing.T) {
	system := NewClassProgressionSystem()

	// Create a warrior entity at level 1
	entity := NewEntity(1)
	entity.AddComponent(&ClassProgressionComponent{
		Class:     ClassWarrior,
		Level:     1,
		Abilities: GetClassAbilities(ClassWarrior),
	})
	entity.AddComponent(&HealthComponent{Current: 150, Max: 150})
	entity.AddComponent(&ManaComponent{Current: 50, Max: 50})
	entity.AddComponent(&StatsComponent{Attack: 12, Defense: 8, MagicPower: 5})
	entity.AddComponent(&AttackComponent{Damage: 20})
	entity.AddComponent(&BaseStatsComponent{
		BaseMaxHealth:  150,
		BaseAttack:     12,
		BaseDefense:    8,
		BaseMagicPower: 5,
	})

	// Apply stat growth for level 5
	system.ApplyStatGrowth(entity, ClassWarrior, 5)

	// Verify stats increased
	healthComp, _ := entity.GetComponent("health")
	health := healthComp.(*HealthComponent)

	// Warriors get +15 HP per level + 5% per level
	// Base 150 + (15 * 4) = 210, then 210 * (1 + 0.05 * 4) = 210 * 1.2 = 252
	if health.Max < 250 {
		t.Errorf("Health Max after level 5 too low: %v, expected at least 250", health.Max)
	}

	statsComp, _ := entity.GetComponent("stats")
	stats := statsComp.(*StatsComponent)

	// Attack should have increased
	if stats.Attack <= 12 {
		t.Errorf("Attack did not increase: %v", stats.Attack)
	}
}

func TestClassProgressionSystem_ChooseSpecialization(t *testing.T) {
	system := NewClassProgressionSystem()

	tests := []struct {
		name           string
		level          int
		class          CharacterClass
		specialization SpecializationType
		expectSuccess  bool
	}{
		{
			name:           "level 10 warrior can specialize to berserker",
			level:          10,
			class:          ClassWarrior,
			specialization: SpecializationBerserker,
			expectSuccess:  true,
		},
		{
			name:           "level 5 warrior cannot specialize",
			level:          5,
			class:          ClassWarrior,
			specialization: SpecializationBerserker,
			expectSuccess:  false,
		},
		{
			name:           "warrior cannot specialize to mage spec",
			level:          10,
			class:          ClassWarrior,
			specialization: SpecializationElementalist,
			expectSuccess:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entity := NewEntity(1)
			entity.AddComponent(&ClassProgressionComponent{
				Class:          tt.class,
				Level:          tt.level,
				Specialization: SpecializationNone,
				Abilities:      GetClassAbilities(tt.class),
			})

			success := system.ChooseSpecialization(entity, tt.specialization)

			if success != tt.expectSuccess {
				t.Errorf("ChooseSpecialization() = %v, want %v", success, tt.expectSuccess)
			}

			if success {
				classComp, _ := entity.GetComponent("class_progression")
				progression := classComp.(*ClassProgressionComponent)
				if progression.Specialization != tt.specialization {
					t.Errorf("Specialization = %v, want %v", progression.Specialization, tt.specialization)
				}

				// Verify specialization abilities were added
				specAbilities := GetSpecializationAbilities(tt.specialization)
				if len(progression.Abilities) < len(specAbilities) {
					t.Errorf("Specialization abilities not added")
				}
			}
		})
	}
}

func TestClassProgressionSystem_GetClassLevel(t *testing.T) {
	system := NewClassProgressionSystem()

	tests := []struct {
		name         string
		hasClassComp bool
		level        int
		expected     int
	}{
		{"entity with level 5", true, 5, 5},
		{"entity with level 1", true, 1, 1},
		{"entity without class comp", false, 0, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entity := NewEntity(1)
			if tt.hasClassComp {
				entity.AddComponent(&ClassProgressionComponent{
					Class: ClassWarrior,
					Level: tt.level,
				})
			}

			result := system.GetClassLevel(entity)
			if result != tt.expected {
				t.Errorf("GetClassLevel() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestClassProgressionSystem_GetClassAbilitiesForEntity(t *testing.T) {
	system := NewClassProgressionSystem()

	entity := NewEntity(1)
	entity.AddComponent(&ClassProgressionComponent{
		Class:     ClassWarrior,
		Level:     10, // Level 10 required for specialization
		Abilities: GetClassAbilities(ClassWarrior),
	})

	abilities := system.GetClassAbilitiesForEntity(entity)

	// Should have 8 warrior abilities
	if len(abilities) != 8 {
		t.Errorf("GetClassAbilitiesForEntity() returned %d abilities, want 8", len(abilities))
	}

	// Add specialization
	success := system.ChooseSpecialization(entity, SpecializationBerserker)
	if !success {
		t.Fatal("Failed to choose specialization")
	}

	abilitiesWithSpec := system.GetClassAbilitiesForEntity(entity)

	// Should have more abilities after specialization
	if len(abilitiesWithSpec) <= len(abilities) {
		t.Errorf("Abilities did not increase after specialization: before=%d, after=%d",
			len(abilities), len(abilitiesWithSpec))
	}
}

func TestClassProgressionSystem_HealthPreservation(t *testing.T) {
	// Test that health percentage is preserved when leveling up
	system := NewClassProgressionSystem()

	entity := NewEntity(1)
	entity.AddComponent(&ClassProgressionComponent{
		Class: ClassWarrior,
		Level: 1,
	})
	entity.AddComponent(&HealthComponent{Current: 75, Max: 150}) // 50% health
	entity.AddComponent(&BaseStatsComponent{BaseMaxHealth: 150})

	// Level up
	system.LevelUp(entity)

	healthComp, _ := entity.GetComponent("health")
	health := healthComp.(*HealthComponent)

	// Health percentage should be preserved (50%)
	healthPercent := health.Current / health.Max
	if healthPercent < 0.49 || healthPercent > 0.51 {
		t.Errorf("Health percentage not preserved: got %v, want ~0.50", healthPercent)
	}
}

func TestClassProgressionSystem_AllClassesHaveAbilities(t *testing.T) {
	// Verify all 6 classes have 8+ abilities
	classes := []CharacterClass{
		ClassWarrior,
		ClassRogue,
		ClassMage,
		ClassRanger,
		ClassCleric,
		ClassNecromancer,
	}

	for _, class := range classes {
		t.Run(class.String(), func(t *testing.T) {
			abilities := GetClassAbilities(class)
			if len(abilities) < 8 {
				t.Errorf("Class %s has only %d abilities, want at least 8",
					class.String(), len(abilities))
			}
		})
	}
}

// Benchmark tests
func BenchmarkClassProgressionSystem_LevelUp(b *testing.B) {
	system := NewClassProgressionSystem()
	entity := NewEntity(1)
	entity.AddComponent(&ClassProgressionComponent{
		Class: ClassWarrior,
		Level: 1,
	})
	entity.AddComponent(&HealthComponent{Current: 150, Max: 150})
	entity.AddComponent(&ManaComponent{Current: 50, Max: 50})
	entity.AddComponent(&StatsComponent{Attack: 12, Defense: 8})
	entity.AddComponent(&BaseStatsComponent{BaseMaxHealth: 150, BaseAttack: 12})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		system.LevelUp(entity)
	}
}

func BenchmarkClassProgressionSystem_ApplyStatGrowth(b *testing.B) {
	system := NewClassProgressionSystem()
	entity := NewEntity(1)
	entity.AddComponent(&HealthComponent{Current: 150, Max: 150})
	entity.AddComponent(&ManaComponent{Current: 50, Max: 50})
	entity.AddComponent(&StatsComponent{Attack: 12, Defense: 8})
	entity.AddComponent(&BaseStatsComponent{BaseMaxHealth: 150, BaseAttack: 12})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		system.ApplyStatGrowth(entity, ClassWarrior, 10)
	}
}
