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
// DESIGN CONTRACT (G18): Progression is event-driven through explicit LevelUp() calls.
// Passive per-frame effects (e.g., class-specific regen ticks) are out of scope until
// a dedicated passive-effects subsystem is added to the class configuration.
func (cps *ClassProgressionSystem) Update(entities []*Entity, deltaTime float64) {
	_ = entities
	_ = deltaTime
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

	baseStats := cps.extractBaseStats(entity)
	newStats := cps.calculateNewStats(growth, baseStats, level)
	cps.applyNewStats(entity, newStats, baseStats)
}

// extractBaseStats retrieves base stats from an entity.
// Returns base HP, mana, attack, defense, and magic power values.
func (cps *ClassProgressionSystem) extractBaseStats(entity *Entity) statValues {
	stats := statValues{}

	if baseStatsComp, ok := entity.GetComponent("base_stats"); ok {
		if baseStats, ok := baseStatsComp.(*BaseStatsComponent); ok {
			stats.hp = baseStats.BaseMaxHealth
			stats.attack = baseStats.BaseAttack
			stats.defense = baseStats.BaseDefense
			stats.magicPower = baseStats.BaseMagicPower
		}
	}

	cps.fillMissingBaseStats(entity, &stats)
	return stats
}

// fillMissingBaseStats fills in base stats from current values if not set.
func (cps *ClassProgressionSystem) fillMissingBaseStats(entity *Entity, stats *statValues) {
	if stats.hp == 0 {
		if healthComp, ok := entity.GetComponent("health"); ok {
			if health, ok := healthComp.(*HealthComponent); ok {
				stats.hp = health.Max
			}
		}
	}
	if stats.mana == 0 {
		if manaComp, ok := entity.GetComponent("mana"); ok {
			if mana, ok := manaComp.(*ManaComponent); ok {
				stats.mana = float64(mana.Max)
			}
		}
	}
}

// calculateNewStats computes new stat values using class growth formulas.
func (cps *ClassProgressionSystem) calculateNewStats(growth *StatGrowth, base statValues, level int) statValues {
	return statValues{
		hp:         growth.CalculateHP(base.hp, level),
		mana:       growth.CalculateMana(base.mana, level),
		attack:     growth.CalculateAttack(base.attack, level),
		defense:    growth.CalculateDefense(base.defense, level),
		magicPower: growth.CalculateMagicPower(base.magicPower, level),
	}
}

// applyNewStats updates entity components with new stat values.
func (cps *ClassProgressionSystem) applyNewStats(entity *Entity, newStats, baseStats statValues) {
	cps.updateHealthComponent(entity, newStats.hp)
	cps.updateManaComponent(entity, newStats.mana)
	cps.updateStatsComponent(entity, newStats)
	cps.updateAttackComponent(entity, newStats.attack)
}

// updateHealthComponent applies new max health while preserving health percentage.
func (cps *ClassProgressionSystem) updateHealthComponent(entity *Entity, newHP float64) {
	if healthComp, ok := entity.GetComponent("health"); ok {
		if health, ok := healthComp.(*HealthComponent); ok {
			healthPercent := 1.0
			if health.Max > 0 {
				healthPercent = health.Current / health.Max
			}
			health.Max = newHP
			health.Current = newHP * healthPercent
			if health.Current < 1 && healthPercent > 0 {
				health.Current = 1
			}
		}
	}
}

// updateManaComponent applies new max mana while preserving mana percentage.
func (cps *ClassProgressionSystem) updateManaComponent(entity *Entity, newMana float64) {
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
}

// updateStatsComponent applies new attack, defense, and magic power values.
func (cps *ClassProgressionSystem) updateStatsComponent(entity *Entity, stats statValues) {
	if statsComp, ok := entity.GetComponent("stats"); ok {
		if s, ok := statsComp.(*StatsComponent); ok {
			s.Attack = stats.attack
			s.Defense = stats.defense
			s.MagicPower = stats.magicPower
		}
	}
}

// updateAttackComponent applies new attack damage value.
func (cps *ClassProgressionSystem) updateAttackComponent(entity *Entity, attackDamage float64) {
	if attackComp, ok := entity.GetComponent("attack"); ok {
		if attack, ok := attackComp.(*AttackComponent); ok {
			attack.Damage = attackDamage
		}
	}
}

