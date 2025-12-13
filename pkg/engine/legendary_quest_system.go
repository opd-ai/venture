package engine

import (
	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/opd-ai/venture/pkg/procgen/legendary"
	"github.com/opd-ai/venture/pkg/world/raids"
	"github.com/sirupsen/logrus"
)

// LegendaryQuestSystem manages legendary quests, player progress, and cross-server validation.
// It integrates with the raid system for raid-based quest phases.
type LegendaryQuestSystem struct {
	world        *World
	questManager *legendary.QuestManager
	worldSeed    int64
	logger       *logrus.Entry
}

// NewLegendaryQuestSystem creates a new legendary quest system.
func NewLegendaryQuestSystem(world *World, worldSeed int64, raidManager *raids.Manager) *LegendaryQuestSystem {
	logger := logrus.WithField("system", "legendary_quest")
	return &LegendaryQuestSystem{
		world:        world,
		questManager: legendary.NewQuestManager(raidManager),
		worldSeed:    worldSeed,
		logger:       logger,
	}
}

// Update processes legendary quest progression and validation.
func (s *LegendaryQuestSystem) Update(entities []*Entity, deltaTime float64) {
	// Process entities with legendary quest components
	for _, entity := range entities {
		if !entity.HasComponent("legendary_quest") {
			continue
		}

		comp, ok := entity.GetComponent("legendary_quest")
		if !ok {
			continue
		}
		questComp := comp.(*LegendaryQuestComponent)
		s.updateQuestProgress(entity, questComp)
	}
}

// updateQuestProgress checks and updates quest phase completion.
func (s *LegendaryQuestSystem) updateQuestProgress(entity *Entity, questComp *LegendaryQuestComponent) {
	// Get player ID from entity
	playerID := getPlayerID(entity)
	if playerID == "" {
		return
	}

	// Get player progress
	progress := s.questManager.GetPlayerProgress(playerID, questComp.QuestID)
	if progress == nil {
		return
	}

	// Update quest component state from progress
	questComp.CurrentPhase = progress.CurrentPhase
}

// StartQuest initiates a legendary quest for a player entity.
func (s *LegendaryQuestSystem) StartQuest(playerEntity *Entity, difficulty float64, depth int, genreID string) error {
	playerID := getPlayerID(playerEntity)

	params := procgen.GenerationParams{
		Difficulty: difficulty,
		Depth:      depth,
		GenreID:    genreID,
		Custom: map[string]interface{}{
			"player_level": depth,
		},
	}

	quest, err := s.questManager.GenerateQuest(playerID, s.worldSeed, params)
	if err != nil {
		return err
	}

	// Add legendary quest component to player
	questComp := &LegendaryQuestComponent{
		QuestID:         quest.ID,
		CurrentPhase:    0,
		PhasesCompleted: make([]bool, len(quest.Phases)),
		StartedAt:       0, // Set by component
	}
	playerEntity.AddComponent(questComp)

	s.logger.WithFields(logrus.Fields{
		"player_id": playerID,
		"quest_id":  quest.ID,
		"phases":    len(quest.Phases),
	}).Info("legendary quest started")

	return nil
}

// CompletePhase marks a quest phase as completed.
func (s *LegendaryQuestSystem) CompletePhase(playerEntity *Entity, phaseIndex int) error {
	if !playerEntity.HasComponent("legendary_quest") {
		return ErrNoActiveQuest
	}

	comp, ok := playerEntity.GetComponent("legendary_quest")
	if !ok {
		return ErrNoActiveQuest
	}
	questComp := comp.(*LegendaryQuestComponent)
	playerID := getPlayerID(playerEntity)

	// Update phase progress to 100%
	err := s.questManager.UpdatePhaseProgress(playerID, questComp.QuestID, phaseIndex, 1.0)
	if err != nil {
		return err
	}

	questComp.CurrentPhase = phaseIndex + 1
	questComp.PhasesCompleted[phaseIndex] = true

	s.logger.WithFields(logrus.Fields{
		"player_id": playerID,
		"quest_id":  questComp.QuestID,
		"phase":     phaseIndex,
	}).Info("legendary quest phase completed")

	return nil
}

// GrantReward awards the legendary quest reward to a player.
func (s *LegendaryQuestSystem) GrantReward(playerEntity *Entity) error {
	if !playerEntity.HasComponent("legendary_quest") {
		return ErrNoActiveQuest
	}

	comp, ok := playerEntity.GetComponent("legendary_quest")
	if !ok {
		return ErrNoActiveQuest
	}
	questComp := comp.(*LegendaryQuestComponent)
	playerID := getPlayerID(playerEntity)

	rewards, err := s.questManager.CompleteQuest(playerID, questComp.QuestID)
	if err != nil {
		return err
	}

	// Create reward items in world (items are item IDs)
	for _, itemID := range rewards.Items {
		s.createRewardItemByID(playerEntity, itemID)
	}

	// Remove quest component (quest complete)
	playerEntity.RemoveComponent("legendary_quest")

	s.logger.WithFields(logrus.Fields{
		"player_id": playerID,
		"quest_id":  questComp.QuestID,
		"items":     len(rewards.Items),
		"gold":      rewards.Gold,
		"titles":    len(rewards.Titles),
	}).Info("legendary quest rewards granted")

	return nil
}

// createRewardItemByID spawns a legendary reward item by ID in the world.
func (s *LegendaryQuestSystem) createRewardItemByID(playerEntity *Entity, itemID string) {
	// Get player position
	comp, ok := playerEntity.GetComponent("position")
	if !ok {
		return
	}
	posComp := comp.(*PositionComponent)

	// Create reward entity
	item := s.world.CreateEntity()
	item.AddComponent(&PositionComponent{X: posComp.X, Y: posComp.Y})
	item.AddComponent(&LegendaryItemComponent{
		RewardID:    itemID,
		Name:        "Legendary Item",
		Description: "A legendary reward",
		Stats:       make(map[string]int),
		Rarity:      3.0,
		Unique:      true,
	})
	item.AddComponent(&EbitenSprite{
		Width:   32,
		Height:  32,
		Visible: true,
		Layer:   3,
	})
}

// getPlayerID extracts player ID from entity.
func getPlayerID(entity *Entity) string {
	// Use entity ID as player ID
	return string(rune(entity.ID))
}
