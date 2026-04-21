// Package engine provides the skill loadout system for managing saved skill configurations.
// This system handles loadout swapping, validation, and application to skill trees.
package engine

import (
	log "github.com/sirupsen/logrus"
)

// SkillLoadoutSystem manages skill loadout operations.
// It processes pending loadout restores and applies skill configurations
// to entities with both SkillLoadoutComponent and SkillTreeComponent.
type SkillLoadoutSystem struct {
	world          *World
	updateInterval int // Frames between updates
	frameCounter   int
}

// NewSkillLoadoutSystem creates a new skill loadout system.
func NewSkillLoadoutSystem(world *World) *SkillLoadoutSystem {
	log.WithFields(log.Fields{
		"system_name":     "skill_loadout",
		"update_interval": 10,
	}).Debug("Creating skill loadout system")

	return &SkillLoadoutSystem{
		world:          world,
		updateInterval: 10, // Check every 10 frames
		frameCounter:   0,
	}
}

// Update processes loadout operations for entities.
// Handles pending loadout restores and applies skill changes.
func (s *SkillLoadoutSystem) Update(entities []*Entity, deltaTime float64) {
	s.frameCounter++
	if s.frameCounter < s.updateInterval {
		return
	}
	s.frameCounter = 0

	currentTime := GetCurrentTime()
	entitiesProcessed := 0

	for _, entity := range entities {
		if s.processEntity(entity, currentTime) {
			entitiesProcessed++
		}
	}

	if entitiesProcessed > 0 && log.GetLevel() >= log.DebugLevel {
		log.WithFields(log.Fields{
			"system_name":        "skill_loadout",
			"entities_processed": entitiesProcessed,
		}).Debug("Processed loadout operations")
	}
}

// processEntity handles loadout operations for a single entity.
// Returns true if any processing was done.
func (s *SkillLoadoutSystem) processEntity(entity *Entity, currentTime float64) bool {
	// Get loadout component
	loadoutComp, hasLoadout := entity.GetComponent("skill_loadout")
	if !hasLoadout {
		return false
	}
	loadout, ok := loadoutComp.(*SkillLoadoutComponent)
	if !ok {
		return false
	}

	// Check for pending restore
	if loadout.PendingRestore < 0 {
		return false
	}

	// Get skill tree component
	treeComp, hasTree := entity.GetComponent("skill_tree")
	if !hasTree {
		log.WithFields(log.Fields{
			"entity_id":   entity.ID,
			"system_name": "skill_loadout",
		}).Debug("Entity missing skill_tree component, canceling restore")
		loadout.PendingRestore = -1
		return false
	}
	skillTree, ok := treeComp.(*SkillTreeComponent)
	if !ok {
		loadout.PendingRestore = -1
		return false
	}

	// Validate loadout compatibility
	treeID := ""
	if skillTree.Tree != nil {
		treeID = skillTree.Tree.ID
	}
	if err := loadout.ValidateLoadoutCompatibility(loadout.PendingRestore, treeID); err != nil {
		log.WithFields(log.Fields{
			"entity_id":     entity.ID,
			"loadout_index": loadout.PendingRestore,
			"error":         err.Error(),
			"system_name":   "skill_loadout",
		}).Warn("Loadout incompatible with skill tree")
		loadout.PendingRestore = -1
		return false
	}

	// Apply the loadout
	if s.applyLoadout(entity, loadout, skillTree, loadout.PendingRestore, currentTime) {
		log.WithFields(log.Fields{
			"entity_id":     entity.ID,
			"loadout_index": loadout.PendingRestore,
			"loadout_name":  loadout.GetLoadout(loadout.PendingRestore).Name,
			"system_name":   "skill_loadout",
		}).Info("Loadout restored successfully")
		return true
	}

	return false
}

