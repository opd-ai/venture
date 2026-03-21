// Package engine provides the build template system for managing character build presets.
// This system handles saving current builds as templates, loading templates to apply
// complete respecs, and generating archetype-based preset builds.
package engine

import (
	"fmt"
	"math/rand"
	"time"

	log "github.com/sirupsen/logrus"
)

// BuildTemplateSystem manages build template operations.
// It processes pending template applications and handles save/load operations.
type BuildTemplateSystem struct {
	world          *World
	seed           int64
	updateInterval int
	frameCounter   int

	// Callbacks for template events
	onTemplateApplied func(entity *Entity, template *BuildTemplate)
	onTemplateSaved   func(entity *Entity, template *BuildTemplate)

	// Preset templates per archetype
	archetypePresets map[BuildArchetype]*BuildTemplate
}

// NewBuildTemplateSystem creates a new build template system.
func NewBuildTemplateSystem(world *World, seed int64) *BuildTemplateSystem {
	log.WithFields(log.Fields{
		"system_name":     "build_template",
		"seed":            seed,
		"update_interval": 10,
	}).Debug("Creating build template system")

	sys := &BuildTemplateSystem{
		world:            world,
		seed:             seed,
		updateInterval:   10,
		frameCounter:     0,
		archetypePresets: make(map[BuildArchetype]*BuildTemplate),
	}

	// Generate preset templates for each archetype
	sys.generateArchetypePresets()

	return sys
}

// Update processes template operations for entities.
func (s *BuildTemplateSystem) Update(entities []*Entity, deltaTime float64) {
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

	if entitiesProcessed > 0 {
		log.WithFields(log.Fields{
			"system_name":        "build_template",
			"entities_processed": entitiesProcessed,
		}).Debug("Processed build template operations")
	}
}

// processEntity handles template operations for a single entity.
func (s *BuildTemplateSystem) processEntity(entity *Entity, currentTime float64) bool {
	comp, hasComp := entity.GetComponent("build_template")
	if !hasComp {
		return false
	}
	buildComp, ok := comp.(*BuildTemplateComponent)
	if !ok || buildComp.PendingApply < 0 {
		return false
	}

	template := buildComp.GetTemplate(buildComp.PendingApply)
	if template == nil {
		buildComp.PendingApply = -1
		return false
	}

	// Check level requirement
	level := s.getEntityLevel(entity)
	if level < template.RequiredLevel {
		log.WithFields(log.Fields{
			"system_name":    "build_template",
			"entity_id":      entity.ID,
			"template_name":  template.Name,
			"required_level": template.RequiredLevel,
			"current_level":  level,
		}).Warn("Level too low to apply build template")
		buildComp.PendingApply = -1
		return false
	}

	// Apply the template
	if s.applyTemplate(entity, template, currentTime) {
		buildComp.MarkApplyComplete(template.ID, currentTime)
		if s.onTemplateApplied != nil {
			s.onTemplateApplied(entity, template)
		}
		log.WithFields(log.Fields{
			"system_name":   "build_template",
			"entity_id":     entity.ID,
			"template_name": template.Name,
			"template_id":   template.ID,
		}).Info("Build template applied successfully")
		return true
	}

	buildComp.PendingApply = -1
	return false
}

// applyTemplate applies a build template to an entity.
func (s *BuildTemplateSystem) applyTemplate(entity *Entity, template *BuildTemplate, currentTime float64) bool {
	// Step 1: Reset attributes
	if !s.resetAndApplyAttributes(entity, template) {
		log.WithFields(log.Fields{
			"system_name": "build_template",
			"entity_id":   entity.ID,
			"step":        "attributes",
		}).Debug("Failed to apply attribute template")
		return false
	}

	// Step 2: Reset and apply talents
	if !s.resetAndApplyTalents(entity, template) {
		log.WithFields(log.Fields{
			"system_name": "build_template",
			"entity_id":   entity.ID,
			"step":        "talents",
		}).Debug("Failed to apply talent template")
		return false
	}

	// Step 3: Reset and apply skills
	if !s.resetAndApplySkills(entity, template) {
		log.WithFields(log.Fields{
			"system_name": "build_template",
			"entity_id":   entity.ID,
			"step":        "skills",
		}).Debug("Failed to apply skill template")
		return false
	}

	return true
}

