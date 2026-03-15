package housing

import (
	"encoding/json"
	"fmt"
	"math/rand"

	log "github.com/sirupsen/logrus"
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
		log.WithFields(log.Fields{
			"error": "housing is disabled",
		}).Warn("failed to place plot: housing disabled")
		return fmt.Errorf("housing is disabled")
	}

	if plot == nil {
		log.WithFields(log.Fields{
			"error": "plot cannot be nil",
		}).Warn("failed to place plot: nil plot")
		return fmt.Errorf("plot cannot be nil")
	}

	if plot.OwnerID == "" {
		log.WithFields(log.Fields{
			"plotID": plot.ID,
			"error":  "plot must have an owner",
		}).Warn("failed to place plot: missing owner")
		return fmt.Errorf("plot must have an owner")
	}

	// Check for overlaps with existing plots
	if err := m.validatePlacement(plot); err != nil {
		log.WithFields(log.Fields{
			"plotID":  plot.ID,
			"ownerID": plot.OwnerID,
			"error":   err.Error(),
		}).Warn("failed to place plot: invalid placement")
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
		log.WithFields(log.Fields{
			"plotID": plotID,
			"error":  "plot not found",
		}).Warn("failed to remove plot")
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
// buildingData provides optional width/height dimensions and position override.
// Pass nil to use default sizing (SizeMedium) and seed-based positioning.
func (m *Manager) CreateHouse(ownerID string, buildingData *HousingBuildingData, seed int64) (string, error) {
	if !m.enabled {
		return "", fmt.Errorf("housing is disabled")
	}

	plotID := fmt.Sprintf("house_%s_%d", ownerID, len(m.playerPlots[ownerID]))
	size := m.determinePlotSizeFromData(buildingData)
	position := m.resolvePlotPosition(ownerID, seed, size, buildingData)

	plot := &Plot{
		ID:       plotID,
		OwnerID:  ownerID,
		Position: position,
		Size:     size,
	}

	if err := m.PlacePlot(plot); err != nil {
		return "", fmt.Errorf("failed to place house plot: %w", err)
	}

	return plotID, nil
}

// determinePlotSize determines the plot size based on building dimensions.
// Returns SizeMedium if buildingData is nil or doesn't implement BuildingInterface.
// determinePlotSizeFromData determines the plot size from a HousingBuildingData.
func (m *Manager) determinePlotSizeFromData(data *HousingBuildingData) BuildingSize {
	if data == nil || (data.Width == 0 && data.Height == 0) {
		return SizeMedium
	}
	if data.Width <= 8 && data.Height <= 8 {
		return SizeSmall
	} else if data.Width <= 16 && data.Height <= 16 {
		return SizeMedium
	} else if data.Width <= 24 && data.Height <= 24 {
		return SizeLarge
	}
	return SizeEstate
}

// resolvePlotPosition returns either the explicit position from buildingData or
// a deterministically generated one based on seed.
func (m *Manager) resolvePlotPosition(ownerID string, seed int64, size BuildingSize, data *HousingBuildingData) Vector2 {
	if data != nil && data.Position != nil {
		return *data.Position
	}
	return m.generatePlotPosition(ownerID, seed, size)
}

// generatePlotPosition generates a deterministic position for a house plot.
// Uses grid-based distribution with randomized offsets within cells.
func (m *Manager) generatePlotPosition(ownerID string, seed int64, size BuildingSize) Vector2 {
	// Note: math/rand is appropriate here for deterministic procedural generation.
	// The goal is reproducibility from the same seed, not cryptographic security.
	// Each call creates a new rand.Rand instance, making this safe for concurrent use
	// as the instance is not shared between goroutines.
	rng := rand.New(rand.NewSource(seed + int64(len(m.playerPlots[ownerID]))))

	const gridSpacing = 64.0
	plotCount := len(m.playerPlots[ownerID])
	gridX := float64(plotCount % 10) // 10 houses per row
	gridY := float64(plotCount / 10)

	offsetX := rng.Float64() * (gridSpacing - float64(size))
	offsetY := rng.Float64() * (gridSpacing - float64(size))

	return Vector2{
		X: gridX*gridSpacing + offsetX,
		Y: gridY*gridSpacing + offsetY,
	}
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
	if !m.enabled {
		return fmt.Errorf("housing is disabled")
	}

	// Deserialize plot data from federation message
	var plot Plot
	if err := json.Unmarshal(data, &plot); err != nil {
		return fmt.Errorf("failed to deserialize house data: %w", err)
	}

	// Reconstruct permission set if nil
	if plot.Permissions == nil {
		plot.Permissions = NewPermissionSet()
	}

	// Check if plot already exists (update case)
	if existing, ok := m.plots[plot.ID]; ok {
		// Update existing plot
		existing.Position = plot.Position
		existing.Size = plot.Size
		existing.Permissions = plot.Permissions
		m.spatialGrid.Update(existing)
		return nil
	}

	// Create new local plot representation (sync case)
	m.plots[plot.ID] = &plot
	m.playerPlots[plot.OwnerID] = append(m.playerPlots[plot.OwnerID], &plot)
	m.spatialGrid.Insert(&plot)

	return nil
}
