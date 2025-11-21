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

	b.serializeCoreComponents(entity, &snapshot)
	b.serializeV4Components(entity, &snapshot)

	return snapshot
}

// serializeCoreComponents serializes core entity components.
func (b *SnapshotBuilder) serializeCoreComponents(entity *engine.Entity, snapshot *EntitySnapshot) {
	b.serializePosition(entity, snapshot)
	b.serializeVelocity(entity, snapshot)
	b.serializeHealth(entity, snapshot)
	b.serializeStats(entity, snapshot)
	b.serializeTeam(entity, snapshot)
	b.serializeLevel(entity, snapshot)
}

// serializeV4Components serializes V4.0 entity components.
func (b *SnapshotBuilder) serializeV4Components(entity *engine.Entity, snapshot *EntitySnapshot) {
	b.serializeVehicle(entity, snapshot)
	b.serializeCompanion(entity, snapshot)
	b.serializeMount(entity, snapshot)
	b.serializeAchievement(entity, snapshot)
	b.serializeExpression(entity, snapshot)
	b.serializeClassProgression(entity, snapshot)
}

// serializePosition serializes position component.
func (b *SnapshotBuilder) serializePosition(entity *engine.Entity, snapshot *EntitySnapshot) {
	if posComp, ok := entity.GetComponent("position"); ok {
		pos := posComp.(*engine.PositionComponent)
		snapshot.Position = Position{X: pos.X, Y: pos.Y}
		snapshot.Components["position"] = b.serializer.SerializePosition(pos.X, pos.Y)
	}
}

// serializeVelocity serializes velocity component.
func (b *SnapshotBuilder) serializeVelocity(entity *engine.Entity, snapshot *EntitySnapshot) {
	if velComp, ok := entity.GetComponent("velocity"); ok {
		vel := velComp.(*engine.VelocityComponent)
		snapshot.Velocity = Velocity{VX: vel.VX, VY: vel.VY}
		snapshot.Components["velocity"] = b.serializer.SerializeVelocity(vel.VX, vel.VY)
	}
}

// serializeHealth serializes health component.
func (b *SnapshotBuilder) serializeHealth(entity *engine.Entity, snapshot *EntitySnapshot) {
	if healthComp, ok := entity.GetComponent("health"); ok {
		health := healthComp.(*engine.HealthComponent)
		snapshot.Components["health"] = b.serializer.SerializeHealth(health.Current, health.Max)
	}
}

// serializeStats serializes stats component.
func (b *SnapshotBuilder) serializeStats(entity *engine.Entity, snapshot *EntitySnapshot) {
	if statsComp, ok := entity.GetComponent("stats"); ok {
		stats := statsComp.(*engine.StatsComponent)
		snapshot.Components["stats"] = b.serializer.SerializeStats(stats.Attack, stats.Defense, stats.MagicPower)
	}
}

// serializeTeam serializes team component.
func (b *SnapshotBuilder) serializeTeam(entity *engine.Entity, snapshot *EntitySnapshot) {
	if teamComp, ok := entity.GetComponent("team"); ok {
		team := teamComp.(*engine.TeamComponent)
		snapshot.Components["team"] = b.serializer.SerializeTeam(uint64(team.TeamID))
	}
}

// serializeLevel serializes experience component as level data.
func (b *SnapshotBuilder) serializeLevel(entity *engine.Entity, snapshot *EntitySnapshot) {
	if expComp, ok := entity.GetComponent("experience"); ok {
		exp := expComp.(*engine.ExperienceComponent)
		snapshot.Components["level"] = b.serializer.SerializeLevel(uint32(exp.Level), uint32(exp.CurrentXP))
	}
}

// serializeVehicle serializes vehicle component.
func (b *SnapshotBuilder) serializeVehicle(entity *engine.Entity, snapshot *EntitySnapshot) {
	if vehicleComp, ok := entity.GetComponent("vehicle"); ok {
		vehicle := vehicleComp.(*engine.VehicleComponent)
		snapshot.Components["vehicle"] = vehicle.Serialize()
	}
}

// serializeCompanion serializes companion component.
func (b *SnapshotBuilder) serializeCompanion(entity *engine.Entity, snapshot *EntitySnapshot) {
	if companionComp, ok := entity.GetComponent("companion"); ok {
		companion := companionComp.(*engine.CompanionComponent)
		snapshot.Components["companion"] = companion.Serialize()
	}
}

// serializeMount serializes mount component.
func (b *SnapshotBuilder) serializeMount(entity *engine.Entity, snapshot *EntitySnapshot) {
	if mountComp, ok := entity.GetComponent("mount"); ok {
		mount := mountComp.(*engine.MountComponent)
		snapshot.Components["mount"] = mount.Serialize()
	}
}

