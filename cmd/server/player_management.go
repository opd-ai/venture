//go:build !android && !ios
// +build !android,!ios

// player_management.go contains player entity management and input handling.
// Code relocated from: main.go
package main

import (
	"image/color"

	"github.com/opd-ai/venture/pkg/combat"
	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/logging"
	"github.com/opd-ai/venture/pkg/network"
	itemgen "github.com/opd-ai/venture/pkg/procgen/item"
	"github.com/opd-ai/venture/pkg/procgen/terrain"
	"github.com/opd-ai/venture/pkg/rendering/sprites"
	"github.com/sirupsen/logrus"
)

// createPlayerEntity creates a player entity for a connected client
func createPlayerEntity(world *engine.World, terrain *terrain.Terrain, playerID uint64, seed int64, genreID string, useAerialSprites bool, logger *logrus.Logger) *engine.Entity {
	// Create player entity
	entity := world.CreateEntity()

	// Find valid spawn position in first room
	spawnX, spawnY := 400.0, 300.0 // Default spawn
	if terrain != nil && len(terrain.Rooms) > 0 {
		room := terrain.Rooms[0]
		// Spawn in center of first room
		spawnX = float64(room.X+room.Width/2) * 32 // Convert to pixel coordinates (32px tiles)
		spawnY = float64(room.Y+room.Height/2) * 32
	}

	// Add core components
	entity.AddComponent(&engine.PositionComponent{X: spawnX, Y: spawnY})
	entity.AddComponent(&engine.VelocityComponent{VX: 0, VY: 0})
	entity.AddComponent(&engine.HealthComponent{Current: 100, Max: 100})
	entity.AddComponent(&engine.TeamComponent{TeamID: 1}) // All players on team 1

	// Add network component to mark as networked entity
	entity.AddComponent(&engine.NetworkComponent{
		PlayerID: playerID,
		Synced:   true,
	})

	// Add sprite for rendering (Phase 45: 64×64 enhanced sprites)
	var playerSprite *engine.EbitenSprite
	const playerSpriteSize = 64   // Phase 45 standard sprite size
	const playerColliderOff = -32 // Center offset (64/2 = 32)
	if useAerialSprites {
		// Generate procedural directional sprites with aerial-view perspective
		spriteGen := sprites.NewGenerator()
		config := sprites.Config{
			Width:   playerSpriteSize,
			Height:  playerSpriteSize,
			Seed:    seed + int64(playerID), // Unique seed per player
			GenreID: genreID,
			Type:    sprites.SpriteEntity,
			Custom: map[string]interface{}{
				"entityType": "humanoid",
				"useAerial":  true,
			},
		}

		directionalSprites, err := spriteGen.GenerateDirectionalSprites(config)
		if err != nil {
			logger.WithError(err).Warn("failed to generate directional sprites, using fallback")
			playerSprite = engine.NewSpriteComponent(playerSpriteSize, playerSpriteSize, color.RGBA{100, 150, 255, 255})
		} else {
			// Create sprite component with initial down-facing direction
			// directionalSprites is map[int]*ebiten.Image with keys 0-3
			playerSprite = &engine.EbitenSprite{
				Image:             directionalSprites[int(engine.DirDown)],
				Width:             playerSpriteSize,
				Height:            playerSpriteSize,
				Visible:           true,
				Layer:             10,
				CurrentDirection:  int(engine.DirDown),
				DirectionalImages: directionalSprites, // Already map[int]*ebiten.Image
			}
			// Add animation component to enable automatic facing updates
			entity.AddComponent(&engine.AnimationComponent{
				Seed:         seed + int64(playerID),
				CurrentState: engine.AnimationStateIdle,
				Playing:      true,
			})
		}
	} else {
		// Use simple colored sprite for side-view
		playerSprite = engine.NewSpriteComponent(playerSpriteSize, playerSpriteSize, color.RGBA{100, 150, 255, 255})
	}

	playerSprite.Layer = 10 // Draw players on top
	entity.AddComponent(playerSprite)

	// Add player stats
	playerStats := engine.NewStatsComponent()
	playerStats.Attack = 10
	playerStats.Defense = 5
	entity.AddComponent(playerStats)

	// Add player experience/progression
	playerExp := engine.NewExperienceComponent()
	entity.AddComponent(playerExp)

	// Add player inventory
	playerInventory := engine.NewInventoryComponent(20, 100.0) // 20 items, 100 weight max
	playerInventory.Gold = 100
	entity.AddComponent(playerInventory)

	// Add player equipment
	playerEquipment := engine.NewEquipmentComponent()
	entity.AddComponent(playerEquipment)

	// Add quest tracker
	questTracker := engine.NewQuestTrackerComponent(5) // Max 5 active quests
	entity.AddComponent(questTracker)

	// Add player attack capability
	entity.AddComponent(&engine.AttackComponent{
		Damage:     15,
		DamageType: combat.DamagePhysical,
		Range:      50,
		Cooldown:   0.5,
	})

	// Add collision for player (Phase 45: 64×64 collider)
	entity.AddComponent(&engine.ColliderComponent{
		Width:     playerSpriteSize,
		Height:    playerSpriteSize,
		Solid:     true,
		IsTrigger: false,
		Layer:     1,
		OffsetX:   playerColliderOff, // Center the collider (64/2 = 32)
		OffsetY:   playerColliderOff,
	})

	if logger.GetLevel() >= logrus.DebugLevel {
		logging.EntityLogger(logger, int(entity.ID)).WithFields(logrus.Fields{
			"playerID": playerID,
			"x":        spawnX,
			"y":        spawnY,
		}).Debug("player entity created")
	}

	return entity
}

