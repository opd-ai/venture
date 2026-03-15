package housing

import (
	"encoding/json"
	"fmt"
	"image/color"
	"sync"
	"sync/atomic"
	"time"
)

// TimeProvider is an interface for obtaining timestamps.
// This enables deterministic timestamps for testing and network synchronization.
// In production, use RealTimeProvider. For tests, use MockTimeProvider.
type TimeProvider interface {
	// Now returns the current timestamp.
	Now() time.Time
}

// RealTimeProvider implements TimeProvider using the system clock.
// Use this in production code.
type RealTimeProvider struct{}

// Now returns the current system time.
func (r RealTimeProvider) Now() time.Time {
	return time.Now()
}

// MockTimeProvider implements TimeProvider with a fixed timestamp.
// Use this for testing and deterministic generation.
type MockTimeProvider struct {
	CurrentTime time.Time
}

// Now returns the mock timestamp.
func (m MockTimeProvider) Now() time.Time {
	return m.CurrentTime
}

// NewMockTimeProvider creates a MockTimeProvider from a seed.
// The timestamp is deterministically derived from the seed.
func NewMockTimeProvider(seed int64) *MockTimeProvider {
	return &MockTimeProvider{
		CurrentTime: time.Unix(seed, 0),
	}
}

// defaultTimeProvider is the package-level time provider used when none is specified.
var defaultTimeProvider TimeProvider = RealTimeProvider{}

// SetDefaultTimeProvider sets the package-level time provider.
// Use this to inject a MockTimeProvider for multiplayer synchronization or testing.
// The time provider affects NewPlot() and NewBlueprint() timestamp generation.
//
// Example for multiplayer:
//
//	housing.SetDefaultTimeProvider(housing.NewMockTimeProvider(gameSeed))
//	defer housing.ResetDefaultTimeProvider()
func SetDefaultTimeProvider(tp TimeProvider) {
	defaultTimeProvider = tp
}

// ResetDefaultTimeProvider restores the package-level time provider to RealTimeProvider.
// Call this after multiplayer sessions or tests to restore normal time behavior.
func ResetDefaultTimeProvider() {
	defaultTimeProvider = RealTimeProvider{}
}

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

