package engine

import (
	"testing"
)

func TestAdaptiveSoundtrackComponent(t *testing.T) {
	comp := NewAdaptiveSoundtrackComponent("fantasy")

	if comp.CurrentIntensity != IntensityCalm {
		t.Errorf("Initial intensity = %v, want IntensityCalm", comp.CurrentIntensity)
	}

	if comp.GenreTheme != "fantasy" {
		t.Errorf("GenreTheme = %s, want fantasy", comp.GenreTheme)
	}

	if comp.Type() != "adaptive_soundtrack" {
		t.Errorf("Type() = %s, want adaptive_soundtrack", comp.Type())
	}
}

func TestAdaptiveSoundtrackSystemCalm(t *testing.T) {
	world := NewWorld()
	system := NewAdaptiveSoundtrackSystem(world)

	// Create player with no threats
	player := world.CreateEntity()
	player.AddComponent(&PositionComponent{X: 100, Y: 100})
	player.AddComponent(&HealthComponent{Current: 100, Max: 100})
	soundtrack := NewAdaptiveSoundtrackComponent("fantasy")
	player.AddComponent(soundtrack)

	// Process pending entities
	world.Update(0)

	// Update - should remain calm
	for i := 0; i < 10; i++ {
		system.Update(0.1)
	}

	if soundtrack.CurrentIntensity != IntensityCalm {
		t.Errorf("Intensity = %v, want IntensityCalm in safe situation", soundtrack.CurrentIntensity)
	}

	// Ambient layer should be active
	if !soundtrack.ActiveLayers[LayerAmbient] {
		t.Error("Ambient layer should be active in calm state")
	}
}

func TestAdaptiveSoundtrackSystemCombat(t *testing.T) {
	world := NewWorld()
	system := NewAdaptiveSoundtrackSystem(world)

	// Create player
	player := world.CreateEntity()
	player.AddComponent(&PositionComponent{X: 100, Y: 100})
	player.AddComponent(&HealthComponent{Current: 100, Max: 100})
	soundtrack := NewAdaptiveSoundtrackComponent("fantasy")
	soundtrack.CombatThreshold = 2
	player.AddComponent(soundtrack)

	// Create nearby enemies with AI components
	for i := 0; i < 3; i++ {
		enemy := world.CreateEntity()
		enemy.AddComponent(&PositionComponent{X: 150 + float64(i*10), Y: 100})
		enemy.AddComponent(&HealthComponent{Current: 50, Max: 50})
		enemy.AddComponent(&AIComponent{State: AIStateAttack})
	}

	// Process pending entities
	world.Update(0)

	// Update - should increase to combat intensity
	for i := 0; i < 20; i++ {
		system.Update(0.1)
	}

	if soundtrack.CurrentIntensity != IntensityCombat {
		t.Errorf("Intensity = %v, want IntensityCombat with 3 nearby enemies", soundtrack.CurrentIntensity)
	}

	// All layers should be active in combat
	if !soundtrack.ActiveLayers[LayerPercussion] {
		t.Error("Percussion layer should be active in combat")
	}
}

func TestAdaptiveSoundtrackSystemLowHealth(t *testing.T) {
	world := NewWorld()
	system := NewAdaptiveSoundtrackSystem(world)

	// Create player with low health
	player := world.CreateEntity()
	player.AddComponent(&PositionComponent{X: 100, Y: 100})
	player.AddComponent(&HealthComponent{Current: 30, Max: 100}) // 30% health
	soundtrack := NewAdaptiveSoundtrackComponent("fantasy")
	player.AddComponent(soundtrack)

	// Process pending entities
	world.Update(0)

	// Update
	for i := 0; i < 10; i++ {
		system.Update(0.1)
	}

	// Should increase intensity due to low health
	if soundtrack.TargetIntensity < IntensityHigh {
		t.Errorf("TargetIntensity = %v, should be at least IntensityHigh with low health", soundtrack.TargetIntensity)
	}
}

func TestAdaptiveSoundtrackSystemLayerFading(t *testing.T) {
	world := NewWorld()
	system := NewAdaptiveSoundtrackSystem(world)

	player := world.CreateEntity()
	player.AddComponent(&PositionComponent{X: 100, Y: 100})
	player.AddComponent(&HealthComponent{Current: 100, Max: 100})
	soundtrack := NewAdaptiveSoundtrackComponent("fantasy")
	soundtrack.CombatThreshold = 2
	player.AddComponent(soundtrack)

	// Create nearby enemies to maintain combat intensity
	for i := 0; i < 3; i++ {
		enemy := world.CreateEntity()
		enemy.AddComponent(&PositionComponent{X: 150 + float64(i*10), Y: 100})
		enemy.AddComponent(&HealthComponent{Current: 50, Max: 50})
		enemy.AddComponent(&AIComponent{State: AIStateAttack})
	}

	// Process pending entities
	world.Update(0)

	// Update to activate combat layers
	for i := 0; i < 5; i++ {
		system.Update(0.1)
	}

	// Should be at combat intensity now
	if soundtrack.CurrentIntensity != IntensityCombat {
		t.Errorf("CurrentIntensity = %v, want IntensityCombat", soundtrack.CurrentIntensity)
	}

	// Check that percussion layer volume is increasing
	initialVolume := soundtrack.LayerVolumes[LayerPercussion]

	// Update multiple times to fade in
	for i := 0; i < 10; i++ {
		system.Update(0.1)
	}

	finalVolume := soundtrack.LayerVolumes[LayerPercussion]

	if finalVolume <= initialVolume {
		t.Errorf("LayerPercussion volume should increase, got %f -> %f", initialVolume, finalVolume)
	}

	// Should approach 1.0
	if finalVolume < 0.5 {
		t.Errorf("LayerPercussion volume should be at least 0.5 after many updates, got %f", finalVolume)
	}
}