// applyInputCommand applies a network input command to a player entity
func applyInputCommand(entity *engine.Entity, cmd *network.InputCommand, logger *logrus.Logger) {
	velComp, hasVel := entity.GetComponent("velocity")
	if !hasVel {
		return
	}
	velocity := velComp.(*engine.VelocityComponent)

	switch cmd.InputType {
	case "move":
		applyMovement(velocity, cmd, logger)
	case "attack":
		applyAttack(entity, cmd, logger)
	case "use_item":
		applyItemUse(entity, cmd, logger)
	default:
		logUnknownInput(cmd, logger)
	}
}

// applyMovement processes movement input and updates entity velocity.
func applyMovement(velocity *engine.VelocityComponent, cmd *network.InputCommand, logger *logrus.Logger) {
	if len(cmd.Data) < 2 {
		return
	}

	moveX := float64(int8(cmd.Data[0])) / 127.0
	moveY := float64(int8(cmd.Data[1])) / 127.0

	if moveX != 0 && moveY != 0 {
		moveX *= 0.707
		moveY *= 0.707
	}

	velocity.VX = moveX * 100.0
	velocity.VY = moveY * 100.0

	if logger.GetLevel() >= logrus.DebugLevel && (moveX != 0 || moveY != 0) {
		logging.NetworkLogger(logger, "", "").WithFields(logrus.Fields{
			"playerID":  cmd.PlayerID,
			"velocityX": velocity.VX,
			"velocityY": velocity.VY,
		}).Debug("player moving")
	}
}

// applyAttack processes attack input and triggers entity attack if cooldown permits.
func applyAttack(entity *engine.Entity, cmd *network.InputCommand, logger *logrus.Logger) {
	if logger.GetLevel() >= logrus.DebugLevel {
		logging.NetworkLogger(logger, "", "").WithField("playerID", cmd.PlayerID).Debug("player attacking")
	}

	attackComp, hasAttack := entity.GetComponent("attack")
	if !hasAttack {
		if logger.GetLevel() >= logrus.WarnLevel {
			logging.NetworkLogger(logger, "", "").WithField("playerID", cmd.PlayerID).Warn("player has no attack component")
		}
		return
	}
	attack := attackComp.(*engine.AttackComponent)

	if !attack.CanAttack() {
		if logger.GetLevel() >= logrus.DebugLevel {
			logging.NetworkLogger(logger, "", "").WithFields(logrus.Fields{
				"playerID":       cmd.PlayerID,
				"cooldownRemain": attack.CooldownTimer,
			}).Debug("player attack on cooldown")
		}
		return
	}

	attack.ResetCooldown()

	if logger.GetLevel() >= logrus.DebugLevel {
		logging.NetworkLogger(logger, "", "").WithFields(logrus.Fields{
			"playerID": cmd.PlayerID,
			"damage":   attack.Damage,
			"range":    attack.Range,
		}).Debug("player attack triggered")
	}
}

