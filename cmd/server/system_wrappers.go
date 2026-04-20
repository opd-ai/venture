//go:build !android && !ios
// +build !android,!ios

// system_wrappers.go contains adapter wrappers for ECS systems.
// These wrappers adapt systems with Update(deltaTime) signatures to the
// standard ECS System interface Update([]*Entity, deltaTime).
//
// Code relocated from: v4_systems.go, v8_systems.go
package main

import (
	"fmt"

	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/engine/physics/fluids"
	"github.com/opd-ai/venture/pkg/engine/prestige"
	"github.com/opd-ai/venture/pkg/integration/trade_routes"
	"github.com/opd-ai/venture/pkg/network/federation"
	"github.com/opd-ai/venture/pkg/world"
	"github.com/sirupsen/logrus"
)

// V4.0 System Wrappers (originally from v4_systems.go)
// These adapt the simpler Update(deltaTime) signature to Update([]*Entity, deltaTime)

type companionAISystemWrapper struct {
	system *engine.CompanionAISystem
}

func (w *companionAISystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}

type companionProgressionSystemWrapper struct {
	system *engine.CompanionProgressionSystem
}

func (w *companionProgressionSystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}

type companionLoyaltySystemWrapper struct {
	system *engine.CompanionLoyaltySystem
}

func (w *companionLoyaltySystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}

type companionInventorySystemWrapper struct {
	system *engine.CompanionInventorySystem
}

func (w *companionInventorySystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}

type skillInheritanceSystemWrapper struct {
	system *engine.SkillInheritanceSystem
}