// applyLoadout applies a saved loadout to the skill tree.
// This performs a full respec and applies the saved configuration.
func (s *SkillLoadoutSystem) applyLoadout(entity *Entity, loadout *SkillLoadoutComponent, skillTree *SkillTreeComponent, index int, currentTime float64) bool {
	savedLoadout := loadout.GetLoadout(index)
	if savedLoadout == nil {
		return false
	}

	// Get available skill points (from experience component or similar)
	totalAvailablePoints := s.getTotalSkillPoints(entity, skillTree)

	// Calculate total points needed for this loadout
	pointsNeeded := savedLoadout.TotalPointsUsed()
	if pointsNeeded > totalAvailablePoints {
		log.WithFields(log.Fields{
			"entity_id":        entity.ID,
			"points_needed":    pointsNeeded,
			"points_available": totalAvailablePoints,
			"system_name":      "skill_loadout",
		}).Warn("Not enough skill points for loadout")
		loadout.PendingRestore = -1
		return false
	}

	// Reset all skills (full respec)
	s.resetAllSkills(skillTree)

	// Apply skills from loadout in order (respecting dependencies)
	if !s.applySkillsFromLoadout(entity, skillTree, savedLoadout, totalAvailablePoints) {
		// Revert if application failed
		s.resetAllSkills(skillTree)
		loadout.PendingRestore = -1
		return false
	}

	// Mark swap complete
	loadout.MarkSwapComplete(index, currentTime)

	// Trigger skill bonus recalculation
	RecalculateSkillBonuses(entity)

	return true
}

// getTotalSkillPoints calculates total skill points available to entity.
func (s *SkillLoadoutSystem) getTotalSkillPoints(entity *Entity, skillTree *SkillTreeComponent) int {
	// Base points from level
	basePoints := 1

	// Get experience component for level-based points
	if expComp, hasExp := entity.GetComponent("experience"); hasExp {
		if exp, ok := expComp.(*ExperienceComponent); ok {
			basePoints = exp.Level // 1 point per level
		}
	}

	// Add back currently used points (they become available after respec)
	return basePoints + skillTree.TotalPointsUsed
}

// resetAllSkills unlearns all skills, returning points to available pool.
func (s *SkillLoadoutSystem) resetAllSkills(skillTree *SkillTreeComponent) {
	// Clear all learned skills
	skillTree.LearnedSkills = make(map[string]bool)
	skillTree.SkillLevels = make(map[string]int)
	skillTree.TotalPointsUsed = 0

	// Reset skill levels in the tree itself
	if skillTree.Tree != nil {
		for _, node := range skillTree.Tree.Nodes {
			if node.Skill != nil {
				node.Skill.Level = 0
			}
		}
	}
}

// applySkillsFromLoadout applies skills from a loadout in dependency order.
func (s *SkillLoadoutSystem) applySkillsFromLoadout(entity *Entity, skillTree *SkillTreeComponent, loadout *SkillLoadout, availablePoints int) bool {
	if skillTree.Tree == nil {
		return false
	}

	// Build dependency graph and apply in order
	applied := make(map[string]bool)
	maxIterations := len(loadout.SkillLevels) * 10 // Prevent infinite loops
	iterations := 0
	remaining := len(loadout.SkillLevels)

	for remaining > 0 && iterations < maxIterations {
		iterations++
		madeProgress := false

		for skillID, targetLevel := range loadout.SkillLevels {
			if applied[skillID] {
				continue
			}

			// Check if we can apply this skill now (prerequisites met)
			skill := skillTree.Tree.GetSkillByID(skillID)
			if skill == nil {
				// Skill not found in tree - skip it
				applied[skillID] = true
				remaining--
				continue
			}

			// Check prerequisites
			prereqsMet := true
			for _, prereqID := range skill.Requirements.PrerequisiteIDs {
				if !skillTree.LearnedSkills[prereqID] {
					// Check if prereq is even in our loadout
					if _, inLoadout := loadout.SkillLevels[prereqID]; inLoadout {
						prereqsMet = false // We'll apply it later
						break
					}
					// Prereq not in loadout - this loadout might be invalid
					prereqsMet = false
					break
				}
			}

			if !prereqsMet {
				continue
			}

			// Apply skill levels one at a time
			for level := 0; level < targetLevel; level++ {
				pointsRemaining := availablePoints - skillTree.TotalPointsUsed
				if !skillTree.LearnSkill(skillID, pointsRemaining) {
					log.WithFields(log.Fields{
						"skill_id":         skillID,
						"target_level":     targetLevel,
						"current_level":    level,
						"points_remaining": pointsRemaining,
						"system_name":      "skill_loadout",
					}).Debug("Failed to apply skill level")
					break
				}
			}

			applied[skillID] = true
			remaining--
			madeProgress = true
		}

		if !madeProgress && remaining > 0 {
			log.WithFields(log.Fields{
				"remaining_skills": remaining,
				"system_name":      "skill_loadout",
			}).Warn("Could not apply all skills from loadout - dependency issues")
			break
		}
	}

	return remaining == 0
}

