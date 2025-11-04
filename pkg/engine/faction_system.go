package engine

import (
	"github.com/sirupsen/logrus"
)

// FactionSystem manages faction relationships and reputation changes
type FactionSystem struct {
	world  *World
	logger *logrus.Logger

	// Factions maps faction IDs to faction data
	Factions map[string]*Faction

	// PendingReputationChanges stores reputation changes to apply
	PendingReputationChanges []ReputationChange
}

// NewFactionSystem creates a new faction system
func NewFactionSystem(world *World, logger *logrus.Logger) *FactionSystem {
	if logger == nil {
		logger = logrus.New()
	}

	return &FactionSystem{
		world:                    world,
		logger:                   logger,
		Factions:                 make(map[string]*Faction),
		PendingReputationChanges: make([]ReputationChange, 0),
	}
}

// Update processes pending reputation changes
func (fs *FactionSystem) Update(entities []*Entity, deltaTime float64) {
	if len(fs.PendingReputationChanges) == 0 {
		return
	}

	// Process all pending reputation changes
	for _, change := range fs.PendingReputationChanges {
		fs.applyReputationChange(change)
	}

	// Clear processed changes
	fs.PendingReputationChanges = fs.PendingReputationChanges[:0]
}

// AddFaction adds a faction to the system
func (fs *FactionSystem) AddFaction(faction *Faction) {
	if faction == nil {
		fs.logger.Warn("Attempted to add nil faction")
		return
	}

	fs.Factions[faction.ID] = faction
	fs.logger.WithFields(logrus.Fields{
		"factionID":   faction.ID,
		"factionName": faction.Name,
		"factionType": faction.Type,
	}).Info("Faction added to system")
}

// GetFaction retrieves a faction by ID
func (fs *FactionSystem) GetFaction(factionID string) *Faction {
	return fs.Factions[factionID]
}

// QueueReputationChange adds a reputation change to the pending queue
func (fs *FactionSystem) QueueReputationChange(change ReputationChange) {
	fs.PendingReputationChanges = append(fs.PendingReputationChanges, change)

	fs.logger.WithFields(logrus.Fields{
		"factionID": change.FactionID,
		"amount":    change.Amount,
		"reason":    change.Reason,
	}).Debug("Reputation change queued")
}

// applyReputationChange applies a reputation change to the player
func (fs *FactionSystem) applyReputationChange(change ReputationChange) {
	// Find the player entity
	playerEntity := fs.findPlayerEntity()
	if playerEntity == nil {
		fs.logger.Warn("Cannot apply reputation change: player entity not found")
		return
	}

	// Find or create player's faction component for this faction
	var factionComp *FactionComponent
	if comp, ok := playerEntity.GetComponent("faction"); ok {
		// Player has faction components, find the right one
		// Note: In the ECS, we'll need to handle multiple faction components
		// For now, we assume one component per faction via a different approach
		fc := comp.(FactionComponent)
		if fc.FactionID == change.FactionID && fc.IsPlayerFaction {
			factionComp = &fc
		}
	}

	if factionComp == nil {
		// Create new faction component for this faction
		factionComp = &FactionComponent{
			FactionID:       change.FactionID,
			Reputation:      0,
			IsPlayerFaction: true,
		}
	}

	// Apply reputation change with bounds checking
	newReputation := factionComp.Reputation + change.Amount
	if newReputation > 100 {
		newReputation = 100
	} else if newReputation < -100 {
		newReputation = -100
	}

	oldLevel := factionComp.GetReputationLevel()
	factionComp.Reputation = newReputation
	newLevel := factionComp.GetReputationLevel()

	// Update player entity
	playerEntity.AddComponent(*factionComp)

	// Log reputation change
	fs.logger.WithFields(logrus.Fields{
		"factionID":     change.FactionID,
		"oldReputation": factionComp.Reputation - change.Amount,
		"newReputation": factionComp.Reputation,
		"oldLevel":      oldLevel,
		"newLevel":      newLevel,
		"reason":        change.Reason,
	}).Info("Reputation changed")

	// Trigger level change event if level changed
	if oldLevel != newLevel {
		fs.onReputationLevelChanged(change.FactionID, oldLevel, newLevel)
	}
}

