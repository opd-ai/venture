package engine

import (
	"fmt"

	"github.com/opd-ai/venture/pkg/world/territory"
	"github.com/sirupsen/logrus"
)

// TerritorySystem manages territory capture, guild warfare, and defensive structures.
type TerritorySystem struct {
	manager        *territory.Manager
	logger         *logrus.Entry
	updateInterval float64
	timeAccum      float64
}

// NewTerritorySystem creates a new territory management system.
func NewTerritorySystem(manager *territory.Manager, logger *logrus.Entry) *TerritorySystem {
	if logger == nil {
		logger = logrus.WithField("system", "territory")
	}
	return &TerritorySystem{
		manager:        manager,
		logger:         logger,
		updateInterval: 1.0, // Update every second
		timeAccum:      0.0,
	}
}

// Update processes territory capture progress for entities in territories.
func (ts *TerritorySystem) Update(entities []*Entity, deltaTime float64) {
	ts.timeAccum += deltaTime
	if ts.timeAccum < ts.updateInterval {
		return
	}
	ts.timeAccum = 0.0

	territoryPresence := ts.countTerritoryPresence(entities)
	ts.processTerritoryCombat(territoryPresence)
}

func (ts *TerritorySystem) countTerritoryPresence(entities []*Entity) map[string]map[string]int {
	territoryPresence := make(map[string]map[string]int)

	for _, entity := range entities {
		guildID := ts.extractEntityGuildID(entity)
		if guildID == "" {
			continue
		}

		territoryID := ts.extractEntityTerritoryID(entity)
		if territoryID == "" {
			continue
		}

		if territoryPresence[territoryID] == nil {
			territoryPresence[territoryID] = make(map[string]int)
		}
		territoryPresence[territoryID][guildID]++
	}

	return territoryPresence
}

func (ts *TerritorySystem) extractEntityGuildID(entity *Entity) string {
	_, hasPos := entity.GetComponent("position")
	guildComp, hasGuild := entity.GetComponent("guild")

	if !hasPos || !hasGuild {
		return ""
	}

	guild := guildComp.(*GuildComponent)
	return guild.GuildID
}

func (ts *TerritorySystem) extractEntityTerritoryID(entity *Entity) string {
	posComp, hasPos := entity.GetComponent("position")
	if !hasPos {
		return ""
	}

	pos := posComp.(*PositionComponent)
	return ts.getTerritoryIDFromPosition(pos.X, pos.Y)
}

func (ts *TerritorySystem) processTerritoryCombat(territoryPresence map[string]map[string]int) {
	for territoryID, presence := range territoryPresence {
		terr, err := ts.manager.GetTerritory(territoryID)
		if err != nil {
			continue
		}

		ownerCount, attackerGuild, attackerCount := ts.analyzeTerritoryForces(terr, presence)
		ts.updateTerritoryCapture(territoryID, terr, attackerGuild, attackerCount, ownerCount)
	}
}

func (ts *TerritorySystem) analyzeTerritoryForces(terr *territory.Territory, presence map[string]int) (int, string, int) {
	ownerCount := 0
	if terr.OwnerGuildID != "" {
		ownerCount = presence[terr.OwnerGuildID]
	}

	attackerGuild := ""
	attackerCount := 0
	for guildID, count := range presence {
		if guildID != terr.OwnerGuildID && count > attackerCount {
			attackerGuild = guildID
			attackerCount = count
		}
	}

	return ownerCount, attackerGuild, attackerCount
}

func (ts *TerritorySystem) updateTerritoryCapture(territoryID string, terr *territory.Territory, attackerGuild string, attackerCount, ownerCount int) {
	if attackerCount == 0 {
		return
	}

	err := ts.manager.UpdateCaptureProgress(territoryID, attackerCount, ownerCount, attackerGuild)
	if err != nil {
		ts.logger.WithError(err).Warn("failed to update capture progress")
		return
	}

	ts.logCaptureProgress(territoryID, attackerGuild, attackerCount, ownerCount)
}

func (ts *TerritorySystem) logCaptureProgress(territoryID, attackerGuild string, attackerCount, ownerCount int) {
	updatedTerr, _ := ts.manager.GetTerritory(territoryID)
	if updatedTerr != nil && updatedTerr.Status == territory.StatusContested {
		ts.logger.WithFields(logrus.Fields{
			"territory":       territoryID,
			"attackers":       attackerCount,
			"defenders":       ownerCount,
			"progress":        updatedTerr.CaptureProgress,
			"attacking_guild": attackerGuild,
		}).Debug("territory contested")
	}
}

// getTerritoryIDFromPosition returns the territory ID for a given position.
// Territories are 5x5 chunk zones (assuming chunk size of 100 units).
func (ts *TerritorySystem) getTerritoryIDFromPosition(x, y float64) string {
	chunkSize := 100.0
	territoryChunks := 5
	territorySize := chunkSize * float64(territoryChunks)

	// Use floor division to handle negative coordinates correctly
	territoryX := int(x / territorySize)
	if x < 0 && x != float64(territoryX)*territorySize {
		territoryX--
	}
	territoryY := int(y / territorySize)
	if y < 0 && y != float64(territoryY)*territorySize {
		territoryY--
	}

	return fmt.Sprintf("territory_%d_%d", territoryX, territoryY)
}

// GetManager returns the underlying territory manager.
func (ts *TerritorySystem) GetManager() *territory.Manager {
	return ts.manager
}

// ProcessCombatDamage applies damage to defensive structures.
func (ts *TerritorySystem) ProcessCombatDamage(territoryID, structureID string, damage float64) error {
	return ts.manager.DamageStructure(territoryID, structureID, damage)
}

// GetTerritoryAtPosition returns the territory at a given position if it exists.
func (ts *TerritorySystem) GetTerritoryAtPosition(x, y float64) (*territory.Territory, error) {
	territoryID := ts.getTerritoryIDFromPosition(x, y)
	return ts.manager.GetTerritory(territoryID)
}

// EnsureTerritoryExists creates a territory at the given position if it doesn't exist.
func (ts *TerritorySystem) EnsureTerritoryExists(x, y float64) (*territory.Territory, error) {
	territoryID := ts.getTerritoryIDFromPosition(x, y)

	terr, err := ts.manager.GetTerritory(territoryID)
	if err == nil {
		return terr, nil
	}

	// Territory doesn't exist, create it
	chunkSize := 100.0
	territoryChunks := 5
	territorySize := chunkSize * float64(territoryChunks)

	territoryX := int(x / territorySize)
	territoryY := int(y / territorySize)

	coords := territory.TerritoryCoords{
		ChunkX: territoryX,
		ChunkZ: territoryY,
	}

	terr, err = ts.manager.CreateTerritory(territoryID, coords)
	if err != nil {
		return nil, fmt.Errorf("failed to create territory: %w", err)
	}

	ts.logger.WithFields(logrus.Fields{
		"territory_id": territoryID,
		"coords_x":     territoryX,
		"coords_z":     territoryY,
	}).Info("created new territory")

	return terr, nil
}

// GetBonusesForGuild returns the resource and XP bonuses for a guild.
// Implements TerritoryBonusProvider interface for HUD display.
func (ts *TerritorySystem) GetBonusesForGuild(guildID string) (resourceBonus, xpBonus float64) {
	if guildID == "" {
		return 0, 0
	}
	return ts.manager.GetBonusesForGuild(guildID)
}

// Compile-time check that TerritorySystem implements TerritoryBonusProvider
var _ TerritoryBonusProvider = (*TerritorySystem)(nil)
