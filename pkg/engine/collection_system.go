// Package engine provides the collection system for the ECS.
// This file implements the CollectionSystem that processes collectible discovery,
// tracks progress, and grants completion rewards.
// Phase 97: Collection System (V18.0)

package engine

import (
	"sync"

	log "github.com/sirupsen/logrus"
)

// CollectionRewardCallback is called when a milestone reward is unlocked.
type CollectionRewardCallback func(entityID uint64, milestone CollectionMilestone, category CollectionCategory)

// CollectionDiscoveryCallback is called when a new collectible is discovered.
type CollectionDiscoveryCallback func(entityID uint64, entry *CollectedEntry)

// CollectionSystem manages collectible discovery and completion tracking.
type CollectionSystem struct {
	mu sync.RWMutex

	// world is a reference to the ECS world
	world *World

	// rewardCallbacks are invoked when milestones are reached
	rewardCallbacks []CollectionRewardCallback

	// discoveryCallbacks are invoked when new collectibles are found
	discoveryCallbacks []CollectionDiscoveryCallback

	// milestones defines the completion thresholds
	milestones []CollectionMilestone

	// autoRegisterFromFishing enables automatic fish collection tracking
	autoRegisterFromFishing bool

	// autoRegisterFromGathering enables automatic resource collection tracking
	autoRegisterFromGathering bool
}

// NewCollectionSystem creates a new collection system.
func NewCollectionSystem(world *World) *CollectionSystem {
	log.WithFields(log.Fields{
		"system_name": "collection",
	}).Debug("Creating collection system")

	return &CollectionSystem{
		world:                     world,
		rewardCallbacks:           make([]CollectionRewardCallback, 0),
		discoveryCallbacks:        make([]CollectionDiscoveryCallback, 0),
		milestones:                DefaultMilestones(),
		autoRegisterFromFishing:   true,
		autoRegisterFromGathering: true,
	}
}

// Update processes all entities with collection-related components.
func (s *CollectionSystem) Update(entities []*Entity, deltaTime float64) {
	for _, entity := range entities {
		// Process collectible pickups
		if entity.HasComponent("collectible") {
			s.processCollectibleEntity(entity)
		}

		// Auto-register from fishing catches
		if s.autoRegisterFromFishing && entity.HasComponent("fishing") && entity.HasComponent("collection") {
			s.syncFishingCollection(entity)
		}

		// Auto-register from gathering harvests
		if s.autoRegisterFromGathering && entity.HasComponent("gathering") && entity.HasComponent("collection") {
			s.syncGatheringCollection(entity)
		}
	}
}

// processCollectibleEntity checks if a collectible should be collected.
func (s *CollectionSystem) processCollectibleEntity(entity *Entity) {
	comp, ok := entity.GetComponent("collectible")
	if !ok {
		return
	}
	collectible, ok := comp.(*CollectibleComponent)
	if !ok || collectible == nil {
		return
	}

	// Skip already collected items
	if collectible.IsAlreadyCollected() {
		return
	}

	// Check for nearby players that can collect this
	if s.world == nil {
		return
	}

	// Get position of collectible
	posComp := entity.GetPosition()
	if posComp == nil {
		return
	}

	collX, collY := posComp.X, posComp.Y

	// Find players within collection range
	for _, playerEntity := range s.world.GetEntitiesWith("player") {
		playerPos := playerEntity.GetPosition()
		if playerPos == nil {
			continue
		}

		playerX, playerY := playerPos.X, playerPos.Y

		// Check distance (collection range of 32 pixels)
		dx := playerX - collX
		dy := playerY - collY
		distSq := dx*dx + dy*dy

		if distSq <= 32*32 {
			s.tryCollectItem(playerEntity, entity, collectible)
			return
		}
	}
}