// resetAndApplyAttributes resets attributes and applies template values.
func (s *BuildTemplateSystem) resetAndApplyAttributes(entity *Entity, template *BuildTemplate) bool {
	attrComp, hasAttr := entity.GetComponent("attribute_allocation")
	if !hasAttr {
		// Create component if missing
		newComp := NewAttributeAllocationComponent()
		entity.AddComponent(newComp)
		attrComp = newComp
	}
	attr, ok := attrComp.(*AttributeAllocationComponent)
	if !ok {
		return false
	}

	// Calculate total available points (unspent + currently allocated)
	totalAvailable := attr.UnspentPoints + attr.TotalAllocatedPoints()

	// Calculate points needed for template
	totalNeeded := template.TotalAttributePoints()
	if totalNeeded > totalAvailable {
		log.WithFields(log.Fields{
			"system_name":      "build_template",
			"entity_id":        entity.ID,
			"points_needed":    totalNeeded,
			"points_available": totalAvailable,
		}).Debug("Not enough attribute points for template")
		return false
	}

	// Reset all allocated points
	for i := 0; i < int(NumCoreAttributes); i++ {
		attr.UnspentPoints += attr.AllocatedPoints[i]
		attr.AllocatedPoints[i] = 0
	}
	attr.RespecCount++
	attr.LastModifiedTime = float64(time.Now().Unix())

	// Apply template attribute allocations
	for attrIndex, points := range template.Attributes {
		if attrIndex < 0 || attrIndex >= int(NumCoreAttributes) {
			continue
		}
		if points > attr.UnspentPoints {
			points = attr.UnspentPoints
		}
		attr.AllocatedPoints[attrIndex] = points
		attr.UnspentPoints -= points
	}

	attr.Dirty = true
	return true
}

// resetAndApplyTalents resets talents and applies template values.
func (s *BuildTemplateSystem) resetAndApplyTalents(entity *Entity, template *BuildTemplate) bool {
	talentComp, hasTalent := entity.GetComponent("talent")
	if !hasTalent {
		// Create component if missing
		newComp := NewTalentComponent()
		entity.AddComponent(newComp)
		talentComp = newComp
	}
	talent, ok := talentComp.(*TalentComponent)
	if !ok {
		return false
	}

	// Calculate total available points
	totalSpent := 0
	for _, ranks := range talent.Allocations {
		totalSpent += ranks
	}
	totalAvailable := talent.UnspentPoints + totalSpent

	// Calculate points needed
	totalNeeded := template.TotalTalentPoints()
	if totalNeeded > totalAvailable {
		log.WithFields(log.Fields{
			"system_name":      "build_template",
			"entity_id":        entity.ID,
			"points_needed":    totalNeeded,
			"points_available": totalAvailable,
		}).Debug("Not enough talent points for template")
		return false
	}

	// Reset all talents
	talent.ResetAll()

	// Get entity level for talent prerequisites
	level := s.getEntityLevel(entity)

	// Apply template talent allocations (in order to respect prerequisites)
	// Multiple passes to handle dependencies
	remaining := make(map[string]int)
	for talentID, ranks := range template.Talents {
		remaining[talentID] = ranks
	}

	maxIterations := len(remaining) * 5
	for iterations := 0; len(remaining) > 0 && iterations < maxIterations; iterations++ {
		madeProgress := false
		for talentID, targetRanks := range remaining {
			def := GetTalentDefinition(talentID)
			if def == nil {
				delete(remaining, talentID)
				continue
			}

			// Apply ranks one at a time
			for talent.Allocations[talentID] < targetRanks {
				if talent.CanAllocate(def, level) {
					talent.AllocatePoint(def, level)
					madeProgress = true
				} else {
					break
				}
			}

			if talent.Allocations[talentID] >= targetRanks {
				delete(remaining, talentID)
			}
		}
		if !madeProgress {
			break
		}
	}

	talent.Dirty = true
	return true
}

