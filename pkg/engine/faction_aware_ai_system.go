// Package engine provides the FactionAwareAISystem which bridges faction reputation
// and AI behavior. This system updates entity team relationships based on faction
// standing, making neutral NPCs become hostile when reputation is low enough.
package engine

import (
	"github.com/sirupsen/logrus"
)

// FactionAwareAISystem synchronizes faction reputation with AI team assignments.
// When a player's reputation with a faction drops to hostile (-50 or below),
// NPCs of that faction will treat the player as an enemy and attack on sight.
//
// This connects the FactionSystem (reputation tracking) with the AISystem
// (enemy detection via TeamComponent) without modifying either core system.
type FactionAwareAISystem struct {
	world  *World
	logger *logrus.Entry

	// updateInterval controls how often we check faction standings (in seconds)
	updateInterval float64
	timeSinceCheck float64

	// cachedPlayerID avoids repeated player lookups
	cachedPlayerID uint64
	playerValid    bool
}

// NewFactionAwareAISystem creates a new faction-aware AI system.
func NewFactionAwareAISystem(world *World) *FactionAwareAISystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "faction_aware_ai")
		logEntry.Debug("FactionAwareAISystem created")
	}

	return &FactionAwareAISystem{
		world:          world,
		logger:         logEntry,
		updateInterval: 1.0, // Check every 1 second (faction changes are infrequent)
	}
}

// Update checks faction standings and updates team assignments.
// This runs at a reduced rate (once per second) since faction reputation
// changes are relatively rare events.
func (f *FactionAwareAISystem) Update(entities []*Entity, deltaTime float64) {
	f.timeSinceCheck += deltaTime

	// Only check once per interval to minimize performance impact
	if f.timeSinceCheck < f.updateInterval {
		return
	}
	f.timeSinceCheck = 0

	// Find player entity
	player := f.findPlayer(entities)
	if player == nil {
		return
	}

	// Get player's faction standings
	playerFactions := f.getPlayerFactionStandings(player)
	if len(playerFactions) == 0 {
		return
	}

	// Check each NPC for faction-based hostility
	for _, entity := range entities {
		if entity == player {
			continue
		}

		f.updateEntityHostility(entity, player, playerFactions)
	}
}

// findPlayer locates the player entity efficiently.
func (f *FactionAwareAISystem) findPlayer(entities []*Entity) *Entity {
	// Use cached ID if valid
	if f.playerValid && f.cachedPlayerID != 0 {
		if entity, ok := f.world.GetEntity(f.cachedPlayerID); ok {
			return entity
		}
		f.playerValid = false
	}

	// Search for player by input component
	for _, entity := range entities {
		if entity.HasComponent("input") {
			f.cachedPlayerID = entity.ID
			f.playerValid = true
			return entity
		}
	}

	return nil
}

// getPlayerFactionStandings retrieves all faction components from the player.
// Returns a map of factionID -> reputation.
func (f *FactionAwareAISystem) getPlayerFactionStandings(player *Entity) map[string]int {
	standings := make(map[string]int)

	// Iterate through components to find faction entries
	for _, comp := range player.Components {
		if factionComp, ok := comp.(*FactionComponent); ok {
			if factionComp.IsPlayerFaction {
				standings[factionComp.FactionID] = factionComp.Reputation
			}
		}
		// Also handle non-pointer faction components
		if factionComp, ok := comp.(FactionComponent); ok {
			if factionComp.IsPlayerFaction {
				standings[factionComp.FactionID] = factionComp.Reputation
			}
		}
	}

	return standings
}

// updateEntityHostility checks if an NPC should become hostile based on faction.
func (f *FactionAwareAISystem) updateEntityHostility(
	entity *Entity,
	player *Entity,
	playerFactions map[string]int,
) {
	// Skip entities without AI or faction
	if !entity.HasComponent("ai") {
		return
	}

	// Get entity's faction membership
	factionComp := f.getEntityFaction(entity)
	if factionComp == nil {
		return
	}

	// Check player's reputation with this entity's faction
	reputation, hasRelation := playerFactions[factionComp.FactionID]
	if !hasRelation {
		return
	}

	// Get or create team component
	team := entity.GetTeam()
	originalTeamID := 0
	if team != nil {
		originalTeamID = team.TeamID
	}

	// Determine new team based on reputation
	newTeamID := f.calculateTeamFromReputation(reputation, originalTeamID)

	// Update team if changed
	if team == nil && newTeamID != 0 {
		// Add team component if hostile
		entity.AddComponent(&TeamComponent{TeamID: newTeamID})
		f.logHostilityChange(entity.ID, factionComp.FactionID, reputation, 0, newTeamID)
	} else if team != nil && team.TeamID != newTeamID {
		// Update existing team
		oldTeamID := team.TeamID
		team.TeamID = newTeamID
		f.logHostilityChange(entity.ID, factionComp.FactionID, reputation, oldTeamID, newTeamID)
	}
}

// getEntityFaction retrieves the faction component from an NPC (non-player faction).
func (f *FactionAwareAISystem) getEntityFaction(entity *Entity) *FactionComponent {
	comp, ok := entity.GetComponent("faction")
	if !ok {
		return nil
	}

	// Handle pointer type
	if fc, ok := comp.(*FactionComponent); ok {
		if !fc.IsPlayerFaction {
			return fc
		}
	}

	// Handle value type
	if fc, ok := comp.(FactionComponent); ok {
		if !fc.IsPlayerFaction {
			return &fc
		}
	}

	return nil
}

// calculateTeamFromReputation determines team ID based on reputation level.
// Team 0 = neutral, Team 1 = player, Team 2 = enemy
func (f *FactionAwareAISystem) calculateTeamFromReputation(reputation, currentTeamID int) int {
	// Hostile reputation makes NPC an enemy (Team 2)
	if reputation <= -50 {
		return 2 // Enemy team
	}

	// If reputation improved from hostile, return to neutral
	if currentTeamID == 2 && reputation > -50 {
		return 0 // Neutral team
	}

	// Keep current team for non-hostile reputations
	return currentTeamID
}

// logHostilityChange logs when an entity's hostility status changes.
func (f *FactionAwareAISystem) logHostilityChange(
	entityID uint64,
	factionID string,
	reputation int,
	oldTeamID, newTeamID int,
) {
	if f.logger == nil {
		return
	}

	action := "became hostile"
	if newTeamID == 0 {
		action = "became neutral"
	}

	f.logger.WithFields(logrus.Fields{
		"entity_id":   entityID,
		"faction_id":  factionID,
		"reputation":  reputation,
		"old_team_id": oldTeamID,
		"new_team_id": newTeamID,
		"action":      action,
	}).Info("Entity faction-based hostility changed")
}

// SetUpdateInterval configures how often faction standings are checked.
// Default is 1.0 seconds. Lower values increase responsiveness but cost more CPU.
func (f *FactionAwareAISystem) SetUpdateInterval(seconds float64) {
	if seconds > 0 {
		f.updateInterval = seconds
	}
}
