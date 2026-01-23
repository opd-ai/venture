// Package engine provides fishing components for the ECS.
// This file implements components for fishing spots and player fishing state.
// Phase 96: Fishing System

package engine

import (
	"encoding/json"
	"math/rand"
	"sync"
)

// WaterType defines the type of water body for fishing.
type WaterType string

const (
	// WaterTypeFreshwater for rivers, lakes, ponds.
	WaterTypeFreshwater WaterType = "freshwater"
	// WaterTypeSaltwater for oceans and seas.
	WaterTypeSaltwater WaterType = "saltwater"
	// WaterTypeMagical for enchanted water bodies.
	WaterTypeMagical WaterType = "magical"
)

// AllWaterTypes returns all valid water types.
func AllWaterTypes() []WaterType {
	return []WaterType{
		WaterTypeFreshwater,
		WaterTypeSaltwater,
		WaterTypeMagical,
	}
}

// DepthLevel defines the depth of fishing spot.
type DepthLevel int

const (
	// DepthShallow for near-shore fishing.
	DepthShallow DepthLevel = iota
	// DepthMedium for standard depth.
	DepthMedium
	// DepthDeep for deep water fishing.
	DepthDeep
)

// TimeOfDay represents fishing time conditions.
type TimeOfDay string

const (
	// TimeDawn for early morning.
	TimeDawn TimeOfDay = "dawn"
	// TimeDay for daytime.
	TimeDay TimeOfDay = "day"
	// TimeDusk for evening.
	TimeDusk TimeOfDay = "dusk"
	// TimeNight for nighttime.
	TimeNight TimeOfDay = "night"
	// TimeAny for fish available at all times.
	TimeAny TimeOfDay = "any"
)

// FishRarity defines rarity tiers for fish.
type FishRarity string

const (
	// FishRarityCommon for frequently caught fish.
	FishRarityCommon FishRarity = "common"
	// FishRarityUncommon for moderately rare fish.
	FishRarityUncommon FishRarity = "uncommon"
	// FishRarityRare for rare fish.
	FishRarityRare FishRarity = "rare"
	// FishRarityEpic for very rare fish.
	FishRarityEpic FishRarity = "epic"
	// FishRarityLegendary for extremely rare fish.
	FishRarityLegendary FishRarity = "legendary"
)

// FishType defines a type of fish that can be caught.
type FishType struct {
	// ID is the unique identifier for this fish type.
	ID string `json:"id"`

	// Name is the display name.
	Name string `json:"name"`

	// Rarity determines catch probability.
	Rarity FishRarity `json:"rarity"`

	// WaterTypes where this fish can be found.
	WaterTypes []WaterType `json:"water_types"`

	// MinDepth is the minimum depth level required.
	MinDepth DepthLevel `json:"min_depth"`

	// BestTime is the optimal time to catch this fish.
	BestTime TimeOfDay `json:"best_time"`

	// MinSkill is the fishing skill required.
	MinSkill int `json:"min_skill"`

	// BaseWeight is the minimum catch weight.
	BaseWeight float64 `json:"base_weight"`

	// MaxWeight is the maximum catch weight.
	MaxWeight float64 `json:"max_weight"`

	// Difficulty affects the minigame tension.
	Difficulty float64 `json:"difficulty"`

	// BiomeTypes where this fish spawns (empty = all biomes).
	BiomeTypes []string `json:"biome_types"`

	// RequiredBait is the bait type needed (empty = any).
	RequiredBait string `json:"required_bait"`

	// WeatherCondition for special fish (empty = any).
	WeatherCondition string `json:"weather_condition"`
}

// CaughtFish represents a fish that has been caught.
type CaughtFish struct {
	// FishTypeID references the fish type.
	FishTypeID string `json:"fish_type_id"`

	// Weight is the actual weight of this catch.
	Weight float64 `json:"weight"`

	// CaughtAt is the unix timestamp of catch.
	CaughtAt int64 `json:"caught_at"`

	// Location is where the fish was caught.
	Location string `json:"location"`

	// IsRecord indicates if this is a personal record.
	IsRecord bool `json:"is_record"`
}

