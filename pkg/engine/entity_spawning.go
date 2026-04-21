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
		return 0, nil
	}

	spawnRooms := selectSpawnRooms(terr.Rooms)
	rng := rand.New(rand.NewSource(seed))

	generatedEntities, err := generateEnemyEntities(seed, params, spawnRooms, rng)
	if err != nil {
		return 0, err
	}
	if len(generatedEntities) == 0 {
		return 0, nil
	}

	return spawnEntitiesInRooms(world, spawnRooms, generatedEntities, seed, rng, params.GenreID), nil
}

// selectSpawnRooms returns rooms for enemy spawning, excluding the first room (player spawn).
func selectSpawnRooms(rooms []*terrain.Room) []*terrain.Room {
	if len(rooms) > 1 {
		return rooms[1:]
	}
	return rooms
}

// generateEnemyEntities creates procedurally generated enemy entities for the given rooms.
func generateEnemyEntities(seed int64, params procgen.GenerationParams, rooms []*terrain.Room, rng *rand.Rand) ([]*entity.Entity, error) {
	totalEnemies := calculateTotalEnemies(rooms, rng)

	params.Custom = make(map[string]interface{})
	params.Custom["count"] = totalEnemies

	entityGen := entity.NewEntityGenerator()
	result, err := entityGen.Generate(seed+1000, params)
	if err != nil {
		return nil, fmt.Errorf("failed to generate entities: %w", err)
	}

	return result.([]*entity.Entity), nil
}

// calculateTotalEnemies determines the total number of enemies to spawn (1-3 per room).
func calculateTotalEnemies(rooms []*terrain.Room, rng *rand.Rand) int {
	total := 0
	for range rooms {
		total += 1 + rng.Intn(3)
	}
	return total
}

// spawnEntitiesInRooms creates and configures ECS entities in each room.
func spawnEntitiesInRooms(world *World, rooms []*terrain.Room, entities []*entity.Entity, seed int64, rng *rand.Rand, genreID string) int {
	entityIndex := 0
	spawned := 0

	for roomIdx, room := range rooms {
		if entityIndex >= len(entities) {
			break
		}

		roomEnemyCount := calculateRoomEnemyCount(rng, len(entities)-entityIndex)

		for i := 0; i < roomEnemyCount; i++ {
			if entityIndex >= len(entities) {
				break
			}

			genEntity := entities[entityIndex]
			entityIndex++

			spawnX, spawnY := calculateSpawnPosition(room, rng)
			enemy := createConfiguredEnemy(world, genEntity, spawnX, spawnY, seed, roomIdx, i, len(rooms), rng, genreID)

			if enemy != nil {
				spawned++
			}
		}
	}

	return spawned
}

// calculateRoomEnemyCount determines how many enemies to spawn in a room (1-3, capped by available).
func calculateRoomEnemyCount(rng *rand.Rand, remainingEntities int) int {
	count := 1 + rng.Intn(3)
	if count > remainingEntities {
		count = remainingEntities
	}
	return count
}

// calculateSpawnPosition determines the spawn location within a room with random offset.
func calculateSpawnPosition(room *terrain.Room, rng *rand.Rand) (float64, float64) {
	cx, cy := room.Center()
	offsetX := rng.Float64()*20 - 10
	offsetY := rng.Float64()*20 - 10
	return float64(cx*32) + offsetX, float64(cy*32) + offsetY
}

// createConfiguredEnemy creates an ECS entity with all necessary components for an enemy.
func createConfiguredEnemy(world *World, genEntity *entity.Entity, x, y float64, seed int64, roomIdx, enemyIdx, totalRooms int, rng *rand.Rand, genreID string) *Entity {
	enemy := world.CreateEntity()

	addCoreComponents(enemy, genEntity, x, y)
	addCombatComponents(enemy, genEntity, x, y)
	enemySize := addVisualComponents(enemy, genEntity, seed)
	addAdvancedComponents(enemy, genEntity, enemySize, seed, roomIdx, enemyIdx, totalRooms, rng, world, genreID)

	return enemy
}

