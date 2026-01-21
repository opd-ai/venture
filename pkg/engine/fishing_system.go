// Package engine provides the fishing system for processing fishing activities.
// This file implements the FishingSystem which handles fishing spot interactions,
// fish generation, catch calculation, and the fishing minigame mechanics.
// Phase 96: Fishing System

package engine

import (
	"math/rand"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

// FishingSystem processes fishing activities for all entities.
type FishingSystem struct {
	mu    sync.RWMutex
	world *World

	// fishingSpots maps entity IDs to fishing spot entities.
	fishingSpots map[uint64]*Entity

	// fishTypes is the registry of all fish types.
	fishTypes map[string]*FishType

	// nextEntityID tracks the next ID for generated entities.
	nextEntityID uint64

	// BaseWaitTime is the base time to wait for a bite (seconds).
	BaseWaitTime float64

	// BiteWindowTime is the time window to hook a fish (seconds).
	BiteWindowTime float64

	// XPPerCatch is base XP gained per successful catch.
	XPPerCatch int

	// SpotCooldown is cooldown time per spot after a catch.
	SpotCooldown float64

	// OnCatchCallback is called when a fish is caught.
	OnCatchCallback func(fisher *Entity, caught *CaughtFish)

	// OnLevelUpCallback is called when fishing skill increases.
	OnLevelUpCallback func(entity *Entity, newLevel int)

	// OnBiteCallback is called when a fish bites.
	OnBiteCallback func(fisher *Entity, fishTypeID string)

	// OnEscapeCallback is called when a fish escapes.
	OnEscapeCallback func(fisher *Entity, fishTypeID, reason string)

	// CurrentTimeOfDay is a function returning the current time of day.
	CurrentTimeOfDay func() TimeOfDay

	// CurrentWeather is a function returning the current weather.
	CurrentWeather func() string
}

// NewFishingSystem creates a new fishing system.
func NewFishingSystem(world *World) *FishingSystem {
	log.WithFields(log.Fields{
		"system_name": "fishing",
	}).Debug("Creating fishing system")

	fs := &FishingSystem{
		world:          world,
		fishingSpots:   make(map[uint64]*Entity),
		fishTypes:      make(map[string]*FishType),
		nextEntityID:   2000000, // High start to avoid collision
		BaseWaitTime:   10.0,    // 10 seconds default
		BiteWindowTime: 2.0,     // 2 second window
		XPPerCatch:     15,
		SpotCooldown:   5.0,
		CurrentTimeOfDay: func() TimeOfDay {
			hour := time.Now().Hour()
			if hour >= 5 && hour < 8 {
				return TimeDawn
			} else if hour >= 8 && hour < 18 {
				return TimeDay
			} else if hour >= 18 && hour < 21 {
				return TimeDusk
			}
			return TimeNight
		},
		CurrentWeather: func() string {
			return "clear"
		},
	}

	// Register default fish types
	fs.registerDefaultFishTypes()

	return fs
}

// registerDefaultFishTypes adds the base fish types to the registry.
func (fs *FishingSystem) registerDefaultFishTypes() {
	// Freshwater common
	fs.RegisterFishType(&FishType{
		ID:         "bass",
		Name:       "Bass",
		Rarity:     FishRarityCommon,
		WaterTypes: []WaterType{WaterTypeFreshwater},
		MinDepth:   DepthShallow,
		BestTime:   TimeDay,
		MinSkill:   1,
		BaseWeight: 0.5,
		MaxWeight:  3.0,
		Difficulty: 0.3,
	})
	fs.RegisterFishType(&FishType{
		ID:         "trout",
		Name:       "Trout",
		Rarity:     FishRarityCommon,
		WaterTypes: []WaterType{WaterTypeFreshwater},
		MinDepth:   DepthShallow,
		BestTime:   TimeDawn,
		MinSkill:   1,
		BaseWeight: 0.3,
		MaxWeight:  2.0,
		Difficulty: 0.2,
	})
	fs.RegisterFishType(&FishType{
		ID:         "catfish",
		Name:       "Catfish",
		Rarity:     FishRarityUncommon,
		WaterTypes: []WaterType{WaterTypeFreshwater},
		MinDepth:   DepthMedium,
		BestTime:   TimeNight,
		MinSkill:   5,
		BaseWeight: 1.0,
		MaxWeight:  8.0,
		Difficulty: 0.4,
	})
	fs.RegisterFishType(&FishType{
		ID:         "pike",
		Name:       "Pike",
		Rarity:     FishRarityRare,
		WaterTypes: []WaterType{WaterTypeFreshwater},
		MinDepth:   DepthMedium,
		BestTime:   TimeDusk,
		MinSkill:   15,
		BaseWeight: 2.0,
		MaxWeight:  15.0,
		Difficulty: 0.6,
	})

	// Saltwater common
	fs.RegisterFishType(&FishType{
		ID:         "mackerel",
		Name:       "Mackerel",
		Rarity:     FishRarityCommon,
		WaterTypes: []WaterType{WaterTypeSaltwater},
		MinDepth:   DepthShallow,
		BestTime:   TimeDay,
		MinSkill:   1,
		BaseWeight: 0.3,
		MaxWeight:  1.5,
		Difficulty: 0.2,
	})
	fs.RegisterFishType(&FishType{
		ID:         "cod",
		Name:       "Cod",
		Rarity:     FishRarityCommon,
		WaterTypes: []WaterType{WaterTypeSaltwater},
		MinDepth:   DepthMedium,
		BestTime:   TimeAny,
		MinSkill:   3,
		BaseWeight: 1.0,
		MaxWeight:  5.0,
		Difficulty: 0.3,
	})
	fs.RegisterFishType(&FishType{
		ID:         "tuna",
		Name:       "Tuna",
		Rarity:     FishRarityUncommon,
		WaterTypes: []WaterType{WaterTypeSaltwater},
		MinDepth:   DepthDeep,
		BestTime:   TimeDay,
		MinSkill:   10,
		BaseWeight: 5.0,
		MaxWeight:  50.0,
		Difficulty: 0.6,
	})
	fs.RegisterFishType(&FishType{
		ID:         "swordfish",
		Name:       "Swordfish",
		Rarity:     FishRarityRare,
		WaterTypes: []WaterType{WaterTypeSaltwater},
		MinDepth:   DepthDeep,
		BestTime:   TimeDusk,
		MinSkill:   20,
		BaseWeight: 20.0,
		MaxWeight:  150.0,
		Difficulty: 0.8,
	})

	// Magical fish
	fs.RegisterFishType(&FishType{
		ID:         "starfish",
		Name:       "Glowing Starfish",
		Rarity:     FishRarityUncommon,
		WaterTypes: []WaterType{WaterTypeMagical},
		MinDepth:   DepthShallow,
		BestTime:   TimeNight,
		MinSkill:   10,
		BaseWeight: 0.1,
		MaxWeight:  0.5,
		Difficulty: 0.4,
	})
	fs.RegisterFishType(&FishType{
		ID:         "moonfish",
		Name:       "Moonfish",
		Rarity:     FishRarityRare,
		WaterTypes: []WaterType{WaterTypeMagical},
		MinDepth:   DepthMedium,
		BestTime:   TimeNight,
		MinSkill:   25,
		BaseWeight: 1.0,
		MaxWeight:  5.0,
		Difficulty: 0.7,
	})
	fs.RegisterFishType(&FishType{
		ID:         "ethereal_eel",
		Name:       "Ethereal Eel",
		Rarity:     FishRarityEpic,
		WaterTypes: []WaterType{WaterTypeMagical},
		MinDepth:   DepthDeep,
		BestTime:   TimeAny,
		MinSkill:   40,
		BaseWeight: 2.0,
		MaxWeight:  10.0,
		Difficulty: 0.85,
	})

	// Legendary fish (all water types)
	fs.RegisterFishType(&FishType{
		ID:         "golden_carp",
		Name:       "Golden Carp",
		Rarity:     FishRarityLegendary,
		WaterTypes: []WaterType{WaterTypeFreshwater, WaterTypeMagical},
		MinDepth:   DepthDeep,
		BestTime:   TimeDawn,
		MinSkill:   50,
		BaseWeight: 5.0,
		MaxWeight:  25.0,
		Difficulty: 0.9,
	})
	fs.RegisterFishType(&FishType{
		ID:         "leviathan",
		Name:       "Leviathan",
		Rarity:     FishRarityLegendary,
		WaterTypes: []WaterType{WaterTypeSaltwater},
		MinDepth:   DepthDeep,
		BestTime:   TimeNight,
		MinSkill:   75,
		BaseWeight: 100.0,
		MaxWeight:  500.0,
		Difficulty: 0.95,
	})

	log.WithFields(log.Fields{
		"fish_count": len(fs.fishTypes),
	}).Debug("Registered default fish types")
}

// RegisterFishType adds a fish type to the registry.
func (fs *FishingSystem) RegisterFishType(fishType *FishType) {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	fs.fishTypes[fishType.ID] = fishType
}

// GetFishType returns a fish type by ID.
func (fs *FishingSystem) GetFishType(id string) *FishType {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	return fs.fishTypes[id]
}

// GetAllFishTypes returns all registered fish types.
func (fs *FishingSystem) GetAllFishTypes() []*FishType {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	types := make([]*FishType, 0, len(fs.fishTypes))
	for _, ft := range fs.fishTypes {
		types = append(types, ft)
	}
	return types
}

// Update processes all fishing entities and fishing spot cooldowns.
func (fs *FishingSystem) Update(entities []*Entity, deltaTime float64) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	// Update fishing spot index and cooldowns
	fs.updateFishingSpots(entities, deltaTime)

	// Process fishers
	for _, entity := range entities {
		if !entity.HasComponent("fishing") {
			continue
		}

		comp, ok := entity.GetComponent("fishing")
		if !ok {
			continue
		}
		fishComp, ok := comp.(*FishingComponent)
		if !ok || fishComp == nil {
			continue
		}

		state := fishComp.GetState()
		if state == FishingStateIdle {
			continue
		}

		fs.processFishing(entity, fishComp, deltaTime)
	}
}

