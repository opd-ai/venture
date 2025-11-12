package engine

import (
	"time"
)

// MiniGameSystem manages mini-game instances
type MiniGameSystem struct {
	world *World
}

// NewMiniGameSystem creates a new mini-game system
func NewMiniGameSystem(world *World) *MiniGameSystem {
	return &MiniGameSystem{world: world}
}

// Update processes active mini-games
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
		
		// Check for timeout
		if gameComp.TimeLimit > 0 && gameComp.TimeElapsed >= gameComp.TimeLimit {
			s.EndGame(entity.ID, false)
		}
	}
}

// StartGame starts a mini-game
func (s *MiniGameSystem) StartGame(entityID uint64, gameType MiniGameType, difficulty float64) {
	entity, ok := s.world.GetEntity(entityID)
	if !ok || entity == nil {
		return
	}
	
	// Create or update mini-game component
	gameComp := &MiniGameComponent{
		GameType:    gameType,
		Active:      true,
		State:       make(map[string]interface{}),
		Difficulty:  difficulty,
		TimeLimit:   s.getTimeLimit(gameType),
		TimeElapsed: 0,
		Reward:      s.generateReward(gameType, difficulty),
	}
	
	entity.AddComponent(gameComp)
}

// EndGame ends a mini-game
func (s *MiniGameSystem) EndGame(entityID uint64, success bool) {
	entity, ok := s.world.GetEntity(entityID)
	if !ok || entity == nil {
		return
	}
	
	gameCompRaw, ok := entity.GetComponent("minigame")
	if !ok {
		return
	}
	
	gameComp := gameCompRaw.(*MiniGameComponent)
	gameComp.Active = false
	
	// Award rewards if successful
	if success && gameComp.Reward != nil {
		s.awardReward(entityID, gameComp.Reward)
	}
}

func (s *MiniGameSystem) getTimeLimit(gameType MiniGameType) float64 {
	switch gameType {
	case MiniGameCard:
		return 600.0 // 10 minutes
	case MiniGameDice:
		return 300.0 // 5 minutes
	case MiniGamePuzzle:
		return 420.0 // 7 minutes
	case MiniGameMemory:
		return 240.0 // 4 minutes
	case MiniGameLockPicking:
		return 120.0 // 2 minutes
	case MiniGameHacking:
		return 180.0 // 3 minutes
	case MiniGameRitual:
		return 300.0 // 5 minutes
	default:
		return 300.0
	}
}

func (s *MiniGameSystem) generateReward(gameType MiniGameType, difficulty float64) *Reward {
	baseGold := 50
	baseXP := 100.0
	
	return &Reward{
		Gold:  int(float64(baseGold) * difficulty * 2.0),
		XP:    baseXP * difficulty * 1.5,
		Items: []uint64{}, // TODO: Generate items
	}
}

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
	
	// Award items
	// TODO: Add items to inventory
	_ = time.Now() // Placeholder to avoid unused import
}