// addCoreComponents adds position, health, stats, team, and velocity components.
func addCoreComponents(enemy *Entity, genEntity *entity.Entity, x, y float64) {
	enemy.AddComponent(&PositionComponent{X: x, Y: y})

	maxHealth := float64(genEntity.Stats.Health)
	enemy.AddComponent(&HealthComponent{Current: maxHealth, Max: maxHealth})

	stats := NewStatsComponent()
	stats.Attack = float64(genEntity.Stats.Damage)
	stats.Defense = float64(genEntity.Stats.Defense)
	enemy.AddComponent(stats)

	enemy.AddComponent(&TeamComponent{TeamID: 2})
	enemy.AddComponent(&VelocityComponent{VX: 0, VY: 0})
}

// addCombatComponents adds attack and AI components based on entity properties.
func addCombatComponents(enemy *Entity, genEntity *entity.Entity, x, y float64) {
	attackRange := calculateAttackRange(genEntity.Size)
	enemy.AddComponent(&AttackComponent{
		Damage:     float64(genEntity.Stats.Damage),
		DamageType: 0,
		Range:      attackRange,
		Cooldown:   1.0,
	})

	aiComp := configureAIComponent(genEntity, x, y)
	enemy.AddComponent(aiComp)
}

// calculateAttackRange returns the attack range based on entity size.
func calculateAttackRange(size entity.EntitySize) float64 {
	if size == entity.SizeLarge || size == entity.SizeHuge {
		return 70.0
	}
	return 50.0
}

// configureAIComponent creates an AI component with detection and speed based on entity type.
func configureAIComponent(genEntity *entity.Entity, x, y float64) *AIComponent {
	aiComp := NewAIComponent(x, y)
	aiComp.DetectionRange = 200.0

	if genEntity.Type == entity.TypeBoss {
		aiComp.DetectionRange = 300.0
		aiComp.ChaseSpeed = 0.8
	} else if genEntity.Type == entity.TypeMinion {
		aiComp.ChaseSpeed = 1.2
	}

	return aiComp
}

// addVisualComponents adds collision, sprite, and animation components, returning enemy size.
func addVisualComponents(enemy *Entity, genEntity *entity.Entity, seed int64) float64 {
	enemySize := calculateEnemySize(genEntity.Size)

	enemy.AddComponent(&ColliderComponent{
		Width:     enemySize,
		Height:    enemySize,
		Solid:     true,
		IsTrigger: false,
		Layer:     1,
		OffsetX:   -enemySize / 2,
		OffsetY:   -enemySize / 2,
	})

	enemy.AddComponent(&EbitenSprite{
		Width:   enemySize,
		Height:  enemySize,
		Visible: true,
		Layer:   5,
	})

	enemyAnim := NewAnimationComponent(seed + int64(enemy.ID))
	enemyAnim.CurrentState = AnimationStateIdle
	enemyAnim.FrameTime = 0.2
	enemyAnim.Loop = true
	enemyAnim.Playing = true
	enemyAnim.FrameCount = 8 // V7.0: 8-frame animations
	enemyAnim.Dirty = true
	enemy.AddComponent(enemyAnim)

	return enemySize
}

// calculateEnemySize returns the visual size based on entity size category.
func calculateEnemySize(size entity.EntitySize) float64 {
	switch size {
	case entity.SizeTiny:
		return 16.0
	case entity.SizeSmall:
		return 24.0
	case entity.SizeLarge:
		return 48.0
	case entity.SizeHuge:
		return 64.0
	default:
		return 32.0
	}
}

// addAdvancedComponents adds feedback, rotation, layer, shadow, behavior, squad, faction, and genre components.
func addAdvancedComponents(enemy *Entity, genEntity *entity.Entity, enemySize float64, seed int64, roomIdx, enemyIdx, totalRooms int, rng *rand.Rand, world *World, genreID string) {
	enemy.AddComponent(NewVisualFeedbackComponent())
	enemy.AddComponent(NewRotationComponent(0, 2.0))

	layerComp := NewLayerComponent()
	layerComp.CurrentLayer = 0
	enemy.AddComponent(&layerComp)

	shadowComp := NewShadowComponent(enemySize)
	shadowComp.CastsShadow = true
	shadowComp.ShadowType = ShadowTypeHard
	enemy.AddComponent(shadowComp)

	archetype := selectEnemyArchetype(genEntity, rng)
	enemy.AddComponent(BuildBehaviorTree(archetype, world))

	squadID := totalRooms - (totalRooms - roomIdx)
	role := selectSquadRole(enemyIdx)
	enemy.AddComponent(NewSquadComponent(squadID, role, 0))

	factionID := selectFactionID(genEntity)
	enemy.AddComponent(&FactionComponent{
		FactionID:       factionID,
		Reputation:      0,
		IsPlayerFaction: false,
	})

	// Creature visual classification from procgen entity data.
	creatureForm := ClassifyCreatureForm(genEntity.Name, genEntity.Tags)
	enemy.AddComponent(&CreatureVisualComponent{
		Form:       creatureForm,
		SizeClass:  genEntity.Size.String(),
		VisualTags: genEntity.Tags,
	})

	// Genre component for genre-based reward scaling and theming (AUDIT.md G6).
	if genreID != "" {
		enemy.AddComponent(NewGenreComponent(genreID))
	}
}