// updateFishingSpots updates the fishing spot index and processes cooldowns.
func (fs *FishingSystem) updateFishingSpots(entities []*Entity, deltaTime float64) {
	// Clear and rebuild index
	fs.fishingSpots = make(map[uint64]*Entity)

	for _, entity := range entities {
		if !entity.HasComponent("fishing_spot") {
			continue
		}

		fs.fishingSpots[entity.ID] = entity

		comp, ok := entity.GetComponent("fishing_spot")
		if !ok {
			continue
		}
		spotComp, ok := comp.(*FishingSpotComponent)
		if !ok || spotComp == nil {
			continue
		}

		// Process cooldown
		fs.SpotUpdateCooldown(spotComp, deltaTime)
	}
}

// processFishing handles active fishing for an entity.
func (fs *FishingSystem) processFishing(fisher *Entity, fishComp *FishingComponent, deltaTime float64) {
	state := fishComp.GetState()

	switch state {
	case FishingStateCasting:
		// Casting is instant for now, handled by CompleteCast
		return

	case FishingStateWaiting:
		fs.processWaiting(fisher, fishComp, deltaTime)

	case FishingStateBite:
		fs.processBite(fisher, fishComp, deltaTime)

	case FishingStateReeling:
		fs.processReeling(fisher, fishComp, deltaTime)

	case FishingStateCaught, FishingStateEscaped:
		// Terminal states - need external reset
		return
	}
}