// HousingBuildingData supplies dimension and optional position information
// to CreateHouse. It replaces the previous interface{} parameter and eliminates
// the need for a type assertion to extract building dimensions.
// Width and Height of zero both default to SizeMedium (16 tiles).
// Position is optional; if nil, CreateHouse generates a seed-based position.
type HousingBuildingData struct {
	Width    int
	Height   int
	Position *Vector2 // optional explicit position; nil = generate from seed
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
// Uses the default time provider for timestamps.
func NewPlot(ownerID string, position Vector2, size BuildingSize) *Plot {
	return NewPlotWithTime(ownerID, position, size, defaultTimeProvider)
}

// NewPlotWithTime creates a new plot with timestamps from the provided TimeProvider.
// Use this for deterministic timestamps in multiplayer synchronization and testing.
func NewPlotWithTime(ownerID string, position Vector2, size BuildingSize, tp TimeProvider) *Plot {
	now := tp.Now()
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

// BuildingDefinition contains the parameters needed to procedurally generate a building.
type BuildingDefinition struct {
	Type   int   // BuildingType from pkg/procgen/building
	Style  int   // ArchitecturalStyle from pkg/procgen/building
	Width  int   // Building width in tiles
	Height int   // Building height in tiles
	Floors int   // Number of floors
	Seed   int64 // Seed for deterministic generation
}

// Blueprint represents a shareable building design with metadata.
// Blueprint instances must not be copied after creation; always use pointers.
// The embedded mutex makes copying unsafe and will cause go vet warnings.
type Blueprint struct {
	mu          sync.Mutex          // Protects mutable fields (rating, ratingCount, downloads)
	ID          string              // Unique blueprint identifier
	Name        string              // User-friendly name
	Description string              // Detailed description
	Author      string              // Player ID who created this blueprint
	GenreID     string              // Genre (fantasy, scifi, horror, etc.)
	Tags        []string            // Searchable tags (medieval, manor, etc.)
	rating      float64             // Average rating (0.0-5.0) - protected by mu
	ratingCount int                 // Number of ratings - protected by mu
	downloads   int                 // Download count - protected by mu
	CreatedAt   time.Time           // Creation timestamp
	ModifiedAt  time.Time           // Last modification timestamp
	BuildingDef *BuildingDefinition // Building generation parameters
}

// GetRating returns the current average rating.
// Thread-safe for concurrent access.
func (bp *Blueprint) GetRating() float64 {
	bp.mu.Lock()
	defer bp.mu.Unlock()
	return bp.rating
}

// GetRatingCount returns the number of ratings.
// Thread-safe for concurrent access.
func (bp *Blueprint) GetRatingCount() int {
	bp.mu.Lock()
	defer bp.mu.Unlock()
	return bp.ratingCount
}

// GetDownloads returns the current download count.
// Thread-safe for concurrent access.
func (bp *Blueprint) GetDownloads() int {
	bp.mu.Lock()
	defer bp.mu.Unlock()
	return bp.downloads
}

// blueprintIDCounter uses atomic operations for thread-safe ID generation
var blueprintIDCounter atomic.Int64

// generateBlueprintID generates a unique blueprint ID.
func generateBlueprintID() string {
	id := blueprintIDCounter.Add(1)
	return fmt.Sprintf("bp_%d", id)
}

// blueprintJSON is a helper struct for JSON marshaling/unmarshaling.
// It exposes the private fields for serialization.
type blueprintJSON struct {
	ID          string              `json:"ID"`
	Name        string              `json:"Name"`
	Description string              `json:"Description"`
	Author      string              `json:"Author"`
	GenreID     string              `json:"GenreID"`
	Tags        []string            `json:"Tags"`
	Rating      float64             `json:"Rating"`
	RatingCount int                 `json:"RatingCount"`
	Downloads   int                 `json:"Downloads"`
	CreatedAt   time.Time           `json:"CreatedAt"`
	ModifiedAt  time.Time           `json:"ModifiedAt"`
	BuildingDef *BuildingDefinition `json:"BuildingDef"`
}

// MarshalJSON implements json.Marshaler for Blueprint.
func (bp *Blueprint) MarshalJSON() ([]byte, error) {
	bp.mu.Lock()
	defer bp.mu.Unlock()

	helper := blueprintJSON{
		ID:          bp.ID,
		Name:        bp.Name,
		Description: bp.Description,
		Author:      bp.Author,
		GenreID:     bp.GenreID,
		Tags:        bp.Tags,
		Rating:      bp.rating,
		RatingCount: bp.ratingCount,
		Downloads:   bp.downloads,
		CreatedAt:   bp.CreatedAt,
		ModifiedAt:  bp.ModifiedAt,
		BuildingDef: bp.BuildingDef,
	}
	return json.Marshal(helper)
}

// UnmarshalJSON implements json.Unmarshaler for Blueprint.
func (bp *Blueprint) UnmarshalJSON(data []byte) error {
	var helper blueprintJSON
	if err := json.Unmarshal(data, &helper); err != nil {
		return err
	}

	bp.ID = helper.ID
	bp.Name = helper.Name
	bp.Description = helper.Description
	bp.Author = helper.Author
	bp.GenreID = helper.GenreID
	bp.Tags = helper.Tags
	bp.rating = helper.Rating
	bp.ratingCount = helper.RatingCount
	bp.downloads = helper.Downloads
	bp.CreatedAt = helper.CreatedAt
	bp.ModifiedAt = helper.ModifiedAt
	bp.BuildingDef = helper.BuildingDef

	return nil
}

// NewBlueprint creates a new blueprint with default metadata.
// Uses the default time provider for timestamps.
func NewBlueprint(name, author, genreID string, buildingDef *BuildingDefinition) *Blueprint {
	return NewBlueprintWithTime(name, author, genreID, buildingDef, defaultTimeProvider)
}

// NewBlueprintWithTime creates a new blueprint with timestamps from the provided TimeProvider.
// Use this for deterministic timestamps in multiplayer synchronization and testing.
func NewBlueprintWithTime(name, author, genreID string, buildingDef *BuildingDefinition, tp TimeProvider) *Blueprint {
	now := tp.Now()
	return &Blueprint{
		ID:          generateBlueprintID(),
		Name:        name,
		Author:      author,
		GenreID:     genreID,
		Tags:        []string{},
		rating:      0.0,
		ratingCount: 0,
		downloads:   0,
		CreatedAt:   now,
		ModifiedAt:  now,
		BuildingDef: buildingDef,
	}
}
