// Package engine provides the core ECS (Entity-Component-System) framework.
// This file implements the MiniGameSystem for Phase 27.1: Mini-Game Framework.
package engine

import (
	"fmt"
	"math/rand"
)

// MiniGameSystem manages mini-game instances and their lifecycle.
// It updates active mini-games, handles timeouts, and awards rewards.
//
// Phase 27.1: Mini-Game Framework
type MiniGameSystem struct {
	world *World
	// rngSource is used for deterministic reward generation
	rngSource rand.Source
}

// NewMiniGameSystem creates a new mini-game system.
// The system manages mini-game lifecycle including updates, timeouts, and rewards.
func NewMiniGameSystem(world *World) *MiniGameSystem {
	return &MiniGameSystem{
		world:     world,
		rngSource: rand.NewSource(42), // Default seed, can be changed
	}
}

// SetSeed sets the random seed for deterministic reward generation.
// This enables reproducible mini-game rewards for testing and multiplayer sync.
func (s *MiniGameSystem) SetSeed(seed int64) {
	s.rngSource = rand.NewSource(seed)
}

// Update processes active mini-games, updating their state and checking for timeouts.
// This method should be called every frame with the time elapsed since last update.
func (s *MiniGameSystem) Update(deltaTime float64) {
	entities := s.world.GetEntitiesWith("minigame")

	for _, entity := range entities {
		gameCompRaw, ok := entity.GetComponent("minigame")
		if !ok {
			continue
		}
		gameComp := gameCompRaw.(*MiniGameComponent)

		if !gameComp.Active {
			continue
		}

		// Update time elapsed
		gameComp.TimeElapsed += deltaTime

		// Update the game instance if present
		if gameComp.GameInstance != nil {
			if err := gameComp.GameInstance.Update(deltaTime); err != nil {
				// On error, end the game as failed
				s.EndGame(entity.ID, false)
				continue
			}

			// Check if game is complete
			if gameComp.GameInstance.IsComplete() {
				// Get reward and determine success
				reward := gameComp.GameInstance.GetReward()
				success := reward != nil
				gameComp.Reward = reward
				s.EndGame(entity.ID, success)
				continue
			}
		}

		// Check for timeout
		if gameComp.TimeLimit > 0 && gameComp.TimeElapsed >= gameComp.TimeLimit {
			s.EndGame(entity.ID, false)
		}
	}
}

// StartGame starts a mini-game on the specified entity.
// Creates or updates the MiniGameComponent with the specified parameters.
// The game instance should be set separately using SetGameInstance.
func (s *MiniGameSystem) StartGame(entityID uint64, gameType MiniGameType, difficulty float64) error {
	entity, ok := s.world.GetEntity(entityID)
	if !ok || entity == nil {
		return fmt.Errorf("entity %d not found", entityID)
	}

	// Validate difficulty
	if difficulty < 0.0 {
		difficulty = 0.0
	}
	if difficulty > 1.0 {
		difficulty = 1.0
	}

	// Create or update mini-game component
	gameComp := &MiniGameComponent{
		GameType:     gameType,
		Active:       true,
		State:        make(map[string]interface{}),
		Difficulty:   difficulty,
		TimeLimit:    s.getTimeLimit(gameType),
		TimeElapsed:  0,
		Reward:       s.generateReward(gameType, difficulty),
		GameInstance: nil, // Will be set by caller
	}

	entity.AddComponent(gameComp)
	return nil
}

// SetGameInstance sets the MiniGame implementation for an active mini-game.
// This allows the system to update and render the specific game implementation.
func (s *MiniGameSystem) SetGameInstance(entityID uint64, instance MiniGame) error {
	entity, ok := s.world.GetEntity(entityID)
	if !ok || entity == nil {
		return fmt.Errorf("entity %d not found", entityID)
	}

	gameCompRaw, ok := entity.GetComponent("minigame")
	if !ok {
		return fmt.Errorf("entity %d has no minigame component", entityID)
	}

	gameComp := gameCompRaw.(*MiniGameComponent)
	gameComp.GameInstance = instance

	return nil
}

// EndGame ends a mini-game, marking it inactive and awarding rewards if successful.
// success indicates whether the player won (true) or lost/timed out (false).
func (s *MiniGameSystem) EndGame(entityID uint64, success bool) error {
	entity, ok := s.world.GetEntity(entityID)
	if !ok || entity == nil {
		return fmt.Errorf("entity %d not found", entityID)
	}

	gameCompRaw, ok := entity.GetComponent("minigame")
	if !ok {
		return fmt.Errorf("entity %d has no minigame component", entityID)
	}

	gameComp := gameCompRaw.(*MiniGameComponent)
	gameComp.Active = false

	// Award rewards if successful
	if success && gameComp.Reward != nil {
		s.awardReward(entityID, gameComp.Reward)
	}

	return nil
}

// IsGameActive returns true if the entity has an active mini-game.
func (s *MiniGameSystem) IsGameActive(entityID uint64) bool {
	entity, ok := s.world.GetEntity(entityID)
	if !ok || entity == nil {
		return false
	}

	gameCompRaw, ok := entity.GetComponent("minigame")
	if !ok {
		return false
	}

	gameComp := gameCompRaw.(*MiniGameComponent)
	return gameComp.Active
}