// processWaiting handles the waiting for bite phase.
func (fs *FishingSystem) processWaiting(fisher *Entity, fishComp *FishingComponent, deltaTime float64) {
	if !fishComp.UpdateWait(deltaTime) {
		return
	}

	// Time for a bite - select a fish
	spotEntity, ok := fs.fishingSpots[fishComp.GetTargetSpotID()]
	if !ok || spotEntity == nil {
		fishComp.StopFishing()
		return
	}

	comp, ok := spotEntity.GetComponent("fishing_spot")
	if !ok {
		fishComp.StopFishing()
		return
	}
	spotComp, ok := comp.(*FishingSpotComponent)
	if !ok || spotComp == nil {
		fishComp.StopFishing()
		return
	}

	// Select fish based on conditions
	selectedFish, weight := fs.selectFish(spotComp, fishComp)
	if selectedFish == nil {
		// No valid fish - restart wait with shorter time
		fishComp.CompleteCast(fs.BaseWaitTime * 0.5)
		return
	}

	// Trigger bite
	fishComp.TriggerBite(selectedFish.ID, weight, selectedFish.Difficulty, fs.BiteWindowTime)

	log.WithFields(log.Fields{
		"entity_id": fisher.ID,
		"fish_type": selectedFish.ID,
		"weight":    weight,
	}).Debug("Fish bite triggered")

	if fs.OnBiteCallback != nil {
		fs.OnBiteCallback(fisher, selectedFish.ID)
	}
}