// tryCollectItem attempts to collect an item for a player.
func (s *CollectionSystem) tryCollectItem(player, collectibleEntity *Entity, collectible *CollectibleComponent) {
	// Get player's collection component
	comp, ok := player.GetComponent("collection")
	if !ok {
		return
	}
	collection, ok := comp.(*CollectionComponent)
	if !ok || collection == nil {
		return
	}

	// Check level requirement
	playerLevel := 1
	if exp := player.GetExperience(); exp != nil {
		playerLevel = exp.Level
	}

	if !collectible.CanCollect(playerLevel) {
		return
	}

	// Get current timestamp
	timestamp := s.getCurrentTimestamp()

	// Collect the item
	if !collectible.Collect(player.ID, timestamp) {
		return
	}

	// Add to player's collection
	isNew := collection.AddCollectible(
		collectible.GetCollectibleID(),
		collectible.GetName(),
		collectible.GetCategory(),
		collectible.GetRarity(),
		collectible.GetDescription(),
		timestamp,
	)

	if isNew {
		log.WithFields(log.Fields{
			"entityID":      player.ID,
			"collectibleID": collectible.GetCollectibleID(),
			"category":      collectible.GetCategory(),
			"rarity":        collectible.GetRarity(),
		}).Debug("New collectible discovered")

		// Notify discovery callbacks
		entry := collection.GetCollectible(collectible.GetCollectibleID())
		s.notifyDiscovery(player.ID, entry)

		// Check milestones
		s.checkMilestones(player.ID, collection, collectible.GetCategory())
	}
}

// syncFishingCollection syncs caught fish to the collection.
func (s *CollectionSystem) syncFishingCollection(entity *Entity) {
	comp, ok := entity.GetComponent("fishing")
	if !ok {
		return
	}
	fishing, ok := comp.(*FishingComponent)
	if !ok || fishing == nil {
		return
	}

	comp, ok = entity.GetComponent("collection")
	if !ok {
		return
	}
	collection, ok := comp.(*CollectionComponent)
	if !ok || collection == nil {
		return
	}

	// Check each caught fish type
	for fishTypeID, count := range fishing.TotalCaught {
		if count > 0 && !collection.HasCollectible(fishTypeID) {
			// Determine rarity based on fish type
			rarity := s.getFishRarity(fishTypeID)
			timestamp := s.getCurrentTimestamp()

			isNew := collection.AddCollectible(
				fishTypeID,
				fishTypeID, // Use ID as name if no lookup available
				CollectionCategoryFish,
				rarity,
				"A fish caught while fishing",
				timestamp,
			)

			if isNew {
				entry := collection.GetCollectible(fishTypeID)
				s.notifyDiscovery(entity.ID, entry)
				s.checkMilestones(entity.ID, collection, CollectionCategoryFish)
			}
		}
	}
}

// syncGatheringCollection syncs harvested resources to the collection.
func (s *CollectionSystem) syncGatheringCollection(entity *Entity) {
	gathering := s.getGatheringComponent(entity)
	if gathering == nil {
		return
	}

	collection := s.getCollectionComponent(entity)
	if collection == nil {
		return
	}

	s.processHarvestedResources(entity, gathering, collection)
}

// getGatheringComponent retrieves and type-asserts the gathering component.
func (s *CollectionSystem) getGatheringComponent(entity *Entity) *GatheringComponent {
	comp, ok := entity.GetComponent("gathering")
	if !ok {
		return nil
	}
	gathering, ok := comp.(*GatheringComponent)
	if !ok || gathering == nil {
		return nil
	}
	return gathering
}

// getCollectionComponent retrieves and type-asserts the collection component.
func (s *CollectionSystem) getCollectionComponent(entity *Entity) *CollectionComponent {
	comp, ok := entity.GetComponent("collection")
	if !ok {
		return nil
	}
	collection, ok := comp.(*CollectionComponent)
	if !ok || collection == nil {
		return nil
	}
	return collection
}

// processHarvestedResources adds harvested resources to the collection.
func (s *CollectionSystem) processHarvestedResources(entity *Entity, gathering *GatheringComponent, collection *CollectionComponent) {
	for resourceType, count := range gathering.TotalHarvested {
		if count > 0 {
			resourceID := string(resourceType)
			if !collection.HasCollectible(resourceID) {
				s.addNewResource(entity, collection, resourceType, resourceID)
			}
		}
	}
}

// addNewResource adds a new resource to the collection and notifies on discovery.
func (s *CollectionSystem) addNewResource(entity *Entity, collection *CollectionComponent, resourceType ResourceType, resourceID string) {
	rarity := s.getResourceRarity(resourceType)
	timestamp := s.getCurrentTimestamp()

	isNew := collection.AddCollectible(
		resourceID,
		string(resourceType),
		CollectionCategoryResources,
		rarity,
		"A resource harvested while gathering",
		timestamp,
	)

	if isNew {
		entry := collection.GetCollectible(resourceID)
		s.notifyDiscovery(entity.ID, entry)
		s.checkMilestones(entity.ID, collection, CollectionCategoryResources)
	}
}

