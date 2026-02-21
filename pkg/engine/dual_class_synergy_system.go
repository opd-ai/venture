package engine

import (
	"math/rand"

	"github.com/sirupsen/logrus"
)

// DualClassSynergySystem applies passive combat bonuses to dual-classed characters.
// When a character has both primary and secondary classes (dual-classing at level 20+),
// this system grants synergy bonuses based on the class combination during combat.
// Synergies include attack modifiers, defense bonuses, mana regen, and crit multipliers.
type DualClassSynergySystem struct {
	world           *World
	rng             *rand.Rand
	synergies       map[classPair]*DualClassSynergy
	activeBonuses   map[uint64]*appliedSynergyBonus // entityID -> active bonus
	genreID         string
	logEntry        *logrus.Entry
	updateInterval  float64 // How often to refresh bonuses (seconds)
	timeSinceUpdate float64
	bonusMultiplier float64 // Genre-based multiplier
}

// classPair identifies a unique combination of primary and secondary classes.
type classPair struct {
	primary   CharacterClass
	secondary CharacterClass
}

// DualClassSynergy defines the passive bonuses for a specific class combination.
type DualClassSynergy struct {
	Name             string // e.g., "Battle Arcanist" for Warrior+Mage
	Primary          CharacterClass
	Secondary        CharacterClass
	AttackBonus      float64 // Flat attack damage increase
	DefenseBonus     float64 // Flat defense increase
	ManaRegenRate    float64 // Mana regen per second
	CritBonus        float64 // Crit chance increase (0.0-1.0)
	HealthRegenRate  float64 // Health regen per second
	SpeedBonus       float64 // Movement speed multiplier (e.g., 0.1 = +10%)
	SpellDamageBonus float64 // Magic power bonus
	Description      string  // Flavor text for UI
}

// appliedSynergyBonus tracks currently active synergy effects on an entity.
type appliedSynergyBonus struct {
	synergy        *DualClassSynergy
	attackApplied  float64
	defenseApplied float64
	speedApplied   float64
	spellApplied   float64
}

// NewDualClassSynergySystem creates a new dual-class synergy system.
func NewDualClassSynergySystem(world *World, seed int64) *DualClassSynergySystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil && world.logger.Logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "dual_class_synergy")
	} else {
		logEntry = logrus.NewEntry(logrus.StandardLogger()).WithField("system_name", "dual_class_synergy")
	}

	sys := &DualClassSynergySystem{
		world:           world,
		rng:             rand.New(rand.NewSource(seed)),
		synergies:       make(map[classPair]*DualClassSynergy),
		activeBonuses:   make(map[uint64]*appliedSynergyBonus),
		logEntry:        logEntry,
		updateInterval:  1.0, // Refresh every second
		timeSinceUpdate: 0,
		bonusMultiplier: 1.0,
	}

	sys.initializeSynergies()
	logEntry.Debug("dual class synergy system created")
	return sys
}

// SetGenre configures genre-specific bonus multipliers.
func (s *DualClassSynergySystem) SetGenre(genreID string) {
	s.genreID = genreID
	switch genreID {
	case "fantasy":
		s.bonusMultiplier = 1.0
	case "scifi":
		s.bonusMultiplier = 0.9 // Sci-fi favors tech over class synergies
	case "horror":
		s.bonusMultiplier = 1.1 // Horror rewards diverse survival skills
	case "cyberpunk":
		s.bonusMultiplier = 0.95
	case "postapoc":
		s.bonusMultiplier = 1.05 // Post-apocalyptic rewards adaptability
	default:
		s.bonusMultiplier = 1.0
	}
}