// selectFish chooses a fish from the spot based on conditions.
func (fs *FishingSystem) selectFish(spot *FishingSpotComponent, fishComp *FishingComponent) (*FishType, float64) {
	baitType, _ := fishComp.GetBait()
	skillLevel := fishComp.GetSkillLevel()
	castDistance := fishComp.CastDistance
	timeOfDay := fs.CurrentTimeOfDay()
	weather := fs.CurrentWeather()

	eligibleFish := fs.buildEligibleFishList(spot, baitType, skillLevel, castDistance, timeOfDay, weather)
	if len(eligibleFish.fish) == 0 || eligibleFish.totalWeight <= 0 {
		return nil, 0
	}

	selected := fs.selectRandomFish(eligibleFish)
	if selected == nil {
		selected = eligibleFish.fish[0].fish
	}

	fishWeight := fs.calculateFishWeight(selected, skillLevel)

	return selected, fishWeight
}

// weightedFish represents a fish type with its selection weight.
type weightedFish struct {
	fish   *FishType
	weight float64
}

// eligibleFishList holds the weighted fish eligible for selection.
type eligibleFishList struct {
	fish        []weightedFish
	totalWeight float64
}

// buildEligibleFishList constructs the list of fish eligible for catching based on conditions.
func (fs *FishingSystem) buildEligibleFishList(spot *FishingSpotComponent, baitType string, skillLevel int, castDistance float64, timeOfDay TimeOfDay, weather string) eligibleFishList {
	eligible := make([]weightedFish, 0)
	totalWeight := 0.0

	for fishID, spawnWeight := range spot.FishPopulation {
		fish := fs.fishTypes[fishID]
		if fish == nil {
			continue
		}

		if !fs.isFishEligible(fish, spot, baitType, skillLevel, castDistance, weather) {
			continue
		}

		effectiveWeight := fs.calculateEffectiveWeight(fish, spot, spawnWeight, skillLevel, timeOfDay)

		if effectiveWeight > 0 {
			eligible = append(eligible, weightedFish{fish: fish, weight: effectiveWeight})
			totalWeight += effectiveWeight
		}
	}

	return eligibleFishList{fish: eligible, totalWeight: totalWeight}
}

// isFishEligible checks if a fish meets all requirements for catching.
func (fs *FishingSystem) isFishEligible(fish *FishType, spot *FishingSpotComponent, baitType string, skillLevel int, castDistance float64, weather string) bool {
	if skillLevel < fish.MinSkill {
		return false
	}

	accessibleDepth := DepthLevel(int(castDistance * 3))
	if fish.MinDepth > accessibleDepth {
		return false
	}

	if !fs.checkWaterTypeMatch(fish, spot.WaterType) {
		return false
	}

	if fish.RequiredBait != "" && fish.RequiredBait != baitType {
		return false
	}

	if fish.WeatherCondition != "" && fish.WeatherCondition != weather {
		return false
	}

	return true
}

// checkWaterTypeMatch verifies if the fish can live in the spot's water type.
func (fs *FishingSystem) checkWaterTypeMatch(fish *FishType, spotWaterType WaterType) bool {
	for _, wt := range fish.WaterTypes {
		if wt == spotWaterType {
			return true
		}
	}
	return false
}

// calculateEffectiveWeight computes the fish selection weight with all modifiers applied.
func (fs *FishingSystem) calculateEffectiveWeight(fish *FishType, spot *FishingSpotComponent, spawnWeight float64, skillLevel int, timeOfDay TimeOfDay) float64 {
	effectiveWeight := spawnWeight

	if fish.BestTime == timeOfDay || fish.BestTime == TimeAny {
		effectiveWeight *= 1.5
	}

	effectiveWeight *= fs.getRarityMultiplier(fish.Rarity, spot.RareFishBonus)

	if fish.Rarity != FishRarityCommon {
		skillBonus := 1.0 + float64(skillLevel-fish.MinSkill)*0.02
		effectiveWeight *= skillBonus
	}

	return effectiveWeight
}

