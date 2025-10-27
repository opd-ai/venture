// Package engine provides terrain modification components for the ECS.
// This file defines components for destructible terrain, fire propagation,
// and constructible walls. The environmental manipulation system enables
// dynamic world changes through player actions (weapons, spells, building).
//
// Design Philosophy:
// - Components contain only data, no behavior (ECS pattern)
// - Terrain modifications are server-authoritative for multiplayer
// - Fire propagation uses cellular automata for emergent behavior
// - Performance target: <5ms per frame with 100 fire entities
package engine

import (
	"time"

	"github.com/opd-ai/venture/pkg/world"
)

// MaterialType represents the material composition of terrain.
type MaterialType int

const (
	// MaterialStone represents stone walls (high durability, not flammable)
	MaterialStone MaterialType = iota
	// MaterialWood represents wooden structures (medium durability, flammable)
	MaterialWood
	// MaterialEarth represents dirt/earth (low durability, not flammable)
	MaterialEarth
	// MaterialMetal represents metal walls (very high durability, not flammable)
	MaterialMetal
	// MaterialGlass represents glass (low durability, not flammable, transparent)
	MaterialGlass
	// MaterialIce represents frozen water (low durability, melts near fire)
	MaterialIce
)

// String returns the string representation of a material type.
func (m MaterialType) String() string {
	switch m {
	case MaterialStone:
		return "stone"
	case MaterialWood:
		return "wood"
	case MaterialEarth:
		return "earth"
	case MaterialMetal:
		return "metal"
	case MaterialGlass:
		return "glass"
	case MaterialIce:
		return "ice"
	default:
		return "unknown"
	}
}

// IsFlammable returns true if the material can catch fire.
func (m MaterialType) IsFlammable() bool {
	return m == MaterialWood
}

// BaseDurability returns the base health for this material type.
// This represents how many hits it takes to destroy terrain of this material.
func (m MaterialType) BaseDurability() float64 {
	switch m {
	case MaterialStone:
		return 100.0
	case MaterialWood:
		return 50.0
	case MaterialEarth:
		return 30.0
	case MaterialMetal:
		return 200.0
	case MaterialGlass:
		return 20.0
	case MaterialIce:
		return 40.0
	default:
		return 50.0
	}
}

// DestructibleComponent marks a tile entity as destructible and tracks its health.
// When health reaches zero, the tile is destroyed and replaced with floor.
type DestructibleComponent struct {
	// Material determines durability and flammability
	Material MaterialType

	// Health is current durability (0 = destroyed)
	Health float64

	// MaxHealth is the starting durability
	MaxHealth float64

	// TileX, TileY are the tile coordinates in the world map
	TileX int
	TileY int

	// IsDestroyed tracks if this tile has been destroyed
	IsDestroyed bool

	// LastDamageTime tracks when the tile was last damaged (for visual feedback)
	LastDamageTime time.Time
}

// Type returns the component type identifier.
func (d *DestructibleComponent) Type() string {
	return "destructible"
}

// NewDestructibleComponent creates a destructible component for a tile.
func NewDestructibleComponent(material MaterialType, tileX, tileY int) *DestructibleComponent {
	maxHealth := material.BaseDurability()
	return &DestructibleComponent{
		Material:       material,
		Health:         maxHealth,
		MaxHealth:      maxHealth,
		TileX:          tileX,
		TileY:          tileY,
		IsDestroyed:    false,
		LastDamageTime: time.Now(),
	}
}

// TakeDamage applies damage to the tile and returns true if destroyed.
func (d *DestructibleComponent) TakeDamage(damage float64) bool {
	d.Health -= damage
	d.LastDamageTime = time.Now()
	if d.Health <= 0 {
		d.Health = 0
		d.IsDestroyed = true
		return true
	}
	return false
}

// HealthPercent returns the health as a percentage (0.0-1.0).
func (d *DestructibleComponent) HealthPercent() float64 {
	if d.MaxHealth <= 0 {
		return 0
	}
	return d.Health / d.MaxHealth
}