// selectEnemyArchetype determines the AI archetype based on entity properties.
func selectEnemyArchetype(genEntity *entity.Entity, rng *rand.Rand) EnemyArchetype {
	if genEntity.Type == entity.TypeBoss {
		return ArchetypeTank
	}
	if genEntity.Size == entity.SizeTiny || genEntity.Size == entity.SizeSmall {
		return ArchetypeStealth
	}
	if genEntity.Stats.Damage > genEntity.Stats.Defense {
		return ArchetypeMelee
	}

	switch rng.Intn(3) {
	case 0:
		return ArchetypeMelee
	case 1:
		return ArchetypeRanged
	case 2:
		return ArchetypeSupport
	default:
		return ArchetypeMelee
	}
}

// selectSquadRole returns the squad role based on position in room (first is leader).
func selectSquadRole(enemyIdx int) SquadRole {
	if enemyIdx == 0 {
		return SquadRoleLeader
	}
	return SquadRoleMember
}

// selectFactionID determines the faction based on entity type.
func selectFactionID(genEntity *entity.Entity) string {
	if genEntity.Type == entity.TypeBoss {
		return "boss_faction"
	}
	if genEntity.Type == entity.TypeNPC {
		return "neutral_faction"
	}
	return "enemy_faction"
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
	enemyAnim.FrameCount = 8 // V7.0: 8-frame animations
	enemyAnim.Dirty = true   // CRITICAL: Mark dirty to trigger initial frame generation
	enemy.AddComponent(enemyAnim)

	// Visual feedback (Phase 10)
	enemy.AddComponent(NewVisualFeedbackComponent())

	// Rotation (Phase 10.1)
	enemy.AddComponent(NewRotationComponent(0, 2.0))

	// Layer (Phase 11.1)
	layerComp := NewLayerComponent()
	layerComp.CurrentLayer = 0
	enemy.AddComponent(&layerComp)

	// Shadow (Phase 14)
	shadowComp := NewShadowComponent(enemySize)
	shadowComp.CastsShadow = true
	shadowComp.ShadowType = ShadowTypeHard
	enemy.AddComponent(shadowComp)

	// Behavior tree (Phase 13.1) - default to melee
	behaviorTree := BuildBehaviorTree(ArchetypeMelee, world)
	enemy.AddComponent(behaviorTree)

	// Squad (Phase 13.2) - solo enemy
	squadComp := NewSquadComponent(int(enemy.ID), SquadRoleLeader, 0)
	enemy.AddComponent(squadComp)

	// Faction (Phase 13.3)
	var factionID string
	if genEntity.Type == entity.TypeBoss {
		factionID = "boss_faction"
	} else if genEntity.Type == entity.TypeNPC {
		factionID = "neutral_faction"
	} else {
		factionID = "enemy_faction"
	}
	factionComp := &FactionComponent{
		FactionID:       factionID,
		Reputation:      0,
		IsPlayerFaction: false,
	}
	enemy.AddComponent(factionComp)

	// Creature visual classification from procgen entity data.
	creatureForm := ClassifyCreatureForm(genEntity.Name, genEntity.Tags)
	enemy.AddComponent(&CreatureVisualComponent{
		Form:       creatureForm,
		SizeClass:  genEntity.Size.String(),
		VisualTags: genEntity.Tags,
	})

	return enemy
}

// GenerateEnemySprite creates a procedural sprite for an enemy entity.
// Uses the sprite generation system to create varied enemy visuals.
func GenerateEnemySprite(genEntity *entity.Entity, seed int64) (color.RGBA, error) {
	// For now, just return color based on entity properties
	// Full sprite generation can be integrated later with sprites package
	return getEnemyColor(genEntity), nil
}
