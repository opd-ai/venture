// Package engine provides entity spawning utilities for procedural content integration.
// This file implements functions to spawn procedurally generated entities into the game world.
package engine

import (
	"fmt"
	"image/color"
	"math/rand"

	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/opd-ai/venture/pkg/procgen/entity"
	"github.com/opd-ai/venture/pkg/procgen/terrain"
)

// SpawnEnemiesInTerrain spawns procedurally generated enemies into terrain rooms.
// It generates entities using the entity generator and places them at room centers.
// Returns the number of enemies spawned.
func SpawnEnemiesInTerrain(world *World, terr *terrain.Terrain, seed int64, params procgen.GenerationParams) (int, error) {
	if terr == nil {
		return 0, fmt.Errorf("terrain cannot be nil")
	}

	if len(terr.Rooms) == 0 {
		return 0, nil // No rooms to spawn in
	}

	// Skip first room (player spawn room)
	spawnRooms := terr.Rooms
	if len(spawnRooms) > 1 {
		spawnRooms = spawnRooms[1:]
	}

	// Generate entities for rooms
	entityGen := entity.NewEntityGenerator()

	// Set count based on number of rooms (1-3 enemies per room)
	rng := rand.New(rand.NewSource(seed))
	totalEnemies := 0
	for range spawnRooms {
		totalEnemies += 1 + rng.Intn(3) // 1-3 enemies per room
	}

	// Update params with entity count
	params.Custom = make(map[string]interface{})
	params.Custom["count"] = totalEnemies

	// Generate entities
	result, err := entityGen.Generate(seed+1000, params)
	if err != nil {
		return 0, fmt.Errorf("failed to generate entities: %w", err)
	}

	generatedEntities := result.([]*entity.Entity)
	if len(generatedEntities) == 0 {
		return 0, nil
	}

	// Spawn entities in rooms
	entityIndex := 0
	spawned := 0

	for _, room := range spawnRooms {
		if entityIndex >= len(generatedEntities) {
			break
		}

		// Number of enemies for this room (1-3)
		roomEnemyCount := 1 + rng.Intn(3)
		if roomEnemyCount > len(generatedEntities)-entityIndex {
			roomEnemyCount = len(generatedEntities) - entityIndex
		}

		for i := 0; i < roomEnemyCount; i++ {
			if entityIndex >= len(generatedEntities) {
				break
			}

			genEntity := generatedEntities[entityIndex]
			entityIndex++

			// Calculate spawn position (room center with slight offset)
			cx, cy := room.Center()
			offsetX := rng.Float64()*20 - 10 // -10 to +10
			offsetY := rng.Float64()*20 - 10
			spawnX := float64(cx*32) + offsetX
			spawnY := float64(cy*32) + offsetY

			// Create ECS entity
			enemy := world.CreateEntity()

			// Position
			enemy.AddComponent(&PositionComponent{
				X: spawnX,
				Y: spawnY,
			})

			// Health (scale from procgen entity stats)
			maxHealth := float64(genEntity.Stats.Health)
			enemy.AddComponent(&HealthComponent{
				Current: maxHealth,
				Max:     maxHealth,
			})

			// Stats
			stats := NewStatsComponent()
			stats.Attack = float64(genEntity.Stats.Damage)
			stats.Defense = float64(genEntity.Stats.Defense)
			enemy.AddComponent(stats)

			// Team (enemy team = 2, player team = 1)
			enemy.AddComponent(&TeamComponent{TeamID: 2})

			// Velocity (required for movement)
			enemy.AddComponent(&VelocityComponent{VX: 0, VY: 0})

			// Attack capability
			attackRange := 50.0 // Base melee range
			if genEntity.Size == entity.SizeLarge || genEntity.Size == entity.SizeHuge {
				attackRange = 70.0 // Larger enemies have longer reach
			}

			enemy.AddComponent(&AttackComponent{
				Damage:     float64(genEntity.Stats.Damage),
				DamageType: 0, // Physical damage
				Range:      attackRange,
				Cooldown:   1.0, // 1 second between attacks
			})

			// AI behavior
			aiComp := NewAIComponent(spawnX, spawnY)
			aiComp.DetectionRange = 200.0 // Can detect player from 200 pixels

			// Boss entities are more aggressive with wider detection
			if genEntity.Type == entity.TypeBoss {
				aiComp.DetectionRange = 300.0
				aiComp.ChaseSpeed = 0.8 // Slower but tankier
			} else if genEntity.Type == entity.TypeMinion {
				aiComp.ChaseSpeed = 1.2 // Faster but weaker
			}

			enemy.AddComponent(aiComp)

			// Collision
			enemySize := 32.0
			if genEntity.Size == entity.SizeTiny {
				enemySize = 16.0
			} else if genEntity.Size == entity.SizeSmall {
				enemySize = 24.0
			} else if genEntity.Size == entity.SizeLarge {
				enemySize = 48.0
			} else if genEntity.Size == entity.SizeHuge {
				enemySize = 64.0
			}

			enemy.AddComponent(&ColliderComponent{
				Width:     enemySize,
				Height:    enemySize,
				Solid:     true,
				IsTrigger: false,
				Layer:     1,
				OffsetX:   -enemySize / 2,
				OffsetY:   -enemySize / 2,
			})

			// Visual sprite (procedurally generated, animated)
			enemySprite := &EbitenSprite{
				Width:   enemySize,
				Height:  enemySize,
				Visible: true,
				Layer:   5, // Enemies drawn below player (layer 10)
			}
			enemy.AddComponent(enemySprite)

			// GAP-018 REPAIR: Add animation component for enemy animations
			enemyAnim := NewAnimationComponent(seed + int64(enemy.ID))
			enemyAnim.CurrentState = AnimationStateIdle
			enemyAnim.FrameTime = 0.2 // Slightly slower than player (~5 FPS)
			enemyAnim.Loop = true
			enemyAnim.Playing = true
			enemyAnim.FrameCount = 4
			enemy.AddComponent(enemyAnim)

			// GAP-012 REPAIR: Add visual feedback for hit flash
			enemy.AddComponent(NewVisualFeedbackComponent())

			spawned++
		}
	}

	return spawned, nil
}

