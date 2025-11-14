// Package network provides snapshot building utilities.
// This file implements snapshot creation with V4 component synchronization.
package network

import (
	"time"

	"github.com/opd-ai/venture/pkg/engine"
)

// SnapshotBuilder creates entity snapshots for network synchronization.
type SnapshotBuilder struct {
	serializer *ComponentSerializer
}

// NewSnapshotBuilder creates a new snapshot builder.
func NewSnapshotBuilder() *SnapshotBuilder {
	return &SnapshotBuilder{
		serializer: NewComponentSerializer(),
	}
}

// BuildEntitySnapshot creates an EntitySnapshot from an engine Entity.
// Serializes all network-synced components including V4.0 components:
// - Position, Velocity, Health, Stats, Team, Level (core)
// - Vehicle, Companion, Mount (V4.0 Phase 21-22)
// - MiniGame, Achievement (V4.0 Phase 26-27)
// - Expression (V4.0 Phase 26)
func (b *SnapshotBuilder) BuildEntitySnapshot(entity *engine.Entity, timestamp time.Time, sequence uint32) EntitySnapshot {
	snapshot := EntitySnapshot{
		EntityID:   entity.ID,
		Timestamp:  timestamp,
		Sequence:   sequence,
		Components: make(map[string][]byte),
	}

	// Core components
	if posComp, ok := entity.GetComponent("position"); ok {
		pos := posComp.(*engine.PositionComponent)
		snapshot.Position = Position{X: pos.X, Y: pos.Y}
		snapshot.Components["position"] = b.serializer.SerializePosition(pos.X, pos.Y)
	}

	if velComp, ok := entity.GetComponent("velocity"); ok {
		vel := velComp.(*engine.VelocityComponent)
		snapshot.Velocity = Velocity{VX: vel.VX, VY: vel.VY}
		snapshot.Components["velocity"] = b.serializer.SerializeVelocity(vel.VX, vel.VY)
	}

	if healthComp, ok := entity.GetComponent("health"); ok {
		health := healthComp.(*engine.HealthComponent)
		snapshot.Components["health"] = b.serializer.SerializeHealth(health.Current, health.Max)
	}

	if statsComp, ok := entity.GetComponent("stats"); ok {
		stats := statsComp.(*engine.StatsComponent)
		snapshot.Components["stats"] = b.serializer.SerializeStats(stats.Attack, stats.Defense, stats.MagicPower)
	}

	if teamComp, ok := entity.GetComponent("team"); ok {
		team := teamComp.(*engine.TeamComponent)
		snapshot.Components["team"] = b.serializer.SerializeTeam(uint64(team.TeamID))
	}

	if expComp, ok := entity.GetComponent("experience"); ok {
		exp := expComp.(*engine.ExperienceComponent)
		snapshot.Components["level"] = b.serializer.SerializeLevel(uint32(exp.Level), uint32(exp.CurrentXP))
	}

	// V4.0 Components (Phase 21-27)
	// These components use their built-in Serialize() methods

	// Vehicle component (Phase 21)
	if vehicleComp, ok := entity.GetComponent("vehicle"); ok {
		vehicle := vehicleComp.(*engine.VehicleComponent)
		snapshot.Components["vehicle"] = vehicle.Serialize()
	}

	// Companion component (Phase 22)
	if companionComp, ok := entity.GetComponent("companion"); ok {
		companion := companionComp.(*engine.CompanionComponent)
		snapshot.Components["companion"] = companion.Serialize()
	}

	// Mount component (Phase 21)
	if mountComp, ok := entity.GetComponent("mount"); ok {
		mount := mountComp.(*engine.MountComponent)
		snapshot.Components["mount"] = mount.Serialize()
	}

	// MiniGame component (Phase 27)
	// Note: MiniGameComponent has interface{} State field which cannot be easily serialized
	// Mini-games are typically single-player and don't need network sync
	// If multi-player mini-games are needed in future, implement custom serialization

	// Achievement component (Phase 27)
	if achievementComp, ok := entity.GetComponent("achievement"); ok {
		achievement := achievementComp.(*engine.AchievementComponent)
		snapshot.Components["achievement"] = achievement.Serialize()
	}

	// Expression component (Phase 26)
	if expressionComp, ok := entity.GetComponent("expression"); ok {
		expression := expressionComp.(*engine.ExpressionComponent)
		snapshot.Components["expression"] = b.serializer.SerializeExpression(
			uint8(expression.ActiveExpression),
			expression.ExpressionTime,
			expression.Cooldown,
		)
	}

	return snapshot
}