// resetAndApplySkills resets skills and applies template values.
func (s *BuildTemplateSystem) resetAndApplySkills(entity *Entity, template *BuildTemplate) bool {
	skillTree, ok := s.validateSkillTree(entity, template)
	if !ok {
		return len(template.Skills) == 0
	}

	if !s.validateSkillTreeCompatibility(entity.ID, skillTree, template) {
		return false
	}

	totalAvailable := s.getTotalSkillPoints(entity, skillTree)
	if !s.validateSkillPoints(entity.ID, totalAvailable, template.TotalSkillPoints()) {
		return false
	}

	s.clearSkillTree(skillTree)
	s.applySkillsFromTemplate(skillTree, template, totalAvailable)

	RecalculateSkillBonuses(entity)
	return true
}

// validateSkillTree retrieves and validates the skill tree component.
func (s *BuildTemplateSystem) validateSkillTree(entity *Entity, template *BuildTemplate) (*SkillTreeComponent, bool) {
	treeComp, hasTree := entity.GetComponent("skill_tree")
	if !hasTree {
		return nil, false
	}
	skillTree, ok := treeComp.(*SkillTreeComponent)
	return skillTree, ok
}

// validateSkillTreeCompatibility checks if the skill tree ID matches the template.
func (s *BuildTemplateSystem) validateSkillTreeCompatibility(entityID uint64, skillTree *SkillTreeComponent, template *BuildTemplate) bool {
	treeID := ""
	if skillTree.Tree != nil {
		treeID = skillTree.Tree.ID
	}
	if template.SkillTreeID != "" && treeID != template.SkillTreeID {
		log.WithFields(log.Fields{
			"system_name":   "build_template",
			"entity_id":     entityID,
			"template_tree": template.SkillTreeID,
			"entity_tree":   treeID,
		}).Debug("Skill tree mismatch for template")
		return false
	}
	return true
}

// validateSkillPoints checks if there are enough points available.
func (s *BuildTemplateSystem) validateSkillPoints(entityID uint64, available, needed int) bool {
	if needed > available {
		log.WithFields(log.Fields{
			"system_name":      "build_template",
			"entity_id":        entityID,
			"points_needed":    needed,
			"points_available": available,
		}).Debug("Not enough skill points for template")
		return false
	}
	return true
}

// clearSkillTree resets all skills to their initial state.
func (s *BuildTemplateSystem) clearSkillTree(skillTree *SkillTreeComponent) {
	skillTree.LearnedSkills = make(map[string]bool)
	skillTree.SkillLevels = make(map[string]int)
	skillTree.TotalPointsUsed = 0
	if skillTree.Tree != nil {
		for _, node := range skillTree.Tree.Nodes {
			if node.Skill != nil {
				node.Skill.Level = 0
			}
		}
	}
}

// applySkillsFromTemplate applies skills respecting prerequisites via iterative resolution.
func (s *BuildTemplateSystem) applySkillsFromTemplate(skillTree *SkillTreeComponent, template *BuildTemplate, totalAvailable int) {
	applied := make(map[string]bool)
	remaining := make(map[string]int)
	for skillID, level := range template.Skills {
		remaining[skillID] = level
	}

	maxIterations := len(remaining) * 10
	for iterations := 0; len(remaining) > 0 && iterations < maxIterations; iterations++ {
		if !s.applySkillIteration(skillTree, remaining, applied, totalAvailable) {
			break
		}
	}
}

// applySkillIteration processes one pass of skill application, returning true if progress was made.
func (s *BuildTemplateSystem) applySkillIteration(skillTree *SkillTreeComponent, remaining map[string]int, applied map[string]bool, totalAvailable int) bool {
	madeProgress := false
	for skillID, targetLevel := range remaining {
		if applied[skillID] {
			delete(remaining, skillID)
			continue
		}

		if !s.checkSkillPrerequisites(skillTree, skillID) {
			continue
		}

		if s.applySkillLevels(skillTree, skillID, targetLevel, totalAvailable) {
			madeProgress = true
		}

		if skillTree.SkillLevels[skillID] >= targetLevel {
			applied[skillID] = true
			delete(remaining, skillID)
		}
	}
	return madeProgress
}