// findPlayerEntity finds the player entity in the world
func (fs *FactionSystem) findPlayerEntity() *Entity {
	// Look for entity with input component (marks player entity)
	entities := fs.world.GetEntitiesWith("input")
	if len(entities) > 0 {
		return entities[0]
	}
	return nil
}

// onReputationLevelChanged handles reputation level transitions
func (fs *FactionSystem) onReputationLevelChanged(factionID, oldLevel, newLevel string) {
	faction := fs.GetFaction(factionID)
	if faction == nil {
		return
	}

	fs.logger.WithFields(logrus.Fields{
		"factionID":   factionID,
		"factionName": faction.Name,
		"oldLevel":    oldLevel,
		"newLevel":    newLevel,
	}).Info("Faction reputation level changed")

	// Future: Trigger UI notifications, quest unlocks, etc.
}

// GetPlayerReputation gets the player's reputation with a faction
func (fs *FactionSystem) GetPlayerReputation(factionID string) int {
	playerEntity := fs.findPlayerEntity()
	if playerEntity == nil {
		return 0
	}

	if comp, ok := playerEntity.GetComponent("faction"); ok {
		fc := comp.(FactionComponent)
		if fc.FactionID == factionID && fc.IsPlayerFaction {
			return fc.Reputation
		}
	}

	return 0 // Default neutral starting reputation
}

// CanTrade checks if player can trade with a faction member
func (fs *FactionSystem) CanTrade(factionID string) bool {
	reputation := fs.GetPlayerReputation(factionID)
	return reputation > -50 // Not hostile
}

// GetTradeDiscount gets the trade discount/markup for a faction
func (fs *FactionSystem) GetTradeDiscount(factionID string) float64 {
	reputation := fs.GetPlayerReputation(factionID)
	fc := FactionComponent{Reputation: reputation}
	return fc.GetPriceMultiplier()
}

// ShouldAttackPlayer checks if faction members should attack player on sight
func (fs *FactionSystem) ShouldAttackPlayer(factionID string) bool {
	reputation := fs.GetPlayerReputation(factionID)
	return reputation <= -50 // Hostile threshold
}

// UpdateNPCHostility updates an NPC's hostility based on faction reputation
func (fs *FactionSystem) UpdateNPCHostility(entity *Entity) {
	// Get NPC's faction
	if comp, ok := entity.GetComponent("faction"); ok {
		fc := comp.(FactionComponent)
		if !fc.IsPlayerFaction {
			// This is an NPC faction member
			if fs.ShouldAttackPlayer(fc.FactionID) {
				// Set NPC as hostile
				if _, ok := entity.GetComponent("ai"); ok {
					// Future: Update AI state to hostile
					fs.logger.WithFields(logrus.Fields{
						"entityID":  entity.ID,
						"factionID": fc.FactionID,
					}).Debug("NPC set to hostile due to faction reputation")
				}
			}
		}
	}
}

// ProcessKillReputation handles reputation changes when a faction member is killed
func (fs *FactionSystem) ProcessKillReputation(killerEntity, victimEntity *Entity) {
	// Get victim's faction
	comp, ok := victimEntity.GetComponent("faction")
	if !ok {
		return // Victim has no faction
	}

	victimFaction := comp.(FactionComponent)
	if victimFaction.IsPlayerFaction {
		return // Don't process if victim is player
	}

	// Check if killer is player
	if _, ok := killerEntity.GetComponent("input"); ok {
		// Player killed faction member - decrease reputation
		fs.QueueReputationChange(ReputationChange{
			FactionID: victimFaction.FactionID,
			Amount:    ReputationKillMember,
			Reason:    "Killed faction member",
		})

		// Check relationships - killing enemy of allied factions increases reputation
		for factionID, faction := range fs.Factions {
			if faction.IsEnemy(victimFaction.FactionID) {
				// This faction is enemy of victim's faction
				playerRep := fs.GetPlayerReputation(factionID)
				if playerRep >= 0 { // Only if not hostile with this faction
					fs.QueueReputationChange(ReputationChange{
						FactionID: factionID,
						Amount:    ReputationKillEnemy,
						Reason:    "Killed enemy faction member",
					})
				}
			} else if faction.IsAlly(victimFaction.FactionID) {
				// This faction is ally of victim's faction
				fs.QueueReputationChange(ReputationChange{
					FactionID: factionID,
					Amount:    ReputationKillAlly,
					Reason:    "Killed allied faction member",
				})
			}
		}
	}
}
