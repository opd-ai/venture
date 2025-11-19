package housing

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"sync"
)

// GuildHallManager manages all guild halls in the world.
type GuildHallManager struct {
	guildHalls map[string]*GuildHall // guildID -> guild hall
	mu         sync.RWMutex
}

// NewGuildHallManager creates a new guild hall manager.
func NewGuildHallManager() *GuildHallManager {
	return &GuildHallManager{
		guildHalls: make(map[string]*GuildHall),
	}
}

// CreateGuildHall starts a new guild hall construction project.
func (m *GuildHallManager) CreateGuildHall(guildID, guildName string, position Vector2, size GuildHallSize, floors int) (*GuildHall, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Validate guild level determines max floors
	if floors < 1 || floors > 5 {
		return nil, fmt.Errorf("invalid floor count: %d (must be 1-5)", floors)
	}

	// Check if guild already has a hall
	if _, exists := m.guildHalls[guildID]; exists {
		return nil, fmt.Errorf("guild %s already has a guild hall", guildID)
	}

	// Check collision with existing guild halls
	halfSize := float64(size.Tiles()) / 2.0
	min := Vector2{X: position.X - halfSize, Y: position.Y - halfSize}
	max := Vector2{X: position.X + halfSize, Y: position.Y + halfSize}

	for _, existing := range m.guildHalls {
		existingMin, existingMax := existing.Bounds()
		// Check for overlap
		if !(max.X < existingMin.X || min.X > existingMax.X ||
			max.Y < existingMin.Y || min.Y > existingMax.Y) {
			return nil, fmt.Errorf("guild hall location overlaps with existing guild hall")
		}
	}

	// Create guild hall
	gh := NewGuildHall(guildID, guildName, position, size, floors)
	m.guildHalls[guildID] = gh

	return gh, nil
}

// GetGuildHall returns a guild hall by guild ID.
func (m *GuildHallManager) GetGuildHall(guildID string) (*GuildHall, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	gh, ok := m.guildHalls[guildID]
	return gh, ok
}

// ContributeMaterial adds materials to a guild hall construction.
func (m *GuildHallManager) ContributeMaterial(guildID, playerID string, materialType MaterialType, amount int) error {
	m.mu.RLock()
	gh, ok := m.guildHalls[guildID]
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("guild hall not found for guild %s", guildID)
	}

	if amount <= 0 {
		return fmt.Errorf("invalid material amount: %d", amount)
	}

	// Add material
	phaseComplete := gh.AddMaterial(playerID, materialType, amount)

	// Auto-advance phase if complete
	if phaseComplete {
		if err := gh.AdvancePhase(); err != nil {
			// Don't fail on advancement error, just log it
			// The materials are still added
		}
	}

	return nil
}

// AdvanceConstruction manually advances a guild hall to the next phase.
func (m *GuildHallManager) AdvanceConstruction(guildID string) error {
	m.mu.RLock()
	gh, ok := m.guildHalls[guildID]
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("guild hall not found for guild %s", guildID)
	}

	return gh.AdvancePhase()
}

// GetAllGuildHalls returns all guild halls (snapshot).
func (m *GuildHallManager) GetAllGuildHalls() []*GuildHall {
	m.mu.RLock()
	defer m.mu.RUnlock()

	halls := make([]*GuildHall, 0, len(m.guildHalls))
	for _, gh := range m.guildHalls {
		halls = append(halls, gh)
	}
	return halls
}

// GetGuildHallsInArea returns guild halls within a bounding box.
func (m *GuildHallManager) GetGuildHallsInArea(min, max Vector2) []*GuildHall {
	m.mu.RLock()
	defer m.mu.RUnlock()

	halls := make([]*GuildHall, 0)
	for _, gh := range m.guildHalls {
		ghMin, ghMax := gh.Bounds()
		// Check for overlap
		if !(ghMax.X < min.X || ghMin.X > max.X ||
			ghMax.Y < min.Y || ghMin.Y > max.Y) {
			halls = append(halls, gh)
		}
	}

	return halls
}

// RemoveGuildHall removes a guild hall (e.g., when guild disbands).
func (m *GuildHallManager) RemoveGuildHall(guildID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.guildHalls[guildID]; !ok {
		return fmt.Errorf("guild hall not found for guild %s", guildID)
	}

	delete(m.guildHalls, guildID)
	return nil
}

// GetStats returns statistics about guild halls.
func (m *GuildHallManager) GetStats() map[string]int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := map[string]int{
		"total":      len(m.guildHalls),
		"foundation": 0,
		"walls":      0,
		"roof":       0,
		"interior":   0,
		"complete":   0,
	}

	for _, gh := range m.guildHalls {
		phase := gh.GetCurrentPhase()
		switch phase {
		case PhaseFoundation:
			stats["foundation"]++
		case PhaseWalls:
			stats["walls"]++
		case PhaseRoof:
			stats["roof"]++
		case PhaseInterior:
			stats["interior"]++
		case PhaseComplete:
			stats["complete"]++
		}
	}

	return stats
}

// Save serializes all guild halls to gzip-compressed JSON.
func (m *GuildHallManager) Save(w io.Writer) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Create gzip writer
	gzWriter := gzip.NewWriter(w)
	defer gzWriter.Close()

	// Encode guild halls
	encoder := json.NewEncoder(gzWriter)
	return encoder.Encode(m.guildHalls)
}

// Load deserializes guild halls from gzip-compressed JSON.
func (m *GuildHallManager) Load(r io.Reader) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Create gzip reader
	gzReader, err := gzip.NewReader(r)
	if err != nil {
		return fmt.Errorf("gzip reader creation failed: %w", err)
	}
	defer gzReader.Close()

	// Decode guild halls
	guildHalls := make(map[string]*GuildHall)
	decoder := json.NewDecoder(gzReader)
	if err := decoder.Decode(&guildHalls); err != nil {
		return fmt.Errorf("guild halls decode failed: %w", err)
	}

	// Replace existing data
	m.guildHalls = guildHalls

	return nil
}

// ValidateProgress checks if a guild hall's progress is valid.
func (m *GuildHallManager) ValidateProgress(guildID string) error {
	m.mu.RLock()
	gh, ok := m.guildHalls[guildID]
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("guild hall not found for guild %s", guildID)
	}

	// Check progress is valid
	progress := gh.GetProgress()
	if progress < 0.0 || progress > 1.0 {
		return fmt.Errorf("invalid progress: %.2f", progress)
	}

	// Check phase consistency
	phase := gh.GetCurrentPhase()
	if phase < PhaseFoundation || phase > PhaseComplete {
		return fmt.Errorf("invalid phase: %d", phase)
	}

	// Check floors are valid
	if gh.Floors < 1 || gh.Floors > 5 {
		return fmt.Errorf("invalid floor count: %d", gh.Floors)
	}

	return nil
}
