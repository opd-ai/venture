// Package engine provides the gathering system for processing resource harvesting.
// This file implements the GatheringSystem which handles resource node interactions,
// yield calculation, experience gains, and respawn mechanics.
// Phase 95: Resource Gathering System

package engine

import (
	"math/rand"
	"sync"

	log "github.com/sirupsen/logrus"
)

// GatheringSystem processes resource gathering for all entities.
type GatheringSystem struct {
	mu    sync.RWMutex
	world *World

	// ResourceNodes maps entity IDs to resource node entities for quick lookup.
	resourceNodes map[uint64]*Entity

	// nextEntityID tracks the next ID for generated entities.
	nextEntityID uint64

	// BaseGatherTime is the default time to gather a resource in seconds.
	BaseGatherTime float64

	// XPPerHarvest is base XP gained per successful harvest.
	XPPerHarvest int

	// OnHarvestCallback is called when a harvest completes.
	OnHarvestCallback func(gatherer, node *Entity, resourceType ResourceType, yield int)

	// OnLevelUpCallback is called when gathering skill increases.
	OnLevelUpCallback func(entity *Entity, newLevel int)
}

// NewGatheringSystem creates a new gathering system.
func NewGatheringSystem(world *World) *GatheringSystem {
	log.WithFields(log.Fields{
		"system_name": "gathering",
	}).Debug("Creating gathering system")

	return &GatheringSystem{
		world:          world,
		resourceNodes:  make(map[uint64]*Entity),
		nextEntityID:   1000000, // High start to avoid collision
		BaseGatherTime: 3.0,     // 3 seconds default
		XPPerHarvest:   10,
	}
}

// Update processes all gathering entities and resource node respawns.
func (gs *GatheringSystem) Update(entities []*Entity, deltaTime float64) {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	// Update resource node index and respawns
	gs.updateResourceNodes(entities, deltaTime)

	// Process gatherers
	for _, entity := range entities {
		if !entity.HasComponent("gathering") {
			continue
		}

		comp, ok := entity.GetComponent("gathering")
		if !ok {
			continue
		}
		gatherComp, ok := comp.(*GatheringComponent)
		if !ok || gatherComp == nil {
			continue
		}

		if !gatherComp.IsCurrentlyGathering() {
			continue
		}

		gs.processGathering(entity, gatherComp, deltaTime)
	}
}

// updateResourceNodes updates the resource node index and processes respawns.
func (gs *GatheringSystem) updateResourceNodes(entities []*Entity, deltaTime float64) {
	// Clear and rebuild index
	gs.resourceNodes = make(map[uint64]*Entity)

	for _, entity := range entities {
		if !entity.HasComponent("resource_node") {
			continue
		}

		gs.resourceNodes[entity.ID] = entity

		comp, ok := entity.GetComponent("resource_node")
		if !ok {
			continue
		}
		nodeComp, ok := comp.(*ResourceNodeComponent)
		if !ok || nodeComp == nil {
			continue
		}

		// Process respawn
		if nodeComp.UpdateRespawn(deltaTime) {
			log.WithFields(log.Fields{
				"entity_id":     entity.ID,
				"resource_type": nodeComp.ResourceType,
			}).Debug("Resource node respawned")
		}
	}
}

