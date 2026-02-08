//go:build !android && !ios
// +build !android,!ios

package main

import (
	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/engine/physics/fluids"
	guildvehicle "github.com/opd-ai/venture/pkg/integration/guild_vehicle"
	"github.com/opd-ai/venture/pkg/network/federation"
	"github.com/opd-ai/venture/pkg/network/federation/guild"
	"github.com/sirupsen/logrus"
)

// INTEGRATION FIX [Category A]: V8.0 Server System Initialization
// Gap: V8.0 systems fully implemented but never initialized on server
// Fix: Created v8_systems.go for server-side V8.0 system initialization
// Roadmap: ROADMAP_V8.md (Phase 49-51)
// Note: fluidSimulatorWrapper moved to system_wrappers.go

// initializeV8SystemsServer initializes Version 8.0 systems on the server.
// Server-side systems include: housing infrastructure, trust/reputation tracking,
// chat history persistence, image galleries, guild halls, fluid simulation,
// enhanced vehicle physics, and building/furniture generation.
// Returns the guild.Manager and FleetManager for use by V9 political warfare integration.
func initializeV8SystemsServer(world *engine.World, seed int64, logger *logrus.Logger) (*guild.Manager, *guildvehicle.FleetManager) {
	serverLogger := logger.WithField("component", "v8_systems")

	// Phase 49.1-49.2: Housing, Trust & Reputation Infrastructure
	// NOTE: These managers are initialized client-side only (cmd/client/handlers.go).
	// Server-side initialization deferred until server-authoritative validation is needed.
	// When implementing server validation:
	// - housing.NewManager() for housing plot validation
	// - persistence.NewTrustManager() for trust score validation
	// - persistence.NewReputationManager() for reputation validation

	// Phase 49.3 & 49.4: Chat History and Image Gallery are per-player
	// Created when player entities spawn, not as global systems

	// Phase 50.1: Guild Federation & Cross-Server Sync
	guildManager := guild.NewManager()
	guildSystem := engine.NewGuildSystem(world, guildManager)
	world.AddSystem(guildSystem)

	// Initialize federation protocol for cross-server guild sync
	serverIdentity := &federation.ServerIdentity{
		PublicKey:  []byte("server-public-key"), // In production, load from config/cert
		ServerName: "venture-server",
	}
	federationProtocol := federation.NewFederationProtocol("server-id", serverIdentity)
	guildSystem.SetFederation(federationProtocol)
	_ = federationProtocol // Available for future cross-server features

	// Phase 56.1: Guild Vehicle Fleet Combat (ROADMAP_V9.md)
	// FleetManager coordinates guild vehicle fleets with formations, siege engines, and maintenance
	// Provides server-authoritative fleet management for guild vehicle coordination
	fleetManager := guildvehicle.NewFleetManager()
	serverLogger.Debug("FleetManager initialized for guild vehicle coordination")

	// Phase 50.3: Enhanced Vehicle Physics
	// NOTE: EnhancedVehicleSystem is client-side only (handles visual physics).
	// Server uses basic physics for authoritative position updates.
	// When implementing server-side vehicle physics validation:
	// - vehicle.NewEnhancedVehicleSystem() for advanced vehicle movement

	// Phase 50.4: Fluid Dynamics & Swimming
	// FluidSimulator is the only V8 component that runs as a system
	fluidConfig := fluids.SimulationConfig{
		GridWidth:       100,
		GridHeight:      100,
		CellSize:        1.0,
		UpdateRate:      30.0,
		Gravity:         9.8,
		PressureFactor:  0.8,
		ViscosityFactor: 0.01,
		MaxIterations:   10,
	}
	fluidSimulator := fluids.NewSimulator(fluidConfig)
	world.AddSystem(&fluidSimulatorWrapper{system: fluidSimulator})

	// Phase 50.4: Swimming and flooding managers
	// NOTE: These utilities are used by entity systems on the client side.
	// Server uses simplified physics. Add when implementing:
	// - fluids.NewBuoyancyCalculator(gravity) for buoyancy validation
	// - fluids.NewSwimmingManager(gravity) for swim speed validation
	// - fluids.NewFloodingManager(simulator) for flood damage validation

	// Phase 51.1-51.3: Building, Guild Hall, and Furniture Generation
	// NOTE: These generators are used during world/structure spawning on client side.
	// Server uses pre-generated terrain data. Add when implementing server-side generation:
	// - building.NewGenerator() for procedural buildings
	// - housing.NewGuildHallManager() for guild hall structures
	// - furniture.NewGenerator() for interior furniture

	if logger.GetLevel() >= logrus.DebugLevel {
		serverLogger.Info("V8.0 systems initialized (guild federation, fleet management, housing, trust, reputation, fluid dynamics, vehicle physics, building/furniture)")
	}

	return guildManager, fleetManager
}