// serializeAchievement serializes achievement component.
func (b *SnapshotBuilder) serializeAchievement(entity *engine.Entity, snapshot *EntitySnapshot) {
	if achievementComp, ok := entity.GetComponent("achievement"); ok {
		achievement := achievementComp.(*engine.AchievementComponent)
		snapshot.Components["achievement"] = achievement.Serialize()
	}
}

// serializeExpression serializes expression component.
func (b *SnapshotBuilder) serializeExpression(entity *engine.Entity, snapshot *EntitySnapshot) {
	if expressionComp, ok := entity.GetComponent("expression"); ok {
		expression := expressionComp.(*engine.ExpressionComponent)
		snapshot.Components["expression"] = expression.Serialize()
	}
}

// serializeClassProgression serializes class progression component.
func (b *SnapshotBuilder) serializeClassProgression(entity *engine.Entity, snapshot *EntitySnapshot) {
	if classComp, ok := entity.GetComponent("class_progression"); ok {
		class := classComp.(*engine.ClassProgressionComponent)
		snapshot.Components["class_progression"] = class.Serialize()
	}
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
	if err := b.applyCoreComponents(entity, snapshot); err != nil {
		return err
	}
	if err := b.applyV4Components(entity, snapshot); err != nil {
		return err
	}
	return nil
}

// applyCoreComponents applies core component snapshots to an entity.
func (b *SnapshotBuilder) applyCoreComponents(entity *engine.Entity, snapshot EntitySnapshot) error {
	if err := b.applyPositionComponent(entity, snapshot); err != nil {
		return err
	}
	if err := b.applyVelocityComponent(entity, snapshot); err != nil {
		return err
	}
	if err := b.applyHealthComponent(entity, snapshot); err != nil {
		return err
	}
	if err := b.applyStatsComponent(entity, snapshot); err != nil {
		return err
	}
	if err := b.applyTeamComponent(entity, snapshot); err != nil {
		return err
	}
	if err := b.applyLevelComponent(entity, snapshot); err != nil {
		return err
	}
	return nil
}

// applyV4Components applies V4.0 component snapshots to an entity.
func (b *SnapshotBuilder) applyV4Components(entity *engine.Entity, snapshot EntitySnapshot) error {
	if err := b.applyVehicleComponent(entity, snapshot); err != nil {
		return err
	}
	if err := b.applyCompanionComponent(entity, snapshot); err != nil {
		return err
	}
	if err := b.applyMountComponent(entity, snapshot); err != nil {
		return err
	}
	if err := b.applyAchievementComponent(entity, snapshot); err != nil {
		return err
	}
	if err := b.applyExpressionComponent(entity, snapshot); err != nil {
		return err
	}
	if err := b.applyClassProgressionComponent(entity, snapshot); err != nil {
		return err
	}
	return nil
}

// applyPositionComponent updates position component from snapshot data.
func (b *SnapshotBuilder) applyPositionComponent(entity *engine.Entity, snapshot EntitySnapshot) error {
	posData, ok := snapshot.Components["position"]
	if !ok {
		return nil
	}
	x, y, err := b.serializer.DeserializePosition(posData)
	if err != nil {
		return err
	}
	if posComp, ok := entity.GetComponent("position"); ok {
		pos := posComp.(*engine.PositionComponent)
		pos.X, pos.Y = x, y
	}
	return nil
}

// applyVelocityComponent updates velocity component from snapshot data.
func (b *SnapshotBuilder) applyVelocityComponent(entity *engine.Entity, snapshot EntitySnapshot) error {
	velData, ok := snapshot.Components["velocity"]
	if !ok {
		return nil
	}
	vx, vy, err := b.serializer.DeserializeVelocity(velData)
	if err != nil {
		return err
	}
	if velComp, ok := entity.GetComponent("velocity"); ok {
		vel := velComp.(*engine.VelocityComponent)
		vel.VX, vel.VY = vx, vy
	}
	return nil
}

// applyHealthComponent updates health component from snapshot data.
func (b *SnapshotBuilder) applyHealthComponent(entity *engine.Entity, snapshot EntitySnapshot) error {
	healthData, ok := snapshot.Components["health"]
	if !ok {
		return nil
	}
	current, max, err := b.serializer.DeserializeHealth(healthData)
	if err != nil {
		return err
	}
	if healthComp, ok := entity.GetComponent("health"); ok {
		health := healthComp.(*engine.HealthComponent)
		health.Current, health.Max = current, max
	}
	return nil
}

