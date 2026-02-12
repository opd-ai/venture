package engine

import (
	"fmt"
	"time"

	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/opd-ai/venture/pkg/world/raids"
)

// RaidSystem manages raid instance lifecycle, boss mechanics, and player lockouts.
// It handles raid creation, boss mechanic execution, phase transitions, and cleanup.
type RaidSystem struct {
	world           *World
	generator       *raids.Generator
	instanceManager *raids.InstanceManager
	lockoutManager  *raids.LockoutManager
	worldSeed       int64
	cleanupInterval float64
	cleanupTimer    float64
}

// NewRaidSystem creates a new raid system with generators and managers.
func NewRaidSystem(world *World, worldSeed int64) *RaidSystem {
	return &RaidSystem{
		world:           world,
		generator:       raids.NewGenerator(worldSeed),
		instanceManager: raids.NewInstanceManager(),
		lockoutManager:  raids.NewLockoutManager(),
		worldSeed:       worldSeed,
		cleanupInterval: 60.0, // Cleanup every 60 seconds
		cleanupTimer:    0.0,
	}
}

// Update processes raid instance cleanup and lockout resets.
func (s *RaidSystem) Update(entities []*Entity, deltaTime float64) {
	// Update boss mechanics
	s.updateBossMechanics(deltaTime)

	// Check boss phases
	s.updateBossPhases()

	// Periodic cleanup of expired instances
	s.cleanupTimer += deltaTime
	if s.cleanupTimer >= s.cleanupInterval {
		s.cleanupExpiredInstances()
		s.cleanupTimer = 0.0
	}
}

// CreateRaidInstance generates a new raid dungeon and spawns the entrance portal.
// Returns the instance ID and entrance entity.
func (s *RaidSystem) CreateRaidInstance(tier raids.RaidTier, groupID string, playerIDs []string, genreID string, seed int64) (string, *Entity, error) {
	// Generate raid dungeon
	params := procgen.GenerationParams{
		Difficulty: tier.DifficultyMultiplier() / 10.0, // Scale 2.0-10.0 to 0.2-1.0
		Depth:      15 + int(tier)*5,                   // 15-35 depth based on tier
		GenreID:    genreID,
		Custom: map[string]interface{}{
			"group_size": len(playerIDs),
		},
	}

	result, err := s.generator.Generate(seed, params)
	if err != nil {
		return "", nil, fmt.Errorf("raid generation failed: %w", err)
	}

	raidDungeon := result.(*raids.RaidDungeon)

	// Create instance
	instance, err := s.instanceManager.CreateInstance(raidDungeon, groupID, playerIDs)
	if err != nil {
		return "", nil, fmt.Errorf("instance creation failed: %w", err)
	}

	// Create entrance portal entity
	entrance := s.world.CreateEntity()
	entrance.AddComponent(&PositionComponent{X: 100, Y: 100})
	entrance.AddComponent(&RaidInstanceComponent{
		InstanceID:   instance.InstanceID,
		RaidDungeon:  raidDungeon,
		Tier:         tier,
		GroupID:      groupID,
		PlayerIDs:    playerIDs,
		CreatedAt:    instance.CreatedAt,
		ExpiresAt:    instance.ExpiresAt,
		Completed:    false,
		ActiveBoss:   0,
		BossesKilled: make([]bool, len(raidDungeon.Bosses)),
	})
	entrance.AddComponent(&EbitenSprite{
		Width:   64,
		Height:  64,
		Visible: true,
		Layer:   1,
	})
	entrance.AddComponent(&ColliderComponent{
		Width:     64,
		Height:    64,
		Solid:     false,
		IsTrigger: true,
	})

	return instance.InstanceID, entrance, nil
}