// GetGameComponent returns the MiniGameComponent for an entity, or nil if not found.
func (s *MiniGameSystem) GetGameComponent(entityID uint64) *MiniGameComponent {
	entity, ok := s.world.GetEntity(entityID)
	if !ok || entity == nil {
		return nil
	}

	gameCompRaw, ok := entity.GetComponent("minigame")
	if !ok {
		return nil
	}

	return gameCompRaw.(*MiniGameComponent)
}

// getTimeLimit returns the default time limit for a game type in seconds.
// Based on Phase 27.2 specifications from ROADMAP_V4.md.
func (s *MiniGameSystem) getTimeLimit(gameType MiniGameType) float64 {
	switch gameType {
	case MiniGameCard:
		return 600.0 // 5-10 min -> 10 minutes
	case MiniGameDice:
		return 300.0 // 2-5 min -> 5 minutes
	case MiniGamePuzzle:
		return 420.0 // 3-7 min -> 7 minutes
	case MiniGameMemory:
		return 240.0 // 2-4 min -> 4 minutes
	case MiniGameLockPicking:
		return 120.0 // 0.5-2 min -> 2 minutes
	case MiniGameHacking:
		return 180.0 // 1-3 min -> 3 minutes
	case MiniGameRitual:
		return 300.0 // 2-5 min -> 5 minutes
	default:
		return 300.0 // Default 5 minutes
	}
}

// generateReward creates a reward based on game type and difficulty.
// Uses deterministic RNG for reproducible rewards in multiplayer.
func (s *MiniGameSystem) generateReward(gameType MiniGameType, difficulty float64) *Reward {
	rng := rand.New(s.rngSource)

	// Base rewards scale with difficulty
	baseGold := 50
	baseXP := 100.0

	// Game type multipliers
	multiplier := 1.0
	switch gameType {
	case MiniGameCard:
		multiplier = 1.5 // Longer games have better rewards
	case MiniGameDice:
		multiplier = 1.2
	case MiniGamePuzzle:
		multiplier = 1.3
	case MiniGameMemory:
		multiplier = 1.0
	case MiniGameLockPicking:
		multiplier = 0.8 // Quick games have lower rewards
	case MiniGameHacking:
		multiplier = 1.4
	case MiniGameRitual:
		multiplier = 1.3
	}

	// Add some randomness (±20%)
	randomFactor := 0.8 + rng.Float64()*0.4

	gold := int(float64(baseGold) * difficulty * multiplier * randomFactor * 2.0)
	xp := baseXP * difficulty * multiplier * randomFactor * 1.5

	return &Reward{
		Gold:  gold,
		XP:    xp,
		Items: []uint64{}, // INTEGRATION FIX [Category E]: Mini-Game Reward Items
		// Gap: Mini-game rewards should include procedurally generated items, not just gold/XP
		// Fix: Call ItemGenerator with difficulty-scaled rarity, add 1-3 items to Items array
		// Roadmap: ROADMAP_V4.md Phase 27.1 - Mini-Game Rewards
		// Integration: Requires itemgen.Generator instance, generate based on game type (puzzle=scroll, combat=weapon)
	}
}

// awardReward gives the reward to the entity, updating gold and XP.
func (s *MiniGameSystem) awardReward(entityID uint64, reward *Reward) {
	entity, ok := s.world.GetEntity(entityID)
	if !ok || entity == nil {
		return
	}

	// Award gold
	if invCompRaw, ok := entity.GetComponent("inventory"); ok {
		invComp := invCompRaw.(*InventoryComponent)
		invComp.Gold += reward.Gold
	}

	// Award XP
	if expCompRaw, ok := entity.GetComponent("experience"); ok {
		expComp := expCompRaw.(*ExperienceComponent)
		expComp.CurrentXP += int(reward.XP)
	}

	// Award items - Phase 27.3 Integration
	// Iterate through reward item entity IDs, get the item from each entity,
	// and add it to the player's inventory. Items that cannot be added due to
	// inventory limits remain as world entities for pickup later.
	if len(reward.Items) > 0 {
		invCompRaw, hasInv := entity.GetComponent("inventory")
		if hasInv {
			invComp := invCompRaw.(*InventoryComponent)
			for _, itemEntityID := range reward.Items {
				itemEntity, exists := s.world.GetEntity(itemEntityID)
				if !exists || itemEntity == nil {
					continue
				}
				itemCompRaw, hasItem := itemEntity.GetComponent("item")
				if !hasItem {
					continue
				}
				itemComp, ok := itemCompRaw.(*ItemComponent)
				if !ok || itemComp.Item == nil {
					continue
				}
				// Add the item to player inventory; if inventory is full,
				// the item remains as a world entity for later pickup
				if !invComp.AddItem(itemComp.Item) {
					// Item could not be added (inventory full by count or weight).
					// The item entity remains in the world for the player to pick up later.
					continue
				}
			}
		}
	}
}
