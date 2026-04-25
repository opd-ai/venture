package engine

import (
	"strconv"

	"github.com/opd-ai/venture/pkg/class/advanced"
)

// AdvancedClassSystem integrates the advanced class manager with ECS.
// Handles multi-classing, prestige classes, and talent trees for deep character customization.
type AdvancedClassSystem struct {
	manager     *advanced.Manager
	world       *World
	lastApplied map[uint64]advanced.StatBonuses // G32: guards per-frame accumulation
}

// NewAdvancedClassSystem creates a new advanced class system
func NewAdvancedClassSystem(world *World) *AdvancedClassSystem {
	return &AdvancedClassSystem{
		manager:     advanced.NewManager(),
		world:       world,
		lastApplied: make(map[uint64]advanced.StatBonuses),
	}
}

// Update applies stat bonuses from classes and talents to entities
// Pre-filters entities with "advanced_class" using the world query cache to avoid
// O(n) uncached map lookups across all entities every frame.
func (acs *AdvancedClassSystem) Update(entities []*Entity, deltaTime float64) {
	// Use the world query cache to skip non-advanced-class entities.
	classEntities := entities
	if acs.world != nil {
		classEntities = acs.world.GetEntitiesWith("advanced_class")
	}

	for _, entity := range classEntities {
		comp, ok := entity.GetComponent("advanced_class")
		if !ok || comp == nil {
			continue
		}

		playerID := strconv.FormatUint(entity.ID, 10)

		stats, err := acs.manager.CalculateTotalStats(playerID)
		if err != nil {
			continue
		}

		acs.applyStatBonuses(entity, stats)
	}
}

// applyStatBonuses applies calculated stat bonuses to entity components.
// G32 fix: subtract the previously applied bonus before adding the new one so
// that calling this method N times per session is idempotent.
func (acs *AdvancedClassSystem) applyStatBonuses(entity *Entity, bonuses advanced.StatBonuses) {
	prev := acs.lastApplied[entity.ID]
	acs.applyHealthBonuses(entity, prev, bonuses)
	acs.applyManaBonuses(entity, prev, bonuses)
	acs.applyStatsBonuses(entity, prev, bonuses)
	acs.lastApplied[entity.ID] = bonuses
}

// applyHealthBonuses applies health bonuses to an entity's health component.
// prev contains the bonuses applied on the previous call; they are subtracted
// before the new bonuses are added so the net change is always (new - prev).
func (acs *AdvancedClassSystem) applyHealthBonuses(entity *Entity, prev, bonuses advanced.StatBonuses) {
	healthComp, ok := entity.GetComponent("health")
	if !ok {
		return
	}
	health, ok := healthComp.(*HealthComponent)
	if !ok {
		return
	}
	health.Max -= float64(prev.Health)
	health.Max += float64(bonuses.Health)
	if health.Current > health.Max {
		health.Current = health.Max
	}
}

// applyManaBonuses applies mana bonuses to an entity's mana component.
// prev contains the bonuses applied on the previous call; they are subtracted
// before the new bonuses are added so the net change is always (new - prev).
func (acs *AdvancedClassSystem) applyManaBonuses(entity *Entity, prev, bonuses advanced.StatBonuses) {
	manaComp, ok := entity.GetComponent("mana")
	if !ok {
		return
	}
	mana, ok := manaComp.(*ManaComponent)
	if !ok {
		return
	}
	mana.Max -= prev.Mana
	mana.Max += bonuses.Mana
	if mana.Current > mana.Max {
		mana.Current = mana.Max
	}
}

// applyStatsBonuses applies attribute bonuses to an entity's stats component.
// prev contains the bonuses applied on the previous call; they are subtracted
// before the new bonuses are added so the net change is always (new - prev).
func (acs *AdvancedClassSystem) applyStatsBonuses(entity *Entity, prev, bonuses advanced.StatBonuses) {
	statsComp, ok := entity.GetComponent("stats")
	if !ok {
		return
	}
	stats, ok := statsComp.(*StatsComponent)
	if !ok {
		return
	}
	stats.Attack -= float64(prev.Strength)
	stats.Defense -= float64(prev.Defense)
	stats.MagicPower -= float64(prev.Intelligence)
	stats.CritChance -= prev.CritChance
	stats.CritDamage -= prev.CritDamage

	stats.Attack += float64(bonuses.Strength)
	stats.Defense += float64(bonuses.Defense)
	stats.MagicPower += float64(bonuses.Intelligence)
	stats.CritChance += bonuses.CritChance
	stats.CritDamage += bonuses.CritDamage
}

