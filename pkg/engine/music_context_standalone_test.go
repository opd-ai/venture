//go:build test
// +build test

package engine

import (
	"testing"
)

// TestMusicContextIntegrationStandalone verifies the music context system
// integrates correctly with AudioManagerSystem without importing the full package.
func TestMusicContextIntegrationStandalone(t *testing.T) {
	// Create entities with different states
	explorationEntities := []*Entity{
		{id: 1, components: make(map[string]Component)},
	}

	combatEntities := []*Entity{
		{id: 1, components: make(map[string]Component)}, // player
		{id: 2, components: make(map[string]Component)}, // enemy
	}
	combatEntities[0].components["position"] = &struct {
		X, Y float64
		Component
	}{X: 0, Y: 0}
	combatEntities[0].components["health"] = &struct {
		Current, Max float64
		Component
	}{Current: 100, Max: 100}
	combatEntities[0].components["team"] = &struct {
		Team string
		Component
	}{Team: "player"}

	combatEntities[1].components["position"] = &struct {
		X, Y float64
		Component
	}{X: 100, Y: 100}
	combatEntities[1].components["health"] = &struct {
		Current, Max float64
		Component
	}{Current: 50, Max: 50}
	combatEntities[1].components["stats"] = &struct {
		Attack, Defense float64
		Component
	}{Attack: 10, Defense: 5}
	combatEntities[1].components["team"] = &struct {
		Team string
		Component
	}{Team: "enemy"}

	// Create detector
	detector := NewMusicContextDetector()

	// Test exploration detection
	ctx := detector.DetectContext(explorationEntities, explorationEntities[0])
	if ctx != MusicContextExploration {
		t.Errorf("Expected Exploration context, got %s", ctx)
	}

	// Test combat detection
	ctx = detector.DetectContext(combatEntities, combatEntities[0])
	if ctx != MusicContextCombat {
		t.Errorf("Expected Combat context, got %s", ctx)
	}

	// Test transition manager
	manager := NewMusicTransitionManager()

	// Should allow first transition
	if !manager.ShouldTransition(MusicContextCombat) {
		t.Error("Should allow first transition")
	}

	manager.BeginTransition(MusicContextCombat)
	manager.CompleteTransition()

	// Should block immediate re-transition
	if manager.ShouldTransition(MusicContextCombat) {
		t.Error("Should block immediate re-transition")
	}
}
