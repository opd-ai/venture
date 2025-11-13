//go:build !android && !ios
// +build !android,!ios

// Package main provides V4.0 system initialization for the dedicated server.
// This file adds all Phase 21-27 systems to the server for full multiplayer support.
package main

import (
	"github.com/opd-ai/venture/pkg/engine"
	"github.com/sirupsen/logrus"
)

// System wrappers for V4 systems that don't match the System interface
// These adapt the simpler Update(deltaTime) signature to Update([]*Entity, deltaTime)

type companionAISystemWrapper struct {
	system *engine.CompanionAISystem
}

func (w *companionAISystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}

type expressionSystemWrapper struct {
	system *engine.ExpressionSystem
}

func (w *expressionSystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}

type expressionComboSystemWrapper struct {
	system *engine.ExpressionComboSystem
}

func (w *expressionComboSystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}

type miniGameSystemWrapper struct {
	system *engine.MiniGameSystem
}

func (w *miniGameSystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}

type achievementSystemWrapper struct {
	system *engine.AchievementSystem
}

func (w *achievementSystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}

// initializeV4Systems adds all V4.0 systems to the server world.
// These systems enable vehicles, companions, books, expanded magic, classes,
// expressions, mini-games, and achievements in multiplayer.
//
// Phase 21: Vehicles & Mounts
// Phase 22: Companions
// Phase 23: Books & Knowledge
// Phase 27: Mini-Games
func initializeV4Systems(world *engine.World, seed int64, logger *logrus.Logger) {
	serverLogger := logger.WithField("component", "v4_systems")

	// Phase 21: Vehicle & Mount Systems
	vehicleCombatSystem := engine.NewVehicleCombatSystem(world)
	world.AddSystem(vehicleCombatSystem)

	// Phase 22: Companion Systems
	companionAISystem := engine.NewCompanionAISystem(world)
	world.AddSystem(&companionAISystemWrapper{system: companionAISystem})

	// Phase 23: Book Systems
	bookReadingSystem := engine.NewBookReadingSystem(world)
	world.AddSystem(bookReadingSystem)

	// Phase 26: Expression Systems
	// Note: ExpressionSystem requires AudioManager which servers don't have
	// Skipping expression systems on server (client-only for now)

	// Phase 27: Mini-Game Systems
	miniGameSystem := engine.NewMiniGameSystem(world)
	world.AddSystem(&miniGameSystemWrapper{system: miniGameSystem})

	// Cross-Phase: Achievement System
	achievementSystem := engine.NewAchievementSystem(world)
	world.AddSystem(&achievementSystemWrapper{system: achievementSystem})

	systemCount := 5 // VehicleCombat, CompanionAI, BookReading, MiniGame, Achievement

	serverLogger.WithFields(logrus.Fields{
		"vehicleSystems":     1, // VehicleCombat
		"companionSystems":   1, // CompanionAI
		"bookSystems":        1, // BookReading
		"miniGameSystems":    1, // MiniGame
		"achievementSystems": 1, // Achievement
		"totalV4Systems":     systemCount,
		"note":               "Expression systems skipped (require AudioManager, client-only)",
	}).Info("V4.0 systems initialized on server")
}