// EnterRaidInstance checks lockouts and teleports players into the raid.
func (s *RaidSystem) EnterRaidInstance(playerEntity, entranceEntity *Entity) error {
	instanceCompRaw, ok := entranceEntity.GetComponent("raid_instance")
	if !ok {
		return fmt.Errorf("entrance entity has no raid instance component")
	}
	instance := instanceCompRaw.(*RaidInstanceComponent)

	// Check player lockout
	lockoutCompRaw, ok := playerEntity.GetComponent("raid_lockout")
	if ok {
		lockout := lockoutCompRaw.(*RaidLockoutComponent)
		if lockout.IsLockedOut(instance.Tier) {
			return fmt.Errorf("player is locked out of %s tier", instance.Tier)
		}
	} else {
		// Create lockout component if player doesn't have one
		playerEntity.AddComponent(NewRaidLockoutComponent())
	}

	// Teleport player to entrance
	posCompRaw, ok := playerEntity.GetComponent("position")
	if ok {
		pos := posCompRaw.(*PositionComponent)
		// Set position to raid entrance (first room in dungeon)
		if len(instance.RaidDungeon.Rooms) > 0 {
			entranceRoom := instance.RaidDungeon.Rooms[0]
			pos.X = float64(entranceRoom.X + entranceRoom.W/2)
			pos.Y = float64(entranceRoom.Y + entranceRoom.H/2)
		}
	}

	return nil
}

// CompleteRaidInstance marks the instance complete and sets player lockouts.
func (s *RaidSystem) CompleteRaidInstance(instanceID string, playerEntities []*Entity) error {
	instance, err := s.findRaidInstance(instanceID)
	if err != nil {
		return err
	}

	instance.Completed = true

	s.applyPlayerLockouts(instance, playerEntities)

	return nil
}

// findRaidInstance locates a raid instance by ID.
func (s *RaidSystem) findRaidInstance(instanceID string) (*RaidInstanceComponent, error) {
	entities := s.world.GetEntitiesWith("raid_instance")

	for _, entity := range entities {
		compRaw, ok := entity.GetComponent("raid_instance")
		if !ok {
			continue
		}
		comp := compRaw.(*RaidInstanceComponent)
		if comp.InstanceID == instanceID {
			return comp, nil
		}
	}

	return nil, fmt.Errorf("instance %s not found", instanceID)
}

// applyPlayerLockouts sets raid lockouts for all players in the instance.
func (s *RaidSystem) applyPlayerLockouts(instance *RaidInstanceComponent, playerEntities []*Entity) {
	for _, player := range playerEntities {
		lockout := s.getOrCreateLockout(player)
		playerID := s.extractPlayerID(player)
		lockout.SetLockout(playerID, instance.Tier)
	}
}

// getOrCreateLockout retrieves or creates a raid lockout component for a player.
func (s *RaidSystem) getOrCreateLockout(player *Entity) *RaidLockoutComponent {
	lockoutCompRaw, ok := player.GetComponent("raid_lockout")
	if !ok {
		lockout := NewRaidLockoutComponent()
		player.AddComponent(lockout)
		return lockout
	}
	return lockoutCompRaw.(*RaidLockoutComponent)
}

// extractPlayerID determines the player ID from entity components.
func (s *RaidSystem) extractPlayerID(player *Entity) string {
	playerID := fmt.Sprintf("player_%d", player.ID) // Fallback to entity ID

	networkCompRaw, ok := player.GetComponent("network")
	if ok {
		networkComp := networkCompRaw.(*NetworkComponent)
		if networkComp.PlayerID != 0 {
			playerID = fmt.Sprintf("player_%d", networkComp.PlayerID)
		}
	}

	return playerID
}

// updateBossMechanics processes boss ability cooldowns and executes ready mechanics.
func (s *RaidSystem) updateBossMechanics(deltaTime float64) {
	entities := s.world.GetEntitiesWith("raid_boss", "health")

	for _, entity := range entities {
		s.updateBossEntity(entity, deltaTime)
	}
}

// updateBossEntity updates mechanics for a single boss entity.
func (s *RaidSystem) updateBossEntity(entity *Entity, deltaTime float64) {
	bossComp, healthComp, ok := s.getBossComponents(entity)
	if !ok {
		return
	}

	s.updateMechanicTimers(bossComp, deltaTime)
	s.executeReadyMechanics(entity, bossComp, healthComp)
}

// getBossComponents retrieves boss and health components from an entity.
func (s *RaidSystem) getBossComponents(entity *Entity) (*RaidBossComponent, *HealthComponent, bool) {
	bossCompRaw, ok := entity.GetComponent("raid_boss")
	if !ok {
		return nil, nil, false
	}
	bossComp := bossCompRaw.(*RaidBossComponent)

	healthCompRaw, ok := entity.GetComponent("health")
	if !ok {
		return nil, nil, false
	}
	healthComp := healthCompRaw.(*HealthComponent)

	return bossComp, healthComp, true
}