// getRarityMultiplier returns the weight multiplier based on fish rarity.
func (fs *FishingSystem) getRarityMultiplier(rarity FishRarity, rareFishBonus float64) float64 {
	switch rarity {
	case FishRarityCommon:
		return 1.0
	case FishRarityUncommon:
		return 0.5
	case FishRarityRare:
		return 0.2 * rareFishBonus
	case FishRarityEpic:
		return 0.05 * rareFishBonus
	case FishRarityLegendary:
		return 0.01 * rareFishBonus
	default:
		return 1.0
	}
}

// selectRandomFish performs weighted random selection from eligible fish.
func (fs *FishingSystem) selectRandomFish(eligible eligibleFishList) *FishType {
	seed := time.Now().UnixNano()
	rng := rand.New(rand.NewSource(seed))

	roll := rng.Float64() * eligible.totalWeight
	cumulative := 0.0

	for _, wf := range eligible.fish {
		cumulative += wf.weight
		if roll <= cumulative {
			return wf.fish
		}
	}

	return nil
}

// calculateFishWeight determines the weight of the caught fish with skill bonuses.
func (fs *FishingSystem) calculateFishWeight(fish *FishType, skillLevel int) float64 {
	seed := time.Now().UnixNano()
	rng := rand.New(rand.NewSource(seed))

	weightRange := fish.MaxWeight - fish.BaseWeight
	fishWeight := fish.BaseWeight + rng.Float64()*weightRange

	skillWeightBonus := 1.0 + float64(skillLevel)*0.005
	fishWeight *= skillWeightBonus

	return fishWeight
}

// processBite handles the bite window phase.
func (fs *FishingSystem) processBite(fisher *Entity, fishComp *FishingComponent, deltaTime float64) {
	if fishComp.UpdateBiteWindow(deltaTime) {
		// Window expired - fish got away
		fishTypeID, _ := fishComp.GetHookedFish()
		fishComp.MissBite()

		log.WithFields(log.Fields{
			"entity_id": fisher.ID,
			"reason":    "missed_bite",
		}).Debug("Fish escaped - missed bite")

		if fs.OnEscapeCallback != nil {
			fs.OnEscapeCallback(fisher, fishTypeID, "missed_bite")
		}
	}
}

// processReeling handles the reeling minigame phase.
func (fs *FishingSystem) processReeling(fisher *Entity, fishComp *FishingComponent, deltaTime float64) {
	// Update fish struggle
	seed := time.Now().UnixNano()
	rng := rand.New(rand.NewSource(seed))
	fishComp.UpdateStruggle(deltaTime, rng)

	// Get reel input from entity (stubbed - would come from input system)
	reelInput := fs.getReelInput(fisher)

	result := fishComp.UpdateReeling(deltaTime, reelInput)

	switch result {
	case 1: // Caught
		fs.completeCatch(fisher, fishComp)
	case -1: // Escaped
		fishTypeID, _ := fishComp.GetHookedFish()
		fishComp.FishEscaped()

		log.WithFields(log.Fields{
			"entity_id": fisher.ID,
			"reason":    "line_break",
		}).Debug("Fish escaped - line break")

		if fs.OnEscapeCallback != nil {
			fs.OnEscapeCallback(fisher, fishTypeID, "line_break")
		}
	}
}

// getReelInput retrieves reel input for an entity.
func (fs *FishingSystem) getReelInput(entity *Entity) float64 {
	// Check for input component
	comp, ok := entity.GetComponent("input")
	if !ok {
		return 0
	}

	// InputProvider interface has IsActionPressed method
	inputProvider, ok := comp.(InputProvider)
	if !ok || inputProvider == nil {
		return 0
	}

	// Use primary action for reeling
	if inputProvider.IsActionPressed() {
		return 1.0
	}
	return 0
}

