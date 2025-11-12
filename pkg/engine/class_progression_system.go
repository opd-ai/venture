package engine

// ClassProgressionSystem manages character class leveling.
// This is a V4 Phase 25 feature that extends the basic CharacterClass.
type ClassProgressionSystem struct {
	specializationLevel int
}

// NewClassProgressionSystem creates a new class progression system.
func NewClassProgressionSystem() *ClassProgressionSystem {
	return &ClassProgressionSystem{
		specializationLevel: 10,
	}
}

// Update processes class progression for all entities.
func (cps *ClassProgressionSystem) Update(entities []*Entity, deltaTime float64) {
	// Currently a stub - progression happens through LevelUp() calls
	// This system could be extended to apply passive effects
}

// LevelUp increases the class level and unlocks new abilities.
func (cps *ClassProgressionSystem) LevelUp(entity *Entity) bool {
	classComp, ok := entity.GetComponent("class_progression")
	if !ok || classComp == nil {
		return false
	}

	progression := classComp.(*ClassProgressionComponent)
	progression.Level++

	return true
}

// ChooseSpecialization sets a character's specialization at level 10+.
func (cps *ClassProgressionSystem) ChooseSpecialization(entity *Entity, spec SpecializationType) bool {
	classComp, ok := entity.GetComponent("class_progression")
	if !ok || classComp == nil {
		return false
	}

	progression := classComp.(*ClassProgressionComponent)

	// Check level requirement
	if progression.Level < cps.specializationLevel {
		return false
	}

	// Check not already specialized
	if progression.Specialization != SpecializationNone {
		return false
	}

	// Validate specialization matches class
	validSpecs := GetAvailableSpecializations(progression.Class)
	valid := false
	for _, validSpec := range validSpecs {
		if validSpec == spec {
			valid = true
			break
		}
	}

	if !valid {
		return false
	}

	// Apply specialization
	progression.Specialization = spec
	progression.Abilities = append(progression.Abilities, GetSpecializationAbilities(spec)...)

	return true
}