// applyItemUse processes item usage from inventory and applies effects.
func applyItemUse(entity *engine.Entity, cmd *network.InputCommand, logger *logrus.Logger) {
	if logger.GetLevel() >= logrus.DebugLevel {
		logging.NetworkLogger(logger, "", "").WithField("playerID", cmd.PlayerID).Debug("player using item")
	}

	invComp, hasInv := entity.GetComponent("inventory")
	if !hasInv {
		if logger.GetLevel() >= logrus.WarnLevel {
			logging.NetworkLogger(logger, "", "").WithField("playerID", cmd.PlayerID).Warn("player has no inventory component")
		}
		return
	}
	inventory := invComp.(*engine.InventoryComponent)

	item, itemIndex, ok := validateItemUse(inventory, cmd, logger)
	if !ok {
		return
	}

	consumeItem(entity, inventory, item, itemIndex, cmd.PlayerID, logger)
}

// validateItemUse validates item index and type for usage.
func validateItemUse(inventory *engine.InventoryComponent, cmd *network.InputCommand, logger *logrus.Logger) (*itemgen.Item, int, bool) {
	if len(cmd.Data) < 1 {
		if logger.GetLevel() >= logrus.WarnLevel {
			logging.NetworkLogger(logger, "", "").WithField("playerID", cmd.PlayerID).Warn("use_item command missing item index")
		}
		return nil, 0, false
	}
	itemIndex := int(cmd.Data[0])

	if itemIndex < 0 || itemIndex >= len(inventory.Items) {
		if logger.GetLevel() >= logrus.WarnLevel {
			logging.NetworkLogger(logger, "", "").WithFields(logrus.Fields{
				"playerID":      cmd.PlayerID,
				"itemIndex":     itemIndex,
				"inventorySize": len(inventory.Items),
			}).Warn("invalid item index")
		}
		return nil, 0, false
	}

	item := inventory.Items[itemIndex]
	if item.Type != itemgen.TypeConsumable {
		if logger.GetLevel() >= logrus.WarnLevel {
			logging.NetworkLogger(logger, "", "").WithFields(logrus.Fields{
				"playerID": cmd.PlayerID,
				"itemName": item.Name,
			}).Warn("attempted to use non-consumable item")
		}
		return nil, 0, false
	}

	return item, itemIndex, true
}

// consumeItem applies consumable item effects and removes it from inventory.
func consumeItem(entity *engine.Entity, inventory *engine.InventoryComponent, item *itemgen.Item, itemIndex int, playerID uint64, logger *logrus.Logger) {
	healthComp, hasHealth := entity.GetComponent("health")
	if !hasHealth {
		return
	}
	health := healthComp.(*engine.HealthComponent)

	// G25 fix: use item.Stats.Healing as the authoritative heal amount.
	// item.Stats.Defense is the armor contribution and is 0 for consumables.
	healAmount := float64(item.Stats.Healing)
	if healAmount <= 0 {
		return
	}

	health.Current += healAmount
	if health.Current > health.Max {
		health.Current = health.Max
	}

	if logger.GetLevel() >= logrus.InfoLevel {
		logging.NetworkLogger(logger, "", "").WithFields(logrus.Fields{
			"playerID":      playerID,
			"itemName":      item.Name,
			"healAmount":    healAmount,
			"currentHealth": health.Current,
			"maxHealth":     health.Max,
		}).Info("player used item")
	}

	inventory.Items = append(inventory.Items[:itemIndex], inventory.Items[itemIndex+1:]...)
}

// logUnknownInput logs warning for unknown input types.
func logUnknownInput(cmd *network.InputCommand, logger *logrus.Logger) {
	if logger.GetLevel() >= logrus.WarnLevel {
		logging.NetworkLogger(logger, "", "").WithFields(logrus.Fields{
			"playerID":  cmd.PlayerID,
			"inputType": cmd.InputType,
		}).Warn("unknown input type")
	}
}
