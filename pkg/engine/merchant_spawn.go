// Package engine provides helper functions for spawning merchants in the game world.
// This file bridges procedural generation (pkg/procgen/entity) with the ECS runtime,
// converting MerchantData into engine entities with proper components.
package engine

import (
	"fmt"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/opd-ai/venture/pkg/procgen/dialog"
	procgenEntity "github.com/opd-ai/venture/pkg/procgen/entity"
	"github.com/opd-ai/venture/pkg/procgen/item"
	"github.com/opd-ai/venture/pkg/procgen/terrain"
	"github.com/sirupsen/logrus"
)

// SpawnMerchantFromData converts procedural MerchantData into an engine entity.
// This function creates the entity, adds all required components (position, sprite,
// collider, merchant, dialog), and registers it with the world.
//
// Returns the spawned merchant entity or nil if spawning fails.
func SpawnMerchantFromData(world *World, merchantData *procgenEntity.MerchantData, x, y float64, seed int64, params procgen.GenerationParams) *Entity {
	if merchantData == nil || merchantData.Entity == nil {
		return nil
	}

	// Create merchant entity
	merchant := world.CreateEntity()

	// Add position
	merchant.AddComponent(&PositionComponent{X: x, Y: y})

	// Add velocity (merchants are stationary by default)
	merchant.AddComponent(&VelocityComponent{VX: 0, VY: 0})

	// Add health (merchants are non-combatants)
	merchant.AddComponent(&HealthComponent{
		Current: float64(merchantData.Entity.Stats.Health),
		Max:     float64(merchantData.Entity.Stats.Health),
	})

	// Add team component (neutral team)
	merchant.AddComponent(&TeamComponent{TeamID: 0})

	// Add sprite (distinct from player/enemies) - Phase 45: 64×64 enhanced sprites
	const merchantSpriteSize = 64   // Phase 45 standard sprite size
	const merchantColliderOff = -32 // Center offset (64/2 = 32)
	merchantSprite := &EbitenSprite{
		Image:   ebiten.NewImage(merchantSpriteSize, merchantSpriteSize),
		Width:   merchantSpriteSize,
		Height:  merchantSpriteSize,
		Visible: true,
		Layer:   10, // Same layer as player
	}
	merchant.AddComponent(merchantSprite)

	// Add animation component with unique seed offset for merchants
	merchantAnim := NewAnimationComponent(merchantData.Entity.Seed)
	merchantAnim.CurrentState = AnimationStateIdle
	merchantAnim.FrameTime = 0.2 // Slower animation for NPCs
	merchantAnim.Loop = true
	merchantAnim.Playing = true
	merchantAnim.FrameCount = 8 // V7.0: 8-frame animations
	merchant.AddComponent(merchantAnim)

	// Add collider (merchants are solid NPCs) - Phase 45: 64×64 collider
	merchant.AddComponent(&ColliderComponent{
		Width:     merchantSpriteSize,
		Height:    merchantSpriteSize,
		Solid:     true,
		IsTrigger: false,
		Layer:     1,
		OffsetX:   merchantColliderOff,
		OffsetY:   merchantColliderOff,
	})

	// Convert MerchantType from procgen to engine
	var engineMerchantType MerchantType
	if merchantData.MerchantType == procgenEntity.MerchantFixed {
		engineMerchantType = MerchantFixed
	} else {
		engineMerchantType = MerchantNomadic
	}

	// Add merchant component
	merchantComp := NewMerchantComponent(
		len(merchantData.Inventory), // Use inventory size as max
		engineMerchantType,
		merchantData.PriceMultiplier,
	)
	merchantComp.MerchantName = merchantData.Entity.Name
	merchantComp.BuyBackPercentage = merchantData.BuyBackPercentage

	// Copy inventory items
	merchantComp.Inventory = make([]*item.Item, 0, len(merchantData.Inventory))
	merchantComp.Inventory = append(merchantComp.Inventory, merchantData.Inventory...)

	merchant.AddComponent(merchantComp)

	// Add both dialog components for backward compatibility and advanced features
	// Markov dialog component for varied shop interactions (Phase 3.1)
	merchantPersonalitySimple := dialog.NewPersonality(dialog.PersonalityMerchant)
	dialogProvider := NewMarkovDialogProvider(seed+int64(merchantData.Entity.Seed), params.GenreID, merchantData.Entity.Name, merchantPersonalitySimple)
	dialogComp := NewDialogComponent(dialogProvider)
	merchant.AddComponent(dialogComp)

	// Advanced procedural dialog component for rich NPC interactions
	// Uses genre-specific corpus and merchant personality for varied dialog
	merchantPersonality := dialog.NewPersonality(dialog.PersonalityMerchant)
	npcDialogComp := NewNPCDialogComponent(params.GenreID, merchantPersonality, seed+int64(merchantData.Entity.Seed))
	merchant.AddComponent(npcDialogComp)

	// Add genre component so genre-aware systems (e.g., quest reward scaling) can
	// read the active genre from this entity. Mirrors the same component added to
	// enemy entities in addAdvancedComponents.
	merchant.AddComponent(NewGenreComponent(params.GenreID))

	return merchant
}

