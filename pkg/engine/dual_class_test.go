package engine

import "testing"

// TestClassProgressionSystem_DualClassing tests the dual-classing system.
// Phase 25.2: Dual-classing unlock at level 20.
func TestClassProgressionSystem_DualClassing(t *testing.T) {
	tests := []struct {
		name           string
		primaryClass   CharacterClass
		primaryLevel   int
		secondaryClass CharacterClass
		wantSuccess    bool
		wantError      string
	}{
		{
			name:           "unlock at level 20",
			primaryClass:   ClassWarrior,
			primaryLevel:   20,
			secondaryClass: ClassMage,
			wantSuccess:    true,
		},
		{
			name:           "unlock at level 25",
			primaryClass:   ClassRogue,
			primaryLevel:   25,
			secondaryClass: ClassRanger,
			wantSuccess:    true,
		},
		{
			name:           "fail below level 20",
			primaryClass:   ClassMage,
			primaryLevel:   19,
			secondaryClass: ClassCleric,
			wantSuccess:    false,
			wantError:      "level requirement",
		},
		{
			name:           "fail same class",
			primaryClass:   ClassWarrior,
			primaryLevel:   20,
			secondaryClass: ClassWarrior,
			wantSuccess:    false,
			wantError:      "same class",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			system := NewClassProgressionSystem()
			entity := NewEntity(1)
			entity.AddComponent(&ClassProgressionComponent{
				Class:     tt.primaryClass,
				Level:     tt.primaryLevel,
				Abilities: GetClassAbilities(tt.primaryClass),
			})

			// Try to unlock secondary class
			success := system.UnlockSecondClass(entity, tt.secondaryClass)

			if success != tt.wantSuccess {
				t.Errorf("UnlockSecondClass() = %v, want %v", success, tt.wantSuccess)
				return
			}

			if success {
				// Verify secondary class is set
				comp, ok := entity.GetComponent("class_progression")
				if !ok {
					t.Fatal("Failed to get class progression component")
				}
				progression := comp.(*ClassProgressionComponent)

				if progression.SecondaryClass == nil {
					t.Error("SecondaryClass is nil after successful unlock")
				} else if *progression.SecondaryClass != tt.secondaryClass {
					t.Errorf("SecondaryClass = %v, want %v", *progression.SecondaryClass, tt.secondaryClass)
				}

				if progression.SecondaryLevel != 1 {
					t.Errorf("SecondaryLevel = %d, want 1", progression.SecondaryLevel)
				}

				// Verify secondary abilities were added
				secondaryAbilities := GetClassAbilities(tt.secondaryClass)
				primaryAbilityCount := len(GetClassAbilities(tt.primaryClass))
				expectedTotal := primaryAbilityCount + len(secondaryAbilities)

				if len(progression.Abilities) != expectedTotal {
					t.Errorf("Abilities count = %d, want %d (primary %d + secondary %d)",
						len(progression.Abilities), expectedTotal, primaryAbilityCount, len(secondaryAbilities))
				}
			}
		})
	}
}

// TestClassProgressionSystem_SecondaryLevelUp tests leveling up secondary class.
func TestClassProgressionSystem_SecondaryLevelUp(t *testing.T) {
	system := NewClassProgressionSystem()
	entity := NewEntity(1)

	// Set up entity with stats
	entity.AddComponent(&HealthComponent{Current: 100, Max: 100})
	entity.AddComponent(&ManaComponent{Current: 50, Max: 50})
	entity.AddComponent(&StatsComponent{Attack: 10, Defense: 8, MagicPower: 5})
	entity.AddComponent(&AttackComponent{Damage: 10})

	// Start as Warrior level 20
	secondaryClass := ClassMage
	entity.AddComponent(&ClassProgressionComponent{
		Class:          ClassWarrior,
		Level:          20,
		Abilities:      GetClassAbilities(ClassWarrior),
		SecondaryClass: &secondaryClass,
		SecondaryLevel: 1,
	})

	// Get initial stats
	comp1, ok := entity.GetComponent("health")
	if !ok {
		t.Fatal("Failed to get health component")
	}
	healthComp := comp1.(*HealthComponent)
	
	comp2, ok := entity.GetComponent("mana")
	if !ok {
		t.Fatal("Failed to get mana component")
	}
	manaComp := comp2.(*ManaComponent)
	
	comp3, ok := entity.GetComponent("stats")
	if !ok {
		t.Fatal("Failed to get stats component")
	}
	statsComp := comp3.(*StatsComponent)
	
	initialHP := healthComp.Max
	initialMana := manaComp.Max
	initialAttack := statsComp.Attack

	// Level up secondary class
	success := system.LevelUpSecondaryClass(entity)
	if !success {
		t.Fatal("LevelUpSecondaryClass failed")
	}

	// Verify level increased
	comp4, ok := entity.GetComponent("class_progression")
	if !ok {
		t.Fatal("Failed to get class progression component")
	}
	progression := comp4.(*ClassProgressionComponent)
	if progression.SecondaryLevel != 2 {
		t.Errorf("SecondaryLevel = %d, want 2", progression.SecondaryLevel)
	}

	// Verify stats increased (should be 50% of normal growth)
	if healthComp.Max <= initialHP {
		t.Error("HP did not increase after secondary level up")
	}

	if manaComp.Max <= initialMana {
		t.Error("Mana did not increase after secondary level up (expected for Mage)")
	}

	if statsComp.Attack <= initialAttack {
		t.Error("Attack did not increase after secondary level up")
	}
}

