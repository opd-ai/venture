package housing

import (
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// GuildHallSize represents the size tier of a guild hall.
type GuildHallSize int

const (
	GuildHallSmall  GuildHallSize = 32 // 32×32 tiles
	GuildHallMedium GuildHallSize = 48 // 48×48 tiles
	GuildHallLarge  GuildHallSize = 64 // 64×64 tiles
)

// String returns a human-readable name for the guild hall size.
func (s GuildHallSize) String() string {
	switch s {
	case GuildHallSmall:
		return "Small"
	case GuildHallMedium:
		return "Medium"
	case GuildHallLarge:
		return "Large"
	default:
		return "Unknown"
	}
}

// Tiles returns the width/height in tiles for this guild hall size.
func (s GuildHallSize) Tiles() int {
	return int(s)
}

// ConstructionPhase represents the current phase of construction.
type ConstructionPhase int

const (
	PhaseFoundation ConstructionPhase = iota
	PhaseWalls
	PhaseRoof
	PhaseInterior
	PhaseComplete
)

// String returns a human-readable name for the construction phase.
func (p ConstructionPhase) String() string {
	switch p {
	case PhaseFoundation:
		return "Foundation"
	case PhaseWalls:
		return "Walls"
	case PhaseRoof:
		return "Roof"
	case PhaseInterior:
		return "Interior"
	case PhaseComplete:
		return "Complete"
	default:
		return "Unknown"
	}
}

// MaterialType represents different building materials.
type MaterialType int

const (
	MaterialWood MaterialType = iota
	MaterialStone
	MaterialMetal
	MaterialCrystal
)

// String returns a human-readable name for the material type.
func (m MaterialType) String() string {
	switch m {
	case MaterialWood:
		return "Wood"
	case MaterialStone:
		return "Stone"
	case MaterialMetal:
		return "Metal"
	case MaterialCrystal:
		return "Crystal"
	default:
		return "Unknown"
	}
}

// MaterialContribution tracks a player's contribution to construction.
type MaterialContribution struct {
	PlayerID     string
	MaterialType MaterialType
	Amount       int
	Timestamp    time.Time
}

// GuildHall represents a guild hall building.
type GuildHall struct {
	ID                string
	GuildID           string
	OwnerGuildName    string
	Position          Vector2
	Size              GuildHallSize
	Floors            int
	Theme             string
	Phase             ConstructionPhase
	Contributions     []MaterialContribution
	Materials         map[MaterialType]int // Current materials
	RequiredMaterials map[MaterialType]int // Materials needed for current phase
	CreatedAt         time.Time
	ModifiedAt        time.Time
	BuildingID        string
	mu                sync.RWMutex
	timeProvider      TimeProvider // Time provider for deterministic timestamps
}

// NewGuildHall creates a new guild hall construction project.
// Uses the default time provider for timestamps.
func NewGuildHall(guildID, guildName string, position Vector2, size GuildHallSize, floors int) *GuildHall {
	return NewGuildHallWithTime(guildID, guildName, position, size, floors, defaultTimeProvider)
}

// NewGuildHallWithTime creates a new guild hall with timestamps from the provided TimeProvider.
// Use this for deterministic timestamps in multiplayer synchronization and testing.
func NewGuildHallWithTime(guildID, guildName string, position Vector2, size GuildHallSize, floors int, tp TimeProvider) *GuildHall {
	now := tp.Now()
	gh := &GuildHall{
		ID:                generateGuildHallID(),
		GuildID:           guildID,
		OwnerGuildName:    guildName,
		Position:          position,
		Size:              size,
		Floors:            floors,
		Phase:             PhaseFoundation,
		Contributions:     []MaterialContribution{},
		Materials:         make(map[MaterialType]int),
		RequiredMaterials: make(map[MaterialType]int),
		CreatedAt:         now,
		ModifiedAt:        now,
		timeProvider:      tp,
	}
	gh.calculateRequiredMaterials()
	return gh
}

// AddMaterial adds materials to the construction, returns true if phase complete.
func (gh *GuildHall) AddMaterial(playerID string, materialType MaterialType, amount int) bool {
	gh.mu.Lock()
	defer gh.mu.Unlock()

	// Get time provider, default to real time if not set
	tp := gh.timeProvider
	if tp == nil {
		tp = defaultTimeProvider
	}

	// Record contribution
	gh.Contributions = append(gh.Contributions, MaterialContribution{
		PlayerID:     playerID,
		MaterialType: materialType,
		Amount:       amount,
		Timestamp:    tp.Now(),
	})

	// Add to materials
	gh.Materials[materialType] += amount
	gh.ModifiedAt = tp.Now()

	// Check if phase is complete
	return gh.checkPhaseComplete()
}

// checkPhaseComplete checks if the current phase has enough materials (must hold lock).
func (gh *GuildHall) checkPhaseComplete() bool {
	for mat, required := range gh.RequiredMaterials {
		if gh.Materials[mat] < required {
			return false
		}
	}
	return true
}

// AdvancePhase moves construction to the next phase.
func (gh *GuildHall) AdvancePhase() error {
	gh.mu.Lock()
	defer gh.mu.Unlock()

	if !gh.checkPhaseComplete() {
		return fmt.Errorf("insufficient materials for phase %s", gh.Phase)
	}

	// Consume materials
	for mat := range gh.RequiredMaterials {
		gh.Materials[mat] = 0
	}

	// Move to next phase
	if gh.Phase == PhaseComplete {
		return fmt.Errorf("guild hall already complete")
	}

	gh.Phase++

	// Get time provider, default to real time if not set
	tp := gh.timeProvider
	if tp == nil {
		tp = defaultTimeProvider
	}
	gh.ModifiedAt = tp.Now()

	// Calculate new requirements
	if gh.Phase != PhaseComplete {
		gh.calculateRequiredMaterials()
	} else {
		gh.RequiredMaterials = make(map[MaterialType]int)
	}

	return nil
}

// calculateRequiredMaterials determines materials needed for current phase (must hold lock).
func (gh *GuildHall) calculateRequiredMaterials() {
	gh.RequiredMaterials = make(map[MaterialType]int)

	// Base requirements scale with size and floors
	sizeMultiplier := gh.Size.Tiles() / 32 // 1x for Small, 1.5x for Medium, 2x for Large
	floorMultiplier := gh.Floors

	switch gh.Phase {
	case PhaseFoundation:
		gh.RequiredMaterials[MaterialStone] = 100 * sizeMultiplier * floorMultiplier
		gh.RequiredMaterials[MaterialMetal] = 50 * sizeMultiplier * floorMultiplier

	case PhaseWalls:
		gh.RequiredMaterials[MaterialStone] = 200 * sizeMultiplier * floorMultiplier
		gh.RequiredMaterials[MaterialWood] = 100 * sizeMultiplier * floorMultiplier

	case PhaseRoof:
		gh.RequiredMaterials[MaterialWood] = 150 * sizeMultiplier
		gh.RequiredMaterials[MaterialMetal] = 75 * sizeMultiplier

	case PhaseInterior:
		gh.RequiredMaterials[MaterialWood] = 100 * sizeMultiplier * floorMultiplier
		gh.RequiredMaterials[MaterialCrystal] = 50 * sizeMultiplier * floorMultiplier
	}
}

// GetProgress returns the construction progress as a percentage (0.0-1.0).
func (gh *GuildHall) GetProgress() float64 {
	gh.mu.RLock()
	defer gh.mu.RUnlock()

	if gh.Phase == PhaseComplete {
		return 1.0
	}

	// Calculate progress within current phase
	phaseWeight := 0.25 // Each phase is 25%
	phaseProgress := gh.calculatePhaseProgress()

	totalProgress := float64(gh.Phase)*phaseWeight + phaseProgress*phaseWeight
	return totalProgress
}

// calculatePhaseProgress calculates progress within the current phase (must hold lock).
func (gh *GuildHall) calculatePhaseProgress() float64 {
	if len(gh.RequiredMaterials) == 0 {
		return 1.0
	}

	totalRequired := 0
	totalCollected := 0

	for mat, required := range gh.RequiredMaterials {
		totalRequired += required
		collected := gh.Materials[mat]
		if collected > required {
			collected = required
		}
		totalCollected += collected
	}

	if totalRequired == 0 {
		return 1.0
	}

	return float64(totalCollected) / float64(totalRequired)
}

// GetMaterialProgress returns the progress for a specific material type.
func (gh *GuildHall) GetMaterialProgress(materialType MaterialType) (int, int, float64) {
	gh.mu.RLock()
	defer gh.mu.RUnlock()

	required := gh.RequiredMaterials[materialType]
	collected := gh.Materials[materialType]
	if collected > required {
		collected = required
	}

	progress := 0.0
	if required > 0 {
		progress = float64(collected) / float64(required)
	} else if gh.Phase == PhaseComplete {
		progress = 1.0
	}

	return collected, required, progress
}

// GetContributors returns a list of unique contributors.
func (gh *GuildHall) GetContributors() []string {
	gh.mu.RLock()
	defer gh.mu.RUnlock()

	contributors := make(map[string]bool)
	for _, contrib := range gh.Contributions {
		contributors[contrib.PlayerID] = true
	}

	result := make([]string, 0, len(contributors))
	for playerID := range contributors {
		result = append(result, playerID)
	}
	return result
}

// GetPlayerContribution returns the total contribution for a player.
func (gh *GuildHall) GetPlayerContribution(playerID string) map[MaterialType]int {
	gh.mu.RLock()
	defer gh.mu.RUnlock()

	contribution := make(map[MaterialType]int)
	for _, contrib := range gh.Contributions {
		if contrib.PlayerID == playerID {
			contribution[contrib.MaterialType] += contrib.Amount
		}
	}
	return contribution
}

// IsComplete returns true if the guild hall is fully constructed.
func (gh *GuildHall) IsComplete() bool {
	gh.mu.RLock()
	defer gh.mu.RUnlock()
	return gh.Phase == PhaseComplete
}

// GetCurrentPhase returns the current construction phase.
func (gh *GuildHall) GetCurrentPhase() ConstructionPhase {
	gh.mu.RLock()
	defer gh.mu.RUnlock()
	return gh.Phase
}

// Bounds returns the axis-aligned bounding box for this guild hall.
func (gh *GuildHall) Bounds() (min, max Vector2) {
	halfSize := float64(gh.Size.Tiles()) / 2.0
	min = Vector2{
		X: gh.Position.X - halfSize,
		Y: gh.Position.Y - halfSize,
	}
	max = Vector2{
		X: gh.Position.X + halfSize,
		Y: gh.Position.Y + halfSize,
	}
	return min, max
}

// MarshalJSON implements json.Marshaler (excludes mutex).
func (gh *GuildHall) MarshalJSON() ([]byte, error) {
	gh.mu.RLock()
	defer gh.mu.RUnlock()

	type Alias GuildHall
	return json.Marshal(&struct {
		*Alias
		mu interface{} `json:"-"`
	}{
		Alias: (*Alias)(gh),
	})
}

// guildHallIDCounter uses atomic operations for thread-safe ID generation
var guildHallIDCounter atomic.Int64

func generateGuildHallID() string {
	id := guildHallIDCounter.Add(1)
	return fmt.Sprintf("guildhall_%d", id)
}