// statValues holds stat values for calculations.
type statValues struct {
	hp         float64
	mana       float64
	attack     float64
	defense    float64
	magicPower float64
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
	progression, ok := cps.validateSecondaryClassProgression(entity)
	if !ok {
		return false
	}

	progression.SecondaryLevel++

	growth := GetClassStatGrowth(*progression.SecondaryClass)
	if growth != nil {
		baseStats := cps.extractCurrentStats(entity)
		statGrowth := cps.calculateSecondaryGrowth(growth, baseStats, progression.SecondaryLevel)
		cps.applySecondaryGrowth(entity, statGrowth)
	}

	return true
}

// validateSecondaryClassProgression checks if entity can level up secondary class.
// Returns the progression component if valid, or false if validation fails.
func (cps *ClassProgressionSystem) validateSecondaryClassProgression(entity *Entity) (*ClassProgressionComponent, bool) {
	classComp, ok := entity.GetComponent("class_progression")
	if !ok || classComp == nil {
		return nil, false
	}

	progression := classComp.(*ClassProgressionComponent)
	if progression.SecondaryClass == nil {
		return nil, false
	}

	return progression, true
}

// extractCurrentStats retrieves current stat values from entity components.
func (cps *ClassProgressionSystem) extractCurrentStats(entity *Entity) statValues {
	stats := statValues{}

	if healthComp, ok := entity.GetComponent("health"); ok {
		if health, ok := healthComp.(*HealthComponent); ok {
			stats.hp = health.Max
		}
	}
	if manaComp, ok := entity.GetComponent("mana"); ok {
		if mana, ok := manaComp.(*ManaComponent); ok {
			stats.mana = float64(mana.Max)
		}
	}
	if statsComp, ok := entity.GetComponent("stats"); ok {
		if s, ok := statsComp.(*StatsComponent); ok {
			stats.attack = s.Attack
			stats.defense = s.Defense
			stats.magicPower = s.MagicPower
		}
	}

	return stats
}

// calculateSecondaryGrowth computes stat growth at 50% effectiveness for secondary class.
func (cps *ClassProgressionSystem) calculateSecondaryGrowth(growth *StatGrowth, base statValues, level int) statValues {
	return statValues{
		hp:         (growth.CalculateHP(base.hp, level) - base.hp) * 0.5,
		mana:       (growth.CalculateMana(base.mana, level) - base.mana) * 0.5,
		attack:     (growth.CalculateAttack(base.attack, level) - base.attack) * 0.5,
		defense:    (growth.CalculateDefense(base.defense, level) - base.defense) * 0.5,
		magicPower: (growth.CalculateMagicPower(base.magicPower, level) - base.magicPower) * 0.5,
	}
}

// applySecondaryGrowth applies stat growth increases to entity components.
func (cps *ClassProgressionSystem) applySecondaryGrowth(entity *Entity, growth statValues) {
	cps.applyHealthGrowth(entity, growth.hp)
	cps.applyManaGrowth(entity, growth.mana)
	cps.applyStatGrowth(entity, growth)
	cps.applyAttackGrowth(entity, growth.attack)
}

// applyHealthGrowth increases max health while preserving current health percentage.
func (cps *ClassProgressionSystem) applyHealthGrowth(entity *Entity, hpGrowth float64) {
	if healthComp, ok := entity.GetComponent("health"); ok {
		if health, ok := healthComp.(*HealthComponent); ok {
			healthPercent := 1.0
			if health.Max > 0 {
				healthPercent = health.Current / health.Max
			}
			health.Max += hpGrowth
			health.Current = health.Max * healthPercent
		}
	}
}

// applyManaGrowth increases max mana while preserving current mana percentage.
func (cps *ClassProgressionSystem) applyManaGrowth(entity *Entity, manaGrowth float64) {
	if manaComp, ok := entity.GetComponent("mana"); ok {
		if mana, ok := manaComp.(*ManaComponent); ok {
			if mana.Max > 0 {
				manaPercent := float64(mana.Current) / float64(mana.Max)
				mana.Max += int(manaGrowth)
				mana.Current = int(float64(mana.Max) * manaPercent)
			}
		}
	}
}

// applyStatGrowth increases attack, defense, and magic power stats.
func (cps *ClassProgressionSystem) applyStatGrowth(entity *Entity, growth statValues) {
	if statsComp, ok := entity.GetComponent("stats"); ok {
		if stats, ok := statsComp.(*StatsComponent); ok {
			stats.Attack += growth.attack
			stats.Defense += growth.defense
			stats.MagicPower += growth.magicPower
		}
	}
}

// applyAttackGrowth increases attack component damage.
func (cps *ClassProgressionSystem) applyAttackGrowth(entity *Entity, attackGrowth float64) {
	if attackComp, ok := entity.GetComponent("attack"); ok {
		if attack, ok := attackComp.(*AttackComponent); ok {
			attack.Damage += attackGrowth
		}
	}
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