// TestClassProgressionSystem_SecondarySpecialization tests specialization for secondary class.
func TestClassProgressionSystem_SecondarySpecialization(t *testing.T) {
	system := NewClassProgressionSystem()
	entity := NewEntity(1)

	secondaryClass := ClassMage
	entity.AddComponent(&ClassProgressionComponent{
		Class:          ClassWarrior,
		Level:          20,
		Specialization: SpecializationBerserker,
		Abilities:      append(GetClassAbilities(ClassWarrior), GetSpecializationAbilities(SpecializationBerserker)...),
		SecondaryClass: &secondaryClass,
		SecondaryLevel: 10, // At level 10, can specialize
		SecondarySpec:  SpecializationNone,
	})

	comp1, ok := entity.GetComponent("class_progression")
	if !ok {
		t.Fatal("Failed to get class progression component")
	}
	initialAbilityCount := len(comp1.(*ClassProgressionComponent).Abilities)

	// Try to specialize in Elementalist (Mage specialization)
	success := system.ChooseSecondarySpecialization(entity, SpecializationElementalist)
	if !success {
		t.Fatal("ChooseSecondarySpecialization failed")
	}

	comp2, ok := entity.GetComponent("class_progression")
	if !ok {
		t.Fatal("Failed to get class progression component")
	}
	progression := comp2.(*ClassProgressionComponent)

	// Verify secondary specialization set
	if progression.SecondarySpec != SpecializationElementalist {
		t.Errorf("SecondarySpec = %v, want %v", progression.SecondarySpec, SpecializationElementalist)
	}

	// Verify specialization abilities added
	specAbilities := GetSpecializationAbilities(SpecializationElementalist)
	expectedTotal := initialAbilityCount + len(specAbilities)

	if len(progression.Abilities) != expectedTotal {
		t.Errorf("Abilities count = %d, want %d", len(progression.Abilities), expectedTotal)
	}
}

// TestClassProgressionSystem_SecondarySpecialization_Requirements tests specialization requirements.
func TestClassProgressionSystem_SecondarySpecialization_Requirements(t *testing.T) {
	tests := []struct {
		name               string
		secondaryLevel     int
		currentSpec        SpecializationType
		requestedSpec      SpecializationType
		wantSuccess        bool
	}{
		{
			name:           "level 10 - can specialize",
			secondaryLevel: 10,
			currentSpec:    SpecializationNone,
			requestedSpec:  SpecializationElementalist,
			wantSuccess:    true,
		},
		{
			name:           "level 15 - can specialize",
			secondaryLevel: 15,
			currentSpec:    SpecializationNone,
			requestedSpec:  SpecializationArcanist,
			wantSuccess:    true,
		},
		{
			name:           "level 9 - cannot specialize",
			secondaryLevel: 9,
			currentSpec:    SpecializationNone,
			requestedSpec:  SpecializationElementalist,
			wantSuccess:    false,
		},
		{
			name:           "already specialized - fail",
			secondaryLevel: 15,
			currentSpec:    SpecializationElementalist,
			requestedSpec:  SpecializationArcanist,
			wantSuccess:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			system := NewClassProgressionSystem()
			entity := NewEntity(1)

			secondaryClass := ClassMage
			entity.AddComponent(&ClassProgressionComponent{
				Class:          ClassWarrior,
				Level:          20,
				Specialization: SpecializationBerserker,
				Abilities:      GetClassAbilities(ClassWarrior),
				SecondaryClass: &secondaryClass,
				SecondaryLevel: tt.secondaryLevel,
				SecondarySpec:  tt.currentSpec,
			})

			success := system.ChooseSecondarySpecialization(entity, tt.requestedSpec)
			if success != tt.wantSuccess {
				t.Errorf("ChooseSecondarySpecialization() = %v, want %v", success, tt.wantSuccess)
			}
		})
	}
}

// TestCharacterClass_LowerName tests the LowerName method.
// Phase 25.2: Used for class restriction matching.
func TestCharacterClass_LowerName(t *testing.T) {
	tests := []struct {
		class CharacterClass
		want  string
	}{
		{ClassWarrior, "warrior"},
		{ClassMage, "mage"},
		{ClassRogue, "rogue"},
		{ClassRanger, "ranger"},
		{ClassCleric, "cleric"},
		{ClassNecromancer, "necromancer"},
	}

	for _, tt := range tests {
		t.Run(tt.class.String(), func(t *testing.T) {
			got := tt.class.LowerName()
			if got != tt.want {
				t.Errorf("LowerName() = %q, want %q", got, tt.want)
			}
		})
	}
}
