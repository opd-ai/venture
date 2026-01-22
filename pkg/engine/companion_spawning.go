// Package engine provides companion spawning utilities for V4.0 integration.
// This file implements functions to spawn procedurally generated companions into the game world.
package engine

import (
	"fmt"
	"image/color"
	"math/rand"

	"github.com/opd-ai/venture/pkg/procgen/dialog"
	"github.com/opd-ai/venture/pkg/procgen/terrain"
)

// CompanionSpawnData contains the necessary information to spawn a companion entity.
// This avoids import cycles by not depending on pkg/procgen/companion directly.
type CompanionSpawnData struct {
	Name          string
	CompanionType CompanionType
	Level         int
	Attack        float64
	Defense       float64
	Speed         float64
	HP            float64
	MaxHP         float64
	Loyalty       float64
	Commands      []CommandType
	Color         color.RGBA
	Size          int     // Sprite size
	ColliderSize  float64 // Collider dimensions
}

// SpawnCompanionsInTerrain spawns procedurally generated companions into terrain rooms.
// Places recruitable companion NPCs in larger rooms that serve as settlements.
// Returns the number of companions spawned.
//
// Note: This function expects companion generation to be done externally to avoid import cycles.
// Call this from cmd/client with companions generated via companiongen.NewGenerator().
func SpawnCompanionsInTerrain(world *World, terr *terrain.Terrain, companions []CompanionSpawnData, seed int64) (int, error) {
	if err := validateTerrainRooms(terr); err != nil {
		return 0, err
	}

	if len(companions) == 0 {
		return 0, nil
	}

	rng := rand.New(rand.NewSource(seed + 6000))
	settlementRooms := selectSettlementRooms(terr, rng)

	if len(settlementRooms) == 0 {
		return 0, fmt.Errorf("no available rooms for companion spawning")
	}

	return spawnCompanionsInRooms(world, settlementRooms, companions, seed, rng), nil
}

// validateTerrainRooms checks if terrain has sufficient rooms for spawning.
func validateTerrainRooms(terr *terrain.Terrain) error {
	if len(terr.Rooms) < 2 {
		return fmt.Errorf("insufficient rooms for companion spawning (need at least 2, got %d)", len(terr.Rooms))
	}
	return nil
}

// selectSettlementRooms finds and shuffles suitable rooms for companion placement.
func selectSettlementRooms(terr *terrain.Terrain, rng *rand.Rand) []*terrain.Room {
	settlementRooms := filterLargeRooms(terr)

	if len(settlementRooms) == 0 {
		settlementRooms = collectNonSpawnRooms(terr)
	}

	rng.Shuffle(len(settlementRooms), func(i, j int) {
		settlementRooms[i], settlementRooms[j] = settlementRooms[j], settlementRooms[i]
	})

	return settlementRooms
}

// filterLargeRooms selects rooms with area >= 35 tiles for settlement placement.
func filterLargeRooms(terr *terrain.Terrain) []*terrain.Room {
	settlementRooms := make([]*terrain.Room, 0)
	for i := 1; i < len(terr.Rooms); i++ {
		room := terr.Rooms[i]
		if room.Width*room.Height >= 35 {
			settlementRooms = append(settlementRooms, room)
		}
	}
	return settlementRooms
}

// collectNonSpawnRooms gathers all non-spawn rooms as fallback.
func collectNonSpawnRooms(terr *terrain.Terrain) []*terrain.Room {
	rooms := make([]*terrain.Room, 0)
	for i := 1; i < len(terr.Rooms); i++ {
		rooms = append(rooms, terr.Rooms[i])
	}
	return rooms
}

// spawnCompanionsInRooms creates companion entities in selected rooms.
func spawnCompanionsInRooms(world *World, rooms []*terrain.Room, companions []CompanionSpawnData, seed int64, rng *rand.Rand) int {
	spawned := 0
	for i, companionData := range companions {
		if i >= len(rooms) {
			break
		}

		spawnX, spawnY := calculateCompanionSpawnPosition(rooms[i], rng)
		entity := createCompanionEntity(world, companionData, spawnX, spawnY, seed, "fantasy")
		if entity != nil {
			spawned++
		}
	}
	return spawned
}

// calculateCompanionSpawnPosition computes a randomized spawn position within a room.
func calculateCompanionSpawnPosition(room *terrain.Room, rng *rand.Rand) (float64, float64) {
	cx, cy := room.Center()
	offsetX := rng.Float64()*30 - 15
	offsetY := rng.Float64()*30 - 15
	return float64(cx*32) + offsetX, float64(cy*32) + offsetY
}

// createCompanionEntity creates a companion entity with the provided data.
func createCompanionEntity(world *World, companionData CompanionSpawnData, x, y float64, seed int64, genreID string) *Entity {
	entity := world.CreateEntity()

	// Add position
	entity.AddComponent(&PositionComponent{X: x, Y: y})

	// Add velocity (companions start stationary)
	entity.AddComponent(&VelocityComponent{VX: 0, VY: 0})

	// Add companion component
	companionComp := &CompanionComponent{
		OwnerID:       0, // No owner initially (can be recruited)
		CompanionType: companionData.CompanionType,
		Loyalty:       companionData.Loyalty,
		Experience:    0,
		Level:         companionData.Level,
		Behavior:      BehaviorPassive, // Start passive until recruited
		Commands:      companionData.Commands,
		Permadeath:    false,
		BondingPerks:  []BondingPerk{},
		TimeWithOwner: 0,
	}
	entity.AddComponent(companionComp)

	// Add companion stats component
	statsComp := &CompanionStatsComponent{
		Attack:  companionData.Attack,
		Defense: companionData.Defense,
		Speed:   companionData.Speed,
		HP:      companionData.HP,
		MaxHP:   companionData.MaxHP,
	}
	entity.AddComponent(statsComp)

	// Add health component (for damage/healing)
	healthComp := &HealthComponent{
		Max:     companionData.MaxHP,
		Current: companionData.HP,
	}
	entity.AddComponent(healthComp)

	// Add sprite
	sprite := NewSpriteComponent(float64(companionData.Size), float64(companionData.Size), companionData.Color)
	sprite.Layer = 9 // Same layer as NPCs
	sprite.Visible = true
	entity.AddComponent(sprite)

	// Add animation component
	animComp := NewAnimationComponent(int64(entity.ID))
	animComp.CurrentState = AnimationStateIdle
	animComp.FrameTime = 0.2 // Standard NPC animation speed
	animComp.Loop = true
	animComp.Playing = true
	animComp.FrameCount = 8 // V7.0: 8-frame animations
	entity.AddComponent(animComp)

	// Add collider
	entity.AddComponent(&ColliderComponent{
		Width:     companionData.ColliderSize,
		Height:    companionData.ColliderSize,
		Solid:     true,
		IsTrigger: false,
		Layer:     1, // NPC layer
		OffsetX:   -companionData.ColliderSize / 2,
		OffsetY:   -companionData.ColliderSize / 2,
	})

	// Add team component (neutral/friendly initially)
	entity.AddComponent(&TeamComponent{TeamID: 1}) // Team 1 = friendly NPCs

	// Add dialog component for recruitment interaction (Phase 3.1)
	companionPersonality := dialog.NewPersonality(dialog.PersonalityHelpful)
	dialogProvider := NewMarkovDialogProvider(seed+int64(entity.ID), genreID, companionData.Name, companionPersonality)
	dialogComp := NewDialogComponent(dialogProvider)
	entity.AddComponent(dialogComp)

	return entity
}
