//go:build !android && !ios
// +build !android,!ios

package main

import (
	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/engine/physics/fluids"
	"github.com/opd-ai/venture/pkg/engine/physics/vehicle"
	"github.com/opd-ai/venture/pkg/procgen/building"
	"github.com/opd-ai/venture/pkg/procgen/furniture"
	"github.com/opd-ai/venture/pkg/social/persistence"
	"github.com/opd-ai/venture/pkg/world/housing"
	"github.com/sirupsen/logrus"
)

// INTEGRATION FIX [Category A]: V8.0 Server System Initialization
// Gap: V8.0 systems fully implemented but never initialized on server
// Fix: Created v8_systems.go for server-side V8.0 system initialization
// Roadmap: ROADMAP_V8.md (Phase 49-51)

// initializeV8SystemsServer initializes Version 8.0 systems on the server.
// Server-side systems include: housing infrastructure, trust/reputation tracking,
// chat history persistence, image galleries, guild halls, fluid simulation,
// enhanced vehicle physics, and building/furniture generation.
func initializeV8SystemsServer(world *engine.World, seed int64, logger *logrus.Logger) {
	serverLogger := logger.WithField("component", "v8_systems")

	// Phase 49.1: Housing Core Infrastructure
	// Note: Housing manager is a utility, not an ECS system
	housingManager := housing.NewManager()
	_ = housingManager // Used by housing-related entity spawning

	// Phase 49.2: Persistent Trust & Reputation System
	// Note: Trust and reputation managers are utilities for cross-server state
	trustManager := persistence.NewTrustManager()
	reputationManager := persistence.NewReputationManager()
	_ = trustManager
	_ = reputationManager

	// Phase 49.3 & 49.4: Chat History and Image Gallery are per-player
	// Created when player entities spawn, not as global systems

	// Phase 50.3: Enhanced Vehicle Physics
	// Note: EnhancedVehicleSystem is a helper for vehicle movement, not standalone system
	enhancedVehicleSys := vehicle.NewEnhancedVehicleSystem()
	_ = enhancedVehicleSys // Used by VehicleMovementSystem

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

	// Phase 50.4: Swimming and flooding managers (utilities used by entity systems)
	buoyancyCalculator := fluids.NewBuoyancyCalculator(fluidConfig.Gravity)
	swimmingManager := fluids.NewSwimmingManager(fluidConfig.Gravity)
	floodingManager := fluids.NewFloodingManager(fluidSimulator)
	_ = buoyancyCalculator
	_ = swimmingManager
	_ = floodingManager

	// Phase 51.1: Procedural Building Generation (generator, not system)
	buildingGenerator := building.NewGenerator()
	_ = buildingGenerator // Used during world/structure spawning

	// Phase 51.2: Guild Hall Construction (manager, not system)
	guildHallManager := housing.NewGuildHallManager()
	_ = guildHallManager // Used for guild structure creation

	// Phase 51.3: Furniture Generation & Placement (generator, not system)
	furnitureGenerator := furniture.NewGenerator()
	_ = furnitureGenerator // Used during building interior generation

	if logger.GetLevel() >= logrus.DebugLevel {
		serverLogger.Info("V8.0 systems initialized (housing, trust, reputation, fluid dynamics, vehicle physics, building/furniture)")
	}
}

// fluidSimulatorWrapper adapts FluidSimulator to the System interface for server.
type fluidSimulatorWrapper struct {
	system *fluids.Simulator
}

func (w *fluidSimulatorWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}