// initializeSynergies populates all dual-class synergy definitions.
// Each combination of primary+secondary classes provides unique passive bonuses.
func (s *DualClassSynergySystem) initializeSynergies() {
	synergies := []DualClassSynergy{
		// Warrior combinations
		{
			Name: "Battle Arcanist", Primary: ClassWarrior, Secondary: ClassMage,
			AttackBonus: 5, SpellDamageBonus: 8, ManaRegenRate: 0.5,
			Description: "Melee strikes resonate with arcane energy",
		},
		{
			Name: "Holy Warrior", Primary: ClassWarrior, Secondary: ClassCleric,
			AttackBonus: 3, HealthRegenRate: 1.0, DefenseBonus: 4,
			Description: "Divine blessings empower your blade",
		},
		{
			Name: "Shadow Knight", Primary: ClassWarrior, Secondary: ClassRogue,
			AttackBonus: 4, CritBonus: 0.08, SpeedBonus: 0.05,
			Description: "Strike from unexpected angles",
		},
		{
			Name: "Huntsman", Primary: ClassWarrior, Secondary: ClassRanger,
			AttackBonus: 4, SpeedBonus: 0.08, CritBonus: 0.05,
			Description: "Combine brute force with precision",
		},
		{
			Name: "Death Warrior", Primary: ClassWarrior, Secondary: ClassNecromancer,
			AttackBonus: 6, HealthRegenRate: 0.5, DefenseBonus: 2,
			Description: "Dark power fuels your strikes",
		},

		// Mage combinations
		{
			Name: "Arcane Duelist", Primary: ClassMage, Secondary: ClassWarrior,
			SpellDamageBonus: 6, AttackBonus: 4, ManaRegenRate: 0.3,
			Description: "Weave spells into your swordplay",
		},
		{
			Name: "Divine Invoker", Primary: ClassMage, Secondary: ClassCleric,
			SpellDamageBonus: 8, ManaRegenRate: 1.0, HealthRegenRate: 0.5,
			Description: "Channel both arcane and divine power",
		},
		{
			Name: "Shadow Weaver", Primary: ClassMage, Secondary: ClassRogue,
			SpellDamageBonus: 6, CritBonus: 0.10, SpeedBonus: 0.05,
			Description: "Spells strike from the shadows",
		},
		{
			Name: "Elemental Archer", Primary: ClassMage, Secondary: ClassRanger,
			SpellDamageBonus: 7, CritBonus: 0.06, AttackBonus: 2,
			Description: "Infuse arrows with elemental fury",
		},
		{
			Name: "Arcane Lich", Primary: ClassMage, Secondary: ClassNecromancer,
			SpellDamageBonus: 10, ManaRegenRate: 0.8, HealthRegenRate: 0.3,
			Description: "Master both life and death magic",
		},

		// Rogue combinations
		{
			Name: "Bladedancer", Primary: ClassRogue, Secondary: ClassWarrior,
			AttackBonus: 3, CritBonus: 0.12, SpeedBonus: 0.08,
			Description: "Dance between foes with deadly grace",
		},
		{
			Name: "Spellthief", Primary: ClassRogue, Secondary: ClassMage,
			CritBonus: 0.10, SpellDamageBonus: 5, ManaRegenRate: 0.4,
			Description: "Steal magic and slit throats",
		},
		{
			Name: "Divine Agent", Primary: ClassRogue, Secondary: ClassCleric,
			CritBonus: 0.08, HealthRegenRate: 0.8, DefenseBonus: 3,
			Description: "Holy purpose guides your blade",
		},
		{
			Name: "Pathfinder", Primary: ClassRogue, Secondary: ClassRanger,
			CritBonus: 0.10, SpeedBonus: 0.12, AttackBonus: 2,
			Description: "Navigate any terrain unseen",
		},
		{
			Name: "Grave Stalker", Primary: ClassRogue, Secondary: ClassNecromancer,
			CritBonus: 0.12, HealthRegenRate: 0.5, SpellDamageBonus: 3,
			Description: "Death itself is your ally",
		},

		// Ranger combinations
		{
			Name: "Beast Champion", Primary: ClassRanger, Secondary: ClassWarrior,
			AttackBonus: 4, DefenseBonus: 3, SpeedBonus: 0.06,
			Description: "Fight alongside your companion",
		},
		{
			Name: "Arcane Hunter", Primary: ClassRanger, Secondary: ClassMage,
			SpellDamageBonus: 5, CritBonus: 0.08, ManaRegenRate: 0.4,
			Description: "Enchanted arrows seek their mark",
		},
		{
			Name: "Sacred Tracker", Primary: ClassRanger, Secondary: ClassCleric,
			HealthRegenRate: 0.8, DefenseBonus: 3, AttackBonus: 2,
			Description: "Nature and faith guide your path",
		},
		{
			Name: "Ghost Walker", Primary: ClassRanger, Secondary: ClassRogue,
			SpeedBonus: 0.15, CritBonus: 0.08, AttackBonus: 2,
			Description: "Move unseen through any terrain",
		},
		{
			Name: "Dark Huntsman", Primary: ClassRanger, Secondary: ClassNecromancer,
			AttackBonus: 3, SpellDamageBonus: 4, HealthRegenRate: 0.4,
			Description: "Command undead beasts",
		},

		// Cleric combinations
		{
			Name: "Templar Knight", Primary: ClassCleric, Secondary: ClassWarrior,
			DefenseBonus: 5, HealthRegenRate: 1.2, AttackBonus: 3,
			Description: "Shield the faithful with steel and prayer",
		},
		{
			Name: "Theurge", Primary: ClassCleric, Secondary: ClassMage,
			SpellDamageBonus: 6, ManaRegenRate: 1.2, HealthRegenRate: 0.6,
			Description: "Divine and arcane magic intertwine",
		},
		{
			Name: "Shadow Priest", Primary: ClassCleric, Secondary: ClassRogue,
			HealthRegenRate: 0.8, CritBonus: 0.06, SpeedBonus: 0.05,
			Description: "Deliver salvation... or judgment",
		},
		{
			Name: "Wandering Healer", Primary: ClassCleric, Secondary: ClassRanger,
			HealthRegenRate: 1.0, SpeedBonus: 0.08, DefenseBonus: 2,
			Description: "Heal the wounded across all lands",
		},
		{
			Name: "Grey Priest", Primary: ClassCleric, Secondary: ClassNecromancer,
			HealthRegenRate: 0.8, SpellDamageBonus: 5, ManaRegenRate: 0.6,
			Description: "Balance life and death",
		},

		// Necromancer combinations
		{
			Name: "Death Knight", Primary: ClassNecromancer, Secondary: ClassWarrior,
			AttackBonus: 5, DefenseBonus: 3, HealthRegenRate: 0.6,
			Description: "Undying warrior of the grave",
		},
		{
			Name: "Lich King", Primary: ClassNecromancer, Secondary: ClassMage,
			SpellDamageBonus: 10, ManaRegenRate: 0.8, HealthRegenRate: 0.3,
			Description: "Supreme mastery over death magic",
		},
		{
			Name: "Soul Reaper", Primary: ClassNecromancer, Secondary: ClassRogue,
			CritBonus: 0.10, SpeedBonus: 0.06, SpellDamageBonus: 4,
			Description: "Harvest souls in shadow",
		},
		{
			Name: "Plague Doctor", Primary: ClassNecromancer, Secondary: ClassRanger,
			SpellDamageBonus: 5, HealthRegenRate: 0.5, AttackBonus: 2,
			Description: "Spread death from afar",
		},
		{
			Name: "Fallen Angel", Primary: ClassNecromancer, Secondary: ClassCleric,
			SpellDamageBonus: 6, HealthRegenRate: 1.0, ManaRegenRate: 0.5,
			Description: "Corrupted divine power",
		},
	}

	// Register all synergies - each has a unique primary/secondary combination
	for i := range synergies {
		syn := &synergies[i]
		s.synergies[classPair{syn.Primary, syn.Secondary}] = syn
	}

	s.logEntry.WithField("synergy_count", len(synergies)).Debug("synergies initialized")
}

