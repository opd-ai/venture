package political_warfare

import (
	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/network/federation/guild"
	"github.com/sirupsen/logrus"
)

// System wraps Manager as an ECS system for integration into the game world
type System struct {
	world   *engine.World
	manager *Manager
	logger  *logrus.Entry
}

// NewSystem creates a new political warfare system
func NewSystem(world *engine.World, guildManager *guild.Manager) *System {
	logger := logrus.WithField("system", "political_warfare")
	return &System{
		world:   world,
		manager: NewManager(world, guildManager),
		logger:  logger,
	}
}

// Update processes time-based state changes for wars and treaties
func (s *System) Update(entities []*engine.Entity, deltaTime float64) {
	s.manager.Update(deltaTime)
}

// GetManager returns the underlying political warfare manager for direct access
func (s *System) GetManager() *Manager {
	return s.manager
}