// processGathering handles active gathering for an entity.
func (gs *GatheringSystem) processGathering(gatherer *Entity, gatherComp *GatheringComponent, deltaTime float64) {
	// Find target node
	targetNodeID := gatherComp.GetTargetNodeID()
	nodeEntity, ok := gs.resourceNodes[targetNodeID]
	if !ok || nodeEntity == nil {
		log.WithFields(log.Fields{
			"entity_id":   gatherer.ID,
			"target_node": targetNodeID,
		}).Debug("Target resource node not found, canceling gather")
		gatherComp.StopGathering()
		return
	}

	comp, ok := nodeEntity.GetComponent("resource_node")
	if !ok {
		gatherComp.StopGathering()
		return
	}
	nodeComp, ok := comp.(*ResourceNodeComponent)
	if !ok || nodeComp == nil {
		gatherComp.StopGathering()
		return
	}

	// Check if still harvestable
	if !nodeComp.CanHarvest(gatherComp.GetSkillLevel()) {
		log.WithFields(log.Fields{
			"entity_id":   gatherer.ID,
			"target_node": targetNodeID,
			"reason":      "node depleted or skill too low",
		}).Debug("Cannot harvest, canceling gather")
		gatherComp.StopGathering()
		return
	}

	// Check tool requirement
	if !gatherComp.HasCorrectTool(nodeComp.RequiredTool) {
		log.WithFields(log.Fields{
			"entity_id":     gatherer.ID,
			"required_tool": nodeComp.RequiredTool,
			"equipped_tool": gatherComp.GetEquippedTool(),
		}).Debug("Wrong tool equipped, canceling gather")
		gatherComp.StopGathering()
		return
	}

	// Update progress
	if gatherComp.UpdateProgress(deltaTime, gs.BaseGatherTime) {
		// Gathering complete
		gs.completeHarvest(gatherer, gatherComp, nodeEntity, nodeComp)
	}
}

// completeHarvest finishes a harvest and grants rewards.
func (gs *GatheringSystem) completeHarvest(gatherer *Entity, gatherComp *GatheringComponent, nodeEntity *Entity, nodeComp *ResourceNodeComponent) {
	// Attempt harvest
	if !nodeComp.Harvest() {
		gatherComp.StopGathering()
		return
	}

	// Calculate yield
	yield := gs.calculateYield(gatherComp, nodeComp)

	// Grant XP
	xpGained := gs.XPPerHarvest + (nodeComp.MinSkillLevel / 2)
	if gatherComp.AddXP(xpGained) {
		newLevel := gatherComp.GetSkillLevel()
		log.WithFields(log.Fields{
			"entity_id": gatherer.ID,
			"new_level": newLevel,
		}).Info("Gathering skill leveled up")

		if gs.OnLevelUpCallback != nil {
			gs.OnLevelUpCallback(gatherer, newLevel)
		}
	}

	// Complete gathering
	gatherComp.CompleteGathering(nodeComp.ResourceType)

	log.WithFields(log.Fields{
		"entity_id":     gatherer.ID,
		"resource_type": nodeComp.ResourceType,
		"yield":         yield,
		"xp_gained":     xpGained,
	}).Debug("Harvest complete")

	// Trigger callback
	if gs.OnHarvestCallback != nil {
		gs.OnHarvestCallback(gatherer, nodeEntity, nodeComp.ResourceType, yield)
	}
}

// calculateYield determines harvest yield based on skill and tool bonus.
func (gs *GatheringSystem) calculateYield(gatherComp *GatheringComponent, nodeComp *ResourceNodeComponent) int {
	baseYield := nodeComp.YieldMin
	yieldRange := nodeComp.YieldMax - nodeComp.YieldMin

	if yieldRange > 0 {
		// Use skill to influence yield within range
		skillBonus := float64(gatherComp.GetSkillLevel()) / 100.0
		baseYield += int(float64(yieldRange) * skillBonus)
	}

	// Apply tool bonus
	toolBonus := gatherComp.GetToolBonus(nodeComp.RequiredTool)
	yield := int(float64(baseYield) * toolBonus)

	if yield < 1 {
		yield = 1
	}

	return yield
}