// updateMechanicTimers decrements all active mechanic timers.
func (s *RaidSystem) updateMechanicTimers(bossComp *RaidBossComponent, deltaTime float64) {
	for mechanicID, timer := range bossComp.MechanicTimer {
		if timer > 0 {
			bossComp.MechanicTimer[mechanicID] = timer - deltaTime
		}
	}
}

// executeReadyMechanics executes mechanics that are off cooldown.
func (s *RaidSystem) executeReadyMechanics(entity *Entity, bossComp *RaidBossComponent, healthComp *HealthComponent) {
	for i, mechanic := range bossComp.Mechanics {
		mechanicID := fmt.Sprintf("mechanic_%d", i)
		timer, exists := bossComp.MechanicTimer[mechanicID]

		if !exists {
			bossComp.MechanicTimer[mechanicID] = 0
			continue
		}

		if timer <= 0 && healthComp.Current > 0 {
			s.executeBossMechanic(entity, &mechanic)
			bossComp.MechanicTimer[mechanicID] = mechanic.Cooldown.Seconds()
		}
	}
}

// executeBossMechanic performs a boss mechanic action.
func (s *RaidSystem) executeBossMechanic(boss *Entity, mechanic *raids.BossMechanic) {
	// Get boss position
	posCompRaw, ok := boss.GetComponent("position")
	if !ok {
		return
	}
	pos := posCompRaw.(*PositionComponent)

	switch mechanic.Type {
	case raids.MechanicSummon:
		// Spawn adds near boss
		s.spawnBossAdd(pos.X, pos.Y, mechanic.Damage)

	case raids.MechanicGroundEffect:
		// Create hazard zone
		s.createGroundEffect(pos.X, pos.Y, mechanic.Radius, mechanic.Damage)

	case raids.MechanicDebuff:
		// Apply debuff to nearby players
		s.applyNearbyDebuff(pos.X, pos.Y, mechanic.Radius)

	case raids.MechanicInstant:
		// Instant damage to nearby players
		s.applyAoEDamage(pos.X, pos.Y, mechanic.Radius, mechanic.Damage)
	}
}

// updateBossPhases checks boss health and transitions to new phases.
func (s *RaidSystem) updateBossPhases() {
	entities := s.world.GetEntitiesWith("raid_boss", "health")

	for _, entity := range entities {
		bossComp, healthComp, ok := s.getBossComponents(entity)
		if !ok {
			continue
		}

		healthPercent := healthComp.Current / healthComp.Max
		s.checkPhaseTransitions(entity, bossComp, healthPercent)
	}
}

// checkPhaseTransitions triggers boss phase transitions based on health thresholds.
func (s *RaidSystem) checkPhaseTransitions(entity *Entity, bossComp *RaidBossComponent, healthPercent float64) {
	for _, phase := range bossComp.Phases {
		if s.shouldTransitionToPhase(phase, bossComp, healthPercent) {
			s.transitionBossToPhase(entity, bossComp, phase)
		}
	}
}

// shouldTransitionToPhase determines if boss should transition to a new phase.
func (s *RaidSystem) shouldTransitionToPhase(phase raids.BossPhase, bossComp *RaidBossComponent, healthPercent float64) bool {
	return phase.Number > bossComp.CurrentPhase && healthPercent <= phase.HealthThresh
}

// transitionBossToPhase executes the transition to a new boss phase.
func (s *RaidSystem) transitionBossToPhase(entity *Entity, bossComp *RaidBossComponent, phase raids.BossPhase) {
	bossComp.CurrentPhase = phase.Number
	bossComp.PhaseEntered = true

	if phase.AddSpawns > 0 {
		s.spawnPhaseAdds(entity, phase.AddSpawns)
	}
}

// spawnPhaseAdds spawns additional enemies when boss enters new phase.
func (s *RaidSystem) spawnPhaseAdds(entity *Entity, addCount int) {
	posCompRaw, ok := entity.GetComponent("position")
	if !ok {
		return
	}
	posComp := posCompRaw.(*PositionComponent)
	for i := 0; i < addCount; i++ {
		offset := float64(i * 50)
		s.spawnBossAdd(posComp.X+offset, posComp.Y, 100)
	}
}