// SpawnMerchantsInTerrain generates and spawns merchants in the game world.
// Uses procgen entity generation to create merchants, then converts them to engine entities.
// Merchants spawn in room centers (fixed) or random walkable locations (nomadic).
//
// Parameters:
//   - world: The ECS world to spawn merchants into
//   - terrain: The generated terrain for spawn location validation
//   - worldSeed: Base seed for deterministic merchant generation
//   - params: Generation parameters (difficulty, depth, genre)
//   - merchantCount: Number of merchants to spawn (typically 1-3 per dungeon level)
//
// Returns the number of merchants spawned.
func SpawnMerchantsInTerrain(world *World, terrain *terrain.Terrain, worldSeed int64, params procgen.GenerationParams, merchantCount int) (int, error) {
	if merchantCount <= 0 {
		return 0, nil
	}

	logger := createMerchantLogger(world, worldSeed, params, merchantCount)
	spawnPoints := generateMerchantSpawnPoints(terrain, worldSeed, merchantCount, logger)

	return spawnMerchantsAtPoints(world, terrain, spawnPoints, worldSeed, params, logger), nil
}

// createMerchantLogger creates a logger for merchant spawning operations.
func createMerchantLogger(world *World, worldSeed int64, params procgen.GenerationParams, merchantCount int) *logrus.Entry {
	if world != nil && world.logger != nil {
		return world.logger.WithFields(logrus.Fields{
			"system": "merchant_spawn",
			"seed":   worldSeed,
			"genre":  params.GenreID,
			"count":  merchantCount,
		})
	}
	return nil
}

// generateMerchantSpawnPoints creates spawn point locations for merchants.
func generateMerchantSpawnPoints(terrain *terrain.Terrain, worldSeed int64, merchantCount int, logger *logrus.Entry) []struct{ X, Y float64 } {
	spawnPoints := procgenEntity.GenerateMerchantSpawnPoints(
		worldSeed,
		terrain.Width,
		terrain.Height,
		procgenEntity.MerchantFixed,
		merchantCount,
	)

	if logger != nil {
		logger.WithField("spawnPoints", len(spawnPoints)).Debug("merchant spawn points generated")
	}

	return spawnPoints
}

// spawnMerchantsAtPoints spawns merchant entities at generated spawn points.
func spawnMerchantsAtPoints(world *World, terrain *terrain.Terrain, spawnPoints []struct{ X, Y float64 }, worldSeed int64, params procgen.GenerationParams, logger *logrus.Entry) int {
	spawned := 0
	merchantGen := procgenEntity.NewEntityGenerator()

	for i, point := range spawnPoints {
		merchantData, worldX, worldY, ok := generateMerchantAtPoint(merchantGen, point, i, worldSeed, params, logger)
		if !ok {
			continue
		}

		if !validateMerchantSpawnPosition(terrain, point, logger) {
			continue
		}

		if spawnSingleMerchant(world, merchantData, worldX, worldY, worldSeed, params, logger) {
			spawned++
		}
	}

	if logger != nil {
		logger.WithField("spawned", spawned).Info("merchant spawning complete")
	}

	return spawned
}

// generateMerchantAtPoint generates merchant data and calculates world coordinates.
func generateMerchantAtPoint(merchantGen *procgenEntity.EntityGenerator, point struct{ X, Y float64 }, index int, worldSeed int64, params procgen.GenerationParams, logger *logrus.Entry) (*procgenEntity.MerchantData, float64, float64, bool) {
	merchantSeed := worldSeed + int64(index*1000) + 500
	merchantData, err := merchantGen.GenerateMerchant(merchantSeed, params, procgenEntity.MerchantFixed)
	if err != nil {
		if logger != nil {
			logger.WithError(err).WithField("index", index).Warn("failed to generate merchant")
		}
		return nil, 0, 0, false
	}

	worldX := point.X * 32.0
	worldY := point.Y * 32.0

	return merchantData, worldX, worldY, true
}