// checkSkillPrerequisites verifies all prerequisites for a skill are met.
func (s *BuildTemplateSystem) checkSkillPrerequisites(skillTree *SkillTreeComponent, skillID string) bool {
	if skillTree.Tree == nil {
		return true
	}
	skill := skillTree.Tree.GetSkillByID(skillID)
	if skill == nil {
		return false
	}
	for _, prereqID := range skill.Requirements.PrerequisiteIDs {
		if !skillTree.LearnedSkills[prereqID] {
			return false
		}
	}
	return true
}

// applySkillLevels attempts to learn skill levels up to the target.
func (s *BuildTemplateSystem) applySkillLevels(skillTree *SkillTreeComponent, skillID string, targetLevel, totalAvailable int) bool {
	madeProgress := false
	for lvl := 0; lvl < targetLevel; lvl++ {
		pointsRemaining := totalAvailable - skillTree.TotalPointsUsed
		if skillTree.LearnSkill(skillID, pointsRemaining) {
			madeProgress = true
		} else {
			break
		}
	}
	return madeProgress
}

// getTotalSkillPoints calculates available skill points for an entity.
func (s *BuildTemplateSystem) getTotalSkillPoints(entity *Entity, skillTree *SkillTreeComponent) int {
	basePoints := 1
	if expComp, hasExp := entity.GetComponent("experience"); hasExp {
		if exp, ok := expComp.(*ExperienceComponent); ok {
			basePoints = exp.Level
		}
	}
	return basePoints + skillTree.TotalPointsUsed
}

// getEntityLevel returns the entity's current level.
func (s *BuildTemplateSystem) getEntityLevel(entity *Entity) int {
	expComp, hasExp := entity.GetComponent("experience")
	if !hasExp {
		return 1
	}
	exp, ok := expComp.(*ExperienceComponent)
	if !ok {
		return 1
	}
	return exp.Level
}

// SaveCurrentBuild saves the entity's current build as a template.
func (s *BuildTemplateSystem) SaveCurrentBuild(entity *Entity, name, description string) (*BuildTemplate, error) {
	buildComp := s.ensureBuildTemplateComponent(entity)
	if buildComp == nil {
		return nil, fmt.Errorf("could not create build template component")
	}

	if buildComp.GetAvailableSlots() <= 0 {
		return nil, fmt.Errorf("no available template slots")
	}

	template := s.createEmptyTemplate(entity.ID, name, description)
	s.captureEntityAttributes(entity, template)
	s.captureEntityTalents(entity, template)
	s.captureEntitySkills(entity, template)
	s.captureEntityClassInfo(entity, template)
	template.RequiredLevel = s.calculateRequiredLevel(template)

	index := buildComp.AddTemplate(template)
	if index < 0 {
		return nil, fmt.Errorf("failed to add template")
	}

	if s.onTemplateSaved != nil {
		s.onTemplateSaved(entity, template)
	}

	log.WithFields(log.Fields{
		"system_name":   "build_template",
		"entity_id":     entity.ID,
		"template_name": name,
		"template_id":   template.ID,
	}).Info("Build template saved")

	return template, nil
}

// createEmptyTemplate initializes a new build template with metadata.
func (s *BuildTemplateSystem) createEmptyTemplate(entityID uint64, name, description string) *BuildTemplate {
	return &BuildTemplate{
		ID:          fmt.Sprintf("user_%d_%d", entityID, time.Now().UnixNano()),
		Name:        name,
		Description: description,
		Archetype:   BuildArchetypeCustom,
		Attributes:  make(map[int]int),
		Talents:     make(map[string]int),
		Skills:      make(map[string]int),
		CreatedAt:   time.Now().Unix(),
		UpdatedAt:   time.Now().Unix(),
		IsPreset:    false,
	}
}