// FireComponent tracks fire on a tile and manages propagation.
// Fire damages entities standing on the tile and can spread to adjacent tiles.
type FireComponent struct {
	// Intensity affects damage and spread chance (0.0-1.0)
	Intensity float64

	// Duration tracks how long the fire has been burning (seconds)
	Duration float64

	// MaxDuration is how long fire burns before extinguishing (seconds)
	MaxDuration float64

	// SpreadChance is base probability of spreading per second (0.0-1.0)
	SpreadChance float64

	// DamagePerSecond is fire damage dealt to entities on this tile
	DamagePerSecond float64

	// TileX, TileY are the tile coordinates in the world map
	TileX int
	TileY int

	// LastSpreadTime tracks when fire last attempted to spread
	LastSpreadTime time.Time

	// IsExtinguished tracks if fire has burned out
	IsExtinguished bool
}

// Type returns the component type identifier.
func (f *FireComponent) Type() string {
	return "fire"
}

// NewFireComponent creates a fire component for a tile.
// intensity: 0.0-1.0 (affects damage and spread)
// maxDuration: how long fire burns (typically 10-15 seconds)
func NewFireComponent(intensity float64, tileX, tileY int, maxDuration float64) *FireComponent {
	if intensity < 0 {
		intensity = 0
	}
	if intensity > 1.0 {
		intensity = 1.0
	}
	if maxDuration <= 0 {
		maxDuration = 12.0 // Default: 12 seconds
	}

	return &FireComponent{
		Intensity:       intensity,
		Duration:        0,
		MaxDuration:     maxDuration,
		SpreadChance:    0.3 * intensity, // Higher intensity = more spread
		DamagePerSecond: 5.0 * intensity, // Higher intensity = more damage
		TileX:           tileX,
		TileY:           tileY,
		LastSpreadTime:  time.Now(),
		IsExtinguished:  false,
	}
}

// Update advances fire time and checks if it should extinguish.
// deltaTime is in seconds.
func (f *FireComponent) Update(deltaTime float64) {
	f.Duration += deltaTime
	if f.Duration >= f.MaxDuration {
		f.IsExtinguished = true
	}
}

// RemainingTime returns how many seconds until fire extinguishes.
func (f *FireComponent) RemainingTime() float64 {
	remaining := f.MaxDuration - f.Duration
	if remaining < 0 {
		return 0
	}
	return remaining
}

// BuildableComponent marks a tile location as available for construction.
// Players can place walls, barriers, or structures at buildable locations.
type BuildableComponent struct {
	// TileX, TileY are the tile coordinates in the world map
	TileX int
	TileY int

	// RequiredMaterials maps material type to quantity needed
	RequiredMaterials map[MaterialType]int

	// ConstructionTime is how long building takes (seconds)
	ConstructionTime float64

	// ElapsedTime tracks construction progress (seconds)
	ElapsedTime float64

	// IsComplete indicates if construction is finished
	IsComplete bool

	// ResultTileType is what tile type to place when complete
	ResultTileType world.TileType
}

// Type returns the component type identifier.
func (b *BuildableComponent) Type() string {
	return "buildable"
}

// NewBuildableComponent creates a buildable component for a tile.
// tileX, tileY: world coordinates
// resultType: what tile to create (usually TileWall)
// constructionTime: seconds to build (default 3.0)
func NewBuildableComponent(tileX, tileY int, resultType world.TileType, constructionTime float64) *BuildableComponent {
	if constructionTime <= 0 {
		constructionTime = 3.0 // Default: 3 seconds
	}

	// Default: wall requires 10 stone
	requiredMaterials := make(map[MaterialType]int)
	requiredMaterials[MaterialStone] = 10

	return &BuildableComponent{
		TileX:             tileX,
		TileY:             tileY,
		RequiredMaterials: requiredMaterials,
		ConstructionTime:  constructionTime,
		ElapsedTime:       0,
		IsComplete:        false,
		ResultTileType:    resultType,
	}
}

// Update advances construction progress.
// deltaTime is in seconds.
func (b *BuildableComponent) Update(deltaTime float64) {
	if b.IsComplete {
		return
	}
	b.ElapsedTime += deltaTime
	if b.ElapsedTime >= b.ConstructionTime {
		b.IsComplete = true
	}
}

// Progress returns construction progress as percentage (0.0-1.0).
func (b *BuildableComponent) Progress() float64 {
	if b.ConstructionTime <= 0 {
		return 1.0
	}
	progress := b.ElapsedTime / b.ConstructionTime
	if progress > 1.0 {
		return 1.0
	}
	return progress
}