// FishingSpotComponent marks a location as fishable with specific fish populations.
type FishingSpotComponent struct {
	mu sync.RWMutex

	// FishPopulation maps fish type IDs to spawn weights.
	FishPopulation map[string]float64 `json:"fish_population"`

	// DepthLevel indicates the depth of this spot.
	DepthLevel DepthLevel `json:"depth_level"`

	// WaterType indicates the water type.
	WaterType WaterType `json:"water_type"`

	// BiomeType is the biome this spot is in.
	BiomeType string `json:"biome_type"`

	// RareFishBonus is a multiplier for rare fish chance.
	RareFishBonus float64 `json:"rare_fish_bonus"`

	// IsActive indicates the spot is currently fishable.
	IsActive bool `json:"is_active"`

	// CooldownTimer tracks time until next fish can be caught.
	CooldownTimer float64 `json:"cooldown_timer"`

	// MaxConcurrentFishers is the maximum number of simultaneous fishers.
	MaxConcurrentFishers int `json:"max_concurrent_fishers"`

	// CurrentFishers tracks active fishers at this spot.
	CurrentFishers int `json:"current_fishers"`
}

// NewFishingSpotComponent creates a new fishing spot with default values.
func NewFishingSpotComponent(waterType WaterType, depthLevel DepthLevel, biomeType string) *FishingSpotComponent {
	return &FishingSpotComponent{
		FishPopulation:       make(map[string]float64),
		DepthLevel:           depthLevel,
		WaterType:            waterType,
		BiomeType:            biomeType,
		RareFishBonus:        1.0,
		IsActive:             true,
		CooldownTimer:        0,
		MaxConcurrentFishers: 5,
		CurrentFishers:       0,
	}
}

// Type returns the component type identifier.
func (f *FishingSpotComponent) Type() string {
	return "fishing_spot"
}

// Serialize converts the component to JSON bytes.
func (f *FishingSpotComponent) Serialize() ([]byte, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return json.Marshal(f)
}

// Deserialize loads the component from JSON bytes.
func (f *FishingSpotComponent) Deserialize(data []byte) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	return json.Unmarshal(data, f)
}

// FishingState represents the current state of fishing activity.
type FishingState string

const (
	// FishingStateIdle for not fishing.
	FishingStateIdle FishingState = "idle"
	// FishingStateCasting for casting the line.
	FishingStateCasting FishingState = "casting"
	// FishingStateWaiting for waiting for a bite.
	FishingStateWaiting FishingState = "waiting"
	// FishingStateBite for when a fish bites.
	FishingStateBite FishingState = "bite"
	// FishingStateReeling for actively reeling in.
	FishingStateReeling FishingState = "reeling"
	// FishingStateCaught for when fish is landed.
	FishingStateCaught FishingState = "caught"
	// FishingStateEscaped for when fish escapes.
	FishingStateEscaped FishingState = "escaped"
)

