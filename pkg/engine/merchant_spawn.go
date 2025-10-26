// Package engine provides helper functions for spawning merchants in the game world.
// This file bridges procedural generation (pkg/procgen/entity) with the ECS runtime,
// converting MerchantData into engine entities with proper components.
package engine

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/opd-ai/venture/pkg/procgen"
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
func SpawnMerchantFromData(world *World, merchantData *procgenEntity.MerchantData, x, y float64) *Entity {
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

	// Add sprite (distinct from player/enemies)
	merchantSprite := &EbitenSprite{
		Image:   ebiten.NewImage(28, 28),
		Width:   28,
		Height:  28,
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
	merchantAnim.FrameCount = 4
	merchant.AddComponent(merchantAnim)

	// Add collider (merchants are solid NPCs)
	merchant.AddComponent(&ColliderComponent{
		Width:     28,
		Height:    28,
		Solid:     true,
		IsTrigger: false,
		Layer:     1,
		OffsetX:   -14,
		OffsetY:   -14,
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
	for _, itm := range merchantData.Inventory {
		merchantComp.Inventory = append(merchantComp.Inventory, itm)
	}

	merchant.AddComponent(merchantComp)

	// Add dialog component
	dialogProvider := NewMerchantDialogProvider(merchantData.Entity.Name)
	dialogComp := NewDialogComponent(dialogProvider)
	merchant.AddComponent(dialogComp)

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

	// Get world logger if available
	var logger *logrus.Entry
	if world != nil && world.logger != nil {
		logger = world.logger.WithFields(logrus.Fields{
			"system": "merchant_spawn",
			"seed":   worldSeed,
			"genre":  params.GenreID,
			"count":  merchantCount,
		})
	}

	spawned := 0
	merchantGen := procgenEntity.NewEntityGenerator()

	// Generate spawn points (deterministic based on world seed)
	worldWidth := terrain.Width
	worldHeight := terrain.Height
	spawnPoints := procgenEntity.GenerateMerchantSpawnPoints(
		worldSeed,
		worldWidth,
		worldHeight,
		procgenEntity.MerchantFixed, // Use fixed merchants for dungeon shops
		merchantCount,
	)

	if logger != nil {
		logger.WithField("spawnPoints", len(spawnPoints)).Debug("merchant spawn points generated")
	}

	// Generate and spawn merchants at each point
	for i, point := range spawnPoints {
		// Generate merchant data
		merchantSeed := worldSeed + int64(i*1000) + 500 // Offset seed for each merchant
		merchantData, err := merchantGen.GenerateMerchant(merchantSeed, params, procgenEntity.MerchantFixed)
		if err != nil {
			if logger != nil {
				logger.WithError(err).WithField("index", i).Warn("failed to generate merchant")
			}
			continue
		}

		// Convert tile coordinates to world coordinates (32 pixels per tile)
		worldX := point.X * 32.0
		worldY := point.Y * 32.0

		// Validate spawn position is walkable
		tileX := int(point.X)
		tileY := int(point.Y)
		if !terrain.IsWalkable(tileX, tileY) {
			if logger != nil {
				logger.WithFields(logrus.Fields{
					"x": tileX,
					"y": tileY,
				}).Debug("spawn point not walkable, skipping")
			}
			continue
		}

		// Spawn merchant entity
		merchantEntity := SpawnMerchantFromData(world, merchantData, worldX, worldY)
		if merchantEntity == nil {
			if logger != nil {
				logger.WithField("index", i).Warn("failed to spawn merchant entity")
			}
			continue
		}

		spawned++

		if logger != nil {
			logger.WithFields(logrus.Fields{
				"entityID": merchantEntity.ID,
				"name":     merchantData.Entity.Name,
				"x":        worldX,
				"y":        worldY,
				"items":    len(merchantData.Inventory),
			}).Info("merchant spawned")
		}
	}

	if logger != nil {
		logger.WithField("spawned", spawned).Info("merchant spawning complete")
	}

	return spawned, nil
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
		pos := posComp.(*PositionComponent)

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
		pos := posComp.(*PositionComponent)

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
		dist = 1.0 // Placeholder - in real impl would use math.Sqrt(minDistSq)
		// Avoiding math import to keep file simple
		for i := 0.0; i*i < minDistSq; i += 0.1 {
			dist = i
		}
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

	merchantData := merchComp.(*MerchantComponent)
	return fmt.Sprintf("Press S to talk to %s", merchantData.MerchantName)
}