// validateMerchantSpawnPosition checks if spawn position is walkable terrain.
func validateMerchantSpawnPosition(terrain *terrain.Terrain, point struct{ X, Y float64 }, logger *logrus.Entry) bool {
	tileX := int(point.X)
	tileY := int(point.Y)

	if !terrain.IsWalkable(tileX, tileY) {
		if logger != nil {
			logger.WithFields(logrus.Fields{
				"x": tileX,
				"y": tileY,
			}).Debug("spawn point not walkable, skipping")
		}
		return false
	}

	return true
}

// spawnSingleMerchant creates and spawns a single merchant entity.
func spawnSingleMerchant(world *World, merchantData *procgenEntity.MerchantData, worldX, worldY float64, worldSeed int64, params procgen.GenerationParams, logger *logrus.Entry) bool {
	merchantEntity := SpawnMerchantFromData(world, merchantData, worldX, worldY, worldSeed, params)

	if merchantEntity == nil {
		if logger != nil {
			logger.Warn("failed to spawn merchant entity")
		}
		return false
	}

	if logger != nil {
		logger.WithFields(logrus.Fields{
			"entityID": merchantEntity.ID,
			"name":     merchantData.Entity.Name,
			"x":        worldX,
			"y":        worldY,
			"items":    len(merchantData.Inventory),
		}).Info("merchant spawned")
	}

	return true
}

// GetNearbyMerchants returns all merchant entities within a specified radius of a position.
// Used for proximity detection to enable player interaction (press S to shop).
//
// Returns a slice of merchant entities and their distances from the position.
func GetNearbyMerchants(world *World, x, y, radius float64) []*Entity {
	if world == nil {
		return nil
	}

	nearby := make([]*Entity, 0)
	radiusSq := radius * radius

	// Iterate all entities
	for _, entity := range world.GetEntities() {
		// Check if entity has merchant component
		if !entity.HasComponent("merchant") {
			continue
		}

		// Get position
		posComp, ok := entity.GetComponent("position")
		if !ok {
			continue
		}
		// Type assert with safety check
		pos, ok := posComp.(*PositionComponent)
		if !ok {
			continue
		}

		// Calculate distance squared (avoid sqrt for performance)
		dx := pos.X - x
		dy := pos.Y - y
		distSq := dx*dx + dy*dy

		if distSq <= radiusSq {
			nearby = append(nearby, entity)
		}
	}

	return nearby
}

// FindClosestMerchant returns the closest merchant to a position within a radius.
// Returns the merchant entity and the distance, or (nil, -1) if none found.
func FindClosestMerchant(world *World, x, y, radius float64) (*Entity, float64) {
	merchants := GetNearbyMerchants(world, x, y, radius)
	if len(merchants) == 0 {
		return nil, -1
	}

	var closest *Entity
	minDistSq := radius * radius

	for _, merchant := range merchants {
		posComp, ok := merchant.GetComponent("position")
		if !ok {
			continue
		}
		// Type assert with safety check
		pos, ok := posComp.(*PositionComponent)
		if !ok {
			continue
		}

		dx := pos.X - x
		dy := pos.Y - y
		distSq := dx*dx + dy*dy

		if distSq < minDistSq {
			minDistSq = distSq
			closest = merchant
		}
	}

	if closest == nil {
		return nil, -1
	}

	// Return actual distance (not squared)
	dist := 0.0
	if minDistSq > 0 {
		dist = math.Sqrt(minDistSq)
	}

	return closest, dist
}

// GetMerchantInteractionPrompt returns UI text to display when near a merchant.
// Format: "Press S to talk to [Merchant Name]"
func GetMerchantInteractionPrompt(merchant *Entity) string {
	if merchant == nil {
		return ""
	}

	merchComp, ok := merchant.GetComponent("merchant")
	if !ok {
		return ""
	}

	// Type assert with safety check
	merchantData, ok := merchComp.(*MerchantComponent)
	if !ok {
		return ""
	}
	return fmt.Sprintf("Press S to talk to %s", merchantData.MerchantName)
}