// applyStatsComponent updates stats component from snapshot data.
func (b *SnapshotBuilder) applyStatsComponent(entity *engine.Entity, snapshot EntitySnapshot) error {
	statsData, ok := snapshot.Components["stats"]
	if !ok {
		return nil
	}
	attack, defense, magicPower, err := b.serializer.DeserializeStats(statsData)
	if err != nil {
		return err
	}
	if statsComp, ok := entity.GetComponent("stats"); ok {
		stats := statsComp.(*engine.StatsComponent)
		stats.Attack, stats.Defense, stats.MagicPower = attack, defense, magicPower
	}
	return nil
}

// applyTeamComponent updates team component from snapshot data.
func (b *SnapshotBuilder) applyTeamComponent(entity *engine.Entity, snapshot EntitySnapshot) error {
	teamData, ok := snapshot.Components["team"]
	if !ok {
		return nil
	}
	teamID, err := b.serializer.DeserializeTeam(teamData)
	if err != nil {
		return err
	}
	if teamComp, ok := entity.GetComponent("team"); ok {
		team := teamComp.(*engine.TeamComponent)
		team.TeamID = int(teamID)
	}
	return nil
}

// applyLevelComponent updates experience component from snapshot data.
func (b *SnapshotBuilder) applyLevelComponent(entity *engine.Entity, snapshot EntitySnapshot) error {
	levelData, ok := snapshot.Components["level"]
	if !ok {
		return nil
	}
	level, xp, err := b.serializer.DeserializeLevel(levelData)
	if err != nil {
		return err
	}
	if expComp, ok := entity.GetComponent("experience"); ok {
		exp := expComp.(*engine.ExperienceComponent)
		exp.Level, exp.CurrentXP = int(level), int(xp)
	}
	return nil
}

// applyVehicleComponent updates vehicle component from snapshot data.
func (b *SnapshotBuilder) applyVehicleComponent(entity *engine.Entity, snapshot EntitySnapshot) error {
	vehicleData, ok := snapshot.Components["vehicle"]
	if !ok {
		return nil
	}
	if vehicleComp, ok := entity.GetComponent("vehicle"); ok {
		vehicle := vehicleComp.(*engine.VehicleComponent)
		if err := vehicle.Deserialize(vehicleData); err != nil {
			return err
		}
	}
	return nil
}

// applyCompanionComponent updates companion component from snapshot data.
func (b *SnapshotBuilder) applyCompanionComponent(entity *engine.Entity, snapshot EntitySnapshot) error {
	companionData, ok := snapshot.Components["companion"]
	if !ok {
		return nil
	}
	if companionComp, ok := entity.GetComponent("companion"); ok {
		companion := companionComp.(*engine.CompanionComponent)
		if err := companion.Deserialize(companionData); err != nil {
			return err
		}
	}
	return nil
}

// applyMountComponent updates mount component from snapshot data.
func (b *SnapshotBuilder) applyMountComponent(entity *engine.Entity, snapshot EntitySnapshot) error {
	mountData, ok := snapshot.Components["mount"]
	if !ok {
		return nil
	}
	if mountComp, ok := entity.GetComponent("mount"); ok {
		mount := mountComp.(*engine.MountComponent)
		if err := mount.Deserialize(mountData); err != nil {
			return err
		}
	}
	return nil
}

// applyAchievementComponent updates achievement component from snapshot data.
func (b *SnapshotBuilder) applyAchievementComponent(entity *engine.Entity, snapshot EntitySnapshot) error {
	achievementData, ok := snapshot.Components["achievement"]
	if !ok {
		return nil
	}
	if achievementComp, ok := entity.GetComponent("achievement"); ok {
		achievement := achievementComp.(*engine.AchievementComponent)
		if err := achievement.Deserialize(achievementData); err != nil {
			return err
		}
	}
	return nil
}

// applyExpressionComponent updates expression component from snapshot data.
func (b *SnapshotBuilder) applyExpressionComponent(entity *engine.Entity, snapshot EntitySnapshot) error {
	expressionData, ok := snapshot.Components["expression"]
	if !ok {
		return nil
	}
	if expressionComp, ok := entity.GetComponent("expression"); ok {
		expression := expressionComp.(*engine.ExpressionComponent)
		if err := expression.Deserialize(expressionData); err != nil {
			return err
		}
	}
	return nil
}

// applyClassProgressionComponent updates class progression component from snapshot data.
func (b *SnapshotBuilder) applyClassProgressionComponent(entity *engine.Entity, snapshot EntitySnapshot) error {
	classData, ok := snapshot.Components["class_progression"]
	if !ok {
		return nil
	}
	if classComp, ok := entity.GetComponent("class_progression"); ok {
		class := classComp.(*engine.ClassProgressionComponent)
		if err := class.Deserialize(classData); err != nil {
			return err
		}
	}
	return nil
}