// completeCatch finalizes a successful catch.
func (fs *FishingSystem) completeCatch(fisher *Entity, fishComp *FishingComponent) {
	// Get spot for location
	spotID := fishComp.GetTargetSpotID()
	location := "unknown"
	if spotEntity, ok := fs.fishingSpots[spotID]; ok {
		if posComp, ok := spotEntity.GetComponent("position"); ok {
			if pos, ok := posComp.(*PositionComponent); ok {
				location = pos.String()
			}
		}
	}

	// Complete the catch
	caught := fishComp.CompleteCatch(location, time.Now().Unix())
	if caught == nil {
		return
	}

	// Calculate XP based on fish rarity and difficulty
	fishType := fs.fishTypes[caught.FishTypeID]
	xpGained := fs.XPPerCatch
	if fishType != nil {
		switch fishType.Rarity {
		case FishRarityUncommon:
			xpGained = int(float64(xpGained) * 1.5)
		case FishRarityRare:
			xpGained *= 2
		case FishRarityEpic:
			xpGained *= 3
		case FishRarityLegendary:
			xpGained *= 5
		}
		xpGained = int(float64(xpGained) * (1.0 + fishType.Difficulty))
	}

	// Grant XP
	if fishComp.AddXP(xpGained) {
		newLevel := fishComp.GetSkillLevel()
		log.WithFields(log.Fields{
			"entity_id": fisher.ID,
			"new_level": newLevel,
		}).Info("Fishing skill leveled up")

		if fs.OnLevelUpCallback != nil {
			fs.OnLevelUpCallback(fisher, newLevel)
		}
	}

	// Set spot cooldown
	if spotEntity, ok := fs.fishingSpots[spotID]; ok {
		if comp, ok := spotEntity.GetComponent("fishing_spot"); ok {
			if spotComp, ok := comp.(*FishingSpotComponent); ok {
				fs.SpotSetCooldown(spotComp, fs.SpotCooldown)
				fs.SpotRemoveFisher(spotComp)
			}
		}
	}

	log.WithFields(log.Fields{
		"entity_id":   fisher.ID,
		"fish_type":   caught.FishTypeID,
		"weight":      caught.Weight,
		"is_record":   caught.IsRecord,
		"xp_gained":   xpGained,
		"skill_level": fishComp.GetSkillLevel(),
	}).Debug("Fish caught")

	if fs.OnCatchCallback != nil {
		fs.OnCatchCallback(fisher, caught)
	}
}

// StartFishing initiates fishing at a spot.
func (fs *FishingSystem) StartFishing(fisher *Entity, spotID uint64) bool {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	comp, ok := fisher.GetComponent("fishing")
	if !ok {
		return false
	}
	fishComp, ok := comp.(*FishingComponent)
	if !ok || fishComp == nil {
		return false
	}

	// Check if already fishing
	if fishComp.GetState() != FishingStateIdle {
		return false
	}

	// Get fishing spot
	spotEntity, ok := fs.fishingSpots[spotID]
	if !ok || spotEntity == nil {
		return false
	}

	spotCompRaw, ok := spotEntity.GetComponent("fishing_spot")
	if !ok {
		return false
	}
	spotComp, ok := spotCompRaw.(*FishingSpotComponent)
	if !ok || spotComp == nil {
		return false
	}

	// Validate can fish
	if !fs.SpotCanFish(spotComp) {
		return false
	}

	// Check bait
	_, baitCount := fishComp.GetBait()
	if baitCount <= 0 {
		log.WithFields(log.Fields{
			"entity_id": fisher.ID,
		}).Debug("Cannot fish - no bait")
		return false
	}

	// Use bait
	fishComp.UseBait()

	// Add fisher to spot
	if !fs.SpotAddFisher(spotComp) {
		return false
	}

	// Start casting
	fishComp.StartCasting(spotID)

	log.WithFields(log.Fields{
		"entity_id":  fisher.ID,
		"spot_id":    spotID,
		"water_type": spotComp.WaterType,
	}).Debug("Started fishing")

	return true
}

// Cast performs the cast with given power and starts waiting.
func (fs *FishingSystem) Cast(fisher *Entity, power float64) bool {
	comp, ok := fisher.GetComponent("fishing")
	if !ok {
		return false
	}
	fishComp, ok := comp.(*FishingComponent)
	if !ok || fishComp == nil {
		return false
	}

	if fishComp.GetState() != FishingStateCasting {
		return false
	}

	fishComp.SetCastDistance(power)

	// Calculate wait time based on cast distance and skill
	skillLevel := fishComp.GetSkillLevel()
	waitTime := fs.BaseWaitTime * (1.0 - power*0.3) // Closer casts wait less
	waitTime *= (1.0 - float64(skillLevel)*0.005)   // Higher skill reduces wait
	if waitTime < 3.0 {
		waitTime = 3.0
	}

	fishComp.CompleteCast(waitTime)

	log.WithFields(log.Fields{
		"entity_id":     fisher.ID,
		"cast_distance": power,
		"wait_time":     waitTime,
	}).Debug("Cast completed")

	return true
}