func TestAdaptiveSoundtrackSystemExplorationBonus(t *testing.T) {
	world := NewWorld()
	system := NewAdaptiveSoundtrackSystem(world)

	player := world.CreateEntity()
	player.AddComponent(&PositionComponent{X: 100, Y: 100})
	player.AddComponent(&HealthComponent{Current: 100, Max: 100})
	soundtrack := NewAdaptiveSoundtrackComponent("fantasy")
	player.AddComponent(soundtrack)

	// Set exploration bonus
	system.SetExplorationBonus(player, 1.0)

	// Process pending entities
	world.Update(0)

	// Update
	system.Update(0.1)

	// Should target low intensity due to exploration
	if soundtrack.TargetIntensity != IntensityLow {
		t.Errorf("TargetIntensity = %v, want IntensityLow with exploration bonus", soundtrack.TargetIntensity)
	}

	// Exploration bonus should decay
	initialBonus := soundtrack.ExplorationBonus
	for i := 0; i < 100; i++ {
		system.Update(0.1)
	}

	if soundtrack.ExplorationBonus >= initialBonus {
		t.Errorf("ExplorationBonus should decay, got %f, initial %f", soundtrack.ExplorationBonus, initialBonus)
	}
}

func TestAdaptiveSoundtrackSystemGetCurrentIntensity(t *testing.T) {
	world := NewWorld()
	system := NewAdaptiveSoundtrackSystem(world)

	player := world.CreateEntity()
	player.AddComponent(&PositionComponent{X: 100, Y: 100})
	player.AddComponent(&HealthComponent{Current: 100, Max: 100})
	soundtrack := NewAdaptiveSoundtrackComponent("fantasy")
	soundtrack.CurrentIntensity = IntensityMedium
	player.AddComponent(soundtrack)

	intensity := system.GetCurrentIntensity(player)
	if intensity != IntensityMedium {
		t.Errorf("GetCurrentIntensity() = %v, want IntensityMedium", intensity)
	}
}

func TestAdaptiveSoundtrackSystemTransitionSpeed(t *testing.T) {
	world := NewWorld()
	system := NewAdaptiveSoundtrackSystem(world)

	player := world.CreateEntity()
	player.AddComponent(&PositionComponent{X: 100, Y: 100})
	player.AddComponent(&HealthComponent{Current: 100, Max: 100})
	soundtrack := NewAdaptiveSoundtrackComponent("fantasy")
	soundtrack.CurrentIntensity = IntensityCalm
	soundtrack.CombatThreshold = 2
	player.AddComponent(soundtrack)

	// Create nearby enemies to trigger combat intensity
	for i := 0; i < 3; i++ {
		enemy := world.CreateEntity()
		enemy.AddComponent(&PositionComponent{X: 150 + float64(i*10), Y: 100})
		enemy.AddComponent(&HealthComponent{Current: 50, Max: 50})
		enemy.AddComponent(&AIComponent{State: AIStateAttack})
	}

	// Process pending entities
	world.Update(0)

	// Transition should be gradual
	initialIntensity := soundtrack.CurrentIntensity

	system.Update(0.1)

	// Should have moved toward combat
	if soundtrack.CurrentIntensity == initialIntensity {
		// First update might not have transitioned yet
	}

	// After a few updates, should reach target (combat)
	for i := 0; i < 10; i++ {
		system.Update(0.1)
	}

	if soundtrack.CurrentIntensity != IntensityCombat {
		t.Errorf("After transition, intensity = %v, want IntensityCombat", soundtrack.CurrentIntensity)
	}
}

func TestAdaptiveSoundtrackSystemNilPositionDefense(t *testing.T) {
	world := NewWorld()
	system := NewAdaptiveSoundtrackSystem(world)

	// Create player
	player := world.CreateEntity()
	player.AddComponent(&PositionComponent{X: 100, Y: 100})
	player.AddComponent(&HealthComponent{Current: 100, Max: 100})
	soundtrack := NewAdaptiveSoundtrackComponent("fantasy")
	soundtrack.CombatThreshold = 2
	player.AddComponent(soundtrack)

	// Create enemy with AI and health but without position component
	// This tests the defensive nil check in countNearbyEnemies
	enemy := world.CreateEntity()
	enemy.AddComponent(&HealthComponent{Current: 50, Max: 50})
	enemy.AddComponent(&AIComponent{State: AIStateAttack})
	// Note: No PositionComponent added - this tests nil defense

	// Process pending entities
	world.Update(0)

	// Update should NOT panic even with missing position component
	for i := 0; i < 10; i++ {
		system.Update(0.1)
	}

	// Should remain calm since the enemy without position is skipped
	if soundtrack.CurrentIntensity != IntensityCalm {
		t.Errorf("Intensity = %v, want IntensityCalm when enemies have no position", soundtrack.CurrentIntensity)
	}
}

func BenchmarkAdaptiveSoundtrackSystemUpdate(b *testing.B) {
	world := NewWorld()
	system := NewAdaptiveSoundtrackSystem(world)

	// Create player with soundtrack
	player := world.CreateEntity()
	player.AddComponent(&PositionComponent{X: 100, Y: 100})
	player.AddComponent(&HealthComponent{Current: 100, Max: 100})
	player.AddComponent(NewAdaptiveSoundtrackComponent("fantasy"))

	// Create some enemies
	for i := 0; i < 5; i++ {
		enemy := world.CreateEntity()
		enemy.AddComponent(&PositionComponent{X: 150 + float64(i*20), Y: 100})
		enemy.AddComponent(&HealthComponent{Current: 50, Max: 50})
		enemy.AddComponent(&AIComponent{})
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		system.Update(0.016)
	}
}