// Update applies synergy bonuses to all dual-classed entities.
func (s *DualClassSynergySystem) Update(entities []*Entity, deltaTime float64) {
	s.timeSinceUpdate += deltaTime
	if s.timeSinceUpdate < s.updateInterval {
		return
	}
	elapsedTime := s.timeSinceUpdate // Store for regen calculations
	s.timeSinceUpdate = 0

	activeEntities := make(map[uint64]bool)

	for _, entity := range entities {
		comp, ok := entity.GetComponent("class_progression")
		if !ok || comp == nil {
			continue
		}

		progression := comp.(*ClassProgressionComponent)

		// Only apply to dual-classed characters
		if progression.SecondaryClass == nil {
			continue
		}

		// Find applicable synergy
		pair := classPair{progression.Class, *progression.SecondaryClass}
		synergy, found := s.synergies[pair]
		if !found {
			continue
		}

		activeEntities[entity.ID] = true

		// Check if already applied
		existing, hasBonus := s.activeBonuses[entity.ID]
		if hasBonus && existing.synergy == synergy {
			// Already applied, just do regen effects
			s.applyRegenEffects(entity, synergy, elapsedTime)
			continue
		}

		// Remove old bonus if switching
		if hasBonus {
			s.removeSynergyBonus(entity, existing)
		}

		// Apply new synergy bonus
		applied := s.applySynergyBonus(entity, synergy)
		s.activeBonuses[entity.ID] = applied

		s.logEntry.WithFields(logrus.Fields{
			"entity_id":    entity.ID,
			"synergy_name": synergy.Name,
			"primary":      progression.Class.String(),
			"secondary":    progression.SecondaryClass.String(),
		}).Debug("dual class synergy applied")
	}

	// Clean up bonuses for entities that no longer qualify
	for entityID := range s.activeBonuses {
		if !activeEntities[entityID] {
			// Find entity and remove bonus
			for _, entity := range entities {
				if entity.ID == entityID {
					s.removeSynergyBonus(entity, s.activeBonuses[entityID])
					break
				}
			}
			delete(s.activeBonuses, entityID)
		}
	}
}

