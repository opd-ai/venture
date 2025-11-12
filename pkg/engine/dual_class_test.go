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

// TestHybridClasses tests all 15 hybrid classes.
// Phase 25.2 Extension: Verify hybrid class configuration.
func TestHybridClasses(t *testing.T) {
hybridClasses := []struct {
class           CharacterClass
expectedName    string
expectedLower   string
minAbilities    int
expectedSpecs   int
}{
{ClassBattlemage, "Battlemage", "battlemage", 6, 2},
{ClassSpellblade, "Spellblade", "spellblade", 6, 2},
{ClassPaladin, "Paladin", "paladin", 6, 2},
{ClassMonk, "Monk", "monk", 6, 2},
{ClassDeathKnight, "Death Knight", "deathknight", 6, 2},
{ClassWitchHunter, "Witch Hunter", "witchhunter", 6, 2},
{ClassBeastlord, "Beastlord", "beastlord", 6, 2},
{ClassArcaneArcher, "Arcane Archer", "arcanearcher", 6, 2},
{ClassShadowPriest, "Shadow Priest", "shadowpriest", 6, 2},
{ClassDruid, "Druid", "druid", 6, 2},
{ClassInquisitor, "Inquisitor", "inquisitor", 6, 2},
{ClassBloodKnight, "Blood Knight", "bloodknight", 6, 2},
{ClassMystic, "Mystic", "mystic", 6, 2},
{ClassWarlock, "Warlock", "warlock", 6, 2},
{ClassNinja, "Ninja", "ninja", 6, 2},
}

for _, tc := range hybridClasses {
t.Run(tc.expectedName, func(t *testing.T) {
// Test String() method
if got := tc.class.String(); got != tc.expectedName {
t.Errorf("String() = %q, want %q", got, tc.expectedName)
}

// Test LowerName() method
if got := tc.class.LowerName(); got != tc.expectedLower {
t.Errorf("LowerName() = %q, want %q", got, tc.expectedLower)
}

// Test abilities count
abilities := GetClassAbilities(tc.class)
if len(abilities) < tc.minAbilities {
t.Errorf("GetClassAbilities() returned %d abilities, want at least %d",
len(abilities), tc.minAbilities)
}

// Test specializations count
specs := GetAvailableSpecializations(tc.class)
if len(specs) != tc.expectedSpecs {
t.Errorf("GetAvailableSpecializations() returned %d specs, want %d",
len(specs), tc.expectedSpecs)
}

// Test stat growth exists
growth := GetClassStatGrowth(tc.class)
if growth == nil {
t.Errorf("GetClassStatGrowth() returned nil for %s", tc.expectedName)
}
})
}
}

// TestHybridClassBalance tests that hybrid classes have balanced stats.
func TestHybridClassBalance(t *testing.T) {
allClasses := []CharacterClass{
ClassWarrior, ClassMage, ClassRogue, ClassRanger, ClassCleric, ClassNecromancer,
ClassBattlemage, ClassSpellblade, ClassPaladin, ClassMonk, ClassDeathKnight,
ClassWitchHunter, ClassBeastlord, ClassArcaneArcher, ClassShadowPriest,
ClassDruid, ClassInquisitor, ClassBloodKnight, ClassMystic, ClassWarlock, ClassNinja,
}

for _, class := range allClasses {
t.Run(class.String(), func(t *testing.T) {
growth := GetClassStatGrowth(class)
if growth == nil {
t.Fatal("GetClassStatGrowth returned nil")
}

// Verify stat growth is positive
if growth.HPPerLevel <= 0 {
t.Errorf("HPPerLevel = %f, want > 0", growth.HPPerLevel)
}
if growth.ManaPerLevel < 0 {
t.Errorf("ManaPerLevel = %f, want >= 0", growth.ManaPerLevel)
}
if growth.AttackPerLevel < 0 {
t.Errorf("AttackPerLevel = %f, want >= 0", growth.AttackPerLevel)
}
if growth.DefensePerLevel < 0 {
t.Errorf("DefensePerLevel = %f, want >= 0", growth.DefensePerLevel)
}

// Verify reasonable ranges (not too overpowered)
if growth.HPPerLevel > 20 {
t.Errorf("HPPerLevel = %f, seems too high (max 20)", growth.HPPerLevel)
}
if growth.ManaPerLevel > 15 {
t.Errorf("ManaPerLevel = %f, seems too high (max 15)", growth.ManaPerLevel)
}
if growth.AttackPerLevel > 2.5 {
t.Errorf("AttackPerLevel = %f, seems too high (max 2.5)", growth.AttackPerLevel)
}
})
}
}