// SaveCurrentAsLoadout is a helper to save the current skill configuration.
// Called externally (e.g., from UI) to create a new loadout.
func (s *SkillLoadoutSystem) SaveCurrentAsLoadout(entity *Entity, name, description string) int {
	loadoutComp, hasLoadout := entity.GetComponent("skill_loadout")
	if !hasLoadout {
		// Create component if missing
		newComp := NewSkillLoadoutComponent()
		entity.AddComponent(newComp)
		loadoutComp = newComp
	}
	loadout, ok := loadoutComp.(*SkillLoadoutComponent)
	if !ok {
		return -1
	}

	// Get skill tree
	treeComp, hasTree := entity.GetComponent("skill_tree")
	if !hasTree {
		return -1
	}
	skillTree, ok := treeComp.(*SkillTreeComponent)
	if !ok {
		return -1
	}

	// Get tree ID
	treeID := ""
	if skillTree.Tree != nil {
		treeID = skillTree.Tree.ID
	}

	currentTime := GetCurrentTime()
	return loadout.SaveLoadout(name, description, skillTree.SkillLevels, treeID, currentTime)
}

// RequestLoadoutSwap is a helper to initiate a loadout swap.
// Called externally (e.g., from UI or hotkey) to begin restoration.
func (s *SkillLoadoutSystem) RequestLoadoutSwap(entity *Entity, loadoutIndex int) bool {
	loadoutComp, hasLoadout := entity.GetComponent("skill_loadout")
	if !hasLoadout {
		return false
	}
	loadout, ok := loadoutComp.(*SkillLoadoutComponent)
	if !ok {
		return false
	}

	currentTime := GetCurrentTime()
	return loadout.RequestLoadoutRestore(loadoutIndex, currentTime)
}

// RequestQuickSwap is a helper to swap to a quick slot loadout.
// Called externally when player presses F1-F3.
func (s *SkillLoadoutSystem) RequestQuickSwap(entity *Entity, quickSlot int) bool {
	loadoutComp, hasLoadout := entity.GetComponent("skill_loadout")
	if !hasLoadout {
		return false
	}
	loadout, ok := loadoutComp.(*SkillLoadoutComponent)
	if !ok {
		return false
	}

	if quickSlot < 0 || quickSlot >= 3 {
		return false
	}

	loadoutIndex := loadout.QuickSlots[quickSlot]
	if loadoutIndex < 0 {
		return false // No loadout assigned
	}

	currentTime := GetCurrentTime()
	return loadout.RequestLoadoutRestore(loadoutIndex, currentTime)
}

// GetLoadoutStatus returns status info for UI display.
func (s *SkillLoadoutSystem) GetLoadoutStatus(entity *Entity) *LoadoutStatus {
	loadoutComp, hasLoadout := entity.GetComponent("skill_loadout")
	if !hasLoadout {
		return nil
	}
	loadout, ok := loadoutComp.(*SkillLoadoutComponent)
	if !ok {
		return nil
	}

	currentTime := GetCurrentTime()
	return &LoadoutStatus{
		LoadoutCount:      loadout.GetLoadoutCount(),
		UnlockedSlots:     loadout.UnlockedSlots,
		AvailableSlots:    loadout.GetAvailableSlots(),
		ActiveIndex:       loadout.ActiveIndex,
		CanSwap:           loadout.CanSwapLoadout(currentTime),
		CooldownRemaining: loadout.GetSwapCooldownRemaining(currentTime),
		SwapCost:          loadout.SwapCost,
		LoadoutNames:      loadout.GetLoadoutNames(),
	}
}

// LoadoutStatus contains UI-friendly loadout information.
type LoadoutStatus struct {
	LoadoutCount      int
	UnlockedSlots     int
	AvailableSlots    int
	ActiveIndex       int
	CanSwap           bool
	CooldownRemaining float64
	SwapCost          int
	LoadoutNames      []string
}

// EnsureLoadoutComponent adds a SkillLoadoutComponent to an entity if missing.
// Called during entity initialization or when accessing loadout features.
func EnsureLoadoutComponent(entity *Entity) *SkillLoadoutComponent {
	if entity == nil {
		return nil
	}

	loadoutComp, hasLoadout := entity.GetComponent("skill_loadout")
	if hasLoadout {
		if loadout, ok := loadoutComp.(*SkillLoadoutComponent); ok {
			return loadout
		}
	}

	newComp := NewSkillLoadoutComponent()
	entity.AddComponent(newComp)
	return newComp
}
