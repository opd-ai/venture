package housing

import (
	"fmt"
	"image/color"
	"sync/atomic"
	"time"
)

// BuildingSize represents the size tier of a building.
type BuildingSize int

const (
	SizeSmall  BuildingSize = 8  // 8×8 tiles (256 square units)
	SizeMedium BuildingSize = 16 // 16×16 tiles (1024 square units)
	SizeLarge  BuildingSize = 24 // 24×24 tiles (2304 square units)
	SizeEstate BuildingSize = 32 // 32×32 tiles (4096 square units)
)

// String returns a human-readable name for the building size.
func (s BuildingSize) String() string {
	switch s {
	case SizeSmall:
		return "Small"
	case SizeMedium:
		return "Medium"
	case SizeLarge:
		return "Large"
	case SizeEstate:
		return "Estate"
	default:
		return "Unknown"
	}
}

// Tiles returns the width/height in tiles for this building size.
func (s BuildingSize) Tiles() int {
	return int(s)
}

// SquareUnits returns the total area in square tiles.
func (s BuildingSize) SquareUnits() int {
	tiles := int(s)
	return tiles * tiles
}

// Vector2 represents a 2D position or offset.
type Vector2 struct {
	X, Y float64
}

// PermissionLevel defines access control for housing.
type PermissionLevel int

const (
	PermissionNone    PermissionLevel = 0 // No access
	PermissionVisit   PermissionLevel = 1 // Can visit (view only)
	PermissionFriend  PermissionLevel = 2 // Can use furniture
	PermissionCoOwner PermissionLevel = 3 // Can modify building
)

// String returns a human-readable name for the permission level.
func (p PermissionLevel) String() string {
	switch p {
	case PermissionNone:
		return "None"
	case PermissionVisit:
		return "Visit"
	case PermissionFriend:
		return "Friend"
	case PermissionCoOwner:
		return "CoOwner"
	default:
		return "Unknown"
	}
}

// PermissionSet tracks player permissions for a building.
type PermissionSet struct {
	DefaultLevel PermissionLevel            // Default for unlisted players
	PlayerPerms  map[string]PermissionLevel // Per-player permissions
}

// NewPermissionSet creates a new permission set with owner-only access.
func NewPermissionSet() *PermissionSet {
	return &PermissionSet{
		DefaultLevel: PermissionNone,
		PlayerPerms:  make(map[string]PermissionLevel),
	}
}

// GetPermission returns the permission level for a player.
func (ps *PermissionSet) GetPermission(playerID string) PermissionLevel {
	if perm, ok := ps.PlayerPerms[playerID]; ok {
		return perm
	}
	return ps.DefaultLevel
}

// SetPermission sets the permission level for a player.
func (ps *PermissionSet) SetPermission(playerID string, level PermissionLevel) {
	ps.PlayerPerms[playerID] = level
}

// Plot represents a placed building in the world.
type Plot struct {
	ID          string         // Unique identifier
	OwnerID     string         // Player who owns this plot
	Position    Vector2        // World position (center of plot)
	Size        BuildingSize   // Size tier
	Permissions *PermissionSet // Access control
	CreatedAt   time.Time      // Creation timestamp
	ModifiedAt  time.Time      // Last modification timestamp
	BuildingID  string         // Reference to procedurally generated building
	Theme       string         // Genre/architectural theme
	Color       color.RGBA     // Primary color
}

// NewPlot creates a new plot with default settings.
func NewPlot(ownerID string, position Vector2, size BuildingSize) *Plot {
	now := time.Now()
	return &Plot{
		ID:          generatePlotID(),
		OwnerID:     ownerID,
		Position:    position,
		Size:        size,
		Permissions: NewPermissionSet(),
		CreatedAt:   now,
		ModifiedAt:  now,
		Color:       color.RGBA{R: 128, G: 128, B: 128, A: 255},
	}
}

// Bounds returns the axis-aligned bounding box for this plot.
// Returns min and max corners in world coordinates.
func (p *Plot) Bounds() (min, max Vector2) {
	halfSize := float64(p.Size.Tiles()) / 2.0
	min = Vector2{
		X: p.Position.X - halfSize,
		Y: p.Position.Y - halfSize,
	}
	max = Vector2{
		X: p.Position.X + halfSize,
		Y: p.Position.Y + halfSize,
	}
	return min, max
}

// Contains checks if a point is within this plot's bounds.
func (p *Plot) Contains(point Vector2) bool {
	min, max := p.Bounds()
	return point.X >= min.X && point.X <= max.X &&
		point.Y >= min.Y && point.Y <= max.Y
}

// Overlaps checks if this plot overlaps with another plot.
func (p *Plot) Overlaps(other *Plot) bool {
	min1, max1 := p.Bounds()
	min2, max2 := other.Bounds()

	// Check for no overlap (easier to reason about)
	noOverlap := max1.X < min2.X || max2.X < min1.X ||
		max1.Y < min2.Y || max2.Y < min1.Y

	return !noOverlap
}

// plotIDCounter uses atomic operations for thread-safe ID generation
var plotIDCounter atomic.Int64

func generatePlotID() string {
	id := plotIDCounter.Add(1)
	return fmt.Sprintf("plot_%d", id)
}

// House represents a player-owned building with interior and exterior data.
// This is a higher-level abstraction over Plot that includes building-specific details.
type House struct {
	ID      string // Unique identifier (matches Plot.ID)
	OwnerID string // Player who owns this house
	Plot    *Plot  // Associated plot in the world
	// Additional fields for building details, rooms, furniture would go here
}

// Dimensions is a helper for the integration tests.
type Dimensions = Vector2
