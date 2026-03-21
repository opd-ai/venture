package main

import (
	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/procgen/terrain"
)

// terrainSystemConnector holds context for connecting terrain-aware systems.
type terrainSystemConnector struct {
	terrain        *terrain.Terrain
	terrainChecker *engine.TerrainCollisionChecker
	sys            *systemsContainer
}

// connectTerrainChecker connects collision-related systems to the terrain checker.
func (c *terrainSystemConnector) connectTerrainChecker(system engine.System) {
	if collisionSys, ok := system.(*engine.CollisionSystem); ok {
		collisionSys.SetTerrainChecker(c.terrainChecker)
	}
	if projSys, ok := system.(*engine.ProjectileSystem); ok {
		projSys.SetTerrainChecker(c.terrainChecker)
	}
}

// connectTerrainMovementSystems connects terrain movement-related systems.
func (c *terrainSystemConnector) connectTerrainMovementSystems(system engine.System) {
	if terrainMoveSys, ok := system.(*engine.TerrainMovementSpeedSystem); ok {
		terrainMoveSys.SetTerrain(c.terrain)
		c.sys.terrainMovementSpeedSystem = terrainMoveSys
	}
}

// connectTerrainCombatSystems connects terrain combat-related systems.
func (c *terrainSystemConnector) connectTerrainCombatSystems(system engine.System) {
	if terrainCombatSys, ok := system.(*engine.TerrainCombatBonusSystem); ok {
		terrainCombatSys.SetTerrain(c.terrain)
		c.sys.terrainCombatBonusSystem = terrainCombatSys
	}
	if terrainCombatParticleSys, ok := system.(*engine.TerrainCombatBonusParticleSystem); ok {
		if c.sys.terrainCombatBonusSystem != nil {
			terrainCombatParticleSys.SetTerrainCombatBonusSystem(c.sys.terrainCombatBonusSystem)
		}
		c.sys.terrainCombatBonusParticleSystem = terrainCombatParticleSys
	}
}

// connectTerrainStealthSystems connects terrain stealth-related systems.
func (c *terrainSystemConnector) connectTerrainStealthSystems(system engine.System) {
	if terrainStealthSys, ok := system.(*engine.TerrainStealthSystem); ok {
		terrainStealthSys.SetTerrain(c.terrain)
		c.sys.terrainStealthSystem = terrainStealthSys
	}
	if stealthIndicatorParticleSys, ok := system.(*engine.StealthIndicatorParticleSystem); ok {
		if c.sys.terrainStealthSystem != nil {
			stealthIndicatorParticleSys.SetTerrainStealthSystem(c.sys.terrainStealthSystem)
		}
		c.sys.stealthIndicatorParticleSystem = stealthIndicatorParticleSys
	}
	if terrainAmbushCritSys, ok := system.(*engine.TerrainAmbushCritSystem); ok {
		if c.sys.terrainStealthSystem != nil {
			terrainAmbushCritSys.SetTerrainStealthSystem(c.sys.terrainStealthSystem)
		}
		c.sys.terrainAmbushCritSystem = terrainAmbushCritSys
	}
}

// connectTerrainEffectSystems connects terrain effect systems (status, mana, spell damage, etc).
func (c *terrainSystemConnector) connectTerrainEffectSystems(system engine.System) {
	if terrainStatusSys, ok := system.(*engine.TerrainStatusEffectSystem); ok {
		terrainStatusSys.SetTerrain(c.terrain)
		c.sys.terrainStatusEffectSystem = terrainStatusSys
	}
	if terrainManaRegenSys, ok := system.(*engine.TerrainManaRegenSystem); ok {
		terrainManaRegenSys.SetTerrain(c.terrain)
		c.sys.terrainManaRegenSystem = terrainManaRegenSys
	}
	if terrainSpellDamageSys, ok := system.(*engine.TerrainSpellDamageSystem); ok {
		terrainSpellDamageSys.SetTerrain(c.terrain)
		c.sys.terrainSpellDamageSystem = terrainSpellDamageSys
	}
	if terrainEquipDurabilitySys, ok := system.(*engine.TerrainEquipmentDurabilitySystem); ok {
		terrainEquipDurabilitySys.SetTerrain(c.terrain)
		c.sys.terrainEquipmentDurabilitySys = terrainEquipDurabilitySys
	}
	if terrainRangedAccSys, ok := system.(*engine.TerrainRangedAccuracySystem); ok {
		terrainRangedAccSys.SetTerrain(c.terrain)
		c.sys.terrainRangedAccuracySys = terrainRangedAccSys
	}
	if terrainReflectionTintSys, ok := system.(*engine.TerrainReflectionTintSystem); ok {
		terrainReflectionTintSys.SetTerrain(c.terrain)
		c.sys.terrainReflectionTintSystem = terrainReflectionTintSys
	}
	if terrainCompanionBonusSys, ok := system.(*engine.TerrainCompanionBonusSystem); ok {
		terrainCompanionBonusSys.SetTerrain(c.terrain)
		c.sys.terrainCompanionBonusSystem = terrainCompanionBonusSys
	}
}