// StartGathering initiates gathering from a resource node.
func (gs *GatheringSystem) StartGathering(gatherer *Entity, nodeID uint64) bool {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	comp, ok := gatherer.GetComponent("gathering")
	if !ok {
		return false
	}
	gatherComp, ok := comp.(*GatheringComponent)
	if !ok || gatherComp == nil {
		return false
	}

	nodeEntity, ok := gs.resourceNodes[nodeID]
	if !ok || nodeEntity == nil {
		return false
	}

	nodeCompRaw, ok := nodeEntity.GetComponent("resource_node")
	if !ok {
		return false
	}
	nodeComp, ok := nodeCompRaw.(*ResourceNodeComponent)
	if !ok || nodeComp == nil {
		return false
	}

	// Validate can harvest
	if !nodeComp.CanHarvest(gatherComp.GetSkillLevel()) {
		return false
	}

	// Validate tool
	if !gatherComp.HasCorrectTool(nodeComp.RequiredTool) {
		return false
	}

	gatherComp.StartGathering(nodeID)
	log.WithFields(log.Fields{
		"entity_id":     gatherer.ID,
		"target_node":   nodeID,
		"resource_type": nodeComp.ResourceType,
	}).Debug("Started gathering")

	return true
}

// CancelGathering stops an active gathering action.
func (gs *GatheringSystem) CancelGathering(gatherer *Entity) {
	comp, ok := gatherer.GetComponent("gathering")
	if !ok {
		return
	}
	gatherComp, ok := comp.(*GatheringComponent)
	if !ok || gatherComp == nil {
		return
	}

	if gatherComp.IsCurrentlyGathering() {
		gatherComp.StopGathering()
		log.WithFields(log.Fields{
			"entity_id": gatherer.ID,
		}).Debug("Gathering canceled")
	}
}

// GetNearbyResourceNodes returns resource nodes within range of a position.
func (gs *GatheringSystem) GetNearbyResourceNodes(x, y, maxRange float64) []*Entity {
	gs.mu.RLock()
	defer gs.mu.RUnlock()

	var nearby []*Entity
	rangeSquared := maxRange * maxRange

	for _, entity := range gs.resourceNodes {
		comp, ok := entity.GetComponent("position")
		if !ok {
			continue
		}
		posComp, ok := comp.(*PositionComponent)
		if !ok || posComp == nil {
			continue
		}

		dx := posComp.X - x
		dy := posComp.Y - y
		distSquared := dx*dx + dy*dy

		if distSquared <= rangeSquared {
			nearby = append(nearby, entity)
		}
	}

	return nearby
}

// GenerateResourceNode creates a resource node entity with procedural properties.
func (gs *GatheringSystem) GenerateResourceNode(seed int64, resourceType ResourceType, biome string, x, y float64) *Entity {
	rng := rand.New(rand.NewSource(seed))

	// Get next entity ID
	gs.mu.Lock()
	entityID := gs.nextEntityID
	gs.nextEntityID++
	gs.mu.Unlock()

	entity := NewEntity(entityID)

	// Position
	entity.AddComponent(&PositionComponent{X: x, Y: y})

	// Resource node
	node := NewResourceNodeComponent(resourceType, biome)

	// Randomize properties based on seed
	node.MaxQuantity = 2 + rng.Intn(4) // 2-5 harvests
	node.Quantity = node.MaxQuantity
	node.RespawnTime = 180.0 + float64(rng.Intn(240)) // 3-7 minutes
	node.MinSkillLevel = 1 + rng.Intn(20)             // 1-20 skill requirement
	node.YieldMin = 1
	node.YieldMax = 2 + rng.Intn(3) // 2-4 max yield

	// Higher skill nodes have better yields
	if node.MinSkillLevel > 10 {
		node.YieldMax += 1
	}
	if node.MinSkillLevel > 15 {
		node.YieldMax += 1
	}

	entity.AddComponent(node)

	log.WithFields(log.Fields{
		"entity_id":     entity.ID,
		"resource_type": resourceType,
		"biome":         biome,
		"seed":          seed,
		"quantity":      node.Quantity,
		"skill_req":     node.MinSkillLevel,
	}).Debug("Generated resource node")

	return entity
}

// GetResourceNodeCount returns the count of tracked resource nodes.
func (gs *GatheringSystem) GetResourceNodeCount() int {
	gs.mu.RLock()
	defer gs.mu.RUnlock()
	return len(gs.resourceNodes)
}