// BuildWorldSnapshot creates a WorldSnapshot from all entities in the world.
func (b *SnapshotBuilder) BuildWorldSnapshot(entities []*engine.Entity, timestamp time.Time, sequence uint32) WorldSnapshot {
	snapshot := WorldSnapshot{
		Timestamp: timestamp,
		Sequence:  sequence,
		Entities:  make(map[uint64]EntitySnapshot),
	}

	for _, entity := range entities {
		entitySnapshot := b.BuildEntitySnapshot(entity, timestamp, sequence)
		snapshot.Entities[entity.ID] = entitySnapshot
	}

	return snapshot
}

// ApplySnapshotToEntity updates an entity's components from a snapshot.
// Deserializes all network-synced components and applies them to the entity.
func (b *SnapshotBuilder) ApplySnapshotToEntity(entity *engine.Entity, snapshot EntitySnapshot) error {
	// Core components
	if posData, ok := snapshot.Components["position"]; ok {
		x, y, err := b.serializer.DeserializePosition(posData)
		if err != nil {
			return err
		}
		if posComp, ok := entity.GetComponent("position"); ok {
			pos := posComp.(*engine.PositionComponent)
			pos.X, pos.Y = x, y
		}
	}

	if velData, ok := snapshot.Components["velocity"]; ok {
		vx, vy, err := b.serializer.DeserializeVelocity(velData)
		if err != nil {
			return err
		}
		if velComp, ok := entity.GetComponent("velocity"); ok {
			vel := velComp.(*engine.VelocityComponent)
			vel.VX, vel.VY = vx, vy
		}
	}

	if healthData, ok := snapshot.Components["health"]; ok {
		current, max, err := b.serializer.DeserializeHealth(healthData)
		if err != nil {
			return err
		}
		if healthComp, ok := entity.GetComponent("health"); ok {
			health := healthComp.(*engine.HealthComponent)
			health.Current, health.Max = current, max
		}
	}

	if statsData, ok := snapshot.Components["stats"]; ok {
		attack, defense, magicPower, err := b.serializer.DeserializeStats(statsData)
		if err != nil {
			return err
		}
		if statsComp, ok := entity.GetComponent("stats"); ok {
			stats := statsComp.(*engine.StatsComponent)
			stats.Attack, stats.Defense, stats.MagicPower = attack, defense, magicPower
		}
	}

	if teamData, ok := snapshot.Components["team"]; ok {
		teamID, err := b.serializer.DeserializeTeam(teamData)
		if err != nil {
			return err
		}
		if teamComp, ok := entity.GetComponent("team"); ok {
			team := teamComp.(*engine.TeamComponent)
			team.TeamID = int(teamID)
		}
	}

	if levelData, ok := snapshot.Components["level"]; ok {
		level, xp, err := b.serializer.DeserializeLevel(levelData)
		if err != nil {
			return err
		}
		if expComp, ok := entity.GetComponent("experience"); ok {
			exp := expComp.(*engine.ExperienceComponent)
			exp.Level, exp.CurrentXP = int(level), int(xp)
		}
	}

	// V4.0 Components - use their built-in Deserialize methods

	if vehicleData, ok := snapshot.Components["vehicle"]; ok {
		if vehicleComp, ok := entity.GetComponent("vehicle"); ok {
			vehicle := vehicleComp.(*engine.VehicleComponent)
			if err := vehicle.Deserialize(vehicleData); err != nil {
				return err
			}
		}
	}

	if companionData, ok := snapshot.Components["companion"]; ok {
		if companionComp, ok := entity.GetComponent("companion"); ok {
			companion := companionComp.(*engine.CompanionComponent)
			if err := companion.Deserialize(companionData); err != nil {
				return err
			}
		}
	}

	if mountData, ok := snapshot.Components["mount"]; ok {
		if mountComp, ok := entity.GetComponent("mount"); ok {
			mount := mountComp.(*engine.MountComponent)
			if err := mount.Deserialize(mountData); err != nil {
				return err
			}
		}
	}

	// MiniGame component - skipped (see note in BuildEntitySnapshot)

	if achievementData, ok := snapshot.Components["achievement"]; ok {
		if achievementComp, ok := entity.GetComponent("achievement"); ok {
			achievement := achievementComp.(*engine.AchievementComponent)
			if err := achievement.Deserialize(achievementData); err != nil {
				return err
			}
		}
	}

	if expressionData, ok := snapshot.Components["expression"]; ok {
		expressionType, expressionTime, cooldown, err := b.serializer.DeserializeExpression(expressionData)
		if err != nil {
			return err
		}
		if expressionComp, ok := entity.GetComponent("expression"); ok {
			expression := expressionComp.(*engine.ExpressionComponent)
			expression.ActiveExpression = engine.ExpressionType(expressionType)
			expression.ExpressionTime = expressionTime
			expression.Cooldown = cooldown
		}
	}

	return nil
}
