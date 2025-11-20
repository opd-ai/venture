package housing

import (
	"fmt"
)

// Manager handles plot placement, validation, and persistence.
type Manager struct {
	plots       map[string]*Plot   // Plot ID -> Plot
	playerPlots map[string][]*Plot // Player ID -> Plots owned
	spatialGrid *SpatialGrid
	enabled     bool
}

// NewManager creates a new housing manager.
func NewManager() *Manager {
	return &Manager{
		plots:       make(map[string]*Plot),
		playerPlots: make(map[string][]*Plot),
		spatialGrid: NewSpatialGrid(64), // 64-unit cells
		enabled:     true,
	}
}

// SetEnabled enables or disables housing placement.
func (m *Manager) SetEnabled(enabled bool) {
	m.enabled = enabled
}

// IsEnabled returns whether housing placement is enabled.
func (m *Manager) IsEnabled() bool {
	return m.enabled
}

// PlacePlot attempts to place a plot in the world.
// Returns an error if placement is invalid.
func (m *Manager) PlacePlot(plot *Plot) error {
	if !m.enabled {
		return fmt.Errorf("housing is disabled")
	}

	if plot == nil {
		return fmt.Errorf("plot cannot be nil")
	}

	if plot.OwnerID == "" {
		return fmt.Errorf("plot must have an owner")
	}

	// Check for overlaps with existing plots
	if err := m.validatePlacement(plot); err != nil {
		return err
	}

	// Add to manager
	m.plots[plot.ID] = plot
	m.playerPlots[plot.OwnerID] = append(m.playerPlots[plot.OwnerID], plot)
	m.spatialGrid.Insert(plot)

	return nil
}

// RemovePlot removes a plot from the world.
func (m *Manager) RemovePlot(plotID string) error {
	plot, ok := m.plots[plotID]
	if !ok {
		return fmt.Errorf("plot %s not found", plotID)
	}

	// Remove from maps
	delete(m.plots, plotID)

	// Remove from player plots
	playerPlots := m.playerPlots[plot.OwnerID]
	for i, p := range playerPlots {
		if p.ID == plotID {
			m.playerPlots[plot.OwnerID] = append(playerPlots[:i], playerPlots[i+1:]...)
			break
		}
	}

	// Remove from spatial grid
	m.spatialGrid.Remove(plot)

	return nil
}

// GetPlot retrieves a plot by ID.
func (m *Manager) GetPlot(plotID string) (*Plot, bool) {
	plot, ok := m.plots[plotID]
	return plot, ok
}

// GetPlayerPlots returns all plots owned by a player.
func (m *Manager) GetPlayerPlots(playerID string) []*Plot {
	return m.playerPlots[playerID]
}

// GetPlotsInArea returns all plots that intersect a given area.
func (m *Manager) GetPlotsInArea(min, max Vector2) []*Plot {
	return m.spatialGrid.Query(min, max)
}

// GetAllPlots returns all plots in the world.
func (m *Manager) GetAllPlots() []*Plot {
	plots := make([]*Plot, 0, len(m.plots))
	for _, plot := range m.plots {
		plots = append(plots, plot)
	}
	return plots
}

// PlotCount returns the total number of plots.
func (m *Manager) PlotCount() int {
	return len(m.plots)
}

// validatePlacement checks if a plot can be placed without conflicts.
func (m *Manager) validatePlacement(plot *Plot) error {
	min, max := plot.Bounds()

	// Add 1-tile margin for spacing requirement
	const margin = 1.0
	min.X -= margin
	min.Y -= margin
	max.X += margin
	max.Y += margin

	// Check for overlaps with nearby plots
	nearbyPlots := m.spatialGrid.Query(min, max)
	for _, existing := range nearbyPlots {
		if plot.ID != existing.ID && plot.Overlaps(existing) {
			return fmt.Errorf("plot overlaps with existing plot %s at position (%.1f, %.1f)",
				existing.ID, existing.Position.X, existing.Position.Y)
		}
	}

	return nil
}

// Clear removes all plots from the manager.
func (m *Manager) Clear() {
	m.plots = make(map[string]*Plot)
	m.playerPlots = make(map[string][]*Plot)
	m.spatialGrid = NewSpatialGrid(64)
}

// CreateHouse creates a new housing plot from building data.
// This is a convenience method for creating plots from procedurally generated buildings.
func (m *Manager) CreateHouse(ownerID string, building interface{}) (string, error) {
	if !m.enabled {
		return "", fmt.Errorf("housing is disabled")
	}

	// Generate plot ID
	plotID := fmt.Sprintf("house_%s_%d", ownerID, len(m.playerPlots[ownerID]))

	// Create plot (simplified - real implementation would extract data from building)
	plot := &Plot{
		ID:       plotID,
		OwnerID:  ownerID,
		Position: Vector2{X: 0, Y: 0}, // TODO: Extract from building or use placement system
		Size:     SizeMedium,          // TODO: Extract from building dimensions
	}

	if err := m.PlacePlot(plot); err != nil {
		return "", fmt.Errorf("failed to place house plot: %w", err)
	}

	return plotID, nil
}

// GetHouse retrieves a house plot by ID.
func (m *Manager) GetHouse(houseID string) *House {
	plot, ok := m.plots[houseID]
	if !ok {
		return nil
	}

	// Convert plot to house (simplified representation)
	return &House{
		ID:      plot.ID,
		OwnerID: plot.OwnerID,
		Plot:    plot,
	}
}

// GetHouseFederated retrieves a house that may have originated from another server.
// serverID is the origin server identifier.
func (m *Manager) GetHouseFederated(houseID, serverID string) *House {
	// For now, just check local plots (federation sync would have replicated it)
	return m.GetHouse(houseID)
}

// SyncHouseFromFederation synchronizes a house from another federated server.
// serverID is the origin server, data contains the serialized house information.
func (m *Manager) SyncHouseFromFederation(serverID string, data []byte) error {
	// TODO: Deserialize house data
	// TODO: Create or update local plot representation
	// For now, return success (placeholder for actual implementation)
	return nil
}