// captureEntityAttributes copies allocated attribute points to the template.
func (s *BuildTemplateSystem) captureEntityAttributes(entity *Entity, template *BuildTemplate) {
	attrComp, hasAttr := entity.GetComponent("attribute_allocation")
	if !hasAttr {
		return
	}
	attr, ok := attrComp.(*AttributeAllocationComponent)
	if !ok {
		return
	}
	for i := 0; i < int(NumCoreAttributes); i++ {
		if attr.AllocatedPoints[i] > 0 {
			template.Attributes[i] = attr.AllocatedPoints[i]
		}
	}
}

// captureEntityTalents copies talent allocations to the template.
func (s *BuildTemplateSystem) captureEntityTalents(entity *Entity, template *BuildTemplate) {
	talentComp, hasTalent := entity.GetComponent("talent")
	if !hasTalent {
		return
	}
	talent, ok := talentComp.(*TalentComponent)
	if !ok {
		return
	}
	for talentID, ranks := range talent.Allocations {
		if ranks > 0 {
			template.Talents[talentID] = ranks
		}
	}
}

// captureEntitySkills copies skill levels and tree ID to the template.
func (s *BuildTemplateSystem) captureEntitySkills(entity *Entity, template *BuildTemplate) {
	treeComp, hasTree := entity.GetComponent("skill_tree")
	if !hasTree {
		return
	}
	skillTree, ok := treeComp.(*SkillTreeComponent)
	if !ok {
		return
	}
	for skillID, level := range skillTree.SkillLevels {
		if level > 0 {
			template.Skills[skillID] = level
		}
	}
	if skillTree.Tree != nil {
		template.SkillTreeID = skillTree.Tree.ID
	}
}

// captureEntityClassInfo copies class and specialization info to the template.
func (s *BuildTemplateSystem) captureEntityClassInfo(entity *Entity, template *BuildTemplate) {
	classComp, hasClass := entity.GetComponent("class_progression")
	if !hasClass {
		return
	}
	progression, ok := classComp.(*ClassProgressionComponent)
	if !ok {
		return
	}
	template.Class = progression.Class
	template.Specialization = progression.Specialization
	if progression.SecondaryClass != nil {
		template.SecondaryClass = progression.SecondaryClass
		template.SecondarySpec = progression.SecondarySpec
	}
}

// calculateRequiredLevel estimates the minimum level for a build.
func (s *BuildTemplateSystem) calculateRequiredLevel(template *BuildTemplate) int {
	// Attributes: 3 points per level
	attrPoints := template.TotalAttributePoints()
	attrLevels := (attrPoints + 2) / 3

	// Talents: 1 point per level
	talentLevels := template.TotalTalentPoints()

	// Skills: 1 point per level
	skillLevels := template.TotalSkillPoints()

	// Return the maximum
	maxLevel := attrLevels
	if talentLevels > maxLevel {
		maxLevel = talentLevels
	}
	if skillLevels > maxLevel {
		maxLevel = skillLevels
	}

	if maxLevel < 1 {
		return 1
	}
	return maxLevel
}

// RequestApplyTemplate queues a template for application.
func (s *BuildTemplateSystem) RequestApplyTemplate(entity *Entity, templateIndex int) bool {
	buildComp := s.ensureBuildTemplateComponent(entity)
	if buildComp == nil {
		return false
	}

	currentTime := GetCurrentTime()
	return buildComp.RequestApply(templateIndex, currentTime)
}

// RequestApplyPreset queues a preset archetype template for application.
func (s *BuildTemplateSystem) RequestApplyPreset(entity *Entity, archetype BuildArchetype) bool {
	preset := s.archetypePresets[archetype]
	if preset == nil {
		return false
	}

	buildComp := s.ensureBuildTemplateComponent(entity)
	if buildComp == nil {
		return false
	}

	// Find or add the preset to the component
	for i, t := range buildComp.Templates {
		if t.ID == preset.ID {
			currentTime := GetCurrentTime()
			return buildComp.RequestApply(i, currentTime)
		}
	}

	// Add preset as a template
	index := buildComp.AddTemplate(preset.Clone())
	if index < 0 {
		return false
	}

	currentTime := GetCurrentTime()
	return buildComp.RequestApply(index, currentTime)
}

