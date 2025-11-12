package engine

// ClassProgressionSystem manages character class leveling.
// This is a V4 Phase 25 feature that extends the basic CharacterClass
// with stat growth, ability unlocking, and specialization mechanics.
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

// LevelUp increases the class level and applies stat growth bonuses.
// Updates health, mana, attack, defense, magic power, and speed based
// on the class's StatGrowth configuration.
func (cps *ClassProgressionSystem) LevelUp(entity *Entity) bool {
	classComp, ok := entity.GetComponent("class_progression")
	if !ok || classComp == nil {
		return false
	}

	progression := classComp.(*ClassProgressionComponent)
	progression.Level++

	// Apply stat growth bonuses
	cps.ApplyStatGrowth(entity, progression.Class, progression.Level)

	return true
}

// ApplyStatGrowth updates an entity's stats based on class and level.
// Uses the StatGrowth system to calculate appropriate stat values.
func (cps *ClassProgressionSystem) ApplyStatGrowth(entity *Entity, class CharacterClass, level int) {
	growth := GetClassStatGrowth(class)
	if growth == nil {
		return
	}

	// Get base stats (if available)
	var baseHP, baseMana, baseAttack, baseDefense, baseMagicPower float64
	if baseStatsComp, ok := entity.GetComponent("base_stats"); ok {
		if baseStats, ok := baseStatsComp.(*BaseStatsComponent); ok {
			baseHP = baseStats.BaseMaxHealth
			baseMana = 0 // BaseStatsComponent doesn't store base mana
			baseAttack = baseStats.BaseAttack
			baseDefense = baseStats.BaseDefense
			baseMagicPower = baseStats.BaseMagicPower
		}
	}

	// If no base stats component, use current values as base
	if baseHP == 0 {
		if healthComp, ok := entity.GetComponent("health"); ok {
			if health, ok := healthComp.(*HealthComponent); ok {
				baseHP = health.Max
			}
		}
	}
	if baseMana == 0 {
		if manaComp, ok := entity.GetComponent("mana"); ok {
			if mana, ok := manaComp.(*ManaComponent); ok {
				baseMana = float64(mana.Max)
			}
		}
	}

	// Apply growth calculations
	newHP := growth.CalculateHP(baseHP, level)
	newMana := growth.CalculateMana(baseMana, level)
	newAttack := growth.CalculateAttack(baseAttack, level)
	newDefense := growth.CalculateDefense(baseDefense, level)
	newMagicPower := growth.CalculateMagicPower(baseMagicPower, level)
	// Note: Speed growth calculation exists but VelocityComponent doesn't support max speed

	// Update health component
	if healthComp, ok := entity.GetComponent("health"); ok {
		if health, ok := healthComp.(*HealthComponent); ok {
			// Preserve health percentage
			healthPercent := health.Current / health.Max
			health.Max = newHP
			health.Current = newHP * healthPercent
			if health.Current < 1 && healthPercent > 0 {
				health.Current = 1 // Ensure at least 1 HP if was alive
			}
		}
	}

	// Update mana component
	if manaComp, ok := entity.GetComponent("mana"); ok {
		if mana, ok := manaComp.(*ManaComponent); ok {
			if mana.Max > 0 {
				manaPercent := float64(mana.Current) / float64(mana.Max)
				mana.Max = int(newMana)
				mana.Current = int(float64(mana.Max) * manaPercent)
			} else {
				mana.Max = int(newMana)
				mana.Current = int(newMana)
			}
		}
	}

	// Update stats component
	if statsComp, ok := entity.GetComponent("stats"); ok {
		if stats, ok := statsComp.(*StatsComponent); ok {
			stats.Attack = newAttack
			stats.Defense = newDefense
			stats.MagicPower = newMagicPower
		}
	}

	// Update attack component
	if attackComp, ok := entity.GetComponent("attack"); ok {
		if attack, ok := attackComp.(*AttackComponent); ok {
			attack.Damage = newAttack
		}
	}

	// Note: VelocityComponent does not have MaxSpeed field
	// Speed scaling would need to be handled differently if needed
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

// GetClassLevel returns the current class level of an entity.
// Returns 1 if entity has no class progression component.
func (cps *ClassProgressionSystem) GetClassLevel(entity *Entity) int {
	classComp, ok := entity.GetComponent("class_progression")
	if !ok || classComp == nil {
		return 1
	}

	progression := classComp.(*ClassProgressionComponent)
	return progression.Level
}

// GetClassAbilitiesForEntity returns all abilities available to an entity
// including base class abilities and specialization abilities.
func (cps *ClassProgressionSystem) GetClassAbilitiesForEntity(entity *Entity) []string {
	classComp, ok := entity.GetComponent("class_progression")
	if !ok || classComp == nil {
		return []string{}
	}

	progression := classComp.(*ClassProgressionComponent)
	return progression.Abilities
}

// UnlockSecondClass enables dual-classing for an entity at level 20+.
// Phase 25.2: Dual-classing system.
func (cps *ClassProgressionSystem) UnlockSecondClass(entity *Entity, secondClass CharacterClass) bool {
	classComp, ok := entity.GetComponent("class_progression")
	if !ok || classComp == nil {
		return false
	}

	progression := classComp.(*ClassProgressionComponent)

	// Check level requirement (level 20+)
	if progression.Level < 20 {
		return false
	}

	// Check not already dual-classed
	if progression.SecondaryClass != nil {
		return false
	}

	// Cannot dual-class into the same class
	if progression.Class == secondClass {
		return false
	}

	// Unlock secondary class
	progression.SecondaryClass = &secondClass
	progression.SecondaryLevel = 1
	progression.SecondarySpec = SpecializationNone

	// Add starting abilities from secondary class
	secondaryAbilities := GetClassAbilities(secondClass)
	progression.Abilities = append(progression.Abilities, secondaryAbilities...)

	return true
}

// LevelUpSecondaryClass increases the secondary class level.
// Phase 25.2: Dual-classing progression.
func (cps *ClassProgressionSystem) LevelUpSecondaryClass(entity *Entity) bool {
	classComp, ok := entity.GetComponent("class_progression")
	if !ok || classComp == nil {
		return false
	}

	progression := classComp.(*ClassProgressionComponent)

	// Must have a secondary class
	if progression.SecondaryClass == nil {
		return false
	}

	progression.SecondaryLevel++

	// Apply stat growth from secondary class (scaled by 50% to avoid overpowered builds)
	growth := GetClassStatGrowth(*progression.SecondaryClass)
	if growth != nil {
		// Get current stats
		var baseHP, baseMana, baseAttack, baseDefense, baseMagicPower float64
		if healthComp, ok := entity.GetComponent("health"); ok {
			if health, ok := healthComp.(*HealthComponent); ok {
				baseHP = health.Max
			}
		}
		if manaComp, ok := entity.GetComponent("mana"); ok {
			if mana, ok := manaComp.(*ManaComponent); ok {
				baseMana = float64(mana.Max)
			}
		}
		if statsComp, ok := entity.GetComponent("stats"); ok {
			if stats, ok := statsComp.(*StatsComponent); ok {
				baseAttack = stats.Attack
				baseDefense = stats.Defense
				baseMagicPower = stats.MagicPower
			}
		}

		// Calculate growth at 50% effectiveness
		hpGrowth := (growth.CalculateHP(baseHP, progression.SecondaryLevel) - baseHP) * 0.5
		manaGrowth := (growth.CalculateMana(baseMana, progression.SecondaryLevel) - baseMana) * 0.5
		attackGrowth := (growth.CalculateAttack(baseAttack, progression.SecondaryLevel) - baseAttack) * 0.5
		defenseGrowth := (growth.CalculateDefense(baseDefense, progression.SecondaryLevel) - baseDefense) * 0.5
		magicGrowth := (growth.CalculateMagicPower(baseMagicPower, progression.SecondaryLevel) - baseMagicPower) * 0.5

		// Apply growth
		if healthComp, ok := entity.GetComponent("health"); ok {
			if health, ok := healthComp.(*HealthComponent); ok {
				healthPercent := health.Current / health.Max
				health.Max += hpGrowth
				health.Current = health.Max * healthPercent
			}
		}
		if manaComp, ok := entity.GetComponent("mana"); ok {
			if mana, ok := manaComp.(*ManaComponent); ok {
				if mana.Max > 0 {
					manaPercent := float64(mana.Current) / float64(mana.Max)
					mana.Max += int(manaGrowth)
					mana.Current = int(float64(mana.Max) * manaPercent)
				}
			}
		}
		if statsComp, ok := entity.GetComponent("stats"); ok {
			if stats, ok := statsComp.(*StatsComponent); ok {
				stats.Attack += attackGrowth
				stats.Defense += defenseGrowth
				stats.MagicPower += magicGrowth
			}
		}
		if attackComp, ok := entity.GetComponent("attack"); ok {
			if attack, ok := attackComp.(*AttackComponent); ok {
				attack.Damage += attackGrowth
			}
		}
	}

	return true
}

// ChooseSecondarySpecialization sets the secondary class specialization.
// Phase 25.2: Dual-classing specialization.
func (cps *ClassProgressionSystem) ChooseSecondarySpecialization(entity *Entity, spec SpecializationType) bool {
	classComp, ok := entity.GetComponent("class_progression")
	if !ok || classComp == nil {
		return false
	}

	progression := classComp.(*ClassProgressionComponent)

	// Must have a secondary class
	if progression.SecondaryClass == nil {
		return false
	}

	// Check level requirement (level 10+ in secondary class)
	if progression.SecondaryLevel < 10 {
		return false
	}

	// Check not already specialized in secondary
	if progression.SecondarySpec != SpecializationNone {
		return false
	}

	// Validate specialization matches secondary class
	validSpecs := GetAvailableSpecializations(*progression.SecondaryClass)
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

	// Apply secondary specialization
	progression.SecondarySpec = spec
	progression.Abilities = append(progression.Abilities, GetSpecializationAbilities(spec)...)

	return true
}