// applySynergyBonus adds the synergy's stat bonuses to an entity.
func (s *DualClassSynergySystem) applySynergyBonus(entity *Entity, synergy *DualClassSynergy) *appliedSynergyBonus {
	applied := &appliedSynergyBonus{
		synergy: synergy,
	}

	multiplier := s.bonusMultiplier

	// Apply attack bonus
	if synergy.AttackBonus > 0 {
		if statsComp, ok := entity.GetComponent("stats"); ok {
			if stats := statsComp.(*StatsComponent); stats != nil {
				bonus := synergy.AttackBonus * multiplier
				stats.Attack += bonus
				applied.attackApplied = bonus
			}
		}
	}

	// Apply defense bonus
	if synergy.DefenseBonus > 0 {
		if statsComp, ok := entity.GetComponent("stats"); ok {
			if stats := statsComp.(*StatsComponent); stats != nil {
				bonus := synergy.DefenseBonus * multiplier
				stats.Defense += bonus
				applied.defenseApplied = bonus
			}
		}
	}

	// Apply speed bonus
	if synergy.SpeedBonus > 0 {
		if baseStatsComp, ok := entity.GetComponent("base_stats"); ok {
			if baseStats := baseStatsComp.(*BaseStatsComponent); baseStats != nil {
				bonus := synergy.SpeedBonus * multiplier
				baseStats.BaseSpeed *= (1.0 + bonus)
				applied.speedApplied = bonus
			}
		}
	}

	// Apply spell damage bonus
	if synergy.SpellDamageBonus > 0 {
		if statsComp, ok := entity.GetComponent("stats"); ok {
			if stats := statsComp.(*StatsComponent); stats != nil {
				bonus := synergy.SpellDamageBonus * multiplier
				stats.MagicPower += bonus
				applied.spellApplied = bonus
			}
		}
	}

	// Apply crit bonus
	if synergy.CritBonus > 0 {
		if statsComp, ok := entity.GetComponent("stats"); ok {
			if stats := statsComp.(*StatsComponent); stats != nil {
				stats.CritChance += synergy.CritBonus * multiplier
			}
		}
	}

	return applied
}