// connectTerrainParticleSystems connects terrain-aware particle systems.
func (c *terrainSystemConnector) connectTerrainParticleSystems(system engine.System) {
	if footstepParticleSys, ok := system.(*engine.FootstepParticleSystem); ok {
		footstepParticleSys.SetTerrain(c.terrain)
	}
	if ambientEnvSys, ok := system.(*engine.AmbientEnvironmentParticleSystem); ok {
		ambientEnvSys.SetTerrain(c.terrain)
	}
}

// connectTimeOfDaySystems connects time-of-day systems and stores references.
func (c *terrainSystemConnector) connectTimeOfDaySystems(system engine.System) {
	if timeOfDayLightingSys, ok := system.(*engine.TimeOfDayLightingSystem); ok {
		c.sys.timeOfDayLightingSystem = timeOfDayLightingSys
	}
	if timeOfDayStealthSys, ok := system.(*engine.TimeOfDayStealthSystem); ok {
		c.sys.timeOfDayStealthSystem = timeOfDayStealthSys
	}
	if timeOfDayXPBonusSys, ok := system.(*engine.TimeOfDayXPBonusSystem); ok {
		c.sys.timeOfDayXPBonusSystem = timeOfDayXPBonusSys
	}
	if timeOfDayManaCostSys, ok := system.(*engine.TimeOfDayManaCostSystem); ok {
		c.sys.timeOfDayManaCostSystem = timeOfDayManaCostSys
	}
	if timeOfDayCritChanceSys, ok := system.(*engine.TimeOfDayCriticalChanceSystem); ok {
		c.sys.timeOfDayCriticalChanceSystem = timeOfDayCritChanceSys
	}
	if timeOfDayCompanionBonusSys, ok := system.(*engine.TimeOfDayCompanionBonusSystem); ok {
		c.sys.timeOfDayCompanionBonusSystem = timeOfDayCompanionBonusSys
	}
	if timeOfDayManaRegenSys, ok := system.(*engine.TimeOfDayManaRegenSystem); ok {
		c.sys.timeOfDayManaRegenSystem = timeOfDayManaRegenSys
	}
	if timeOfDayBlockChanceSys, ok := system.(*engine.TimeOfDayBlockChanceSystem); ok {
		c.sys.timeOfDayBlockChanceSystem = timeOfDayBlockChanceSys
	}
	if timeOfDayEvasionSys, ok := system.(*engine.TimeOfDayEvasionSystem); ok {
		c.sys.timeOfDayEvasionSystem = timeOfDayEvasionSys
	}
	if timeOfDayAttackSpeedSys, ok := system.(*engine.TimeOfDayAttackSpeedSystem); ok {
		c.sys.timeOfDayAttackSpeedSystem = timeOfDayAttackSpeedSys
	}
	if timeOfDayShadowDirSys, ok := system.(*engine.TimeOfDayShadowDirectionSystem); ok {
		c.sys.timeOfDayShadowDirectionSystem = timeOfDayShadowDirSys
	}
}

// connectAllSystems connects all terrain-aware systems for a single system instance.
func (c *terrainSystemConnector) connectAllSystems(system engine.System) {
	c.connectTerrainChecker(system)
	c.connectTerrainMovementSystems(system)
	c.connectTerrainCombatSystems(system)
	c.connectTerrainStealthSystems(system)
	c.connectTerrainEffectSystems(system)
	c.connectTerrainParticleSystems(system)
	c.connectTimeOfDaySystems(system)
}

// finalizeTimeOfDayConnections connects systems that depend on other time-of-day systems.
func (c *terrainSystemConnector) finalizeTimeOfDayConnections() {
	if c.sys.timeOfDayFishingBonusSystem != nil && c.sys.timeOfDayLightingSystem != nil {
		c.sys.timeOfDayFishingBonusSystem.SetLightingSystem(c.sys.timeOfDayLightingSystem)
	}
	if c.sys.timeOfDayShadowDirectionSystem != nil && c.sys.timeOfDayLightingSystem != nil {
		c.sys.timeOfDayShadowDirectionSystem.SetLightingSystem(c.sys.timeOfDayLightingSystem)
	}
}