// InitializePlayerClass sets up a player's advanced class configuration
func (acs *AdvancedClassSystem) InitializePlayerClass(entity *Entity, primary advanced.ClassID, level int) error {
	playerID := strconv.FormatUint(entity.ID, 10)

	if err := acs.manager.SetPrimaryClass(playerID, primary); err != nil {
		return err
	}

	if err := acs.manager.SetLevel(playerID, level); err != nil {
		return err
	}

	entity.AddComponent(&advanced.AdvancedClassComponent{
		PrimaryClass: primary,
		Level:        level,
		TalentPoints: advanced.TalentAllocation{
			Talents:     make(map[advanced.TalentID]int),
			PointsTotal: level,
		},
	})

	return nil
}

// SetSecondaryClass enables multi-classing for a player
func (acs *AdvancedClassSystem) SetSecondaryClass(entity *Entity, secondary advanced.ClassID) error {
	playerID := strconv.FormatUint(entity.ID, 10)

	if err := acs.manager.SetSecondaryClass(playerID, secondary); err != nil {
		return err
	}

	comp, ok := entity.GetComponent("advanced_class")
	if ok && comp != nil {
		advClass := comp.(*advanced.AdvancedClassComponent)
		advClass.SecondaryClass = secondary
	}

	return nil
}

// SetPrestigeClass assigns a prestige class if requirements are met
func (acs *AdvancedClassSystem) SetPrestigeClass(entity *Entity, prestige advanced.PrestigeClassID) error {
	playerID := strconv.FormatUint(entity.ID, 10)

	if err := acs.manager.SetPrestigeClass(playerID, prestige); err != nil {
		return err
	}

	comp, ok := entity.GetComponent("advanced_class")
	if ok && comp != nil {
		advClass := comp.(*advanced.AdvancedClassComponent)
		advClass.PrestigeClass = prestige
	}

	return nil
}

// AllocateTalent adds a point to a talent
func (acs *AdvancedClassSystem) AllocateTalent(entity *Entity, talentID advanced.TalentID) error {
	playerID := strconv.FormatUint(entity.ID, 10)

	if err := acs.manager.AllocateTalent(playerID, talentID); err != nil {
		return err
	}

	comp, ok := entity.GetComponent("advanced_class")
	if ok && comp != nil {
		advClass := comp.(*advanced.AdvancedClassComponent)
		advClass.TalentPoints.Talents[talentID]++
		advClass.TalentPoints.PointsSpent++
	}

	return nil
}

// RespecTalents resets all talent points for a gold cost
func (acs *AdvancedClassSystem) RespecTalents(entity *Entity, gold int) error {
	playerID := strconv.FormatUint(entity.ID, 10)

	if err := acs.manager.RespecTalents(playerID, gold); err != nil {
		return err
	}

	comp, ok := entity.GetComponent("advanced_class")
	if ok && comp != nil {
		advClass := comp.(*advanced.AdvancedClassComponent)
		advClass.TalentPoints.Talents = make(map[advanced.TalentID]int)
		advClass.TalentPoints.PointsSpent = 0
		advClass.RespecCount++
	}

	return nil
}

// GetRespecCost returns the gold cost for a player's next respec
func (acs *AdvancedClassSystem) GetRespecCost(entity *Entity) int {
	playerID := strconv.FormatUint(entity.ID, 10)
	return acs.manager.GetRespecCost(playerID)
}

// GetTalentTree returns the talent tree for a class
func (acs *AdvancedClassSystem) GetTalentTree(classID advanced.ClassID) (*advanced.TalentTree, error) {
	return acs.manager.GetTalentTree(classID)
}

// GetPlayerClass returns a player's class configuration
func (acs *AdvancedClassSystem) GetPlayerClass(entity *Entity) (*advanced.AdvancedClassComponent, error) {
	playerID := strconv.FormatUint(entity.ID, 10)
	return acs.manager.GetPlayerClass(playerID)
}

// LevelUp increases the player's level and awards talent points
func (acs *AdvancedClassSystem) LevelUp(entity *Entity) error {
	playerID := strconv.FormatUint(entity.ID, 10)

	comp, ok := entity.GetComponent("advanced_class")
	if !ok || comp == nil {
		return nil
	}

	advClass := comp.(*advanced.AdvancedClassComponent)
	newLevel := advClass.Level + 1

	if err := acs.manager.SetLevel(playerID, newLevel); err != nil {
		return err
	}

	advClass.Level = newLevel
	advClass.TalentPoints.PointsTotal = newLevel

	return nil
}

// GetAllSynergies returns all class synergy bonuses
func (acs *AdvancedClassSystem) GetAllSynergies() []advanced.SynergyBonus {
	return acs.manager.GetAllSynergies()
}