// Hook attempts to hook a fish when bite detected.
func (fs *FishingSystem) Hook(fisher *Entity) bool {
	comp, ok := fisher.GetComponent("fishing")
	if !ok {
		return false
	}
	fishComp, ok := comp.(*FishingComponent)
	if !ok || fishComp == nil {
		return false
	}

	if fishComp.GetState() != FishingStateBite {
		return false
	}

	if fishComp.HookFish() {
		log.WithFields(log.Fields{
			"entity_id": fisher.ID,
		}).Debug("Fish hooked")
		return true
	}
	return false
}

// CancelFishing stops an active fishing session.
func (fs *FishingSystem) CancelFishing(fisher *Entity) {
	comp, ok := fisher.GetComponent("fishing")
	if !ok {
		return
	}
	fishComp, ok := comp.(*FishingComponent)
	if !ok || fishComp == nil {
		return
	}

	if fishComp.GetState() == FishingStateIdle {
		return
	}

	// Remove fisher from spot
	spotID := fishComp.GetTargetSpotID()
	fs.mu.RLock()
	if spotEntity, ok := fs.fishingSpots[spotID]; ok {
		if spotCompRaw, ok := spotEntity.GetComponent("fishing_spot"); ok {
			if spotComp, ok := spotCompRaw.(*FishingSpotComponent); ok {
				fs.SpotRemoveFisher(spotComp)
			}
		}
	}
	fs.mu.RUnlock()

	fishComp.StopFishing()
	log.WithFields(log.Fields{
		"entity_id": fisher.ID,
	}).Debug("Fishing canceled")
}