// GetArchetypePreset returns a copy of the preset for an archetype.
func (s *BuildTemplateSystem) GetArchetypePreset(archetype BuildArchetype) *BuildTemplate {
	preset := s.archetypePresets[archetype]
	if preset == nil {
		return nil
	}
	return preset.Clone()
}

// ensureBuildTemplateComponent creates the component if missing.
func (s *BuildTemplateSystem) ensureBuildTemplateComponent(entity *Entity) *BuildTemplateComponent {
	if entity == nil {
		return nil
	}

	comp, hasComp := entity.GetComponent("build_template")
	if hasComp {
		if buildComp, ok := comp.(*BuildTemplateComponent); ok {
			return buildComp
		}
	}

	newComp := NewBuildTemplateComponent()
	entity.AddComponent(newComp)
	return newComp
}

// generateArchetypePresets creates the predefined archetype templates.
func (s *BuildTemplateSystem) generateArchetypePresets() {
	rng := rand.New(rand.NewSource(s.seed))

	// Tank preset: High Vitality and Endurance
	s.archetypePresets[BuildArchetypeTank] = s.createTankPreset(rng)

	// DPS preset: High Strength and Luck
	s.archetypePresets[BuildArchetypeDPS] = s.createDPSPreset(rng)

	// Support preset: High Intelligence and Vitality
	s.archetypePresets[BuildArchetypeSupport] = s.createSupportPreset(rng)

	// Hybrid preset: Balanced attributes
	s.archetypePresets[BuildArchetypeHybrid] = s.createHybridPreset(rng)

	// Battlemage preset: Strength and Intelligence
	s.archetypePresets[BuildArchetypeBattlemage] = s.createBattlemagePreset(rng)

	// Assassin preset: Agility and Luck
	s.archetypePresets[BuildArchetypeAssassin] = s.createAssassinPreset(rng)

	// Paladin preset: Vitality, Strength, and Intelligence
	s.archetypePresets[BuildArchetypePaladin] = s.createPaladinPreset(rng)

	log.WithFields(log.Fields{
		"system_name":  "build_template",
		"preset_count": len(s.archetypePresets),
	}).Debug("Generated archetype presets")
}

// createTankPreset generates the Tank archetype build.
func (s *BuildTemplateSystem) createTankPreset(rng *rand.Rand) *BuildTemplate {
	return &BuildTemplate{
		ID:             "preset_tank",
		Name:           "Iron Guardian",
		Description:    BuildArchetypeTank.Description(),
		Archetype:      BuildArchetypeTank,
		Class:          ClassWarrior,
		Specialization: SpecializationNone,
		Attributes: map[int]int{
			int(AttrStrength):     5,
			int(AttrAgility):      2,
			int(AttrIntelligence): 1,
			int(AttrVitality):     12,
			int(AttrEndurance):    10,
			int(AttrLuck):         0,
		},
		Talents: map[string]int{
			"defense_iron_skin":      3,
			"defense_thick_hide":     2,
			"defense_last_stand":     1,
			"utility_quick_recovery": 2,
		},
		Skills:        make(map[string]int),
		RequiredLevel: 10,
		CreatedAt:     time.Now().Unix(),
		UpdatedAt:     time.Now().Unix(),
		IsPreset:      true,
	}
}

// createDPSPreset generates the DPS archetype build.
func (s *BuildTemplateSystem) createDPSPreset(rng *rand.Rand) *BuildTemplate {
	return &BuildTemplate{
		ID:             "preset_dps",
		Name:           "Blade Fury",
		Description:    BuildArchetypeDPS.Description(),
		Archetype:      BuildArchetypeDPS,
		Class:          ClassWarrior,
		Specialization: SpecializationNone,
		Attributes: map[int]int{
			int(AttrStrength):     15,
			int(AttrAgility):      8,
			int(AttrIntelligence): 0,
			int(AttrVitality):     2,
			int(AttrEndurance):    0,
			int(AttrLuck):         5,
		},
		Talents: map[string]int{
			"offense_brutal_strikes": 3,
			"offense_critical_edge":  3,
			"offense_armor_pierce":   2,
		},
		Skills:        make(map[string]int),
		RequiredLevel: 10,
		CreatedAt:     time.Now().Unix(),
		UpdatedAt:     time.Now().Unix(),
		IsPreset:      true,
	}
}

