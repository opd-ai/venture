//go:build !android && !ios
// +build !android,!ios

// system_wrappers.go contains adapter wrappers for ECS systems.
// These wrappers adapt systems with Update(deltaTime) signatures to the
// standard ECS System interface Update([]*Entity, deltaTime).
//
// Code relocated from: v4_systems.go, v8_systems.go
package main

import (
	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/engine/physics/fluids"
	"github.com/opd-ai/venture/pkg/network/federation"
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

type miniGameSystemWrapper struct {
	system *engine.MiniGameSystem
}

func (w *miniGameSystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
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
	w.system.Update(deltaTime)
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