// FishingComponent tracks a player's fishing state and skill.
type FishingComponent struct {
	mu sync.RWMutex

	// FishingSkill is the player's fishing skill level (1-100).
	FishingSkill int `json:"fishing_skill"`

	// FishingXP tracks experience toward next skill level.
	FishingXP int `json:"fishing_xp"`

	// State is the current fishing activity state.
	State FishingState `json:"state"`

	// TargetSpotID is the entity ID of the fishing spot.
	TargetSpotID uint64 `json:"target_spot_id"`

	// CastDistance is how far the line was cast (0.0-1.0).
	CastDistance float64 `json:"cast_distance"`

	// BaitType is the currently equipped bait.
	BaitType string `json:"bait_type"`

	// BaitCount is the remaining bait.
	BaitCount int `json:"bait_count"`

	// RodQuality affects catch rates and minigame (1.0 = standard).
	RodQuality float64 `json:"rod_quality"`

	// TensionLevel is line tension during reeling (0.0-1.0).
	TensionLevel float64 `json:"tension_level"`

	// MaxTension is the breaking point (based on rod quality).
	MaxTension float64 `json:"max_tension"`

	// WaitTimer counts time waiting for a bite.
	WaitTimer float64 `json:"wait_timer"`

	// BiteWindowTimer counts down from bite detection.
	BiteWindowTimer float64 `json:"bite_window_timer"`

	// ReelProgress tracks how much fish is reeled in (0.0-1.0).
	ReelProgress float64 `json:"reel_progress"`

	// HookedFishTypeID is the fish currently on the line.
	HookedFishTypeID string `json:"hooked_fish_type_id"`

	// HookedFishWeight is the weight of the hooked fish.
	HookedFishWeight float64 `json:"hooked_fish_weight"`

	// HookedFishDifficulty affects minigame tension.
	HookedFishDifficulty float64 `json:"hooked_fish_difficulty"`

	// CaughtFish is the list of fish caught this session.
	CaughtFish []CaughtFish `json:"caught_fish"`

	// TotalCaught tracks lifetime catch count per fish type.
	TotalCaught map[string]int `json:"total_caught"`

	// PersonalRecords tracks heaviest catch per fish type.
	PersonalRecords map[string]float64 `json:"personal_records"`

	// FishStruggleDirection is the direction fish is fighting (-1, 0, 1).
	FishStruggleDirection int `json:"fish_struggle_direction"`

	// StruggleTimer tracks time until next struggle change.
	StruggleTimer float64 `json:"struggle_timer"`
}

// NewFishingComponent creates a new fishing component with defaults.
func NewFishingComponent() *FishingComponent {
	return &FishingComponent{
		FishingSkill:          1,
		FishingXP:             0,
		State:                 FishingStateIdle,
		TargetSpotID:          0,
		CastDistance:          0,
		BaitType:              "basic",
		BaitCount:             10,
		RodQuality:            1.0,
		TensionLevel:          0,
		MaxTension:            1.0,
		WaitTimer:             0,
		BiteWindowTimer:       0,
		ReelProgress:          0,
		HookedFishTypeID:      "",
		HookedFishWeight:      0,
		HookedFishDifficulty:  0,
		CaughtFish:            make([]CaughtFish, 0),
		TotalCaught:           make(map[string]int),
		PersonalRecords:       make(map[string]float64),
		FishStruggleDirection: 0,
		StruggleTimer:         0,
	}
}

// Type returns the component type identifier.
func (f *FishingComponent) Type() string {
	return "fishing"
}

// StartCasting begins the casting phase.
func (f *FishingComponent) StartCasting(spotID uint64) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.State = FishingStateCasting
	f.TargetSpotID = spotID
	f.CastDistance = 0
	f.TensionLevel = 0
	f.ReelProgress = 0
	f.HookedFishTypeID = ""
}

// SetCastDistance sets the cast power (0.0-1.0).
func (f *FishingComponent) SetCastDistance(distance float64) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if distance < 0 {
		distance = 0
	}
	if distance > 1 {
		distance = 1
	}
	f.CastDistance = distance
}

// CompleteCast finishes casting and starts waiting.
func (f *FishingComponent) CompleteCast(baseWaitTime float64) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.State = FishingStateWaiting
	// Wait time varies based on cast distance and randomness
	f.WaitTimer = baseWaitTime
}

// GetState returns the current fishing state.
func (f *FishingComponent) GetState() FishingState {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.State
}

// GetTargetSpotID returns the target fishing spot ID.
func (f *FishingComponent) GetTargetSpotID() uint64 {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.TargetSpotID
}

// UpdateWait decrements wait timer, returns true when fish should bite.
func (f *FishingComponent) UpdateWait(deltaTime float64) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.State != FishingStateWaiting {
		return false
	}
	f.WaitTimer -= deltaTime
	return f.WaitTimer <= 0
}

// TriggerBite transitions to bite state with a time window.
func (f *FishingComponent) TriggerBite(fishTypeID string, weight, difficulty, windowTime float64) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.State = FishingStateBite
	f.HookedFishTypeID = fishTypeID
	f.HookedFishWeight = weight
	f.HookedFishDifficulty = difficulty
	f.BiteWindowTimer = windowTime
}