// getEnemyColor determines sprite color based on entity properties.
func getEnemyColor(e *entity.Entity) color.RGBA {
	// Base color on entity type
	var baseColor color.RGBA

	switch e.Type {
	case entity.TypeBoss:
		baseColor = color.RGBA{200, 50, 50, 255} // Dark red for bosses
	case entity.TypeMinion:
		baseColor = color.RGBA{100, 100, 150, 255} // Purple-ish for minions
	case entity.TypeNPC:
		baseColor = color.RGBA{100, 200, 100, 255} // Green for NPCs
	default: // Monster
		baseColor = color.RGBA{180, 80, 80, 255} // Red for monsters
	}

	// Modify based on rarity
	switch e.Rarity {
	case entity.RarityUncommon:
		baseColor.G += 30
	case entity.RarityRare:
		baseColor.B += 50
	case entity.RarityEpic:
		baseColor.R += 40
		baseColor.B += 40
	case entity.RarityLegendary:
		baseColor.R += 60
		baseColor.G += 60
		baseColor.B += 60
	}

	return baseColor
}

// SpawnEnemyFromTemplate spawns a single enemy from a procedurally generated entity.
// This is a helper for spawning individual enemies with full control.
func SpawnEnemyFromTemplate(world *World, genEntity *entity.Entity, x, y float64) *Entity {
	enemy := world.CreateEntity()

	// Position
	enemy.AddComponent(&PositionComponent{X: x, Y: y})

	// Health
	maxHealth := float64(genEntity.Stats.Health)
	enemy.AddComponent(&HealthComponent{Current: maxHealth, Max: maxHealth})

	// Stats
	stats := NewStatsComponent()
	stats.Attack = float64(genEntity.Stats.Damage)
	stats.Defense = float64(genEntity.Stats.Defense)
	enemy.AddComponent(stats)

	// Team
	enemy.AddComponent(&TeamComponent{TeamID: 2})

	// Velocity
	enemy.AddComponent(&VelocityComponent{VX: 0, VY: 0})

	// Attack
	enemy.AddComponent(&AttackComponent{
		Damage:     float64(genEntity.Stats.Damage),
		DamageType: 0,
		Range:      50.0,
		Cooldown:   1.0,
	})

	// AI
	aiComp := NewAIComponent(x, y)
	aiComp.DetectionRange = 200.0
	enemy.AddComponent(aiComp)

	// Collision
	enemySize := 32.0
	enemy.AddComponent(&ColliderComponent{
		Width:     enemySize,
		Height:    enemySize,
		Solid:     true,
		IsTrigger: false,
		Layer:     1,
		OffsetX:   -enemySize / 2,
		OffsetY:   -enemySize / 2,
	})

	// Sprite (animated)
	enemySprite := &EbitenSprite{
		Width:   enemySize,
		Height:  enemySize,
		Visible: true,
		Layer:   5,
	}
	enemy.AddComponent(enemySprite)

	// Animation
	enemyAnim := NewAnimationComponent(12345 + int64(enemy.ID))
	enemyAnim.CurrentState = AnimationStateIdle
	enemyAnim.FrameTime = 0.2
	enemyAnim.Loop = true
	enemyAnim.Playing = true
	enemyAnim.FrameCount = 4
	enemy.AddComponent(enemyAnim)

	return enemy
}

// GenerateEnemySprite creates a procedural sprite for an enemy entity.
// Uses the sprite generation system to create varied enemy visuals.
func GenerateEnemySprite(genEntity *entity.Entity, seed int64) (color.RGBA, error) {
	// For now, just return color based on entity properties
	// Full sprite generation can be integrated later with sprites package
	return getEnemyColor(genEntity), nil
}