// createSupportPreset generates the Support archetype build.
func (s *BuildTemplateSystem) createSupportPreset(rng *rand.Rand) *BuildTemplate {
	return &BuildTemplate{
		ID:             "preset_support",
		Name:           "Divine Healer",
		Description:    BuildArchetypeSupport.Description(),
		Archetype:      BuildArchetypeSupport,
		Class:          ClassMage,
		Specialization: SpecializationNone,
		Attributes: map[int]int{
			int(AttrStrength):     0,
			int(AttrAgility):      3,
			int(AttrIntelligence): 15,
			int(AttrVitality):     8,
			int(AttrEndurance):    2,
			int(AttrLuck):         2,
		},
		Talents: map[string]int{
			"utility_mana_flow":      3,
			"utility_quick_recovery": 2,
			"defense_magic_barrier":  2,
			"mastery_arcane_insight": 1,
		},
		Skills:        make(map[string]int),
		RequiredLevel: 10,
		CreatedAt:     time.Now().Unix(),
		UpdatedAt:     time.Now().Unix(),
		IsPreset:      true,
	}
}

// createHybridPreset generates the Hybrid archetype build.
func (s *BuildTemplateSystem) createHybridPreset(rng *rand.Rand) *BuildTemplate {
	return &BuildTemplate{
		ID:             "preset_hybrid",
		Name:           "Versatile Adventurer",
		Description:    BuildArchetypeHybrid.Description(),
		Archetype:      BuildArchetypeHybrid,
		Class:          ClassRanger,
		Specialization: SpecializationNone,
		Attributes: map[int]int{
			int(AttrStrength):     5,
			int(AttrAgility):      5,
			int(AttrIntelligence): 5,
			int(AttrVitality):     5,
			int(AttrEndurance):    5,
			int(AttrLuck):         5,
		},
		Talents: map[string]int{
			"offense_brutal_strikes": 2,
			"defense_thick_hide":     2,
			"utility_swift_feet":     2,
			"mastery_jack_of_trades": 2,
		},
		Skills:        make(map[string]int),
		RequiredLevel: 10,
		CreatedAt:     time.Now().Unix(),
		UpdatedAt:     time.Now().Unix(),
		IsPreset:      true,
	}
}

// createBattlemagePreset generates the Battlemage archetype build.
func (s *BuildTemplateSystem) createBattlemagePreset(rng *rand.Rand) *BuildTemplate {
	return &BuildTemplate{
		ID:             "preset_battlemage",
		Name:           "Spellblade",
		Description:    BuildArchetypeBattlemage.Description(),
		Archetype:      BuildArchetypeBattlemage,
		Class:          ClassMage,
		Specialization: SpecializationNone,
		Attributes: map[int]int{
			int(AttrStrength):     8,
			int(AttrAgility):      4,
			int(AttrIntelligence): 10,
			int(AttrVitality):     5,
			int(AttrEndurance):    3,
			int(AttrLuck):         0,
		},
		Talents: map[string]int{
			"offense_brutal_strikes": 2,
			"offense_spell_surge":    3,
			"mastery_arcane_blade":   2,
			"utility_mana_flow":      1,
		},
		Skills:        make(map[string]int),
		RequiredLevel: 10,
		CreatedAt:     time.Now().Unix(),
		UpdatedAt:     time.Now().Unix(),
		IsPreset:      true,
	}
}