// UpdateBiteWindow decrements bite window, returns true when window expires.
func (f *FishingComponent) UpdateBiteWindow(deltaTime float64) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.State != FishingStateBite {
		return false
	}
	f.BiteWindowTimer -= deltaTime
	return f.BiteWindowTimer <= 0
}

// HookFish transitions from bite to reeling.
func (f *FishingComponent) HookFish() bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.State != FishingStateBite || f.HookedFishTypeID == "" {
		return false
	}
	f.State = FishingStateReeling
	f.ReelProgress = 0
	f.TensionLevel = 0.3 // Initial hook tension
	return true
}

// UpdateReeling processes reeling minigame, returns: 0=continue, 1=caught, -1=escaped.
func (f *FishingComponent) UpdateReeling(deltaTime, reelInput float64) int {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.State != FishingStateReeling {
		return 0
	}

	// Fish struggle affects tension
	struggleTension := float64(f.FishStruggleDirection) * f.HookedFishDifficulty * 0.1 * deltaTime

	// Reeling adds tension and progress
	reelTension := reelInput * 0.5 * deltaTime
	reelProgressGain := reelInput * (1.0 + float64(f.FishingSkill)*0.01) * 0.1 * deltaTime

	// Apply tension changes
	f.TensionLevel += struggleTension + reelTension

	// Tension naturally decreases when not reeling
	if reelInput == 0 {
		f.TensionLevel -= 0.2 * deltaTime
	}

	// Clamp tension
	if f.TensionLevel < 0 {
		f.TensionLevel = 0
	}

	// Check line break
	if f.TensionLevel > f.MaxTension {
		f.State = FishingStateEscaped
		return -1
	}

	// Progress increases only when tension is in good range (0.2-0.8)
	if f.TensionLevel >= 0.2 && f.TensionLevel <= 0.8 {
		f.ReelProgress += reelProgressGain
	}

	// Check if fish is landed
	if f.ReelProgress >= 1.0 {
		f.State = FishingStateCaught
		return 1
	}

	return 0
}

// UpdateStruggle changes fish struggle direction randomly.
func (f *FishingComponent) UpdateStruggle(deltaTime float64, rng *rand.Rand) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.State != FishingStateReeling {
		return
	}

	f.StruggleTimer -= deltaTime
	if f.StruggleTimer <= 0 {
		// Randomize struggle direction
		roll := rng.Float64()
		if roll < 0.3 {
			f.FishStruggleDirection = -1
		} else if roll > 0.7 {
			f.FishStruggleDirection = 1
		} else {
			f.FishStruggleDirection = 0
		}
		// Set next struggle change time (based on difficulty)
		f.StruggleTimer = 0.5 + rng.Float64()*(1.0-f.HookedFishDifficulty*0.5)
	}
}

// GetTensionLevel returns current line tension.
func (f *FishingComponent) GetTensionLevel() float64 {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.TensionLevel
}

// GetReelProgress returns current reel progress.
func (f *FishingComponent) GetReelProgress() float64 {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.ReelProgress
}

// GetHookedFish returns the hooked fish info.
func (f *FishingComponent) GetHookedFish() (string, float64) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.HookedFishTypeID, f.HookedFishWeight
}

// CompleteCatch finalizes a successful catch.
func (f *FishingComponent) CompleteCatch(location string, timestamp int64) *CaughtFish {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.HookedFishTypeID == "" {
		return nil
	}

	// Check if personal record
	isRecord := false
	if prev, ok := f.PersonalRecords[f.HookedFishTypeID]; !ok || f.HookedFishWeight > prev {
		f.PersonalRecords[f.HookedFishTypeID] = f.HookedFishWeight
		isRecord = true
	}

	caught := CaughtFish{
		FishTypeID: f.HookedFishTypeID,
		Weight:     f.HookedFishWeight,
		CaughtAt:   timestamp,
		Location:   location,
		IsRecord:   isRecord,
	}

	f.CaughtFish = append(f.CaughtFish, caught)
	f.TotalCaught[f.HookedFishTypeID]++

	// Reset state
	f.State = FishingStateIdle
	f.HookedFishTypeID = ""
	f.HookedFishWeight = 0
	f.ReelProgress = 0
	f.TensionLevel = 0

	return &caught
}

