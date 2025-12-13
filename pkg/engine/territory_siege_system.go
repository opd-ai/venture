package engine

import (
	"github.com/opd-ai/venture/pkg/world/territory"
)

// TerritorySiegeSystem wraps the territory.SiegeManager for ECS integration.
type TerritorySiegeSystem struct {
	manager *territory.SiegeManager
}

// NewTerritorySiegeSystem creates a new territory siege system wrapper.
func NewTerritorySiegeSystem(manager *territory.SiegeManager) *TerritorySiegeSystem {
	return &TerritorySiegeSystem{
		manager: manager,
	}
}

// Update processes all active sieges.
func (s *TerritorySiegeSystem) Update(entities []*Entity, deltaTime float64) {
	s.manager.Update(deltaTime)
}

// GetSiegeManager returns the underlying siege manager for direct access.
func (s *TerritorySiegeSystem) GetSiegeManager() *territory.SiegeManager {
	return s.manager
}