// createAssassinPreset generates the Assassin archetype build.
func (s *BuildTemplateSystem) createAssassinPreset(rng *rand.Rand) *BuildTemplate {
	return &BuildTemplate{
		ID:             "preset_assassin",
		Name:           "Shadow Striker",
		Description:    BuildArchetypeAssassin.Description(),
		Archetype:      BuildArchetypeAssassin,
		Class:          ClassRogue,
		Specialization: SpecializationNone,
		Attributes: map[int]int{
			int(AttrStrength):     6,
			int(AttrAgility):      12,
			int(AttrIntelligence): 0,
			int(AttrVitality):     2,
			int(AttrEndurance):    2,
			int(AttrLuck):         8,
		},
		Talents: map[string]int{
			"offense_critical_edge": 3,
			"offense_backstab":      3,
			"utility_swift_feet":    2,
		},
		Skills:        make(map[string]int),
		RequiredLevel: 10,
		CreatedAt:     time.Now().Unix(),
		UpdatedAt:     time.Now().Unix(),
		IsPreset:      true,
	}
}

// createPaladinPreset generates the Paladin archetype build.
func (s *BuildTemplateSystem) createPaladinPreset(rng *rand.Rand) *BuildTemplate {
	return &BuildTemplate{
		ID:             "preset_paladin",
		Name:           "Holy Protector",
		Description:    BuildArchetypePaladin.Description(),
		Archetype:      BuildArchetypePaladin,
		Class:          ClassWarrior,
		Specialization: SpecializationNone,
		Attributes: map[int]int{
			int(AttrStrength):     8,
			int(AttrAgility):      2,
			int(AttrIntelligence): 6,
			int(AttrVitality):     8,
			int(AttrEndurance):    4,
			int(AttrLuck):         2,
		},
		Talents: map[string]int{
			"offense_brutal_strikes": 2,
			"defense_iron_skin":      2,
			"utility_divine_grace":   2,
			"mastery_holy_might":     2,
		},
		Skills:        make(map[string]int),
		RequiredLevel: 10,
		CreatedAt:     time.Now().Unix(),
		UpdatedAt:     time.Now().Unix(),
		IsPreset:      true,
	}
}

// SetOnTemplateApplied sets the callback for template application events.
func (s *BuildTemplateSystem) SetOnTemplateApplied(callback func(entity *Entity, template *BuildTemplate)) {
	s.onTemplateApplied = callback
}

// SetOnTemplateSaved sets the callback for template save events.
func (s *BuildTemplateSystem) SetOnTemplateSaved(callback func(entity *Entity, template *BuildTemplate)) {
	s.onTemplateSaved = callback
}

// GetTemplateStatus returns status info for UI display.
func (s *BuildTemplateSystem) GetTemplateStatus(entity *Entity) *BuildTemplateStatus {
	comp, hasComp := entity.GetComponent("build_template")
	if !hasComp {
		return nil
	}
	buildComp, ok := comp.(*BuildTemplateComponent)
	if !ok {
		return nil
	}

	currentTime := GetCurrentTime()
	return &BuildTemplateStatus{
		TemplateCount:     buildComp.GetTemplateCount(),
		MaxTemplates:      buildComp.MaxTemplates,
		AvailableSlots:    buildComp.GetAvailableSlots(),
		ActiveTemplateID:  buildComp.ActiveTemplateID,
		CanApply:          buildComp.CanApplyTemplate(currentTime),
		CooldownRemaining: buildComp.GetApplyCooldownRemaining(currentTime),
		TemplateNames:     buildComp.GetTemplateNames(),
	}
}

// BuildTemplateStatus contains UI-friendly template information.
type BuildTemplateStatus struct {
	TemplateCount     int
	MaxTemplates      int
	AvailableSlots    int
	ActiveTemplateID  string
	CanApply          bool
	CooldownRemaining float64
	TemplateNames     []string
}

// GetAvailableArchetypes returns a list of available archetype presets.
func (s *BuildTemplateSystem) GetAvailableArchetypes() []BuildArchetype {
	archetypes := make([]BuildArchetype, 0, len(s.archetypePresets))
	for arch := range s.archetypePresets {
		archetypes = append(archetypes, arch)
	}
	return archetypes
}