// MissBite handles missed bite window.
func (f *FishingComponent) MissBite() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.State = FishingStateEscaped
	f.HookedFishTypeID = ""
	f.HookedFishWeight = 0
}

// FishEscaped handles line break.
func (f *FishingComponent) FishEscaped() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.State = FishingStateEscaped
	f.HookedFishTypeID = ""
	f.HookedFishWeight = 0
	f.ReelProgress = 0
	f.TensionLevel = 0
}

// StopFishing cancels all fishing activity.
func (f *FishingComponent) StopFishing() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.State = FishingStateIdle
	f.TargetSpotID = 0
	f.CastDistance = 0
	f.TensionLevel = 0
	f.ReelProgress = 0
	f.HookedFishTypeID = ""
	f.HookedFishWeight = 0
	f.WaitTimer = 0
	f.BiteWindowTimer = 0
}

// AddXP adds fishing experience and handles level ups.
func (f *FishingComponent) AddXP(amount int) bool {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.FishingXP += amount
	xpNeeded := f.xpForNextLevel()

	if f.FishingXP >= xpNeeded && f.FishingSkill < 100 {
		f.FishingXP -= xpNeeded
		f.FishingSkill++
		return true // Leveled up
	}
	return false
}

// xpForNextLevel calculates XP needed for next level.
func (f *FishingComponent) xpForNextLevel() int {
	return 100 + (f.FishingSkill * 50)
}

// GetSkillLevel returns current skill level safely.
func (f *FishingComponent) GetSkillLevel() int {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.FishingSkill
}

// GetTotalCaught returns lifetime catch count for a fish type.
func (f *FishingComponent) GetTotalCaught(fishTypeID string) int {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.TotalCaught[fishTypeID]
}

// GetPersonalRecord returns the heaviest catch for a fish type.
func (f *FishingComponent) GetPersonalRecord(fishTypeID string) float64 {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.PersonalRecords[fishTypeID]
}

// GetSessionCaught returns fish caught this session.
func (f *FishingComponent) GetSessionCaught() []CaughtFish {
	f.mu.RLock()
	defer f.mu.RUnlock()
	result := make([]CaughtFish, len(f.CaughtFish))
	copy(result, f.CaughtFish)
	return result
}

// UseBait consumes one bait, returns false if no bait.
func (f *FishingComponent) UseBait() bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.BaitCount <= 0 {
		return false
	}
	f.BaitCount--
	return true
}

// AddBait adds bait to inventory. Replaces type if current bait is empty.
func (f *FishingComponent) AddBait(baitType string, count int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	// Replace bait type if empty or same type
	if f.BaitCount == 0 || f.BaitType == "" || f.BaitType == baitType {
		f.BaitType = baitType
		f.BaitCount += count
	}
}

// GetBait returns current bait type and count.
func (f *FishingComponent) GetBait() (string, int) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.BaitType, f.BaitCount
}

// SetRodQuality sets the fishing rod quality multiplier.
func (f *FishingComponent) SetRodQuality(quality float64) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if quality < 0.1 {
		quality = 0.1
	}
	f.RodQuality = quality
	f.MaxTension = 0.8 + (quality * 0.4) // Higher quality = higher max tension
}

// GetRodQuality returns the current rod quality.
func (f *FishingComponent) GetRodQuality() float64 {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.RodQuality
}

// ClearSession clears session-specific catches.
func (f *FishingComponent) ClearSession() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.CaughtFish = make([]CaughtFish, 0)
}

// Serialize converts the component to JSON bytes.
func (f *FishingComponent) Serialize() ([]byte, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return json.Marshal(f)
}

// Deserialize loads the component from JSON bytes.
func (f *FishingComponent) Deserialize(data []byte) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	return json.Unmarshal(data, f)
}