// getFishRarity returns the collection rarity for a fish type.
func (s *CollectionSystem) getFishRarity(fishTypeID string) CollectionRarity {
	// Map fish types to rarities based on naming convention
	// This could be expanded with a proper lookup table
	switch {
	case len(fishTypeID) > 0 && fishTypeID[len(fishTypeID)-1] == '1':
		return CollectionRarityCommon
	case len(fishTypeID) > 0 && fishTypeID[len(fishTypeID)-1] == '2':
		return CollectionRarityUncommon
	case len(fishTypeID) > 0 && fishTypeID[len(fishTypeID)-1] == '3':
		return CollectionRarityRare
	default:
		return CollectionRarityCommon
	}
}

// getResourceRarity returns the collection rarity for a resource type.
func (s *CollectionSystem) getResourceRarity(resourceType ResourceType) CollectionRarity {
	switch resourceType {
	case ResourceTypeOre, ResourceTypeWood, ResourceTypeFiber:
		return CollectionRarityCommon
	case ResourceTypeHerb:
		return CollectionRarityUncommon
	case ResourceTypeGem:
		return CollectionRarityRare
	case ResourceTypeEssence:
		return CollectionRarityEpic
	default:
		return CollectionRarityCommon
	}
}

// checkMilestones checks if any milestones are reached for a category.
func (s *CollectionSystem) checkMilestones(entityID uint64, collection *CollectionComponent, category CollectionCategory) {
	for _, m := range s.milestones {
		milestone := collection.CheckMilestone(category, m.Threshold)
		if milestone != nil {
			log.WithFields(log.Fields{
				"entityID":  entityID,
				"category":  category,
				"milestone": milestone.Name,
				"threshold": milestone.Threshold,
			}).Info("Collection milestone reached")

			// Add bonus points
			if milestone.Points > 0 {
				// Points are added internally via CheckMilestone
			}

			// Notify reward callbacks
			s.notifyMilestoneReward(entityID, *milestone, category)
		}
	}
}

// RegisterRewardCallback adds a callback for milestone rewards.
func (s *CollectionSystem) RegisterRewardCallback(callback CollectionRewardCallback) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.rewardCallbacks = append(s.rewardCallbacks, callback)
}

// RegisterDiscoveryCallback adds a callback for new discoveries.
func (s *CollectionSystem) RegisterDiscoveryCallback(callback CollectionDiscoveryCallback) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.discoveryCallbacks = append(s.discoveryCallbacks, callback)
}

// notifyMilestoneReward notifies all registered reward callbacks.
func (s *CollectionSystem) notifyMilestoneReward(entityID uint64, milestone CollectionMilestone, category CollectionCategory) {
	s.mu.RLock()
	callbacks := make([]CollectionRewardCallback, len(s.rewardCallbacks))
	copy(callbacks, s.rewardCallbacks)
	s.mu.RUnlock()

	for _, callback := range callbacks {
		callback(entityID, milestone, category)
	}
}

// notifyDiscovery notifies all registered discovery callbacks.
func (s *CollectionSystem) notifyDiscovery(entityID uint64, entry *CollectedEntry) {
	s.mu.RLock()
	callbacks := make([]CollectionDiscoveryCallback, len(s.discoveryCallbacks))
	copy(callbacks, s.discoveryCallbacks)
	s.mu.RUnlock()

	for _, callback := range callbacks {
		callback(entityID, entry)
	}
}

// getCurrentTimestamp returns the current game time or real time.
func (s *CollectionSystem) getCurrentTimestamp() int64 {
	// If world has a clock, use it
	if s.world != nil && s.world.Clock != nil {
		return s.world.Clock.Now().Unix()
	}
	// Fallback to a default
	return 0
}

// SetAutoRegisterFishing enables/disables automatic fish collection tracking.
func (s *CollectionSystem) SetAutoRegisterFishing(enabled bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.autoRegisterFromFishing = enabled
}