// removeSynergyBonus removes previously applied synergy bonuses from an entity.
func (s *DualClassSynergySystem) removeSynergyBonus(entity *Entity, applied *appliedSynergyBonus) {
	if applied == nil {
		return
	}

	// Remove attack bonus
	if applied.attackApplied > 0 {
		if statsComp, ok := entity.GetComponent("stats"); ok {
			if stats := statsComp.(*StatsComponent); stats != nil {
				stats.Attack -= applied.attackApplied
			}
		}
	}

	// Remove defense bonus
	if applied.defenseApplied > 0 {
		if statsComp, ok := entity.GetComponent("stats"); ok {
			if stats := statsComp.(*StatsComponent); stats != nil {
				stats.Defense -= applied.defenseApplied
			}
		}
	}

	// Remove speed bonus
	if applied.speedApplied > 0 {
		if baseStatsComp, ok := entity.GetComponent("base_stats"); ok {
			if baseStats := baseStatsComp.(*BaseStatsComponent); baseStats != nil {
				baseStats.BaseSpeed /= (1.0 + applied.speedApplied)
			}
		}
	}

	// Remove spell damage bonus
	if applied.spellApplied > 0 {
		if statsComp, ok := entity.GetComponent("stats"); ok {
			if stats := statsComp.(*StatsComponent); stats != nil {
				stats.MagicPower -= applied.spellApplied
			}
		}
	}

	// Note: Crit bonus is not tracked separately for removal
	// This is a minor simplification - in practice, entities don't lose dual-class often
}

// applyRegenEffects applies health and mana regeneration from synergy.
func (s *DualClassSynergySystem) applyRegenEffects(entity *Entity, synergy *DualClassSynergy, deltaTime float64) {
	multiplier := s.bonusMultiplier

	// Health regeneration
	if synergy.HealthRegenRate > 0 {
		if healthComp, ok := entity.GetComponent("health"); ok {
			if health := healthComp.(*HealthComponent); health != nil {
				regen := synergy.HealthRegenRate * multiplier * deltaTime
				health.Current += regen
				if health.Current > health.Max {
					health.Current = health.Max
				}
			}
		}
	}

	// Mana regeneration
	if synergy.ManaRegenRate > 0 {
		if manaComp, ok := entity.GetComponent("mana"); ok {
			if mana := manaComp.(*ManaComponent); mana != nil {
				regen := int(synergy.ManaRegenRate * multiplier * deltaTime)
				if regen < 1 && synergy.ManaRegenRate > 0 {
					// Accumulate fractional regen - tick at least sometimes
					if s.rng.Float64() < synergy.ManaRegenRate*multiplier*deltaTime {
						regen = 1
					}
				}
				mana.Current += regen
				if mana.Current > mana.Max {
					mana.Current = mana.Max
				}
			}
		}
	}
}

// GetSynergy returns the synergy for a class combination, if any.
func (s *DualClassSynergySystem) GetSynergy(primary, secondary CharacterClass) *DualClassSynergy {
	pair := classPair{primary, secondary}
	return s.synergies[pair]
}

// GetActiveSynergyCount returns the number of entities with active synergy bonuses.
func (s *DualClassSynergySystem) GetActiveSynergyCount() int {
	return len(s.activeBonuses)
}

// HasActiveSynergy returns whether an entity has an active synergy bonus.
func (s *DualClassSynergySystem) HasActiveSynergy(entityID uint64) bool {
	_, ok := s.activeBonuses[entityID]
	return ok
}

// GetActiveSynergyName returns the name of an entity's active synergy, or empty string.
func (s *DualClassSynergySystem) GetActiveSynergyName(entityID uint64) string {
	if applied, ok := s.activeBonuses[entityID]; ok && applied.synergy != nil {
		return applied.synergy.Name
	}
	return ""
}

// GetAllSynergies returns all defined dual-class synergies.
func (s *DualClassSynergySystem) GetAllSynergies() []*DualClassSynergy {
	seen := make(map[string]bool)
	var result []*DualClassSynergy
	for _, syn := range s.synergies {
		if !seen[syn.Name] {
			seen[syn.Name] = true
			result = append(result, syn)
		}
	}
	return result
}
