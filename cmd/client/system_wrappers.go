//go:build !android && !ios
// +build !android,!ios

// Package main (system_wrappers.go) contains all system wrapper types that adapt
// various game systems to the engine.System interface for registration with the
// ECS World. These wrappers handle signature differences between system-specific
// Update methods and the unified System.Update([]*Entity, float64) interface.
//
// Wrappers are organized by game version/phase:
// - Core Systems (V1-V3)
// - V4.0 Systems (Phase 21-27): Vehicles, Companions, Spells, Classes
// - V5.0 Systems (Phase 32-36): Chat, Mail, Audio, Terrain
// - V6.0 Systems (Phase 39-42): Federation, Portals, Politics
// - V8.0 Systems (Phase 49-51): Housing, Physics, Fluids
// - V19.0 Systems (Phase 99-101): Dormant Package Integration

package main

import (
	"fmt"

	"github.com/opd-ai/venture/pkg/companion/learning"
	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/engine/physics/destruction"
	"github.com/opd-ai/venture/pkg/engine/physics/fluids"
	"github.com/opd-ai/venture/pkg/engine/physics/vehicle"
	"github.com/opd-ai/venture/pkg/engine/prestige"
	"github.com/opd-ai/venture/pkg/integration/world_events"
	"github.com/opd-ai/venture/pkg/network/chat"
	"github.com/opd-ai/venture/pkg/network/federation"
	"github.com/opd-ai/venture/pkg/network/trade"
	"github.com/opd-ai/venture/pkg/world/economy"
	"github.com/sirupsen/logrus"
)

// =============================================================================
// Core System Wrappers
// =============================================================================

// animationSystemWrapper adapts AnimationSystem (returns error) to System interface (no return)
type animationSystemWrapper struct {
	system *engine.AnimationSystem
	logger *logrus.Entry
}

func (w *animationSystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	if err := w.system.Update(entities, deltaTime); err != nil {
		if w.logger != nil && w.logger.Logger.GetLevel() >= logrus.DebugLevel {
			w.logger.WithError(err).Debug("animation system error")
		}
	}
}

// rotationSystemWrapper adapts RotationSystem to System interface
type rotationSystemWrapper struct {
	system *engine.RotationSystem
}

func (w *rotationSystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}

// squadSystemWrapper adapts SquadSystem to System interface
type squadSystemWrapper struct {
	system *engine.SquadSystem
}

func (w *squadSystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}

// reputationSystemWrapper adapts ReputationSystem to System interface
type reputationSystemWrapper struct {
	system *engine.ReputationSystem
}

func (w *reputationSystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}

// alignmentSystemWrapper adapts AlignmentSystem to System interface
type alignmentSystemWrapper struct {
	system *engine.AlignmentSystem
}

func (w *alignmentSystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}

// factionReactionSystemWrapper adapts FactionReactionSystem to System interface
type factionReactionSystemWrapper struct {
	system *engine.FactionReactionSystem
}

func (w *factionReactionSystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}

// =============================================================================
// V4.0 System Wrappers (Phase 21-27): Vehicles, Companions, Spells, Classes
// =============================================================================

// companionAISystemWrapper adapts CompanionAISystem to System interface
type companionAISystemWrapper struct {
	system *engine.CompanionAISystem
}

func (w *companionAISystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}

// companionProgressionSystemWrapper adapts CompanionProgressionSystem to System interface
type companionProgressionSystemWrapper struct {
	system *engine.CompanionProgressionSystem
}

func (w *companionProgressionSystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}

// companionLoyaltySystemWrapper adapts CompanionLoyaltySystem to System interface
type companionLoyaltySystemWrapper struct {
	system *engine.CompanionLoyaltySystem
}

func (w *companionLoyaltySystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}

// companionInventorySystemWrapper adapts CompanionInventorySystem to System interface
type companionInventorySystemWrapper struct {
	system *engine.CompanionInventorySystem
}

func (w *companionInventorySystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}

// skillInheritanceSystemWrapper adapts SkillInheritanceSystem to System interface
type skillInheritanceSystemWrapper struct {
	system *engine.SkillInheritanceSystem
}

func (w *skillInheritanceSystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}

// expressionSystemWrapper adapts ExpressionSystem to System interface
type expressionSystemWrapper struct {
	system *engine.ExpressionSystem
}

func (w *expressionSystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}

// expressionComboSystemWrapper adapts ExpressionComboSystem to System interface
type expressionComboSystemWrapper struct {
	system *engine.ExpressionComboSystem
}

func (w *expressionComboSystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}

// achievementSystemWrapper adapts AchievementSystem to System interface
type achievementSystemWrapper struct {
	system *engine.AchievementSystem
}

func (w *achievementSystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}

// discoverySystemWrapper adapts DiscoverySystem to the System interface.
type discoverySystemWrapper struct {
	system *engine.DiscoverySystem
}

func (w *discoverySystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}

// moralChoiceSystemWrapper adapts MoralChoiceSystem to the System interface.
type moralChoiceSystemWrapper struct {
	system *engine.MoralChoiceSystem
}

func (w *moralChoiceSystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}

// companionSystemWrapper adapts CompanionSystem (high-level manager) to System interface
type companionSystemWrapper struct {
	system *engine.CompanionSystem
}

func (w *companionSystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}

// vehicleSystemWrapper adapts VehicleSystem (high-level manager) to System interface
type vehicleSystemWrapper struct {
	system *engine.VehicleSystem
}

func (w *vehicleSystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}

// adaptiveSoundtrackSystemWrapper adapts AdaptiveSoundtrackSystem to System interface
type adaptiveSoundtrackSystemWrapper struct {
	system *engine.AdaptiveSoundtrackSystem
}

func (w *adaptiveSoundtrackSystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}

// =============================================================================
// V5.0 System Wrappers (Phase 32-36): Chat, Mail, Audio, Terrain
// =============================================================================

type enhancedChatSystemWrapper struct {
	system *engine.EnhancedChatSystem
}

func (w *enhancedChatSystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}

type chatSystemWrapper struct {
	system *engine.EnhancedChatSystem
}

func (w *chatSystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
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

type investigationSystemWrapper struct {
	system *engine.InvestigationSystem
}

func (w *investigationSystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}

type npcDialogSystemWrapper struct {
	system *engine.NPCDialogSystem
}

func (w *npcDialogSystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}

type musicTriggerSystemWrapper struct {
	system *engine.MusicTriggerSystem
}

func (w *musicTriggerSystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(entities, deltaTime)
}

type positionalAudioSystemWrapper struct {
	system *engine.PositionalAudioSystem
}

func (w *positionalAudioSystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}

type reverbSystemWrapper struct {
	system *engine.ReverbSystem
}

func (w *reverbSystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}

type qualitySystemWrapper struct {
	system *engine.QualitySystem
}

func (w *qualitySystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}

type tradeSystemWrapper struct {
	system *engine.TradeSystem
}

func (w *tradeSystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}

type terrainConstructionSystemWrapper struct {
	system *engine.TerrainConstructionSystem
}

func (w *terrainConstructionSystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(entities, deltaTime)
}

type terrainModificationSystemWrapper struct {
	system *engine.TerrainModificationSystem
}

func (w *terrainModificationSystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(entities, deltaTime)
}

type merchantCaravanSystemWrapper struct {
	system *engine.MerchantCaravanSystem
}

func (w *merchantCaravanSystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}

// =============================================================================
// V6.0 System Wrappers (Phase 39-42): Federation, Portals, Politics
// =============================================================================

// portalSystemWrapper adapts PortalSystem to the System interface.
type portalSystemWrapper struct {
	system *federation.PortalSystem
}

func (w *portalSystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}

// bountySystemWrapper adapts BountySystem to the System interface.
type bountySystemWrapper struct {
	system *engine.BountySystem
}

func (w *bountySystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}

// politicsSystemWrapper adapts PoliticsSystem to the System interface.
type politicsSystemWrapper struct {
	system *engine.PoliticsSystem
}

func (w *politicsSystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}

// territorySystemWrapper adapts TerritorySystem to the System interface.
type territorySystemWrapper struct {
	system *engine.TerritorySystem
}

func (w *territorySystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(entities, deltaTime)
}

// =============================================================================
// V8.0 System Wrappers (Phase 49-51): Housing, Physics, Fluids
// =============================================================================

// fluidSimulatorWrapper adapts fluids.Simulator to the System interface.
type fluidSimulatorWrapper struct {
	system *fluids.Simulator
}

func (w *fluidSimulatorWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}

// destructionSystemWrapper adapts destruction.System to the System interface.
type destructionSystemWrapper struct {
	system *destruction.System
}

func (w *destructionSystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}

// =============================================================================
// Network System Wrappers
// =============================================================================

// networkChatSystemWrapper adapts chat.ChatSystem to World.System interface
type networkChatSystemWrapper struct {
	system *chat.ChatSystem
}

func (w *networkChatSystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}

// networkTradeSystemWrapper adapts trade.TradeSystem to World.System interface
type networkTradeSystemWrapper struct {
	system *trade.TradeSystem
}

func (w *networkTradeSystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}

// =============================================================================
// Prestige System Wrapper (Phase 1.3)
// =============================================================================

// prestigeSystemWrapper adapts prestige.System to the engine.System interface.
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

// =============================================================================
// V19.0 System Wrappers (Phase 99-101): Dormant Package Integration
// =============================================================================

// companionLearningSystemWrapper adapts CompanionLearningSystem to the System interface.
type companionLearningSystemWrapper struct {
	system *learning.CompanionLearningSystem
}

func (w *companionLearningSystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}

// worldEconomySystemWrapper adapts economy.System to the System interface.
type worldEconomySystemWrapper struct {
	system *economy.System
}

func (w *worldEconomySystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}

// worldEventManagerWrapper adapts world_events.EventManager to the System interface.
type worldEventManagerWrapper struct {
	system *world_events.EventManager
}

func (w *worldEventManagerWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}

// =============================================================================
// Vehicle Physics System Wrapper (Phase 50.3)
// =============================================================================

// vehicleEntityAdapter adapts engine.Entity to vehicle.Entity interface.
// The vehicle package defines a minimal Entity interface requiring only
// HasComponent and GetComponent, which engine.Entity already implements.
type vehicleEntityAdapter struct {
	entity *engine.Entity
}

func (a *vehicleEntityAdapter) HasComponent(componentType string) bool {
	return a.entity.HasComponent(componentType)
}

func (a *vehicleEntityAdapter) GetComponent(componentType string) interface{} {
	comp, _ := a.entity.GetComponent(componentType)
	return comp
}

// enhancedVehicleSystemWrapper adapts vehicle.EnhancedVehicleSystem to the System interface.
// AUDIT.md REM-019: EnhancedVehicleSystem was initialized but never registered in World.
// This wrapper converts []*engine.Entity to []vehicle.Entity for the system.
type enhancedVehicleSystemWrapper struct {
	system *vehicle.EnhancedVehicleSystem
}

func (w *enhancedVehicleSystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	// Convert []*engine.Entity to []vehicle.Entity
	vehicleEntities := make([]vehicle.Entity, len(entities))
	for i, e := range entities {
		vehicleEntities[i] = &vehicleEntityAdapter{entity: e}
	}
	w.system.Update(vehicleEntities, deltaTime)
}