// GetNearbyFishingSpots returns fishing spots within range of a position.
func (fs *FishingSystem) GetNearbyFishingSpots(x, y, maxRange float64) []*Entity {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	var nearby []*Entity
	rangeSquared := maxRange * maxRange

	for _, entity := range fs.fishingSpots {
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

// GenerateFishingSpot creates a fishing spot entity with procedural fish population.
func (fs *FishingSystem) GenerateFishingSpot(seed int64, waterType WaterType, depthLevel DepthLevel, biome string, x, y float64) *Entity {
	rng := rand.New(rand.NewSource(seed))

	// Get next entity ID
	fs.mu.Lock()
	entityID := fs.nextEntityID
	fs.nextEntityID++
	fs.mu.Unlock()

	entity := NewEntity(entityID)

	// Position
	entity.AddComponent(&PositionComponent{X: x, Y: y})

	// Fishing spot
	spot := NewFishingSpotComponent(waterType, depthLevel, biome)

	// Randomize properties
	spot.MaxConcurrentFishers = 3 + rng.Intn(5)  // 3-7 fishers
	spot.RareFishBonus = 0.8 + rng.Float64()*0.4 // 0.8-1.2

	// Populate with fish based on water type and depth
	fs.mu.RLock()
	for id, fishType := range fs.fishTypes {
		// Check water type match
		waterMatch := false
		for _, wt := range fishType.WaterTypes {
			if wt == waterType {
				waterMatch = true
				break
			}
		}
		if !waterMatch {
			continue
		}

		// Check depth
		if fishType.MinDepth > depthLevel {
			continue
		}

		// Check biome if specified
		if len(fishType.BiomeTypes) > 0 {
			biomeMatch := false
			for _, bt := range fishType.BiomeTypes {
				if bt == biome {
					biomeMatch = true
					break
				}
			}
			if !biomeMatch {
				continue
			}
		}

		// Add to population with weighted spawn rate
		baseWeight := 1.0
		switch fishType.Rarity {
		case FishRarityCommon:
			baseWeight = 10.0
		case FishRarityUncommon:
			baseWeight = 5.0
		case FishRarityRare:
			baseWeight = 2.0
		case FishRarityEpic:
			baseWeight = 0.5
		case FishRarityLegendary:
			baseWeight = 0.1
		}

		// Add some randomness
		spawnWeight := baseWeight * (0.8 + rng.Float64()*0.4)
		fs.SpotAddFishType(spot, id, spawnWeight)
	}
	fs.mu.RUnlock()

	entity.AddComponent(spot)

	log.WithFields(log.Fields{
		"entity_id":   entity.ID,
		"water_type":  waterType,
		"depth_level": depthLevel,
		"biome":       biome,
		"seed":        seed,
		"fish_count":  len(spot.FishPopulation),
	}).Debug("Generated fishing spot")

	return entity
}

// GetFishingSpotCount returns the count of tracked fishing spots.
func (fs *FishingSystem) GetFishingSpotCount() int {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	return len(fs.fishingSpots)
}

// GetFishTypeCount returns the count of registered fish types.
func (fs *FishingSystem) GetFishTypeCount() int {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	return len(fs.fishTypes)
}

// --- FishingSpot Component Helpers (ECS Pattern) ---
// These methods operate on FishingSpotComponent data following ECS pattern
// where systems contain all logic and components are pure data.

// SpotAddFishType adds a fish type to a fishing spot's population.
func (fs *FishingSystem) SpotAddFishType(spot *FishingSpotComponent, fishTypeID string, spawnWeight float64) {
	spot.mu.Lock()
	defer spot.mu.Unlock()
	spot.FishPopulation[fishTypeID] = spawnWeight
}

// SpotRemoveFishType removes a fish type from a fishing spot.
func (fs *FishingSystem) SpotRemoveFishType(spot *FishingSpotComponent, fishTypeID string) {
	spot.mu.Lock()
	defer spot.mu.Unlock()
	delete(spot.FishPopulation, fishTypeID)
}

// SpotGetFishTypes returns all fish type IDs at a fishing spot.
func (fs *FishingSystem) SpotGetFishTypes(spot *FishingSpotComponent) []string {
	spot.mu.RLock()
	defer spot.mu.RUnlock()
	types := make([]string, 0, len(spot.FishPopulation))
	for id := range spot.FishPopulation {
		types = append(types, id)
	}
	return types
}

// SpotGetSpawnWeight returns the spawn weight for a fish type at a fishing spot.
func (fs *FishingSystem) SpotGetSpawnWeight(spot *FishingSpotComponent, fishTypeID string) float64 {
	spot.mu.RLock()
	defer spot.mu.RUnlock()
	return spot.FishPopulation[fishTypeID]
}

// SpotCanFish checks if another fisher can use a fishing spot.
func (fs *FishingSystem) SpotCanFish(spot *FishingSpotComponent) bool {
	spot.mu.RLock()
	defer spot.mu.RUnlock()
	return spot.IsActive && spot.CurrentFishers < spot.MaxConcurrentFishers && spot.CooldownTimer <= 0
}

// SpotAddFisher increments the current fisher count at a fishing spot.
// Returns false if spot is at capacity.
func (fs *FishingSystem) SpotAddFisher(spot *FishingSpotComponent) bool {
	spot.mu.Lock()
	defer spot.mu.Unlock()
	if spot.CurrentFishers >= spot.MaxConcurrentFishers {
		return false
	}
	spot.CurrentFishers++
	return true
}

// SpotRemoveFisher decrements the current fisher count at a fishing spot.
func (fs *FishingSystem) SpotRemoveFisher(spot *FishingSpotComponent) {
	spot.mu.Lock()
	defer spot.mu.Unlock()
	if spot.CurrentFishers > 0 {
		spot.CurrentFishers--
	}
}

// SpotGetCurrentFishers returns the current fisher count at a fishing spot.
func (fs *FishingSystem) SpotGetCurrentFishers(spot *FishingSpotComponent) int {
	spot.mu.RLock()
	defer spot.mu.RUnlock()
	return spot.CurrentFishers
}

// SpotUpdateCooldown processes cooldown timer for a fishing spot.
func (fs *FishingSystem) SpotUpdateCooldown(spot *FishingSpotComponent, deltaTime float64) {
	spot.mu.Lock()
	defer spot.mu.Unlock()
	if spot.CooldownTimer > 0 {
		spot.CooldownTimer -= deltaTime
		if spot.CooldownTimer < 0 {
			spot.CooldownTimer = 0
		}
	}
}

// SpotSetCooldown sets the cooldown timer for a fishing spot.
func (fs *FishingSystem) SpotSetCooldown(spot *FishingSpotComponent, seconds float64) {
	spot.mu.Lock()
	defer spot.mu.Unlock()
	spot.CooldownTimer = seconds
}