func (w *skillInheritanceSystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
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

type reputationSystemWrapper struct {
	system *engine.ReputationSystem
}

func (w *reputationSystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}

type alignmentSystemWrapper struct {
	system *engine.AlignmentSystem
}

func (w *alignmentSystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}

type factionReactionSystemWrapper struct {
	system *engine.FactionReactionSystem
}

func (w *factionReactionSystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}

type moralChoiceSystemWrapper struct {
	system *engine.MoralChoiceSystem
}

func (w *moralChoiceSystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}

type musicTriggerSystemWrapper struct {
	system *engine.MusicTriggerSystem
}

func (w *musicTriggerSystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(entities, deltaTime)
}

type discoverySystemWrapper struct {
	system *engine.DiscoverySystem
}

func (w *discoverySystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}

type achievementSystemWrapper struct {
	system *engine.AchievementSystem
}

func (w *achievementSystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}

type npcDialogSystemWrapper struct {
	system *engine.NPCDialogSystem
}

func (w *npcDialogSystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}

// V5.0 System Wrappers (Social & Communication)

type chatSystemWrapper struct {
	system *engine.ChatSystem
}

func (w *chatSystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}

type enhancedChatSystemWrapper struct {
	system *engine.EnhancedChatSystem
}

func (w *enhancedChatSystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}

type mailSystemWrapper struct {
	system *engine.MailSystem
}

func (w *mailSystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}

type courierSystemWrapper struct {
	system *engine.CourierSystem
}

func (w *courierSystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}

// V6.0 System Wrappers (Persistent Worlds & Federation)

type portalSystemWrapper struct {
	system *federation.PortalSystem
}

func (w *portalSystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}

type bountySystemWrapper struct {
	system *engine.BountySystem
}

func (w *bountySystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}

type politicsSystemWrapper struct {
	system *engine.PoliticsSystem
}

func (w *politicsSystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}

// Core Gameplay System Wrappers (originally from v4_systems.go)
// These systems query world internally, not via entities parameter

type investigationSystemWrapper struct {
	system *engine.InvestigationSystem
}

func (w *investigationSystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}

type merchantCaravanSystemWrapper struct {
	system *engine.MerchantCaravanSystem
}

func (w *merchantCaravanSystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}

type rotationSystemWrapper struct {
	system *engine.RotationSystem
}

func (w *rotationSystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}

type squadSystemWrapper struct {
	system *engine.SquadSystem
}

func (w *squadSystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}

type tradeSystemWrapper struct {
	system *engine.TradeSystem
}

func (w *tradeSystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}

type vehicleSystemWrapper struct {
	system *engine.VehicleSystem
}

func (w *vehicleSystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}

// V8.0 System Wrappers (originally from v8_systems.go)

// fluidSimulatorWrapper adapts FluidSimulator to the System interface for server.
type fluidSimulatorWrapper struct {
	system *fluids.Simulator
}

func (w *fluidSimulatorWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}

// tradeRouteManagerWrapper adapts trade_routes.RouteManager to the System interface for server.
// This enables automated trade route updates via the ECS instead of manual API calls.
type tradeRouteManagerWrapper struct {
	system *trade_routes.RouteManager
}

func (w *tradeRouteManagerWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.UpdateRoutes()
}

// =============================================================================
// Prestige System Wrappers (server-side multiplayer sync)
// AUDIT.md Priority 2: Server-side prestige system registration for multiplayer
// =============================================================================

// prestigeSystemWrapper adapts prestige.System to the engine.System interface.
// This enables prestige data synchronization in multiplayer sessions.
type prestigeSystemWrapper struct {
	system *prestige.System
}

func (w *prestigeSystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	// Convert []*engine.Entity to []prestige.Entity using adapters
	prestigeEntities := make([]prestige.Entity, len(entities))
	for i, e := range entities {
		prestigeEntities[i] = &prestigeEntityAdapter{entity: e}
	}
	w.system.Update(prestigeEntities, deltaTime)
}

// GetSystem returns the underlying prestige.System for direct API access.
// Used for player initialization, XP awards, and paragon point allocation.
func (w *prestigeSystemWrapper) GetSystem() *prestige.System {
	return w.system
}

// prestigeEntityAdapter adapts engine.Entity to prestige.Entity interface.
type prestigeEntityAdapter struct {
	entity *engine.Entity
}

func (a *prestigeEntityAdapter) GetID() string {
	return fmt.Sprintf("%d", a.entity.ID)
}

func (a *prestigeEntityAdapter) HasComponent(componentType string) bool {
	return a.entity.HasComponent(componentType)
}

func (a *prestigeEntityAdapter) GetComponent(componentType string) interface{} {
	comp, _ := a.entity.GetComponent(componentType)
	return comp
}

func (a *prestigeEntityAdapter) AddComponent(component interface{ Type() string }) {
	// Convert to engine.Component (they have the same interface)
	if c, ok := component.(engine.Component); ok {
		a.entity.AddComponent(c)
	}
}

func (a *prestigeEntityAdapter) RemoveComponent(componentType string) {
	a.entity.RemoveComponent(componentType)
}

// Chunk system wrapper — adapts world.ChunkLoaderSystem to ECS System interface.
// ChunkLoaderSystem.Update requires player positions in tile coordinates;
// the wrapper converts pixel positions from entities with "input" component.
// The default tile size is 32 pixels (matching pkg/engine terrain rendering).
const chunkTilePixelSize = 32.0

type chunkLoaderSystemWrapper struct {
	loader *world.ChunkLoaderSystem
}

func (w *chunkLoaderSystemWrapper) Update(entities []*engine.Entity, _ float64) {
	positions := make(map[uint64]struct{ X, Y float64 })
	for _, e := range entities {
		if !e.HasComponent("input") {
			continue
		}
		pos := e.GetPosition()
		if pos == nil {
			continue
		}
		// Convert pixel coordinates to tile coordinates before passing to chunk loader
		positions[e.ID] = struct{ X, Y float64 }{
			pos.X / chunkTilePixelSize,
			pos.Y / chunkTilePixelSize,
		}
	}
	// Always forward the current frame's tracked player positions, including
	// the empty set, so the chunk loader can reconcile disconnects/removals.
	if err := w.loader.Update(positions); err != nil {
		logrus.WithFields(logrus.Fields{
			"system":       "chunk_loader",
			"player_count": len(positions),
		}).WithError(err).Error("chunk loading failed")
	}
}
