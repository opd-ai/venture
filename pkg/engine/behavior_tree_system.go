// Package engine provides the behavior tree system for AI.
// This file implements BehaviorTreeSystem which executes behavior trees for entities.
package engine

import (
	"github.com/sirupsen/logrus"
)

// BehaviorTreeSystem manages behavior tree execution for entities.
type BehaviorTreeSystem struct {
	world  *World
	logger *logrus.Entry
}

// NewBehaviorTreeSystem creates a new behavior tree system.
func NewBehaviorTreeSystem(world *World) *BehaviorTreeSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system", "behavior_tree")
		if logEntry.Logger.GetLevel() >= logrus.DebugLevel {
			logEntry.Debug("Behavior tree system created")
		}
	}
	return &BehaviorTreeSystem{
		world:  world,
		logger: logEntry,
	}
}

// Update executes behavior trees for all entities with behavior tree components.
func (bt *BehaviorTreeSystem) Update(entities []*Entity, deltaTime float64) {
	for _, entity := range entities {
		// Check if entity has behavior tree component
		treeComp, ok := entity.GetComponent("behaviortree")
		if !ok {
			continue
		}

		btComp := treeComp.(*BehaviorTreeComponent)
		if !btComp.Enabled {
			continue
		}

		// Execute the behavior tree
		status := btComp.Tick(entity, deltaTime)

		// Log status changes in debug mode
		if bt.logger != nil && bt.logger.Logger.GetLevel() >= logrus.DebugLevel {
			bt.logger.WithFields(logrus.Fields{
				"entity": entity.ID,
				"tree":   btComp.TreeName,
				"status": status.String(),
			}).Debug("Behavior tree tick")
		}
	}
}