// SetAutoRegisterGathering enables/disables automatic resource collection tracking.
func (s *CollectionSystem) SetAutoRegisterGathering(enabled bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.autoRegisterFromGathering = enabled
}

// SetMilestones sets custom milestones for the collection system.
func (s *CollectionSystem) SetMilestones(milestones []CollectionMilestone) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.milestones = milestones
}

// GetMilestones returns the current milestone definitions.
func (s *CollectionSystem) GetMilestones() []CollectionMilestone {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]CollectionMilestone, len(s.milestones))
	copy(result, s.milestones)
	return result
}

// RegisterCollectible manually registers a collectible discovery for an entity.
func (s *CollectionSystem) RegisterCollectible(entityID uint64, id, name string, category CollectionCategory, rarity CollectionRarity, description string) bool {
	if s.world == nil {
		return false
	}

	entity, ok := s.world.GetEntity(entityID)
	if !ok || entity == nil {
		return false
	}

	comp, ok := entity.GetComponent("collection")
	if !ok {
		return false
	}
	collection, ok := comp.(*CollectionComponent)
	if !ok || collection == nil {
		return false
	}

	timestamp := s.getCurrentTimestamp()

	isNew := collection.AddCollectible(id, name, category, rarity, description, timestamp)
	if isNew {
		entry := collection.GetCollectible(id)
		s.notifyDiscovery(entityID, entry)
		s.checkMilestones(entityID, collection, category)
	}

	return isNew
}

// GetCollectionProgress returns collection progress for an entity.
func (s *CollectionSystem) GetCollectionProgress(entityID uint64) (discovered, total int, percent float64) {
	if s.world == nil {
		return 0, 0, 0
	}

	entity, ok := s.world.GetEntity(entityID)
	if !ok || entity == nil {
		return 0, 0, 0
	}

	comp, ok := entity.GetComponent("collection")
	if !ok {
		return 0, 0, 0
	}
	collection, ok := comp.(*CollectionComponent)
	if !ok || collection == nil {
		return 0, 0, 0
	}

	discovered = collection.GetDiscoveredCount()
	for _, cat := range AllCollectionCategories() {
		_, catTotal := collection.GetCategoryProgress(cat)
		total += catTotal
	}
	percent = collection.GetOverallCompletionPercent()

	return discovered, total, percent
}

// GetCategoryProgress returns progress for a specific category.
func (s *CollectionSystem) GetCategoryProgress(entityID uint64, category CollectionCategory) (discovered, total int, percent float64) {
	if s.world == nil {
		return 0, 0, 0
	}

	entity, ok := s.world.GetEntity(entityID)
	if !ok || entity == nil {
		return 0, 0, 0
	}

	comp, ok := entity.GetComponent("collection")
	if !ok {
		return 0, 0, 0
	}
	collection, ok := comp.(*CollectionComponent)
	if !ok || collection == nil {
		return 0, 0, 0
	}

	discovered, total = collection.GetCategoryProgress(category)
	percent = collection.GetCategoryCompletionPercent(category)

	return discovered, total, percent
}

// ExportCollectionData exports a player's collection as a serializable format.
func (s *CollectionSystem) ExportCollectionData(entityID uint64) ([]byte, error) {
	if s.world == nil {
		return nil, nil
	}

	entity, ok := s.world.GetEntity(entityID)
	if !ok || entity == nil {
		return nil, nil
	}

	comp, ok := entity.GetComponent("collection")
	if !ok {
		return nil, nil
	}
	collection, ok := comp.(*CollectionComponent)
	if !ok || collection == nil {
		return nil, nil
	}

	return collection.Serialize()
}

// ImportCollectionData imports collection data to a player.
func (s *CollectionSystem) ImportCollectionData(entityID uint64, data []byte) error {
	if s.world == nil {
		return nil
	}

	entity, ok := s.world.GetEntity(entityID)
	if !ok || entity == nil {
		return nil
	}

	comp, ok := entity.GetComponent("collection")
	if !ok {
		// Create a new collection component if one doesn't exist
		collection := NewCollectionComponent()
		entity.AddComponent(collection)
		return collection.Deserialize(data)
	}
	collection, ok := comp.(*CollectionComponent)
	if !ok || collection == nil {
		collection = NewCollectionComponent()
		entity.AddComponent(collection)
		return collection.Deserialize(data)
	}

	return collection.Deserialize(data)
}