// cleanupExpiredInstances removes raid instances past their expiration time.
// Note: Uses time.Now() for instance expiration - this is intentional for real-time cleanup.
func (s *RaidSystem) cleanupExpiredInstances() {
	entities := s.world.GetEntitiesWith("raid_instance")
	now := time.Now()

	for _, entity := range entities {
		instanceRaw, ok := entity.GetComponent("raid_instance")
		if !ok {
			continue
		}
		instance := instanceRaw.(*RaidInstanceComponent)

		if now.After(instance.ExpiresAt) || instance.Completed {
			// Remove instance entity
			s.world.RemoveEntity(entity.ID)

			// Clean up instance manager
			s.instanceManager.RemoveInstance(instance.InstanceID)
		}
	}
}

// Helper methods for mechanic execution

// spawnBossAdd creates a hostile add entity near the boss position.
func (s *RaidSystem) spawnBossAdd(x, y float64, damage int) {
	add := s.world.CreateEntity()
	add.AddComponent(&PositionComponent{X: x, Y: y})
	add.AddComponent(&HealthComponent{Current: 500, Max: 500})
	add.AddComponent(&AttackComponent{
		Damage:        float64(damage),
		DamageType:    0, // Default damage type
		Range:         50.0,
		Cooldown:      1.5,
		CooldownTimer: 0,
	})
	add.AddComponent(&AIComponent{
		State:                AIStateIdle,
		DetectionRange:       200,
		FleeHealthThreshold:  0.0,
		MaxChaseDistance:     0,
		DecisionTimer:        0,
		DecisionInterval:     0.5,
		StateTimer:           0,
		PatrolSpeed:          50,
		ChaseSpeed:           100,
		FleeSpeed:            150,
		ReturnSpeed:          75,
		PatrolWaypoints:      nil,
		CurrentWaypointIndex: 0,
		PatrolReverse:        false,
		PatrolDirection:      1,
	})
	// Note: Sprite rendering is handled by the render system
	// No SpriteComponent exists in this ECS architecture
}

// createGroundEffect spawns a damaging hazard zone at the specified location.
func (s *RaidSystem) createGroundEffect(x, y, radius float64, damage int) {
	hazard := s.world.CreateEntity()
	hazard.AddComponent(&PositionComponent{X: x, Y: y})
	hazard.AddComponent(&HazardComponent{
		HazardType:         HazardPoison, // Use poison as default damaging hazard
		Duration:           10.0,         // 10 second duration
		DamagePerSecond:    float64(damage),
		MovementMultiplier: 1.0,
		Radius:             radius,
		IsLingering:        true,
	})
	hazard.AddComponent(&LifetimeComponent{Duration: 10.0}) // 10 second duration
}

// applyNearbyDebuff applies a slow effect to all players within radius.
func (s *RaidSystem) applyNearbyDebuff(x, y, radius float64) {
	players := s.world.GetEntitiesWith("player", "position")

	for _, player := range players {
		posCompRaw, ok := player.GetComponent("position")
		if !ok {
			continue
		}
		pos := posCompRaw.(*PositionComponent)
		distance := ((pos.X - x) * (pos.X - x)) + ((pos.Y - y) * (pos.Y - y))

		if distance <= radius*radius {
			// Apply slow debuff
			player.AddComponent(&StatusEffectComponent{
				EffectType:   "slow",
				Duration:     5.0,
				Magnitude:    0.5, // 50% slow
				TickInterval: 0,
				NextTick:     0,
			})
		}
	}
}

// applyAoEDamage deals instant damage to all players within radius.
func (s *RaidSystem) applyAoEDamage(x, y, radius float64, damage int) {
	players := s.world.GetEntitiesWith("player", "position", "health")

	for _, player := range players {
		posCompRaw, ok := player.GetComponent("position")
		if !ok {
			continue
		}
		pos := posCompRaw.(*PositionComponent)
		distance := ((pos.X - x) * (pos.X - x)) + ((pos.Y - y) * (pos.Y - y))

		if distance <= radius*radius {
			healthCompRaw, ok := player.GetComponent("health")
			if !ok {
				continue
			}
			health := healthCompRaw.(*HealthComponent)
			health.Current -= float64(damage)
			if health.Current < 0 {
				health.Current = 0
			}
		}
	}
}
